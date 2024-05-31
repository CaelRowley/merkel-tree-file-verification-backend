package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/pkg/fileutil"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/pkg/merkletree"

	"github.com/jackc/pgx/v5"
)

type Handler struct {
	DB *pgx.Conn
}

func (h *Handler) UploadFiles(w http.ResponseWriter, r *http.Request) {
	batchId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	queryParams := r.URL.Query()
	isBatchComplete := queryParams.Get("batch-complete")

	decoder := json.NewDecoder(r.Body)
	var files []fileutil.File
	err = decoder.Decode(&files)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var rows [][]interface{}

	for _, file := range files {
		rows = append(rows, []interface{}{batchId, file.Name, file.Data})
	}

	copyCount, err := h.DB.CopyFrom(
		context.Background(),
		pgx.Identifier{"files"},
		[]string{"batch_id", "name", "file"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if copyCount != int64(len(files)) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if isBatchComplete == "true" {
		err = h.RegenerateTree(batchId)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	query := `SELECT batch_id, name, file FROM files WHERE id = $1`
	var batchId uuid.UUID
	var fileName string
	var fileData []byte

	err := h.DB.QueryRow(context.Background(), query, id).Scan(&batchId, &fileName, &fileData)
	if err == pgx.ErrNoRows {
		fmt.Println("No file found with id:", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	contentType := http.DetectContentType(fileData)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.WriteHeader(http.StatusOK)
	w.Write(fileData)
}

func (h *Handler) DeleteAllFiles(w http.ResponseWriter, r *http.Request) {
	query := `TRUNCATE TABLE files`
	_, err := h.DB.Exec(context.Background(), query)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	query = `ALTER SEQUENCE files_id_seq RESTART WITH 1`
	_, err = h.DB.Exec(context.Background(), query)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetFileProof(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	query := `SELECT batch_id, name, file, original_hash FROM files WHERE id = $1`
	var batchId uuid.UUID
	var fileName string
	var fileData []byte
	var originalHash []byte

	err := h.DB.QueryRow(context.Background(), query, id).Scan(&batchId, &fileName, &fileData, &originalHash)
	if err == pgx.ErrNoRows {
		fmt.Println("No file found with id:", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tree := merkletree.GetTree(batchId)
	fileHash := sha256.Sum256(fileData)

	if originalHash != nil && !bytes.Equal(fileHash[:], originalHash) {
		fileHash = [32]byte(originalHash)
		fmt.Println("Server is acting maliciously and giving a false proof")
	}

	proof, err := merkletree.CreateMerkleProof(tree.Root, fileHash[:])
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(proof)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h *Handler) CorruptFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	decoder := json.NewDecoder(r.Body)
	var file []byte
	err := decoder.Decode(&file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	query := `SELECT batch_id, name, file FROM files WHERE id = $1`
	var batchId uuid.UUID
	var fileName string
	var fileData []byte

	err = h.DB.QueryRow(context.Background(), query, id).Scan(&batchId, &fileName, &fileData)
	if err == pgx.ErrNoRows {
		fmt.Println("No file found with id:", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	originalHash := sha256.Sum256(fileData)

	// Update file content but store the hash of the original, for a malicious server
	_, err = h.DB.Exec(context.Background(), "UPDATE files SET file = $1, original_hash = $2 WHERE id = $3", file, originalHash[:], id)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// If the server was genuine it would rebuild the Merkle tree with the new file
	// err := h.RegenerateTree(batchId)

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) RegenerateTree(batchId uuid.UUID) error {
	query := `SELECT name, file FROM files WHERE batch_id = $1`

	rows, err := h.DB.Query(context.Background(), query, batchId)
	if err == pgx.ErrNoRows {
		fmt.Println("No files found with batch_id:", batchId)
		return err
	}
	defer rows.Close()

	var files [][]byte
	for rows.Next() {
		var name string
		var file []byte
		err := rows.Scan(&name, &file)
		if err != nil {
			return err
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	var fileHashes [][]byte
	for _, file := range files {
		fileHash := sha256.Sum256([]byte(file))
		fileHashes = append(fileHashes, fileHash[:])
	}

	sort.Slice(fileHashes, func(i int, j int) bool {
		return bytes.Compare(fileHashes[i], fileHashes[j]) < 0
	})

	root := merkletree.BuildTree(fileHashes)
	merkletree.UpdateTree(merkletree.MerkleTree{ID: batchId, Root: root})

	return nil
}

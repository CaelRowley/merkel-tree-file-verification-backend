package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/utils/fileutil"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/utils/merkletree"

	"github.com/jackc/pgx/v5"
)

type Handler struct {
	DB *pgx.Conn
}

func (h *Handler) UploadFiles(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var files []fileutil.File
	err := decoder.Decode(&files)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	batchId, err := uuid.NewV7()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var rows [][]interface{}
	var fileHashes [][]byte

	for _, file := range files {
		fileHash := sha256.Sum256([]byte(file.Data))
		fileHashes = append(fileHashes, fileHash[:])
		rows = append(rows, []interface{}{batchId, file.Name, file.Data, fileHash[:]})
	}

	copyCount, err := h.DB.CopyFrom(
		context.Background(),
		pgx.Identifier{"files"},
		[]string{"batch_id", "name", "file", "false_hash"},
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

	root := merkletree.BuildTree(fileHashes)
	merkletree.AddTree(merkletree.MerkleTree{ID: batchId, Root: root})
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
}

func (h *Handler) GetFileProof(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	query := `SELECT batch_id, name, file, false_hash FROM files WHERE id = $1`
	var batchId uuid.UUID
	var fileName string
	var fileData []byte
	var falseHash []byte

	err := h.DB.QueryRow(context.Background(), query, id).Scan(&batchId, &fileName, &fileData, &falseHash)
	if err == pgx.ErrNoRows {
		fmt.Println("No file found with id:", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tree := merkletree.GetTree(batchId)

	realHash := sha256.Sum256(fileData)

	if bytes.Equal(realHash[:], falseHash) {
		fmt.Println("Server is acting maliciously and giving a false proof")
	}

	proof, err := merkletree.CreateMerkleProof(tree.Root, falseHash)
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

	query := `SELECT batch_id FROM files WHERE id = $1`
	var batchId uuid.UUID

	err = h.DB.QueryRow(context.Background(), query, id).Scan(&batchId)
	if err == pgx.ErrNoRows {
		fmt.Println("No file found with id:", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Update file without changing the hash, for a malicious server
	_, err = h.DB.Exec(context.Background(), "UPDATE files SET file = $1 WHERE id = $2", file, id)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// If the server was genuine it would rebuild the Merkle tree with the new file
	// query = `SELECT file FROM files`

	// rows, err := h.DB.Query(context.Background(), query)
	// if err == pgx.ErrNoRows {
	// 	fmt.Println("No file found with id:", id)
	// 	w.WriteHeader(http.StatusNotFound)
	// 	return
	// }

	// defer rows.Close()

	// var files [][]byte
	// for rows.Next() {
	// 	var file []byte
	// 	err := rows.Scan(&file)
	// 	if err != nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		return
	// 	}
	// 	files = append(files, file)
	// }

	// if err := rows.Err(); err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// var fileHashes [][]byte

	// for _, file := range files {
	// 	fileHash := sha256.Sum256([]byte(file))
	// 	fileHashes = append(fileHashes, fileHash[:])
	// }

	// root := merkletree.BuildTree(fileHashes)
	// merkletree.UpdateTree(merkletree.MerkleTree{ID: batchId, Root: root})

	w.WriteHeader(http.StatusOK)
}

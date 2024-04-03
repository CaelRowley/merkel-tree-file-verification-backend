package handlers

import (
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
	batchId, err := uuid.NewV7()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var rows [][]interface{}

	files := fileutil.GetFiles()

	var fileHashes [][]byte
	for _, file := range files {
		rows = append(rows, []interface{}{batchId, file.Name, file.Data})
		fileHash := sha256.Sum256([]byte(file.Data))
		fileHashes = append(fileHashes, fileHash[:])
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
	fmt.Println(copyCount)

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

	// err := os.WriteFile(fileName, fileData, 0644)
	// if err != nil {
	// 	fmt.Println(err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	contentType := http.DetectContentType(fileData)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

	w.WriteHeader(http.StatusOK)
	w.Write(fileData)
}

func (h *Handler) GetFileProof(w http.ResponseWriter, r *http.Request) {
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

	tree := merkletree.GetTree(batchId)

	proof := map[string]interface{}{
		"Tree": tree,
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

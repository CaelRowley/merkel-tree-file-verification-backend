package handlers

import (
	"context"
	"crypto/sha256"
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

	root := merkletree.BuildTree(fileHashes)
	merkletree.AddTree(merkletree.MerkleTree{ID: batchId, Root: root})

	fmt.Println(copyCount)
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	uuid, err := uuid.Parse(idParam)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tree := merkletree.GetTree(uuid)
	fmt.Println(tree)
}

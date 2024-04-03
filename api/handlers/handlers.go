package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/utils/fileutil"

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

	files := fileutil.GetFiles()

	for _, file := range files {
		insertFile := `
			INSERT INTO files (batch_id, name, file)
			VALUES ($1, $2, $3)
		`
		_, err := h.DB.Exec(context.Background(), insertFile, batchId, file.Name, file.Data)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	fmt.Println("Download file: ", idParam)
	fmt.Println(h.DB)
}

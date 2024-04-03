package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/jackc/pgx/v5"
)

type Handler struct {
	DB *pgx.Conn
}

func (h *Handler) UploadFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Upload files")
	fmt.Println(h.DB)
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	fmt.Println("Download file: ", idParam)
	fmt.Println(h.DB)
}

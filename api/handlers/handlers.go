package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type Server struct {
}

func (s *Server) UploadFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Upload files")
}

func (s *Server) DownloadFile(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	fmt.Println("Download file: ", idParam)
}

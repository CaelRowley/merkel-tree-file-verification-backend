package app

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/api/handlers"
)

func (a *App) loadRouter() {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/files", a.loadFileRoutes)

	a.router = router
}

func (a *App) loadFileRoutes(router chi.Router) {
	handlers := &handlers.Handler{
		DB: a.db,
	}

	router.Post("/upload", handlers.UploadFiles)
	router.Get("/download/{id}", handlers.DownloadFile)
	router.Get("/get-proof/{id}", handlers.GetFileProof)
}

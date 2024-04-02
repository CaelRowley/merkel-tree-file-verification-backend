package routes

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gitlab.com/CaelRowley/merkel-tree-file-verification-backend/api/handlers"
)

func LoadRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/files", loadOrderRoutes)

	return router
}

func loadOrderRoutes(router chi.Router) {
	handlers := &handlers.Server{}

	router.Post("/upload", handlers.UploadFiles)
	router.Get("/download/{id}", handlers.DownloadFile)
}

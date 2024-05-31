package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jackc/pgx/v5"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/pkg/handlers"
)

type Server struct {
	router http.Handler
	db     *pgx.Conn
}

var (
	dbURL = os.Getenv("DB_URL")
	port  = "8080"
)

func New() *Server {
	if dbURL == "" {
		dbURL = "postgresql://admin:admin@localhost:5432"
	}

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}

	server := &Server{
		db: conn,
	}

	server.loadRouter()

	return server
}

func (s *Server) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":" + port,
		Handler: s.router,
	}

	err := s.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	// TODO: setup migrations for table creation
	err = s.createTables()
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	defer func() {
		if err := s.db.Close(context.Background()); err != nil {
			fmt.Println("failed to close db", err)
		}
	}()

	fmt.Println("Starting server")

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return server.Shutdown(timeout)
	}
}

func (s *Server) loadRouter() {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/files", s.loadFileRoutes)

	s.router = router
}

func (s *Server) loadFileRoutes(router chi.Router) {
	handlers := &handlers.Handler{
		DB: s.db,
	}

	router.Post("/upload-batch/{id}", handlers.UploadFiles)
	router.Post("/delete-all", handlers.DeleteAllFiles)
	router.Get("/download/{id}", handlers.DownloadFile)
	router.Get("/get-proof/{id}", handlers.GetFileProof)
	router.Post("/corrupt-file/{id}", handlers.CorruptFile)
}

func (s *Server) createTables() error {
	dropTable := `DROP TABLE IF EXISTS files`
	_, err := s.db.Exec(context.Background(), dropTable)
	if err != nil {
		return err
	}

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS files (
				id SERIAL PRIMARY KEY,
				batch_id UUID NOT NULL,
				name TEXT NOT NULL,
				file BYTEA NOT NULL,
				original_hash BYTEA
		)
	`
	_, err = s.db.Exec(context.Background(), createTableQuery)
	if err != nil {
		return err
	}

	return nil
}

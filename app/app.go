package app

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/CaelRowley/merkel-tree-file-verification-backend/api/routes"
)

type App struct {
	router http.Handler
}

func New() *App {
	app := &App{
		router: routes.LoadRouter(),
	}

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":8080",
		Handler: a.router,
	}

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

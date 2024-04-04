package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/app"
)

func main() {
	app := app.New()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println(err)
	}
}

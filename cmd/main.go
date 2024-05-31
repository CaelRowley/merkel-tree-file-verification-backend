package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/cmd/server"
)

func main() {
	server := server.New()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := server.Start(ctx)
	if err != nil {
		fmt.Println(err)
	}
}

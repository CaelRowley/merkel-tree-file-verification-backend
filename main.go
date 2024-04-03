package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/app"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/utils/fileutil"
	"gitlab.com/CaelRowley/merkle-tree-file-verification-backend/utils/merkletree"
)

func main() {
	allHashes := fileutil.GetTestFileHashes()
	root := merkletree.BuildTree(allHashes)
	fmt.Println(root)

	app := app.New()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println(err)
	}
}

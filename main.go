package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/d00p1/filtrate-backups/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, os.Args[1:]); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

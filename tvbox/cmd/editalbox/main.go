package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/marcio/editalbox/tvbox/internal/app"
	"github.com/marcio/editalbox/tvbox/internal/config"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	defer application.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		log.Fatalf("run app: %v", err)
	}
}

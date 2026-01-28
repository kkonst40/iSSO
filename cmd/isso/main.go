package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kkonst40/isso/internal/app"
	"github.com/kkonst40/isso/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config loading error: %v", err.Error())
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("App creating error: %v", err.Error())
	}

	appCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	go func() {
		if err := application.Run(); err != nil {
			log.Fatalf("App running error: %v", err.Error())
		}
	}()

	<-appCtx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application.Shutdown(ctx)
}

package main

import (
	"log"

	"github.com/kkonst40/isso/internal/app"
	"github.com/kkonst40/isso/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config loading error: %v", err.Error())
	}
	log.Println(cfg.JWT.SecretKey)

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("App creating error: %v", err.Error())
	}

	err = application.Run()
	if err != nil {
		log.Fatalf("App running error: %v", err.Error())
	}
}

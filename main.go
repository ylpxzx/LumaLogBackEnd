package main

import (
	"context"
	"log"
	"os"
	"time"

	"lumalog-backend/config"
	"lumalog-backend/database"
	"lumalog-backend/handler"
	"lumalog-backend/patch"
	"lumalog-backend/repository"
	"lumalog-backend/router"
	"lumalog-backend/service"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Open(ctx, cfg)
	if err != nil {
		log.Fatalf("initialize database: %v", err)
	}
	defer db.Close()

	repo := repository.New(db)
	svc := service.New(repo)
	h := handler.New(db, repo, svc, cfg.JWTSecret)

	if len(os.Args) >= 2 && os.Args[1] == "patch" {
		if len(os.Args) < 3 {
			log.Fatalf("missing patch name, available: %s", patch.AvailableNames())
		}
		if err := patch.NewRunner(db, repo).Run(ctx, os.Args[2]); err != nil {
			log.Fatalf("run patch: %v", err)
		}
		log.Printf("patch %q completed", os.Args[2])
		return
	}

	r := router.New(h)
	log.Printf("LumaLog API listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}

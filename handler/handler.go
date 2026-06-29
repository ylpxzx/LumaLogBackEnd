package handler

import (
	"lumalog-backend/repository"
	"lumalog-backend/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB        *pgxpool.Pool
	Repo      *repository.Repository
	Service   *service.Service
	JWTSecret []byte
}

func New(db *pgxpool.Pool, repo *repository.Repository, svc *service.Service, jwtSecret string) *Handler {
	return &Handler{
		DB:        db,
		Repo:      repo,
		Service:   svc,
		JWTSecret: []byte(jwtSecret),
	}
}

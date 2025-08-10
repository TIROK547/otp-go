package main

import (
	"context"
	"log"
	"os"
	"time"

	"otp-go/internal/db"
	"otp-go/internal/server"
)

func main() {
	ctx := context.Background()
	pgURL := os.Getenv("DATABASE_URL")
	redisAddr := os.Getenv("REDIS_ADDR")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret"
	}
	pg, err := db.NewPostgres(ctx, pgURL)
	if err != nil {
		log.Fatal(err)
	}
	rd, err := db.NewRedis(ctx, redisAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := server.NewServer(pg, rd, jwtSecret)
	s := srv.Listen(":8080")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.ListenAndServe(); err != nil && err.Error() != "http: Server closed" {
		log.Fatal(err)
	}
	_ = pg.Close(shutdownCtx)
	_ = rd.Close()
}

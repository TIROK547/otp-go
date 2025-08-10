package server

import (
	"net/http"
	"time"

	"otp-go/internal/auth"
	"otp-go/internal/db"
	"otp-go/internal/handlers"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	pg  *db.Postgres
	rd  *db.Redis
	jwt string
}

func NewServer(pg *db.Postgres, rd *db.Redis, jwtSecret string) *Server {
	return &Server{pg: pg, rd: rd, jwt: jwtSecret}
}

func (s *Server) Listen(addr string) *http.Server {
	r := mux.NewRouter()
	j := auth.NewJWTService(s.jwt)
	otpH := handlers.NewOTPHandler(s.pg, s.rd, j)
	userH := handlers.NewUserHandler(s.pg)
	otpH.RegisterRoutes(r)
	userH.RegisterRoutes(r)
	handler := cors.Default().Handler(r)
	srv := &http.Server{Addr: addr, Handler: handler, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second}
	return srv
}

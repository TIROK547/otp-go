package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"otp-go/internal/db"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	pg *db.Postgres
}

func NewUserHandler(pg *db.Postgres) *UserHandler {
	return &UserHandler{pg: pg}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	u, err := h.pg.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(u)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit := 20
	offset := 0
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = v
		}
	}
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil {
			offset = v
		}
	}
	users, total, err := h.pg.ListUsers(r.Context(), limit, offset, q)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"users": users, "total": total})
}

func (h *UserHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/v1/users/{id}", h.GetUser).Methods("GET")
	r.HandleFunc("/v1/users", h.ListUsers).Methods("GET")
}

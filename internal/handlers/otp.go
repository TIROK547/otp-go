package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"otp-go/internal/auth"
	"otp-go/internal/db"

	"github.com/gorilla/mux"
)

type OTPHandler struct {
	rd         *db.Redis
	pg         *db.Postgres
	jwt        *auth.JWTService
	otpTTL     time.Duration
	rateLimit  int64
	rateWindow time.Duration
}

func NewOTPHandler(pg *db.Postgres, rd *db.Redis, jwt *auth.JWTService) *OTPHandler {
	return &OTPHandler{rd: rd, pg: pg, jwt: jwt, otpTTL: 2 * time.Minute, rateLimit: 3, rateWindow: 10 * time.Minute}
}

func (h *OTPHandler) GenerateOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" {
		http.Error(w, "invalid", http.StatusBadRequest)
		return
	}
	count, err := h.rd.IncrementRate(req.Phone, h.rateWindow)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if count > h.rateLimit {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	otp := h.makeOTP()
	_ = h.rd.SetOTP(req.Phone, otp, h.otpTTL)
	_, err = h.rd.IncrementRate("sent:"+req.Phone, h.rateWindow)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status": "ok", "message": "OTP generated (printed to console)"})
	println("OTP for", req.Phone, ":", otp)
}

func (h *OTPHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		OTP   string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" || req.OTP == "" {
		http.Error(w, "invalid", http.StatusBadRequest)
		return
	}
	stored, err := h.rd.GetOTP(req.Phone)
	if err != nil {
		http.Error(w, "invalid or expired otp", http.StatusUnauthorized)
		return
	}
	if stored != req.OTP {
		http.Error(w, "invalid otp", http.StatusUnauthorized)
		return
	}
	if err := h.rd.DeleteOTP(req.Phone); err != nil {
	}
	u, err := h.pg.CreateUserIfNotExists(r.Context(), req.Phone)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	token, err := h.jwt.GenerateToken(u.ID, u.Phone, 24*time.Hour)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"token": token, "user": u})
}

func (h *OTPHandler) makeOTP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(900000) + 100000
	return strconv.Itoa(n)
}

func (h *OTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *OTPHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/v1/otp/send", h.GenerateOTP).Methods("POST")
	r.HandleFunc("/v1/otp/verify", h.VerifyOTP).Methods("POST")
	r.HandleFunc("/health", h.Health).Methods("GET")
}

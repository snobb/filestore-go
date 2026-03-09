package user

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Handler struct {
	service Service
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("invalid request", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid request",
		})
		return
	}

	if req.Email == "" || req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "email, username and password are required",
		})
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		slog.Error("unable to register user", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "unable to register user",
		})
		return
	}

	token, err := h.service.GenerateToken(user)
	if err != nil {
		slog.Error("unable to generate token", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "unable to generate token",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("invalid request", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid request",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "email and password are required",
		})
		return
	}

	user, err := h.service.ValidatePassword(r.Context(), req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid credentials",
		})
		return
	}

	token, err := h.service.GenerateToken(user)
	if err != nil {
		slog.Error("unable to generate token", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "unable to generate token",
		})
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  user,
	})
}

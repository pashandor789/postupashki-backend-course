package handler

import (
	"encoding/json"
	"hw2/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse the request body || validation kinda
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error": "username and password are required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(req.Username, req.Password)
	if err != nil {
		if err.Error() == "username already taken" {
			http.Error(w, `{"error": "username already taken"}`, http.StatusConflict)
		} else {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	// Automatically log the user in after registration (return a token)
	token, err := h.authService.Login(user.Username, req.Password)
	if err != nil {
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Send success response (201 Created)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{Token: token})

}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error": "username and password are required"}`, http.StatusBadRequest)
		return
	}

	// Call the service to login
	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, `{"error": "invalid username or password"}`, http.StatusUnauthorized)
		return
	}

	// Send success response (200 OK)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}

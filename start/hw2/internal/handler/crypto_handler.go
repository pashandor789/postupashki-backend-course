package handler

import (
	"encoding/json"
	"hw2/internal/models"
	"hw2/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type CryptoHandler struct {
	cryptoService *service.CryptoService
}

type CreateRequest struct {
	Symbol string `json:"symbol"`
}

type CreateResponse struct {
	Crypto *models.Crypto `json:"crypto"`
}

type ListAllResponse struct {
	Cryptos []*models.Crypto `json:"cryptos"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewCryptoHandler(cryptoService *service.CryptoService) *CryptoHandler {
	return &CryptoHandler{cryptoService: cryptoService}
}

// writeError sends a consistent JSON error response
// Uses json.NewEncoder to safely escape special characters
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(ErrorResponse{Error: message}); err != nil {
		// Fallback in case encoding fails (should never happen)
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
	}
}

func (h *CryptoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Symbol == "" {
		writeError(w, http.StatusBadRequest, "symbol cannot be empty")
		return
	}

	crypto, err := h.cryptoService.AddCrypto(req.Symbol)

	if err != nil {
		if err.Error() == "crypto already exists" {
			writeError(w, http.StatusConflict, "crypto already exists")
			return
		}

		writeError(w, http.StatusInternalServerError, "something went wrong while trying to add crypto")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateResponse{Crypto: crypto})
}

func (h *CryptoHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	cryptos := h.cryptoService.ListAllCryptos()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListAllResponse{Cryptos: cryptos})
}

// GetBySymbol handles GET /api/crypto/{symbol}
func (h *CryptoHandler) GetBySymbol(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")

	if symbol == "" {
		writeError(w, http.StatusBadRequest, "symbol is required")
		return
	}

	crypto, err := h.cryptoService.GetCryptoBySymbol(symbol)

	if err != nil {
		writeError(w, http.StatusNotFound, "crypto not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(crypto)
}

func (h *CryptoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")

	if symbol == "" {
		writeError(w, http.StatusBadRequest, "symbol is required")
		return
	}

	err := h.cryptoService.DeleteCrypto(symbol)

	if err != nil {
		writeError(w, http.StatusNotFound, "crypto not found")
		return
	}

	// ✅ FIX: Return empty JSON object, not JSON string
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

// GetHistory handles GET /api/crypto/{symbol}/history
func (h *CryptoHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	if symbol == "" {
		writeError(w, http.StatusBadRequest, "symbol is required")
		return
	}

	history, err := h.cryptoService.GetHistory(symbol)
	if err != nil {
		// ✅ SAFE: err.Error() is safely encoded by json.NewEncoder
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbol":  symbol,
		"history": history,
	})
}

// GetStats handles GET /api/crypto/{symbol}/stats
func (h *CryptoHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	if symbol == "" {
		writeError(w, http.StatusBadRequest, "symbol is required")
		return
	}

	stats, err := h.cryptoService.GetStats(symbol)
	if err != nil {
		// ✅ SAFE: err.Error() is safely encoded by json.NewEncoder
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// Refresh handles POST /api/crypto/{symbol}/refresh
func (h *CryptoHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")

	if symbol == "" {
		writeError(w, http.StatusBadRequest, "symbol is required")
		return
	}

	crypto, err := h.cryptoService.RefreshPrice(symbol)

	if err != nil {
		// ✅ SAFE: err.Error() is safely encoded by json.NewEncoder
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"crypto": crypto,
	})
}

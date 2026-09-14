package repository

import (
	"hw2/internal/models"
	"log"
	"maps"
	"slices"
	"sync"
)

type CryptoRepository struct {
	mu      sync.RWMutex
	cryptos map[string]*models.Crypto
	history map[string][]models.PriceRecord
}

func NewCryptoRepository() *CryptoRepository {
	return &CryptoRepository{
		cryptos: make(map[string]*models.Crypto),
		history: make(map[string][]models.PriceRecord),
	}
}

func (r *CryptoRepository) SaveCrypto(crypto *models.Crypto) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("Saving crypto: %s", crypto.Symbol)
	r.cryptos[crypto.Symbol] = crypto
}

func (r *CryptoRepository) FindCrypto(symbol string) *models.Crypto {
	r.mu.RLock()
	defer r.mu.RUnlock()

	log.Printf("Looking for symbol: %s", symbol)
	log.Printf("Current map keys: %v", getMapKeys(r.cryptos))

	return r.cryptos[symbol]
}

func (r *CryptoRepository) DeleteCrypto(symbol string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.cryptos, symbol)
	delete(r.history, symbol)
}

func (r *CryptoRepository) AddHistory(symbol string, record models.PriceRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hist := r.history[symbol]
	if hist == nil {
		hist = []models.PriceRecord{}
	}

	hist = append(hist, record)

	// Keep only last 100
	if len(hist) > 100 {
		hist = hist[len(hist)-100:] // Drop first 100
	}

	r.history[symbol] = hist
}

// GetHistory returns a DEEP COPY of the history slice to prevent race conditions
func (r *CryptoRepository) GetHistory(symbol string) []models.PriceRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hist := r.history[symbol]
	if hist == nil {
		return []models.PriceRecord{}
	}

	// ✅ FIX: Return a deep copy with a new backing array
	copyHist := make([]models.PriceRecord, len(hist))
	copy(copyHist, hist)
	return copyHist
}

func (r *CryptoRepository) GetAll() []*models.Crypto {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Collect(maps.Values(r.cryptos))
}

func (r *CryptoRepository) Exists(symbol string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.cryptos[symbol]
	return exists
}

// Helper function for debugging
func getMapKeys(m map[string]*models.Crypto) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

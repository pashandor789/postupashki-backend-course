package service

import (
	"errors"
	"fmt"
	"hw2/internal/models"
	"hw2/internal/repository"
	"strings"
	"time"
)

type CryptoService struct {
	repo  *repository.CryptoRepository
	gecko *CoinGeckoClient
}

func NewCryptoService(repo *repository.CryptoRepository, gecko *CoinGeckoClient) *CryptoService {
	return &CryptoService{
		repo:  repo,
		gecko: gecko,
	}
}

func (s *CryptoService) AddCrypto(symbol string) (*models.Crypto, error) {
	// Convert to uppercase
	symbol = strings.ToUpper(symbol)

	// Check if crypto already exists
	if s.repo.Exists(symbol) {
		return nil, errors.New("crypto already exists")
	}

	// Get CoinGecko ID
	coinID, err := s.gecko.GetCoinID(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get CoinGecko ID: %w", err)
	}

	// Fetch real price
	price, err := s.gecko.FetchPrice(coinID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price: %w", err)
	}

	// Create crypto with real data
	crypto := &models.Crypto{
		Symbol:       symbol,
		Name:         symbol, // You could fetch the name from CoinGecko too
		CurrentPrice: price,
		LastUpdated:  time.Now(),
	}

	// Save to repository
	s.repo.SaveCrypto(crypto)
	return crypto, nil
}

func (s *CryptoService) RefreshPrice(symbol string) (*models.Crypto, error) {
	symbol = strings.ToUpper(symbol)

	// Find the crypto
	crypto := s.repo.FindCrypto(symbol)
	if crypto == nil {
		return nil, errors.New("crypto not found")
	}

	// Get CoinGecko ID
	coinID, err := s.gecko.GetCoinID(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get CoinGecko ID: %w", err)
	}

	// Fetch new price
	price, err := s.gecko.FetchPrice(coinID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price: %w", err)
	}

	// Update crypto
	crypto.CurrentPrice = price
	crypto.LastUpdated = time.Now()
	s.repo.SaveCrypto(crypto)

	// Add to history
	s.repo.AddHistory(symbol, models.PriceRecord{
		Price:     price,
		Timestamp: time.Now(),
	})

	return crypto, nil
}

func (s *CryptoService) ListAllCryptos() []*models.Crypto {
	return s.repo.GetAll()
}

func (s *CryptoService) GetCryptoBySymbol(symbol string) (*models.Crypto, error) {
	// ✅ FIX: Convert to uppercase for case-insensitive lookup
	symbol = strings.ToUpper(symbol)

	if exists := s.repo.Exists(symbol); !exists {
		return nil, errors.New("Crypto " + symbol + " not found")
	}

	crypto := s.repo.FindCrypto(symbol)
	return crypto, nil
}

func (s *CryptoService) DeleteCrypto(symbol string) error {
	// ✅ FIX: Convert to uppercase for case-insensitive lookup
	symbol = strings.ToUpper(symbol)

	if exists := s.repo.Exists(symbol); !exists {
		return errors.New("No such crypto yet")
	}

	s.repo.DeleteCrypto(symbol)
	return nil
}

func (s *CryptoService) GetHistory(symbol string) ([]models.PriceRecord, error) {
	symbol = strings.ToUpper(symbol)

	if !s.repo.Exists(symbol) {
		return nil, errors.New("crypto not found")
	}

	history := s.repo.GetHistory(symbol)
	return history, nil
}

// GetStats calculates statistics for a crypto's price history
func (s *CryptoService) GetStats(symbol string) (*models.CryptoStats, error) {
	symbol = strings.ToUpper(symbol)

	// Check if crypto exists
	crypto := s.repo.FindCrypto(symbol)
	if crypto == nil {
		return nil, errors.New("crypto not found")
	}

	// Get history
	history := s.repo.GetHistory(symbol)
	if len(history) == 0 {
		return &models.CryptoStats{
			Symbol:       symbol,
			CurrentPrice: crypto.CurrentPrice,
			Stats: models.Stats{
				MinPrice:           0,
				MaxPrice:           0,
				AvgPrice:           0,
				PriceChange:        0,
				PriceChangePercent: 0,
				RecordsCount:       0,
			},
		}, nil
	}

	// Calculate stats
	var minPrice, maxPrice, sumPrice float64
	minPrice = history[0].Price
	maxPrice = history[0].Price

	for _, record := range history {
		if record.Price < minPrice {
			minPrice = record.Price
		}
		if record.Price > maxPrice {
			maxPrice = record.Price
		}
		sumPrice += record.Price
	}

	avgPrice := sumPrice / float64(len(history))

	// Calculate price change (first vs last)
	firstPrice := history[0].Price
	lastPrice := history[len(history)-1].Price
	priceChange := lastPrice - firstPrice

	var priceChangePercent float64
	if firstPrice != 0 {
		priceChangePercent = (priceChange / firstPrice) * 100
	}

	return &models.CryptoStats{
		Symbol:       symbol,
		CurrentPrice: crypto.CurrentPrice,
		Stats: models.Stats{
			MinPrice:           minPrice,
			MaxPrice:           maxPrice,
			AvgPrice:           avgPrice,
			PriceChange:        priceChange,
			PriceChangePercent: priceChangePercent,
			RecordsCount:       len(history),
		},
	}, nil
}

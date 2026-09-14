package service

import (
	"encoding/json"
	"fmt"
	"hw2/internal/config"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

// CoinGeckoClient handles all CoinGecko API interactions
type CoinGeckoClient struct {
	httpClient *http.Client
	coinCache  map[string]string // symbol -> coin_id (symbol is uppercase)
	apiKey     string
	mu         sync.RWMutex
}

// NewCoinGeckoClient creates a new CoinGecko client
func NewCoinGeckoClient(cfg *config.Config) *CoinGeckoClient {
	return &CoinGeckoClient{
		httpClient: &http.Client{},
		coinCache:  make(map[string]string),
		apiKey:     cfg.CG_KEY,
	}
}

// CoinGeckoCoin represents a coin from the /coins/list endpoint
type CoinGeckoCoin struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

// LoadCoinList fetches the coin list and caches it
func (c *CoinGeckoClient) LoadCoinList() error {
	url := "https://api.coingecko.com/api/v3/coins/list"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return err
	}

	// Add API key header if provided
	if c.apiKey != "" {
		req.Header.Set("x-cg-demo-api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("CoinGecko API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return err
	}

	// Parse JSON into slice of CoinGeckoCoin
	var coins []CoinGeckoCoin
	if err := json.Unmarshal(body, &coins); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return err
	}

	// Cache the coins (symbol -> id)
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, coin := range coins {
		// Convert symbol to uppercase for case-insensitive lookups
		symbol := strings.ToUpper(coin.Symbol)
		c.coinCache[symbol] = coin.ID
	}

	log.Printf("Loaded %d coins from CoinGecko", len(coins))
	return nil
}

// GetCoinID returns the CoinGecko ID for a symbol
func (c *CoinGeckoClient) GetCoinID(symbol string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Convert to uppercase for lookup
	symbol = strings.ToUpper(symbol)

	id, exists := c.coinCache[symbol]
	if !exists {
		return "", fmt.Errorf("coin symbol '%s' not found in CoinGecko cache", symbol)
	}

	return id, nil
}

// FetchPrice fetches the current price for a coin ID in USD
func (c *CoinGeckoClient) FetchPrice(coinID string) (float64, error) {
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", coinID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("x-cg-demo-api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("CoinGecko API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response: %w", err)
	}

	// Parse the price response
	// Response format: {"bitcoin":{"usd":45000.50}}
	var priceResponse map[string]map[string]float64
	if err := json.Unmarshal(body, &priceResponse); err != nil {
		return 0, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Extract the price
	coinData, exists := priceResponse[coinID]
	if !exists {
		return 0, fmt.Errorf("coin ID '%s' not found in price response", coinID)
	}

	price, exists := coinData["usd"]
	if !exists {
		return 0, fmt.Errorf("USD price not found for coin ID '%s'", coinID)
	}

	return price, nil
}

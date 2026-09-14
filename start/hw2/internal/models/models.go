package models

import (
	"time"
)

type User struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashedpassword"`
}

type Crypto struct {
	Symbol       string    `json:"symbol"`
	Name         string    `json:"name"`
	CurrentPrice float64   `json:"current_price"`
	LastUpdated  time.Time `json:"last_updated"`
}

type PriceRecord struct {
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// Stats contains all statistical calculations
type Stats struct {
	MinPrice           float64 `json:"min_price"`
	MaxPrice           float64 `json:"max_price"`
	AvgPrice           float64 `json:"avg_price"`
	PriceChange        float64 `json:"price_change"`
	PriceChangePercent float64 `json:"price_change_percent"`
	RecordsCount       int     `json:"records_count"`
}

// CryptoStats is the response for the /stats endpoint
type CryptoStats struct {
	Symbol       string  `json:"symbol"`
	CurrentPrice float64 `json:"current_price"`
	Stats        Stats   `json:"stats"`
}

// ScheduleConfig represents the auto-refresh schedule settings
type ScheduleConfig struct {
	Enabled         bool      `json:"enabled"`
	IntervalSeconds int       `json:"interval_seconds"`
	LastUpdate      time.Time `json:"last_update"`
	NextUpdate      time.Time `json:"next_update"`
}

// ScheduleRequest is the request body for PUT /schedule
type ScheduleRequest struct {
	Enabled         *bool `json:"enabled"`          // Pointer to allow optional fields
	IntervalSeconds *int  `json:"interval_seconds"` // Pointer to allow optional fields
}

// ScheduleResponse is the response for GET /schedule
type ScheduleResponse struct {
	Enabled         bool      `json:"enabled"`
	IntervalSeconds int       `json:"interval_seconds"`
	LastUpdate      time.Time `json:"last_update"`
	NextUpdate      time.Time `json:"next_update"`
}

// TriggerResponse is the response for POST /schedule/trigger
type TriggerResponse struct {
	UpdatedCount int       `json:"updated_count"`
	Timestamp    time.Time `json:"timestamp"`
}

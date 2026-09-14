package service

import (
	"errors"
	"hw2/internal/models"
	"sync"
	"time"
)

type ScheduleService struct {
	mu            sync.RWMutex
	config        *models.ScheduleConfig
	cryptoService *CryptoService
	cancel        chan struct{} // Renamed from stopChan for clarity
	ticker        *time.Ticker
	wg            sync.WaitGroup // Track goroutine lifetime
	isRunning     bool           // Track if scheduler is actively running
}

func NewScheduleService(cryptoService *CryptoService) *ScheduleService {
	return &ScheduleService{
		config: &models.ScheduleConfig{
			Enabled:         true,
			IntervalSeconds: 30, // Default: 30 seconds
			LastUpdate:      time.Now(),
			NextUpdate:      time.Now().Add(30 * time.Second),
		},
		cryptoService: cryptoService,
		cancel:        make(chan struct{}),
	}
}

// Start begins the background scheduler
func (s *ScheduleService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If already running, stop everything first
	if s.isRunning {
		s.stopTickerLocked()
	}

	if s.config.Enabled {
		s.startTickerLocked()
	}
}

// stopTickerLocked stops the ticker and signals the goroutine to exit
// Must be called with the lock held
func (s *ScheduleService) stopTickerLocked() {
	// Stop the ticker first
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}

	// Signal the goroutine to stop by closing the channel
	if s.cancel != nil {
		close(s.cancel)
		s.cancel = nil
	}

	s.isRunning = false
}

// startTickerLocked starts a new ticker and goroutine
// Must be called with the lock held
func (s *ScheduleService) startTickerLocked() {
	// Ensure any old goroutine is cleaned up first
	if s.isRunning {
		s.stopTickerLocked()
	}

	interval := time.Duration(s.config.IntervalSeconds) * time.Second
	s.ticker = time.NewTicker(interval)

	// Create new cancel channel
	s.cancel = make(chan struct{})
	s.isRunning = true

	// Start the goroutine with WaitGroup tracking
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ticker.C:
				s.performUpdate()
			case <-s.cancel:
				// Channel closed, exit goroutine
				return
			}
		}
	}()

	// Update next update time
	s.config.NextUpdate = time.Now().Add(interval)
}

// performUpdate refreshes prices for all cryptos
func (s *ScheduleService) performUpdate() {
	// Don't hold the lock during the entire operation
	// Just check if enabled and get cryptos
	s.mu.RLock()
	if !s.config.Enabled {
		s.mu.RUnlock()
		return
	}
	cryptos := s.cryptoService.ListAllCryptos()
	s.mu.RUnlock()

	if len(cryptos) == 0 {
		return
	}

	// Refresh each crypto without holding the lock
	for _, crypto := range cryptos {
		_, err := s.cryptoService.RefreshPrice(crypto.Symbol)
		if err != nil {
			// Log error but continue with other cryptos
			// In production, you'd want proper logging here
			continue
		}
	}

	// Update timestamps after all refreshes
	s.mu.Lock()
	s.config.LastUpdate = time.Now()
	s.config.NextUpdate = time.Now().Add(time.Duration(s.config.IntervalSeconds) * time.Second)
	s.mu.Unlock()
}

// TriggerManualUpdate forces an immediate update
func (s *ScheduleService) TriggerManualUpdate() (int, error) {
	cryptos := s.cryptoService.ListAllCryptos()
	if len(cryptos) == 0 {
		return 0, nil
	}

	updatedCount := 0
	for _, crypto := range cryptos {
		_, err := s.cryptoService.RefreshPrice(crypto.Symbol)
		if err == nil {
			updatedCount++
		}
	}

	// Update last update time
	s.mu.Lock()
	s.config.LastUpdate = time.Now()
	if s.config.Enabled {
		s.config.NextUpdate = time.Now().Add(time.Duration(s.config.IntervalSeconds) * time.Second)
	}
	s.mu.Unlock()

	return updatedCount, nil
}

// GetConfig returns the current schedule configuration
func (s *ScheduleService) GetConfig() *models.ScheduleConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid external modifications
	config := *s.config
	return &config
}

// UpdateConfig updates the schedule configuration
func (s *ScheduleService) UpdateConfig(updated *models.ScheduleRequest) error {
	s.mu.Lock()

	// Update enabled if provided
	if updated.Enabled != nil {
		s.config.Enabled = *updated.Enabled
	}

	// Update interval if provided (with validation)
	if updated.IntervalSeconds != nil {
		interval := *updated.IntervalSeconds
		if interval < 10 || interval > 3600 {
			s.mu.Unlock()
			return errors.New("interval must be between 10 and 3600 seconds")
		}
		s.config.IntervalSeconds = interval
	}

	// Store old cancel channel and ticker for cleanup
	oldCancel := s.cancel
	oldTicker := s.ticker
	wasRunning := s.isRunning

	// Stop old scheduler completely
	if wasRunning {
		// Signal old goroutine to stop
		if oldCancel != nil {
			close(oldCancel)
			s.cancel = nil
		}
		if oldTicker != nil {
			oldTicker.Stop()
			s.ticker = nil
		}
		s.isRunning = false
	}

	// Start new scheduler if enabled
	if s.config.Enabled {
		s.startTickerLocked()
	}

	s.mu.Unlock()

	// Wait for old goroutine to finish (optional but good practice)
	// This ensures the old goroutine has truly exited before we continue
	if wasRunning && oldCancel != nil {
		// Wait with timeout to avoid deadlock
		done := make(chan struct{})
		go func() {
			s.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Old goroutine finished successfully
		case <-time.After(5 * time.Second):
			// Timeout - old goroutine might be stuck, but we can proceed
			// In production, you'd log this warning
		}
	}

	return nil
}

// Stop gracefully stops the scheduler
func (s *ScheduleService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		s.stopTickerLocked()
		// Wait for goroutine to finish
		s.mu.Unlock()
		s.wg.Wait()
		s.mu.Lock()
	}
}

// IsRunning returns whether the scheduler is currently running
func (s *ScheduleService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetStatus returns the current status of the scheduler
func (s *ScheduleService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"enabled":      s.config.Enabled,
		"is_running":   s.isRunning,
		"interval_sec": s.config.IntervalSeconds,
		"last_update":  s.config.LastUpdate,
		"next_update":  s.config.NextUpdate,
		"crypto_count": len(s.cryptoService.ListAllCryptos()),
	}
}

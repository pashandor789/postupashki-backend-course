package handler

import (
	"encoding/json"
	"hw2/internal/models"
	"hw2/internal/service"
	"net/http"
	"time"
)

// ScheduleHandler handles schedule-related HTTP requests
type ScheduleHandler struct {
	scheduleService *service.ScheduleService
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(scheduleService *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
	}
}

// Get handles GET /schedule
func (h *ScheduleHandler) Get(w http.ResponseWriter, r *http.Request) {
	config := h.scheduleService.GetConfig()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.ScheduleResponse{
		Enabled:         config.Enabled,
		IntervalSeconds: config.IntervalSeconds,
		LastUpdate:      config.LastUpdate,
		NextUpdate:      config.NextUpdate,
	})
}

// Update handles PUT /schedule
func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req models.ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Validate at least one field is provided
	if req.Enabled == nil && req.IntervalSeconds == nil {
		http.Error(w, `{"error": "at least one field (enabled or interval_seconds) must be provided"}`, http.StatusBadRequest)
		return
	}

	// Validate interval if provided
	if req.IntervalSeconds != nil {
		interval := *req.IntervalSeconds
		if interval < 10 || interval > 3600 {
			http.Error(w, `{"error": "interval must be between 10 and 3600 seconds"}`, http.StatusBadRequest)
			return
		}
	}

	if err := h.scheduleService.UpdateConfig(&req); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Get updated config to return
	config := h.scheduleService.GetConfig()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled":          config.Enabled,
		"interval_seconds": config.IntervalSeconds,
	})
}

// Trigger handles POST /schedule/trigger
func (h *ScheduleHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	updatedCount, err := h.scheduleService.TriggerManualUpdate()
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.TriggerResponse{
		UpdatedCount: updatedCount,
		Timestamp:    time.Now(),
	})
}

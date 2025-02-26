package http

import (
	"http_server/api/http/types"
	"http_server/usecases"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Object struct {
	service usecases.Object
}

func NewHandler(service usecases.Object) *Object {
	return &Object{service: service}
}

// @Summary Get a result
// @Description Get a result by UUID
// @Tags object
// @Accept json
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} types.GetResultObjectHandlerResponse
// @Failure 400 {string} string "Bad request"
// @Failure 404 {string} string "Object not found"
// @Router /result/{task_id} [get]
func (s *Object) getResultHandler(w http.ResponseWriter, r *http.Request) {
	req, err := types.CreateGetObjectHandlerRequest(r)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	value, err := s.service.GetResult(req.Key)
	types.ProcessError(w, err, &types.GetResultObjectHandlerResponse{Result: value})
}

// @Summary Get a status
// @Description Get a status by UUID
// @Tags object
// @Accept json
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} types.GetStatusObjectHandlerResponse
// @Failure 400 {string} string "Bad request"
// @Failure 404 {string} string "Object not found"
// @Router /status/{task_id} [get]
func (s *Object) getStatusHandler(w http.ResponseWriter, r *http.Request) {
	req, err := types.CreateGetObjectHandlerRequest(r)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	value, err := s.service.GetStatus(req.Key)
	types.ProcessError(w, err, &types.GetStatusObjectHandlerResponse{Status: value})
}

// @Summary Create or update an object
// @Description Create or update an object with the provided key, result, and status
// @Tags object
// @Accept json
// @Produce json
// @Param request body types.PutObjectHandlerRequest true "Object data"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /task [put]
func (s *Object) putHandler(w http.ResponseWriter, r *http.Request) {
	req, err := types.CreatePutObjectHandlerRequest(r)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	err = s.service.Put(req.Key, req.Result, req.Status)
	types.ProcessError(w, err, nil)
}

// @Summary Create a new object
// @Description Create a new object with the provided key, result, and status
// @Tags object
// @Accept json
// @Produce json
// @Param request body types.PostObjectHandlerRequest true "Object data"
// @Success 200 {object} types.PostObjectHandlerResponse
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /task [post]
func (s *Object) postHandler(w http.ResponseWriter, r *http.Request) {
	req, err := types.CreatePostObjectHandlerRequest(r)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	err = s.service.Post(req.Key, req.Result, req.Status)
	types.ProcessError(w, err, &types.PostObjectHandlerResponse{Key: &req.Key})

	// Бурная деятельность
	time.Sleep(2 * time.Second)

	err = s.service.Put(req.Key, rand.Intn(100), "ready")
	types.ProcessError(w, err, nil)
}

// @Summary Delete an object
// @Description Delete an object by UUID
// @Tags object
// @Accept json
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad request"
// @Failure 404 {string} string "Object not found"
// @Router /task/{task_id} [delete]
func (s *Object) deleteHandler(w http.ResponseWriter, r *http.Request) {
	req, err := types.CreateDeleteObjectHandlerRequest(r)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	err = s.service.Delete(req.Key)
	types.ProcessError(w, err, nil)
}

func (s *Object) WithObjectHandlers(r chi.Router) {
	r.Route("/task", func(r chi.Router) {
		r.Post("/", s.postHandler)
	})

	r.Route("/status", func(r chi.Router) {
		r.Get("/{task_id}", s.getStatusHandler)
	})

	r.Route("/result", func(r chi.Router) {
		r.Get("/{task_id}", s.getResultHandler)
	})
}

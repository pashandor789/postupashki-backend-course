package types

import (
	"encoding/json"
	"fmt"
	"http_server/domain"
	"http_server/repository"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GetObjectHandlerRequest struct {
	Key string `json:"key"`
}

func CreateGetObjectHandlerRequest(r *http.Request) (*GetObjectHandlerRequest, error) {
	key := chi.URLParam(r, "task_id")
	if key == "" {
		return nil, fmt.Errorf("missing key")
	}
	return &GetObjectHandlerRequest{Key: key}, nil
}

type PutObjectHandlerRequest struct {
	domain.Object
}

func CreatePutObjectHandlerRequest(r *http.Request) (*PutObjectHandlerRequest, error) {
	var req PutObjectHandlerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("error while decoding json: %v", err)
	}
	return &req, nil
}

type DeleteObjectHandlerRequest struct {
	Key string `json:"key"`
}

func CreateDeleteObjectHandlerRequest(r *http.Request) (*DeleteObjectHandlerRequest, error) {
	key := r.URL.Query().Get("key")
	if key == "" {
		return nil, fmt.Errorf("missing key")
	}
	return &DeleteObjectHandlerRequest{Key: key}, nil
}

type PostObjectHandlerRequest struct {
	domain.Object
}

func CreatePostObjectHandlerRequest(r *http.Request) (*PostObjectHandlerRequest, error) {
	key := uuid.New().String()
	req := PostObjectHandlerRequest{
		Object: domain.Object{
			Key:    key,
			Result: -1,
			Status: "in_progress",
		},
	}
	return &req, nil
}

type GetResultObjectHandlerResponse struct {
	Result *int `json:"result"`
}

type GetStatusObjectHandlerResponse struct {
	Status *string `json:"status"`
}

type PostObjectHandlerResponse struct {
	Key *string `json:"key"`
}

func ProcessError(w http.ResponseWriter, err error, resp any) {
	if err == repository.NotFound {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if resp != nil {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
	}
}

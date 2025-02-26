package service

import (
	"http_server/repository"
)

type Object struct {
	repo repository.Object
}

func NewObject(repo repository.Object) *Object {
	return &Object{
		repo: repo,
	}
}

func (rs *Object) GetStatus(key string) (*string, error) {
	return rs.repo.GetStatus(key)
}

func (rs *Object) GetResult(key string) (*int, error) {
	return rs.repo.GetResult(key)
}

func (rs *Object) Put(key string, result int, status string) error {
	return rs.repo.Put(key, result, status)
}

func (rs *Object) Post(key string, result int, status string) error {
	return rs.repo.Post(key, result, status)
}

func (rs *Object) Delete(key string) error {
	return rs.repo.Delete(key)
}

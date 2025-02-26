package ram_storage

import (
	"errors"
	"http_server/repository"
)

type Pair struct {
	result int
	status string
}

type Object struct {
	data map[string]Pair
}

func NewObject() *Object {
	return &Object{
		data: make(map[string]Pair),
	}
}

func (rs *Object) GetStatus(key string) (*string, error) {
	value, exists := rs.data[key]
	if !exists {
		return nil, repository.NotFound
	}
	return &value.status, nil
}

func (rs *Object) GetResult(key string) (*int, error) {
	value, exists := rs.data[key]
	if !exists {
		return nil, repository.NotFound
	}
	return &value.result, nil
}

func (rs *Object) Put(key string, result int, status string) error {
	rs.data[key] = Pair{result: result, status: status}
	return nil
}

func (rs *Object) Post(key string, result int, status string) error {
	if _, exists := rs.data[key]; exists {
		return errors.New("key already exists")
	}
	rs.data[key] = Pair{result: result, status: status}
	return nil
}

func (rs *Object) Delete(key string) error {
	if _, exists := rs.data[key]; !exists {
		return errors.New("key not found")
	}
	delete(rs.data, key)
	return nil
}

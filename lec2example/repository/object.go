package repository

type Object interface {
	GetStatus(key string) (*string, error)
	GetResult(key string) (*int, error)
	Put(key string, result int, status string) error
	Post(key string, result int, status string) error
	Delete(key string) error
}

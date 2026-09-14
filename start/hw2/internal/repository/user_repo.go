package repository

import (
	"hw2/internal/models"
	"sync"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*models.User // Key: Username
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*models.User),
	}
}

func (r *UserRepository) SaveUser(user *models.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.Username] = user
}

func (r *UserRepository) FindUser(username string) *models.User {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.users[username]
}

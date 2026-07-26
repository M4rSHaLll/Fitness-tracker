package memory

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"fmt"
	"sync"
)

type UserRepository struct {
	mu     sync.RWMutex
	users  map[int64]*model.User
	nextID int64
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:  make(map[int64]*model.User),
		nextID: 1,
	}
}

func (r *UserRepository) Create(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existingUser := range r.users {
		if existingUser.TelegramID == user.TelegramID {
			return fmt.Errorf("user already exists: %w", repository.ErrAlreadyExists)
		}
	}

	user.ID = r.nextID
	r.users[user.ID] = user
	r.nextID++
	return nil
}

func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %w", repository.ErrNotFound)
	}

	return user, nil
}

func (r *UserRepository) GetByTelegramID(telegramID int64) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.TelegramID == telegramID {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found: %w", repository.ErrNotFound)
}

func (r *UserRepository) Update(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.users[user.ID]
	if !exists {
		return fmt.Errorf("user not found: %w", repository.ErrNotFound)
	}

	r.users[user.ID] = user
	return nil
}

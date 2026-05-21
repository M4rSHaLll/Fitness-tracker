package memory

import (
	"Fitness-tracker/internal/model"
	"errors"
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
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (r *UserRepository) Update(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.users[user.ID]
	if !exists {
		return errors.New("user not found")
	}

	r.users[user.ID] = user
	return nil
}

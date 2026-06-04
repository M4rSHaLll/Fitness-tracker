package repository

import (
	"Fitness-tracker/internal/model"
	"errors"
)

var ErrNotFound = errors.New("not found")

type UserRepository interface {
	Create(user *model.User) error
	GetByID(id int64) (*model.User, error)
	Update(user *model.User) error
}

type WorkoutRepository interface {
	Create(workout *model.Workout) error
	GetByID(id int64) (*model.Workout, error)
	GetByUserID(userID int64) ([]*model.Workout, error)
	Delete(id int64) error
}

type SetRepository interface {
	Create(set *model.Set) error
	GetByWorkoutID(workoutID int64) ([]*model.Set, error)
	Delete(id int64) error
}

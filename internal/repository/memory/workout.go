package memory

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"fmt"
	"sync"
)

type WorkoutRepository struct {
	workouts map[int64]*model.Workout
	mu       sync.RWMutex
	nextID   int64
}

func NewWorkoutRepository() *WorkoutRepository {
	return &WorkoutRepository{
		workouts: make(map[int64]*model.Workout),
		nextID:   1,
	}
}

func (r *WorkoutRepository) Create(workout *model.Workout) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	workout.ID = r.nextID
	r.workouts[workout.ID] = workout
	r.nextID++
	return nil
}

func (r *WorkoutRepository) GetByID(id int64) (*model.Workout, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	workout, exists := r.workouts[id]
	if !exists {
		return nil, fmt.Errorf("workout not found: %w", repository.ErrNotFound)
	}

	return workout, nil
}

func (r *WorkoutRepository) GetByUserID(userID int64) ([]*model.Workout, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userWorkouts []*model.Workout
	for _, workout := range r.workouts {
		if workout.UserID == userID {
			userWorkouts = append(userWorkouts, workout)
		}
	}

	return userWorkouts, nil
}

func (r *WorkoutRepository) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.workouts[id]
	if !exists {
		return fmt.Errorf("workout not found: %w", repository.ErrNotFound)
	}

	delete(r.workouts, id)
	return nil
}

package memory

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type ExerciseRepository struct {
	mu        sync.RWMutex
	exercises map[int64]*model.Exercise
	nextID    int64
}

func NewExerciseRepository() *ExerciseRepository {
	return &ExerciseRepository{
		exercises: make(map[int64]*model.Exercise),
		nextID:    1,
	}
}

func (r *ExerciseRepository) Create(exercise *model.Exercise) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existingExercise := range r.exercises {
		if strings.EqualFold(existingExercise.Name, exercise.Name) {
			return fmt.Errorf("exercise already exists: %w", repository.ErrAlreadyExists)
		}
	}

	exercise.ID = r.nextID
	r.exercises[exercise.ID] = exercise
	r.nextID++
	return nil
}

func (r *ExerciseRepository) GetByID(id int64) (*model.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	exercise, exists := r.exercises[id]
	if !exists {
		return nil, fmt.Errorf("exercise not found: %w", repository.ErrNotFound)
	}

	return exercise, nil
}

func (r *ExerciseRepository) GetAll() ([]*model.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	exercises := make([]*model.Exercise, 0, len(r.exercises))
	for _, exercise := range r.exercises {
		exercises = append(exercises, exercise)
	}
	sort.Slice(exercises, func(i, j int) bool {
		return exercises[i].ID < exercises[j].ID
	})

	return exercises, nil
}

func (r *ExerciseRepository) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.exercises[id]
	if !exists {
		return fmt.Errorf("exercise not found: %w", repository.ErrNotFound)
	}

	delete(r.exercises, id)
	return nil
}

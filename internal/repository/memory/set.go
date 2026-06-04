package memory

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"fmt"
	"sync"
)

type SetRepository struct {
	mu     sync.RWMutex
	sets   map[int64]*model.Set
	nextID int64
}

func NewSetRepository() *SetRepository {
	return &SetRepository{
		sets:   make(map[int64]*model.Set),
		nextID: 1,
	}
}

func (r *SetRepository) Create(set *model.Set) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	set.ID = r.nextID
	r.sets[set.ID] = set
	r.nextID++
	return nil
}

func (r *SetRepository) GetByWorkoutID(workoutID int64) ([]*model.Set, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var workoutSets []*model.Set
	for _, set := range r.sets {
		if set.WorkoutID == workoutID {
			workoutSets = append(workoutSets, set)
		}
	}

	return workoutSets, nil
}

func (r *SetRepository) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.sets[id]
	if !exists {
		return fmt.Errorf("set not found: %w", repository.ErrNotFound)
	}

	delete(r.sets, id)
	return nil
}

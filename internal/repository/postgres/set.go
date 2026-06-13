package postgres

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SetRepository struct {
	pool *pgxpool.Pool
}

func NewSetRepository(pool *pgxpool.Pool) *SetRepository {
	return &SetRepository{
		pool: pool,
	}
}

func (r *SetRepository) Create(set *model.Set) error {
	ctx, cancel := newQueryContext()
	defer cancel()

	err := r.pool.QueryRow(ctx, `
		INSERT INTO sets (exercise_id, workout_id, reps, weight, rpe)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, set.ExerciseID, set.WorkoutID, set.Reps, set.Weight, set.RPE).Scan(
		&set.ID,
		&set.CreatedAt,
	)
	if err != nil {
		return mapConstraintError(err, "set parent not found")
	}

	return nil
}

func (r *SetRepository) GetByWorkoutID(workoutID int64) ([]*model.Set, error) {
	ctx, cancel := newQueryContext()
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT id, exercise_id, workout_id, created_at, reps, weight, rpe
		FROM sets
		WHERE workout_id = $1
		ORDER BY id
	`, workoutID)
	if err != nil {
		return nil, fmt.Errorf("get workout sets: %w", err)
	}
	defer rows.Close()

	sets := make([]*model.Set, 0)
	for rows.Next() {
		set := &model.Set{}
		if err := rows.Scan(
			&set.ID,
			&set.ExerciseID,
			&set.WorkoutID,
			&set.CreatedAt,
			&set.Reps,
			&set.Weight,
			&set.RPE,
		); err != nil {
			return nil, fmt.Errorf("scan set: %w", err)
		}
		sets = append(sets, set)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sets: %w", err)
	}

	return sets, nil
}

func (r *SetRepository) Delete(id int64) error {
	ctx, cancel := newQueryContext()
	defer cancel()

	result, err := r.pool.Exec(ctx, `
		DELETE FROM sets
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete set: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("set not found: %w", repository.ErrNotFound)
	}

	return nil
}

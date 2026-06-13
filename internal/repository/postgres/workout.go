package postgres

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkoutRepository struct {
	pool *pgxpool.Pool
}

func NewWorkoutRepository(pool *pgxpool.Pool) *WorkoutRepository {
	return &WorkoutRepository{
		pool: pool,
	}
}

func (r *WorkoutRepository) Create(workout *model.Workout) error {
	ctx, cancel := newQueryContext()
	defer cancel()

	err := r.pool.QueryRow(ctx, `
		INSERT INTO workouts (user_id, description)
		VALUES ($1, $2)
		RETURNING id, created_at
	`, workout.UserID, workout.Description).Scan(
		&workout.ID,
		&workout.CreatedAt,
	)
	if err != nil {
		return mapConstraintError(err, "workout parent not found")
	}

	return nil
}

func (r *WorkoutRepository) GetByID(id int64) (*model.Workout, error) {
	ctx, cancel := newQueryContext()
	defer cancel()

	workout := &model.Workout{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, description, created_at
		FROM workouts
		WHERE id = $1
	`, id).Scan(
		&workout.ID,
		&workout.UserID,
		&workout.Description,
		&workout.CreatedAt,
	)
	if err != nil {
		return nil, mapNotFound(err, "workout not found")
	}

	return workout, nil
}

func (r *WorkoutRepository) GetByUserID(userID int64) ([]*model.Workout, error) {
	ctx, cancel := newQueryContext()
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, description, created_at
		FROM workouts
		WHERE user_id = $1
		ORDER BY id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get user workouts: %w", err)
	}
	defer rows.Close()

	workouts := make([]*model.Workout, 0)
	for rows.Next() {
		workout := &model.Workout{}
		if err := rows.Scan(
			&workout.ID,
			&workout.UserID,
			&workout.Description,
			&workout.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan workout: %w", err)
		}
		workouts = append(workouts, workout)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workouts: %w", err)
	}

	return workouts, nil
}

func (r *WorkoutRepository) Delete(id int64) error {
	ctx, cancel := newQueryContext()
	defer cancel()

	result, err := r.pool.Exec(ctx, `
		DELETE FROM workouts
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete workout: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("workout not found: %w", repository.ErrNotFound)
	}

	return nil
}

package postgres

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ExerciseRepository struct {
	pool *pgxpool.Pool
}

func NewExerciseRepository(pool *pgxpool.Pool) *ExerciseRepository {
	return &ExerciseRepository{
		pool: pool,
	}
}

func (r *ExerciseRepository) Create(exercise *model.Exercise) error {
	ctx, cancel := newQueryContext()
	defer cancel()

	err := r.pool.QueryRow(ctx, `
		INSERT INTO exercises (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`, exercise.Name).Scan(
		&exercise.ID,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	if err != nil {
		return mapConstraintError(err, "exercise already exists")
	}

	return nil
}

func (r *ExerciseRepository) GetByID(id int64) (*model.Exercise, error) {
	ctx, cancel := newQueryContext()
	defer cancel()

	exercise := &model.Exercise{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, created_at, updated_at
		FROM exercises
		WHERE id = $1
	`, id).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	if err != nil {
		return nil, mapNotFound(err, "exercise not found")
	}

	return exercise, nil
}

func (r *ExerciseRepository) GetAll() ([]*model.Exercise, error) {
	ctx, cancel := newQueryContext()
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, created_at, updated_at
		FROM exercises
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("get exercises: %w", err)
	}
	defer rows.Close()

	exercises := make([]*model.Exercise, 0)
	for rows.Next() {
		exercise := &model.Exercise{}
		if err := rows.Scan(
			&exercise.ID,
			&exercise.Name,
			&exercise.CreatedAt,
			&exercise.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}
		exercises = append(exercises, exercise)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exercises: %w", err)
	}

	return exercises, nil
}

func (r *ExerciseRepository) Delete(id int64) error {
	ctx, cancel := newQueryContext()
	defer cancel()

	result, err := r.pool.Exec(ctx, `
		DELETE FROM exercises
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete exercise: %w", mapResourceInUse(err, "exercise is used"))
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("exercise not found: %w", repository.ErrNotFound)
	}

	return nil
}

package postgres

import (
	"Fitness-tracker/internal/model"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) Create(user *model.User) error {
	ctx, cancel := newQueryContext()
	defer cancel()

	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (telegram_id, username, weight, height, age)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, user.TelegramID, user.Username, user.Weight, user.Height, user.Age).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return mapConstraintError(err, "user already exists")
	}

	return nil
}

func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	ctx, cancel := newQueryContext()
	defer cancel()

	user := &model.User{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, telegram_id, username, created_at, updated_at, weight, height, age
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Weight,
		&user.Height,
		&user.Age,
	)
	if err != nil {
		return nil, mapNotFound(err, "user not found")
	}

	return user, nil
}

func (r *UserRepository) Update(user *model.User) error {
	ctx, cancel := newQueryContext()
	defer cancel()

	err := r.pool.QueryRow(ctx, `
		UPDATE users
		SET weight = $2, height = $3, age = $4, updated_at = now()
		WHERE id = $1
		RETURNING updated_at
	`, user.ID, user.Weight, user.Height, user.Age).Scan(&user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update user: %w", mapNotFound(err, "user not found"))
	}

	return nil
}

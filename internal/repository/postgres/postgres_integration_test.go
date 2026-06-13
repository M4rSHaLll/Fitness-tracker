package postgres

import (
	"Fitness-tracker/internal/config"
	"Fitness-tracker/internal/database"
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestPostgresRepositoriesFlow(t *testing.T) {
	databaseURL := os.Getenv("POSTGRES_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("POSTGRES_TEST_DATABASE_URL is not set")
	}

	pool, err := database.ConnectPostgres(config.DatabaseConfig{URL: databaseURL})
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	userRepo := NewUserRepository(pool)
	exerciseRepo := NewExerciseRepository(pool)
	workoutRepo := NewWorkoutRepository(pool)
	setRepo := NewSetRepository(pool)

	suffix := time.Now().UnixNano()
	user := model.NewUser(suffix, "pg_test_user")
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	defer pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID)

	duplicateUser := model.NewUser(suffix, "pg_test_user_duplicate")
	if err := userRepo.Create(duplicateUser); !errors.Is(err, repository.ErrAlreadyExists) {
		t.Fatalf("expected duplicate user ErrAlreadyExists, got %v", err)
	}

	exercise := model.NewExercise("pg_test_exercise_" + strconv.FormatInt(suffix, 10))
	if err := exerciseRepo.Create(exercise); err != nil {
		t.Fatalf("create exercise: %v", err)
	}

	workout := model.NewWorkout(user.ID, "Postgres flow")
	if err := workoutRepo.Create(workout); err != nil {
		t.Fatalf("create workout: %v", err)
	}

	set := model.NewSet(exercise.ID, workout.ID, 5, 100, 8.5)
	if err := setRepo.Create(set); err != nil {
		t.Fatalf("create set: %v", err)
	}

	if err := exerciseRepo.Delete(exercise.ID); !errors.Is(err, repository.ErrResourceInUse) {
		t.Fatalf("expected ErrResourceInUse, got %v", err)
	}

	if err := workoutRepo.Delete(workout.ID); err != nil {
		t.Fatalf("delete workout: %v", err)
	}
	if err := exerciseRepo.Delete(exercise.ID); err != nil {
		t.Fatalf("delete exercise after workout cascade: %v", err)
	}
}

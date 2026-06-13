package main

import (
	"Fitness-tracker/internal/config"
	"Fitness-tracker/internal/database"
	"Fitness-tracker/internal/handler"
	"Fitness-tracker/internal/repository"
	"Fitness-tracker/internal/repository/memory"
	postgresrepo "Fitness-tracker/internal/repository/postgres"
	"Fitness-tracker/internal/router"
	"Fitness-tracker/internal/service"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	userRepo, workoutRepo, setRepo, exerciseRepo, closeStorage, err := buildRepositories(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer closeStorage()

	userService := service.NewUserService(userRepo)
	workoutService := service.NewWorkoutService(workoutRepo, userRepo)
	exerciseService := service.NewExerciseService(exerciseRepo)
	setService := service.NewSetService(setRepo, workoutRepo, exerciseRepo)
	statsService := service.NewStatsService(setRepo, workoutRepo, userRepo)

	h := handler.NewHandler(
		userService,
		workoutService,
		setService,
		statsService,
		exerciseService)

	r := router.NewRouter(h, cfg.Storage.Driver)

	addr := cfg.App.Address()
	log.Printf("app env: %s", cfg.App.Env)
	log.Printf("storage driver: %s", cfg.Storage.Driver)
	log.Printf("server started on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func buildRepositories(cfg *config.Config) (
	repository.UserRepository,
	repository.WorkoutRepository,
	repository.SetRepository,
	repository.ExerciseRepository,
	func(),
	error,
) {
	switch cfg.Storage.Driver {
	case "memory":
		userRepo := memory.NewUserRepository()
		workoutRepo := memory.NewWorkoutRepository()
		setRepo := memory.NewSetRepository()
		exerciseRepo := memory.NewExerciseRepository()
		return userRepo, workoutRepo, setRepo, exerciseRepo, func() {}, nil
	case "postgres":
		pool, err := database.ConnectPostgres(cfg.Database)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		userRepo := postgresrepo.NewUserRepository(pool)
		workoutRepo := postgresrepo.NewWorkoutRepository(pool)
		setRepo := postgresrepo.NewSetRepository(pool)
		exerciseRepo := postgresrepo.NewExerciseRepository(pool)
		return userRepo, workoutRepo, setRepo, exerciseRepo, pool.Close, nil
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported storage driver: %s", cfg.Storage.Driver)
	}
}

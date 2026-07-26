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
	telegrambot "Fitness-tracker/internal/telegram"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	httpReadHeaderTimeout = 5 * time.Second
	httpIdleTimeout       = 60 * time.Second
	shutdownTimeout       = 10 * time.Second
	telegramRetryInterval = 15 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Printf("application stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	userRepo, workoutRepo, setRepo, exerciseRepo, closeStorage, err := buildRepositories(cfg)
	if err != nil {
		return err
	}
	defer closeStorage()

	userService := service.NewUserService(userRepo)
	workoutService := service.NewWorkoutService(workoutRepo, userRepo)
	exerciseService := service.NewExerciseService(exerciseRepo)
	setService := service.NewSetService(setRepo, workoutRepo, exerciseRepo)
	statsService := service.NewStatsService(setRepo, workoutRepo, userRepo)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var botWG sync.WaitGroup
	if cfg.Telegram.BotToken != "" {
		if cfg.Telegram.Mode != "polling" {
			return fmt.Errorf("unsupported telegram mode: %s", cfg.Telegram.Mode)
		}

		botServices := telegrambot.Services{
			Users:     userService,
			Workouts:  workoutService,
			Sets:      setService,
			Stats:     statsService,
			Exercises: exerciseService,
		}

		botWG.Add(1)
		go func() {
			defer botWG.Done()
			runTelegramBot(ctx, cfg.Telegram.BotToken, botServices, log.Default())
		}()
	} else {
		log.Print("telegram bot disabled: TELEGRAM_BOT_TOKEN is not set")
	}

	h := handler.NewHandler(
		userService,
		workoutService,
		setService,
		statsService,
		exerciseService)

	r := router.NewRouter(h, cfg.Storage.Driver, cfg.Security.InternalAPIToken, cfg.Security.APIToken)

	addr := cfg.App.Address()
	server := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: httpReadHeaderTimeout,
		IdleTimeout:       httpIdleTimeout,
	}

	log.Printf("app env: %s", cfg.App.Env)
	log.Printf("storage driver: %s", cfg.Storage.Driver)
	log.Printf("server started on %s", addr)

	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("serve HTTP: %w", err)
		}
	}()

	var runErr error
	select {
	case <-ctx.Done():
		log.Print("shutdown signal received")
	case runErr = <-serverErr:
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil && runErr == nil {
		runErr = fmt.Errorf("shutdown HTTP server: %w", err)
	}
	botWG.Wait()

	log.Print("application stopped")
	return runErr
}

func runTelegramBot(ctx context.Context, token string, services telegrambot.Services, logger *log.Logger) {
	for {
		if ctx.Err() != nil {
			return
		}

		bot, err := telegrambot.NewBot(token, services, logger)
		if err == nil {
			bot.Start(ctx)
			if ctx.Err() != nil {
				return
			}
			logger.Printf("telegram polling stopped; retrying in %s", telegramRetryInterval)
		} else {
			logger.Printf("telegram bot unavailable: %v; retrying in %s", err, telegramRetryInterval)
		}

		timer := time.NewTimer(telegramRetryInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-timer.C:
		}
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

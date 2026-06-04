package main

import (
	"Fitness-tracker/internal/handler"
	"Fitness-tracker/internal/repository/memory"
	"Fitness-tracker/internal/router"
	"Fitness-tracker/internal/service"
	"log"
	"net/http"
)

func main() {

	userRepo := memory.NewUserRepository()
	workoutRepo := memory.NewWorkoutRepository()
	setRepo := memory.NewSetRepository()

	userService := service.NewUserService(userRepo)

	workoutService := service.NewWorkoutService(workoutRepo, userRepo)

	setService := service.NewSetService(setRepo, workoutRepo)

	statsService := service.NewStatsService(setRepo, workoutRepo, userRepo)

	h := handler.NewHandler(
		userService,
		workoutService,
		setService,
		statsService)

	r := router.NewRouter(h)

	log.Println("server started on :8080")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

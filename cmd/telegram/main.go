package main

import (
	"Fitness-tracker/internal/handler"
	"Fitness-tracker/internal/repository/memory"
	"Fitness-tracker/internal/router"
	"Fitness-tracker/internal/service"
	"fmt"
	"net/http"
)

func main() {

	userRepo := memory.NewUserRepository()

	userService := service.NewUserService(
		userRepo,
	)

	h := handler.NewHandler(
		userService,
		nil,
		nil,
	)

	r := router.NewRouter(h)

	fmt.Println("server started on :8080")

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}

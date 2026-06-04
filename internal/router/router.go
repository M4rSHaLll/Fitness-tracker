package router

import (
	"Fitness-tracker/internal/handler"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *handler.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/users", h.CreateUser)
	r.Get("/users/{id}", h.GetUser)
	r.Patch("/users/{id}", h.UpdateUser)

	r.Post("/workouts", h.CreateWorkout)
	r.Get("/users/{id}/workouts", h.GetWorkouts)
	r.Delete("/workouts/{id}", h.DeleteWorkout)

	r.Post("/sets", h.CreateSet)
	r.Get("/workouts/{id}/sets", h.GetSets)
	r.Delete("/sets/{id}", h.DeleteSet)

	r.Get("/users/{id}/stats", h.GetStats)

	return r

}

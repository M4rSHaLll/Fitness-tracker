package router

import (
	"Fitness-tracker/internal/dto/response"
	"Fitness-tracker/internal/handler"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const internalTokenHeader = "X-Internal-Token"

func NewRouter(h *handler.Handler, storageDriver, internalToken, apiToken string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", handler.NewHealthHandler(storageDriver))

	r.Route("/telegram", func(r chi.Router) {
		r.Use(internalAuth(internalToken))
		r.Post("/users", h.GetOrCreateTelegramUser)
		r.Post("/users/{telegram_id}/workouts", h.CreateTelegramWorkout)
		r.Get("/users/{telegram_id}/workouts", h.GetTelegramWorkouts)
		r.Post("/users/{telegram_id}/sets", h.CreateTelegramSet)
		r.Get("/users/{telegram_id}/workouts/{workout_id}/sets", h.GetTelegramSets)
		r.Get("/users/{telegram_id}/stats", h.GetTelegramStats)
	})

	r.Group(func(r chi.Router) {
		r.Use(bearerAuth(apiToken))

		r.Post("/users", h.CreateUser)
		r.Get("/users/{id}", h.GetUser)
		r.Patch("/users/{id}", h.UpdateUser)

		r.Post("/exercises", h.CreateExercise)
		r.Get("/exercises", h.GetExercises)
		r.Get("/exercises/{id}", h.GetExercise)
		r.Delete("/exercises/{id}", h.DeleteExercise)

		r.Post("/workouts", h.CreateWorkout)
		r.Get("/users/{id}/workouts", h.GetWorkouts)
		r.Delete("/workouts/{id}", h.DeleteWorkout)

		r.Post("/sets", h.CreateSet)
		r.Get("/workouts/{id}/sets", h.GetSets)
		r.Delete("/sets/{id}", h.DeleteSet)

		r.Get("/users/{id}/stats", h.GetStats)
	})
	return r

}

func internalAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				writeRouterError(w, http.StatusServiceUnavailable, "Internal API token is not configured", "internal_auth_not_configured")
				return
			}
			if !tokenMatches(r.Header.Get(internalTokenHeader), token) {
				writeRouterError(w, http.StatusUnauthorized, "Invalid internal token", "invalid_internal_token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bearerAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				writeRouterError(w, http.StatusServiceUnavailable, "API token is not configured", "api_auth_not_configured")
				return
			}

			authorization := strings.TrimSpace(r.Header.Get("Authorization"))
			const prefix = "Bearer "
			if !strings.HasPrefix(authorization, prefix) ||
				!tokenMatches(strings.TrimSpace(strings.TrimPrefix(authorization, prefix)), token) {
				writeRouterError(w, http.StatusUnauthorized, "Invalid API token", "invalid_api_token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func tokenMatches(provided, expected string) bool {
	if len(provided) != len(expected) || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func writeRouterError(w http.ResponseWriter, status int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response.Error{
		Error: message,
		Code:  code,
	})
}

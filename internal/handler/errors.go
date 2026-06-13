package handler

import (
	"Fitness-tracker/internal/dto/response"
	"Fitness-tracker/internal/service"
	"errors"
	"net/http"
)

func writeBadRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, response.Error{Error: message, Code: "invalid_request"})
}

func writeServiceError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, response.Error{Error: "Invalid request data", Code: "invalid_request"})
	case errors.Is(err, service.ErrAlreadyExists):
		writeJSON(w, http.StatusConflict, response.Error{Error: "Resource already exists", Code: "already_exists"})
	case errors.Is(err, service.ErrResourceInUse):
		writeJSON(w, http.StatusConflict, response.Error{Error: "Resource is in use", Code: "resource_in_use"})
	case errors.Is(err, service.ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, response.Error{Error: "User not found", Code: "user_not_found"})
	case errors.Is(err, service.ErrWorkoutNotFound):
		writeJSON(w, http.StatusNotFound, response.Error{Error: "Workout not found", Code: "workout_not_found"})
	case errors.Is(err, service.ErrSetNotFound):
		writeJSON(w, http.StatusNotFound, response.Error{Error: "Set not found", Code: "set_not_found"})
	case errors.Is(err, service.ErrExerciseNotFound):
		writeJSON(w, http.StatusNotFound, response.Error{Error: "Exercise not found", Code: "exercise_not_found"})
	default:
		writeJSON(w, http.StatusInternalServerError, response.Error{Error: fallback, Code: "internal_error"})
	}
}

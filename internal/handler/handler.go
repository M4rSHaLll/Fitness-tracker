package handler

import (
	"Fitness-tracker/internal/dto/request"
	"Fitness-tracker/internal/dto/response"
	"Fitness-tracker/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	userService    *service.UserService
	workoutService *service.WorkoutService
	setService     *service.SetService
}

func NewHandler(userService *service.UserService, workoutService *service.WorkoutService, setService *service.SetService) *Handler {
	return &Handler{
		userService:    userService,
		workoutService: workoutService,
		setService:     setService,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req request.CreateUser

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid request body",
		})
		return
	}

	user, err := h.userService.CreateUser(req.TelegramID, req.Username)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response.Error{
			Error: "Failed to create user",
		})
		return
	}

	resp := response.User{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		Username:   user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	userID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid user ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, response.Error{
			Error: "User not found",
		})
		return
	}

	resp := response.User{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		Username:   user.Username,
		Weight:     user.Weight,
		Height:     user.Height,
		Age:        user.Age,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	idParam := chi.URLParam(r, "id")

	userID, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid user ID",
		})
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req request.UpdateUser

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid request body",
		})
		return
	}

	err = h.userService.UpdateProfile(userID, req.Weight, req.Height, req.Age)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response.Error{
			Error: "Failed to update user",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "user updated successfully",
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}
}

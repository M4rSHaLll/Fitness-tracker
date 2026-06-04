package handler

import (
	"Fitness-tracker/internal/dto/request"
	"Fitness-tracker/internal/dto/response"
	"Fitness-tracker/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	userService    *service.UserService
	workoutService *service.WorkoutService
	setService     *service.SetService
	statsService   *service.StatsService
}

func NewHandler(userService *service.UserService, workoutService *service.WorkoutService, setService *service.SetService, statsService *service.StatsService) *Handler {
	return &Handler{
		userService:    userService,
		workoutService: workoutService,
		setService:     setService,
		statsService:   statsService,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req request.CreateUser

	if err := decoderJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid request body",
		})
		return
	}

	user, err := h.userService.CreateUser(req.TelegramID, req.Username)

	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			writeJSON(w, http.StatusBadRequest, response.Error{
				Error: "Invalid user data",
			})
			return
		}
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

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r, "id")

	if err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid user ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if !errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusInternalServerError, response.Error{
				Error: "Failed to get user",
			})
			return
		}
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

	userID, err := parseID(r, "id")

	if err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid user ID",
		})
		return
	}

	var req request.UpdateUser

	if err := decoderJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid request body",
		})
		return
	}

	err = h.userService.UpdateProfile(userID, req.Weight, req.Height, req.Age)

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, response.Error{
				Error: "User not found",
			})
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			writeJSON(w, http.StatusBadRequest, response.Error{
				Error: "Invalid user data",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, response.Error{
			Error: "Failed to update user",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "user updated successfully",
	})
}

func (h *Handler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req request.CreateWorkout

	if err := decoderJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid request body",
		})
		return
	}

	workout, err := h.workoutService.CreateWorkout(req.UserID, req.Description)

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, response.Error{
				Error: "User not found",
			})
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			writeJSON(w, http.StatusBadRequest, response.Error{
				Error: "Invalid workout data",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, response.Error{
			Error: "Failed to create workout",
		})
		return
	}

	resp := response.Workout{
		ID:          workout.ID,
		UserID:      workout.UserID,
		Description: workout.Description,
		CreatedAt:   workout.CreatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetWorkouts(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r, "id")

	if err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid user ID",
		})
		return
	}

	workouts, err := h.workoutService.GetUserWorkouts(userID)

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, response.Error{
				Error: "User not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, response.Error{
			Error: "Failed to get workouts",
		})
		return
	}

	resp := make([]response.Workout, 0, len(workouts))

	for _, workout := range workouts {
		resp = append(resp, response.Workout{
			ID:          workout.ID,
			UserID:      workout.UserID,
			Description: workout.Description,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	workoutID, err := parseID(r, "id")

	if err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid workout ID",
		})
		return
	}

	err = h.workoutService.DeleteWorkout(workoutID)

	if err != nil {
		if errors.Is(err, service.ErrWorkoutNotFound) {
			writeJSON(w, http.StatusNotFound, response.Error{
				Error: "Workout not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, response.Error{
			Error: "Failed to delete workout",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "workout deleted successfully",
	})
}

func (h *Handler) CreateSet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req request.CreateSet

	if err := decoderJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid request body",
		})
		return
	}

	set, err := h.setService.CreateSet(req.ExerciseID, req.WorkoutID, req.Reps, req.Weight, req.RPE)

	if err != nil {
		if errors.Is(err, service.ErrWorkoutNotFound) {
			writeJSON(w, http.StatusNotFound, response.Error{
				Error: "Workout not found",
			})
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			writeJSON(w, http.StatusBadRequest, response.Error{
				Error: "Invalid set data",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, response.Error{
			Error: "Failed to create set",
		})
		return
	}

	resp := response.Set{
		ID:         set.ID,
		WorkoutID:  set.WorkoutID,
		ExerciseID: set.ExerciseID,
		Reps:       set.Reps,
		Weight:     set.Weight,
		RPE:        set.RPE,
		CreatedAt:  set.CreatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetSets(w http.ResponseWriter, r *http.Request) {
	workoutID, err := parseID(r, "id")

	if err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid workout ID",
		})
		return
	}

	sets, err := h.setService.GetWorkoutSets(workoutID)

	if err != nil {
		if errors.Is(err, service.ErrWorkoutNotFound) {
			writeJSON(w, http.StatusNotFound, response.Error{
				Error: "Workout not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, response.Error{
			Error: "Failed to get sets",
		})
		return
	}

	resp := make([]response.Set, 0, len(sets))

	for _, set := range sets {
		resp = append(resp, response.Set{
			ID:         set.ID,
			WorkoutID:  set.WorkoutID,
			ExerciseID: set.ExerciseID,
			Reps:       set.Reps,
			Weight:     set.Weight,
			RPE:        set.RPE,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeleteSet(w http.ResponseWriter, r *http.Request) {
	setID, err := parseID(r, "id")

	if err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid set ID",
		})
		return
	}

	err = h.setService.DeleteSet(setID)

	if err != nil {
		if errors.Is(err, service.ErrSetNotFound) {
			writeJSON(w, http.StatusNotFound, response.Error{
				Error: "Set not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, response.Error{
			Error: "Failed to delete set",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "set deleted successfully",
	})

}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r, "id")

	if err != nil {
		writeJSON(w, http.StatusBadRequest, response.Error{
			Error: "Invalid user ID",
		})
		return
	}

	stats, err := h.statsService.GetUserStats(userID)

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, response.Error{
				Error: "User not found",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, response.Error{
			Error: "Failed to get stats",
		})
		return
	}

	resp := response.UserStats{
		TotalWorkouts: stats.TotalWorkouts,
		TotalVolume:   stats.TotalVolume,
		AverageRPE:    stats.AverageRPE,
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	body, err := json.Marshal(data)
	if err != nil {

		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)

		return
	}

	w.WriteHeader(statusCode)

	w.Write(body)
}

func decoderJSON(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(dst)
}

func parseID(r *http.Request, param string) (int64, error) {
	idParam := chi.URLParam(r, param)
	return strconv.ParseInt(idParam, 10, 64)
}

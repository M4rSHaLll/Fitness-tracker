package handler

import (
	"Fitness-tracker/internal/dto/request"
	"Fitness-tracker/internal/dto/response"
	"Fitness-tracker/internal/model"
	"net/http"
)

func (h *Handler) GetOrCreateTelegramUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req request.TelegramUser
	if err := decoderJSON(r, &req); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	user, err := h.userService.GetOrCreateTelegramUser(req.TelegramID, req.Username)
	if err != nil {
		writeServiceError(w, err, "Failed to get or create telegram user")
		return
	}

	writeJSON(w, http.StatusOK, userResponse(user))
}

func (h *Handler) CreateTelegramWorkout(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	telegramID, err := parseID(r, "telegram_id")
	if err != nil {
		writeBadRequest(w, "Invalid telegram ID")
		return
	}

	var req request.CreateTelegramWorkout
	if err := decoderJSON(r, &req); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	user, err := h.userService.GetOrCreateTelegramUser(telegramID, req.Username)
	if err != nil {
		writeServiceError(w, err, "Failed to get or create telegram user")
		return
	}

	workout, err := h.workoutService.CreateWorkout(user.ID, req.Description)
	if err != nil {
		writeServiceError(w, err, "Failed to create workout")
		return
	}

	writeJSON(w, http.StatusCreated, workoutResponse(workout))
}

func (h *Handler) GetTelegramWorkouts(w http.ResponseWriter, r *http.Request) {
	user, ok := h.telegramUserFromPath(w, r)
	if !ok {
		return
	}

	workouts, err := h.workoutService.GetUserWorkouts(user.ID)
	if err != nil {
		writeServiceError(w, err, "Failed to get workouts")
		return
	}

	resp := make([]response.Workout, 0, len(workouts))
	for _, workout := range workouts {
		resp = append(resp, workoutResponse(workout))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) CreateTelegramSet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	telegramID, err := parseID(r, "telegram_id")
	if err != nil {
		writeBadRequest(w, "Invalid telegram ID")
		return
	}

	var req request.CreateTelegramSet
	if err := decoderJSON(r, &req); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	user, err := h.userService.GetOrCreateTelegramUser(telegramID, req.Username)
	if err != nil {
		writeServiceError(w, err, "Failed to get or create telegram user")
		return
	}

	if ok := h.ensureWorkoutOwner(w, user.ID, req.WorkoutID); !ok {
		return
	}

	set, err := h.setService.CreateSet(req.ExerciseID, req.WorkoutID, req.Reps, req.Weight, req.RPE)
	if err != nil {
		writeServiceError(w, err, "Failed to create set")
		return
	}

	writeJSON(w, http.StatusCreated, setResponse(set))
}

func (h *Handler) GetTelegramSets(w http.ResponseWriter, r *http.Request) {
	user, ok := h.telegramUserFromPath(w, r)
	if !ok {
		return
	}

	workoutID, err := parseID(r, "workout_id")
	if err != nil {
		writeBadRequest(w, "Invalid workout ID")
		return
	}

	if ok := h.ensureWorkoutOwner(w, user.ID, workoutID); !ok {
		return
	}

	sets, err := h.setService.GetWorkoutSets(workoutID)
	if err != nil {
		writeServiceError(w, err, "Failed to get sets")
		return
	}

	resp := make([]response.Set, 0, len(sets))
	for _, set := range sets {
		resp = append(resp, setResponse(set))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetTelegramStats(w http.ResponseWriter, r *http.Request) {
	user, ok := h.telegramUserFromPath(w, r)
	if !ok {
		return
	}

	stats, err := h.statsService.GetUserStats(user.ID)
	if err != nil {
		writeServiceError(w, err, "Failed to get stats")
		return
	}

	writeJSON(w, http.StatusOK, response.UserStats{
		TotalWorkouts: stats.TotalWorkouts,
		TotalVolume:   stats.TotalVolume,
		AverageRPE:    stats.AverageRPE,
	})
}

func (h *Handler) telegramUserFromPath(w http.ResponseWriter, r *http.Request) (*model.User, bool) {
	telegramID, err := parseID(r, "telegram_id")
	if err != nil {
		writeBadRequest(w, "Invalid telegram ID")
		return nil, false
	}

	user, err := h.userService.GetUserByTelegramID(telegramID)
	if err != nil {
		writeServiceError(w, err, "Failed to get telegram user")
		return nil, false
	}

	return user, true
}

func (h *Handler) ensureWorkoutOwner(w http.ResponseWriter, userID, workoutID int64) bool {
	workout, err := h.workoutService.GetWorkoutByID(workoutID)
	if err != nil {
		writeServiceError(w, err, "Failed to get workout")
		return false
	}
	if workout.UserID != userID {
		writeForbidden(w, "Workout belongs to another user")
		return false
	}
	return true
}

func userResponse(user *model.User) response.User {
	return response.User{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		Username:   user.Username,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
		Weight:     user.Weight,
		Height:     user.Height,
		Age:        user.Age,
	}
}

func workoutResponse(workout *model.Workout) response.Workout {
	return response.Workout{
		ID:          workout.ID,
		UserID:      workout.UserID,
		Description: workout.Description,
		CreatedAt:   workout.CreatedAt,
	}
}

func setResponse(set *model.Set) response.Set {
	return response.Set{
		ID:         set.ID,
		WorkoutID:  set.WorkoutID,
		ExerciseID: set.ExerciseID,
		Reps:       set.Reps,
		Weight:     set.Weight,
		RPE:        set.RPE,
		CreatedAt:  set.CreatedAt,
	}
}

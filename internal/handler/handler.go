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
	userService     *service.UserService
	workoutService  *service.WorkoutService
	setService      *service.SetService
	statsService    *service.StatsService
	exerciseService *service.ExerciseService
}

func NewHandler(userService *service.UserService, workoutService *service.WorkoutService, setService *service.SetService, statsService *service.StatsService, exerciseService *service.ExerciseService) *Handler {
	return &Handler{
		userService:     userService,
		workoutService:  workoutService,
		setService:      setService,
		statsService:    statsService,
		exerciseService: exerciseService,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req request.CreateUser

	if err := decoderJSON(r, &req); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	user, err := h.userService.CreateUser(req.TelegramID, req.Username)

	if err != nil {
		writeServiceError(w, err, "Failed to create user")
		return
	}

	resp := response.User{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		Username:   user.Username,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r, "id")

	if err != nil {
		writeBadRequest(w, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		writeServiceError(w, err, "Failed to get user")
		return
	}

	resp := response.User{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		Username:   user.Username,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
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
		writeBadRequest(w, "Invalid user ID")
		return
	}

	var req request.UpdateUser

	if err := decoderJSON(r, &req); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	err = h.userService.UpdateProfile(userID, req.Weight, req.Height, req.Age)

	if err != nil {
		writeServiceError(w, err, "Failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "user updated successfully",
	})
}

func (h *Handler) CreateExercise(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req request.CreateExercise

	if err := decoderJSON(r, &req); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	exercise, err := h.exerciseService.CreateExercise(req.Name)
	if err != nil {
		writeServiceError(w, err, "Failed to create exercise")
		return
	}

	resp := response.Exercise{
		ID:        exercise.ID,
		Name:      exercise.Name,
		CreatedAt: exercise.CreatedAt,
		UpdatedAt: exercise.UpdatedAt,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetExercise(w http.ResponseWriter, r *http.Request) {
	exerciseID, err := parseID(r, "id")

	if err != nil {
		writeBadRequest(w, "Invalid exercise ID")
		return
	}

	exercise, err := h.exerciseService.GetExerciseByID(exerciseID)
	if err != nil {
		writeServiceError(w, err, "Failed to get exercise")
		return
	}

	resp := response.Exercise{
		ID:        exercise.ID,
		Name:      exercise.Name,
		CreatedAt: exercise.CreatedAt,
		UpdatedAt: exercise.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetExercises(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.exerciseService.GetExercises()
	if err != nil {
		writeServiceError(w, err, "Failed to get exercises")
		return
	}

	resp := make([]response.Exercise, 0, len(exercises))

	for _, exercise := range exercises {
		resp = append(resp, response.Exercise{
			ID:        exercise.ID,
			Name:      exercise.Name,
			CreatedAt: exercise.CreatedAt,
			UpdatedAt: exercise.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) DeleteExercise(w http.ResponseWriter, r *http.Request) {
	exerciseID, err := parseID(r, "id")

	if err != nil {
		writeBadRequest(w, "Invalid exercise ID")
		return
	}

	if err := h.exerciseService.DeleteExercise(exerciseID); err != nil {
		writeServiceError(w, err, "Failed to delete exercise")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "exercise deleted successfully",
	})
}

func (h *Handler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req request.CreateWorkout

	if err := decoderJSON(r, &req); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	workout, err := h.workoutService.CreateWorkout(req.UserID, req.Description)

	if err != nil {
		writeServiceError(w, err, "Failed to create workout")
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
		writeBadRequest(w, "Invalid user ID")
		return
	}

	workouts, err := h.workoutService.GetUserWorkouts(userID)

	if err != nil {
		writeServiceError(w, err, "Failed to get workouts")
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
		writeBadRequest(w, "Invalid workout ID")
		return
	}

	err = h.workoutService.DeleteWorkout(workoutID)

	if err != nil {
		writeServiceError(w, err, "Failed to delete workout")
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
		writeBadRequest(w, "Invalid request body")
		return
	}

	set, err := h.setService.CreateSet(req.ExerciseID, req.WorkoutID, req.Reps, req.Weight, req.RPE)

	if err != nil {
		writeServiceError(w, err, "Failed to create set")
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
		writeBadRequest(w, "Invalid workout ID")
		return
	}

	sets, err := h.setService.GetWorkoutSets(workoutID)

	if err != nil {
		writeServiceError(w, err, "Failed to get sets")
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
		writeBadRequest(w, "Invalid set ID")
		return
	}

	err = h.setService.DeleteSet(setID)

	if err != nil {
		writeServiceError(w, err, "Failed to delete set")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "set deleted successfully",
	})

}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r, "id")

	if err != nil {
		writeBadRequest(w, "Invalid user ID")
		return
	}

	stats, err := h.statsService.GetUserStats(userID)

	if err != nil {
		writeServiceError(w, err, "Failed to get stats")
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

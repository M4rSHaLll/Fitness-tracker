package request

type TelegramUser struct {
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username"`
}

type CreateTelegramWorkout struct {
	Username    string `json:"username"`
	Description string `json:"description"`
}

type CreateTelegramSet struct {
	Username   string  `json:"username"`
	WorkoutID  int64   `json:"workout_id"`
	ExerciseID int64   `json:"exercise_id"`
	Weight     float64 `json:"weight"`
	Reps       int64   `json:"reps"`
	RPE        float64 `json:"rpe"`
}

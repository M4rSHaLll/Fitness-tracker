package request

type CreateSet struct {
	WorkoutID  int64 `json:"workout_id"`
	ExerciseID int64 `json:"exercise_id"`

	Weight float64 `json:"weight"`
	Reps   int64   `json:"reps"`
	RPE    float64 `json:"rpe"`
}

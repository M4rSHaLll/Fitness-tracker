package response

type UserStats struct {
	TotalWorkouts int64   `json:"total_workouts"`
	TotalVolume   float64 `json:"total_volume"`
	AverageRPE    float64 `json:"average_rpe"`
}

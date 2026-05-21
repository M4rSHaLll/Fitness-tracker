package request

type CreateWorkout struct {
	UserID      int64  `json:"user_id"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

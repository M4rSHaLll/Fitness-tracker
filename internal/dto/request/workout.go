package request

type CreateWorkout struct {
	UserID      int64  `json:"user_id"`
	Description string `json:"description"`
}

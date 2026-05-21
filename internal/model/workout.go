package model

import (
	"time"
)

type Workout struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewWorkout(userID int64, description string) *Workout {
	return &Workout{
		UserID:      userID,
		Description: description,
		CreatedAt:   time.Now(),
	}
}

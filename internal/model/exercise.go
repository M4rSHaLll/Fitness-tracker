package model

import "time"

type Exercise struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewExercise(name string) *Exercise {
	now := time.Now()
	return &Exercise{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

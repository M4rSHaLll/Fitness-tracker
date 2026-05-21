package model

type Exercise struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewExercise(name string) *Exercise {
	return &Exercise{
		Name: name,
	}
}

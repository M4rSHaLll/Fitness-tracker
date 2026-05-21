package validation

import "errors"

const (
	maxExerciseWeight = 10000
	maxReps           = 1000
)

func ValidationExerciseWeight(weight float64) error {
	if weight <= 0 || weight > maxExerciseWeight {
		return errors.New("invalid set weight, weight must be between 1 and 10000 kg")
	}
	return nil
}

func ValidationReps(reps int64) error {
	if reps < 0 || reps > maxReps {
		return errors.New("invalid reps, reps must be between 0 and 1000")
	}
	return nil
}

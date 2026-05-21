package validation

import "errors"

const (
	maxWeight = 500
	maxHeight = 300
	maxAge    = 150
)

func ValidationWeight(weight float64) error {
	if weight <= 0 || weight > maxWeight {
		return errors.New("invalid weight, weight must be between 1 and 500 kg")
	}
	return nil
}

func ValidationHeight(height float64) error {
	if height <= 0 || height > maxHeight {
		return errors.New("invalid height, height must be between 1 and 300 cm")
	}
	return nil
}

func ValidationAge(age int64) error {
	if age <= 0 || age > maxAge {
		return errors.New("invalid age, age must be between 1 and 150")
	}
	return nil
}

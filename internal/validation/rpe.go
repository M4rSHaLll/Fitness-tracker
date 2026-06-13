package validation

import (
	"errors"
	"math"
)

func ValidationRPE(rpe float64) error {
	if rpe < 6 || rpe > 10 {
		return errors.New("invalid RPE, RPE must be between 6 and 10")
	}
	if math.Abs(rpe*2-math.Round(rpe*2)) > 1e-9 {
		return errors.New("invalid RPE, RPE step must be 0.5")
	}
	return nil
}

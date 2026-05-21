package validation

import "errors"

func ValidationRPE(rpe float64) error {
	if rpe < 6 || rpe > 10 {
		return errors.New("invalid RPE, RPE must be between 6 and 10")
	}
	return nil
}

package validation

import "testing"

func TestValidationRPEAcceptsHalfStepValues(t *testing.T) {
	values := []float64{6, 6.5, 8, 9.5, 10}

	for _, value := range values {
		if err := ValidationRPE(value); err != nil {
			t.Fatalf("expected %v to be valid, got %v", value, err)
		}
	}
}

func TestValidationRPERejectsInvalidStep(t *testing.T) {
	if err := ValidationRPE(7.25); err == nil {
		t.Fatal("expected invalid RPE step error")
	}
}

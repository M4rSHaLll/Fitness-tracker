package service

import "fmt"

func calculateMaxReps(weight float32, reps int32, RPE float32) (float32, int32, error) {
	iRPE := int32(RPE * 10)

	switch iRPE % 10 {
	case 5:
		return weight + 2.5, reps + (10 - iRPE/10), nil
	case 0:
		return weight, reps + (10 - iRPE/10), nil
	}

	return 0, 0, fmt.Errorf("incorrect RPE value")
}

func Calculate1RM(weight float32, reps int32, RPE float32) (float32, error) {
	corWeight, corReps, err := calculateMaxReps(weight, reps, RPE)
	if err != nil{
		return 0, err
	}

	repMax := corWeight * (1 + float32(corReps) / 30)

	return repMax, nil
}

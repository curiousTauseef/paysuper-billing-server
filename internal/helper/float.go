package helper

import (
	"fmt"
	"strconv"
)

func Round(v float64) float64 {
	roundedTxt := fmt.Sprintf("%0.2f", v)
	value, err := strconv.ParseFloat(roundedTxt, 64)
	if err != nil {
		return 0
	}

	return value
}

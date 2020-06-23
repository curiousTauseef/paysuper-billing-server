package helper

import (
	"go.uber.org/zap"
	"time"
)

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	zap.L().Info(
		"function execution time",
		zap.String("name", name),
		zap.Duration("time", elapsed),
	)
}

package testutils

import (
	"testing"
	"time"

	"github.com/GoGstickGo/terratest-helpers/core"
	"github.com/gruntwork-io/terratest/modules/logger"
)

type Logger interface {
	Log(t *testing.T, args ...interface{})
}

type RealLogger struct{}

func (RealLogger) Log(t *testing.T, args ...interface{}) {
	logger.Log(t, args...)
}

type Sleeper interface {
	Sleep(duration time.Duration)
}

type RealSleeper struct{}

func (RealSleeper) Sleep(duration time.Duration) {
	time.Sleep(duration)
}
func PauseTest(t *testing.T, config core.RunTime, logger Logger, sleeper Sleeper) {
	logger.Log(t, "Pause test for", config.Pause, "before starting destruction of the environment")
	sleeper.Sleep(config.Pause)
}

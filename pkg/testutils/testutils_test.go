package testutils_test

import (
	"testing"
	"time"

	"github.com/GoGstickGo/terratest-helpers/core"
	"github.com/GoGstickGo/terratest-helpers/pkg/testutils"
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Log(t *testing.T, args ...interface{}) {
	combinedArgs := append([]interface{}{t}, args...)
	m.Called(combinedArgs...)
}

type MockSleeper struct {
	mock.Mock
}

func (m *MockSleeper) Sleep(duration time.Duration) {
	m.Called(duration)
}
func TestMockPauseTest(t *testing.T) {
	t.Parallel()
	// Create mock logger
	mockLogger := new(MockLogger)

	// Create mock sleeper
	mockSleeper := new(MockSleeper)

	// Prepare config
	pauseDuration := 2 * time.Second
	config := core.RunTime{
		Pause: pauseDuration,
	}

	// Set up expectations for the logger
	mockLogger.On(
		"Log",
		t,
		"Pause test for",
		pauseDuration,
		"before starting destruction of the environment",
	).Return()

	// Set up expectations for the sleeper
	mockSleeper.On("Sleep", pauseDuration).Return()

	// Call the function under test
	testutils.PauseTest(t, config, mockLogger, mockSleeper)

	// Assertions
	mockLogger.AssertExpectations(t)
	mockSleeper.AssertExpectations(t)
}

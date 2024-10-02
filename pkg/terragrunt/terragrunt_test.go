package terragrunt

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/GoGstickGo/terratest-helpers/core"
	"github.com/GoGstickGo/terratest-helpers/pkg/parameters"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTerragruntExecutor struct {
	mock.Mock
}

func (m *MockTerragruntExecutor) TgApplyAllE(t *testing.T, options *terraform.Options) (string, error) {
	args := m.Called(t, options)
	return args.String(0), args.Error(1)
}

func (m *MockTerragruntExecutor) TgDestroyAllE(t *testing.T, options *terraform.Options) (string, error) {
	args := m.Called(t, options)
	return args.String(0), args.Error(1)
}

type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) RunCommand(cmdName string, args []string, dir string, envVars map[string]string) ([]byte, error) {
	argsCalled := m.Called(cmdName, args, dir, envVars)
	return argsCalled.Get(0).([]byte), argsCalled.Error(1)
}

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

func TestMockTgApply_Success(t *testing.T) {
	// Create a mock executors
	mockExecutor := new(MockTerragruntExecutor)

	cmdMockExecutor := new(MockCommandExecutor)

	// Set up expected behavior
	mockExecutor.On("TgApplyAllE", t, mock.AnythingOfType("*terraform.Options")).Return("Mocked output", nil)

	// Prepare other parameters
	options := &terraform.Options{
		// Initialize with necessary options
	}
	config := core.RunTime{
		// Initialize config
	}

	// Call the function under test
	err := TgApply(t, options, mockExecutor, config, cmdMockExecutor)

	// Assertions
	assert.NoError(t, err)
	mockExecutor.AssertExpectations(t)
}

func TestMockTgDestroy_Success(t *testing.T) {
	// Create a mock executors
	mockExecutor := new(MockTerragruntExecutor)

	cmdMockExecutor := new(MockCommandExecutor)

	// Set up expected behavior
	mockExecutor.On("TgDestroyAllE", t, mock.AnythingOfType("*terraform.Options")).Return("Mocked output", nil)

	// Prepare other parameters
	options := &terraform.Options{
		// Initialize with necessary options
	}
	config := core.RunTime{
		// Initialize config
	}

	// Call the function under test
	err := TgDestroy(t, options, mockExecutor, config, cmdMockExecutor, false)

	// Assertions
	assert.NoError(t, err)
	mockExecutor.AssertExpectations(t)
}

func TestMockTgDestroy_Failure(t *testing.T) {
	// Create a mock executors
	mockExecutor := new(MockTerragruntExecutor)

	cmdMockExecutor := new(MockCommandExecutor)

	// Set up expected behavior
	mockExecutor.On("TgDestroyAllE", t, mock.AnythingOfType("*terraform.Options")).Return("", fmt.Errorf("Mocked error"))

	// Prepare other parameters
	options := &terraform.Options{
		// Initialize with necessary options
	}
	config := core.RunTime{}

	// Call the function under test
	err := TgDestroy(t, options, mockExecutor, config, cmdMockExecutor, false)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "restore vars file failed")
	mockExecutor.AssertExpectations(t)
}

func TestMockTgApply_Failure(t *testing.T) {
	// Create a mock executors
	mockExecutor := new(MockTerragruntExecutor)

	cmdMockExecutor := new(MockCommandExecutor)

	// Set up expected behavior
	mockExecutor.On("TgApplyAllE", t, mock.AnythingOfType("*terraform.Options")).Return("", fmt.Errorf("Mocked error"))

	// Prepare other parameters
	options := &terraform.Options{
		// Initialize with necessary options
	}
	config := core.RunTime{}

	// Mock plugin_cache if necessary
	// You can create a mock for plugin_cache.ClearFolder if it interacts with external systems

	// Call the function under test
	err := TgApply(t, options, mockExecutor, config, cmdMockExecutor)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "restore vars file failed")
	mockExecutor.AssertExpectations(t)
}

func TestMockPauseTest(t *testing.T) {
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
	PauseTest(t, config, mockLogger, mockSleeper)

	// Assertions
	mockLogger.AssertExpectations(t)
	mockSleeper.AssertExpectations(t)
}

func TestTerragrunt(t *testing.T) {
	t.Parallel()

	os.Setenv("TT_TERRAGRUNT_ROOT_DIR", "../../example")
	os.Setenv("TT_PAUSE", "2")
	config := core.NewConfig()

	originalContent, err := core.UpdateVarsFile(t, config, core.OsFileSystem{})
	if err != nil {
		t.Fatalf("failed to update %s file, err:%v", config.VarsFile, err)
	}
	config.Content = string(originalContent)

	iamOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir:    "../../example/app/iam",
		TerraformBinary: "terragrunt",
		Vars: map[string]interface{}{
			"iam_policy_name": "TestDummy-" + parameters.AWSRegion,
		},
	})

	// Create an instance of the real executor
	executor := &RealTerragruntExecutor{}
	cmdExecutor := &RealCommandExecutor{}
	logger := &RealLogger{}
	sleeper := &RealSleeper{}

	defer func() {
		if err := TgDestroy(t, iamOptions, executor, config, cmdExecutor, true); err != nil {
			t.Fatalf("Error: %v\n", err)
		}
	}()

	if err := TgApply(t, iamOptions, executor, config, cmdExecutor); err != nil {
		t.Fatalf("Error: %v\n", err)
	}

	// IAM policy test cases
	policy_arn := terraform.Output(t, iamOptions, "policy_arn")
	assert.Equal(t, policy_arn, "arn:aws:iam::"+parameters.AWSAccountID+":policy/TestDummy-us-east-1", "Policy arn should match arn:aws:iam::"+parameters.AWSAccountID+":policy/TestDummy-us-east-1")

	iam2Options := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir:    "../../example/app/iam2",
		TerraformBinary: "terragrunt",
	})

	defer func() {
		if err := TgDestroy(t, iam2Options, executor, config, cmdExecutor, false); err != nil {
			t.Fatalf("Error: %v\n", err)
		}
	}()

	if err := TgApply(t, iam2Options, executor, config, cmdExecutor); err != nil {
		t.Fatalf("Error: %v\n", err)
	}

	// IAM policy test cases
	policy2_arn := terraform.Output(t, iam2Options, "policy_arn")
	assert.Equal(t, policy2_arn, "arn:aws:iam::"+parameters.AWSAccountID+":policy/DummyTest2-us-east-1", "Policy arn should match arn:aws:iam::"+parameters.AWSAccountID+":policy/DummyTest-us-east-1")

	//pause test
	PauseTest(t, config, logger, sleeper)
}

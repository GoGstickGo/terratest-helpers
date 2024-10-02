package terragrunt_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/GoGstickGo/terratest-helpers/core"
	"github.com/GoGstickGo/terratest-helpers/pkg/parameters"
	"github.com/GoGstickGo/terratest-helpers/pkg/terragrunt"
	"github.com/GoGstickGo/terratest-helpers/pkg/testutils"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	t.Parallel()
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
	err := terragrunt.Apply(t, options, mockExecutor, config, cmdMockExecutor)

	// Assertions
	require.NoError(t, err)
	mockExecutor.AssertExpectations(t)
}

func TestMockTgDestroy_Success(t *testing.T) {
	t.Parallel()
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
	err := terragrunt.Destroy(t, options, mockExecutor, config, cmdMockExecutor, false)

	// Assertions
	require.NoError(t, err)
	mockExecutor.AssertExpectations(t)
}

func TestMockTgDestroy_Failure(t *testing.T) {
	t.Parallel()
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
	err := terragrunt.Destroy(t, options, mockExecutor, config, cmdMockExecutor, false)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "restore vars file failed")
	mockExecutor.AssertExpectations(t)
}

func TestMockTgApply_Failure(t *testing.T) {
	t.Parallel()
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
	err := terragrunt.Apply(t, options, mockExecutor, config, cmdMockExecutor)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "restore vars file failed")
	mockExecutor.AssertExpectations(t)
}

func TestTerragrunt(t *testing.T) {
	t.Parallel()

	t.Setenv("TT_TERRAGRUNT_ROOT_DIR", "../../example")
	t.Setenv("TT_PAUSE", "2")
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
	executor := &terragrunt.RealTerragruntExecutor{}
	cmdExecutor := &terragrunt.RealCommandExecutor{}
	logger := &testutils.RealLogger{}
	sleeper := &testutils.RealSleeper{}

	defer func() {
		if err := terragrunt.Destroy(t, iamOptions, executor, config, cmdExecutor, true); err != nil {
			t.Fatalf("Error: %v\n", err)
		}
	}()

	if err := terragrunt.Apply(t, iamOptions, executor, config, cmdExecutor); err != nil {
		t.Fatalf("Error: %v\n", err)
	}

	// IAM policy test cases
	policyArn := terraform.Output(t, iamOptions, "policy_arn")
	assert.Equal(t, policyArn, "arn:aws:iam::"+parameters.AWSAccountID+":policy/TestDummy-us-east-1", "Policy arn should match arn:aws:iam::"+parameters.AWSAccountID+":policy/TestDummy-us-east-1")

	iam2Options := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir:    "../../example/app/iam2",
		TerraformBinary: "terragrunt",
	})

	defer func() {
		if err := terragrunt.Destroy(t, iam2Options, executor, config, cmdExecutor, false); err != nil {
			t.Fatalf("Error: %v\n", err)
		}
	}()

	if err := terragrunt.Apply(t, iam2Options, executor, config, cmdExecutor); err != nil {
		t.Fatalf("Error: %v\n", err)
	}

	// IAM policy test cases
	policy2Arn := terraform.Output(t, iam2Options, "policy_arn")
	assert.Equal(t, policy2Arn, "arn:aws:iam::"+parameters.AWSAccountID+":policy/DummyTest2-us-east-1", "Policy arn should match arn:aws:iam::"+parameters.AWSAccountID+":policy/DummyTest-us-east-1")

	// Pause test
	testutils.PauseTest(t, config, logger, sleeper)
}

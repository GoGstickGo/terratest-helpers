# Terratest-Helpers

## Overview

Terratest-Helpers is a collection of utility functions and tools to streamline and simplify testing infrastructure as code (IaC) with Terratest. It provides helper functions for common tasks such as setting up testing environments, executing Terraform or Terragrunt commands, asserting test outcomes, and more.


## Installation

To use Terratest-Helpers in your Go project, you can include it as a dependency using Go modules. Add the following import statement to your Go code:

```go
import "github.com/GoGstickGo/terratest-helpers"
```

Then, run `go mod tidy` to ensure the module is added to your `go.mod` file.

## Features

- **Wrapper Functions**: Terratest-Helpers provides wrapper functions around Terratest and Terraform functionalities, making it easier to write and manage tests.
- **Configuration Management**: Offers utilities for managing configuration files, environment variables, and test fixtures.
- **Logging and Reporting**: Provides logging functions to output test results and detailed information during test execution.
- **Common Patterns**: Implements common testing patterns and best practices to accelerate test development and maintainability.
- **Sustainability**: Each test spin up  completely environments then it delete itself. Tests run in chain to avoid any clashesh.

## Usage

1. **Import the Library**: Import the `terratest-helpers` package in your Go code.
2. **Use Helper Functions**: Utilize the provided helper functions for your test cases, such as setting up testing environments, executing commands, and making assertions.
3. **Run Tests**: Run your tests using the `go test -timeout 20m` command as usual.

## Examples

[Example](example) folder shows terragrunt environment what terratest-helpers were optimised. 

Below are some examples of how to use Terratest-Helpers:

```go
// Example test case using Terratest-Helpers
func TestTerragrunt(t *testing.T) {
	t.Parallel()

	config := NewConfig("", "", "", "", false, false)

	originalContent, err := UpdateRootVars(t, config, config.Content)
	if err != nil {
		t.Fatalf("Error updating root_vars.hcl: %v", err)
	}

	iamOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir:    config.Paths.TerragruntDir + "app/iam",
		TerraformBinary: "terragrunt",
		Vars: map[string]interface{}{
			"iam_policy_name": "TestDummy-" + parameters.AWSRegion,
		},
	})

	defer func() {
		if err := TgDestroy(t, iamOptions, config, originalContent, true); err != nil {
			t.Fatalf("Error: %v\n", err)
		}
	}()

	if err := TgApply(t, iamOptions, config, originalContent); err != nil {
		t.Fatalf("Error: %v\n", err)
	}

	// IAM policy test cases
	policy_arn := terraform.Output(t, iamOptions, "policy_arn")
	assert.Equal(t, policy_arn, "arn:aws:iam::" + parameters.AWSAccountID + ":policy/TestDummy-us-east-1", "Policy arn should match arn:aws:iam::" + parameters.AWSAccountID + ":policy/TestDummy-us-east-1")
}
```

## Testing

1. **Unit Tests:**
   - Implemented using **mocking** and **dependency injection** to ensure components are tested in isolation.
   
2. **Integration Tests:**
   - **Availability:** Integration tests can be executed to verify the interaction between different system components.
   - **Requirements:**
     - An active **AWS account**.
     - Updates to the [parameters.go](pkg/parameters/parameters.go) file to configure necessary parameters.
## Ci/Cd

There is runnner setup to run unit tests. See [go-test.yml](.github/workflows/go-test.yml).

## Contributing

If you'd like to contribute to Terratest-Helpers, please fork the repository, make your changes, and submit a pull request. Contributions, bug reports, and feature requests are welcome!

## License

Terratest-Helpers is licensed under the [MIT License](LICENSE).

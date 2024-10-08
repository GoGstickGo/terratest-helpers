package terragrunt

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/GoGstickGo/terratest-helpers/core"
	"github.com/GoGstickGo/terratest-helpers/pkg/awsutils"
	"github.com/GoGstickGo/terratest-helpers/pkg/parameters"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

// TerragruntExecutor abstracts Terragrunt execution methods.
type Executor interface {
	TgApplyAllE(t *testing.T, options *terraform.Options) (string, error)
	TgDestroyAllE(t *testing.T, options *terraform.Options) (string, error)
	// Add other methods like TgInitAllE, TgDestroyAllE if needed.
}

type RealTerragruntExecutor struct{}

func (e *RealTerragruntExecutor) TgApplyAllE(t *testing.T, options *terraform.Options) (string, error) {
	return terraform.TgApplyAllE(t, options)
}

func (e *RealTerragruntExecutor) TgDestroyAllE(t *testing.T, options *terraform.Options) (string, error) {
	return terraform.TgDestroyAllE(t, options)
}

// CommandExecutor abstracts command execution.
type CommandExecutor interface {
	RunCommand(cmdName string, args []string, dir string, envVars map[string]string) ([]byte, error)
}

// RealCommandExecutor executes real system commands.
type RealCommandExecutor struct{}

func (e *RealCommandExecutor) RunCommand(cmdName string, args []string, dir string, envVars map[string]string) ([]byte, error) {
	cmd := exec.Command(cmdName, args...)
	cmd.Dir = dir

	// Prepare environment variables.
	cmd.Env = os.Environ() // Start with existing environment variables.
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	output, err := cmd.CombinedOutput()

	return output, err
}

func tGiNit(t *testing.T, terra *terraform.Options, config core.RunTime, executor CommandExecutor) error {
	logger.Log(t, "TerraGrunt init in progress")

	// Collect environment variables
	envVars := map[string]string{
		"TF_PLUGIN_CACHE_MAY_BREAK_DEPENDENCY_LOCK_FILE": "true",
		"TERRAGRUNT_DOWNLOAD":                            config.Paths.TgDownloadDir,
		"TF_PLUGIN_CACHE_DIR":                            config.Paths.TfPluginDir,
		"TERRAGRUNT_NO_AUTO_INIT":                        "true",
	}
	if config.IsDebug {
		envVars["TERRAGRUNT_LOG_LEVEL"] = "debug"
		envVars["TERRAGRUNT_DEBUG"] = ""
		envVars["TF_LOG"] = "DEBUG"
	}

	// Command and arguments.
	cmdName := "terragrunt"
	args := []string{"run-all", "init", "--terragrunt-non-interactive"}

	// Run the command
	output, err := executor.RunCommand(cmdName, args, terra.TerraformDir, envVars)
	if config.IsDebug {
		logger.Log(t, "init output: %s\n", string(output))
	}

	// Check for misconfigured plugin-cache.
	checkStr := []string{"cache", "previously"}
	for _, subStr := range checkStr {
		if !strings.ContainsAny(string(output), subStr) {
			if err := core.ClearFolder(t, config, core.OsFileSystem{}); err != nil {

				return fmt.Errorf("clearing chache folder failed: %w", err)
			}
			if err := core.RestoreVarsFile(t, config, core.OsFileSystem{}); err != nil {

				return fmt.Errorf("restore vars file failed: %w", err)
			}

			return fmt.Errorf("\nplugin cache out of order:\n %s", string(output))
		}
	}

	if err != nil {
		if err := core.ClearFolder(t, config, core.OsFileSystem{}); err != nil {

			return fmt.Errorf("clearing chache folder failed: %w", err)
		}
		if err := core.RestoreVarsFile(t, config, core.OsFileSystem{}); err != nil {

			return fmt.Errorf("restore vars file failed: %w", err)
		}

		return fmt.Errorf("%w\nOutput:\n%s", err, output)
	}
	logger.Log(t, "terragrunt init completed")

	return nil
}

func Apply(t *testing.T, options *terraform.Options, executor Executor, config core.RunTime, cmdExecutor CommandExecutor) error {

	if config.IsPluginCache {
		if err := tGiNit(t, options, config, cmdExecutor); err != nil {

			return fmt.Errorf("terragrunt init failed: %w", err)
		}
	}

	if config.IsDebug && !config.IsPluginCache {
		os.Setenv("TERRAGRUNT_LOG_LEVEL", "debug")
		os.Setenv("TERRAGRUNT_DEBUG", "")
		os.Setenv("TF_LOG", "DEBUG")
	}

	logger.Log(t, "TerraGrunt Apply in progress")
	output, err := executor.TgApplyAllE(t, options)
	if err != nil {
		if config.IsPluginCache {
			// Remove cached files.
			if err := core.ClearFolder(t, config, core.OsFileSystem{}); err != nil {

				return fmt.Errorf("error clearing cache folder: %w", err)
			}
		}
		if err := core.RestoreVarsFile(t, config, core.OsFileSystem{}); err != nil {

			return fmt.Errorf("restore vars file failed: %w", err)
		}

		return fmt.Errorf("failed to apply Terragrunt ,output: %s, error: %w", output, err)
	}

	return nil
}

func Destroy(t *testing.T, options *terraform.Options, executor Executor, config core.RunTime, cmdExecutor CommandExecutor, restore bool) error {
	logger.Log(t, "Defer func started")

	if config.IsPluginCache {
		if err := tGiNit(t, options, config, cmdExecutor); err != nil {

			return fmt.Errorf("terragrunt init failed: %w", err)
		}
	}

	logger.Log(t, "TerraGrunt destroy in progress")
	stdout, err := executor.TgDestroyAllE(t, options)
	if err != nil {
		if err := core.RestoreVarsFile(t, config, core.OsFileSystem{}); err != nil {

			return fmt.Errorf("restore vars file failed: %w", err)
		}

		return fmt.Errorf("failed to destroy with Terragrunt ,output: %s, error: %w", stdout, err)
	}

	if config.IsPluginCache {
		// Remove cached files.
		if err := core.ClearFolder(t, config, core.OsFileSystem{}); err != nil {

			return fmt.Errorf("error clearing cache folder: %w", err)
		}
	}

	var errs []error

	ec2Client, err := awsutils.LoadEC2Client(parameters.AWSRegion)
	if err != nil {

		return fmt.Errorf("error loading EC2 client: %w", err)
	}

	if restore {
		// Restore the original content of root_vars.hcl.
		if err = core.RestoreVarsFile(t, config, core.OsFileSystem{}); err != nil {
			errs = append(errs, fmt.Errorf("error restoring %s: %w", config.VarsFile, err))
		}
		if _, err := awsutils.RemoveENI(t, parameters.VPCId, ec2Client); err != nil {
			errs = append(errs, fmt.Errorf("error deleting AWS EC2 ENIs: %w", err))
		}
	}
	if len(errs) > 0 {

		return fmt.Errorf("restore failed: %v", errs)
	}

	return nil
} /*func UpdateTerraformHook(dir, key, newLine string) error {
	log.Print("Update terraform_hook")
	rootTGPath := filepath.Join(dir, "terragrunt.hcl")
	// Read the content of the file
	content, err := os.ReadFile(rootTGPath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Convert content to string
	fileContent := string(content)

	// Update the line in the content
	lines := strings.Split(fileContent, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, key) {
			lines[i] = newLine
		}
	}

	// Write the updated content back to the file
	err = os.WriteFile(rootTGPath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}*/

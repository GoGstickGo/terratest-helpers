package core

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/GoGstickGo/terratest-helpers/pkg/parameters"
)

type FolderPaths struct {
	HomeDir       string
	TgDownloadDir string
	TfPluginDir   string
	TerragruntDir string
}

type RunTime struct {
	Paths         FolderPaths
	Content       string
	VarsFile      string
	IsPluginCache bool
	IsDebug       bool
	Pause         time.Duration
}

// NewFolderConfig creates a new instance of FolderConfig with default values.
func NewConfig() RunTime {

	terragruntDir := getEnvVar("TT_TERRAGRUNT_ROOT_DIR", "../../")
	homeDir := getEnvVar("TT_HOME_DIR", "")
	tgDownloadDir := getEnvVar("TT_TERRAGRUNT_DOWNLOAD_DIR", "")
	tfPluginDir := getEnvVar("TT_TERRAGRUNT_PLUGIN_DIR", "")
	content := getEnvVar("TT_CONTENT", parameters.TGRootVars)
	varsFile := getEnvVar("TT_VARS_FILE", "root_vars.hcl")
	isPluginCache := getEnvVarBool("TT_PLUGIN_CACHE", false)
	isDebug := getEnvVarBool("TT_DEBUG", false)
	pause := getEnvVarDuration("TT_PAUSE", 0)

	// Set default values
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}

	if tgDownloadDir == "" {
		tgDownloadDir = filepath.Join(homeDir, ".terragrunt-cache")
	}

	if tfPluginDir == "" {
		tfPluginDir = filepath.Join(tgDownloadDir, ".plugins")
	}

	return RunTime{
		Paths: FolderPaths{
			TerragruntDir: terragruntDir,
			HomeDir:       homeDir,
			TgDownloadDir: tgDownloadDir,
			TfPluginDir:   tfPluginDir,
		},
		VarsFile:      varsFile,
		IsDebug:       isDebug,
		Content:       content,
		IsPluginCache: isPluginCache,
		Pause:         pause,
	}
}

func getEnvVar(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func getEnvVarBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	temp, _ := strconv.ParseBool(value)

	return temp
}

func getEnvVarDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	tempInt, _ := strconv.Atoi(value)

	return time.Duration(tempInt) * time.Minute
}

var ErrFailedToReadDirectory = errors.New("failed to read directory")

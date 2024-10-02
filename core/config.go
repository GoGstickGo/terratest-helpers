package core

import (
	"flag"
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

	terragruntDir := flag.String("terragruntDir", "../../", "Terragrunt directory")
	homeDir := flag.String("homeDir", "", "Home directory default return os.UserHomeDir()")
	tgDownloadDir := flag.String("tgDownloadDir", "", "Terragrunt download directory defaults to homeDir/.terragrunt-cache")
	tfPluginDir := flag.String("tfPluginDir", "", "Terraform plugin directory, defaults to tgDownloadDir/.plugins")
	content := flag.String("content", parameters.TGRootVars, "Content")
	varsFile := flag.String("varsFile", "root_vars.hcl", "Variable file for AWS account/region details.")
	isPluginCache := flag.Bool("isPluginCache", false, "Enable plugin cache functionality")
	isDebug := flag.Bool("isDebug", false, "Enable debug mode")
	pause := flag.Duration("pause", 0, "Amount of minutes to pause test execution before destroying environments")

	flag.Parse()

	//set default values
	if *homeDir == "" {
		*homeDir, _ = os.UserHomeDir()
	}

	if *tgDownloadDir == "" {
		*tgDownloadDir = filepath.Join(*homeDir, ".terragrunt-cache")
	}

	if *tfPluginDir == "" {
		*tfPluginDir = filepath.Join(*tgDownloadDir, ".plugins")
	}

	// handle env vars
	if os.Getenv("TT_TERRAGRUNT_ROOT_DIR") != "" {
		*terragruntDir = os.Getenv("TT_TERRAGRUNT_ROOT_DIR")
	}

	if os.Getenv("TT_HOME_DIR") != "" {
		*homeDir = os.Getenv("TT_HOME_DIR")
	}

	if os.Getenv("TT_TERRAGRUNT_DOWNLOAD_DIR") != "" {
		*tgDownloadDir = os.Getenv("TT_TERRAGRUNT_DOWNLOAD_DIR")
	}

	if os.Getenv("TT_TERRAGRUNT_PLUGIN_DIR") != "" {
		*tfPluginDir = os.Getenv("TT_TERRAGRUNT_DOWNLOAD_DIR")
	}

	if os.Getenv("TT_VARS_FILE") != "" {
		*varsFile = os.Getenv("TT_VARS_FILE")
	}

	if os.Getenv("TT_CONTENT") != "" {
		*content = os.Getenv("TT_CONTENT")
	}

	if os.Getenv("TT_DEBUG") != "" {
		temp := os.Getenv("TT_DEBUG")
		*isDebug, _ = strconv.ParseBool(temp)
	}

	if os.Getenv("TT_PLUGIN_CACHE") != "" {
		temp := os.Getenv("TT_PLUGIN_CACHE")
		*isPluginCache, _ = strconv.ParseBool(temp)
	}

	if os.Getenv("TT_PAUSE") != "" {
		temp := os.Getenv("TT_PAUSE")
		tempInt, _ := strconv.Atoi(temp)
		*pause = time.Duration(tempInt) * time.Minute
	}

	return RunTime{
		Paths: FolderPaths{
			TerragruntDir: *terragruntDir,
			HomeDir:       *homeDir,
			TgDownloadDir: *tgDownloadDir,
			TfPluginDir:   *tfPluginDir,
		},
		VarsFile:      *varsFile,
		IsDebug:       *isDebug,
		Content:       *content,
		IsPluginCache: *isPluginCache,
		Pause:         *pause,
	}
}

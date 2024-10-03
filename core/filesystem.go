package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
)

// FileSystem interface abstracts file system operations.
type FileSystem interface {
	ReadDir(dirname string) ([]os.DirEntry, error)
	RemoveAll(path string) error
	ReadFile(path string) ([]byte, error)
	WriteFile(filename string, data []byte, perm fs.FileMode) error
}

// OsFileSystem is a real implementation of FileSystem.
type OsFileSystem struct{}

func (OsFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) {
	return os.ReadDir(dirname)
}

func (OsFileSystem) RemoveAll(path string) error {

	return os.RemoveAll(path)
}

func (OsFileSystem) ReadFile(path string) ([]byte, error) {

	return os.ReadFile(path)
}

func (OsFileSystem) WriteFile(filename string, data []byte, perm fs.FileMode) error {

	return os.WriteFile(filename, data, perm)
}

// clearFolder removes all subfolders within the specified directory, excluding a specific subfolder.
// It takes a FolderConfig struct as input containing the paths to relevant directories.
func ClearFolder(t *testing.T, cfg RunTime, fs FileSystem) error {
	// Log the start of cache folder clearing
	logger.Log(t, "Cache folder clearing in progress")

	// Read directory entries within the target download directory
	entries, err := fs.ReadDir(cfg.Paths.TgDownloadDir)
	if err != nil {
		// Return an error if reading directory entries fails
		return fmt.Errorf("failed to read directory: %v", err)
	}

	// Iterate over directory entries.
	for _, entry := range entries {
		// Check if the entry is a directory and not the ".plugins" directory.
		if entry.IsDir() && entry.Name() != ".plugins" {
			// Construct the full path of the subfolder
			subfolderPath := filepath.Join(cfg.Paths.TgDownloadDir, entry.Name())

			// Attempt to remove the subfolder and its contents.
			if err := fs.RemoveAll(subfolderPath); err != nil {
				// Return an error if removal fails.
				return fmt.Errorf("failed to remove subfolder %s: %v", subfolderPath, err)
			}
		}
	}
	logger.Log(t, "Cache folder cleared")
	// Return nil to indicate success.
	return nil
}

func UpdateVarsFile(t *testing.T, cfg RunTime, fs FileSystem) ([]byte, error) {
	logger.Log(t, "Update "+cfg.VarsFile)
	rootVarsPath := filepath.Join(cfg.Paths.TerragruntDir, cfg.VarsFile)

	// Read the current content.
	currentContent, err := fs.ReadFile(rootVarsPath)
	if err != nil {
		return nil, fmt.Errorf("readFile func failed to read %s: %v", cfg.VarsFile, err)
	}

	// Store the original content.
	originalContent := make([]byte, len(currentContent))
	copy(originalContent, currentContent)

	// Append or overwrite the content.
	err = fs.WriteFile(rootVarsPath, []byte(cfg.Content), 0644)
	if err != nil {
		return nil, fmt.Errorf("writeFile func failed to write %s: %v", cfg.VarsFile, err)
	}

	logger.Log(t, "Updated "+cfg.VarsFile)

	return originalContent, nil
}

// RestoreVarsFile restores the original content.
func RestoreVarsFile(t *testing.T, cfg RunTime, fs FileSystem) error {
	rootVarsPath := filepath.Join(cfg.Paths.TerragruntDir, cfg.VarsFile)
	logger.Log(t, "Restore "+cfg.VarsFile)

	return fs.WriteFile(rootVarsPath, []byte(cfg.Content), 0644)
}

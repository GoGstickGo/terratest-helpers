package core

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockFileSystem struct {
	mock.Mock
}

// Mock ReadDir method
func (m *MockFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) {
	args := m.Called(dirname)
	return args.Get(0).([]os.DirEntry), args.Error(1)
}

// Mock RemoveAll method
func (m *MockFileSystem) RemoveAll(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

// Mock ReadFile method
func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	args := m.Called(path)
	return args.Get(0).([]byte), args.Error(1)
}

// Mock WriteFile method
func (m *MockFileSystem) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	args := m.Called(filename, data, perm)
	return args.Error(0)
}

type MockDirEntry struct {
	name  string
	isDir bool
}

func (m MockDirEntry) Name() string {
	return m.name
}

func (m MockDirEntry) IsDir() bool {
	return m.isDir
}

func (m MockDirEntry) Type() fs.FileMode {
	if m.isDir {
		return fs.ModeDir
	}
	return 0
}

func (m MockDirEntry) Info() (fs.FileInfo, error) {
	return nil, nil
}

func TestMockClearFolder(t *testing.T) {
	// Create a mock file system
	mockFS := new(MockFileSystem)

	// Define the configuration
	cfg := RunTime{
		Paths: FolderPaths{
			TgDownloadDir: "test/download-dir",
		},
	}

	// Set up mock directory entries
	mockEntries := []os.DirEntry{
		MockDirEntry{name: "folder1", isDir: true},
		MockDirEntry{name: "folder2", isDir: true},
		MockDirEntry{name: ".plugins", isDir: true},
		MockDirEntry{name: "file1.txt", isDir: false},
	}

	// Set expectations for ReadDir
	mockFS.On("ReadDir", cfg.Paths.TgDownloadDir).Return(mockEntries, nil)

	// Set expectations for RemoveAll (only for directories that are not ".plugins")
	mockFS.On("RemoveAll", filepath.Join(cfg.Paths.TgDownloadDir, "folder1")).Return(nil)
	mockFS.On("RemoveAll", filepath.Join(cfg.Paths.TgDownloadDir, "folder2")).Return(nil)

	err := ClearFolder(t, cfg, mockFS)
	if err != nil {
		t.Errorf("ClearFolder returned error: %v", err)
	}

	mockFS.AssertExpectations(t)
}

func TestMockUpdateVarsFile(t *testing.T) {

	// Create a mock file system
	mockFS := new(MockFileSystem)

	// Define the configuration
	cfg := RunTime{
		VarsFile: "root_vars.hcl",
		Paths: FolderPaths{
			TerragruntDir: "test/terragrunt",
		},
		Content: "new content",
	}

	rootVarsPath := filepath.Join(cfg.Paths.TerragruntDir, cfg.VarsFile)

	// Set up the expected behavior for ReadFile
	originalContent := []byte("original content")
	mockFS.On("ReadFile", rootVarsPath).Return(originalContent, nil)

	// Set up the expected behavior for WriteFile
	mockFS.On("WriteFile", rootVarsPath, []byte(cfg.Content), fs.FileMode(0644)).Return(nil)

	// Call the function under test
	result, err := UpdateVarsFile(t, cfg, mockFS)
	if err != nil {
		t.Errorf("UpdateVarsFile returned error: %v", err)
	}

	// Assert that the returned original content is correct
	if !bytes.Equal(result, originalContent) {
		t.Errorf("Expected original content to be %s, got %s", originalContent, result)
	}

	// Assert that all expectations were met
	mockFS.AssertExpectations(t)

}

func TestMockRestoreVarsFile(t *testing.T) {

	// Create a mock file system
	mockFS := new(MockFileSystem)

	// Define the configuration
	cfg := RunTime{
		VarsFile: "root_vars.hcl",
		Paths: FolderPaths{
			TerragruntDir: "test/terragrunt",
		},
		Content: "changed content",
	}

	rootVarsPath := filepath.Join(cfg.Paths.TerragruntDir, cfg.VarsFile)

	// Set up the expected behavior for WriteFile
	mockFS.On("WriteFile", rootVarsPath, []byte(cfg.Content), fs.FileMode(0644)).Return(nil)

	if err := RestoreVarsFile(t, cfg, mockFS); err != nil {
		t.Errorf("RestoreVarsFile returned error: %v", err)
	}

	// Assert that all expectations were met
	mockFS.AssertExpectations(t)
}

// nolint:testpackage
package utils

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// GetSourceListTestSuite defines the test suite for GetSourceList function.
type GetSourceListTestSuite struct {
	suite.Suite
	tempDir string
}

// SetupTest is called before each test method, creating a fresh test environment.
func (suite *GetSourceListTestSuite) SetupTest() {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_source_list")
	suite.Require().NoError(err, "Failed to create temp dir")
	suite.tempDir = tempDir

	// Create test files and directories
	testFiles := []string{
		"file1.go",
		"file2.txt",
		"dir1/file3.go",
		"dir1/file4.txt",
		"dir2/file5.js",
		".hidden",
		"node_modules/package.json",
		"build/output.bin",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(suite.tempDir, file)
		// Create directory if it doesn't exist
		dirErr := os.MkdirAll(filepath.Dir(filePath), 0755)
		suite.Require().NoError(dirErr, "Failed to create directory for %s", file)
		// Create file
		fileErr := os.WriteFile(filePath, []byte("test content"), 0644)
		suite.Require().NoError(fileErr, "Failed to create file %s", file)
	}

	// Create .gitignore file
	gitignoreContent := `node_modules/
build/
*.bin
.hidden
`
	gitignorePath := filepath.Join(suite.tempDir, ".gitignore")
	err = os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	suite.Require().NoError(err, "Failed to create .gitignore")
}

// TearDownTest is called after each test method, cleaning up the test environment.
func (suite *GetSourceListTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// getRelativeFiles converts absolute file paths to relative paths for easier testing.
func (suite *GetSourceListTestSuite) getRelativeFiles(files []string, skipGitignore bool) []string {
	relativeFiles := make([]string, 0, len(files))
	for _, file := range files {
		relPath, err := filepath.Rel(suite.tempDir, file)
		suite.Require().NoError(err, "Failed to get relative path for %s", file)
		// Skip gitignore files if requested
		if skipGitignore && strings.HasSuffix(relPath, ".gitignore") {
			continue
		}
		relativeFiles = append(relativeFiles, filepath.ToSlash(relPath))
	}
	sort.Strings(relativeFiles)
	return relativeFiles
}

// TestWithGitignore tests file discovery with gitignore rules enabled.
func (suite *GetSourceListTestSuite) TestWithGitignore() {
	options := &GetSourceListOptions{
		RespectGitignore: true,
		IncludeHidden:    false,
	}
	files, err := GetSourceList(suite.tempDir, options)
	suite.Require().NoError(err, "GetSourceList failed")

	relativeFiles := suite.getRelativeFiles(files, false)

	expected := []string{
		"dir1/file3.go",
		"dir1/file4.txt",
		"dir2/file5.js",
		"file1.go",
		"file2.txt",
	}
	sort.Strings(expected)

	suite.Equal(expected, relativeFiles, "Files should match expected list with gitignore rules")
}

// TestWithoutGitignore tests file discovery with gitignore rules disabled.
func (suite *GetSourceListTestSuite) TestWithoutGitignore() {
	options := &GetSourceListOptions{
		RespectGitignore: false,
		IncludeHidden:    false,
	}
	files, err := GetSourceList(suite.tempDir, options)
	suite.Require().NoError(err, "GetSourceList failed")

	relativeFiles := suite.getRelativeFiles(files, true)

	expected := []string{
		"build/output.bin",
		"dir1/file3.go",
		"dir1/file4.txt",
		"dir2/file5.js",
		"file1.go",
		"file2.txt",
		"node_modules/package.json",
	}
	sort.Strings(expected)

	suite.Equal(expected, relativeFiles, "Files should include gitignore-excluded files when RespectGitignore is false")
}

// TestWithExtensionFilter tests file discovery with extension filtering.
func (suite *GetSourceListTestSuite) TestWithExtensionFilter() {
	options := &GetSourceListOptions{
		RespectGitignore: true,
		IncludeHidden:    false,
		Extensions:       []string{".go"},
	}
	files, err := GetSourceList(suite.tempDir, options)
	suite.Require().NoError(err, "GetSourceList failed")

	relativeFiles := suite.getRelativeFiles(files, false)

	expected := []string{
		"dir1/file3.go",
		"file1.go",
	}
	sort.Strings(expected)

	suite.Equal(expected, relativeFiles, "Should only return .go files when extension filter is applied")
}

// TestWithHiddenFiles tests file discovery with hidden files included.
func (suite *GetSourceListTestSuite) TestWithHiddenFiles() {
	options := &GetSourceListOptions{
		RespectGitignore: false,
		IncludeHidden:    true,
	}
	files, err := GetSourceList(suite.tempDir, options)
	suite.Require().NoError(err, "GetSourceList failed")

	// Check if hidden files are included
	var hasHiddenFile bool
	for _, file := range files {
		if strings.Contains(file, ".hidden") {
			hasHiddenFile = true
			break
		}
	}

	suite.True(hasHiddenFile, "Hidden files should be included when IncludeHidden is true")
}

// TestWithCustomGitignoreFilePath tests file discovery with custom gitignore file path.
func (suite *GetSourceListTestSuite) TestWithCustomGitignoreFilePath() {
	// Create a custom gitignore file with different rules
	customGitignoreContent := `*.txt
dir2/
`
	customGitignorePath := filepath.Join(suite.tempDir, "custom.gitignore")
	err := os.WriteFile(customGitignorePath, []byte(customGitignoreContent), 0644)
	suite.Require().NoError(err, "Failed to create custom .gitignore")

	options := &GetSourceListOptions{
		RespectGitignore:  true,
		IncludeHidden:     false,
		GitignoreFilePath: customGitignorePath,
	}
	files, err := GetSourceList(suite.tempDir, options)
	suite.Require().NoError(err, "GetSourceList failed")

	relativeFiles := suite.getRelativeFiles(files, true)

	// With custom gitignore rules: *.txt and dir2/ should be ignored
	// So we should only have .go files, node_modules/, and build/ files
	// Note: .hidden is excluded because IncludeHidden is false
	expected := []string{
		"build/output.bin",
		"dir1/file3.go",
		"file1.go",
		"node_modules/package.json",
	}
	sort.Strings(expected)

	suite.Equal(expected, relativeFiles, "Should respect custom gitignore file rules")
}

// TestWithNonExistentGitignoreFilePath tests graceful handling of non-existent gitignore file.
func (suite *GetSourceListTestSuite) TestWithNonExistentGitignoreFilePath() {
	options := &GetSourceListOptions{
		RespectGitignore:  true,
		IncludeHidden:     false,
		GitignoreFilePath: "/non/existent/path/.gitignore",
	}

	// Should not fail, just ignore the gitignore rules
	files, err := GetSourceList(suite.tempDir, options)
	suite.Require().NoError(err, "GetSourceList should not fail with non-existent gitignore file")

	relativeFiles := suite.getRelativeFiles(files, true)

	// Should have at least the non-hidden files
	suite.GreaterOrEqual(len(relativeFiles), 6,
		"Should have at least 6 files when gitignore file doesn't exist: %v", relativeFiles)
}

// TestWithNilOptions tests that GetSourceList works with nil options (using defaults).
func (suite *GetSourceListTestSuite) TestWithNilOptions() {
	// Test with nil options - should use defaults
	files, err := GetSourceList(suite.tempDir, nil)
	suite.Require().NoError(err, "GetSourceList failed with nil options")

	suite.NotEmpty(files, "Should find files when using default options")

	relativeFiles := suite.getRelativeFiles(files, false)

	// With default options (RespectGitignore=true, IncludeHidden=false),
	// should behave same as TestWithGitignore
	expected := []string{
		"dir1/file3.go",
		"dir1/file4.txt",
		"dir2/file5.js",
		"file1.go",
		"file2.txt",
	}
	sort.Strings(expected)

	suite.Equal(expected, relativeFiles, "Nil options should use default behavior")
}

// TestWithMultipleExtensions tests file discovery with multiple extension filters.
func (suite *GetSourceListTestSuite) TestWithMultipleExtensions() {
	options := &GetSourceListOptions{
		RespectGitignore: true,
		IncludeHidden:    false,
		Extensions:       []string{".go", ".js"},
	}
	files, err := GetSourceList(suite.tempDir, options)
	suite.Require().NoError(err, "GetSourceList failed")

	relativeFiles := suite.getRelativeFiles(files, false)

	expected := []string{
		"dir1/file3.go",
		"dir2/file5.js",
		"file1.go",
	}
	sort.Strings(expected)

	suite.Equal(expected, relativeFiles, "Should return files with .go and .js extensions only")
}

// EmptyDirectoryTestSuite tests behavior with an empty directory.
type EmptyDirectoryTestSuite struct {
	suite.Suite
	tempDir string
}

// SetupTest creates an empty directory for testing.
func (suite *EmptyDirectoryTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "test_empty_dir")
	suite.Require().NoError(err, "Failed to create temp dir")
	suite.tempDir = tempDir
}

// TearDownTest cleans up the test directory.
func (suite *EmptyDirectoryTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// TestEmptyDirectory tests that empty directories are handled correctly.
func (suite *EmptyDirectoryTestSuite) TestEmptyDirectory() {
	options := &GetSourceListOptions{
		RespectGitignore: true,
		IncludeHidden:    false,
	}
	files, err := GetSourceList(suite.tempDir, options)
	suite.Require().NoError(err, "GetSourceList failed")

	suite.Empty(files, "Empty directory should return no files")
}

// TestGetSourceList runs all the test suites.
func TestGetSourceList(t *testing.T) {
	suite.Run(t, new(GetSourceListTestSuite))
	suite.Run(t, new(EmptyDirectoryTestSuite))
}

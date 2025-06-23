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

// TestWithNonExistentDirectory tests behavior when directory doesn't exist.
func (suite *GetSourceListTestSuite) TestWithNonExistentDirectory() {
	nonExistentDir := filepath.Join(suite.tempDir, "non_existent_directory")

	options := &GetSourceListOptions{
		RespectGitignore: true,
		IncludeHidden:    false,
	}

	files, err := GetSourceList(nonExistentDir, options)
	suite.Error(err, "Should return error when directory doesn't exist")
	suite.Empty(files, "Should return empty slice when directory doesn't exist")
}

// TestWithGitDirectory tests that .git directories are skipped during traversal.
func (suite *GetSourceListTestSuite) TestWithGitDirectory() {
	// Create a project directory structure
	projectDir := filepath.Join(suite.tempDir, "project")
	err := os.Mkdir(projectDir, 0755)
	suite.Require().NoError(err, "Failed to create project directory")

	// Create some normal files
	normalFile := filepath.Join(projectDir, "main.go")
	err = os.WriteFile(normalFile, []byte("package main"), 0644)
	suite.Require().NoError(err, "Failed to create normal file")

	// Create a .git directory with some contents
	gitDir := filepath.Join(projectDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	suite.Require().NoError(err, "Failed to create .git directory")

	// Create some files inside .git directory
	gitConfigFile := filepath.Join(gitDir, "config")
	err = os.WriteFile(gitConfigFile, []byte("git config content"), 0644)
	suite.Require().NoError(err, "Failed to create git config file")

	gitObjectsDir := filepath.Join(gitDir, "objects")
	err = os.Mkdir(gitObjectsDir, 0755)
	suite.Require().NoError(err, "Failed to create git objects directory")

	gitObjectFile := filepath.Join(gitObjectsDir, "somehash")
	err = os.WriteFile(gitObjectFile, []byte("git object content"), 0644)
	suite.Require().NoError(err, "Failed to create git object file")

	options := &GetSourceListOptions{
		RespectGitignore: false,
		IncludeHidden:    false,
	}

	files, err := GetSourceList(projectDir, options)
	suite.Require().NoError(err, "GetSourceList failed")

	// Should only have the normal file, not any files from .git directory
	suite.Len(files, 1, "Should only have one file (excluding .git contents)")

	// Verify the file contains what we expect (main.go but not any .git files)
	hasMainGo := false
	hasGitFiles := false
	for _, file := range files {
		if strings.Contains(file, "main.go") {
			hasMainGo = true
		}
		if strings.Contains(file, ".git") {
			hasGitFiles = true
		}
	}

	suite.True(hasMainGo, "Should contain main.go file")
	suite.False(hasGitFiles, "Should not contain any files from .git directory")
}

// TestGetSourceList runs all the test suites.
func TestGetSourceList(t *testing.T) {
	suite.Run(t, new(GetSourceListTestSuite))
	suite.Run(t, new(EmptyDirectoryTestSuite))
}

// FuzzGetSourceList implements fuzz testing for GetSourceList function
// to test robustness with various random inputs and edge cases.
func FuzzGetSourceList(f *testing.F) {
	// Create a controlled temporary directory for safer fuzzing
	tempDir, err := os.MkdirTemp("", "fuzz_test_controlled")
	if err != nil {
		f.Fatal("Failed to create temp dir for fuzz test:", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files in the controlled environment
	testFiles := []string{"test.go", "test.txt", ".hidden", "sub/test.js"}
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		os.MkdirAll(filepath.Dir(filePath), 0755)
		os.WriteFile(filePath, []byte("test"), 0644)
	}

	// Add seed data for fuzzing - use relative paths within our controlled directory
	f.Add("test_subdir", true, false, "")
	f.Add("nonexistent", false, true, ".go")
	f.Add("", true, false, ".txt")
	f.Add(".", false, false, "")
	f.Add("sub", true, true, ".js")

	f.Fuzz(func(t *testing.T, relativeDir string, respectGitignore bool, includeHidden bool, extension string) {
		// Sanitize and construct safe directory path within our controlled temp directory
		// Avoid dangerous paths like "../", "/", etc.
		sanitizedRelDir := filepath.Clean(relativeDir)
		if strings.Contains(sanitizedRelDir, "..") ||
			strings.HasPrefix(sanitizedRelDir, "/") ||
			len(sanitizedRelDir) > 100 { // Limit path length
			// Use a safe default instead of dangerous paths
			sanitizedRelDir = "safe_subdir"
		}

		var testDir string
		if sanitizedRelDir == "" || sanitizedRelDir == "." {
			testDir = tempDir
		} else {
			testDir = filepath.Join(tempDir, sanitizedRelDir)
		}

		// Create test options from fuzz inputs
		var extensions []string
		if extension != "" && len(extension) <= 10 { // Limit extension length
			extensions = []string{extension}
		}

		options := &GetSourceListOptions{
			RespectGitignore: respectGitignore,
			IncludeHidden:    includeHidden,
			Extensions:       extensions,
		}

		// Test that GetSourceList doesn't panic with any input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetSourceList panicked with dir=%q, options=%+v: %v", testDir, options, r)
			}
		}()

		// Call GetSourceList - should not panic
		files, err := GetSourceList(testDir, options)

		// Basic invariant checks
		if err == nil {
			// If no error, files should not be nil
			if files == nil {
				t.Errorf("GetSourceList returned nil files slice without error for dir=%q", testDir)
			}

			// Limit the number of files we check for performance
			maxFilesToCheck := 100
			filesToCheck := files
			if len(files) > maxFilesToCheck {
				filesToCheck = files[:maxFilesToCheck]
			}

			// All returned files should be valid paths
			for _, file := range filesToCheck {
				if file == "" {
					t.Errorf("GetSourceList returned empty file path for dir=%q", testDir)
				}

				// If extension filter is specified, check that files match
				if len(extensions) > 0 {
					fileExt := filepath.Ext(file)
					found := false
					for _, ext := range extensions {
						if fileExt == ext {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("GetSourceList returned file %q with extension %q, but only %v extensions were requested",
							file, fileExt, extensions)
					}
				}
			}
		}
	})
}

// FuzzGetSourceListWithSpecialChars tests GetSourceList with various special characters
// in directory paths and file extensions.
func FuzzGetSourceListWithSpecialChars(f *testing.F) {
	// Add seed data with special characters
	f.Add("test\x00dir", ".go\x00")
	f.Add("dir\nwith\nnewlines", ".txt\n")
	f.Add("dir\twith\ttabs", ".js\t")
	f.Add("dir with spaces", ".py ")
	f.Add("dir-with-unicode-测试", ".测试")
	f.Add("dir/with/../../paths", "../.go")
	f.Add("dir\\with\\backslashes", ".exe\\")
	f.Add("very"+strings.Repeat("long", 100)+"path", ".verylongextension")

	f.Fuzz(func(t *testing.T, dir string, extension string) {
		// Create options with potentially problematic extension
		var extensions []string
		if extension != "" {
			extensions = []string{extension}
		}

		options := &GetSourceListOptions{
			RespectGitignore: true,
			IncludeHidden:    false,
			Extensions:       extensions,
		}

		// Test that function doesn't panic with special characters
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetSourceList panicked with special chars dir=%q, ext=%q: %v", dir, extension, r)
			}
		}()

		// Call function - main goal is to ensure no panic
		_, err := GetSourceList(dir, options)

		// We don't check the error here since many special character paths
		// are expected to fail - we just want to ensure no panic occurs
		_ = err
	})
}

// FuzzGetSourceListOptions tests various combinations of GetSourceListOptions
// to ensure robustness with different option configurations.
func FuzzGetSourceListOptions(f *testing.F) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fuzz_test")
	if err != nil {
		f.Fatal("Failed to create temp dir for fuzz test:", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files
	testFiles := []string{"test.go", "test.txt", ".hidden", "dir/nested.js"}
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		os.MkdirAll(filepath.Dir(filePath), 0755)
		os.WriteFile(filePath, []byte("test"), 0644)
	}

	// Add seed data for options fuzzing
	f.Add(true, false, 1, "/nonexistent/.gitignore")
	f.Add(false, true, 0, "")
	f.Add(true, true, 2, tempDir+"/.gitignore")
	f.Add(false, false, 5, "/dev/null")

	f.Fuzz(func(t *testing.T, respectGitignore bool, includeHidden bool, numExtensions int, gitignorePath string) {
		// Generate extensions based on numExtensions
		var extensions []string
		extOptions := []string{".go", ".txt", ".js", ".py", ".java", ".cpp", ".h", ".md", ".json", ".xml"}
		for i := 0; i < numExtensions && i < len(extOptions); i++ {
			extensions = append(extensions, extOptions[i])
		}

		options := &GetSourceListOptions{
			RespectGitignore:  respectGitignore,
			IncludeHidden:     includeHidden,
			Extensions:        extensions,
			GitignoreFilePath: gitignorePath,
		}

		// Test that function doesn't panic with various option combinations
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("GetSourceList panicked with options=%+v: %v", options, r)
			}
		}()

		files, err := GetSourceList(tempDir, options)

		// Basic checks when no error occurs
		if err == nil {
			if files == nil {
				t.Errorf("GetSourceList returned nil files slice without error")
			}

			// Check extension filtering
			for _, file := range files {
				if len(extensions) > 0 {
					ext := filepath.Ext(file)
					found := false
					for _, validExt := range extensions {
						if ext == validExt {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("File %q has extension %q but only %v were requested", file, ext, extensions)
					}
				}

				// Check hidden file handling
				if !includeHidden {
					basename := filepath.Base(file)
					if strings.HasPrefix(basename, ".") && basename != ".." && basename != "." {
						t.Errorf("Hidden file %q returned when IncludeHidden=false", file)
					}
				}
			}
		}
	})
}

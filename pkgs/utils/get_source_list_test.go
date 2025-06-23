// nolint:testpackage
package utils

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// setupTestEnvironment creates a temporary directory with test files and returns the temp dir path.
func setupTestEnvironment(t *testing.T) string {
	t.Helper()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_source_list")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

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
		filePath := filepath.Join(tempDir, file)
		// Create directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", file, err)
		}
		// Create file
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// Create .gitignore file
	gitignoreContent := `node_modules/
build/
*.bin
.hidden
`
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	return tempDir
}

// getRelativeFiles converts absolute file paths to relative paths for easier testing.
func getRelativeFiles(t *testing.T, tempDir string, files []string, skipGitignore bool) []string {
	t.Helper()

	relativeFiles := make([]string, 0, len(files))
	for _, file := range files {
		relPath, err := filepath.Rel(tempDir, file)
		if err != nil {
			t.Fatalf("Failed to get relative path for %s: %v", file, err)
		}
		// Skip gitignore files if requested
		if skipGitignore && strings.HasSuffix(relPath, ".gitignore") {
			continue
		}
		relativeFiles = append(relativeFiles, filepath.ToSlash(relPath))
	}
	sort.Strings(relativeFiles)
	return relativeFiles
}

// TestGetSourceList_WithGitignore tests file discovery with gitignore rules enabled.
func TestGetSourceList_WithGitignore(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	defer os.RemoveAll(tempDir)

	options := &GetSourceListOptions{
		RespectGitignore: true,
		IncludeHidden:    false,
	}
	files, err := GetSourceList(tempDir, options)
	if err != nil {
		t.Fatalf("GetSourceList failed: %v", err)
	}

	relativeFiles := getRelativeFiles(t, tempDir, files, false)

	expected := []string{
		"dir1/file3.go",
		"dir1/file4.txt",
		"dir2/file5.js",
		"file1.go",
		"file2.txt",
	}
	sort.Strings(expected)

	if len(relativeFiles) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relativeFiles))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relativeFiles)
		return
	}

	for i, expectedFile := range expected {
		if relativeFiles[i] != expectedFile {
			t.Errorf("Expected file %s, got %s", expectedFile, relativeFiles[i])
		}
	}
}

// TestGetSourceList_WithoutGitignore tests file discovery with gitignore rules disabled.
func TestGetSourceList_WithoutGitignore(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	defer os.RemoveAll(tempDir)

	options := &GetSourceListOptions{
		RespectGitignore: false,
		IncludeHidden:    false,
	}
	files, err := GetSourceList(tempDir, options)
	if err != nil {
		t.Fatalf("GetSourceList failed: %v", err)
	}

	relativeFiles := getRelativeFiles(t, tempDir, files, true)

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

	if len(relativeFiles) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relativeFiles))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relativeFiles)
		return
	}

	for i, expectedFile := range expected {
		if relativeFiles[i] != expectedFile {
			t.Errorf("Expected file %s, got %s", expectedFile, relativeFiles[i])
		}
	}
}

// TestGetSourceList_WithExtensionFilter tests file discovery with extension filtering.
func TestGetSourceList_WithExtensionFilter(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	defer os.RemoveAll(tempDir)

	options := &GetSourceListOptions{
		RespectGitignore: true,
		IncludeHidden:    false,
		Extensions:       []string{".go"},
	}
	files, err := GetSourceList(tempDir, options)
	if err != nil {
		t.Fatalf("GetSourceList failed: %v", err)
	}

	relativeFiles := getRelativeFiles(t, tempDir, files, false)

	expected := []string{
		"dir1/file3.go",
		"file1.go",
	}
	sort.Strings(expected)

	if len(relativeFiles) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relativeFiles))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relativeFiles)
		return
	}

	for i, expectedFile := range expected {
		if relativeFiles[i] != expectedFile {
			t.Errorf("Expected file %s, got %s", expectedFile, relativeFiles[i])
		}
	}
}

// TestGetSourceList_WithHiddenFiles tests file discovery with hidden files included.
func TestGetSourceList_WithHiddenFiles(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	defer os.RemoveAll(tempDir)

	options := &GetSourceListOptions{
		RespectGitignore: false,
		IncludeHidden:    true,
	}
	files, err := GetSourceList(tempDir, options)
	if err != nil {
		t.Fatalf("GetSourceList failed: %v", err)
	}

	// Check if hidden files are included
	var hasHiddenFile bool
	for _, file := range files {
		if strings.Contains(file, ".hidden") {
			hasHiddenFile = true
			break
		}
	}

	if !hasHiddenFile {
		t.Error("Expected hidden files to be included when IncludeHidden is true")
	}
}

// TestGetSourceList_WithCustomGitignoreFilePath tests file discovery with custom gitignore file path.
func TestGetSourceList_WithCustomGitignoreFilePath(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	defer os.RemoveAll(tempDir)

	// Create a custom gitignore file with different rules
	customGitignoreContent := `*.txt
dir2/
`
	customGitignorePath := filepath.Join(tempDir, "custom.gitignore")
	if err := os.WriteFile(customGitignorePath, []byte(customGitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create custom .gitignore: %v", err)
	}

	options := &GetSourceListOptions{
		RespectGitignore:  true,
		IncludeHidden:     false,
		GitignoreFilePath: customGitignorePath,
	}
	files, err := GetSourceList(tempDir, options)
	if err != nil {
		t.Fatalf("GetSourceList failed: %v", err)
	}

	relativeFiles := getRelativeFiles(t, tempDir, files, true)

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

	if len(relativeFiles) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relativeFiles))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relativeFiles)
		return
	}

	for i, expectedFile := range expected {
		if relativeFiles[i] != expectedFile {
			t.Errorf("Expected file %s, got %s", expectedFile, relativeFiles[i])
		}
	}
}

// TestGetSourceList_WithNonExistentGitignoreFilePath tests graceful handling of non-existent gitignore file.
func TestGetSourceList_WithNonExistentGitignoreFilePath(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	defer os.RemoveAll(tempDir)

	options := &GetSourceListOptions{
		RespectGitignore:  true,
		IncludeHidden:     false,
		GitignoreFilePath: "/non/existent/path/.gitignore",
	}

	// Should not fail, just ignore the gitignore rules
	files, err := GetSourceList(tempDir, options)
	if err != nil {
		t.Fatalf("GetSourceList failed: %v", err)
	}

	relativeFiles := getRelativeFiles(t, tempDir, files, true)

	// Should have at least the non-hidden files
	if len(relativeFiles) < 6 { // file1.go, dir1/file3.go, dir1/file4.txt, dir2/file5.js, node_modules/package.json, build/output.bin
		t.Errorf("Expected at least 6 files when gitignore file doesn't exist, got %d: %v", len(relativeFiles), relativeFiles)
	}
}

// TestGetSourceList_WithNilOptions tests that GetSourceList works with nil options (using defaults).
func TestGetSourceList_WithNilOptions(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	defer os.RemoveAll(tempDir)

	// Test with nil options - should use defaults
	files, err := GetSourceList(tempDir, nil)
	if err != nil {
		t.Fatalf("GetSourceList failed: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected to find files when using default options")
	}

	relativeFiles := getRelativeFiles(t, tempDir, files, false)

	// With default options (RespectGitignore=true, IncludeHidden=false),
	// should behave same as TestGetSourceList_WithGitignore
	expected := []string{
		"dir1/file3.go",
		"dir1/file4.txt",
		"dir2/file5.js",
		"file1.go",
		"file2.txt",
	}
	sort.Strings(expected)

	if len(relativeFiles) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relativeFiles))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relativeFiles)
	}
}

// TestGetSourceList_WithMultipleExtensions tests file discovery with multiple extension filters.
func TestGetSourceList_WithMultipleExtensions(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	defer os.RemoveAll(tempDir)

	options := &GetSourceListOptions{
		RespectGitignore: true,
		IncludeHidden:    false,
		Extensions:       []string{".go", ".js"},
	}
	files, err := GetSourceList(tempDir, options)
	if err != nil {
		t.Fatalf("GetSourceList failed: %v", err)
	}

	relativeFiles := getRelativeFiles(t, tempDir, files, false)

	expected := []string{
		"dir1/file3.go",
		"dir2/file5.js",
		"file1.go",
	}
	sort.Strings(expected)

	if len(relativeFiles) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relativeFiles))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relativeFiles)
		return
	}

	for i, expectedFile := range expected {
		if relativeFiles[i] != expectedFile {
			t.Errorf("Expected file %s, got %s", expectedFile, relativeFiles[i])
		}
	}
}

// TestGetSourceList_WithEmptyDirectory tests behavior with an empty directory.
func TestGetSourceList_WithEmptyDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_empty_dir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	options := &GetSourceListOptions{
		RespectGitignore: true,
		IncludeHidden:    false,
	}
	files, err := GetSourceList(tempDir, options)
	if err != nil {
		t.Fatalf("GetSourceList failed: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d: %v", len(files), files)
	}
}

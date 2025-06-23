// nolint:testpackage
package utils

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestGetSourceList(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_source_list")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

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

	t.Run("WithGitignore", func(t *testing.T) {
		options := &GetSourceListOptions{
			RespectGitignore: true,
			IncludeHidden:    false,
		}
		files, err := GetSourceList(tempDir, options)
		if err != nil {
			t.Fatalf("GetSourceList failed: %v", err)
		}

		// Convert to relative paths for easier testing
		var relativeFiles []string
		for _, file := range files {
			relPath, _ := filepath.Rel(tempDir, file)
			relativeFiles = append(relativeFiles, filepath.ToSlash(relPath))
		}
		sort.Strings(relativeFiles)

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
	})

	t.Run("WithoutGitignore", func(t *testing.T) {
		options := &GetSourceListOptions{
			RespectGitignore: false,
			IncludeHidden:    false,
		}
		files, err := GetSourceList(tempDir, options)
		if err != nil {
			t.Fatalf("GetSourceList failed: %v", err)
		}

		// Convert to relative paths for easier testing
		var relativeFiles []string
		for _, file := range files {
			relPath, _ := filepath.Rel(tempDir, file)
			// Skip .gitignore file for this test
			if filepath.Base(relPath) == ".gitignore" {
				continue
			}
			relativeFiles = append(relativeFiles, filepath.ToSlash(relPath))
		}
		sort.Strings(relativeFiles)

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
	})

	t.Run("WithExtensionFilter", func(t *testing.T) {
		options := &GetSourceListOptions{
			RespectGitignore: true,
			IncludeHidden:    false,
			Extensions:       []string{".go"},
		}
		files, err := GetSourceList(tempDir, options)
		if err != nil {
			t.Fatalf("GetSourceList failed: %v", err)
		}

		// Convert to relative paths for easier testing
		var relativeFiles []string
		for _, file := range files {
			relPath, _ := filepath.Rel(tempDir, file)
			relativeFiles = append(relativeFiles, filepath.ToSlash(relPath))
		}
		sort.Strings(relativeFiles)

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
	})

	t.Run("WithHiddenFiles", func(t *testing.T) {
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
	})
}

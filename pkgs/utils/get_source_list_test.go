package utils // nolint:testpackage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetSourceList(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"main.go",
		"README.md",
		"config.json",
		"src/app.go",
		"src/utils.go",
		"tests/test.go",
		".env",
		".gitignore",
		"node_modules/package.json",
		"dist/bundle.js",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		// Create file
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create .gitignore file
	gitignoreContent := `node_modules/
dist/
*.log
.env
`
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Test with gitignore rules
	t.Run("WithGitignore", func(t *testing.T) {
		options := &GetSourceListOptions{
			RespectGitignore: true,
			IncludeHidden:    false,
		}

		files, err := GetSourceList(tempDir, options)
		if err != nil {
			t.Fatalf("GetSourceList failed: %v", err)
		}

		// Convert to relative paths for easier comparison
		relFiles := make([]string, 0, len(files))
		for _, file := range files {
			rel, _ := filepath.Rel(tempDir, file)
			relFiles = append(relFiles, rel)
		}

		// Check that gitignored files are excluded
		for _, file := range relFiles {
			if filepath.Base(file) == ".env" {
				t.Errorf("Expected .env to be gitignored, but found: %s", file)
			}
			if filepath.Dir(file) == "node_modules" || strings.HasPrefix(file, "node_modules/") {
				t.Errorf("Expected node_modules to be gitignored, but found: %s", file)
			}
			if filepath.Dir(file) == "dist" || strings.HasPrefix(file, "dist/") {
				t.Errorf("Expected dist to be gitignored, but found: %s", file)
			}
		}

		// Check that non-gitignored files are included
		expectedFiles := []string{"main.go", "README.md", "config.json", "src/app.go", "src/utils.go", "tests/test.go"}
		for _, expected := range expectedFiles {
			found := false
			for _, file := range relFiles {
				if file == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected file %s to be included, but not found", expected)
			}
		}
	})

	// Test without gitignore rules
	t.Run("WithoutGitignore", func(t *testing.T) {
		options := &GetSourceListOptions{
			RespectGitignore: false,
			IncludeHidden:    true,
		}

		files, err := GetSourceList(tempDir, options)
		if err != nil {
			t.Fatalf("GetSourceList failed: %v", err)
		}

		// Should include all files (except .git which we always skip)
		if len(files) < len(testFiles) {
			t.Errorf("Expected at least %d files, got %d", len(testFiles), len(files))
		}
	})

	// Test with file extension filter
	t.Run("WithExtensionFilter", func(t *testing.T) {
		options := &GetSourceListOptions{
			RespectGitignore: false,
			IncludeHidden:    false,
			Extensions:       []string{".go"},
		}

		files, err := GetSourceList(tempDir, options)
		if err != nil {
			t.Fatalf("GetSourceList failed: %v", err)
		}

		// Should only include .go files
		for _, file := range files {
			if filepath.Ext(file) != ".go" {
				t.Errorf("Expected only .go files, but found: %s", file)
			}
		}
	})
}

// TestLoadGitignoreRules tests the internal loadGitignoreRules function by creating
// a public wrapper since it's not exported from the utils package.
func TestLoadGitignoreRules(t *testing.T) {
	tempDir := t.TempDir()

	gitignoreContent := `# This is a comment
node_modules/
*.log

dist/
.env
`
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Test the functionality through GetSourceList since loadGitignoreRules is not exported
	options := &GetSourceListOptions{
		RespectGitignore: true,
		IncludeHidden:    false,
	}

	// Create some test files
	testFiles := []string{"test.log", "main.go", "node_modules/pkg.json", "dist/bundle.js", ".env"}
	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	files, err := GetSourceList(tempDir, options)
	if err != nil {
		t.Fatalf("GetSourceList failed: %v", err)
	}

	// Convert to relative paths
	relFiles := make([]string, 0, len(files))
	for _, file := range files {
		rel, _ := filepath.Rel(tempDir, file)
		relFiles = append(relFiles, rel)
	}

	// Should only contain main.go (others should be gitignored)
	expectedCount := 1
	if len(relFiles) != expectedCount {
		t.Errorf("Expected %d files after gitignore filtering, got %d: %v", expectedCount, len(relFiles), relFiles)
	}
}

// TestMatchesGitignoreRule tests the internal matchesGitignoreRule function indirectly.
func TestMatchesGitignoreRule(t *testing.T) {
	tests := []struct {
		name         string
		gitignore    string
		testFile     string
		shouldIgnore bool
	}{
		{"Directory rule", "node_modules/", "node_modules/package.json", true},
		{"Wildcard rule", "*.log", "test.log", true},
		{"Nested wildcard", "*.log", "src/test.log", true},
		{"Go files", "*.go", "main.go", true},
		{"No match", "*.log", "README.md", false},
		{"Exact match", ".env", ".env", true},
		{"Nested exact", ".env", "config/.env", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a unique temp directory for each subtest
			tempDir := t.TempDir()

			// Create .gitignore with specific rule
			gitignorePath := filepath.Join(tempDir, ".gitignore")
			if err := os.WriteFile(gitignorePath, []byte(test.gitignore), 0644); err != nil {
				t.Fatalf("Failed to create .gitignore: %v", err)
			}

			// Create test file
			fullPath := filepath.Join(tempDir, test.testFile)
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create directory %s: %v", dir, err)
			}
			if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", fullPath, err)
			}

			// Test through GetSourceList
			options := &GetSourceListOptions{
				RespectGitignore: true,
				IncludeHidden:    true,
			}

			files, err := GetSourceList(tempDir, options)
			if err != nil {
				t.Fatalf("GetSourceList failed: %v", err)
			}

			// Check if file is included or excluded
			found := false
			for _, file := range files {
				rel, _ := filepath.Rel(tempDir, file)
				if rel == test.testFile {
					found = true
					break
				}
			}

			if test.shouldIgnore && found {
				t.Errorf("Expected file %s to be ignored by rule %s, but it was included", test.testFile, test.gitignore)
			} else if !test.shouldIgnore && !found {
				t.Errorf("Expected file %s to not be ignored by rule %s, but it was excluded", test.testFile, test.gitignore)
			}
		})
	}
}

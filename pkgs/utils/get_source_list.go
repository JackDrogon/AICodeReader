package utils

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// GetSourceListOptions represents options for GetSourceList function.
type GetSourceListOptions struct {
	// RespectGitignore whether to respect .gitignore rules
	RespectGitignore bool
	// IncludeHidden whether to include hidden files (starting with .)
	IncludeHidden bool
	// Extensions list of file extensions to include (e.g., []string{".go", ".js"})
	// If empty, all files are included
	Extensions []string
}

// GetSourceList returns a list of file paths in the given directory,
// optionally respecting .gitignore rules.
func GetSourceList(dir string, options *GetSourceListOptions) ([]string, error) {
	if options == nil {
		options = &GetSourceListOptions{
			RespectGitignore: true,
			IncludeHidden:    false,
		}
	}

	var files []string
	var gitignoreRules []string

	// Load .gitignore rules if requested
	if options.RespectGitignore {
		gitignoreRules = loadGitignoreRules(dir)
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip if it's a directory
		if d.IsDir() {
			// Skip .git directory
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files if not included
		if !options.IncludeHidden && strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Check file extension if specified
		if len(options.Extensions) > 0 {
			ext := filepath.Ext(path)
			if !contains(options.Extensions, ext) {
				return nil
			}
		}

		// Check against gitignore rules
		if options.RespectGitignore && isIgnored(path, dir, gitignoreRules) {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// loadGitignoreRules loads gitignore rules from .gitignore file in the directory.
func loadGitignoreRules(dir string) []string {
	var rules []string
	gitignorePath := filepath.Join(dir, ".gitignore")

	// #nosec G304 - gitignore path is constructed from trusted input
	file, err := os.Open(gitignorePath)
	if err != nil {
		// No .gitignore file found, return empty rules
		return rules
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rules = append(rules, line)
	}

	return rules
}

// isIgnored checks if a file path should be ignored based on gitignore rules.
func isIgnored(filePath, rootDir string, rules []string) bool {
	// Convert absolute path to relative path from root directory
	relPath, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		return false
	}

	// Normalize path separators for cross-platform compatibility
	relPath = filepath.ToSlash(relPath)

	for _, rule := range rules {
		if matchesGitignoreRule(relPath, rule) {
			return true
		}
	}

	return false
}

// matchesGitignoreRule checks if a file path matches a gitignore rule.
// This is a simplified implementation of gitignore pattern matching.
func matchesGitignoreRule(filePath, rule string) bool {
	// Handle negation (rules starting with !)
	if strings.HasPrefix(rule, "!") {
		return false // Simplified: don't handle negation for now
	}

	// Handle directory patterns (ending with /)
	if strings.HasSuffix(rule, "/") {
		// Remove trailing slash for directory matching
		dirRule := strings.TrimSuffix(rule, "/")
		// Check if any directory in the path matches
		pathParts := strings.Split(filePath, "/")
		for i := range pathParts {
			dirPath := strings.Join(pathParts[:i+1], "/")
			if dirPath == dirRule || filepath.Base(dirPath) == dirRule {
				return true
			}
		}
		return false
	}

	// Simple pattern matching
	if rule == filePath {
		return true
	}

	// Handle wildcard patterns
	if strings.Contains(rule, "*") {
		// Check full path match
		matched, _ := filepath.Match(rule, filePath)
		if matched {
			return true
		}

		// Check if filename matches
		matched, _ = filepath.Match(rule, filepath.Base(filePath))
		if matched {
			return true
		}

		// Check if any parent directory matches
		dir := filepath.Dir(filePath)
		for dir != "." && dir != "/" {
			matched, _ := filepath.Match(rule, filepath.Base(dir))
			if matched {
				return true
			}
			dir = filepath.Dir(dir)
		}
	}

	// Check if the rule matches any part of the path
	pathParts := strings.Split(filePath, "/")
	for _, part := range pathParts {
		if rule == part {
			return true
		}
		if strings.Contains(rule, "*") {
			matched, _ := filepath.Match(rule, part)
			if matched {
				return true
			}
		}
	}

	return false
}

// contains checks if a slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

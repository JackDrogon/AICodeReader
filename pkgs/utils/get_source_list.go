package utils

import (
	"io/fs"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
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
	// GitignoreFilePath custom path to .gitignore file
	// If empty and RespectGitignore is true, will use .gitignore in the target directory
	GitignoreFilePath string
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

	var gitIgnore *ignore.GitIgnore
	var err error

	// Load .gitignore rules if requested
	if options.RespectGitignore {
		var gitignorePath string

		if options.GitignoreFilePath != "" {
			// Use custom gitignore file path
			gitignorePath = options.GitignoreFilePath
		} else {
			// Use default .gitignore in the target directory
			gitignorePath = filepath.Join(dir, ".gitignore")
		}

		gitIgnore, err = ignore.CompileIgnoreFile(gitignorePath)
		if err != nil {
			// If .gitignore file doesn't exist or can't be read, create an empty GitIgnore
			gitIgnore = ignore.CompileIgnoreLines()
		}
	}

	files := make([]string, 0, 100) // Preallocate slice for better performance

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
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

		// Check against gitignore rules if enabled
		if gitIgnore != nil {
			// Convert to relative path from the directory
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			// Normalize path separators for cross-platform compatibility
			relPath = filepath.ToSlash(relPath)

			if gitIgnore.MatchesPath(relPath) {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	return files, err
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

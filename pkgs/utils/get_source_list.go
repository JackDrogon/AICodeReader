package utils

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// GetSourceListOptions represents configuration options for the GetSourceList function.
// It provides fine-grained control over file discovery behavior.
type GetSourceListOptions struct {
	// RespectGitignore determines whether to respect .gitignore rules during file discovery.
	// When true, files matching .gitignore patterns will be excluded from the results.
	// When false, all files (subject to other filters) will be included regardless of .gitignore rules.
	RespectGitignore bool

	// IncludeHidden determines whether to include hidden files (files starting with '.').
	// When true, hidden files like .env, .config, etc. will be included in the results.
	// When false, hidden files will be filtered out.
	// Note: .git directories are always excluded regardless of this setting.
	IncludeHidden bool

	// Extensions specifies a list of file extensions to include in the results.
	// Only files with extensions matching this list will be returned.
	// Extensions should include the dot prefix (e.g., []string{".go", ".js", ".py"}).
	// If empty or nil, all file extensions will be included (subject to other filters).
	Extensions []string

	// GitignoreFilePath specifies a custom path to a .gitignore file.
	// When RespectGitignore is true:
	//   - If GitignoreFilePath is empty: uses .gitignore in the target directory (dir parameter)
	//   - If GitignoreFilePath is set: uses the specified .gitignore file path
	//   - If the specified file doesn't exist: silently continues without gitignore rules
	// When RespectGitignore is false: this field is ignored.
	GitignoreFilePath string
}

// GetSourceList recursively scans a directory and returns a list of file paths
// that match the specified criteria. It provides flexible filtering options
// including gitignore support, file extension filtering, and hidden file handling.
//
// Parameters:
//   - dir: The root directory path to scan. Can be absolute or relative path.
//   - options: Configuration options for filtering behavior. If nil, uses default settings:
//     RespectGitignore=true, IncludeHidden=false, Extensions=nil, GitignoreFilePath=""
//
// Returns:
//   - []string: A slice of file paths that match the specified criteria.
//     Paths are returned as provided by filepath.WalkDir (absolute if dir is absolute).
//   - error: An error if directory traversal fails or other filesystem errors occur.
//     Gitignore file read errors are handled gracefully and don't cause function failure.
//
// Behavior:
//   - Always excludes .git directories from traversal for performance
//   - Respects gitignore rules when RespectGitignore=true
//   - Filters by file extensions when Extensions is specified
//   - Filters hidden files when IncludeHidden=false
//   - Returns empty slice (not nil) when no files match criteria
//
// Example usage:
//
//	// Get all Go files respecting .gitignore
//	options := &GetSourceListOptions{
//		RespectGitignore: true,
//		Extensions:       []string{".go"},
//	}
//	files, err := GetSourceList("./src", options)
//
//	// Get all files including hidden ones, no gitignore
//	options := &GetSourceListOptions{
//		RespectGitignore: false,
//		IncludeHidden:    true,
//	}
//	files, err := GetSourceList(".", options)
//
//	// Use custom gitignore file
//	options := &GetSourceListOptions{
//		RespectGitignore:  true,
//		GitignoreFilePath: "/path/to/custom/.gitignore",
//		Extensions:        []string{".js", ".ts"},
//	}
//	files, err := GetSourceList("./project", options)
func GetSourceList(dir string, options *GetSourceListOptions) ([]string, error) {
	if options == nil {
		options = &GetSourceListOptions{
			RespectGitignore: true,
			IncludeHidden:    false,
		}
	}

	var gitIgnore *ignore.GitIgnore
	var extSet map[string]struct{}

	// Precompute extension set if needed
	if len(options.Extensions) > 0 {
		extSet = extensionSet(options.Extensions)
	}

	// Load .gitignore rules if requested
	if options.RespectGitignore {
		gitIgnore = loadGitignore(dir, options.GitignoreFilePath)
	}

	files := make([]string, 0, 512) // Preallocate larger initial capacity

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
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
		if extSet != nil {
			ext := filepath.Ext(path)
			if _, exists := extSet[ext]; !exists {
				return nil
			}
		}

		// Check against gitignore rules if enabled
		if gitIgnore != nil {
			// Convert to relative path from the directory
			relPath, _ := filepath.Rel(dir, path)
			relPath = filepath.ToSlash(relPath) // Normalize to slash separators

			if gitIgnore.MatchesPath(relPath) {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// loadGitignore handles gitignore file loading with error logging.
func loadGitignore(dir, customPath string) *ignore.GitIgnore {
	gitignorePath := customPath
	if gitignorePath == "" {
		gitignorePath = filepath.Join(dir, ".gitignore")
	}

	gitIgnore, err := ignore.CompileIgnoreFile(gitignorePath)
	if err != nil {
		// Log error but continue with empty rules
		log.Printf("WARNING: Could not load gitignore file at %q: %v", gitignorePath, err)
		return ignore.CompileIgnoreLines()
	}
	return gitIgnore
}

// extensionSet creates a map for fast extension lookups.
func extensionSet(extensions []string) map[string]struct{} {
	set := make(map[string]struct{}, len(extensions))
	for _, ext := range extensions {
		set[ext] = struct{}{}
	}
	return set
}

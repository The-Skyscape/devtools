package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// RepoSummarizer is a tool for summarizing repository structure and content
type RepoSummarizer struct {
	name        string
	description string
}

// NewRepoSummarizer creates a new repository summarizer tool
func NewRepoSummarizer() *RepoSummarizer {
	return &RepoSummarizer{
		name:        "repo_summarizer",
		description: "Summarizes repository structure, languages, and key files",
	}
}

// Name returns the tool name
func (r *RepoSummarizer) Name() string {
	return r.name
}

// Description returns the tool description
func (r *RepoSummarizer) Description() string {
	return r.description
}

// Parameters returns the tool's parameter schema
func (r *RepoSummarizer) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"repo_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the repository",
			},
			"max_depth": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum depth to traverse",
				"default":     3,
			},
		},
		"required": []string{"repo_path"},
	}
}

// Execute runs the repository summarizer
func (r *RepoSummarizer) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	repoPath, ok := params["repo_path"].(string)
	if !ok {
		return nil, fmt.Errorf("repo_path parameter is required")
	}
	
	maxDepth := 3
	if md, ok := params["max_depth"].(float64); ok {
		maxDepth = int(md)
	}
	
	summary := map[string]interface{}{
		"path":      repoPath,
		"structure": map[string]interface{}{},
		"languages": map[string]int{},
		"stats":     map[string]interface{}{},
		"key_files": []string{},
	}
	
	// Analyze repository structure
	structure, err := r.analyzeStructure(repoPath, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze structure: %w", err)
	}
	summary["structure"] = structure
	
	// Detect languages
	languages := r.detectLanguages(repoPath)
	summary["languages"] = languages
	
	// Calculate statistics
	stats := r.calculateStats(repoPath)
	summary["stats"] = stats
	
	// Find key files
	keyFiles := r.findKeyFiles(repoPath)
	summary["key_files"] = keyFiles
	
	// Check for README
	readmeContent := r.getReadmeContent(repoPath)
	if readmeContent != "" {
		summary["readme_excerpt"] = r.truncateString(readmeContent, 500)
	}
	
	return summary, nil
}

// analyzeStructure analyzes the directory structure
func (r *RepoSummarizer) analyzeStructure(path string, maxDepth int) (map[string]interface{}, error) {
	structure := make(map[string]interface{})
	
	err := r.walkDir(path, path, structure, 0, maxDepth)
	if err != nil {
		return nil, err
	}
	
	return structure, nil
}

// walkDir recursively walks the directory structure
func (r *RepoSummarizer) walkDir(basePath, currentPath string, structure map[string]interface{}, depth, maxDepth int) error {
	if depth > maxDepth {
		return nil
	}
	
	// Skip common ignored directories
	ignoreDirs := []string{".git", "node_modules", "vendor", ".venv", "__pycache__", "target", "build", "dist"}
	
	entries, err := ioutil.ReadDir(currentPath)
	if err != nil {
		return err
	}
	
	dirs := []string{}
	files := []string{}
	
	for _, entry := range entries {
		name := entry.Name()
		
		// Skip hidden files and ignored directories
		if strings.HasPrefix(name, ".") && name != ".gitignore" && name != ".env.example" {
			continue
		}
		
		if entry.IsDir() {
			// Check if directory should be ignored
			shouldIgnore := false
			for _, ignore := range ignoreDirs {
				if name == ignore {
					shouldIgnore = true
					break
				}
			}
			
			if !shouldIgnore {
				dirs = append(dirs, name)
				
				// Recursively analyze subdirectory
				subPath := filepath.Join(currentPath, name)
				subStructure := make(map[string]interface{})
				err := r.walkDir(basePath, subPath, subStructure, depth+1, maxDepth)
				if err == nil && len(subStructure) > 0 {
					structure[name+"/"] = subStructure
				}
			}
		} else {
			files = append(files, name)
		}
	}
	
	if len(files) > 0 {
		structure["_files"] = files
	}
	
	return nil
}

// detectLanguages detects programming languages in the repository
func (r *RepoSummarizer) detectLanguages(repoPath string) map[string]int {
	languages := make(map[string]int)
	
	// Language mapping by file extension
	langMap := map[string]string{
		".go":     "Go",
		".js":     "JavaScript",
		".jsx":    "JavaScript",
		".ts":     "TypeScript",
		".tsx":    "TypeScript",
		".py":     "Python",
		".rb":     "Ruby",
		".java":   "Java",
		".c":      "C",
		".h":      "C",
		".cpp":    "C++",
		".cc":     "C++",
		".hpp":    "C++",
		".cs":     "C#",
		".php":    "PHP",
		".swift":  "Swift",
		".kt":     "Kotlin",
		".rs":     "Rust",
		".html":   "HTML",
		".css":    "CSS",
		".scss":   "SCSS",
		".sql":    "SQL",
		".sh":     "Shell",
		".yml":    "YAML",
		".yaml":   "YAML",
		".json":   "JSON",
		".xml":    "XML",
		".md":     "Markdown",
	}
	
	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		// Skip common ignored paths
		if strings.Contains(path, "node_modules") || 
		   strings.Contains(path, ".git") || 
		   strings.Contains(path, "vendor") {
			return nil
		}
		
		ext := strings.ToLower(filepath.Ext(path))
		if lang, ok := langMap[ext]; ok {
			languages[lang]++
		}
		
		return nil
	})
	
	return languages
}

// calculateStats calculates repository statistics
func (r *RepoSummarizer) calculateStats(repoPath string) map[string]interface{} {
	stats := map[string]interface{}{
		"total_files":       0,
		"total_directories": 0,
		"total_size_bytes":  int64(0),
		"largest_file":      "",
		"largest_file_size": int64(0),
	}
	
	totalFiles := 0
	totalDirs := 0
	totalSize := int64(0)
	largestFile := ""
	largestSize := int64(0)
	
	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		// Skip .git directory
		if strings.Contains(path, ".git") {
			return nil
		}
		
		if info.IsDir() {
			totalDirs++
		} else {
			totalFiles++
			size := info.Size()
			totalSize += size
			
			if size > largestSize {
				largestSize = size
				largestFile = strings.TrimPrefix(path, repoPath+"/")
			}
		}
		
		return nil
	})
	
	stats["total_files"] = totalFiles
	stats["total_directories"] = totalDirs
	stats["total_size_bytes"] = totalSize
	stats["largest_file"] = largestFile
	stats["largest_file_size"] = largestSize
	
	return stats
}

// findKeyFiles finds important files in the repository
func (r *RepoSummarizer) findKeyFiles(repoPath string) []string {
	keyFilePatterns := []string{
		"README.md", "README.txt", "README",
		"LICENSE", "LICENSE.md", "LICENSE.txt",
		"package.json", "go.mod", "requirements.txt", "Gemfile", "pom.xml",
		"Dockerfile", "docker-compose.yml", "docker-compose.yaml",
		"Makefile", "CMakeLists.txt",
		".gitignore", ".env.example",
		"main.go", "main.py", "main.js", "index.js", "app.py", "app.js",
	}
	
	keyFiles := []string{}
	
	for _, pattern := range keyFilePatterns {
		path := filepath.Join(repoPath, pattern)
		if _, err := os.Stat(path); err == nil {
			keyFiles = append(keyFiles, pattern)
		}
	}
	
	return keyFiles
}

// getReadmeContent reads the README file content
func (r *RepoSummarizer) getReadmeContent(repoPath string) string {
	readmeFiles := []string{"README.md", "README.txt", "README", "readme.md"}
	
	for _, filename := range readmeFiles {
		path := filepath.Join(repoPath, filename)
		content, err := ioutil.ReadFile(path)
		if err == nil {
			return string(content)
		}
	}
	
	return ""
}

// truncateString truncates a string to the specified length
func (r *RepoSummarizer) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
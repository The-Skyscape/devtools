package tools

import "context"

// RepositoryProvider is an interface for repository operations
// This allows tools to work with repositories without depending on specific implementations
type RepositoryProvider interface {
	// ListFiles lists files in a repository
	ListFiles(ctx context.Context, repoID, path, branch string) ([]FileInfo, error)
	
	// GetFile retrieves a file's content
	GetFile(ctx context.Context, repoID, path, branch string) (string, error)
	
	// GetCommits retrieves commit history
	GetCommits(ctx context.Context, repoID, branch string, limit int) ([]CommitInfo, error)
	
	// GetBranches lists repository branches
	GetBranches(ctx context.Context, repoID string) ([]string, string, error)
	
	// SearchCode searches for code patterns
	SearchCode(ctx context.Context, repoID, searchTerm string, limit int) ([]SearchResult, error)
}

// IssueProvider is an interface for issue operations
type IssueProvider interface {
	// ListIssues lists repository issues
	ListIssues(ctx context.Context, repoID, status string) ([]IssueInfo, error)
	
	// GetIssue retrieves a specific issue
	GetIssue(ctx context.Context, repoID, issueID string) (*IssueDetail, error)
	
	// CreateIssue creates a new issue
	CreateIssue(ctx context.Context, repoID, title, description string, labels []string) (string, error)
	
	// UpdateIssue updates an existing issue
	UpdateIssue(ctx context.Context, repoID, issueID string, updates map[string]interface{}) error
	
	// SearchIssues searches for issues
	SearchIssues(ctx context.Context, repoID, query string) ([]IssueInfo, error)
}

// FileInfo represents file information
type FileInfo struct {
	Name string
	Type string // "file" or "directory"
	Size int64
	Path string
}

// CommitInfo represents commit information
type CommitInfo struct {
	Hash    string
	Message string
	Author  string
	Date    string
}

// SearchResult represents a code search result
type SearchResult struct {
	File       string
	LineNumber int
	Content    string
}

// IssueInfo represents basic issue information
type IssueInfo struct {
	ID          string
	Title       string
	Status      string
	CreatedAt   string
	Author      string
	Labels      []string
}

// IssueDetail represents detailed issue information
type IssueDetail struct {
	IssueInfo
	Description string
	Comments    []Comment
	Assignees   []string
}

// Comment represents an issue comment
type Comment struct {
	ID        string
	Author    string
	Content   string
	CreatedAt string
}
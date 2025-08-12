package tools

import (
	"context"
	"fmt"
	"strings"
)

// RepositoryTool provides AI capabilities for repository operations
type RepositoryTool struct {
	name        string
	description string
	provider    RepositoryProvider
}

// NewRepositoryTool creates a new repository tool
func NewRepositoryTool(provider RepositoryProvider) *RepositoryTool {
	return &RepositoryTool{
		name:        "repository_tool",
		description: "Interact with git repositories - browse files, view commits, analyze code structure",
		provider:    provider,
	}
}

// Name returns the tool name
func (r *RepositoryTool) Name() string {
	return r.name
}

// Description returns the tool description
func (r *RepositoryTool) Description() string {
	return r.description
}

// Parameters returns the tool's parameter schema
func (r *RepositoryTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform",
				"enum":        []string{"list_files", "get_file", "get_commits", "get_branches", "get_diff", "search_code"},
			},
			"repo_id": map[string]interface{}{
				"type":        "string",
				"description": "Repository ID",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File or directory path (for file operations)",
			},
			"branch": map[string]interface{}{
				"type":        "string",
				"description": "Branch name (default: main/master)",
			},
			"search_term": map[string]interface{}{
				"type":        "string",
				"description": "Search term for code search",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Limit number of results",
				"default":     20,
			},
		},
		"required": []string{"action", "repo_id"},
	}
}

// Execute runs the repository tool
func (r *RepositoryTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	action, ok := params["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action parameter is required")
	}

	repoID, ok := params["repo_id"].(string)
	if !ok {
		return nil, fmt.Errorf("repo_id parameter is required")
	}

	switch action {
	case "list_files":
		return r.listFiles(repoID, params)
	case "get_file":
		return r.getFile(repoID, params)
	case "get_commits":
		return r.getCommits(repoID, params)
	case "get_branches":
		return r.getBranches(repoID, params)
	case "get_diff":
		return r.getDiff(repoID, params)
	case "search_code":
		return r.searchCode(repoID, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// listFiles lists files in a repository
func (r *RepositoryTool) listFiles(repoID string, params map[string]interface{}) (interface{}, error) {
	path := ""
	if p, ok := params["path"].(string); ok {
		path = p
	}

	branch := "HEAD"
	if b, ok := params["branch"].(string); ok {
		branch = b
	}

	if r.provider == nil {
		// Fallback if no provider is set
		return map[string]interface{}{
			"repo_id": repoID,
			"path":    path,
			"branch":  branch,
			"files":   []map[string]interface{}{},
			"error":   "No repository provider configured",
		}, nil
	}

	files, err := r.provider.ListFiles(context.Background(), repoID, path, branch)
	if err != nil {
		return nil, err
	}

	fileList := make([]map[string]interface{}, len(files))
	for i, f := range files {
		fileList[i] = map[string]interface{}{
			"name": f.Name,
			"type": f.Type,
			"size": f.Size,
			"path": f.Path,
		}
	}

	return map[string]interface{}{
		"repo_id": repoID,
		"path":    path,
		"branch":  branch,
		"files":   fileList,
	}, nil
}

// getFile retrieves a file's content
func (r *RepositoryTool) getFile(repoID string, params map[string]interface{}) (interface{}, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required for get_file action")
	}

	branch := "HEAD"
	if b, ok := params["branch"].(string); ok {
		branch = b
	}

	if r.provider == nil {
		return map[string]interface{}{
			"repo_id": repoID,
			"path":    path,
			"branch":  branch,
			"content": "",
			"error":   "No repository provider configured",
		}, nil
	}

	content, err := r.provider.GetFile(context.Background(), repoID, path, branch)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"repo_id": repoID,
		"path":    path,
		"branch":  branch,
		"content": content,
		"size":    len(content),
		"type":    r.detectFileType(path),
	}, nil
}

// getCommits retrieves commit history
func (r *RepositoryTool) getCommits(repoID string, params map[string]interface{}) (interface{}, error) {
	branch := "HEAD"
	if b, ok := params["branch"].(string); ok {
		branch = b
	}

	limit := 20
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	// This would integrate with the actual repository model
	return map[string]interface{}{
		"repo_id": repoID,
		"branch":  branch,
		"limit":   limit,
		"commits": []map[string]interface{}{
			{
				"hash":    "abc123",
				"message": "Initial commit",
				"author":  "developer",
				"date":    "2024-01-01",
			},
		},
	}, nil
}

// getBranches lists repository branches
func (r *RepositoryTool) getBranches(repoID string, params map[string]interface{}) (interface{}, error) {
	// This would integrate with the actual repository model
	return map[string]interface{}{
		"repo_id": repoID,
		"branches": []string{
			"main",
			"develop",
			"feature/new-feature",
		},
		"default_branch": "main",
	}, nil
}

// getDiff gets diff between commits or branches
func (r *RepositoryTool) getDiff(repoID string, params map[string]interface{}) (interface{}, error) {
	// This would integrate with the actual repository model
	return map[string]interface{}{
		"repo_id": repoID,
		"diff": `diff --git a/file.go b/file.go
index abc123..def456 100644
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
+// New line added
 package main`,
	}, nil
}

// searchCode searches for code in the repository
func (r *RepositoryTool) searchCode(repoID string, params map[string]interface{}) (interface{}, error) {
	searchTerm, ok := params["search_term"].(string)
	if !ok {
		return nil, fmt.Errorf("search_term parameter is required for search_code action")
	}

	limit := 20
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	// This would integrate with the actual repository model
	return map[string]interface{}{
		"repo_id":     repoID,
		"search_term": searchTerm,
		"limit":       limit,
		"results": []map[string]interface{}{
			{
				"file":        "src/main.go",
				"line_number": 42,
				"content":     "func main() { // " + searchTerm,
			},
		},
		"total_matches": 1,
	}, nil
}

// detectFileType detects the file type from extension
func (r *RepositoryTool) detectFileType(path string) string {
	path = strings.ToLower(path)
	
	if strings.HasSuffix(path, ".go") {
		return "go"
	} else if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".jsx") {
		return "javascript"
	} else if strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
		return "typescript"
	} else if strings.HasSuffix(path, ".py") {
		return "python"
	} else if strings.HasSuffix(path, ".java") {
		return "java"
	} else if strings.HasSuffix(path, ".md") {
		return "markdown"
	} else if strings.HasSuffix(path, ".json") {
		return "json"
	} else if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
		return "yaml"
	}
	
	return "text"
}

// IssueTool provides AI capabilities for issue management
type IssueTool struct {
	name        string
	description string
	provider    IssueProvider
}

// NewIssueTool creates a new issue management tool
func NewIssueTool(provider IssueProvider) *IssueTool {
	return &IssueTool{
		name:        "issue_tool",
		description: "Manage repository issues - create, update, search, and analyze issues",
		provider:    provider,
	}
}

// Name returns the tool name
func (i *IssueTool) Name() string {
	return i.name
}

// Description returns the tool description
func (i *IssueTool) Description() string {
	return i.description
}

// Parameters returns the tool's parameter schema
func (i *IssueTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform",
				"enum":        []string{"list", "get", "create", "update", "search", "analyze"},
			},
			"repo_id": map[string]interface{}{
				"type":        "string",
				"description": "Repository ID",
			},
			"issue_id": map[string]interface{}{
				"type":        "string",
				"description": "Issue ID (for get/update actions)",
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Issue title (for create/update)",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Issue description (for create/update)",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Issue status",
				"enum":        []string{"open", "closed", "in_progress"},
			},
			"labels": map[string]interface{}{
				"type":        "array",
				"description": "Issue labels",
				"items":       map[string]interface{}{"type": "string"},
			},
			"search_query": map[string]interface{}{
				"type":        "string",
				"description": "Search query for finding issues",
			},
		},
		"required": []string{"action", "repo_id"},
	}
}

// Execute runs the issue tool
func (i *IssueTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	action, ok := params["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action parameter is required")
	}

	repoID, ok := params["repo_id"].(string)
	if !ok {
		return nil, fmt.Errorf("repo_id parameter is required")
	}

	switch action {
	case "list":
		return i.listIssues(repoID, params)
	case "get":
		return i.getIssue(repoID, params)
	case "create":
		return i.createIssue(repoID, params)
	case "update":
		return i.updateIssue(repoID, params)
	case "search":
		return i.searchIssues(repoID, params)
	case "analyze":
		return i.analyzeIssues(repoID, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// listIssues lists repository issues
func (i *IssueTool) listIssues(repoID string, params map[string]interface{}) (interface{}, error) {
	status := ""
	if s, ok := params["status"].(string); ok {
		status = s
	}

	// This would integrate with the actual issue model
	return map[string]interface{}{
		"repo_id": repoID,
		"status":  status,
		"issues": []map[string]interface{}{
			{
				"id":          "1",
				"title":       "Sample Issue",
				"status":      "open",
				"created_at":  "2024-01-01",
				"author":      "user1",
			},
		},
		"total": 1,
	}, nil
}

// getIssue retrieves a specific issue
func (i *IssueTool) getIssue(repoID string, params map[string]interface{}) (interface{}, error) {
	issueID, ok := params["issue_id"].(string)
	if !ok {
		return nil, fmt.Errorf("issue_id parameter is required for get action")
	}

	// This would integrate with the actual issue model
	return map[string]interface{}{
		"repo_id":     repoID,
		"issue_id":    issueID,
		"title":       "Sample Issue",
		"description": "This is a sample issue description",
		"status":      "open",
		"created_at":  "2024-01-01",
		"author":      "user1",
		"labels":      []string{"bug", "enhancement"},
	}, nil
}

// createIssue creates a new issue
func (i *IssueTool) createIssue(repoID string, params map[string]interface{}) (interface{}, error) {
	title, ok := params["title"].(string)
	if !ok {
		return nil, fmt.Errorf("title parameter is required for create action")
	}

	description := ""
	if d, ok := params["description"].(string); ok {
		description = d
	}

	// This would integrate with the actual issue model
	return map[string]interface{}{
		"repo_id":     repoID,
		"issue_id":    "new-issue-id",
		"title":       title,
		"description": description,
		"status":      "open",
		"created_at":  "2024-01-01",
	}, nil
}

// updateIssue updates an existing issue
func (i *IssueTool) updateIssue(repoID string, params map[string]interface{}) (interface{}, error) {
	issueID, ok := params["issue_id"].(string)
	if !ok {
		return nil, fmt.Errorf("issue_id parameter is required for update action")
	}

	// This would integrate with the actual issue model
	return map[string]interface{}{
		"repo_id":  repoID,
		"issue_id": issueID,
		"updated":  true,
		"message":  "Issue updated successfully",
	}, nil
}

// searchIssues searches for issues
func (i *IssueTool) searchIssues(repoID string, params map[string]interface{}) (interface{}, error) {
	query := ""
	if q, ok := params["search_query"].(string); ok {
		query = q
	}

	// This would integrate with the actual issue model
	return map[string]interface{}{
		"repo_id": repoID,
		"query":   query,
		"results": []map[string]interface{}{
			{
				"id":    "1",
				"title": "Issue matching query: " + query,
				"score": 0.95,
			},
		},
		"total": 1,
	}, nil
}

// analyzeIssues analyzes repository issues
func (i *IssueTool) analyzeIssues(repoID string, params map[string]interface{}) (interface{}, error) {
	// This would integrate with the actual issue model and perform analysis
	return map[string]interface{}{
		"repo_id": repoID,
		"analysis": map[string]interface{}{
			"total_issues":      10,
			"open_issues":       6,
			"closed_issues":     4,
			"avg_resolution_time": "3.5 days",
			"most_active_labels": []string{"bug", "enhancement"},
			"trends": map[string]interface{}{
				"new_issues_this_week": 2,
				"closed_this_week":      3,
			},
		},
	}, nil
}
package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/The-Skyscape/devtools/pkg/database"
)

// DatabaseQuery is a tool for querying databases with AI assistance
type DatabaseQuery struct {
	name        string
	description string
	db          database.Database
}

// NewDatabaseQuery creates a new database query tool
func NewDatabaseQuery(db database.Database) *DatabaseQuery {
	return &DatabaseQuery{
		name:        "database_query",
		description: "Query and analyze database records using natural language",
		db:          db,
	}
}

// Name returns the tool name
func (d *DatabaseQuery) Name() string {
	return d.name
}

// Description returns the tool description
func (d *DatabaseQuery) Description() string {
	return d.description
}

// Parameters returns the tool's parameter schema
func (d *DatabaseQuery) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Natural language query describing what to search for",
			},
			"table": map[string]interface{}{
				"type":        "string",
				"description": "Table name to query (optional, will be inferred if not provided)",
			},
			"filters": map[string]interface{}{
				"type":        "object",
				"description": "Additional filters to apply",
				"additionalProperties": true,
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return",
				"default":     10,
			},
		},
		"required": []string{"query"},
	}
}

// Execute runs the database query tool
func (d *DatabaseQuery) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter is required")
	}

	limit := 10
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	// Convert natural language query to SQL-like conditions
	conditions := d.parseNaturalQuery(query)
	
	// Add any additional filters
	if filters, ok := params["filters"].(map[string]interface{}); ok {
		for key, value := range filters {
			conditions = append(conditions, fmt.Sprintf("%s = '%v'", key, value))
		}
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// If table is specified, use it; otherwise try to infer from query
	table := ""
	if t, ok := params["table"].(string); ok {
		table = t
	} else {
		table = d.inferTableFromQuery(query)
	}

	if table == "" {
		return nil, fmt.Errorf("could not determine table to query")
	}

	// Execute the search using the database Search method
	searchQuery := fmt.Sprintf("%s LIMIT %d", whereClause, limit)
	
	// Use a generic interface to handle the search
	results, err := d.executeSearch(table, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	return map[string]interface{}{
		"table":   table,
		"query":   searchQuery,
		"count":   len(results),
		"results": results,
	}, nil
}

// parseNaturalQuery converts natural language to SQL conditions
func (d *DatabaseQuery) parseNaturalQuery(query string) []string {
	var conditions []string
	
	// Convert common natural language patterns to SQL
	lowerQuery := strings.ToLower(query)
	
	// Handle "created after/before" patterns
	if strings.Contains(lowerQuery, "created after") {
		// Extract date and add condition
		conditions = append(conditions, "CreatedAt > datetime('now', '-7 days')")
	}
	if strings.Contains(lowerQuery, "created today") {
		conditions = append(conditions, "date(CreatedAt) = date('now')")
	}
	
	// Handle "contains" patterns
	if strings.Contains(lowerQuery, "contains") || strings.Contains(lowerQuery, "with") {
		// Extract the term to search for
		parts := strings.Split(lowerQuery, "contains")
		if len(parts) > 1 {
			term := strings.TrimSpace(parts[1])
			term = strings.Trim(term, `"'`)
			if term != "" {
				conditions = append(conditions, fmt.Sprintf("(Name LIKE '%%%s%%' OR Description LIKE '%%%s%%')", term, term))
			}
		}
	}
	
	// Handle "by user" patterns
	if strings.Contains(lowerQuery, "by user") || strings.Contains(lowerQuery, "by ") {
		if idx := strings.Index(lowerQuery, "by "); idx != -1 {
			user := strings.TrimSpace(lowerQuery[idx+3:])
			// Remove any trailing words
			if spaceIdx := strings.Index(user, " "); spaceIdx != -1 {
				user = user[:spaceIdx]
			}
			if user != "" {
				conditions = append(conditions, fmt.Sprintf("UserID = '%s'", user))
			}
		}
	}
	
	// Handle status patterns
	if strings.Contains(lowerQuery, "open") {
		conditions = append(conditions, "Status = 'open'")
	}
	if strings.Contains(lowerQuery, "closed") {
		conditions = append(conditions, "Status = 'closed'")
	}
	if strings.Contains(lowerQuery, "active") {
		conditions = append(conditions, "(Status = 'active' OR Status = 'open')")
	}
	
	// Handle visibility patterns
	if strings.Contains(lowerQuery, "public") {
		conditions = append(conditions, "Visibility = 'public'")
	}
	if strings.Contains(lowerQuery, "private") {
		conditions = append(conditions, "Visibility = 'private'")
	}
	
	return conditions
}

// inferTableFromQuery tries to determine the table from the query
func (d *DatabaseQuery) inferTableFromQuery(query string) string {
	lowerQuery := strings.ToLower(query)
	
	// Common table mappings
	tableKeywords := map[string][]string{
		"repositories": {"repository", "repo", "repos", "project", "projects"},
		"users":        {"user", "users", "member", "members", "developer"},
		"issues":       {"issue", "issues", "bug", "bugs", "ticket", "tickets"},
		"pull_requests": {"pull request", "pr", "prs", "merge request"},
		"comments":     {"comment", "comments", "reply", "replies"},
		"activities":   {"activity", "activities", "action", "actions", "event", "events"},
		"permissions":  {"permission", "permissions", "access", "role", "roles"},
	}
	
	for table, keywords := range tableKeywords {
		for _, keyword := range keywords {
			if strings.Contains(lowerQuery, keyword) {
				return table
			}
		}
	}
	
	return ""
}

// executeSearch performs the actual database search
func (d *DatabaseQuery) executeSearch(table string, searchQuery string) ([]map[string]interface{}, error) {
	// This is a simplified implementation
	// In a real implementation, you would use the actual database Search method
	// based on the model type for the given table
	
	// For now, return a placeholder
	// The actual implementation would need to be integrated with the specific
	// database repository types (e.g., Repositories, Issues, etc.)
	
	return []map[string]interface{}{
		{
			"note": "This would return actual database results",
			"table": table,
			"query": searchQuery,
		},
	}, nil
}

// AnalyzeResults analyzes query results and provides insights
func (d *DatabaseQuery) AnalyzeResults(results []map[string]interface{}) map[string]interface{} {
	analysis := map[string]interface{}{
		"total_count": len(results),
		"summary":     fmt.Sprintf("Found %d results", len(results)),
	}
	
	if len(results) > 0 {
		// Analyze patterns in the results
		// Count by different fields, find trends, etc.
		fieldCounts := make(map[string]map[string]int)
		
		for _, result := range results {
			for key, value := range result {
				if fieldCounts[key] == nil {
					fieldCounts[key] = make(map[string]int)
				}
				valueStr := fmt.Sprintf("%v", value)
				fieldCounts[key][valueStr]++
			}
		}
		
		analysis["field_distributions"] = fieldCounts
	}
	
	return analysis
}
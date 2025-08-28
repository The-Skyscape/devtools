package audit

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/The-Skyscape/devtools/pkg/application"
	"github.com/The-Skyscape/devtools/pkg/database"
)

// Log tracks all admin and sensitive actions for compliance and security
type Log struct {
	application.Model
	
	// Who performed the action
	UserID    string `json:"user_id"`
	UserEmail string `json:"user_email,omitempty"`
	UserName  string `json:"user_name,omitempty"`
	
	// What action was performed
	Action     string `json:"action"`      // e.g., "user.create", "settings.update"
	Category   string `json:"category"`    // e.g., "user_management", "security"
	TargetType string `json:"target_type"` // e.g., "user", "workspace", "setting"
	TargetID   string `json:"target_id"`   // ID of the affected resource
	TargetName string `json:"target_name,omitempty"` // Human-readable name
	
	// Action details
	Method      string          `json:"method,omitempty"`       // HTTP method
	Path        string          `json:"path,omitempty"`         // Request path
	IPAddress   string          `json:"ip_address,omitempty"`   // Client IP
	UserAgent   string          `json:"user_agent,omitempty"`   // Browser info
	OldValue    json.RawMessage `json:"old_value,omitempty"`    // Previous value
	NewValue    json.RawMessage `json:"new_value,omitempty"`    // New value
	Description string          `json:"description,omitempty"`  // Human-readable description
	
	// Result
	Status  string `json:"status"`           // "success", "failed", "unauthorized"
	Error   string `json:"error,omitempty"`  // Error message if failed
	Latency int    `json:"latency,omitempty"` // Response time in ms
	
	// Metadata
	SessionID   string          `json:"session_id,omitempty"`
	RequestID   string          `json:"request_id,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	RiskLevel   string          `json:"risk_level"`  // "low", "medium", "high", "critical"
	RequiresMFA bool            `json:"requires_mfa"`
}

// Table returns the database table name
func (a *Log) Table() string {
	return "audit_logs"
}

// Action categories
const (
	CategoryUserManagement = "user_management"
	CategorySecurity       = "security"
	CategorySettings       = "settings"
	CategoryBilling        = "billing"
	CategoryWorkspace      = "workspace"
	CategorySystem         = "system"
	CategoryAuthentication = "authentication"
	CategoryAuthorization  = "authorization"
)

// Action statuses
const (
	StatusSuccess      = "success"
	StatusFailed       = "failed"
	StatusUnauthorized = "unauthorized"
)

// Risk levels
const (
	RiskLow      = "low"
	RiskMedium   = "medium"
	RiskHigh     = "high"
	RiskCritical = "critical"
)

// Common actions
const (
	ActionUserCreate       = "user.create"
	ActionUserUpdate       = "user.update"
	ActionUserDelete       = "user.delete"
	ActionUserPromote      = "user.promote"
	ActionUserDemote       = "user.demote"
	ActionUserSuspend      = "user.suspend"
	ActionUserReactivate   = "user.reactivate"
	
	ActionAuthLogin        = "auth.login"
	ActionAuthLogout       = "auth.logout"
	ActionAuthFailedLogin  = "auth.failed_login"
	ActionAuthPasswordReset = "auth.password_reset"
	
	ActionSettingsUpdate   = "settings.update"
	ActionSecurityUpdate   = "security.update"
	
	ActionWorkspaceCreate  = "workspace.create"
	ActionWorkspaceDelete  = "workspace.delete"
	ActionWorkspaceStop    = "workspace.stop"
	ActionWorkspaceStart   = "workspace.start"
)

// Logger provides audit logging functionality
type Logger struct {
	Repository *database.Collection[*Log]
}

// NewLogger creates a new audit logger
func NewLogger(db *database.DynamicDB) *Logger {
	return &Logger{
		Repository: database.Manage(db, new(Log)),
	}
}

// LogAction logs an audit event
func (l *Logger) LogAction(log *Log) error {
	// Auto-set timestamp if not set
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	
	// Auto-determine risk level if not set
	if log.RiskLevel == "" {
		log.RiskLevel = l.determineRiskLevel(log.Action, log.Category)
	}
	
	_, err := l.Repository.Insert(log)
	return err
}

// LogHTTPAction logs an HTTP request as an audit event
func (l *Logger) LogHTTPAction(r *http.Request, userID, action string, status string, latency time.Duration) error {
	log := &Log{
		UserID:    userID,
		Action:    action,
		Method:    r.Method,
		Path:      r.URL.Path,
		IPAddress: l.getClientIP(r),
		UserAgent: r.UserAgent(),
		Status:    status,
		Latency:   int(latency.Milliseconds()),
		SessionID: l.getSessionID(r),
		RequestID: l.getRequestID(r),
	}
	
	return l.LogAction(log)
}

// LogSuccess logs a successful action
func (l *Logger) LogSuccess(userID, action, category, targetType, targetID, description string) error {
	return l.LogAction(&Log{
		UserID:      userID,
		Action:      action,
		Category:    category,
		TargetType:  targetType,
		TargetID:    targetID,
		Description: description,
		Status:      StatusSuccess,
	})
}

// LogFailure logs a failed action
func (l *Logger) LogFailure(userID, action, category, targetType, targetID, errorMsg string) error {
	return l.LogAction(&Log{
		UserID:     userID,
		Action:     action,
		Category:   category,
		TargetType: targetType,
		TargetID:   targetID,
		Status:     StatusFailed,
		Error:      errorMsg,
	})
}

// LogUnauthorized logs an unauthorized attempt
func (l *Logger) LogUnauthorized(userID, action, category, targetType, targetID string) error {
	return l.LogAction(&Log{
		UserID:     userID,
		Action:     action,
		Category:   category,
		TargetType: targetType,
		TargetID:   targetID,
		Status:     StatusUnauthorized,
		RiskLevel:  RiskHigh,
	})
}

// determineRiskLevel determines the risk level based on action
func (l *Logger) determineRiskLevel(action, category string) string {
	// Critical risk actions
	criticalActions := []string{
		ActionUserDelete,
		ActionWorkspaceDelete,
		"security.disable_mfa",
		"security.change_password",
		"billing.change_payment",
	}
	
	// High risk actions
	highActions := []string{
		ActionUserPromote,
		ActionUserSuspend,
		ActionSecurityUpdate,
		ActionAuthPasswordReset,
	}
	
	// Medium risk actions
	mediumActions := []string{
		ActionUserUpdate,
		ActionSettingsUpdate,
		ActionWorkspaceStop,
	}
	
	for _, a := range criticalActions {
		if action == a {
			return RiskCritical
		}
	}
	
	for _, a := range highActions {
		if action == a {
			return RiskHigh
		}
	}
	
	for _, a := range mediumActions {
		if action == a {
			return RiskMedium
		}
	}
	
	// Check category-based risk
	if category == CategorySecurity || category == CategoryBilling {
		return RiskHigh
	}
	
	if category == CategoryAuthentication || category == CategoryAuthorization {
		return RiskMedium
	}
	
	return RiskLow
}

// getClientIP gets the client's IP address from the request
func (l *Logger) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// getSessionID extracts session ID from request
func (l *Logger) getSessionID(r *http.Request) string {
	// Try to get from context or cookie
	if cookie, err := r.Cookie("session_id"); err == nil {
		return cookie.Value
	}
	return ""
}

// getRequestID extracts request ID from request
func (l *Logger) getRequestID(r *http.Request) string {
	// Check X-Request-ID header
	return r.Header.Get("X-Request-ID")
}

// Query provides methods for querying audit logs
type Query struct {
	logger *Logger
}

// NewQuery creates a new query builder
func (l *Logger) Query() *Query {
	return &Query{logger: l}
}

// ByUser returns logs for a specific user
func (q *Query) ByUser(userID string) ([]*Log, error) {
	return q.logger.Repository.Search("WHERE UserID = ? ORDER BY CreatedAt DESC", userID)
}

// ByCategory returns logs for a specific category
func (q *Query) ByCategory(category string) ([]*Log, error) {
	return q.logger.Repository.Search("WHERE Category = ? ORDER BY CreatedAt DESC", category)
}

// ByAction returns logs for a specific action
func (q *Query) ByAction(action string) ([]*Log, error) {
	return q.logger.Repository.Search("WHERE Action = ? ORDER BY CreatedAt DESC", action)
}

// ByRiskLevel returns logs with a specific risk level or higher
func (q *Query) ByRiskLevel(minLevel string) ([]*Log, error) {
	return q.logger.Repository.Search("WHERE RiskLevel IN (?, ?, ?) ORDER BY CreatedAt DESC",
		q.getRiskLevelsAbove(minLevel)...)
}

// getRiskLevelsAbove returns risk levels at or above the given level
func (q *Query) getRiskLevelsAbove(level string) []interface{} {
	switch level {
	case RiskCritical:
		return []interface{}{RiskCritical}
	case RiskHigh:
		return []interface{}{RiskHigh, RiskCritical}
	case RiskMedium:
		return []interface{}{RiskMedium, RiskHigh, RiskCritical}
	default:
		return []interface{}{RiskLow, RiskMedium, RiskHigh, RiskCritical}
	}
}

// Recent returns the most recent logs
func (q *Query) Recent(limit int) ([]*Log, error) {
	return q.logger.Repository.Search("ORDER BY CreatedAt DESC LIMIT ?", limit)
}

// Failed returns failed actions
func (q *Query) Failed() ([]*Log, error) {
	return q.logger.Repository.Search("WHERE Status = ? ORDER BY CreatedAt DESC", StatusFailed)
}

// Unauthorized returns unauthorized attempts
func (q *Query) Unauthorized() ([]*Log, error) {
	return q.logger.Repository.Search("WHERE Status = ? ORDER BY CreatedAt DESC", StatusUnauthorized)
}
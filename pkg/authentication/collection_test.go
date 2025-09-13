package authentication

import (
	"database/sql"
	"testing"

	"github.com/The-Skyscape/devtools/pkg/database"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestCollection(t *testing.T) (*Collection, *sql.DB) {
	t.Helper()
	
	// Create in-memory SQLite database
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	
	// Enable foreign keys
	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}
	
	// Create DynamicDB wrapper
	dynamicDB := database.NewDynamicDB(sqlDB)
	
	// Create collection
	collection := Manage(dynamicDB)
	
	// Cleanup function
	t.Cleanup(func() {
		sqlDB.Close()
	})
	
	return collection, sqlDB
}

func TestManage(t *testing.T) {
	collection, _ := setupTestCollection(t)
	
}

func TestGetUser(t *testing.T) {
	collection, _ := setupTestCollection(t)
	
	// Create test user
	testUser := &User{
		Avatar:   "https://robohash.org/test@example.com?set=set4",
		Name:     "Test User",
		Email:    "test@example.com",
		Handle:   "testuser",
		PassHash: []byte("hashedpassword"),
		IsAdmin:  false,
		Role:     "developer",
	}
	
	insertedUser, err := collection.Users.Insert(testUser)
	
	tests := []struct {
		name        string
		identifier  string
		expectError bool
	}{
		{
			name:        "get by ID",
			identifier:  insertedUser.ID,
			expectError: false,
		},
		{
			name:        "get by email",
			identifier:  "test@example.com",
			expectError: false,
		},
		{
			name:        "get by handle",
			identifier:  "testuser",
			expectError: false,
		},
		{
			name:        "nonexistent user",
			identifier:  "nonexistent",
			expectError: true,
		},
		{
			name:        "empty identifier",
			identifier:  "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := collection.GetUser(tt.identifier)
			
			if tt.expectError {
			} else {
			}
		})
	}
}

func TestSignup(t *testing.T) {
	tests := []struct {
		name        string
		userName    string
		email       string
		handle      string
		password    string
		isAdmin     bool
		expectError bool
	}{
		{
			name:        "valid regular user",
			userName:    "John Doe",
			email:       "john@example.com",
			handle:      "johndoe",
			password:    "password123",
			isAdmin:     false,
			expectError: false,
		},
		{
			name:        "valid admin user",
			userName:    "Admin User",
			email:       "admin@example.com",
			handle:      "admin",
			password:    "adminpass",
			isAdmin:     true,
			expectError: false,
		},
		{
			name:        "empty name",
			userName:    "",
			email:       "test@example.com",
			handle:      "test",
			password:    "password",
			isAdmin:     false,
			expectError: false, // bcrypt will handle empty password, name validation should be done at higher level
		},
		{
			name:        "empty email",
			userName:    "Test User",
			email:       "",
			handle:      "test",
			password:    "password",
			isAdmin:     false,
			expectError: false, // Collection doesn't validate email format
		},
		{
			name:        "empty handle",
			userName:    "Test User",
			email:       "test@example.com",
			handle:      "",
			password:    "password",
			isAdmin:     false,
			expectError: false, // Collection doesn't validate handle format
		},
		{
			name:        "empty password",
			userName:    "Test User",
			email:       "test@example.com",
			handle:      "test",
			password:    "",
			isAdmin:     false,
			expectError: true, // bcrypt should fail with empty password
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection, _ := setupTestCollection(t)
			
			user, err := collection.Signup(tt.userName, tt.email, tt.handle, tt.password, tt.isAdmin)
			
			if tt.expectError {
			} else {
				
				// Verify user fields
				
				// Verify role assignment
				expectedRole := "developer"
				if tt.isAdmin {
					expectedRole = "admin"
				}
				
				// Verify avatar URL format
				expectedAvatar := "https://robohash.org/" + tt.email + "?set=set4"
				
				// Verify password hash exists and is not the original password
				
				// Verify password can be verified
				if tt.password != "" {
				}
				
				// Verify user has ID and timestamps
			}
		})
	}
}

func TestSignupUniqueConstraints(t *testing.T) {
	collection, _ := setupTestCollection(t)
	
	// Create first user
	_, err := collection.Signup("User One", "test@example.com", "testuser", "password", false)
	
	tests := []struct {
		name     string
		userName string
		email    string
		handle   string
		password string
	}{
		{
			name:     "duplicate email",
			userName: "User Two",
			email:    "test@example.com", // Same email
			handle:   "differenthandle",
			password: "password",
		},
		{
			name:     "duplicate handle",
			userName: "User Three",
			email:    "different@example.com",
			handle:   "testuser", // Same handle
			password: "password",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := collection.Signup(tt.userName, tt.email, tt.handle, tt.password, false)
		})
	}
}

func TestSignin(t *testing.T) {
	collection, _ := setupTestCollection(t)
	
	// Create test user
	password := "testpassword"
	user, err := collection.Signup("Test User", "test@example.com", "testuser", password, false)
	
	tests := []struct {
		name        string
		identifier  string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "signin with email and correct password",
			identifier:  "test@example.com",
			password:    password,
			expectError: false,
		},
		{
			name:        "signin with handle and correct password",
			identifier:  "testuser",
			password:    password,
			expectError: false,
		},
		{
			name:        "signin with ID and correct password",
			identifier:  user.ID,
			password:    password,
			expectError: false,
		},
		{
			name:        "signin with wrong password",
			identifier:  "test@example.com",
			password:    "wrongpassword",
			expectError: true,
			errorMsg:    "user not found",
		},
		{
			name:        "signin with nonexistent user",
			identifier:  "nonexistent@example.com",
			password:    password,
			expectError: true,
			errorMsg:    "user not found",
		},
		{
			name:        "signin with empty password",
			identifier:  "test@example.com",
			password:    "",
			expectError: true,
			errorMsg:    "user not found",
		},
		{
			name:        "signin with empty identifier",
			identifier:  "",
			password:    password,
			expectError: true,
			errorMsg:    "user not found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultUser, err := collection.Signin(tt.identifier, tt.password)
			
			if tt.expectError {
				if tt.errorMsg != "" {
				}
			} else {
			}
		})
	}
}

func TestSigninConsistentErrorMessages(t *testing.T) {
	collection, _ := setupTestCollection(t)
	
	// Create test user
	_, err := collection.Signup("Test User", "test@example.com", "testuser", "password", false)
	
	// Test that both nonexistent user and wrong password return the same error message
	// This prevents user enumeration attacks
	tests := []struct {
		name       string
		identifier string
		password   string
	}{
		{
			name:       "nonexistent user",
			identifier: "nonexistent@example.com",
			password:   "password",
		},
		{
			name:       "existing user with wrong password",
			identifier: "test@example.com",
			password:   "wrongpassword",
		},
	}
	
	var errorMessages []string
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := collection.Signin(tt.identifier, tt.password)
			errorMessages = append(errorMessages, err.Error())
		})
	}
	
	// Verify both errors have the same message
	if len(errorMessages) == 2 {
			"Error messages should be identical to prevent user enumeration")
	}
}

func TestCollectionIntegration(t *testing.T) {
	collection, _ := setupTestCollection(t)
	
	// Test full workflow: signup -> signin -> get user
	name := "Integration Test User"
	email := "integration@example.com"
	handle := "integration"
	password := "integrationpass"
	
	// 1. Signup
	user, err := collection.Signup(name, email, handle, password, false)
	
	userID := user.ID
	
	// 2. Signin
	signinUser, err := collection.Signin(email, password)
	
	// 3. Get user by different identifiers
	getUserByID, err := collection.GetUser(userID)
	
	getUserByEmail, err := collection.GetUser(email)
	
	getUserByHandle, err := collection.GetUser(handle)
	
	// Verify all retrieved users have the same data
	users := []*User{user, signinUser, getUserByID, getUserByEmail, getUserByHandle}
	for i, u := range users {
	}
}

func TestCollectionSessionsIntegration(t *testing.T) {
	collection, _ := setupTestCollection(t)
	
	// Create a user
	user, err := collection.Signup("Session User", "session@example.com", "sessionuser", "password", false)
	
	// Create a session for the user
	session := &Session{
		UserID: user.ID,
	}
	
	insertedSession, err := collection.Sessions.Insert(session)
	
	// Retrieve the session
	retrievedSession, err := collection.Sessions.Get(insertedSession.ID)
}

func TestCollectionMultipleUsers(t *testing.T) {
	collection, _ := setupTestCollection(t)
	
	// Create multiple users
	users := []struct {
		name    string
		email   string
		handle  string
		isAdmin bool
	}{
		{"User One", "user1@example.com", "user1", false},
		{"User Two", "user2@example.com", "user2", false},
		{"Admin User", "admin@example.com", "admin", true},
	}
	
	var createdUsers []*User
	for _, userData := range users {
		user, err := collection.Signup(userData.name, userData.email, userData.handle, "password", userData.isAdmin)
		createdUsers = append(createdUsers, user)
	}
	
	// Verify all users can be retrieved correctly
	for i, expectedUser := range users {
		createdUser := createdUsers[i]
		
		// Get by email
		user, err := collection.GetUser(expectedUser.email)
		
		// Get by handle
		user, err = collection.GetUser(expectedUser.handle)
		
		// Verify admin status
		
		expectedRole := "developer"
		if expectedUser.isAdmin {
			expectedRole = "admin"
		}
	}
}
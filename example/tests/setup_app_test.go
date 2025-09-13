package tests

// init sets up the test environment automatically when tests are run
func init() {
	// Test database is already initialized in db_test.go
	// Create initial test data using test database
	SetupTestData()
}

// SetupTestData creates standard test data using test database
func SetupTestData() {
	InsertTestDucks()
}

// ClearData removes all test data
func ClearData() {
	ClearTestData()
}

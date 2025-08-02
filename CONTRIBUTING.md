# Contributing to TheSkyscape DevTools

Thank you for your interest in contributing to TheSkyscape DevTools! This document provides guidelines and information for contributors.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:

- **Be respectful** - Treat everyone with respect and kindness
- **Be inclusive** - Welcome newcomers and diverse perspectives  
- **Be collaborative** - Work together towards common goals
- **Be constructive** - Provide helpful feedback and suggestions

## Getting Started

### Prerequisites

- **Go 1.24+** - Latest Go version with generics support
- **Docker** - For container-related functionality
- **Git** - For version control

### Development Setup

1. **Fork and clone the repository**:
   ```bash
   git clone https://github.com/your-username/devtools.git
   cd devtools
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Run tests**:
   ```bash
   go test ./...
   ```

4. **Run the example application**:
   ```bash
   go run ./example
   ```

5. **Build command-line tools**:
   ```bash
   go build ./cmd/create-app
   go build ./cmd/launch-app
   ```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-description
```

### 2. Make Your Changes

- Follow the existing code style and patterns
- Add tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 3. Test Your Changes

```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Test the example application
go run ./example
```

### 4. Commit Your Changes

Follow [Conventional Commits](https://www.conventionalcommits.org/) format:

```bash
git commit -m "feat: add new container management feature"
git commit -m "fix: resolve authentication session timeout"
git commit -m "docs: update API reference for hosting package"
```

### 5. Submit a Pull Request

1. Push your branch to GitHub
2. Create a pull request with a clear title and description
3. Link any related issues
4. Wait for review and address feedback

## Project Structure

```
devtools/
├── pkg/                        # Core packages
│   ├── application/            # Web framework
│   ├── authentication/        # User auth & sessions
│   ├── containers/            # Docker management
│   ├── database/              # Database & ORM
│   ├── hosting/               # Cloud platforms
│   └── coding/                # Git & workspaces
├── cmd/                       # Command-line tools
├── example/                   # Working example app
├── docs/                      # Documentation
└── README.md
```

## Coding Standards

### Go Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small
- Use interfaces for abstraction

### Package Design

- **Single Responsibility** - Each package has one clear purpose
- **Interface-based** - Define behavior through interfaces
- **Composable** - Packages work well together
- **Testable** - Easy to write unit tests

### Error Handling

- Use the `github.com/pkg/errors` package for error wrapping
- Provide meaningful error messages
- Return errors rather than panicking
- Handle errors at appropriate levels

### Example Code Style

```go
// Good: Clear function with error handling
func CreateUser(name, email string) (*User, error) {
    if name == "" {
        return nil, errors.New("name is required")
    }
    
    user := &User{Name: name, Email: email}
    if err := users.Insert(user); err != nil {
        return nil, errors.Wrap(err, "failed to create user")
    }
    
    return user, nil
}

// Good: Interface definition
type Platform interface {
    Launch(*Server) (*Server, error)
    Server(id string) (*Server, error)
}
```

## Testing Guidelines

### Unit Tests

- Test file naming: `*_test.go`
- Test function naming: `TestFunctionName`
- Use table-driven tests for multiple scenarios
- Mock external dependencies

### Integration Tests

- Test real interactions between components
- Use temporary databases and directories
- Clean up resources after tests

### Example Test

```go
func TestUserCreation(t *testing.T) {
    tests := []struct {
        name      string
        inputName string
        inputEmail string
        wantError bool
    }{
        {"valid user", "John Doe", "john@example.com", false},
        {"empty name", "", "john@example.com", true},
        {"empty email", "John Doe", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user, err := CreateUser(tt.inputName, tt.inputEmail)
            if tt.wantError {
                assert.Error(t, err)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
                assert.Equal(t, tt.inputName, user.Name)
            }
        })
    }
}
```

## Documentation

### Code Documentation

- Document all exported functions, types, and variables
- Use clear, concise language
- Include usage examples for complex functions
- Document package purpose in `doc.go` files

### User Documentation

- Update relevant documentation in `docs/`
- Include code examples
- Update the main README if needed
- Add tutorial sections for new features

## Security Guidelines

### Defensive Programming

- **Input Validation** - Validate all user inputs
- **SQL Injection** - Use parameterized queries
- **Path Traversal** - Validate file paths
- **Authentication** - Never expose auth secrets
- **Authorization** - Check permissions properly

### Security Review Checklist

- [ ] No secrets or credentials in code
- [ ] Input validation on all endpoints
- [ ] Proper error handling (no information leaks)
- [ ] Secure defaults for configurations
- [ ] Appropriate file permissions

## Performance Considerations

### Optimization Guidelines

- **Profile before optimizing** - Use Go's profiling tools
- **Benchmark critical paths** - Write benchmark tests
- **Memory efficiency** - Avoid unnecessary allocations
- **Concurrent safety** - Ensure thread-safe operations
- **Database optimization** - Use efficient queries

### Benchmarking

```go
func BenchmarkUserSearch(b *testing.B) {
    // Setup
    db := setupTestDB()
    defer db.Close()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := SearchUsers(db, "test query")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Issue Reporting

### Bug Reports

Include the following information:

1. **Environment** - Go version, OS, Docker version
2. **Steps to reproduce** - Clear, numbered steps
3. **Expected behavior** - What should happen
4. **Actual behavior** - What actually happens
5. **Code samples** - Minimal reproducing example
6. **Logs/Errors** - Full error messages

### Feature Requests

Include the following information:

1. **Use case** - Why is this feature needed?
2. **Proposed solution** - How should it work?
3. **Alternatives** - Other approaches considered
4. **Breaking changes** - Impact on existing code

## Pull Request Guidelines

### Before Submitting

- [ ] Tests pass locally (`go test ./...`)
- [ ] Code follows project style
- [ ] Documentation updated
- [ ] CHANGELOG.md updated (if needed)
- [ ] Commit messages follow convention
- [ ] Branch is up to date with main

### Pull Request Description

- **Clear title** - Summarize the change
- **Description** - Explain what and why
- **Testing** - How was this tested?
- **Documentation** - What docs were updated?
- **Breaking changes** - Any compatibility issues?

### Review Process

1. **Automated checks** - CI/CD pipeline runs
2. **Code review** - Maintainer reviews code
3. **Feedback** - Address review comments
4. **Approval** - Review approved by maintainer
5. **Merge** - Changes merged to main branch

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **Major** (v2.0.0) - Breaking changes
- **Minor** (v1.1.0) - New features, backwards compatible
- **Patch** (v1.0.1) - Bug fixes, backwards compatible

### Release Checklist

- [ ] Update version numbers
- [ ] Update CHANGELOG.md
- [ ] Create release notes
- [ ] Tag release in Git
- [ ] Publish to Go module registry

## Getting Help

### Resources

- **Documentation**: Check the `docs/` directory
- **Examples**: Look at the `example/` application
- **Issues**: Search existing GitHub issues
- **Discussions**: Use GitHub Discussions for questions

### Communication

- **GitHub Issues** - Bug reports and feature requests
- **GitHub Discussions** - Questions and general discussion
- **Pull Requests** - Code contributions and reviews

## Recognition

Contributors are recognized in several ways:

- **CONTRIBUTORS.md** - Listed as project contributors
- **Release notes** - Mentioned in release announcements
- **Git history** - Permanent record of contributions

Thank you for contributing to TheSkyscape DevTools! 🚀
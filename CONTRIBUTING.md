# Contributing to prconflict

Thank you for your interest in contributing to `prconflict`! This document provides guidelines and instructions for contributing to the project.

## 🚀 Quick Start for Contributors

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/your-username/prconflict.git
   cd prconflict
   ```
3. **Set up development environment**:
   ```bash
   go mod download
   export GITHUB_TOKEN=your_token_here
   ```
4. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```
5. **Make your changes** and test them
6. **Submit a pull request**

## 📋 Types of Contributions

We welcome various types of contributions:

### 🐛 Bug Fixes
- Fix issues reported in GitHub Issues
- Add regression tests for the bug
- Update documentation if needed

### ✨ New Features
- Propose new features via GitHub Discussions first
- Implement features with comprehensive tests
- Update documentation and examples

### 📚 Documentation
- Improve README, code comments, or guides
- Add examples and usage patterns
- Fix typos and clarify instructions

### 🧪 Testing
- Add test cases for uncovered scenarios
- Improve E2E test coverage
- Performance and reliability improvements

### 🔧 Infrastructure
- CI/CD improvements
- Build and release automation
- Development tooling enhancements

## 🛠 Development Setup

### Prerequisites

- **Go 1.21+**: [Install Go](https://golang.org/dl/)
- **Git**: Standard Git installation
- **GitHub CLI**: [Install gh CLI](https://cli.github.com/) (recommended)
- **GitHub Token**: Personal Access Token with `repo` scope

### Environment Setup

```bash
# Clone and enter the repository
git clone https://github.com/teddyknox/prconflict.git
cd prconflict

# Install dependencies
go mod download

# Set up GitHub token for testing
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx

# Verify setup
go build ./cmd/prconflict
go test -short ./...
```

### Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/amazing-feature
   ```

2. **Make your changes** with tests
3. **Run the test suite**:
   ```bash
   # Unit tests (fast)
   go test -short ./...
   
   # Integration tests (requires GITHUB_TOKEN)
   go test ./cmd/prconflict -run TestIntegration
   
   # E2E tests (optional, creates GitHub repos)
   ./scripts/run-e2e-tests.sh --test=TestE2E_BasicUnresolvedComments
   ```

4. **Check code quality**:
   ```bash
   go vet ./...
   go fmt ./...
   ```

5. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```

6. **Push and create PR**:
   ```bash
   git push origin feature/amazing-feature
   ```

## 🧪 Testing Guidelines

### Test Types

1. **Unit Tests**: Fast tests for individual functions
   - Location: `*_test.go` files alongside source code
   - Run with: `go test -short ./...`
   - Should not require external dependencies

2. **Integration Tests**: Test real GitHub API interactions
   - Location: `cmd/prconflict/integration_test.go`
   - Run with: `go test ./cmd/prconflict -run TestIntegration`
   - Requires `GITHUB_TOKEN` environment variable

3. **E2E Tests**: Complete workflow tests
   - Location: `cmd/prconflict/e2e_*_test.go`
   - Run with: `./scripts/run-e2e-tests.sh`
   - Creates temporary GitHub repositories

### Writing Tests

#### Unit Test Example
```go
func TestSplitRepo(t *testing.T) {
    tests := []struct {
        input    string
        wantOwner string
        wantRepo  string
        wantOK    bool
    }{
        {"owner/repo", "owner", "repo", true},
        {"invalid", "", "", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            owner, repo, ok := splitRepo(tt.input)
            if owner != tt.wantOwner || repo != tt.wantRepo || ok != tt.wantOK {
                t.Errorf("splitRepo(%q) = (%q, %q, %v), want (%q, %q, %v)",
                    tt.input, owner, repo, ok, tt.wantOwner, tt.wantRepo, tt.wantOK)
            }
        })
    }
}
```

#### E2E Test Example
```go
func TestE2E_YourFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    framework := NewE2ETestFramework(t)
    if err := framework.Setup(t); err != nil {
        t.Fatalf("Framework setup failed: %v", err)
    }
    defer framework.Cleanup()

    scenario := TestScenario{
        Name: "Your feature description",
        // ... scenario definition
    }

    if err := framework.RunScenario(t, scenario); err != nil {
        t.Fatalf("Scenario failed: %v", err)
    }
}
```

### Test Best Practices

- **Fast by default**: Use `testing.Short()` to skip slow tests
- **Deterministic**: Tests should not depend on external state
- **Isolated**: Each test should be independent
- **Descriptive**: Use clear test names and failure messages
- **Comprehensive**: Test both success and error cases

## 📝 Code Guidelines

### Go Style

Follow standard Go conventions:

- **gofmt**: Use `go fmt` for consistent formatting
- **golint**: Follow Go lint recommendations
- **go vet**: Pass static analysis checks
- **Effective Go**: Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines

### Code Structure

```go
// Package-level documentation
package main

import (
    // Standard library imports first
    "context"
    "fmt"
    
    // Third-party imports
    "github.com/google/go-github/v72/github"
    
    // Local imports last
    "./internal/package"
)

// Public functions have comprehensive documentation
// Example processes the given input and returns a result.
// It returns an error if the input is invalid.
func Example(input string) (string, error) {
    // Implementation
}
```

### Error Handling

- Use descriptive error messages with context
- Wrap errors when adding context: `fmt.Errorf("operation failed: %w", err)`
- Handle errors at appropriate levels
- Prefer explicit error checking over panic

### Comments and Documentation

- **Public APIs**: Must have comprehensive godoc comments
- **Complex Logic**: Add inline comments explaining why, not what
- **TODOs**: Include issue numbers or context
- **Examples**: Provide usage examples in documentation

### Commit Message Format

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or fixing tests
- `chore`: Maintenance tasks

**Examples**:
```bash
feat(cli): add --dry-run flag for preview mode
fix(graphql): handle rate limiting in thread resolution
docs(readme): update installation instructions
test(e2e): add scenario for conversation threads
```

## 🔄 Pull Request Process

### Before Submitting

1. **Ensure all tests pass**:
   ```bash
   go test ./...
   ```

2. **Check code quality**:
   ```bash
   go vet ./...
   go fmt ./...
   ```

3. **Update documentation** if needed
4. **Add tests** for new functionality
5. **Verify your changes** work with the E2E framework

### PR Requirements

- **Clear description**: Explain what and why
- **Link related issues**: Use "Fixes #123" or "Closes #123"
- **Small, focused changes**: One feature/fix per PR
- **Tests included**: All new code should have tests
- **Documentation updated**: Update README, comments, etc.

### PR Review Process

1. **Automated checks**: CI must pass
2. **Code review**: At least one maintainer approval
3. **Testing**: Verify functionality works as expected
4. **Documentation**: Ensure docs are accurate and complete

### After Submission

- **Respond to feedback**: Address reviewer comments promptly
- **Keep PR updated**: Rebase on main if needed
- **Be patient**: Reviews may take a few days

## 🚀 Release Process

### Version Scheme

We use [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality, backwards compatible
- **PATCH**: Bug fixes, backwards compatible

### Release Steps

1. **Update version** in relevant files
2. **Update CHANGELOG** with new features and fixes
3. **Create release tag**: `git tag v1.2.3`
4. **Push tag**: `git push origin v1.2.3`
5. **GitHub Actions** automatically creates release with binaries

## 📞 Getting Help

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Pull Request Comments**: Code review discussions

### Asking Questions

When asking for help, please provide:

- **Clear description** of the problem
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Environment details** (OS, Go version, etc.)
- **Relevant code snippets** or error messages

### Reporting Bugs

Use the bug report template and include:

- **Minimal reproduction case**
- **Complete error messages**
- **System information**
- **Expected behavior**

## 🎯 Areas for Contribution

Looking for ways to contribute? Consider these areas:

### High Priority
- **Performance improvements**: Optimize API calls and processing
- **Error handling**: Better error messages and recovery
- **Cross-platform support**: Windows compatibility improvements
- **Documentation**: More examples and tutorials

### Medium Priority
- **Configuration options**: More flexible tool configuration
- **Output formats**: Alternative output formats (JSON, etc.)
- **Editor integrations**: VS Code, Vim, Emacs plugins
- **Metrics and analytics**: Usage statistics and insights

### Good First Issues
- **Code cleanup**: Refactoring and simplification
- **Test coverage**: Adding tests for edge cases
- **Documentation fixes**: Typos, clarifications, examples
- **Build improvements**: Scripts and automation

## 📄 License

By contributing to prconflict, you agree that your contributions will be licensed under the MIT License.

## 🙏 Recognition

Contributors are recognized in several ways:

- **Contributors section** in the README
- **Release notes** mention significant contributions
- **GitHub insights** track contribution statistics

Thank you for making prconflict better! 🎉 
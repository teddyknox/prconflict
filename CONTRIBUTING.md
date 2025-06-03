# Contributing to prconflict

We welcome pull requests for bug fixes, new features and documentation improvements.

## Getting Started

1. Fork the repo and clone your fork.
2. Install Go 1.21 or later.
3. Run `go mod download` to fetch dependencies.
4. Export a `GITHUB_TOKEN` with `repo` scope for integration tests.

## Development Workflow

Create a feature branch, make your changes and add tests. Use `go test ./...` for unit tests. Integration and end-to-end tests are in `cmd/prconflict` and require GitHub access. Run `./scripts/run-e2e-tests.sh` to execute the full suite.

Format your code with `go fmt` and check for issues with `go vet` before committing.

## Commit Messages

Use short, descriptive messages. Follow the style `feat: add dry-run flag` or `fix: handle rate limiting`.

## Pull Requests

Include a clear description, link any related issues and ensure all tests pass. Small, focused PRs are easier to review. Documentation updates are appreciated when behaviour changes.

## Need Help?

Open an issue or start a discussion on GitHub if you run into problems.

By contributing you agree that your work will be licensed under the MIT License.

# End-to-End Testing

This directory contains integration and end-to-end tests for `prconflict`.

## Test Suites

- **Integration tests** (`integration_test.go`)
  - Validate GitHub API calls.
  - Run with `go test ./cmd/prconflict -run TestIntegration`.
- **E2E tests** (`e2e_test.go`, `e2e_scenarios_test.go`)
  - Create temporary repositories and run the tool against real pull requests.
  - Run with `./scripts/run-e2e-tests.sh`.

Both suites require `GITHUB_TOKEN` with repo permissions.

## Running Tests

```bash
# Unit tests only
go test ./...

# Integration tests
GITHUB_TOKEN=token go test ./cmd/prconflict -run TestIntegration

# All E2E scenarios
./scripts/run-e2e-tests.sh
```

## Scenario Coverage

The E2E tests cover unresolved comments, resolved comments, multiple files, conversation threads and dry-run mode. Additional scenarios exercise edge cases such as unicode content and boundary line numbers.

## Developing Tests

Use `E2ETestFramework` from `e2e_test.go` to create scenarios. Each scenario sets up the repository, adds comments and verifies the conflict markers. Temporary repositories are deleted at the end of each run.

Refer to the existing scenarios in `e2e_scenarios_test.go` for examples.

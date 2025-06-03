# prconflict

[![CI](https://github.com/teddyknox/prconflict/actions/workflows/ci.yml/badge.svg)](https://github.com/teddyknox/prconflict/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/teddyknox/prconflict)](https://goreportcard.com/report/github.com/teddyknox/prconflict)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/teddyknox/prconflict.svg)](https://pkg.go.dev/github.com/teddyknox/prconflict)

`prconflict` inserts unresolved GitHub Pull Request review comments directly into your source files using Git conflict markers.

## Features

- Fetches only unresolved review threads via GraphQL
- Supports multiple files and preserves comment order
- Automatically detects repository, PR number and branch
- Dry run mode for previewing changes

## Installation

```bash
# Install with Go
go install github.com/teddyknox/prconflict/cmd/prconflict@latest

# Or download a release binary from GitHub
```

## Quick Start

```bash
export GITHUB_TOKEN=your_token
prconflict            # auto-detect repo and PR
```

## Usage

```bash
# Specify repo and PR explicitly
prconflict --repo owner/repo --pr 123

# Preview without writing changes
prconflict --dry-run
```

## Project Layout

```
prconflict/
├── cmd/prconflict        # CLI application and tests
├── scripts/              # Helper scripts
├── README.md             # This file
└── go.mod / go.sum       # Go module files
```

## Testing

```bash
go test ./...
```

Integration and end-to-end tests require a valid `GITHUB_TOKEN`. See `cmd/prconflict/E2E_TESTING.md` for details.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

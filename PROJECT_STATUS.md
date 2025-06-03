# Project Status

`prconflict` is a CLI that inserts unresolved GitHub review comments into your files using conflict markers. The project is stable and tested on Linux, macOS and Windows.

## Completed Features

- GraphQL and REST API integration
- Processes only unresolved threads
- Conflict markers with chronological ordering
- Auto-detection of repo and PR
- Cross-file support and dry-run mode

## Testing

- Unit tests for core logic
- Integration tests against GitHub
- End-to-end tests that create real repositories
- GitHub Actions run the full test suite

## Documentation

The README covers installation and usage. `cmd/prconflict/E2E_TESTING.md` describes the testing framework. Contribution guidelines are in `CONTRIBUTING.md`.

## Future Work

Upcoming releases will focus on better error messages, configuration files and additional output formats. Editor integrations and metrics collection are longer term goals.

## Community

Issues and pull requests are welcome. The project follows the MIT License.

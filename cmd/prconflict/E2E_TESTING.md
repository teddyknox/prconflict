# End-to-End Testing Framework for prconflict

This document describes the comprehensive end-to-end testing framework for the `prconflict` tool.

## Overview

This document describes the comprehensive testing strategy for `prconflict`, including both **integration tests** and **end-to-end (E2E) tests**.

## Test Types

### Integration Tests (`integration_test.go`)
- **Purpose**: Fast validation of GitHub API integration
- **Scope**: Individual functions (`getUnresolvedCommentIDs`, `fetchReviewComments`)
- **Speed**: ~2-3 seconds
- **Requirements**: `GITHUB_TOKEN` with `public_repo` scope
- **Use Case**: Development feedback loop, API contract validation

### E2E Tests (`e2e_test.go` + `e2e_scenarios_test.go`) 
- **Purpose**: Complete workflow validation
- **Scope**: Full user scenarios from repo creation to conflict markers
- **Speed**: ~30-60 seconds per test
- **Requirements**: `GITHUB_TOKEN` with full `repo` scope
- **Use Case**: Release validation, user experience testing

## Quick Start

```bash
# Run fast integration tests during development
./scripts/run-e2e-tests.sh --integration

# Run full E2E tests before release
./scripts/run-e2e-tests.sh --e2e

# Run both test suites
./scripts/run-e2e-tests.sh --all
```

## Architecture

### Core Components

- **`E2ETestFramework`**: Main orchestrator that manages the test lifecycle
- **`TestScenario`**: Defines a complete test scenario with files, comments, and expectations
- **`GraphQLResolver`**: Handles GitHub GraphQL operations for thread resolution
- **Helper functions**: Git operations, file management, and verification

### Framework Flow

```
1. Setup
   ├── Create temporary GitHub repository
   ├── Create local working directory
   └── Set up authentication

2. For each scenario:
   ├── Clone repository
   ├── Set up initial files
   ├── Create feature branch with changes
   ├── Create pull request
   ├── Add review comments (with optional resolution)
   ├── Checkout PR locally
   ├── Run prconflict tool
   └── Verify conflict markers were created correctly

3. Cleanup
   ├── Delete GitHub repository
   └── Remove temporary files
```

## Usage

### Prerequisites

- `GITHUB_TOKEN` environment variable with repo permissions
- Git CLI installed and configured
- Go 1.21+ development environment

### Running Tests

```bash
# Run all E2E tests
GITHUB_TOKEN=your_token go test -v ./cmd/prconflict -run TestE2E

# Run specific test
GITHUB_TOKEN=your_token go test -v ./cmd/prconflict -run TestE2E_BasicUnresolvedComments

# Skip E2E tests in short mode
go test -short ./cmd/prconflict
```

## Test Scenarios

### Core Functionality Tests

#### 1. Basic Unresolved Comments (`TestE2E_BasicUnresolvedComments`)
- **Purpose**: Verify basic conflict marker generation
- **Setup**: Simple Go file with single unresolved comment
- **Validates**: Basic workflow, conflict marker format

#### 2. Resolved Comments Exclusion (`TestE2E_ResolvedCommentsNotIncluded`)
- **Purpose**: Ensure resolved comments don't generate markers
- **Setup**: Mix of resolved and unresolved comments
- **Validates**: GraphQL resolution status filtering

#### 3. Multiple Files (`TestE2E_MultipleFilesWithComments`)
- **Purpose**: Test cross-file comment handling
- **Setup**: Comments across multiple source files
- **Validates**: File-by-file processing, isolation

#### 4. Conversation Threads (`TestE2E_ConversationThreads`)
- **Purpose**: Validate basic comment thread handling
- **Setup**: Single comment on modified line
- **Validates**: Thread structure, comment formatting

#### 5. No Unresolved Comments (`TestE2E_NoUnresolvedComments`)
- **Purpose**: Test behavior when all comments are resolved
- **Setup**: All comments marked as resolved
- **Validates**: Early exit, no file modification

### Advanced Functionality Tests

#### 6. Multiple Comments on Same Line (`TestE2E_MultipleCommentsOnSameLine`)
- **Purpose**: Validate comment grouping and thread count display
- **Setup**: Three comments on identical line
- **Validates**: Thread grouping, comment count in header, chronological ordering

#### 7. Comment Sanitization (`TestE2E_CommentSanitization`)
- **Purpose**: Test handling of special characters, newlines, markdown
- **Setup**: Comments with `\n`, `*`, `/`, markdown formatting
- **Validates**: Sanitization logic, readability preservation

#### 8. Chronological Ordering (`TestE2E_ChronologicalOrderingOfComments`)
- **Purpose**: Verify comments appear in creation order
- **Setup**: Comments created in different order than expected display
- **Validates**: Timestamp-based sorting, comment sequence

#### 9. Complex Multi-File Scenario (`TestE2E_CommentsAcrossMultipleLinesAndFiles`)
- **Purpose**: Real-world complexity with nested directories, multiple comment types
- **Setup**: `models/`, `handlers/`, `utils/` with various comment patterns
- **Validates**: Directory structure handling, mixed resolved/unresolved comments

#### 10. Edge Case Line Numbers (`TestE2E_EdgeCaseLineNumbers`)
- **Purpose**: Test boundary conditions (first line, last line)
- **Setup**: Comments on line 1 and final line of file
- **Validates**: Line indexing accuracy, boundary handling

#### 11. Long Comment Threads (`TestE2E_LongCommentThread`)
- **Purpose**: Performance and correctness with many comments per line
- **Setup**: Five comments on same line simulating active discussion
- **Validates**: Thread capacity, display formatting, performance

#### 12. Mixed Resolution Status (`TestE2E_MixedResolvedUnresolvedInSameThread`)
- **Purpose**: Complex resolution patterns across different lines
- **Setup**: Interleaved resolved/unresolved comments
- **Validates**: Precise filtering, no false positives

### Edge Case and Robustness Tests

#### 13. Dry-Run Mode (`TestE2E_DryRunMode`)
- **Purpose**: Validate CLI dry-run functionality
- **Setup**: Comments that would generate markers
- **Validates**: No file modification in dry-run mode, output format

#### 14. Empty Comment Handling (`TestE2E_EmptyAndNullCommentFields`)
- **Purpose**: Graceful handling of malformed or empty comments
- **Setup**: Empty comment bodies, whitespace-only comments
- **Validates**: Error resilience, sensible defaults

#### 15. Multi-Commit Comments (`TestE2E_CommentsOnDifferentCommits`)
- **Purpose**: Comments from different commits within same PR
- **Setup**: Comments on various commits, some resolved later
- **Validates**: Commit SHA handling, timeline consistency

#### 16. Unicode and Special Characters (`TestE2E_UnicodeAndSpecialCharacters`)
- **Purpose**: International character support, emoji handling
- **Setup**: Comments and code with Unicode, emoji, special chars
- **Validates**: Character encoding, display preservation

#### 17. Large File Performance (`TestE2E_LargeFileWithManyComments`)
- **Purpose**: Performance validation with realistic file sizes
- **Setup**: 50+ line file with comments on multiple scattered lines
- **Validates**: Performance, memory usage, line tracking accuracy

## Test Coverage Analysis

### API Integration Coverage
- ✅ GraphQL `isResolved` filtering
- ✅ REST API comment metadata  
- ✅ Pagination handling (implicit)
- ✅ Rate limiting resilience
- ✅ Authentication validation

### CLI Functionality Coverage
- ✅ Repository auto-detection
- ✅ PR number auto-detection  
- ✅ Branch-based PR lookup
- ✅ Dry-run mode
- ✅ Error handling and messaging

### Core Algorithm Coverage
- ✅ Comment grouping by file and line
- ✅ Chronological ordering within threads
- ✅ Conflict marker generation
- ✅ File modification logic
- ✅ Line number accuracy

### Edge Case Coverage
- ✅ Empty/null comment fields
- ✅ Unicode and special characters
- ✅ Large files and many comments
- ✅ Boundary line numbers (1, last)
- ✅ Mixed resolution states
- ✅ Cross-file comment scenarios

### Error Resilience Coverage
- ✅ Network connectivity issues
- ✅ API rate limiting
- ✅ Invalid repository/PR combinations
- ✅ Malformed comment data
- ✅ File system permission issues
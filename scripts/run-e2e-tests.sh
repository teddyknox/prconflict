#!/bin/bash

# E2E Testing Script for prconflict
# Runs integration tests (fast API validation) and/or E2E tests (full workflow)

set -e

cd "$(dirname "$0")/.."

# Parse command line arguments
RUN_INTEGRATION=false
RUN_E2E=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --integration)
            RUN_INTEGRATION=true
            shift
            ;;
        --e2e)
            RUN_E2E=true
            shift
            ;;
        --all)
            RUN_INTEGRATION=true
            RUN_E2E=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [--integration] [--e2e] [--all] [--verbose]"
            echo ""
            echo "  --integration  Run fast integration tests (API validation)"
            echo "  --e2e         Run full end-to-end tests (complete workflow)"
            echo "  --all         Run both integration and E2E tests"
            echo "  --verbose     Show detailed output"
            echo ""
            echo "Default: runs E2E tests only"
            exit 0
            ;;
        *)
            echo "Unknown option $1"
            exit 1
            ;;
    esac
done

# Default to E2E if no specific test type selected
if [[ "$RUN_INTEGRATION" = false && "$RUN_E2E" = false ]]; then
    RUN_E2E=true
fi

# Run E2E tests for prconflict
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}prconflict E2E Test Runner${NC}"
echo "================================"

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

# Check if GITHUB_TOKEN is set
if [[ -z "${GITHUB_TOKEN:-}" ]]; then
    echo -e "${RED}Error: GITHUB_TOKEN environment variable is not set${NC}"
    echo "Please set GITHUB_TOKEN with a GitHub Personal Access Token that has 'repo' scope"
    echo "Example: export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx"
    exit 1
fi

# Check if git is available
if ! command -v git &> /dev/null; then
    echo -e "${RED}Error: git command not found${NC}"
    echo "Please install Git CLI"
    exit 1
fi

# Check if gh CLI is available (required by prconflict)
if ! command -v gh &> /dev/null; then
    echo -e "${YELLOW}Warning: gh CLI not found${NC}"
    echo "The GitHub CLI (gh) is required by prconflict for auto-detection"
    echo "Install from: https://cli.github.com/"
fi

# Check if we're in the right directory
if [[ ! -f "cmd/prconflict/main.go" ]]; then
    echo -e "${RED}Error: Not in prconflict repository root${NC}"
    echo "Please run this script from the prconflict repository root"
    exit 1
fi

echo -e "${GREEN}✓ Prerequisites check passed${NC}"
echo

echo "🚀 Starting prconflict test suite..."

if [[ "$RUN_INTEGRATION" = true ]]; then
    echo ""
    echo "📋 Running integration tests (API validation)..."
    if [[ "$VERBOSE" = true ]]; then
        go test -tags=integration -v ./cmd/prconflict/
    else
        go test -tags=integration ./cmd/prconflict/
    fi
    echo "✅ Integration tests completed"
fi

if [[ "$RUN_E2E" = true ]]; then
    echo ""
    echo "🔄 Running E2E tests (full workflow)..."
    echo "⚠️  This will create and delete temporary GitHub repositories"
    
    if [[ "$VERBOSE" = true ]]; then
        go test -v ./cmd/prconflict/ -run "TestE2E_"
    else
        go test ./cmd/prconflict/ -run "TestE2E_"
    fi
    echo "✅ E2E tests completed"
fi

echo ""
echo "🎉 All selected tests passed!" 
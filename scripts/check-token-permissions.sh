#!/bin/bash

# Check GitHub token permissions for prconflict testing
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}GitHub Token Permission Checker for prconflict${NC}"
echo "=================================================="

# Check if GITHUB_TOKEN is set
if [[ -z "${GITHUB_TOKEN:-}" ]]; then
    echo -e "${RED}❌ GITHUB_TOKEN environment variable is not set${NC}"
    echo "Please set your GitHub token:"
    echo "export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx"
    exit 1
fi

echo -e "${GREEN}✓ GITHUB_TOKEN is set${NC}"

# Check if token is valid and get user info
echo -e "\n${YELLOW}Checking token validity...${NC}"
if ! USER_INFO=$(curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user); then
    echo -e "${RED}❌ Failed to validate token${NC}"
    exit 1
fi

if echo "$USER_INFO" | grep -q '"message": "Bad credentials"'; then
    echo -e "${RED}❌ Invalid GitHub token${NC}"
    exit 1
fi

USERNAME=$(echo "$USER_INFO" | grep '"login"' | cut -d'"' -f4)
echo -e "${GREEN}✓ Token is valid for user: ${USERNAME}${NC}"

# Check scopes
echo -e "\n${YELLOW}Checking token scopes...${NC}"
SCOPES_RESPONSE=$(curl -s -I -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user)
SCOPES=$(echo "$SCOPES_RESPONSE" | grep -i "x-oauth-scopes:" | cut -d' ' -f2- | tr -d '\r\n')

echo "Available scopes: $SCOPES"

# Check for required permissions
echo -e "\n${YELLOW}Permission Analysis:${NC}"

# Integration tests (minimal permissions)
if echo "$SCOPES" | grep -E "(repo|public_repo)" > /dev/null; then
    echo -e "${GREEN}✓ Integration Tests: Supported${NC} (repo/public_repo scope found)"
    INTEGRATION_SUPPORTED=true
else
    echo -e "${RED}❌ Integration Tests: NOT Supported${NC} (needs repo or public_repo scope)"
    INTEGRATION_SUPPORTED=false
fi

# E2E tests (full repo permissions)
if echo "$SCOPES" | grep -E "repo" > /dev/null && ! echo "$SCOPES" | grep -E "public_repo" | grep -v "repo" > /dev/null; then
    echo -e "${GREEN}✓ E2E Tests: Supported${NC} (full repo scope found)"
    E2E_SUPPORTED=true
    
    # Test repository creation capability
    echo -e "\n${YELLOW}Testing repository creation capability...${NC}"
    TEST_REPO_NAME="prconflict-permission-test-$(date +%s)"
    
    CREATE_RESPONSE=$(curl -s -X POST \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$TEST_REPO_NAME\",\"description\":\"Temporary test repo\",\"private\":false}" \
        https://api.github.com/user/repos)
    
    if echo "$CREATE_RESPONSE" | grep -q '"id"'; then
        echo -e "${GREEN}✓ Repository creation: Successful${NC}"
        
        # Clean up test repository
        sleep 1
        DELETE_RESPONSE=$(curl -s -X DELETE \
            -H "Authorization: token $GITHUB_TOKEN" \
            "https://api.github.com/repos/$USERNAME/$TEST_REPO_NAME")
        echo -e "${GREEN}✓ Repository deletion: Successful${NC}"
    else
        echo -e "${RED}❌ Repository creation: Failed${NC}"
        echo "Response: $CREATE_RESPONSE"
        E2E_SUPPORTED=false
    fi
elif echo "$SCOPES" | grep -E "public_repo" > /dev/null; then
    echo -e "${YELLOW}⚠ E2E Tests: Limited Support${NC} (public_repo only - can't create private repos)"
    E2E_SUPPORTED=true
else
    echo -e "${RED}❌ E2E Tests: NOT Supported${NC} (needs full repo scope)"
    E2E_SUPPORTED=false
fi

# Rate limit check
echo -e "\n${YELLOW}Checking rate limits...${NC}"
RATE_LIMIT=$(curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/rate_limit)
REMAINING=$(echo "$RATE_LIMIT" | grep '"remaining"' | head -1 | cut -d':' -f2 | cut -d',' -f1 | tr -d ' ')
LIMIT=$(echo "$RATE_LIMIT" | grep '"limit"' | head -1 | cut -d':' -f2 | cut -d',' -f1 | tr -d ' ')

echo "Rate limit: $REMAINING/$LIMIT remaining"
if [[ $REMAINING -lt 100 ]]; then
    echo -e "${YELLOW}⚠ Warning: Low rate limit remaining${NC}"
else
    echo -e "${GREEN}✓ Rate limit: Sufficient${NC}"
fi

# Summary
echo -e "\n${BLUE}Summary:${NC}"
echo "=========================="
if [[ "$INTEGRATION_SUPPORTED" == true ]]; then
    echo -e "${GREEN}✓ Integration Tests: Ready${NC}"
    echo "  Run with: go test ./cmd/prconflict -run TestIntegration"
else
    echo -e "${RED}❌ Integration Tests: Not Ready${NC}"
    echo "  Need: repo or public_repo scope"
fi

if [[ "$E2E_SUPPORTED" == true ]]; then
    echo -e "${GREEN}✓ E2E Tests: Ready${NC}"
    echo "  Run with: ./scripts/run-e2e-tests.sh"
else
    echo -e "${RED}❌ E2E Tests: Not Ready${NC}"
    echo "  Need: full repo scope"
fi

echo -e "\n${YELLOW}Token Requirements:${NC}"
echo "• Integration Tests: public_repo or repo scope"
echo "• E2E Tests: repo scope (full repository access)"
echo ""
echo "Generate token at: https://github.com/settings/tokens"
echo "Select 'repo' scope for full functionality"

if [[ "$INTEGRATION_SUPPORTED" == true && "$E2E_SUPPORTED" == true ]]; then
    echo -e "\n${GREEN}🎉 All tests supported! Your token is ready.${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Token needs additional permissions${NC}"
    exit 1
fi 
#!/bin/bash

echo "🔍 GitHub Token Type Checker"
echo "============================"

if [[ -z "$GITHUB_TOKEN" ]]; then
    echo "❌ GITHUB_TOKEN not set"
    exit 1
fi

TOKEN_PREFIX="${GITHUB_TOKEN:0:4}"
echo "Token prefix: ${TOKEN_PREFIX}..."

if [[ "$TOKEN_PREFIX" == "ghp_" ]]; then
    echo "✅ Classic Personal Access Token (recommended for E2E tests)"
elif [[ "$TOKEN_PREFIX" == "gith" ]]; then
    echo "⚠️  Fine-grained Personal Access Token (limited functionality)"
    echo "   E2E tests require a classic token for repository creation"
else
    echo "❓ Unknown token type"
fi

echo ""
echo "Testing repository creation..."
TEST_REPO="prconflict-token-test-$(date +%s)"

RESPONSE=$(curl -s -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"$TEST_REPO\",\"description\":\"Token test\",\"private\":false}" \
    https://api.github.com/user/repos)

if echo "$RESPONSE" | grep -q '"id"'; then
    echo "✅ Repository creation: SUCCESS"
    
    # Get username and cleanup
    USERNAME=$(curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user | grep '"login"' | cut -d'"' -f4)
    sleep 1
    curl -s -X DELETE -H "Authorization: token $GITHUB_TOKEN" "https://api.github.com/repos/$USERNAME/$TEST_REPO" > /dev/null
    echo "✅ Repository deletion: SUCCESS"
    echo ""
    echo "🎉 Token is ready for E2E tests!"
elif echo "$RESPONSE" | grep -q "Resource not accessible"; then
    echo "❌ Repository creation: FAILED"
    echo "   Error: Resource not accessible by personal access token"
    echo "   Solution: Use a classic personal access token with 'repo' scope"
else
    echo "❌ Repository creation: FAILED"
    echo "   Response: $RESPONSE"
fi 
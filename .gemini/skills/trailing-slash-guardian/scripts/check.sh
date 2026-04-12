#!/bin/bash
# trailing-slash-check.sh - Scans the project for API routes/calls missing trailing slashes.

echo "🔍 Scanning for missing trailing slashes..."
FAILED=0

# 1. Backend Routes
echo "--- Backend (internal/adapter/http/handler.go) ---"
MISSING_BACKEND=$(grep -E 'r\.(Get|Post|Put|Patch|Delete|Route)\("[^"]+[^/]"' backend/internal/adapter/http/handler.go | grep -v "/health" | grep -v "/api/v1")
if [ ! -z "$MISSING_BACKEND" ]; then
    echo "❌ Found routes missing trailing slashes:"
    echo "$MISSING_BACKEND"
    FAILED=1
else
    echo "✅ Backend routes OK."
fi

# 2. Frontend API Calls
echo "--- Frontend (src/api/client.ts) ---"
MISSING_FRONTEND=$(grep -E 'api\.(get|post|put|patch|delete)\((["`])[^"`]+[^/]\2' frontend/src/api/client.ts | grep -v "/api/v1" | grep -v "\?window=")
if [ ! -z "$MISSING_FRONTEND" ]; then
    echo "❌ Found API calls missing trailing slashes:"
    echo "$MISSING_FRONTEND"
    FAILED=1
else
    echo "✅ Frontend API calls OK."
fi

# 3. Mobile Repository Calls
echo "--- Mobile (lib/data/repositories/) ---"
MISSING_MOBILE=$(grep -rE '_dio\.(get|post|put|patch|delete)\((["`])[^"`]+[^/]\2' mobile/lib/data/repositories/ | grep -v "/$")
if [ ! -z "$MISSING_MOBILE" ]; then
    echo "❌ Found repository calls missing trailing slashes:"
    echo "$MISSING_MOBILE"
    FAILED=1
else
    echo "✅ Mobile repository calls OK."
fi

if [ $FAILED -eq 1 ]; then
    echo "❌ Trailing slash check failed."
    exit 1
else
    echo "✅ All trailing slash checks passed."
    exit 0
fi

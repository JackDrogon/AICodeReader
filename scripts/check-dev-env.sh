#!/bin/bash

# Check development environment for CI tools

set -e

echo "🔍 Checking development environment..."
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if command exists
check_command() {
    if command -v "$1" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ $1 is installed${NC}"
        if [ "$2" ]; then
            version=$($1 $2 2>/dev/null || echo "version unknown")
            echo -e "   ${YELLOW}Version: $version${NC}"
        fi
        return 0
    else
        echo -e "${RED}❌ $1 is not installed${NC}"
        return 1
    fi
}

# Function to check Go version
check_go_version() {
    if command -v go >/dev/null 2>&1; then
        version=$(go version | cut -d' ' -f3)
        echo -e "${GREEN}✅ Go is installed${NC}"
        echo -e "   ${YELLOW}Version: $version${NC}"
        
        # Check if it's Go 1.24.x
        if [[ $version == go1.24* ]]; then
            echo -e "   ${GREEN}✅ Go version is compatible${NC}"
        else
            echo -e "   ${YELLOW}⚠️  Recommended Go version is 1.24.x${NC}"
        fi
        return 0
    else
        echo -e "${RED}❌ Go is not installed${NC}"
        return 1
    fi
}

errors=0

echo "📋 Required tools:"
echo ""

# Check Go
check_go_version || errors=$((errors + 1))

# Check golangci-lint
if check_command "golangci-lint" "--version"; then
    echo ""
else
    errors=$((errors + 1))
    echo -e "   ${YELLOW}💡 Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest${NC}"
    echo ""
fi

# Check make
check_command "make" "--version" || errors=$((errors + 1))
echo ""

# Check git
check_command "git" "--version" || errors=$((errors + 1))
echo ""

echo "📋 Optional tools:"
echo ""

# Check additional tools
check_command "curl" "--version" >/dev/null 2>&1 && echo -e "${GREEN}✅ curl is available${NC}" || echo -e "${YELLOW}⚠️  curl not available${NC}"

echo ""
echo "🧪 Testing CI commands..."
echo ""

# Test if golangci-lint config is valid
if command -v golangci-lint >/dev/null 2>&1; then
    if golangci-lint config path >/dev/null 2>&1; then
        echo -e "${GREEN}✅ golangci-lint config is valid${NC}"
    else
        echo -e "${RED}❌ golangci-lint config has issues${NC}"
        errors=$((errors + 1))
    fi
else
    echo -e "${YELLOW}⚠️  Cannot test golangci-lint config (not installed)${NC}"
fi

# Test if go modules are working
if go mod verify >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Go modules are valid${NC}"
else
    echo -e "${RED}❌ Go modules have issues${NC}"
    errors=$((errors + 1))
fi

echo ""
if [ $errors -eq 0 ]; then
    echo -e "${GREEN}🎉 Development environment is ready for CI!${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo "  1. Run 'make ci' to test all CI checks locally"
    echo "  2. Run 'make help' to see all available commands"
    exit 0
else
    echo -e "${RED}❌ Found $errors issue(s) in development environment${NC}"
    echo ""
    echo -e "${YELLOW}Please fix the issues above before running CI commands.${NC}"
    exit 1
fi 
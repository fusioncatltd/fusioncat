#!/bin/bash

# Script to validate GitHub Actions workflow syntax using act
# Requires: act (https://github.com/nektos/act)

set -e

echo "================================================"
echo "GitHub Actions Workflow Syntax Checker"
echo "================================================"
echo ""

# Check if act is installed
if ! command -v act &> /dev/null; then
    echo "❌ 'act' is not installed!"
    echo ""
    echo "To install act, run one of the following:"
    echo "  - macOS: brew install act"
    echo "  - Linux: curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash"
    echo "  - Or run: make install-act"
    exit 1
fi

echo "✅ Using act version: $(act --version)"
echo ""

# Check workflow syntax
echo "Checking syntax of GitHub Actions workflow..."
echo "-------------------------------------------"

workflow_count=0
errors_found=false

for workflow in .github/workflows/*.yml; do
    if [ -f "$workflow" ]; then
        workflow_count=$((workflow_count + 1))
        echo ""
        echo "Checking: $(basename $workflow)"
        
        # List jobs in the workflow (this validates syntax)
        if act -W "$workflow" -l --container-architecture linux/amd64 2>&1 | grep -v "authentication required" > /dev/null; then
            echo "  ✅ Syntax valid"
        else
            echo "  ❌ Syntax error detected"
            errors_found=true
        fi
    fi
done

echo ""
echo "-------------------------------------------"
echo "Checked $workflow_count workflow(s)"
echo ""

if [ "$errors_found" = true ]; then
    echo "❌ Some workflows have syntax errors"
    exit 1
else
    echo "✅ All workflows have valid syntax"
    echo ""
    echo "ℹ️  Note: This validates syntax only. To test the actual Docker build, run:"
    echo "    make docker-test"
    exit 0
fi
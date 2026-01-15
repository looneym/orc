#!/bin/bash
# preflight-check.sh - Validate environment before orchestration test
set -euo pipefail

echo "=== Orchestration Test Pre-flight Check ==="
echo ""

# Run orc doctor
echo "Running orc doctor..."
echo ""

if orc doctor; then
    echo ""
    echo "✓ Environment validation passed"
    echo ""
    exit 0
else
    echo ""
    echo "✗ Environment validation failed"
    echo ""
    echo "ERROR: Fix issues reported by 'orc doctor' before running test"
    echo "See INSTALL.md for workspace trust setup instructions"
    echo ""
    exit 1
fi

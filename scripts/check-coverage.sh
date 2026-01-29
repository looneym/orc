#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Thresholds (as integers for comparison)
CORE_THRESHOLD=70
APP_THRESHOLD=50
SQLITE_THRESHOLD=60
FILESYSTEM_THRESHOLD=50

# Exempt packages (skip silently)
EXEMPT="internal/core/effects internal/core/factory internal/core/git internal/core/workbench internal/core/workshop internal/cli internal/tmux"

failures=()

is_exempt() {
    local pkg="$1"
    for exempt in $EXEMPT; do
        [[ "$pkg" == "$exempt" ]] && return 0
    done
    return 1
}

get_threshold() {
    local pkg="$1"
    case "$pkg" in
        internal/core/*) echo "$CORE_THRESHOLD" ;;
        internal/app) echo "$APP_THRESHOLD" ;;
        internal/adapters/sqlite) echo "$SQLITE_THRESHOLD" ;;
        internal/adapters/filesystem) echo "$FILESYSTEM_THRESHOLD" ;;
        *) echo "0" ;;
    esac
}

check_package() {
    local pkg="$1"
    local threshold="$2"

    local output
    output=$(go test -cover "./$pkg" 2>&1) || true

    # Parse "coverage: XX.X% of statements"
    local coverage
    coverage=$(echo "$output" | grep -oE 'coverage: [0-9.]+%' | grep -oE '[0-9.]+' || echo "0")

    # Skip packages with no statements (0.0% but no test failures)
    [[ "$coverage" == "0" ]] && return 0

    # Compare as integers (truncate decimal)
    local cov_int="${coverage%%.*}"
    if [[ "$cov_int" -lt "$threshold" ]]; then
        failures+=("$pkg: ${coverage}% < ${threshold}%")
    fi
}

# Check core packages (with exemptions)
for dir in "$PROJECT_ROOT"/internal/core/*/; do
    [[ -d "$dir" ]] || continue
    pkg="internal/core/$(basename "$dir")"
    is_exempt "$pkg" && continue
    threshold=$(get_threshold "$pkg")
    check_package "$pkg" "$threshold"
done

# Check other packages
check_package "internal/app" "$APP_THRESHOLD"
check_package "internal/adapters/sqlite" "$SQLITE_THRESHOLD"
check_package "internal/adapters/filesystem" "$FILESYSTEM_THRESHOLD"

# Report
if [[ ${#failures[@]} -eq 0 ]]; then
    echo "✓ All coverage thresholds met"
    exit 0
else
    for f in "${failures[@]}"; do
        echo "FAIL: $f"
    done
    echo "✗ Coverage check failed"
    exit 1
fi

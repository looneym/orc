#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ALLOWLIST="$SCRIPT_DIR/test-presence-allowlist.txt"

missing=()

# Check if a path is in the allowlist
is_allowed() {
    local path="$1"
    if [[ -f "$ALLOWLIST" ]]; then
        grep -qxF "$path" "$ALLOWLIST" 2>/dev/null && return 0
    fi
    return 1
}

check_test() {
    local test="$1"
    if [[ ! -f "$PROJECT_ROOT/$test" ]]; then
        if ! is_allowed "$test"; then
            missing+=("$test")
        fi
    fi
}

# Check app services
for f in "$PROJECT_ROOT"/internal/app/*_service.go; do
    [[ -f "$f" ]] || continue
    base=$(basename "$f" .go)
    check_test "internal/app/${base}_test.go"
done

# Check sqlite repos
for f in "$PROJECT_ROOT"/internal/adapters/sqlite/*_repo.go; do
    [[ -f "$f" ]] || continue
    base=$(basename "$f" .go)
    check_test "internal/adapters/sqlite/${base}_test.go"
done

# Check core guards
for dir in "$PROJECT_ROOT"/internal/core/*/; do
    [[ -f "$dir/guards.go" ]] || continue
    subdir=$(basename "$dir")
    check_test "internal/core/$subdir/guards_test.go"
done

# Report
if [[ ${#missing[@]} -eq 0 ]]; then
    echo "✓ All test files present"
    exit 0
else
    for f in "${missing[@]}"; do
        echo "MISSING: $f"
    done
    echo "✗ ${#missing[@]} test files missing"
    exit 1
fi

#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TUI_FILE="$PROJECT_ROOT/internal/cli/summary_tui.go"
TEST_FILE="$PROJECT_ROOT/internal/cli/summary_tui_test.go"
DOCS_FILE="$PROJECT_ROOT/docs/dev/tui.md"

errors=()

echo "Checking TUI entity-action matrix consistency..."

# --- Extract allEntityTypes from test ---
entity_types=()
while IFS= read -r line; do
    type=$(echo "$line" | sed -n 's/.*"\([A-Z][A-Z]*\)".*/\1/p')
    if [[ -n "$type" ]]; then
        entity_types+=("$type")
    fi
done < <(sed -n '/allEntityTypes.*\[\]string{/,/}/p' "$TEST_FILE" | grep '"')

# --- Extract allActions from test ---
actions=()
while IFS= read -r line; do
    action=$(echo "$line" | sed -n 's/.*"\([a-z][a-z]*\)".*/\1/p')
    if [[ -n "$action" ]]; then
        actions+=("$action")
    fi
done < <(sed -n '/allActions.*\[\]string{/,/}/p' "$TEST_FILE" | grep '"')

if [[ ${#entity_types[@]} -eq 0 ]]; then
    errors+=("PARSE: Could not extract allEntityTypes from test file")
fi

if [[ ${#actions[@]} -eq 0 ]]; then
    errors+=("PARSE: Could not extract allActions from test file")
fi

# --- Check 1: Every action has a formatHint call in renderStatusBar ---
echo "  Checking status bar hints..."
for action in "${actions[@]}"; do
    # Skip expand — it's shown as "expand" but gated differently (always available in status bar)
    if [[ "$action" == "expand" ]]; then
        continue
    fi
    if ! grep -q "formatHint.*\"$action\"" "$TUI_FILE"; then
        errors+=("HINT: action '$action' has no formatHint call in renderStatusBar")
    fi
done

# --- Check 2: Every action has a case handler or is handled by enter/l ---
echo "  Checking key handlers..."
for action in "${actions[@]}"; do
    case "$action" in
        yank|open|focus|close|goblin|note|review|run|deploy)
            if ! grep -q "entityHasAction.*\"$action\"" "$TUI_FILE"; then
                errors+=("HANDLER: action '$action' has no entityHasAction guard in key handler")
            fi
            ;;
        expand)
            # expand is handled via isExpandable which delegates to entityHasAction
            if ! grep -q 'isExpandable' "$TUI_FILE"; then
                errors+=("HANDLER: 'expand' action requires isExpandable function")
            fi
            ;;
    esac
done

# --- Check 3: Entity-action matrix in docs matches code ---
echo "  Checking docs table..."
if [[ -f "$DOCS_FILE" ]]; then
    for etype in "${entity_types[@]}"; do
        if ! grep -q "| $etype " "$DOCS_FILE"; then
            errors+=("DOCS: entity type '$etype' not in docs/dev/tui.md matrix table")
        fi
    done
else
    errors+=("DOCS: docs/dev/tui.md does not exist")
fi

# --- Report ---
if [[ ${#errors[@]} -eq 0 ]]; then
    echo "✓ TUI entity-action matrix is consistent"
    exit 0
else
    echo ""
    for err in "${errors[@]}"; do
        echo "  $err"
    done
    echo ""
    echo "✗ ${#errors[@]} TUI action issues found"
    exit 1
fi

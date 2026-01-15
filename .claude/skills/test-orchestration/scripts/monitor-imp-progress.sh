#!/bin/bash
# monitor-imp-progress.sh - Watch IMP activity and task progress
# Usage: ./monitor-imp-progress.sh <grove-id>

set -euo pipefail

GROVE_ID="${1:-}"

if [[ -z "$GROVE_ID" ]]; then
    echo "status=ERROR"
    echo "error=Missing grove ID argument"
    exit 1
fi

# Get grove info
GROVE_INFO=$(orc grove show "$GROVE_ID" 2>/dev/null || echo "")

if [[ -z "$GROVE_INFO" ]]; then
    echo "status=FAIL"
    echo "grove_exists=false"
    echo "error=Grove '$GROVE_ID' not found"
    exit 1
fi

echo "grove_exists=true"

# Extract grove path and mission
GROVE_PATH=$(echo "$GROVE_INFO" | grep "Path:" | awk '{print $2}')
MISSION_ID=$(echo "$GROVE_INFO" | grep "Mission:" | awk '{print $2}')

echo "grove_path=$GROVE_PATH"
echo "mission_id=$MISSION_ID"

# Check if grove directory exists
if [[ ! -d "$GROVE_PATH" ]]; then
    echo "grove_dir_exists=false"
    echo "error=Grove directory not found at $GROVE_PATH"
    exit 1
fi

echo "grove_dir_exists=true"

# Check for assigned work
ASSIGNED_WORK="$GROVE_PATH/.orc/assigned-work.json"

if [[ -f "$ASSIGNED_WORK" ]]; then
    echo "assigned_work_exists=true"

    # Extract epic ID from assignment
    EPIC_ID=$(jq -r '.epic_id' "$ASSIGNED_WORK" 2>/dev/null || echo "")

    if [[ -n "$EPIC_ID" ]]; then
        echo "epic_id=$EPIC_ID"

        # Get task counts using orc task list
        TASK_LIST=$(orc task list --epic "$EPIC_ID" 2>/dev/null || echo "")

        if [[ -n "$TASK_LIST" ]]; then
            TOTAL_TASKS=$(echo "$TASK_LIST" | grep -c "TASK-" || echo "0")
            READY=$(echo "$TASK_LIST" | grep -c "\[ready\]" || echo "0")
            IN_PROGRESS=$(echo "$TASK_LIST" | grep -c "\[implement\]" || echo "0")
            COMPLETED=$(echo "$TASK_LIST" | grep -c "\[complete\]" || echo "0")

            echo "tasks=$TOTAL_TASKS"
            echo "ready=$READY"
            echo "in_progress=$IN_PROGRESS"
            echo "completed=$COMPLETED"

            # Calculate progress percentage
            if [[ "$TOTAL_TASKS" -gt 0 ]]; then
                PROGRESS=$((COMPLETED * 100 / TOTAL_TASKS))
                echo "progress_percent=$PROGRESS"
            else
                echo "progress_percent=0"
            fi
        else
            echo "tasks=0"
            echo "ready=0"
            echo "in_progress=0"
            echo "completed=0"
            echo "progress_percent=0"
        fi
    else
        echo "epic_id=unknown"
        echo "tasks=unknown"
    fi
else
    echo "assigned_work_exists=false"
    echo "tasks=0"
fi

# Check for git changes in grove
cd "$GROVE_PATH"

if git diff --quiet 2>/dev/null; then
    echo "git_changes=false"
else
    echo "git_changes=true"
    CHANGED_FILES=$(git diff --name-only | wc -l)
    echo "files_changed=$CHANGED_FILES"
fi

if git diff --cached --quiet 2>/dev/null; then
    echo "git_staged=false"
else
    echo "git_staged=true"
fi

# Check for untracked files
UNTRACKED=$(git ls-files --others --exclude-standard | wc -l)
echo "untracked_files=$UNTRACKED"

# Overall status
echo "status=OK"
echo "timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

exit 0

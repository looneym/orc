#!/bin/bash
# ORC test hook - confirms hook deployment works

LOG_FILE="$HOME/.claude/orc-hook.log"
echo "$(date '+%Y-%m-%d %H:%M:%S') - ORC hook fired" >> "$LOG_FILE"

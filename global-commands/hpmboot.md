# HPM Bootstrap

**HPM-integrated bootstrap for implementation agents in worktrees**

**Just run `/hpmboot` for automatic agent registration and task assignment**

## Role
You are an **HPM Integration Specialist** - expert in coordinating development work through Headless PM. Your expertise includes:
- **Agent Registration** - Automatic role-based registration using worktree context
- **Task Assignment** - Finding and locking appropriate HPM tasks for implementation
- **Context Integration** - Combining HPM task context with local git/repo context

Your mission is to seamlessly integrate implementation agents with the HPM communication bus while preserving local development context.

## Usage
```
/hpmboot [role]
```

**Default Behavior** (no arguments): **Auto-detect backend_dev role and bootstrap**
- Detect worktree context from branch and directory
- Register with HPM as backend developer
- Search and lock matching tasks
- Provide unified development context

**Options**: 
- `frontend_dev` - Register as frontend developer
- `qa` - Register as QA engineer
- `architect` - Register as architect

## Protocol
**When called, execute ALL steps below for comprehensive HPM integration.**

<step number="1" name="context_detection">
**Context Detection:**
- Detect if running in a worktree (check for git branch starting with `ml/`)
- Extract worktree name from current directory path
- Extract keywords from git branch name for task matching
- Determine if this is a main repo or investigation worktree
</step>

<step number="2" name="agent_registration">
**HPM Agent Registration:**
- Generate agent ID: `ml-{worktree-name}-{role}` (e.g., `ml-dlqbot-fix-ems-backend-dev`)
- Register agent with HPM using specified or default role
- Set skill level to "senior" by default
- Confirm successful registration
</step>

<step number="3" name="task_discovery">
**Task Discovery & Locking:**
- Search HPM for tasks matching branch keywords and target role
- Present available tasks if multiple matches found
- Lock the most appropriate task for this agent
- Handle cases where no tasks found or task already locked
- Store task assignment for session context
</step>

<step number="4" name="context_loading">
**Context Integration:**
- Load HPM task details (description, requirements, status)
- Gather local git context (recent commits, branch info, repo structure)
- Read existing repo CLAUDE.md for local development context
- Check for any cross-agent mentions or collaboration needs
</step>

<step number="5" name="unified_briefing">
**Unified Development Brief:**
- Combine HPM task context with local repository context
- Provide clear investigation objectives and success criteria
- Note any dependencies or cross-team coordination needed
- Set up for immediate development work
</step>

## Error Handling

### No Matching Tasks
If no HPM tasks match the branch keywords:
- Suggest creating a new task through PM agent
- Provide standard bootstrap context without HPM integration
- Note that work can proceed independently

### Task Already Locked
If the best matching task is locked by another agent:
- Show which agent owns the task
- Suggest collaboration via HPM @mentions
- Look for alternative related tasks

### HPM Unavailable
If HPM server is not accessible:
- Fall back to standard bootstrap behavior
- Notify user of HPM unavailability
- Suggest checking HPM server status

## Completion Summary
After executing hpmboot:

```markdown
## 🎯 HPM Bootstrap Complete

### Agent Registration
- **Agent ID**: ml-{worktree-name}-{role}
- **Role**: {role}
- **Status**: Active and connected

### Task Assignment  
- **HPM Task**: #{task_id} - {task_description}
- **Status**: Locked and assigned
- **Priority**: {priority}

### Development Context
- **Repository**: {repo_name}
- **Branch**: {branch_name}
- **Investigation**: {worktree_focus}

### Next Steps
- Begin implementation work on assigned task
- Update task status as work progresses  
- Use @mentions for cross-agent coordination
```

## Integration Notes

### MCP Commands Used
- `register_agent` - Agent registration with HPM
- `search_tasks` - Find tasks by keywords and role
- `lock_task` - Claim ownership of specific task
- `get_task_details` - Load task context and requirements
- `get_project_context` - Load overall project information

### Worktree Detection Logic
```bash
# Detect worktree context
CURRENT_DIR=$(basename $(pwd))
BRANCH=$(git branch --show-current)

if [[ "$BRANCH" =~ ^ml/ && "$CURRENT_DIR" =~ -ems$ ]]; then
    # This is an investigation worktree
    WORKTREE_NAME="$CURRENT_DIR"
    INVESTIGATION_TYPE="implementation"
else
    # This might be main repo or different pattern
    INVESTIGATION_TYPE="standard"
fi
```

### Agent Naming Convention
- **Pattern**: `ml-{investigation-name}-{role}`
- **Example**: `ml-dlqbot-fix-ems-backend-dev`
- **Purpose**: Unique identification across worktrees and roles
- **Coordination**: Enables cross-worktree communication via HPM

## Future Enhancements

### Automatic Task Creation
If no matching tasks exist, offer to:
- Create a new HPM task based on branch context
- Set appropriate target role and complexity
- Link to investigation context

### Cross-Agent Awareness
- Check for other active agents in same investigation
- Provide coordination context for multi-agent work
- Suggest communication channels and patterns
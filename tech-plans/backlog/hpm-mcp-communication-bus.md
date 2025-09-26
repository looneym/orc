# HPM MCP Communication Bus Integration

**Status**: investigating

## Problem & Solution
**Current Issue:** ORC worktrees operate in isolation with no cross-context communication - findings, blockers, and dependencies in one investigation are invisible to other parallel work streams
**Solution:** Integrate Headless PM MCP server as a communication bus allowing Claude sessions across worktrees to coordinate via @mentions, shared documents, and status updates

## Integration Context

### Current ORC Architecture
- **TMux Layout**: Vim + Claude + Shell per worktree (preserved)
- **Worktree Focus**: Single repository investigations with isolated contexts  
- **Tech Plans**: Symlinked local planning with central ORC storage
- **Parallel Development**: Multiple independent investigation streams

### Headless PM Communication Layer
- **MCP Integration**: Natural language interface via Claude Code CLI
- **Agent Registration**: Each worktree Claude session registers with unique agent ID
- **Cross-Session Communication**: @mentions, documents, notifications across worktrees
- **Role-Based Context**: Dynamic roles based on worktree investigation type

## Implementation
### Approach
**MCP-First Integration**: Use Headless PM's existing MCP server as communication backbone without changing core ORC workflow patterns.

Each worktree's Claude session:
1. **Auto-registers** with HPM using worktree-derived agent ID
2. **Declares role** based on investigation type (backend_dev, qa, frontend_dev, etc.)
3. **Maintains connection** to communication bus throughout session
4. **Preserves local focus** while enabling cross-worktree coordination

### MCP Integration Points

#### Agent Registration (Per Worktree)
```python
# Embedded in each worktree's CLAUDE.md
agent_id = f"ml-{worktree_name}"
role = detect_role_from_worktree_context()  # backend_dev, qa, frontend_dev, etc.
client = claude_register(agent_id, role, "senior")
```

#### Communication Commands (Natural Language via MCP)
- **"Share findings with other investigations"** → Creates HPM document with @mentions
- **"Check for notifications from other worktrees"** → Polls HPM for @mentions  
- **"Alert all sessions about security issue"** → Broadcasts to all registered agents
- **"What are other investigations working on?"** → Query active agent status

#### Context Bridging
- Cross-worktree dependency notification
- Shared findings repository
- Coordinated investigation efforts
- Automatic blocker alerts

## Testing Strategy

### Phase 1 Validation
1. **Setup HPM server** locally alongside existing ORC
2. **Test MCP integration** with single worktree Claude session
3. **Verify registration** and basic communication functionality
4. **Validate TMux integration** doesn't interfere with existing workflow

### Phase 2 Multi-Session Testing  
1. **Two worktree setup** with different investigation contexts
2. **Cross-session @mentions** and document sharing
3. **Role-based task assignment** and coordination
4. **Context preservation** during communication events

### Phase 3 Production Integration
1. **Migration of active investigations** to HPM communication bus
2. **Performance testing** with multiple concurrent sessions
3. **Workflow optimization** based on real usage patterns

## Implementation Plan

### Phase 1: HPM Setup and MCP Integration (Week 1-2)
- **Install Headless PM** server in ORC ecosystem
- **Configure MCP server** for Claude Code integration
- **Test basic agent registration** and communication
- **Document integration patterns** for worktree contexts

### Phase 2: CLAUDE.md Enhancement (Week 2-3)  
- **Update worktree templates** with HPM agent registration
- **Add communication commands** to standard Claude context
- **Implement role detection** based on worktree patterns
- **Create notification polling** integration

### Phase 3: ORC Command Integration (Week 3-4)
- **New /hpm-* commands** for communication management
- **Enhanced /bootstrap** with HPM agent setup
- **Modified /janitor** for HPM connection cleanup
- **TMux integration hooks** for automatic agent registration

### Phase 4: Advanced Coordination (Week 4-6)
- **Cross-worktree context queries** via MCP tools
- **Dependency tracking** and blocker notifications  
- **Shared investigation findings** repository
- **Performance optimization** and usage analytics

## Technical Considerations

### Worktree Agent ID Strategy
```
Pattern: ml-{investigation-name}-{primary-role}
Examples:
- ml-auth-api-backend-dev
- ml-perf-investigation-qa  
- ml-ui-redesign-frontend-dev
- ml-security-audit-architect
```

### Role Detection Logic
```python
def detect_role_from_context():
    if "api" in worktree_name or "backend" in tech_plans:
        return "backend_dev"
    elif "ui" in worktree_name or "frontend" in tech_plans:
        return "frontend_dev" 
    elif "test" in worktree_name or "debug" in worktree_name:
        return "qa"
    else:
        return "backend_dev"  # default
```

### MCP Configuration
```json
{
  "mcpServers": {
    "headless-pm": {
      "command": "python",
      "args": ["-m", "headless_pm_mcp_server"],
      "env": {
        "HEADLESS_PM_URL": "http://localhost:6969"
      }
    }
  }
}
```

## Benefits Analysis

### Preserved ORC Strengths
- ✅ TMux pane organization unchanged
- ✅ Worktree isolation and focus maintained  
- ✅ Individual Claude sessions per investigation
- ✅ Tech plans symlink architecture preserved
- ✅ Existing /bootstrap, /janitor, /muxup workflows

### New Communication Capabilities
- ✅ Cross-worktree findings sharing
- ✅ Dependency and blocker coordination
- ✅ Security alert broadcasting
- ✅ Investigation status visibility
- ✅ Natural language coordination via MCP

### Risk Mitigation
- **Incremental rollout** preserves existing workflow
- **Optional communication layer** - core ORC functions independently
- **Graceful degradation** if HPM server unavailable
- **No architectural changes** to proven TMux/worktree patterns

## Success Metrics

### Week 2: Basic Integration
- HPM server running stable alongside ORC
- Single worktree Claude session successfully registers and communicates
- MCP tools accessible via natural language commands

### Week 4: Multi-Session Coordination  
- Multiple worktree sessions coordinate via @mentions
- Cross-context dependency notification working
- No disruption to existing investigation workflows

### Week 6: Production Ready
- All active investigations using HPM communication bus
- Measurable improvement in cross-investigation coordination
- Documentation and templates updated for new workflow patterns
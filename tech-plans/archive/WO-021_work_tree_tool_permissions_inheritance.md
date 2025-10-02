# Work Order #021: Work Tree Tool Permissions Inheritance

**Created**: 2025-08-21  
**Category**: 🛠️ Tooling  
**Priority**: Medium  
**Effort**: S  
**IMP Assignment**: Unassigned

## Problem Statement

Currently, every new work tree requires manual tool permission granting to IMP Claudes, resulting in repeated "Yes, you can use env/Bash/etc." conversations that break workflow momentum. This creates friction in the Forest Factory system where IMPs should have immediate access to the tools they need for implementation work.

**Current Pain Point**: El Presidente must manually grant tool permissions (env, Bash, file operations, etc.) in every new work tree, despite having global tool allowlists configured. This suggests either global tool configuration isn't being inherited properly or Claude's tool permission system has changed recently.

**Goal**: Implement automatic tool permissions inheritance so that all work tree IMPs have immediate access to the same toolset as the orchestrating ORC, eliminating permission friction and enabling immediate productive work.

## Acceptance Criteria

### Phase 1: Global Tool Configuration Analysis
- [ ] **Current Configuration Audit**: Review El Presidente's global Claude tool allowlists and settings
- [ ] **Inheritance Investigation**: Determine why global tools aren't being applied to work tree sessions
- [ ] **Claude Changes Analysis**: Research if recent Claude updates changed tool permission inheritance behavior
- [ ] **Permission Flow Mapping**: Document how tool permissions should flow from global → work tree contexts

### Phase 2: Tool Permissions Template System
- [ ] **Work Tree Template Enhancement**: Add tool permissions specification to work tree CLAUDE.md templates
- [ ] **Standard Toolset Definition**: Define standard IMP toolset (Bash, env, file operations, etc.)
- [ ] **Permission Documentation**: Clear specification of what tools IMPs need and why
- [ ] **Inheritance Mechanism**: Implement automatic tool permission setup in work tree creation

### Phase 3: Seamless Work Tree Creation
- [ ] **ORC Work Tree Setup**: Modify work tree creation process to include tool permissions
- [ ] **IMP Onboarding**: Ensure IMPs immediately have access to required tools upon work tree activation
- [ ] **Permission Validation**: Test that tool permissions work correctly in new work trees
- [ ] **Documentation Update**: Update work tree templates and procedures with tool inheritance

### Phase 4: Permission Learning System
- [ ] **Worktree Cleanup Enhancement**: During worktree cleanup, ORC captures any permissions granted to IMP agents
- [ ] **Global Permission Aggregation**: Add discovered permissions to global Claude configuration
- [ ] **Permission Evolution**: System continuously enhances itself with newly granted permissions
- [ ] **Audit Trail**: Track what permissions were added and when for troubleshooting

## Technical Context

**Current Tool Permission Issues**:
- Global Claude tool configuration not inheriting to work tree sessions
- Manual permission granting required for basic tools (env, Bash, file ops)
- Workflow friction when IMPs need immediate tool access for implementation
- Possible recent changes to Claude's tool permission system

**Expected Standard IMP Toolset**:
- **Bash**: Command execution, git operations, build/test commands
- **Environment Variables**: Access to env for configuration and debugging
- **File Operations**: Read, Edit, Write, Glob for codebase modification
- **Repository Tools**: Git operations, GitHub CLI access
- **Development Tools**: Package managers, testing frameworks, linting tools

**Global Configuration Locations**:
- El Presidente's global Claude settings and tool allowlists
- Work tree CLAUDE.md templates and tool specifications
- Session initialization and permission inheritance mechanisms

## Resources & References

- **Global Claude Configuration**: El Presidente's current tool allowlist settings
- **Work Tree Templates**: Current CLAUDE.md template patterns in `/orc/templates/`
- **Session Management**: How Claude Code handles tool permissions across contexts
- **Recent Changes**: Investigation into Claude platform changes affecting tool inheritance

## Implementation Notes

**Investigation Areas**:
1. **Global Settings Review**: Check current global Claude tool configuration
2. **Inheritance Testing**: Test if global tools should inherit to new sessions
3. **Template Enhancement**: Add explicit tool permissions to work tree CLAUDE.md
4. **ORC Process Update**: Modify work tree creation to include tool setup

**Expected Tool Permission Pattern**:
```markdown
# In work tree CLAUDE.md template:
## Available Tools
You have access to the following tools for implementation work:
- **Bash**: Execute commands, git operations, build/test processes
- **Environment**: Access env variables for configuration and debugging  
- **File Operations**: Read, Edit, Write, Glob for codebase modification
- **Repository Tools**: Git commands, GitHub CLI, package managers
```

**ORC Workflow Enhancement**:
- Work tree creation automatically includes tool permission specification
- IMPs receive clear tool access documentation upon worktree activation
- No manual permission granting required for standard development tools
- **Permission Learning**: During worktree cleanup, ORC captures any permissions granted to IMP and adds to global allowlist
- **Self-Enhancement**: System evolves its tool permissions automatically based on actual usage patterns

**Success Metrics**:
- Zero manual tool permission grants required in new work trees
- IMPs immediately productive with full toolset access
- Clear documentation of tool permissions in all work tree templates
- Reduced friction in Forest Factory workflow initiation

---

## Work Order Lifecycle

### Status History
- **2025-08-21**: Created → 02-NEXT (workflow enhancement ready for implementation)

### IMP Notes
**Status**: 📅 **NEXT** - Tool permissions inheritance system ready for implementation

**Problem Context**: El Presidente experiencing repeated tool permission friction in work trees, suggesting global tool configuration isn't inheriting properly or Claude's permission system has changed recently.

**Implementation Scope**: 
1. Investigate current global tool configuration and inheritance behavior
2. Enhance work tree templates with explicit tool permissions
3. Modify ORC work tree creation process to include automatic tool setup
4. Test and validate seamless tool access in new work trees

**Expected Outcome**: 
- Eliminated manual tool permission granting in work trees
- IMPs immediately productive with full development toolset
- Streamlined Forest Factory workflow with zero permission friction

**Investigation Priority**: Determine if this is global configuration issue, Claude platform change, or template enhancement need

**Next Steps**: 
1. Audit El Presidente's current global Claude tool configuration
2. Test tool permission inheritance in new Claude sessions
3. Research recent Claude platform changes affecting tool permissions
4. Design enhanced work tree template with explicit tool specifications

---
*Work Order #021 - Forest Manufacturing System*
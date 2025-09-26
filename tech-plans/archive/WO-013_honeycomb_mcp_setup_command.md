# Work Order #013: Honeycomb MCP Setup Claude Command

**Created**: 2025-08-18  
**Category**: 🛠️ Tooling  
**Priority**: Medium  
**Effort**: M  
**IMP Assignment**: Unassigned

## Problem Statement

Engineers need easy access to Honeycomb observability data through Claude for DLQ investigation, performance analysis, and operational troubleshooting. Currently, setting up the Honeycomb MCP server requires manual configuration of global Claude settings, API token management, and understanding of MCP server architecture.

This creates a barrier to adoption and limits the team's ability to leverage Claude for systematic data analysis and investigation workflows.

We need a self-service Claude command that any engineer can run to automatically configure Honeycomb MCP server access without requiring deep knowledge of MCP configuration or access to El Presidente's spellbook system.

## Acceptance Criteria

### Phase 1: Core Setup Command
- [ ] **Claude Command Creation**: Create `.claude/commands/setup-honeycomb-mcp.md` in intercom repo
- [ ] **Global Config Integration**: Command modifies `~/.claude.json` to add Honeycomb MCP server configuration
- [ ] **API Token Management**: Secure handling of Honeycomb API tokens (likely from existing env vars or secure storage)
- [ ] **Endpoint Configuration**: Automatic setup of correct Honeycomb server endpoints and parameters

### Phase 2: User Experience
- [ ] **Self-Service Execution**: Any engineer can run command without additional setup knowledge
- [ ] **Validation and Testing**: Command includes verification steps to confirm MCP server is working
- [ ] **Clear Documentation**: Complete usage instructions and troubleshooting guidance
- [ ] **Error Handling**: Graceful handling of common setup failures (missing tokens, config conflicts)

### Phase 3: Team Adoption Support
- [ ] **Usage Examples**: Sample Honeycomb queries and investigation patterns
- [ ] **Integration Documentation**: How to use Honeycomb MCP in investigation workflows
- [ ] **Rollout Strategy**: Communication and adoption plan for team members
- [ ] **Support Documentation**: Troubleshooting and maintenance guidance

## Technical Context

**MCP Server Configuration Requirements**:
- Global Claude config at `~/.claude.json` (not `~/.claude/settings.json`)
- Honeycomb MCP server endpoints and authentication
- Node.js based MCP server (likely `@modelcontextprotocol/server-honeycomb` or similar)

**Dependencies**: 
- Understanding of El Presidente's spellbook MCP configuration procedures
- Access to Honeycomb API token management (likely env vars like `HONEYCOMB_DEVELOPMENT_KEY`)
- Knowledge of team's Honeycomb datasets and common investigation patterns
- Integration with existing Claude command structure in intercom repo

**Repositories**: 
- **intercom**: Primary location for Claude command (`.claude/commands/`)
- **Reference**: El Presidente's spellbook MCP configuration spells for implementation guidance

**Complexity Notes**: 
- Must work independently of spellbook access (self-contained)
- Secure API token handling without exposing credentials
- Cross-platform compatibility for team members' different setups
- Integration with existing MCP servers without conflicts

## Resources & References

- **Spellbook Reference**: `orc/spellbook/integrations/mcp-server-setup.md` - Implementation guidance
- **Global CLAUDE.md**: Current MCP server configuration examples and patterns
- **Existing MCP Setup**: El Presidente's working Honeycomb MCP configuration as template
- **Team Adoption**: Target audience is all Infrastructure Services Group engineers

## Implementation Notes

**Command Structure**:
```markdown
# Setup Honeycomb MCP Server

## Purpose
Automatically configure Honeycomb MCP server for Claude Code integration

## Prerequisites
- Claude Code installed and configured
- Access to Honeycomb development environment
- [Specific token/access requirements]

## Usage
[Step-by-step execution instructions]

## Verification
[How to test the setup worked correctly]

## Troubleshooting
[Common issues and solutions]
```

**Key Implementation Areas**:
1. **JSON Configuration**: Modify `~/.claude.json` safely without breaking existing setup
2. **Token Management**: Secure API token handling and validation
3. **Server Setup**: MCP server installation and configuration
4. **Testing Integration**: Built-in verification that setup is working correctly

**Success Metrics**:
- Any team member can run command and get working Honeycomb MCP access
- Command is self-contained and doesn't require spellbook knowledge
- Setup process is reliable and includes proper error handling
- Team adoption increases access to observability data for investigations

---

## Work Order Lifecycle

### Status History
- **2025-08-18**: Created → 03-IN-PROGRESS (immediate implementation priority)

### IMP Notes
**Status**: 🔄 **IN-PROGRESS** - Self-service MCP setup command development

**Immediate Priority**: Create Claude command that enables any engineer to set up Honeycomb MCP server access without requiring El Presidente's spellbook system.

**Key Implementation Focus**:
- Self-contained command in intercom repo `.claude/commands/`
- Safe modification of global Claude configuration
- Secure API token handling and validation
- Built-in verification and troubleshooting guidance

**Expected Outcome**: 
- Team-wide access to Honeycomb observability data through Claude
- Reduced setup friction for investigation and analysis workflows
- Foundation for broader MCP server adoption across the team

**Target Users**: All Infrastructure Services Group engineers

**Next Steps**: 
1. Review spellbook MCP configuration procedures for implementation guidance
2. Design command structure and user experience flow
3. Implement safe global configuration modification
4. Add verification and error handling
5. Test with team members for usability validation

---
*Work Order #013 - Forest Manufacturing System*
# Work Order #020: Claude Code DLQ Investigation Workflow

**Created**: 2025-08-21  
**Category**: 🤖 Automation  
**Priority**: Medium  
**Effort**: L  
**IMP Assignment**: Unassigned

## Problem Statement

Building on the foundation of automated DLQ issue creation (WO-019), we need to implement the automated investigation workflow using Claude Code GitHub Actions. This completes the DLQ automation pipeline by enabling AI-powered investigation and remediation directly within GitHub issues.

**Dependency**: WO-019 (CloudBot DLQ Issue Creation) must be completed first to provide the GitHub issues that trigger this investigation workflow.

**Architecture Evolution**: Following the architectural decision from WO-012, we're implementing investigation automation via "Claude Code GitHub Actions workflow" rather than the original "Refactorer MCP server" approach.

## Acceptance Criteria

### Phase 1: GitHub Actions Integration
- [ ] **Claude Code Workflow**: Configure GitHub Actions workflow to trigger on DLQ investigation issue creation
- [ ] **Issue Detection**: Automatically detect newly created DLQ investigation issues via labels/templates
- [ ] **Investigation Trigger**: Launch Claude Code investigation using existing proven `investigate-dlq` methodology
- [ ] **Honeycomb Access**: Ensure Claude has Honeycomb MCP access for systematic diagnostic queries

### Phase 2: Automated Investigation
- [ ] **Diagnostic Analysis**: Leverage Honeycomb MCP for systematic telemetry analysis
- [ ] **Root Cause Identification**: Automated analysis of queue patterns, error rates, and system health
- [ ] **Code Review**: Relevant application code reading to understand processing logic
- [ ] **Investigation Documentation**: Structured findings and analysis added to GitHub issue

### Phase 3: Remediation and PR Generation
- [ ] **Remediation Planning**: Develop fix strategies based on investigation findings
- [ ] **PR Creation**: Generate pull requests with fixes when clear remediation is identified
- [ ] **Issue Updates**: Update investigation issue with findings, PRs, and next steps
- [ ] **Slack Notifications**: Optional Slack thread updates when investigation completes

## Technical Context

**Foundation Dependencies**:
- **WO-012**: ✅ CloudBot message detection infrastructure complete
- **WO-019**: 🔄 CloudBot DLQ issue creation (prerequisite)
- **WO-013**: ✅ Honeycomb MCP setup command available for team

**Repository**: intercom (GitHub Actions workflow and investigation commands)

**Integration Architecture**:
- GitHub Actions triggered by DLQ investigation issue creation
- Claude Code execution with Honeycomb MCP access
- Existing `investigate-dlq` command methodology as foundation
- PR generation and issue updates for complete workflow

**Workflow Trigger**:
```yaml
# Expected GitHub Actions trigger
on:
  issues:
    types: [opened, labeled]
    # Trigger when DLQ investigation issues are created
```

## Resources & References

- **WO-012 Foundation**: Completed CloudBot infrastructure and architectural decisions
- **WO-013 Reference**: Honeycomb MCP setup command for team access patterns
- **Existing Methodology**: Proven `investigate-dlq` command in intercom repository
- **GitHub Actions Experience**: Existing Claude Code GitHub Actions integration patterns

## Implementation Notes

**GitHub Actions Workflow Design**:
1. **Issue Detection**: Filter for DLQ investigation issues (labels: "dlq-alarm", "automated")
2. **Claude Code Execution**: Launch investigation using existing proven methodology
3. **Honeycomb Integration**: Leverage MCP server for systematic diagnostic analysis
4. **Results Documentation**: Update issue with findings and recommended actions

**Investigation Process**:
```bash
# Expected investigation flow
claude investigate-dlq --queue=[queue-name] --region=[region] --issue=[issue-number]
```

**Claude Code Capabilities**:
- Honeycomb MCP access for telemetry analysis
- Codebase reading for understanding processing logic
- GitHub API access for issue updates and PR creation
- Slack integration for optional status notifications

**Success Metrics**:
- DLQ investigation issues automatically processed by Claude Code
- Systematic diagnostic analysis using Honeycomb data
- Investigation findings documented in GitHub issues
- PR generation for clear remediation opportunities
- Complete automation pipeline from alarm to investigation

**Architectural Benefits**:
- Leverages existing GitHub Actions infrastructure
- Utilizes proven `investigate-dlq` methodology
- Integrates with team's existing Claude Code workflows
- Maintains familiar GitHub-centric operational model

---

## Work Order Lifecycle

### Status History
- **2025-08-21**: Created → 02-NEXT (dependent on WO-019 completion)

### IMP Notes
**Status**: 📅 **NEXT** - Claude Code investigation workflow ready after WO-019

**Dependency**: WO-019 (CloudBot DLQ Issue Creation) must complete first to provide GitHub issues that trigger this investigation workflow.

**Architecture Foundation**: Based on WO-012 architectural decision to use "Claude Code GitHub Actions workflow" instead of "Refactorer MCP server" approach.

**Implementation Scope**: GitHub Actions workflow enhancement to automate DLQ investigation:
1. Detect newly created DLQ investigation issues
2. Launch Claude Code with existing `investigate-dlq` methodology
3. Leverage Honeycomb MCP for systematic analysis
4. Document findings and generate remediation PRs

**Expected Outcome**: 
- Complete automation from DLQ alarm to investigation completion
- AI-powered diagnostic analysis using proven methodology
- PR generation for identified remediation opportunities
- End-to-end audit trail from alarm detection to fix implementation

**Integration Points**: GitHub Actions, Claude Code, Honeycomb MCP, existing investigation commands

**Next Steps**: 
1. Wait for WO-019 completion (CloudBot issue creation)
2. Design GitHub Actions workflow for issue detection
3. Integrate existing `investigate-dlq` methodology with automation
4. Test end-to-end investigation and remediation workflow

---
*Work Order #020 - Forest Manufacturing System*
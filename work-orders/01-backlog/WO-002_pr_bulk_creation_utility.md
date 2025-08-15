# Work Order #002: PR Bulk Creation Utility

**Created**: 2025-08-15  
**Category**: 🛠️ Tooling  
**Priority**: Medium  
**Effort**: L  
**IMP Assignment**: Unassigned

## Problem Statement

Creating PRs for multi-repository features requires manually running `pro` (GitHub CLI) in each repository directory. This is time-consuming and error-prone for worktrees spanning 2-3 repositories. Need automated PR creation across all repositories with changes in the current worktree.

## Acceptance Criteria

- [ ] `wtpr-create` command creates PRs for all repositories with uncommitted changes
- [ ] Automatically generates appropriate PR titles based on branch names or commits
- [ ] Uses consistent PR description template across repositories
- [ ] Handles repositories with no changes gracefully (skip PR creation)
- [ ] Integrates with existing `pro` command and GitHub CLI workflow
- [ ] Links related PRs together in descriptions for cross-repository context

## Technical Context

**Dependencies**: 
- GitHub CLI (`gh`) and `pro` alias functionality
- Git repository detection and change identification
- Access to commit messages for PR title/description generation
- Understanding of current PR creation workflow

**Repositories**: 
- Shell scripting integration with existing git aliases
- GitHub CLI API usage for PR creation and linking

**Complexity Notes**: 
- Must handle authentication and permissions across repositories
- Need to generate meaningful PR titles without manual input
- Cross-repository PR linking requires GitHub API understanding
- Error handling for failed PR creation (branch conflicts, etc.)

## Resources & References

- Current `pro` alias implementation and GitHub CLI usage
- Existing git workflow: `publish` → `pro` → `pru` → `prfeed`
- Multi-repository feature patterns and PR relationships

## Implementation Notes

Should integrate with existing workflow:
1. **Pre-checks**: Verify all repos have publishable changes
2. **Bulk Publishing**: Use existing `wtpublish` pattern to push all branches  
3. **PR Creation**: Create PRs with consistent naming and cross-references
4. **Link Generation**: Add "Related PRs" section to each PR description

Example cross-repository PR linking:
```markdown
## Related PRs
- Infrastructure changes: intercom/infrastructure#12345
- Backend changes: intercom/intercom#67890
```

Consider integration with `pru` command for bulk PR description updates.

---

## Work Order Lifecycle

### Status History
- **2025-08-15**: Created → 01-BACKLOG

### IMP Notes
*Space for assigned IMP to add progress notes, blockers, discoveries*

---
*Work Order #002 - Forest Manufacturing System*
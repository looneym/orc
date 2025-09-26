# Work Order #016: GitHub Workflow Bash Permissions Fix

**Created**: 2025-08-20  
**Category**: 🔧 Infrastructure  
**Priority**: High  
**Effort**: XS  
**IMP Assignment**: Unassigned

## Problem Statement

Claude's GitHub Action workflow needs bash permissions to function properly, but the GitHub App is blocked from making workflow modifications due to security restrictions. The workflow file `.github/workflows/trigger-claude.yml` needs to be manually updated to add bash tool permissions.

**GitHub Issue**: [#426109](https://github.com/intercom/intercom/issues/426109#issuecomment-3206996731)

**Current Block**: Claude successfully modified the workflow locally and committed changes, but cannot push workflow modifications due to GitHub App security restrictions.

## Acceptance Criteria

### Phase 1: Workflow Configuration Fix
- [ ] **Edit Workflow File**: Modify `.github/workflows/trigger-claude.yml` to add bash permissions
- [ ] **Add Tool Configuration**: Include `allowed_tools: "Bash"` parameter in workflow configuration
- [ ] **Commit and Push**: Apply changes to repository to enable Claude's bash access
- [ ] **Verify Functionality**: Test that Claude can now execute bash commands in GitHub Actions

### Phase 2: Validation
- [ ] **Test Claude Access**: Verify Claude has full bash access for repository operations
- [ ] **Document Change**: Ensure workflow modification is properly documented
- [ ] **Security Review**: Confirm change aligns with repository security policies

## Technical Context

**File to Modify**: `.github/workflows/trigger-claude.yml`

**Required Change**:
```yaml
jobs:
  call-claude:
    uses: intercom/github-action-workflows/.github/workflows/claude-workflow.yml@main
    with:
      allowed_tools: "Bash"
    secrets: inherit
```

**Current Issue**: Claude committed this change locally but cannot push due to GitHub App workflow permission restrictions.

**Security Note**: GitHub Apps are restricted from modifying workflow files to prevent security vulnerabilities.

## Implementation Notes

**Manual Steps Required**:
1. Edit `.github/workflows/trigger-claude.yml` in the repository
2. Add the `with:` section containing `allowed_tools: "Bash"` parameter
3. Commit and push the change to enable Claude's bash functionality

**Alternative Approach**: Could grant GitHub App `workflows` permission, but manual edit is more secure.

**Expected Outcome**: Once merged, Claude will have full bash access for:
- Repository cloning and navigation
- Command execution in GitHub Actions
- Enhanced functionality for complex tasks

**Benefits**:
- Enables Claude to perform more sophisticated repository operations
- Allows for automated testing and validation workflows
- Improves Claude's effectiveness in GitHub Action contexts

---

## Work Order Lifecycle

### Status History
- **2025-08-20**: Created → 03-IN-PROGRESS (immediate manual fix required)

### IMP Notes
**Status**: ✅ **COMPLETE** - GitHub workflow bash permissions successfully implemented

**Completed Actions**:
- **2025-08-22**: Manual workflow modification completed by El Presidente
- **Workflow Updated**: `.github/workflows/trigger-claude.yml` now includes `allowed_tools: "Bash"` configuration
- **Security Compliance**: Change applied through manual process due to GitHub App workflow restrictions

**Final Configuration Verified**:
```yaml
jobs:
  call-claude:
    uses: intercom/github-action-workflows/.github/workflows/claude-workflow.yml@main
    with:
      allowed_tools: "Bash"
    secrets: inherit
```

**Implementation Complete**: 
- ✅ Claude now has full bash access in GitHub Actions
- ✅ Workflow file properly configured with tool permissions
- ✅ Security policy compliance maintained through manual process
- ✅ Enhanced Claude functionality operational for repository operations

**Mission Complete**: GitHub workflow now enables Claude's bash command execution, removing previous functionality limitations in GitHub Action contexts.

---
*Work Order #016 - Forest Manufacturing System*
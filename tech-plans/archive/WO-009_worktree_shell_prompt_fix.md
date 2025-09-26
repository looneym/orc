# Work Order #009: Work Tree Shell Prompt Fix

**Created**: 2025-08-15  
**Category**: 🐛 Bug Fix  
**Priority**: High  
**Effort**: S  
**IMP Assignment**: Immediate (blocking work tree usage)

## Problem Statement

Work tree shell prompts are displaying error: `prompt_lambda:8: command not found: get_claude_status`. This is caused by changes made to shell profile inside work trees that reference a function that doesn't exist or isn't properly loaded. This is blocking effective use of work trees for development.

**Impact**: Cannot use work trees effectively without shell prompt errors appearing constantly.

## Acceptance Criteria

- [ ] **Error Elimination**: Remove `command not found: get_claude_status` error from work tree shell prompts
- [ ] **Proper Function Loading**: Ensure `get_claude_status` function is available in work tree environments
- [ ] **Prompt Functionality**: Work tree prompt displays correct information (user, worktree, branch, repos)
- [ ] **No Regression**: Standard shell prompts outside work trees continue working
- [ ] **All Work Trees**: Fix applies to all existing work trees consistently

## Technical Context

**Current Error**: `prompt_lambda:8: command not found: get_claude_status`

**Prompt Elements Working**:
- ✅ User: `looneym@moneymaker`
- ✅ Work tree path: `~/src/worktrees/ml-dlq-alarm-review-automation/intercom`  
- ✅ Branch: `ml/intercom-dlq-alarm-review-automation`
- ✅ Repos: `infrastructure, intercom`

**Issue**: `get_claude_status` function not found/loaded

**Dependencies**: 
- Custom dotfiles configuration
- Shell profile setup in work tree environments
- ZSH prompt configuration system

**Repositories**: 
- Likely custom dotfiles repository or shell configuration files
- May need updates to work tree setup process

## Implementation Strategy

**Investigation Areas**:
1. **Function Definition**: Locate where `get_claude_status` should be defined
2. **Loading Issue**: Check if function isn't being sourced in work tree environments
3. **Path Problems**: Verify dotfiles/shell config paths in work tree contexts
4. **Missing Dependency**: Function may depend on tools not available in work trees

**Debugging Steps**:
1. **Check Function**: `type get_claude_status` in work tree vs normal shell
2. **Source Analysis**: Review which shell config files are loaded in work trees
3. **Dotfiles Review**: Check custom dotfiles for `get_claude_status` definition
4. **Fix Application**: Either fix function loading or remove reference from prompt

**Likely Solutions**:
- Add `get_claude_status` function definition to work tree shell config
- Fix sourcing of dotfiles in work tree environments  
- Remove broken `get_claude_status` reference if function is obsolete
- Update work tree setup to properly load all needed shell functions

## Resources & References

- **Error Location**: `prompt_lambda:8` - line 8 of prompt_lambda function/file
- **Work Tree Context**: All work trees showing this error
- **Custom Dotfiles**: User's shell configuration system

## Risk Assessment

**High Priority Fix**:
- Blocking effective work tree usage
- Constant error messages disrupt workflow
- Simple shell configuration issue with straightforward fix

**Low Risk**:
- Shell prompt fixes typically safe
- Can test changes in single work tree first
- Easy to revert if changes cause issues

---

## Work Order Lifecycle

### Status History
- **2025-08-15**: Created as immediate fix → 03-IN-PROGRESS (blocking work tree usage)

### IMP Notes
**Status**: ✅ **COMPLETE** - Shell prompt error eliminated, work trees functional

**Key Actions Completed**:
- **2025-08-18**: Identified error in dotfiles/themes/looneym.zsh-theme:146
- **2025-08-18**: IMP removed `get_claude_status` function call from prompt_lambda
- **2025-08-18**: Verified fix resolves command not found error
- **2025-08-18**: Work trees now display clean multi-line prompts without errors

**Final Status**: Work tree shell prompts working correctly, blocking issue resolved

---
*Work Order #009 - Forest Manufacturing System*
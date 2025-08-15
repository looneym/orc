# Work Order #001: Worktree Status Command

**Created**: 2025-08-15  
**Category**: 🛠️ Tooling  
**Priority**: High  
**Effort**: M  
**IMP Assignment**: Unassigned

## Problem Statement

Need a unified command to show git status across all repositories in the current worktree, eliminating the need to manually check each repo directory individually. This will streamline the development workflow by providing instant visibility into the state of all repos in a multi-repository feature.

## Acceptance Criteria

- [ ] `wtstatus` command shows git status for all repositories in current worktree
- [ ] Output clearly identifies each repository and its status
- [ ] Command works from any directory within the worktree
- [ ] Performance optimized - uses efficient git commands
- [ ] Handles repositories with no changes gracefully
- [ ] Color-coded output for easy scanning (clean, dirty, untracked files)

## Technical Context

**Dependencies**: 
- ZSH prompt worktree detection logic (already implemented by IMP-ZSH)
- Git repository identification within worktree directories
- Efficient git status checking (avoid slow `git status` calls)

**Repositories**: 
- Likely implemented as shell function/script in dotfiles
- May integrate with existing ZSH theme functions

**Complexity Notes**: 
- Must handle multiple git repositories efficiently
- Need to differentiate between clean, dirty, and untracked states
- Should work across different worktree structures (2-repo, 3-repo, etc.)

## Resources & References

- IMP-ZSH implementation of worktree detection in ZSH prompt
- Existing git performance optimizations using `git diff-index` and `git ls-files`
- Current worktree structure patterns

## Implementation Notes

Consider leveraging the optimized git status checking already implemented in the ZSH prompt:
- `git diff-index --quiet HEAD --` for staged/unstaged changes
- `git ls-files --others --exclude-standard` for untracked files

Output format should be scannable and actionable, possibly showing:
```
worktree: ~/src/worktrees/ml-feature-name
├── intercom: clean ✓
├── infrastructure: dirty (2 modified, 1 untracked) ⚠️  
└── muster: clean ✓
```

---

## Work Order Lifecycle

### Status History
- **2025-08-15**: Created → 01-BACKLOG

### IMP Notes
*Space for assigned IMP to add progress notes, blockers, discoveries*

---
*Work Order #001 - Forest Manufacturing System*
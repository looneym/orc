# Forest Backlog - Future Work Items

## Extracted from IMP-ZSH Grove (2025-08-15)

**Source**: IMP-ZSH completed ZSH prompt improvements and identified comprehensive worktree management utilities

### Status & Information Commands

#### `wtstatus` - Multi-Repository Status View
**Description**: Show git status for all repositories in current worktree
**Value**: Quick overview of all repo states in one command
**Complexity**: Medium (requires git status aggregation across repos)
**Dependencies**: Existing worktree detection logic from ZSH prompt

#### `wtdiff` - Multi-Repository Diff View  
**Description**: Display git diff for all repositories in current worktree
**Value**: See all changes across worktree in unified view
**Complexity**: Medium (git diff aggregation, potentially large output)
**Dependencies**: Repository detection, git diff formatting

#### `wtinfo` - Investigation Status Display
**Description**: Show current investigation status from CLAUDE.md
**Value**: Quick check on current grove status without file reading
**Complexity**: Low (file parsing, status extraction)
**Dependencies**: CLAUDE.md format standards

#### `wtrepos` - Repository Listing with Status
**Description**: List all available repositories with their current branch and status
**Value**: High-level worktree overview with actionable information
**Complexity**: Medium (combines repo detection with git status)
**Dependencies**: Git branch detection, status checking

### PR & Review Management Commands

#### `wtprs` - Multi-Repository PR Listing
**Description**: List all open PRs for repositories in current worktree
**Value**: Central view of all worktree-related review activity
**Complexity**: High (requires GitHub API integration)
**Dependencies**: GitHub CLI, PR-branch correlation logic

#### `wtpr-create` - Bulk PR Creation
**Description**: Create PRs for all repositories with changes in current worktree
**Value**: Streamline PR creation for multi-repo features
**Complexity**: High (automated PR creation, description generation)
**Dependencies**: GitHub CLI, branch publishing, PR templates

#### `wtpr-update` - PR Description Synchronization
**Description**: Update PR descriptions from latest commit messages across repos
**Value**: Keep PRs in sync with evolving commit messages
**Complexity**: High (GitHub API, commit message parsing)
**Dependencies**: GitHub CLI, PR identification, commit history

#### `wtpr-status` - PR Review Dashboard
**Description**: Show PR status, review status, and CI status for worktree PRs
**Value**: Comprehensive worktree delivery tracking
**Complexity**: High (GitHub API, CI system integration)
**Dependencies**: GitHub CLI, CI status APIs, review state tracking

### Git Operations Commands

#### `wtcommit` - Multi-Repository Commit
**Description**: Commit changes across all dirty repositories with same message
**Value**: Unified commit workflow for related changes
**Complexity**: Medium (multi-repo git operations, error handling)
**Dependencies**: Repository state detection, git commit coordination

#### `wtpublish` - Multi-Repository Publishing
**Description**: Publish (push) all repositories in worktree to remote
**Value**: One-command publishing for entire feature
**Complexity**: Medium (git push coordination, upstream management)
**Dependencies**: Branch tracking, remote configuration

#### `wtresync` - Multi-Repository Resync
**Description**: Resync all repositories in worktree with their remote branches
**Value**: Handle "stale info" errors across entire worktree
**Complexity**: High (rebase coordination, conflict resolution)
**Dependencies**: Git resync logic, error handling

#### `wtclean` - Worktree Cleanup
**Description**: Clean up worktree - remove merged branches, prune references
**Value**: Maintain clean worktree state post-completion
**Complexity**: Medium (git cleanup operations, safety checks)
**Dependencies**: Branch merge detection, reference pruning

### Investigation Management Commands

#### `wtnote` - Progress Note Addition
**Description**: Add timestamped note to CLAUDE.md progress section
**Value**: Quick progress logging without file editing
**Complexity**: Low (file manipulation, timestamp generation)
**Dependencies**: CLAUDE.md format standards, file editing

#### `wtstatus-update` - Status Section Updates
**Description**: Update CLAUDE.md status section with new progress
**Value**: Maintain accurate status without manual file editing
**Complexity**: Medium (status parsing, section replacement)
**Dependencies**: CLAUDE.md status format, section identification

#### `wtsummary` - Work Summary Generation
**Description**: Generate summary of work completed in current worktree
**Value**: Quick completion reporting for reviews/handoffs
**Complexity**: High (git history analysis, commit summarization)
**Dependencies**: Git history, CLAUDE.md parsing, summary generation

---

**Total Items**: 12 commands across 4 categories
**Source Grove**: ml-zsh-prompt-improvements (IMP-ZSH)
**Extraction Date**: 2025-08-15
**Status**: Backlog (not yet prioritized or assigned)
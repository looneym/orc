# Squash Branch Command

Transform messy development branches with multiple commits into clean, single-commit release branches ready for merge.

## Role

You are a **Git Branch Squashing Specialist** that expertly consolidates multiple development commits into clean, professional release branches. You understand merge strategies, commit verification, and create comprehensive commit messages that capture all improvements in a structured format.

## Usage

```
/squash-branch [--suffix <suffix>] [--base <base-branch>]
```

**Purpose**: Take a working development branch with many incremental commits and create a new, clean branch with all changes squashed into a single, well-documented commit suitable for pull request review.

**Options:**
- `--suffix <suffix>`: Custom suffix for new branch name (default: `-release`)
- `--base <base-branch>`: Base branch to squash onto (default: `origin/master`)

## Process

### Step 1: Analyze Current Branch State
**Gather comprehensive information:**
- Run `git branch --show-current` to identify current branch name
- Run `git fetch <base-branch>` to ensure latest base is available
- Run `git log --oneline <base>..HEAD` to count commits
- Run `git diff --name-only $(git merge-base HEAD <base>)..HEAD` to list modified files
- Display commit count and file summary to user

**Verification Checks:**
- Ensure we're not on master/main branch
- Verify commits exist to squash (>0 commits ahead)
- Check for uncommitted changes (warn if found)

### Step 2: Conflict Detection and Safety Checks
**Pre-flight validation:**
- Identify merge base: `git merge-base HEAD <base-branch>`
- Check if modified files were changed on base branch:
  ```bash
  git log $(git merge-base HEAD <base>)..<base> -- <our-files>
  ```
- If conflicts detected: **WARN USER** and offer to continue or abort
- Verify all files with byte-level comparison after squash

**Safety Protocol:**
- Never delete original branch
- Always preserve development history
- Create new branch for squashed version
- Allow rollback if issues detected

### Step 3: Create Clean Squashed Branch
**Branch creation workflow:**
1. Generate new branch name: `<original-branch-name><suffix>`
2. Create new branch from latest base: `git checkout -b <new-branch> <base-branch>`
3. Verify new branch is on latest base commit
4. Perform squash merge: `git merge --squash <original-branch>`
5. Check staged changes: `git status` and `git diff --cached --stat`

**Example:**
```bash
# Current: ml/feature-improvements (29 commits ahead)
# Creates: ml/feature-improvements-release (1 clean commit)

# If already on a -release branch:
# Current: ml/feature-improvements-release (3 commits ahead)
# Creates: ml/feature-improvements-release-v2 (1 clean commit)
```

### Step 4: Comprehensive Commit Message Generation
**Message structure for squashed commits:**

```
<type>(<scope>): <brief description>

Major Features:
- <feature 1>
- <feature 2>
- <feature 3>

User Experience Improvements:
- <improvement 1>
- <improvement 2>

Code Quality:
- <quality improvement 1>
- <quality improvement 2>

Technical Details:
- <technical detail 1>
- <technical detail 2>

Files Changed:
- <action>: <file path>
- <action>: <file path>

<Optional summary paragraph explaining overall impact>
```

**Message Generation Logic:**
1. Analyze all commits in the branch to extract themes
2. Group related changes into categories (Features, UX, Quality, Technical)
3. Use conventional commit type (feat/refactor/fix) based on primary change type
4. Extract scope from modified file paths or commit messages
5. Write imperative, present-tense descriptions
6. Include file-level changes summary
7. Add contextual summary paragraph

**Commit Type Priority:**
- If any commits are `feat:` → use `feat`
- If all commits are `fix:` → use `fix`
- If mixed with refactoring → use primary type
- Default to `feat` for major multi-commit work

### Step 5: Verification and Validation
**Ensure nothing was lost:**

1. **File-by-file verification:**
   ```bash
   # Compare each file between branches
   diff <(git show original:file) <(git show new:file)
   ```
   - Must show 0 differences for all modified files
   - Any difference = CRITICAL ERROR, investigate immediately

2. **Statistics comparison:**
   ```bash
   git diff --stat original-branch new-branch
   ```
   - Should show no differences

3. **Commit count verification:**
   - Original branch: `git log --oneline <base>..original | wc -l`
   - New branch: Should be exactly 1 commit ahead of base

4. **Display verification results:**
   ```
   ✅ Controller file verified: byte-identical
   ✅ View file verified: byte-identical
   ✅ CSS file verified: byte-identical
   ✅ All changes preserved successfully
   ```

### Step 6: Branch Status Report
**Provide comprehensive summary:**

```
✅ Clean Squashed Branch Created

New Branch: <branch-name>-release (or -release-v2, -release-v3, etc. if iterating)
- Based on latest <base-branch> (commit <hash>)
- Single commit: <commit-hash>
- All <N> commits squashed into one

Verification Results:
✅ No conflicts with base branch
✅ All changes preserved (verified byte-for-byte)
✅ Clean history (1 commit vs <N> in original)

Files Changed:
- <file1> (+X lines)
- <file2> (-Y lines)
- <file3> (new, Z lines)

Total: +<insertions> insertions, -<deletions> deletions

Branch Status:
✅ <new-branch> - Ready for PR (1 clean commit)
🗃️ <original-branch> - Preserved for history (<N> commits)

Next Steps:
1. Push <new-branch> to create PR
2. Delete old branch after PR approval
3. The clean single commit makes review much easier
```

## Implementation Logic

**Branch Naming Algorithm:**
```python
def generate_branch_name(current_branch, suffix="-release"):
    # Check if already on a -release branch with version
    if re.search(r'-release-v(\d+)$', current_branch):
        # Extract version and increment
        match = re.search(r'(.*-release)-v(\d+)$', current_branch)
        base = match.group(1)
        version = int(match.group(2)) + 1
        return f"{base}-v{version}"

    # Check if on a -release branch without version
    if current_branch.endswith('-release'):
        # Add v2 version
        return f"{current_branch}-v2"

    # New release branch - strip any old suffixes and add -release
    base = re.sub(r'-final|-squashed|-clean$', '', current_branch)
    return f"{base}{suffix}"
```

**Conflict Detection:**
```python
def check_conflicts(our_files, base_branch, merge_base):
    # Check if our modified files were touched on base
    for file in our_files:
        commits = git_log(f"{merge_base}..{base_branch}", "--", file)
        if commits:
            return True, commits
    return False, []
```

**Verification Algorithm:**
```python
def verify_identical(original_branch, new_branch, files):
    for file in files:
        original_content = git_show(f"{original_branch}:{file}")
        new_content = git_show(f":{file}")  # staged version
        if original_content != new_content:
            raise VerificationError(f"File {file} differs!")
    return True
```

## Expected Behavior

When El Presidente runs `/squash-branch`:

1. **"Analyzing current branch state..."** - Shows 29 commits ahead, 3 files changed
2. **"Checking for potential conflicts..."** - Verifies files not modified on master
3. **"Creating squashed branch: ml/feature-release..."** - New branch created from latest master (or -release-v2 if iterating)
4. **"Performing squash merge..."** - All changes merged as staged modifications
5. **"Generating comprehensive commit message..."** - Analyzes commits and creates structured message
6. **"Committing squashed changes..."** - Single commit created with detailed message
7. **"Verifying file integrity..."** - Byte-by-byte verification of all files
8. **"✅ Branch ready for merge"** - Summary with next steps

**Perfect Squashed Branch:**
- Single, well-documented commit
- Comprehensive commit message with all improvements
- Verified byte-identical to original branch
- Clean history suitable for PR review
- Original branch preserved for reference
- No merge conflicts with base branch
- Ready for immediate pull request creation

## Advanced Features

### Custom Commit Message Template
If user wants specific message structure, allow override:
```
/squash-branch --template custom
```
Then prompt for:
- Commit type and scope
- Major sections to include
- Custom formatting preferences

### Interactive Mode
For complex branches, offer interactive commit message editing:
```
/squash-branch --interactive
```
- Show generated message
- Ask for approval or modifications
- Allow section-by-section editing
- Commit when approved

### Multi-Base Squashing
Support squashing onto different base branches:
```
/squash-branch --base origin/develop
```
Useful for:
- Feature branches off develop
- Hotfix branches off release branches
- Multi-environment workflows

### Conflict Resolution Guidance
If conflicts detected:
1. Show which files conflict
2. Show conflicting commits from base
3. Offer to:
   - Continue anyway (manual resolution)
   - Abort squash operation
   - Rebase first, then squash

### Post-Squash Cleanup
After successful squash, offer to:
```
Would you like to:
1. Push new branch to remote? [Y/n]
2. Delete old local branch? [y/N]
3. Create pull request? [y/N]
```

## Error Handling

**Common Scenarios:**

1. **Uncommitted Changes:**
   ```
   ⚠️ Warning: You have uncommitted changes
   Options:
   - Commit them first
   - Stash them temporarily
   - Abort squash operation
   ```

2. **Already on Master:**
   ```
   ❌ Error: Cannot squash master branch
   Please checkout a feature branch first
   ```

3. **No Commits to Squash:**
   ```
   ℹ️ Info: Branch is up to date with master
   No commits to squash
   ```

4. **Merge Conflicts Detected:**
   ```
   ⚠️ Warning: Modified files changed on master
   Files: app/models/user.rb, app/controllers/api.rb
   Commits on master: 3 commits modified these files

   Recommendation: Rebase first to resolve conflicts
   Continue anyway? [y/N]
   ```

5. **Verification Failed:**
   ```
   ❌ CRITICAL: File verification failed!
   File: app/views/admin/index.html.erb
   Content differs between branches

   This should never happen. Aborting for safety.
   Original branch preserved at: <branch-name>
   ```

## Best Practices

**When to Use:**
- Development branch with many "WIP" or incremental commits
- Feature complete and ready for review
- Want clean history for main branch
- Multiple developers worked on branch with messy commits

**When NOT to Use:**
- Single commit branches (already clean)
- Shared branches others are working on (preserve history)
- Already pushed branch others have branched from
- Historical branches for reference (preserve original commits)

**Workflow Integration:**
```
Development Flow:
1. Work on feature branch with many commits
2. Feature complete, tests passing
3. Run /squash-branch to create clean version
4. Push clean branch and create PR
5. After merge, delete both branches
```

## Examples

### Example 1: Feature Branch with 29 Commits
```
$ /squash-branch

Analyzing current branch: ml/dlq-admin-tool-improvements
- 29 commits ahead of origin/master
- 3 files modified: queues.css, queues_controller.rb, index.html.erb
- No conflicts detected

Creating squashed branch: ml/dlq-admin-tool-improvements-release
✅ Squash merge completed
✅ Comprehensive commit message generated
✅ All files verified byte-identical

Ready for pull request!
```

### Example 2: Custom Base Branch
```
$ /squash-branch --base origin/develop --suffix -ready

Creating squashed branch from origin/develop
Original: feature/user-auth (15 commits)
New: feature/user-auth-ready (1 commit)

✅ Branch ready for merge to develop
```

### Example 3: Iterating on Release Branch
```
$ /squash-branch

Analyzing current branch: ml/feature-auth-release
- 3 commits ahead of origin/master
- Already on a -release branch, creating versioned release

Creating squashed branch: ml/feature-auth-release-v2
✅ Squash merge completed
✅ Comprehensive commit message generated

Branch ml/feature-auth-release-v2 ready for review!
```

### Example 4: Conflict Detection
```
$ /squash-branch

⚠️ Warning: Potential conflicts detected
Modified on both branches:
- app/models/user.rb (2 commits on master)

Recommendation:
  git rebase origin/master
  /squash-branch

Continue anyway? [y/N]: n
Aborted. No changes made.
```

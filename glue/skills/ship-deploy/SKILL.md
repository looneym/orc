---
name: ship-deploy
description: |
  Merge current worktree branch to master at ~/src/orc with full validation.
  Use when user says "/ship-deploy", "ship deploy", "deploy shipment", "merge to master", or wants to integrate their worktree changes into the main ORC repository.
  Handles pre-flight checks, merge, hook installation, and post-merge cleanup.
---

# Ship Deploy

Merge the current worktree branch into master at `~/src/orc` with full validation, then transition the shipment to deployed status.

## ORC Repository Only

This skill is specific to the ORC repository (direct-to-master workflow). PR-based repositories are not yet supported.

## Workflow

### 0. Verify ORC Repository

Before proceeding, verify we're in the ORC repo:

```bash
# Check if we're in the ORC repository
git remote get-url origin 2>/dev/null | grep -q "orc"
```

If not in ORC repo, stop immediately:
```
This skill is specific to the ORC repository.
Current repo: [detected repo from git remote]

For other repositories, use standard git merge workflow or PR-based deployment.
```

### 1. Pre-flight Checks

Before merging, verify:

```bash
# Must be clean
git status --porcelain  # Should be empty

# Must pass lint
make lint

# Detect current branch (the one to merge)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
```

If any check fails, stop and report the issue.

### 2. Merge to Master

```bash
cd ~/src/orc
git checkout master
git pull origin master
git merge <BRANCH> --no-edit
```

Report merge result (fast-forward or merge commit).

### 3. Post-merge Build Steps

The post-merge hook shows reminders but doesn't auto-execute. Run the checklist manually:

```bash
# Initialize and rebuild
make init          # Install dependencies and hooks
make install       # Build and install orc binary
make deploy-glue   # Deploy Claude Code skills/hooks
make test          # Verify everything works
```

### 4. Schema Sync (if needed)

If the hook warned about schema drift:

```bash
make schema-diff   # Preview changes
make schema-apply  # Apply to local DB
```

### 5. Push to Origin

```bash
git push origin master
```

### 6. Rebase Worktree Branch

Return to the worktree and rebase:

```bash
cd <original-worktree-path>
git rebase master
```

### 7. Update Shipment Status

Transition the shipment to deployed status:

```bash
orc status  # Get focused shipment
orc shipment deploy SHIP-XXX
```

This marks the shipment as deployed (merged to master / deployed to prod).

### 8. Notify Goblin

Mail and nudge the goblin with a summary of completed work:

```bash
# Find the gatehouse for the current workshop
orc workshop show  # Note the gatehouse ID (GATE-XXX)

orc mail send "<summary of completed tasks and commit hash>" \
  --to GOBLIN-GATE-XXX \
  --subject "SHIP-XXX Deployed" \
  --nudge
```

Include in the message:
- Shipment ID and title
- List of completed tasks with brief descriptions
- Commit hash on master

## Success Output

Report completion with:
- Branch merged
- Build completed (init, install, deploy-glue, test)
- Schema synced (if needed)
- Master pushed
- Worktree rebased
- Shipment status → deployed
- Goblin notified and nudged

## Error Handling

| Error | Action |
|-------|--------|
| Not ORC repo | Report and suggest standard git workflow |
| Dirty working tree | List uncommitted files, ask to commit or stash |
| Lint fails | Show failures, do not proceed |
| Merge conflicts | Report conflicts, do not auto-resolve |
| Push rejected | Check if remote has new commits, suggest pull --rebase |

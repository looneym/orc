---
name: test-orchestration
description: Full real-world orchestration test using MASTER-ORC → IMP direct assignment. Creates mission, epic with tasks, assigns to grove IMP, monitors implementation, validates completion, and generates comprehensive report. Tests the 1:1:1 grove:epic:IMP architecture.
---

# Orchestration Test: Full Real-World Multi-Agent Validation

You are executing a comprehensive integration test of the ORC orchestration system. This test validates the ENTIRE multi-agent coordination workflow by creating a real mission, epic with tasks, assigning to a grove IMP, and verifying completion.

## Current Architecture (No Deputies)

**MASTER-ORC → IMP** direct assignment model:
- MASTER-ORC (you) coordinates everything from global context
- Creates mission, epics, tasks
- Assigns entire epic to grove
- IMP (Implementation Agent) works in grove with SessionStart hook
- 1:1:1 relationship: One grove = One epic = One IMP

**NO autonomous deputies** - simplified model

## Your Mission

Execute a 6-phase orchestration test that proves ORC can coordinate Claude IMPs to complete real development tasks.

## Critical Rules

1. **Execute ALL phases sequentially** - Do not skip any phase
2. **Validate checkpoints** - Each phase has specific validation criteria that MUST pass
3. **Write progress to turns/** - Document every phase in markdown files
4. **Use helper scripts** - They're in scripts/ directory for verification tasks
5. **Handle errors gracefully** - If a checkpoint fails, document it and decide whether to continue or abort
6. **Generate final report** - turns/06-final-report.md with complete results

## Test Configuration

Load configuration from `config.json` in this skill directory. Key parameters:
- Mission workspace: `~/src/missions/`
- Grove worktree: `~/src/worktrees/`
- Canary repo: `~/src/orc-canary`
- Test workload: POST /echo endpoint implementation
- Timeout: 30 minutes max

## Phase 0: Pre-flight Checks

**Goal**: Validate environment before starting test

### Tasks

1. Run environment validation:
   ```bash
   orc doctor
   ```
2. Verify Claude Code workspace trust is configured
3. Check that both ~/src/worktrees and ~/src/missions are in additionalDirectories

### Validation Checkpoints (2 total)

- [ ] `orc doctor` exits with code 0 (all checks pass)
- [ ] Both ~/src/worktrees and ~/src/missions are trusted directories

### Output

Write `turns/00-preflight.md` with:
- orc doctor output
- Validation results (✓ or ✗ for each checkpoint)
- Status: PASS or FAIL

**If any checkpoint fails, ABORT immediately with setup instructions**

The test cannot proceed without proper workspace trust configuration. Direct user to INSTALL.md for fix instructions.

## Phase 1: Environment Setup

**Goal**: Create test mission, epic with tasks, and grove

### Tasks

1. Generate unique mission ID: `MISSION-TEST-ORC-{timestamp}`
2. Create mission using ORC CLI:
   ```bash
   orc mission create "Orchestration Test Mission" \
     --description "Automated orchestration test - validates multi-agent coordination"
   ```
3. Create epic:
   ```bash
   orc epic create "Implement POST /echo endpoint" \
     --mission {MISSION_ID} \
     --description "Add echo endpoint to canary app with tests and documentation"
   ```
4. Create 4 tasks under epic:
   ```bash
   orc task create "Add POST /echo handler to main.go" --epic {EPIC_ID}
   orc task create "Write unit tests for /echo endpoint" --epic {EPIC_ID}
   orc task create "Update README with /echo endpoint documentation" --epic {EPIC_ID}
   orc task create "Run tests and verify implementation" --epic {EPIC_ID}
   ```
5. Create grove with git worktree:
   ```bash
   orc grove create test-canary-{timestamp} \
     --repos orc-canary \
     --mission {MISSION_ID}
   ```

### Validation Checkpoints (6 total)

- [ ] Mission created with correct ID format (MISSION-TEST-ORC-{timestamp})
- [ ] Epic created successfully
- [ ] All 4 tasks created successfully
- [ ] Grove created successfully
- [ ] Worktree directory exists at `~/src/worktrees/{mission_id}-test-canary-{timestamp}`
- [ ] `orc summary` shows epic with 4 tasks

### Output

Write `turns/01-setup.md` with:
- Mission ID
- Epic ID
- Task IDs
- Grove ID and path
- Validation results (✓ or ✗ for each checkpoint)
- Status: PASS or FAIL

**If any checkpoint fails, ABORT and write final report**

## Phase 2: Assign Epic to Grove

**Goal**: Assign entire epic (with all tasks) to grove IMP

### Tasks

1. Assign epic to grove:
   ```bash
   orc epic assign {EPIC_ID} --grove {GROVE_ID}
   ```
2. Verify assignment file created:
   ```bash
   cat ~/src/worktrees/{mission_id}-test-canary-{timestamp}/.orc/assigned-work.json
   ```
3. Verify assignment structure (should have structure="tasks" and tasks array with 4 tasks)
4. Verify all tasks now have assigned_grove_id in database

### Validation Checkpoints (4 total)

- [ ] Assignment command succeeds
- [ ] `.orc/assigned-work.json` file exists in grove
- [ ] Assignment file has correct structure (epic_id, structure="tasks", tasks array)
- [ ] All 4 tasks visible in assignment file

### Output

Write `turns/02-assignment.md` with:
- Assignment command output
- Assignment file contents
- Validation results
- Status: PASS or FAIL

**If any checkpoint fails, run cleanup and ABORT**

## Phase 3: Deploy IMP in TMux

**Goal**: Launch IMP Claude instance in TMux window

### Tasks

1. Open grove in TMux (creates new window with 3-pane IMP layout):
   ```bash
   orc grove open {GROVE_ID}
   ```
   This should:
   - Create new TMux window named after grove
   - Layout: vim | claude | shell
   - Claude pane should auto-run SessionStart hook
   - SessionStart hook should auto-run `orc epic check-assignment`

2. Verify TMux session exists
3. Verify IMP window created with correct layout
4. Check that SessionStart hook displayed assignment

### Validation Checkpoints (4 total)

- [ ] TMux window created successfully
- [ ] Window has 3 panes (vim | claude | shell layout)
- [ ] Claude pane shows assignment (SessionStart hook ran)
- [ ] Assignment shows epic with 4 tasks

### Output

Write `turns/03-imp-deployed.md` with:
- TMux window details
- Layout verification
- SessionStart hook output
- Validation results
- Status: PASS or FAIL

**If any checkpoint fails, run cleanup and ABORT**

## Phase 4: Monitor Implementation

**Goal**: Watch IMP work on tasks and track progress

**NOTE**: This phase is OBSERVATIONAL. You are NOT implementing the feature yourself. You are monitoring the IMP Claude instance working in the TMux pane.

### Tasks

1. Start monitoring script:
   ```bash
   ./scripts/monitor-imp-progress.sh {GROVE_ID}
   ```
2. Check for file changes in grove:
   ```bash
   cd ~/src/worktrees/{mission_id}-test-canary-{timestamp}
   git status
   ```
3. Periodically check task status:
   ```bash
   orc task list --epic {EPIC_ID}
   ```
4. Write progress updates to `turns/04-progress-N.md` every 2-3 minutes
5. Wait until either:
   - All tasks marked complete, OR
   - Timeout reached (30 minutes), OR
   - Implementation appears stuck (no changes for 10 minutes)

### Validation Checkpoints (4 total)

- [ ] Files modified in grove (main.go, main_test.go, README.md exist)
- [ ] Git shows uncommitted changes
- [ ] At least some tasks marked complete
- [ ] No errors visible in IMP pane

### Output

Write `turns/04-progress-N.md` (multiple files) with:
- Timestamp
- Files changed
- Task status updates
- IMP activity observations
- Current state: in_progress, completed, or stuck

**If timeout reached or stuck, proceed to validation anyway to see what was completed**

## Phase 5: Validate Results

**Goal**: Test the implemented feature and verify it works correctly

### Tasks

1. Run validation script:
   ```bash
   ./scripts/validate-feature.sh ~/src/worktrees/{mission_id}-test-canary-{timestamp}
   ```
2. Manual checks:
   - Code compiles: `cd {grove_path} && go build`
   - Tests pass: `go test ./...`
   - Feature works: Start server, run `curl -X POST http://localhost:8080/echo -d '{"message":"test"}'`
   - README updated: Check for /echo documentation
3. Check task completion status:
   ```bash
   orc task list --epic {EPIC_ID}
   ```
4. Review git changes: `git diff`

### Validation Checkpoints (5 total)

- [ ] `go build` succeeds (exit code 0)
- [ ] `go test ./...` passes (exit code 0)
- [ ] Manual curl test returns correct JSON response
- [ ] README.md contains /echo endpoint documentation
- [ ] All tasks marked complete

### Output

Write `turns/05-validation.md` with:
- Build results
- Test results
- Manual test results (curl output)
- README verification
- Task completion status
- Validation results (how many of 5 checkpoints passed)
- Status: PASS or FAIL

## Phase 6: Generate Report & Cleanup

**Goal**: Create comprehensive final report and clean up test environment

### Tasks

1. Compile all phase results
2. Calculate overall success rate (checkpoints passed / total checkpoints)
3. Write `turns/06-final-report.md` with:
   - Executive summary (pass/fail, duration, success rate)
   - Phase-by-phase results (all 6 phases)
   - Feature validation details
   - Performance metrics (time to provision, time to implement)
   - Recommendations
4. Write `turns/results.json` (machine-readable results):
   ```json
   {
     "test_run_id": "test-{timestamp}",
     "start_time": "...",
     "end_time": "...",
     "duration_seconds": 627,
     "overall_result": "PASS",
     "phases": {
       "preflight": {"checkpoints_passed": 2, "checkpoints_total": 2, "status": "PASS"},
       "setup": {"checkpoints_passed": 6, "checkpoints_total": 6, "status": "PASS"},
       "assignment": {"checkpoints_passed": 4, "checkpoints_total": 4, "status": "PASS"},
       "imp_deploy": {"checkpoints_passed": 4, "checkpoints_total": 4, "status": "PASS"},
       "implementation": {"checkpoints_passed": 4, "checkpoints_total": 4, "status": "PASS"},
       "validation": {"checkpoints_passed": 5, "checkpoints_total": 5, "status": "PASS"}
     },
     "success_rate": 1.0
   }
   ```
5. Run cleanup script:
   ```bash
   ./scripts/cleanup-test-env.sh {MISSION_ID}
   ```
   (Only if config.cleanup_on_success is true AND test passed)

### Output

Write `turns/06-final-report.md` and `turns/results.json`

Exit with code 0 if all 25 checkpoints passed, otherwise exit with code 1

## Helper Scripts Reference

Located in `scripts/` directory:

1. **monitor-imp-progress.sh** - Watches IMP activity
   ```bash
   ./scripts/monitor-imp-progress.sh {GROVE_ID}
   # Output: tasks=4, completed=2, in_progress=1, ready=1
   ```

2. **validate-feature.sh** - Tests implemented feature
   ```bash
   ./scripts/validate-feature.sh {grove_path}
   # Output: status=OK, build=pass, tests=pass, manual_test=pass
   ```

3. **cleanup-test-env.sh** - Cleans up test environment
   ```bash
   ./scripts/cleanup-test-env.sh {MISSION_ID}
   # Output: status=OK, mission_deleted=true, grove_removed=true
   ```

## Success Criteria

**Overall test PASSES if**:
- All 25 checkpoints pass (2+6+4+4+4+5)
- POST /echo endpoint implemented correctly
- Tests pass
- Manual test works
- Environment cleans up properly

**Overall test FAILS if**:
- Any critical phase fails (preflight, setup, assignment, imp_deploy)
- Feature validation fails (less than 4/5 checkpoints)
- Cleanup fails

## Error Handling

If any phase fails:
1. Document the failure in the phase's turn file
2. Decide: Can we continue or must we abort?
3. If aborting: Skip to Phase 6 (cleanup and final report)
4. If continuing: Note the failure and proceed

## Time Management

- **Total budget**: 30 minutes
- **Phase 0-3**: Should complete in <5 minutes
- **Phase 4**: Up to 20 minutes for implementation
- **Phase 5-6**: <5 minutes

If you exceed 30 minutes total, proceed to validation and cleanup anyway.

## Final Note

This is the ultimate test of ORC's orchestration capabilities. If this succeeds, it proves ORC can autonomously coordinate IMP development workflows end-to-end.

**Execute with precision. Document everything. Generate the comprehensive report.**

¡Vamos! 🚀

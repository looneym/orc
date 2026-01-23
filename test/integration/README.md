# ORC Integration Tests

Comprehensive integration test suite for validating ORC core functionality with real-world scenarios.

## Quick Start

```bash
# Run all integration tests
cd ~/src/orc
./test/integration/run-all-tests.sh

# Run individual test
./test/integration/01-test-commission-creation.sh
./test/integration/02-test-workbench-tmux.sh
```

## Test Coverage

### 01-test-commission-creation.sh (7 tests)
Tests commission creation and context bootstrap workflow:
- ✓ Create commission
- ✓ Create commission workspace directory structure
- ✓ Write .orc-commission marker (JSON format)
- ✓ Write workspace config.json
- ✓ Commission context detection from .orc-commission file
- ✓ Create work order in commission context (auto-scoping)
- ✓ Command auto-scoping to commission

**Validates:**
- Commission lifecycle
- Commission context detection
- File-based commission markers
- Auto-scoping of commands to commission

### 02-test-workbench-tmux.sh (6 tests)
Tests workbench creation and TMux integration:
- ✓ Create workbench with git worktree
- ✓ Verify workbench metadata in .orc/ directory
- ✓ Verify commission marker propagation to workbench
- ✓ Test commands from workbench directory
- ✓ TMux session basics
- ✓ Workbench show command

**Validates:**
- Git worktree integration
- Workbench database registration
- Metadata propagation
- Cross-directory command execution
- TMux session management

## Test Framework

### test-helpers.sh
Provides reusable test utilities:

**Assertion Functions:**
- `assert_command_succeeds` - Test command success
- `assert_command_fails` - Test expected failure
- `assert_file_exists` - Verify file existence
- `assert_directory_exists` - Verify directory existence
- `assert_contains` - Check string containment

**Test Lifecycle:**
- `run_test` - Execute test with pass/fail tracking
- `register_cleanup` - Register cleanup functions
- `run_cleanup` - Execute all cleanup functions
- `print_test_summary` - Display test results

**Logging:**
- `log_info` - Informational messages (blue)
- `log_success` - Success messages (green)
- `log_error` - Error messages (red)
- `log_warn` - Warning messages (yellow)
- `log_section` - Section headers

### Test Isolation

Each test:
- Uses COMM-420 as the dedicated test commission (avoids DB pollution)
- Creates unique workbenches (using timestamps)
- Registers cleanup functions
- Cleans up workbenches/worktrees on exit (success or failure)
- Does not interfere with other tests

## Prerequisites

- ORC installed and in PATH (`cd ~/src/orc && go build && install orc`)
- ORC initialized (`orc init` creates ~/.orc/orc.db)
- TMux installed
- orc-canary repository at ~/src/orc-canary
- Git configured
- COMM-420 exists as the dedicated test commission (`orc commission create "Integration Test Commission" --id COMM-420`)

## Running Tests

### Full Test Suite
```bash
cd ~/src/orc
./test/integration/run-all-tests.sh
```

**Output:**
```
━━━ ORC Integration Test Suite ━━━
Running: 01-test-commission-creation.sh
  ✓ 7/7 tests passed
Running: 02-test-workbench-tmux.sh
  ✓ 6/6 tests passed

✓✓✓ ALL TESTS PASSED ✓✓✓
Test Files Run: 2
Test Files Passed: 2
Test Files Failed: 0
```

### Individual Tests
```bash
# Commission creation tests only
./test/integration/01-test-commission-creation.sh

# Workbench and TMux tests only
./test/integration/02-test-workbench-tmux.sh
```

## Test Development

### Adding New Tests

1. Create test file: `test/integration/XX-test-name.sh`
2. Source helpers: `source "$SCRIPT_DIR/test-helpers.sh"`
3. Define test functions: `test_something() { ... }`
4. Register cleanup: `cleanup() { ... }; trap cleanup EXIT`
5. Run tests: `run_test "Description" test_something`
6. Add to `run-all-tests.sh` TEST_FILES array

### Example Test Function
```bash
test_something() {
    log_info "Testing something"

    # Do test actions
    orc command ...

    # Make assertions
    assert_command_succeeds "orc status" "Status command works"
    assert_contains "$output" "expected" "Output contains text"

    return $?
}
```

### Cleanup Pattern
```bash
TEST_COMMISSION_ID=""

cleanup() {
    log_section "Cleanup"
    if [[ -n "$TEST_COMMISSION_ID" ]]; then
        # Clean up test data
        rm -rf "$HOME/src/factories/$TEST_COMMISSION_ID"
    fi
}

trap cleanup EXIT
```

## Test Scenarios Validated

### Commission Lifecycle
- ✓ Commission creation with auto-generated IDs
- ✓ Commission workspace directory structure
- ✓ .orc-commission marker (JSON format with commission_id)
- ✓ Workspace config.json (active_commission_id)
- ✓ Commission context detection from any subdirectory

### Commission Context
- ✓ Context auto-detection from .orc/config.json file
- ✓ Command auto-scoping to commission (no --commission flag needed)
- ✓ Work order creation scoped to commission
- ✓ Summary and status commands show correct commission

### Workbench Management
- ✓ Workbench creation with database registration
- ✓ Git worktree creation and validation
- ✓ Config propagation (.orc/config.json)
- ✓ Commission marker propagation (.orc-commission)
- ✓ Commands work from workbench directories
- ✓ Workbench show command displays correct info

### TMux Integration
- ✓ TMux session creation
- ✓ Window/pane management
- ✓ Directory context preservation

## Known Limitations

1. **TMux Pane Testing**: Tests verify TMux sessions and windows exist but don't test actual pane content or IMP spawning (requires interactive environment)

2. **Commission Deletion**: Currently only cleans up directories, not database records (DeleteCommission function exists but not exposed via CLI yet)

3. **Parallel Execution**: Tests should be run sequentially to avoid TMux session conflicts

## CI/CD Integration

These tests are designed to run in CI environments:

```bash
# In CI pipeline
- name: Run ORC Integration Tests
  run: |
    cd ~/src/orc
    ./test/integration/run-all-tests.sh
```

**Exit Codes:**
- 0 = All tests passed
- 1 = Some tests failed

## Troubleshooting

### "orc command not found"
```bash
cd ~/src/orc
go build -o orc cmd/orc/main.go
# Add to PATH or use absolute path
```

### "ORC database not found"
```bash
orc init
```

### "orc-canary repository not found"
```bash
cd ~/src
git clone git@github.com:example/orc-canary.git
```

### "COMM-420 not found"
```bash
# Create dedicated test commission
orc commission create "Integration Test Commission" --description "Dedicated commission for integration tests"
# Note: Commission ID will be auto-generated; rename to COMM-420 or update tests
```

### "TMux session already exists"
```bash
# Kill existing test session
tmux kill-session -t test-orc-session
```

### Stale test commissions/workbenches
```bash
# Manual cleanup
rm -rf ~/src/factories/COMM-*
rm -rf ~/src/worktrees/test-canary-*
cd ~/src/orc-canary && git worktree prune
```

## Test Results

Last run: 2026-01-14

```
Test Files: 2
Total Tests: 13
Passed: 13 ✓
Failed: 0
Success Rate: 100%
```

## Roadmap

Future test additions:
- [ ] Work order state transitions
- [ ] Handoff creation and retrieval
- [ ] Workbench open command (IMP layout validation)
- [ ] Proto-mail system (WO-061 ↔ WO-065)
- [ ] Cross-workbench coordination
- [ ] Error cases and edge conditions
- [ ] Performance benchmarks

## Contributing

When adding new ORC features:
1. Write integration tests first (TDD)
2. Ensure tests are isolated and clean up properly
3. Update this README with new test coverage
4. Run full test suite before committing

---

**Status**: ✓ Operational - All core functionality validated
**Maintained by**: ORC Development Team
**Last Updated**: 2026-01-14

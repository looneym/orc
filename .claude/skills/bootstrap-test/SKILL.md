---
name: bootstrap-test
description: |
  Test make bootstrap in a fresh macOS VM using Tart.
  Use when user says "/bootstrap-test", "test bootstrap", or "run bootstrap test".
---

# Bootstrap Test

Run `make bootstrap` in a clean macOS VM to verify the bootstrap experience works on a fresh system.

## Usage

```
/bootstrap-test                   (run test, cleanup on success)
/bootstrap-test --keep-on-failure (keep VM if test fails for debugging)
/bootstrap-test --verbose         (show detailed progress)
```

## Prerequisites

- **tart** - macOS VM manager (`brew install cirruslabs/cli/tart`)
- **sshpass** - Non-interactive SSH (`brew install sshpass`)

## What It Tests

The test validates the full first-run experience:

1. Creates a fresh macOS VM (Tahoe base image)
2. Installs Go via Homebrew
3. Copies ORC repo into VM
4. Runs `make bootstrap`
5. Verifies `orc` is in PATH and works
6. Cleans up VM on success

## Flow

### Step 1: Check Prerequisites

```bash
command -v tart
command -v sshpass
```

If missing, show installation instructions.

### Step 2: Run Test

```bash
make bootstrap-test
```

Or with flags:

```bash
./scripts/bootstrap-test.sh --verbose
./scripts/bootstrap-test.sh --keep-on-failure
```

### Step 3: Report Results

**On success:**
```
✓ Bootstrap test PASSED in Xs

Verified:
- VM boots and SSH works
- Go installs via brew
- make bootstrap completes
- orc command works via PATH
```

**On failure:**
```
✗ Bootstrap test FAILED

[Error details from script output]

To debug:
  ./scripts/bootstrap-test.sh --keep-on-failure --verbose

Then SSH to the VM:
  ssh admin@$(tart ip orc-bootstrap-test-XXXX)
  Password: admin
```

## When to Use

- Before releases (validates bootstrap works for new users)
- After changing Makefile bootstrap target
- After changing PATH or shell configuration
- When debugging bootstrap issues

## Timing

Typical run: ~60-70 seconds
- VM boot: ~10s
- Go install: ~20s
- make bootstrap: ~30s

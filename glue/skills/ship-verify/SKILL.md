---
name: ship-verify
description: |
  Post-deploy verification for shipments. Use when user says "/ship-verify",
  "verify deployment", "verify shipment", or wants to confirm a deploy was successful.
  Runs verification checks and transitions shipment to verified status.
---

# Ship Verify

Post-deploy verification that runs smoke tests and transitions shipment to verified status.

## ORC Repository Only

This skill is specific to the ORC repository. For other repositories, implement your own verification steps.

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

For other repositories, implement your own verification steps.
```

### 1. Get Deployed Shipment

```bash
orc status  # Get focused shipment
orc shipment show SHIP-XXX
```

Verify shipment status is `deployed`. If not deployed:
- If `implemented`: "Shipment not yet deployed. Run /ship-deploy first."
- If `verified`: "Shipment already verified."
- If `complete`: "Shipment already complete."
- Other: "Shipment must be in deployed status to verify."

### 2. Run Verification Checks

Run the following verification commands:

```bash
# Health check
orc doctor

# Basic smoke tests - commands should not error
orc status
orc commission list
orc shipment list

# Build verification (if in worktree with Makefile)
make test
```

Report each check as PASS/FAIL:
```
Verification Results:
  [PASS] orc doctor
  [PASS] orc status
  [PASS] orc commission list
  [PASS] orc shipment list
  [PASS] make test
```

### 3. Handle Results

**If all checks pass:**
```bash
orc shipment verify SHIP-XXX
```

Output:
```
Shipment SHIP-XXX verified successfully.

Next steps:
  /ship-complete SHIP-XXX  # Complete shipment (terminal state)
```

**If any check fails:**
Do NOT transition status. Report:
```
Verification failed:
  [FAIL] make test - exit code 1

Fix the failing checks and re-run /ship-verify
```

### 4. Notify Goblin (Optional)

If verification passes, optionally notify goblin:

```bash
orc workshop show  # Get gatehouse ID
orc mail send "SHIP-XXX verified - all checks pass" \
  --to GOBLIN-GATE-XXX \
  --subject "SHIP-XXX Verified"
```

## Success Output

```
Verifying SHIP-XXX: Feature Name

Running verification checks...
  [PASS] orc doctor
  [PASS] orc status
  [PASS] orc commission list
  [PASS] orc shipment list
  [PASS] make test

All checks passed!

Transitioning shipment to verified...
 Shipment SHIP-XXX verified

Next steps:
  /ship-complete SHIP-XXX  # Complete shipment (terminal state)
```

## Error Handling

| Error | Action |
|-------|--------|
| Not ORC repo | Report and suggest implementing custom verification |
| Shipment not deployed | Report required status, suggest /ship-deploy |
| Check fails | Report which check failed, do not transition |
| No focused shipment | Ask for shipment ID or suggest `orc focus SHIP-XXX` |

## Usage

```
/ship-verify              (verify focused shipment)
/ship-verify SHIP-xxx     (verify specific shipment)
```

---
name: ship-complete
description: Complete a shipment after verification. Use when user says /ship-complete or wants to mark a shipment as finished.
---

# Ship Complete Skill

Mark a shipment as complete (terminal state) after verification is done.

## Usage

```
/ship-complete              (complete focused shipment)
/ship-complete SHIP-xxx     (complete specific shipment)
/ship-complete --force      (complete even from non-verified status)
```

## Status Lifecycle

This is the final step in the shipment lifecycle:

```
implementing → implemented → deployed → verified → complete
```

Shipments should normally be in "verified" status before completing.

## Flow

### Step 1: Get Shipment

If argument provided:
- Use specified SHIP-xxx

If no argument:
- Get focused shipment from `orc focus --show`
- If no focus, ask which shipment to complete

### Step 2: Verify Readiness

```bash
orc shipment show <SHIP-xxx>
```

Check:
- Shipment status is "verified" (preferred) or "deployed"/"implemented" (with --force)
- Shipment is not pinned

### Step 3: Handle Issues

If status is not "verified" and no --force:
```
Cannot complete SHIP-xxx: shipment is in '<status>' status

The standard lifecycle is:
  implemented → deployed → verified → complete

If you want to skip verification, use:
  /ship-complete SHIP-xxx --force
```

If shipment is pinned:
```
Cannot complete SHIP-xxx: shipment is pinned

Unpin first:
  orc shipment unpin SHIP-xxx
```

### Step 4: Complete Shipment

```bash
orc shipment complete <SHIP-xxx>
```

Or with force:
```bash
orc shipment complete <SHIP-xxx> --force
```

### Step 5: Clear Focus

```bash
orc focus --clear
```

### Step 6: Summary

Output:
```
Shipment completed:
  SHIP-xxx: <Title>
  Status: complete (terminal)

Next steps:
  orc summary              - View remaining work
  /ship-queue claim        - Claim next from queue
  /ship-new "Title"        - Start new shipment
```

## Example Session

```
User: /ship-complete

Agent: [gets focused shipment SHIP-250]
       [runs orc shipment show SHIP-250]

Agent: Completing SHIP-250: Core Model Simplification

       Status: verified
       Ready for completion.

       [runs orc shipment complete SHIP-250]
       [runs orc focus --clear]

Agent: Shipment completed:
         SHIP-250: Core Model Simplification
         Status: complete (terminal)

       Next steps:
         orc summary - View remaining work
         /ship-queue claim - Claim next from queue
```

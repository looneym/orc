---
name: ship-queue
description: View and manage the shipyard queue. Use when user says /ship-queue or wants to see pending shipments and priorities.
---

# Ship Queue Skill

View and manage shipments available for IMP pickup.

## Usage

```
/ship-queue                     (view available shipments)
/ship-queue SHIP-xxx            (mark shipment ready for IMP)
/ship-queue priority SHIP-xxx N (set priority)
```

## Flow

### Step 1: Determine Action

If SHIP-xxx argument (without priority):
- Go to Ready flow (mark shipment ready_for_imp)

If "priority SHIP-xxx N" arguments:
- Go to Priority flow

Otherwise:
- Go to View flow

---

## View Flow

### Step 2: Get Available Shipments

```bash
orc shipment list --available
```

This shows shipments in `ready_for_imp` status.

### Step 3: Show Options

Output:
```
Available Shipments:
  SHIP-xxx: Title (5 tasks) - COMM-001
  SHIP-yyy: Title (4 tasks) - COMM-001
  SHIP-zzz: Title (3 tasks) - COMM-002

Actions:
  orc focus SHIP-xxx             - Focus and start work
  /ship-queue SHIP-xxx           - Mark another shipment ready
  /ship-queue priority SHIP-xxx 1 - Set priority
```

---

## Ready Flow

Mark a shipment as ready for IMP pickup.

### Step 2: Mark Ready

```bash
orc shipment ready SHIP-xxx
```

This sets status to `ready_for_imp`, signaling the shipment is ready for IMP work.

### Step 3: Confirm

Output:
```
Shipment SHIP-xxx marked ready_for_imp

IMP can now claim this shipment:
  orc focus SHIP-xxx
  orc task claim TASK-yyy
```

---

## Priority Flow

### Step 2: Set Priority

```bash
orc shipment update SHIP-xxx --priority N
```

Priority 1 = highest, NULL = default FIFO

### Step 3: Confirm

Output:
```
Priority set:
  SHIP-xxx now at priority N
```

## Example Session

```
User: /ship-queue

Agent: [runs orc shipment list --available]

Agent: Available Shipments (2):

       ID        Title                    Tasks   Commission
       SHIP-248  Test Feature             5       COMM-001
       SHIP-247  Bugfix Batch             2       COMM-001

       Actions:
         orc focus SHIP-248 - Start work on first
         /ship-queue SHIP-249 - Mark another shipment ready

User: /ship-queue SHIP-249

Agent: [runs orc shipment ready SHIP-249]

Agent: ✓ Shipment SHIP-249 marked ready_for_imp

       IMP can now pick up this shipment:
         orc focus SHIP-249
         orc task list --shipment SHIP-249
```

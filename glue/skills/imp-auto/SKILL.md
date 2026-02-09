---
name: imp-auto
description: Toggle autonomous mode for current shipment. Use when IMP wants to enable/disable hook-propelled autonomous work.
---

# IMP Auto Mode

Toggle between autonomous and manual implementation modes. When auto mode is enabled, a Watchdog agent is spawned to monitor IMP progress.

## Usage

```
/imp-auto       (enable auto mode)
/imp-auto on    (enable auto mode)
/imp-auto off   (enable manual mode)
```

## Modes

| Mode | Status | Behavior |
|------|--------|----------|
| **Auto** | `auto_implementing` | Stop hook blocks until shipment complete. Watchdog monitors IMP. |
| **Manual** | `implementing` | IMP can stop freely. No Watchdog monitoring. |

## Flow

### Step 1: Get Current Context

```bash
orc status
```

Check for focused shipment and workbench. If no shipment focused:
```
No shipment focused. Focus a shipment first:
  orc focus SHIP-xxx
```

### Step 2: Determine Action

Parse arguments:
- No args or "on" → Enable auto mode
- "off" → Enable manual mode

### Step 3: Toggle Mode

**Enable Auto Mode:**
```bash
# 1. Update shipment status
orc shipment auto SHIP-xxx

# 2. Start patrol (creates monitoring session for watchdog)
orc patrol start BENCH-xxx

# 3. Apply infrastructure to spawn watchdog pane
orc infra apply WORK-xxx --yes
```

Output:
```
✓ Auto mode enabled for SHIP-xxx
✓ Patrol started: PATROL-xxx
✓ Watchdog spawned in pane 4

The Stop hook will now block until this shipment is complete.
Watchdog is monitoring your progress and will nudge when idle.
IMP workflow:
  /imp-plan-create → /imp-plan-submit → implement → /imp-rec

Stay focused until all tasks are complete!
```

**Enable Manual Mode:**
```bash
# 1. Get active patrol for this workbench
orc patrol status

# 2. End the patrol
orc patrol end PATROL-xxx

# 3. Apply infrastructure to remove watchdog pane
orc infra apply WORK-xxx --yes

# 4. Update shipment status
orc shipment manual SHIP-xxx
```

Output:
```
✓ Patrol ended: PATROL-xxx
✓ Watchdog pane removed
✓ Manual mode enabled for SHIP-xxx

You can now stop at any time. The Stop hook will not block.
Resume work anytime with:
  /imp-nudge
```

## Example Session

```
> /imp-auto

[runs orc status, finds SHIP-265 focused, BENCH-044 current]
[runs orc shipment auto SHIP-265]
[runs orc patrol start BENCH-044]
[runs orc infra apply WORK-010 --yes]

✓ Auto mode enabled for SHIP-265
✓ Patrol started: PATROL-042
✓ Watchdog spawned in pane 4

The Stop hook will now block until this shipment is complete.
Watchdog is monitoring your progress and will nudge when idle.

> /imp-auto off

[runs orc patrol status, finds PATROL-042 active]
[runs orc patrol end PATROL-042]
[runs orc infra apply WORK-010 --yes]
[runs orc shipment manual SHIP-265]

✓ Patrol ended: PATROL-042
✓ Watchdog pane removed
✓ Manual mode enabled for SHIP-265

You can now stop at any time.
```

## Watchdog Integration

When auto mode is enabled:
1. **Patrol record** is created to track the monitoring session
2. **Infra apply** spawns a watchdog pane (index 4) in the TMux window
3. **Watchdog agent** runs `/watchdog-monitor` skill to:
   - Check `orc shipment should-continue` periodically
   - Capture IMP pane and detect state (working/idle/menu/typed/error)
   - Nudge IMP when idle, dismiss menus, report errors
   - End patrol when shipment completes

## Error Handling

- No shipment focused → "No shipment focused. Use `orc focus SHIP-xxx` first."
- Invalid argument → "Usage: /imp-auto [on|off]"
- Already in requested mode → "Shipment SHIP-xxx is already in [auto|manual] mode."
- Patrol start fails → Report error, don't change shipment status
- Infra apply fails → Report error, end patrol, don't change shipment status

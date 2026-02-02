---
name: imp-auto
description: Toggle autonomous mode for current shipment. Use when IMP wants to enable/disable hook-propelled autonomous work.
---

# IMP Auto Mode

Toggle between autonomous and manual implementation modes.

## Usage

```
/imp-auto       (enable auto mode)
/imp-auto on    (enable auto mode)
/imp-auto off   (enable manual mode)
```

## Modes

| Mode | Status | Behavior |
|------|--------|----------|
| **Auto** | `auto_implementing` | Stop hook blocks until shipment complete. IMP propelled through workflow. |
| **Manual** | `implementing` | IMP can stop freely. Human oversight mode. |

## Flow

### Step 1: Get Current Shipment

```bash
orc status
```

Check for focused shipment. If no shipment focused:
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
orc shipment auto SHIP-xxx
```

Output:
```
✓ Auto mode enabled for SHIP-xxx

The Stop hook will now block until this shipment is complete.
IMP workflow:
  /imp-plan-create → /imp-plan-submit → implement → /imp-rec

Stay focused until all tasks are complete!
```

**Enable Manual Mode:**
```bash
orc shipment manual SHIP-xxx
```

Output:
```
✓ Manual mode enabled for SHIP-xxx

You can now stop at any time. The Stop hook will not block.
Resume work anytime with:
  /imp-nudge
```

## Example Session

```
> /imp-auto

[runs orc status, finds SHIP-265 focused]
[runs orc shipment auto SHIP-265]

✓ Auto mode enabled for SHIP-265

The Stop hook will now block until this shipment is complete.
IMP workflow:
  /imp-plan-create → /imp-plan-submit → implement → /imp-rec

> /imp-auto off

[runs orc shipment manual SHIP-265]

✓ Manual mode enabled for SHIP-265

You can now stop at any time.
```

## Error Handling

- No shipment focused → "No shipment focused. Use `orc focus SHIP-xxx` first."
- Invalid argument → "Usage: /imp-auto [on|off]"
- Already in requested mode → "Shipment SHIP-xxx is already in [auto|manual] mode."

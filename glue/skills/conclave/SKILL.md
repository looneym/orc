---
name: conclave
description: "DEPRECATED: Use /ship-new instead. Conclaves are being replaced by shipments."
---

# Conclave Skill (DEPRECATED)

**This skill is deprecated.** Conclaves are being replaced by shipments.

## Migration

Use `/ship-new` instead:

```
/ship-new "Topic Name"
```

Shipments replace conclaves with a unified lifecycle:
- **draft** → **exploring** → **specced** → **tasked** → **in_progress** → **complete**

## What Changed

| Old (Conclave) | New (Shipment) |
|----------------|----------------|
| `orc conclave create` | `orc shipment create` |
| `/conclave` | `/ship-new` |
| `/exorcism` | `/ship-plan` then `/ship-complete` |
| CON-xxx | SHIP-xxx |

## Existing Conclaves

Existing conclaves remain accessible:
- `orc conclave list` - View existing conclaves
- `orc conclave show CON-xxx` - View conclave details
- `orc conclave migrate CON-xxx` - Migrate to shipment

To migrate all open conclaves:
```bash
orc conclave migrate --all
```

## Flow (If Invoked)

If user says `/conclave`, respond:

```
The /conclave skill is deprecated. Use /ship-new instead:

  /ship-new "Topic Name"

Shipments provide a complete lifecycle from exploration through completion.

To migrate existing conclaves:
  orc conclave migrate --all
```

Then offer to run `/ship-new` with their topic.

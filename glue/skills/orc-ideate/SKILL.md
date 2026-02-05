# ORC Ideate

Rapid idea capture for brainstorming mode.

## Usage

```
/ideate "My quick idea"
```

## Behavior

1. Get focused entity from `orc focus --show`
2. If focused on shipment: use that shipment ID
3. If no shipment focused: error "Focus a shipment first"
4. Create note: `orc note create "<idea>" --shipment <SHIP-xxx> --type idea`
5. Output: "Idea captured: NOTE-xxx"

## Flow

```bash
# Get current focus
orc focus --show
```

Parse output to extract shipment ID:
- If focused on shipment: use that ID
- If focused on commission only: error (need shipment for ideas)
- If no focus: error

```bash
# Create idea note on shipment
orc note create "<idea>" --shipment <SHIP-xxx> --type idea
```

## Example

```
> /ideate "Maybe we need a way to bulk-close notes"

Idea captured: NOTE-603
  Type: idea
  Shipment: SHIP-297
```

## Notes

- Ideas are always created at shipment level (not commission)
- Type is always 'idea'
- Minimal friction - just capture and continue
- Good for quick thoughts during implementation

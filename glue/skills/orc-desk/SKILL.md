---
name: orc-desk
description: Open the desk popup for note review. Use when user says /orc-desk, "review note", "open desk", or wants to edit a note interactively.
---

# ORC Desk

Open a note for interactive review in the desk popup. The desk review workflow
lets you edit note content in vim, with changes persisted back to the database.

## Usage

```
/orc-desk NOTE-xxx     (review specific note)
/orc-desk              (open desk popup for current workbench)
```

## Flow

### Step 1: Determine Target

If a NOTE-xxx argument is provided, use that note ID.

If no argument, check if there is a focused entity:
```bash
orc focus --show
```

### Step 2: Open Review

Run the desk review command:

```bash
orc desk review NOTE-xxx
```

This opens the note in your editor. The file format is:

```
# Note Title

Note content goes here...
```

Edit the title and/or content as needed. Save and quit to persist changes.

### Step 3: Verify

The review command:
- Persists title and content changes back to the database
- Emits a `desk.review.complete` operational event
- Prints confirmation of what was updated

You can verify the changes:
```bash
orc note show NOTE-xxx
```

### Step 4: Resume Work

After review completion, continue with your current task. The operational event
can be used by automated processes to detect that a review is complete.

Check recent events:
```bash
orc events tail --source desk
```

## Notes

- Only NOTE entities are reviewable (other entity types use `orc <type> show` for read-only viewing)
- When running inside the desk TUI (ledger), press `d` on a NOTE line to open a review window
- The review window is ephemeral and closes automatically when vim exits
- If no changes are made, the note is not updated

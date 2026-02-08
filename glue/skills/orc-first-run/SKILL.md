---
name: orc-first-run
description: Interactive first-run walkthrough for new users. Guides through creating first commission, workshop, and shipment.
---

# ORC First Run

Interactive walkthrough for new ORC users. Uses guided execution - the skill runs commands for you while explaining what's happening.

## Usage

```
/orc-first-run
```

Run this after `make bootstrap` to get oriented and create your first project.

## Flow

### Step 1: Welcome

Display welcome message:

```
Welcome to ORC - The Forest Factory!

ORC helps you orchestrate software development work through a hierarchy:

  Factory     → Your development environment
    Workshop  → A project (tmux session with workbenches)
      Workbench → A git worktree where IMPs work
        Shipment  → A unit of work (exploration → tasks → code)

Let's set up your first project together.
```

### Step 2: Create First Commission

A commission is a body of work - think of it as a project or initiative.

Ask the user for a commission name using AskUserQuestion:
- Question: "What would you like to call your first commission?"
- Header: "Commission"
- Options:
  1. "Getting Started" (Recommended) - Good for learning ORC
  2. "My Project" - Generic project name
  3. Other - Enter custom name

Then run:
```bash
orc commission create "<name>"
```

Capture the COMM-xxx ID and explain:
```
Created commission: COMM-xxx

A commission groups related work together. You might have one for
"Q1 Features" or "Backend Refactor" - whatever makes sense for your work.
```

### Step 3: Create First Workshop

A workshop is where the actual work happens - it's a tmux session with workbenches.

Ask the user for a workshop name using AskUserQuestion:
- Question: "What would you like to call your first workshop?"
- Header: "Workshop"
- Options:
  1. "my-workshop" (Recommended) - Simple default
  2. "dev" - Short and common
  3. Other - Enter custom name

Then run:
```bash
orc workshop create "<name>" --commission COMM-xxx
```

Explain:
```
Created workshop: WORK-xxx

A workshop is a tmux session where you and your IMP agents work.
Each workshop can have multiple workbenches (git worktrees).
```

### Step 4: Explain TMux Layout

Display explanation of tmux navigation (no commands needed yet):

```
TMux Navigation

When you connect to a workshop, you'll see tmux windows:
- Window 0: Gatehouse (Goblin/coordinator view)
- Window 1+: Workbenches (where IMPs work)

Key bindings:
- Ctrl+b n     → Next window
- Ctrl+b p     → Previous window
- Ctrl+b w     → Window list
- Right-click  → ORC context menu

You can connect to your workshop later with:
  orc tmux connect WORK-xxx
```

### Step 5: Create First Shipment

A shipment is a unit of work that moves from exploration to implementation.

```bash
orc shipment create "My First Shipment" --commission COMM-xxx --description "Learning ORC"
```

Then focus it:
```bash
orc focus SHIP-xxx
```

Explain:
```
Created and focused: SHIP-xxx

A shipment tracks work through stages:
  exploring → tasked → implementing → deployed → complete

You're now focused on this shipment. Let's add a note.
```

### Step 6: Add First Note

```bash
orc note create "Welcome note" --shipment SHIP-xxx --type idea --content "This is my first note in ORC. I'm learning the system!"
```

Explain:
```
Added note: NOTE-xxx

Notes capture ideas, decisions, and learnings as you work.
They help you (and your IMP) remember context later.
```

### Step 7: Show Summary

```bash
orc summary
```

Explain:
```
The summary shows your work hierarchy. You'll use this constantly to
see what's happening across your commissions and shipments.
```

### Step 8: Quick Reference Card

Display completion message and reference:

```
You're all set! Here's a quick reference:

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Essential Commands:
  orc summary              See all your work
  orc status               Current context
  orc focus SHIP-xxx       Focus on a shipment
  orc prime                Restore context (start of session)

Shipment Workflow:
  /ship-new "Title"        Create new shipment
  /ship-plan               Plan tasks from notes
  /imp-start               Begin autonomous work

Getting Help:
  /orc-help                Show all ORC skills
  orc doctor               Check environment health

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Happy building!
```

## Notes

- Uses AskUserQuestion for interactive prompts with suggestions
- Runs all orc commands directly (guided execution)
- Does not reference Watchtower (deferred feature)
- Designed to run after `make bootstrap`

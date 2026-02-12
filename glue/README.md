# ORC Glue

Glue between ORC and Claude Code - skills, hooks, and other integrations.

## Structure

```
glue/
├── skills/           # Claude Code skills (globally deployed)
├── hooks/            # Claude Code hooks (empty after orc-debug removal)
└── hooks.json        # Hook configuration (Stop hook only)
```

## Deployment

Skills are deployed by copying to `~/.claude/skills/`:

```bash
make deploy-glue
```

This copies all skills to the global Claude Code skills directory. Changes take effect on next skill invocation (hot reload - no restart needed).

## Adding Skills

1. Create `glue/skills/<skill-name>/SKILL.md`
2. Run `make deploy-glue`
3. Use `/skill-name` in Claude Code

## SKILL.md Format

```markdown
---
name: skill-name
description: Clear description of when to use this skill. Include trigger words and use cases.
---

# Skill Title

Instructions for Claude when this skill is invoked.
```

## Why Not Plugin Marketplace?

The Claude Code plugin marketplace doesn't support autocomplete for directory-based sources. Direct deployment to `~/.claude/skills/` provides:

- ✓ Autocomplete support
- ✓ Hot reload (re-deploy updates without restart)
- ✓ Simple copy-based deployment

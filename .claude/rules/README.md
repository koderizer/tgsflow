# Modular Rules (`.claude/rules/`)

Claude Code loads rules from this directory automatically. Each file can be **path-scoped** via YAML frontmatter — rules only load when the agent touches matching files.

## Why Modular Rules?

- CLAUDE.md should be <200 lines (community consensus, Anthropic best practice)
- Path-scoped rules reduce token waste — Rust rules don't load for TypeScript work
- Rules with "**Why:**" rationale generalize better to novel situations
- Deterministic hooks (100% enforcement) complement advisory rules (~80% adherence)

## Structure

```yaml
---
description: "One-line description shown in rule listings"
paths:                          # Optional — omit for global rules
  - "packages/backend/**"
  - "packages/shared/**"
---

## Rule Title

Rule content here. Every rule should have a "**Why:**" line.
```

## Examples

- `example-data-privacy.md` — Global hard rule (no path scope)
- `example-conventions.md` — Path-scoped to specific directories

## Integration with CLAUDE.md

Keep CLAUDE.md lean (50-100 lines). Reference rules:

```markdown
### Rules (loaded from `.claude/rules/`)
- **Data privacy** → `.claude/rules/data-privacy.md`
- **Code conventions** → `.claude/rules/conventions.md`
```

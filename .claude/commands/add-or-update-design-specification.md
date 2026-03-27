---
name: add-or-update-design-specification
description: Workflow command scaffold for add-or-update-design-specification in ecs-engine.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /add-or-update-design-specification

Use this workflow when working on **add-or-update-design-specification** in `ecs-engine`.

## Goal

Adds or updates one or more design specification documents for engine systems or features, and registers or updates them in the main INDEX.md.

## Common Files

- `.design/main/specifications/*.md`
- `.design/main/INDEX.md`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Create or update one or more files in .design/main/specifications/
- Update .design/main/INDEX.md to register the new or changed specification(s)
- Optionally update related cross-references in other spec files

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.
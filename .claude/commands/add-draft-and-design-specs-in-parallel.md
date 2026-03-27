---
name: add-draft-and-design-specs-in-parallel
description: Workflow command scaffold for add-draft-and-design-specs-in-parallel in ecs-engine.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /add-draft-and-design-specs-in-parallel

Use this workflow when working on **add-draft-and-design-specs-in-parallel** in `ecs-engine`.

## Goal

Adds both draft documents (early-stage or research) and formal design specifications for new features or systems, often for physics or engine subsystems.

## Common Files

- `.design/main/specifications/*.md`
- `.draft/**/*.md`
- `.design/main/INDEX.md`
- `.design/main/PLAN.md`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Create or update one or more files in .design/main/specifications/
- Create or update corresponding draft documents in .draft/ (e.g., .draft/physics/...)
- Update .design/main/INDEX.md and/or .design/main/PLAN.md to reference the new work

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.
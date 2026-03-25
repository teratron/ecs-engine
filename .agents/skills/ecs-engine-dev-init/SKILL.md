---
name: ecs-engine-dev-init
description: Initialize ECS engine development environment with junctions and hardlinks.
---

# ECS Engine Dev Init Skill

This skill sets up the mapping between project-level directories and agent-level interfaces. It creates junctions and hardlinks so Claude Code, Gemini, and Qwen can discover workflows, skills, and rules through `.claude/`, `.qwen/`, and respective instruction files, while keeping `.agents/` as the canonical source.

## Bootstrap (First Run)

This skill cannot be invoked via `/ecs-engine-dev-init` until the environment is initialized — `.claude/skills/` does not exist yet. Run the setup script directly from the project root:

**Windows:**
```powershell
powershell -NoProfile -File .agents/skills/ecs-engine-dev-init/scripts/setup_windows.ps1
```

**Unix:**
```bash
bash .agents/skills/ecs-engine-dev-init/scripts/setup_unix.sh
```

After the script runs, `.claude/skills/` will be a junction pointing to `.agents/skills/`, and the skill becomes available as `/ecs-engine-dev-init` for subsequent re-initialization.

## What It Does

1. Links `.claude/` and `.qwen/` subdirectories (`commands`, `skills`, `rules`) to `.agents/` counterparts (`workflows`, `skills`, `rules`).
2. Creates hardlinks `CLAUDE.md`, `GEMINI.md`, `QWEN.md` → `AGENTS.md`.
3. Removes linked paths from the git index to prevent accidental tracking.
4. Verifies all junctions and hardlinks resolve correctly.

## Resources

- [scripts/setup_windows.ps1](scripts/setup_windows.ps1) - PowerShell command sequence for Windows.
- [scripts/setup_unix.sh](scripts/setup_unix.sh) - Bash command sequence for Linux/macOS.

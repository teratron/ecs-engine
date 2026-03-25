---
name: ecs-engine-dev-init
description: Initialize ECS engine development environment with junctions and hardlinks.
---

# ECS Engine Dev Init Skill

This skill sets up the mapping between project-level directories and agent-level interfaces. It creates junctions and hardlinks so Claude Code, Gemini, and Qwen can discover workflows, skills, and rules through `.claude/`, `.qwen/`, and respective instruction files, while keeping `.agents/` as the canonical source.

## Procedures

### 1. Identify Environment

Identify the operating system (Windows vs. Unix/macOS).

### 2. Execute Initialization Script

Establish all links using the platform-specific script. Key tasks include:

- Linking `.claude/commands`, `.claude/skills`, `.claude/rules`, `.qwen/commands`, `.qwen/skills`, and `.qwen/rules` to `.agents/` subdirectories.
- Linking root `CLAUDE.md`, `GEMINI.md`, and `QWEN.md` to `AGENTS.md` as hardlinks so agent instructions stay in sync.
- Maintaining the git index by removing linked paths from tracking.

**On Windows (PowerShell):**

```powershell
pwsh -NoProfile -File .agents/skills/ecs-engine-dev-init/scripts/setup_windows.ps1
```

**On Unix (Bash):**

```bash
bash .agents/skills/ecs-engine-dev-init/scripts/setup_unix.sh
```

### 3. Verification

Confirm that all junctions resolve correctly and that `CLAUDE.md`, `GEMINI.md`, `QWEN.md` hardlinks all point to `AGENTS.md`.

## Resources

- [scripts/setup_windows.ps1](scripts/setup_windows.ps1) - PowerShell command sequence for Windows.
- [scripts/setup_unix.sh](scripts/setup_unix.sh) - Bash command sequence for Linux/macOS.

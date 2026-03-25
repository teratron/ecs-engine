#!/usr/bin/env bash

# ───────────────────────────────────────────────────────────────────────────────
# ECS ENGINE DEV INIT (UNIX)
# ───────────────────────────────────────────────────────────────────────────────

set -e

echo ">>> Initializing Unix/macOS Agent Environment..."

# 1. Git Index Maintenance (MUST run BEFORE creating symlinks)
echo "Synchronizing git index (pre-link)..."
agentFiles="CLAUDE.md GEMINI.md QWEN.md"
links=".claude/commands .claude/skills .claude/rules .qwen/commands .qwen/skills .qwen/rules"
for f in $agentFiles; do
  links="$links $f"
done
git rm -r --cached --ignore-unmatch $links 2>/dev/null || true

# 2. Create agent symlinks (.claude)
mkdir -p .claude

rm -rf .claude/commands .claude/skills .claude/rules
ln -s ../.agents/workflows .claude/commands
ln -s ../.agents/skills .claude/skills
ln -s ../.agents/rules .claude/rules

# 3. Create agent symlinks (.qwen)
mkdir -p .qwen

rm -rf .qwen/commands .qwen/skills .qwen/rules
ln -s ../.agents/workflows .qwen/commands
ln -s ../.agents/skills .qwen/skills
ln -s ../.agents/rules .qwen/rules

# 4. Global Agent Instructions (hardlinks to AGENTS.md)
echo "Linking agent instruction files..."
for f in $agentFiles; do
  rm -f $f
  ln AGENTS.md $f
done

echo -e "\n>>> Verification:"
verifyLinks=".claude/commands .claude/skills .claude/rules .qwen/commands .qwen/skills .qwen/rules"
for f in $agentFiles; do
  verifyLinks="$verifyLinks $f"
done
ls -ld $verifyLinks

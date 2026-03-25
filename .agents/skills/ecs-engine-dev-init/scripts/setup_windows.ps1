# ═══════════════════════════════════════════════════════════════════════════════
# ECS ENGINE DEV INIT (WINDOWS)
# ═══════════════════════════════════════════════════════════════════════════════

# Safe junction/hardlink creation for Windows environment.
# Note: Junctions (/J) work without admin or Developer Mode.
# Hardlinks (/H) also work without admin for files.

# ───────────────────────────────────────────────────────────────────────────────
# 1. Configuration
# ───────────────────────────────────────────────────────────────────────────────

$agentFiles = @("CLAUDE.md", "GEMINI.md", "QWEN.md")

# ───────────────────────────────────────────────────────────────────────────────
# 2. Cleanup function
# ───────────────────────────────────────────────────────────────────────────────

function Remove-Existing($path) {
    if (Test-Path $path) {
        Write-Host "Removing: $path" -ForegroundColor Yellow
        if ((Get-Item $path).Attributes -match "ReparsePoint") {
            if ((Get-Item $path).PSIsContainer) { cmd /c "rmdir ""$path""" } else { cmd /c "del ""$path""" }
        } else {
            Remove-Item -Recurse -Force $path
        }
    }
}

# ───────────────────────────────────────────────────────────────────────────────
# 3. Main Execution
# ───────────────────────────────────────────────────────────────────────────────

Write-Host ">>> Initializing Windows Agent Environment..." -ForegroundColor Cyan

# 3.1. Git Index Maintenance (MUST run BEFORE creating junctions)
Write-Host "Synchronizing git index (pre-link)..." -ForegroundColor Cyan
$linksToRemove = @(
    ".claude/commands",
    ".claude/skills",
    ".claude/rules",
    ".qwen/commands",
    ".qwen/skills",
    ".qwen/rules"
)
foreach ($f in $agentFiles) { $linksToRemove += "$f" }
git rm -r --cached --ignore-unmatch $linksToRemove 2>$null

# 3.2. .claude junctions
if (-not (Test-Path ".claude")) { New-Item -ItemType Directory -Path ".claude" -Force }
Remove-Existing ".claude\commands"
Remove-Existing ".claude\skills"
Remove-Existing ".claude\rules"
cmd /c 'mklink /J ".claude\commands" ".agents\workflows"'
cmd /c 'mklink /J ".claude\skills" ".agents\skills"'
cmd /c 'mklink /J ".claude\rules" ".agents\rules"'

# 3.3. .qwen junctions
if (-not (Test-Path ".qwen")) { New-Item -ItemType Directory -Path ".qwen" -Force }
Remove-Existing ".qwen\commands"
Remove-Existing ".qwen\skills"
Remove-Existing ".qwen\rules"
cmd /c 'mklink /J ".qwen\commands" ".agents\workflows"'
cmd /c 'mklink /J ".qwen\skills" ".agents\skills"'
cmd /c 'mklink /J ".qwen\rules" ".agents\rules"'

# 3.4. Global Agent Instructions (hardlinks to AGENTS.md)
Write-Host "Linking agent instruction files..." -ForegroundColor Cyan
foreach ($f in $agentFiles) {
    Remove-Existing $f
    cmd /c "mklink /H ""$f"" AGENTS.md"
}

Write-Host "`n>>> Verification:" -ForegroundColor Green
cmd /c "dir .claude\commands .claude\skills .claude\rules .qwen\commands .qwen\skills .qwen\rules /AL"

Write-Host "`n>>> Hardlink Integrity Check (AGENTS.md):" -ForegroundColor Cyan
cmd /c "fsutil hardlink list AGENTS.md"

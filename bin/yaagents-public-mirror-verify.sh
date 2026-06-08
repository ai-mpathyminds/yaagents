#!/usr/bin/env bash
# bin/yaagents-public-mirror-verify.sh
# PI3-yaa TRACK SCRUB acceptance: verify no internal AimpathyMinds planning
# artifacts leaked into any public-mirror working-tree.
#
# Usage:
#   bin/yaagents-public-mirror-verify.sh <path1> [path2 ...]
#
#   Each path is a local working-tree root of one of the 8 public repos.
#   Run AFTER SC-1 + SC-2 (scrub WIs) land; BEFORE LA-PUBLIC-FLIP fires.
#
# Returns:
#   exit 0  — 0 hits on all repos (PASS; safe to proceed with LA-PUBLIC-FLIP)
#   exit 1  — 1+ hits found (FAIL; fix SCRUB WIs before flipping repos public)
#   exit 2  — bad argument / path not found
#
# Rule ref:
#   yaagents/docs/PI3-yaa/roadmap.md §Success criteria #8
#   yaagents/docs/PI3-yaa/platform-engineer.md WI-3yaa.NFR-SUP-1
#   portfolio/RUNBOOKS/pi3-yaa-planning.yml A-4 exit_check
#
# CI integration:
#   .github/workflows/scrub-verify.yml — triggered on push to meta-repo main.
#   Passes each of 8 checked-out submodule paths as args; fails PR if non-zero.
#
# Exclusions:
#   - CHANGELOG* files: may legitimately reference past PI cycles as version history.
#   - .git/ directories: always skipped.
#   - Binary file extensions: png/jpg/gif/woff/ttf/otf/ico/svg.

set -uo pipefail

# ── colour helpers ───────────────────────────────────────────────────────────
_red()    { printf '\033[31m%s\033[0m' "$*"; }
_green()  { printf '\033[32m%s\033[0m' "$*"; }
_yellow() { printf '\033[33m%s\033[0m' "$*"; }

pass_line() { printf "  [%s] %s\n" "$(_green PASS)" "$1"; }
fail_line() { printf "  [%s] %s\n" "$(_red   FAIL)" "$1"; }

# ── args ─────────────────────────────────────────────────────────────────────
if [ $# -eq 0 ]; then
  echo "Usage: $(basename "$0") <repo-path1> [repo-path2 ...]" >&2
  echo "" >&2
  echo "Default 8-repo set (relative to portfolio root):" >&2
  echo "  yaagents  yaagents-gateway  yaagents-sdk-fastapi  yaagents-sdk-go" >&2
  echo "  yaagents-client-python  yaagents-client-ts  yaagents-client-go  yaagents-cli" >&2
  exit 2
fi

TOTAL_HITS=0
CHECKED=0

for REPO_PATH in "$@"; do
  if [ ! -d "$REPO_PATH" ]; then
    printf "[%s] path not found: %s\n" "$(_red ERROR)" "$REPO_PATH" >&2
    exit 2
  fi

  REPO_HITS=0
  CHECKED=$((CHECKED + 1))
  echo ""
  echo "── checking: $REPO_PATH ────────────────────────────────────────────────"

  # ── 1. File-system presence checks ────────────────────────────────────────
  if [ -d "$REPO_PATH/.claude" ]; then
    fail_line ".claude/ directory present at repo root"
    REPO_HITS=$((REPO_HITS + 1))
  fi

  if [ -f "$REPO_PATH/CLAUDE.md" ]; then
    fail_line "CLAUDE.md file present at repo root"
    REPO_HITS=$((REPO_HITS + 1))
  fi

  if [ -d "$REPO_PATH/system-refs" ]; then
    fail_line "system-refs/ directory present at repo root"
    REPO_HITS=$((REPO_HITS + 1))
  fi

  if [ -d "$REPO_PATH/portfolio" ]; then
    fail_line "portfolio/ directory present at repo root"
    REPO_HITS=$((REPO_HITS + 1))
  fi

  # PI*-yaa directories anywhere in the tree
  PI_DIRS=$(find "$REPO_PATH" -not -path "*/.git/*" -name "PI[0-9]*-yaa" -type d 2>/dev/null || true)
  if [ -n "$PI_DIRS" ]; then
    fail_line "PI*-yaa directories found:"
    echo "$PI_DIRS" | sed 's/^/    /'
    REPO_HITS=$((REPO_HITS + 1))
  fi

  # *.seed.md / *_detailed.md / *_onepager.md anywhere in tree
  LEAK_FILES=$(find "$REPO_PATH" -not -path "*/.git/*" \
    \( -name "*.seed.md" -o -name "*_detailed.md" -o -name "*_onepager.md" \) \
    2>/dev/null || true)
  if [ -n "$LEAK_FILES" ]; then
    fail_line "Internal PRD artifact files found:"
    echo "$LEAK_FILES" | sed 's/^/    /'
    REPO_HITS=$((REPO_HITS + 1))
  fi

  # ── 2. Content grep (exclude CHANGELOG* + .git/ + binary extensions) ──────
  GREP_EXCLUDE_DIR="--exclude-dir=.git"
  # Exclude this script itself from grep — it necessarily contains all 5 patterns
  # as literal pattern strings + commentary, which would self-match (round 3 found this).
  GREP_EXCLUDE_FILES="--exclude=yaagents-public-mirror-verify.sh --exclude=CHANGELOG* --exclude=*.png --exclude=*.jpg \
--exclude=*.gif --exclude=*.woff --exclude=*.woff2 --exclude=*.ttf \
--exclude=*.otf --exclude=*.ico --exclude=*.svg"

  # Allowlist filter for the PI pattern only — accepts legitimate ADR/regression citations.
  PI_ALLOWLIST_FILTER='ADR PI[0-9]+-yaa-[0-9]{4}|PI[0-9]+-yaa regression|spec/VERSION = [0-9.]+ \(ADR PI[0-9]+-yaa-[0-9]{4}'

  for PATTERN in '\.claude/' 'portfolio/' 'PI[0-9][0-9]*-yaa' 'system-refs/' 'CLAUDE\.md'; do
    # shellcheck disable=SC2086
    HITS=$(grep -rn $GREP_EXCLUDE_DIR $GREP_EXCLUDE_FILES \
      --include="*.go" --include="*.py" --include="*.ts" \
      --include="*.md" --include="*.yaml" --include="*.yml" \
      --include="*.json" --include="*.toml" --include="*.mod" \
      "$PATTERN" "$REPO_PATH" 2>/dev/null | head -20 || true)
    # Allowlist post-filter applies ONLY to the PI*-yaa content pattern
    if [ "$PATTERN" = 'PI[0-9][0-9]*-yaa' ] && [ -n "$HITS" ]; then
      HITS=$(echo "$HITS" | grep -vE "$PI_ALLOWLIST_FILTER" || true)
    fi
    if [ -n "$HITS" ]; then
      fail_line "content pattern '$PATTERN' found:"
      echo "$HITS" | sed 's/^/    /'
      REPO_HITS=$((REPO_HITS + 1))
    fi
  done

  if [ "$REPO_HITS" -eq 0 ]; then
    pass_line "$REPO_PATH — clean"
  fi

  TOTAL_HITS=$((TOTAL_HITS + REPO_HITS))
done

# ── summary ──────────────────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════════════════════════════════"
echo "  Checked : $CHECKED repo(s)"
echo "  Hits    : $TOTAL_HITS"

if [ "$TOTAL_HITS" -eq 0 ]; then
  printf "  Result  : %s\n" "$(_green "PASS — no internal artifacts leaked")"
  echo "  Action  : Safe to proceed with LA-PUBLIC-FLIP (flip 8 repos PRIVATE→PUBLIC)."
  exit 0
else
  printf "  Result  : %s\n" "$(_red "FAIL — $TOTAL_HITS hit(s); scrub incomplete")"
  echo "  Action  : Fix SC-1 / SC-2 WIs. Re-run this script before LA-PUBLIC-FLIP."
  exit 1
fi

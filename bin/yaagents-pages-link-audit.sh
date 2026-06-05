#!/usr/bin/env bash
# yaagents-pages-link-audit.sh
#
# Audit internal Markdown links in yaagents Starlight docs.
# Catches broken /yaagents/... hrefs before they reach the live site.
#
# Origin: PC-5-07 (PI3-yaa, 2026-06-05) — systemic discovery that
# B-45h sidebar restructure introduced 5 broken links by writing
# sidebar group names as URL paths. Starlight URL = file path, not
# sidebar group name. Five concrete defects: /start-here/why-yaagents/,
# /reference/profile-v03/, /start-here/production-agent-api-checklist/,
# /examples/campaign-api/ — all 404 on the live site.
#
# Usage:
#   bash bin/yaagents-pages-link-audit.sh           # check built docs/dist/
#   bash bin/yaagents-pages-link-audit.sh --live    # force curl against live Pages URL
#
# GHA context: run this step AFTER pnpm build (docs/dist/ present) and
# BEFORE actions/upload-pages-artifact.
#
# Exit: 0 = clean; 1 = broken link(s) found. Broken links to stderr.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Locate yaagents docs source and built dist from either location:
#   - portfolio/bin/  → ../yaagents/docs
#   - yaagents/bin/   → ../docs
if [[ -d "$SCRIPT_DIR/../yaagents/docs/src/content/docs" ]]; then
  DOCS_SRC="$SCRIPT_DIR/../yaagents/docs/src/content/docs"
  DIST_DIR="$SCRIPT_DIR/../yaagents/docs/dist"
elif [[ -d "$SCRIPT_DIR/../docs/src/content/docs" ]]; then
  DOCS_SRC="$SCRIPT_DIR/../docs/src/content/docs"
  DIST_DIR="$SCRIPT_DIR/../docs/dist"
else
  echo "ERROR: cannot locate yaagents docs/src/content/docs from $SCRIPT_DIR" >&2
  exit 1
fi

USE_LIVE=0
[[ "${1:-}" == "--live" ]] && USE_LIVE=1

LIVE_BASE="https://ai-mpathyminds.github.io"

broken=0
checked=0
declare -a broken_links=()

while IFS= read -r mdx_file; do
  # Extract all /yaagents/... link targets from Markdown href syntax ](...)
  while IFS= read -r link; do
    [[ -z "$link" ]] && continue
    checked=$((checked + 1))

    if [[ $USE_LIVE -eq 0 ]] && [[ -d "$DIST_DIR" ]]; then
      # Check built dist: /yaagents/some/path/ → docs/dist/some/path/index.html
      rel="${link#/yaagents}"   # strip /yaagents prefix
      rel="${rel%/}"            # strip trailing slash
      target="${DIST_DIR}${rel}/index.html"
      if [[ ! -f "$target" ]]; then
        lineno=$(grep -n "](${link}" "$mdx_file" 2>/dev/null | head -1 | cut -d: -f1)
        relfile="${mdx_file#$DOCS_SRC/}"
        broken_links+=("BROKEN: ${link}  in  ${relfile}:${lineno:-?}")
        broken=$((broken + 1))
      fi
    else
      # Curl against live Pages site
      url="${LIVE_BASE}${link}"
      http_code=$(curl -fIs --max-time 15 -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
      if [[ "$http_code" != "200" ]]; then
        lineno=$(grep -n "](${link}" "$mdx_file" 2>/dev/null | head -1 | cut -d: -f1)
        relfile="${mdx_file#$DOCS_SRC/}"
        broken_links+=("BROKEN: ${link}  (HTTP ${http_code})  in  ${relfile}:${lineno:-?}")
        broken=$((broken + 1))
      fi
    fi
  done < <(grep -oE '\]\(/yaagents/[a-z0-9./_-]+(/?)?\)' "$mdx_file" 2>/dev/null \
            | sed 's/^](\(.*\))$/\1/' \
            | sort -u)
done < <(find "$DOCS_SRC" -name "*.mdx" | sort)

if [[ $broken -gt 0 ]]; then
  printf '%s\n' "${broken_links[@]}" >&2
  echo "LINK-AUDIT: FAIL — ${broken} broken link(s) found (checked ${checked} links across docs)" >&2
  exit 1
fi

echo "LINK-AUDIT: PASS — ${checked} internal /yaagents/ links verified OK"

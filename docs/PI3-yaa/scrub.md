# PI3-yaa — Component: Portfolio scrub (internal artifacts move + cleanup)

Owner lane: **operator-driven / mechanical** (file moves + `git rm`). Sprint 1 (must land
before RP-* WIs in S3-S4, because the submodule init copies use working-tree-mirror semantics
and any internal artifact still present in `yaagents/` would leak into the public submodule
during `cp -r`).

> **Library gate (Gate 3) — applies to SC-* WIs**: `library_ref: ADR PI3-yaa-0003` (orphan-baseline
> squashed history; trade-off of commit-history-loss for cleanliness is the gating decision).
>
> The companion verification script `bin/yaagents-public-mirror-verify.sh` is authored by
> **platform-engineer at A-4** per planning runbook A-4 artifact spec — NOT in this file.

---

### WI-3yaa.SC-1: Move internal planning artifacts into `portfolio/yaagents-internal/` [DRAFT] — Sprint 1
service: portfolio/yaagents-internal (target) + yaagents (source)
parent_feature: F-SCRUB
brief: Move every internal planning artifact out of `yaagents/` into
`portfolio/yaagents-internal/` (a new directory in the private portfolio repo). This is a
one-shot move — internal artifacts MUST NOT be in the public-bound source tree at any point
after this WI lands. Per PRD §7.1 mapping:

| Source (current `yaagents/`) | Destination (`portfolio/yaagents-internal/`) |
|------------------------------|---------------------------------------------|
| `docs/PI*-yaa/` (PI1-yaa, PI2-yaa, PI3-yaa rollups) | `docs/PI*-yaa/` |
| `docs/adr/PI*-yaa-*.md` (all internal ADRs) | `adr/` |
| `system-refs/*.seed.md` (chief-architect seeds) | `seeds/` |
| `system-refs/*_detailed.md` (full PRDs) | `specs/` |
| `system-refs/*_onepager.md` (PRD one-pagers) | `specs/` |

**Public-surface artifacts that STAY in `yaagents/`** (do NOT move):
- `yaagents/system-refs/YAAgents_PRD_README.md` — community-facing PRD overview (per intake §A-2 + PRD §5.1; this file is **public-facing** and carries the normative Response Profile table).
- `yaagents/system-refs/YAAgents_GTM_README.md` — GTM positioning doc; **public-facing** per GTM README §12 license-and-disclaimer footer.
- `yaagents/system-refs/gtm-protocol-positioning.seed.md` — wait, this IS a seed. **MOVES** to `portfolio/yaagents-internal/seeds/`.

Concrete commands (run from portfolio repo root, before any PI3-yaa S2 work):
```bash
mkdir -p portfolio/yaagents-internal/{docs,adr,seeds,specs}
git -C portfolio mv yaagents/docs/PI1-yaa portfolio/yaagents-internal/docs/
git -C portfolio mv yaagents/docs/PI2-yaa portfolio/yaagents-internal/docs/
git -C portfolio mv yaagents/docs/PI3-yaa portfolio/yaagents-internal/docs/
git -C portfolio mv yaagents/docs/adr/PI*-yaa-*.md portfolio/yaagents-internal/adr/
git -C portfolio mv yaagents/system-refs/*.seed.md portfolio/yaagents-internal/seeds/
git -C portfolio mv yaagents/system-refs/*_detailed.md portfolio/yaagents-internal/specs/
git -C portfolio mv yaagents/system-refs/*_onepager.md portfolio/yaagents-internal/specs/
# YAAgents_PRD_README.md + YAAgents_GTM_README.md STAY in yaagents/system-refs/ (public-facing)
```

**Operator note**: `yaagents/` is a separate git repo (per portfolio convention "each product is a
separate git repository — never make cross-product commits"). So this move is actually TWO commits:
- 1 commit in `yaagents/` repo with `git rm` of all internal artifacts.
- 1 commit in `portfolio/` (private workspace) repo creating `portfolio/yaagents-internal/` with
  the moved content.

The two commits MUST land together (PR-pair or atomic operator action) so no window exists where
internal artifacts are deleted from `yaagents/` but not yet captured in `portfolio/yaagents-internal/`.

**Important: this WI lands on the `pi3-yaa` branch of `yaagents/` repo BEFORE any submodule split
happens.** Once SC-1 lands on `pi3-yaa`, the working tree is clean and ready for RP-* per-submodule
mirroring.

acceptance:
- `find yaagents/docs/ -name 'PI*-yaa*'` returns 0 hits (all moved out).
- `find yaagents/docs/adr/ -name 'PI*-yaa-*.md'` returns 0 hits.
- `find yaagents/system-refs/ -name '*.seed.md' -o -name '*_detailed.md' -o -name '*_onepager.md'` returns 0 hits.
- `find portfolio/yaagents-internal/ -name 'PI*-yaa*' -o -name '*.seed.md' -o -name '*_detailed.md' -o -name '*_onepager.md'` shows the moved content.
- `yaagents/system-refs/YAAgents_PRD_README.md` + `yaagents/system-refs/YAAgents_GTM_README.md` REMAIN (public-facing; do not move).
- Both commits land on consistent branches (`yaagents/pi3-yaa` + `portfolio/pi3-yaa-prep` or equivalent operator-chosen branch); no orphan state where internal artifacts are deleted without being captured.
- bin/yaagents-public-mirror-verify.sh (authored by platform-engineer at A-4) on a fresh `yaagents/` working tree returns 0 portfolio-marker hits after this WI lands (verified at A-4 dry-run).
library_ref: ADR PI3-yaa-0003 (orphan-baseline squashed history + portfolio-scrub trade-off).
depends_on: [WI-3yaa.SP-1]

### WI-3yaa.SC-2: Delete `yaagents/.claude/` + `yaagents/CLAUDE.md` from working tree [DRAFT] — Sprint 1
service: yaagents
parent_feature: F-SCRUB
brief: Remove `yaagents/.claude/` directory and `yaagents/CLAUDE.md` from the `yaagents/` repo
working tree. Per intake §A-4 + PRD §7.1: the `.claude/` resolves to the portfolio-root
`.claude/` directory in operator usage (parent directory traversal); the per-product
`yaagents/CLAUDE.md` consolidates into the portfolio-root `CLAUDE.md`'s yaagents section.

Concrete commands (from `yaagents/` repo root):
```bash
git rm -rf .claude/
git rm CLAUDE.md
git commit -m "Scrub yaagents/.claude + yaagents/CLAUDE.md ahead of public mirror"
```

**Pre-condition** (verified at A-4): the agents that were previously read from
`yaagents/.claude/agents/` (yaagents-architect, yaagents product-manager, yaagents-side
platform-engineer, go-developer, python-developer, frontend-developer) MUST be either:
- (a) Re-homed to `portfolio/yaagents-internal/.claude/agents/` OR
- (b) Reduced to nothing (lane-table entry in `portfolio-conventions.md` already maps the
  agents to writable yaagents paths; the agent file content is the only thing that moves).

**Recommended path**: option (a) — preserves the agent-body content under
`portfolio/yaagents-internal/.claude/agents/` so future PI{n}-yaa planning can continue to load
the agents. Operator runs SC-2 with a paired commit in `portfolio/` repo that creates
`portfolio/yaagents-internal/.claude/agents/` with the moved content.

acceptance:
- `ls yaagents/.claude/ 2>&1 | head -1` returns "No such file or directory".
- `ls yaagents/CLAUDE.md 2>&1` returns "No such file or directory".
- `bin/sync-portfolio-skills.sh` (which previously synced into `yaagents/.claude/skills/`) is updated to skip yaagents OR yaagents is removed from its default product list (operator-driven; the change lands in portfolio repo). This is a downstream cross-ref handled at A-4 NFR pass.
- The agent-body content has been preserved in `portfolio/yaagents-internal/.claude/agents/` (option a path).
- bin/yaagents-public-mirror-verify.sh returns 0 `.claude/` hits + 0 `CLAUDE.md` hits on the `yaagents/` repo.
library_ref: ADR PI3-yaa-0003 (orphan-baseline squashed history + portfolio-scrub trade-off).
depends_on: [WI-3yaa.SC-1]

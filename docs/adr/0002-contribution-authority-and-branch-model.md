# ADR 0002 — The meta-repo's root `CONTRIBUTING.md` governs all eight repositories, and branch names carry no mandated prefix

Status: Accepted (proposed by the architect agent; ratified by the reviewer who merges it)
Date: 2026-09-10
Work item: `01a089f7-38ce-7148-86ad-afcf678e82b9` · Packet: `01a08c2f-01ff-7194-8c32-dd73dd3b7f47`
Depends on: ADR 0001 (`docs/adr/0001-decision-record-location.md`, artifact
`01a089fd-b3a3-7310-b610-8cb401882677`), which sets this folder and this file's form. It is
committed at `db420a2` on branch `wakalix/wi-01a089f738cb` and is **not yet on `main`**.
Supersedes: nothing in `docs/adr/`. It overrides the issue-routing text in the `gateway` and
`sdk-go` contribution documents, and line 81 of the root `CONTRIBUTING.md`; see *Decision*.

## Context

YAAgents is one meta-repository and seven submodules (`.gitmodules`): `gateway`, `sdk-go`,
`sdk-fastapi`, `client-python`, `client-ts`, `client-go`, `cli`. Each submodule is its own
GitHub repository, so a contributor can arrive at any of the eight without passing through the
others. Each has a `CONTRIBUTING.md`, in four shapes:

| Shape | Where | What it says | Read by me? |
|---|---|---|---|
| Full guide, 206 lines | root `CONTRIBUTING.md` | All issues go to the meta-repo tracker (lines 12–15); `feature/`, `fix/`, `docs/` branch prefixes (line 81); Conventional Commits and a DCO sign-off (lines 83–109) | Yes, in full |
| Near-complete second copy | `client-ts/CONTRIBUTING.md` | The root's content, maintained separately, so it can drift without anyone noticing | **No** — as the packet describes it |
| 18-line "file bugs on this repo" | `gateway/`, `sdk-go/` | Issues go to that component's own repository, which contradicts the root | **No** — as the packet describes it |
| 9-line pointer stub | `cli/`, `client-go/`, `client-python/`, `sdk-fastapi/` | Points at the root | **No** — as the packet describes it |

Nothing in any of the trees says which document wins. A coding agent asked to change
`client-ts` today has two standards and no rule for choosing between them.

**What the meta-repo already does** decides most of this:

- **The public site already treats the root as the only source.**
  `docs/src/components/ContributingContent.astro` reads `../../../CONTRIBUTING.md` at build
  time and renders it; its header says "Single source of truth — no copy/paste".
  `docs/src/content/docs/contributing.mdx` is that component plus a legal caution, and it is
  the *Contributing* sidebar entry (`docs/astro.config.mjs:65`). `community.mdx` and
  `plugins.mdx` both send readers there.
- **`README.md` points at the root** for contributing (lines 162–164) and at the meta-repo
  tracker for issues (line 138).
- **The meta-repo tracker was built to take issues for every component.** The issue forms live
  in the meta-repo's `.github/ISSUE_TEMPLATE/`, and `bug.yml` opens with a *Component* dropdown.
  Line 30 of the root asks every bug report to name one of the eight components.
- **Four of the seven submodules already use the pointer shape.**

**The branch rule is written down but nobody follows it:**

- Commit `684240f` (2026-06-07) added `feature/`, `fix/`, `docs/` to the root. The earlier
  versions at `2fb028c` and `cbd6df3` had no branch rule.
- The meta-repo's history has 14 pull-request merges, #12 to #25, all after that date. **None**
  came from a prefixed branch: `workflow-fixes`, `scrub-self-exclude`,
  `client-go-pointer-bump`, `gotest-advisory` and so on — short, lowercase, kebab-case.
- The other remote branches are unprefixed topic branches, three long-lived per-cycle branches,
  and the automated workflow's `wakalix/wi-<id>` and `wakalix/run-<id>`. The workflow's
  dispatch names those branches, not whoever writes the change. This record itself is
  delivered on `wakalix/wi-01a089f738ce`.
- **Nothing enforces a branch name.** `ci.yml` runs on a push to `'**'` and on every pull
  request. No workflow filters on a prefix.

**Why these are one record.** Once one document governs, every rule in it binds, including the
branch rule that nobody follows. Declaring the root authoritative without deciding its branch
clause would put every branch in the log in breach of the governing document. So the two
questions are the same decision — *what binds a contributor* — and the second follows from
the first.

No requirement set came with this ask, so there is no stated NFR for this record to answer.

## Options considered — which document governs

### A. The root governs all eight; each submodule's `CONTRIBUTING.md` is a pointer to it

- **For — it is the existing pattern.** The site, the README, the issue forms and four of the
  seven submodules already behave this way. This option writes down what is already true
  instead of inventing something new.
- **For — drift becomes impossible, not just unlikely.** A pointer has no rules in it to go
  stale. The `client-ts` problem ends because only one text exists.
- **For — it matches how the product is built and released.** The root's *Dev setup* already
  covers every component (lines 45–75). A release tags all eight at one version, in one wave
  (lines 160–164).
- **Against — one click, not zero.** A contributor who lands on, say, `yaagents-gateway` on
  GitHub reads a pointer, not the rules. The pointer has to use the absolute URL
  `https://github.com/ai-mpathyminds/yaagents/blob/main/CONTRIBUTING.md`. A relative link
  breaks when the submodule is viewed as its own repository.
- **Against — anything specific to one component has to move.** If the `gateway` or `sdk-go`
  documents hold guidance the root lacks, it has to go into the root or into that component's
  `README.md` rather than be deleted. I have not read them, so I cannot say whether they do.
- **Against — one tracker for everything.** Gateway and Go SDK issues share a queue with the
  rest. Triage then depends on the *Component* dropdown and the labels.

### B. Each repository governs itself

- **For:** each component repository stands alone, and `gateway` and `sdk-go` already work this
  way for issues.
- **Against — it contradicts what is already published.** The site, the README and the issue
  forms all send contributors to the meta-repo.
- **Against — it multiplies the drift problem by eight.** The rules every repository shares
  (DCO, the licensing disclaimer, security disclosure, the Code of Conduct) would live in eight
  copies. That is the `client-ts` failure made the norm.
- **Against — it splits bugs that span components.** Releases cross all eight components, and a
  bug found through `client-ts` often lives in `gateway`. A tracker for each repository forces
  the reporter to guess.

### C. The root governs, and each submodule carries a full copy kept in sync mechanically

- **For:** zero clicks from any repository.
- **Against — it needs tooling that does not exist:** a sync job or a diff check across eight
  repositories, and one more gate.
- **Against — a copy is only as current as its submodule pin.** `client-ts` is pinned at
  `7a11174`, so its copy lags until someone bumps the pointer, which means two texts exist in
  every release.
- **Against — the `client-ts` copy is this option without the check,** and that copy is the
  problem this record was asked to settle.

## Options considered — which branch convention wins

### A. Enforce the documented prefixes

- **For:** the rule is already written, and a branch list would show each change's type at a
  glance.
- **Against — every merged pull request in the history would have failed it.**
- **Against — the automated workflow names its own branches** under `wakalix/`. Enforcing the
  prefixes means either an exemption or a change to the dispatcher.
- **Against — it needs a new workflow.** Nothing checks branch names today.
- **Against — the prefix repeats what the pull request's title and labels already say.** For a
  fork-based contribution, the branch lives in the contributor's fork and appears upstream only
  in the merge commit's subject.

### B. Drop the mandate, and describe the practice

- **For — it matches all 14 merged branches and the automation,** with nothing to build and
  nothing to exempt.
- **Against:** the change's type is no longer readable from the branch name. If a tool is ever
  adopted that keys off a branch prefix (none is today), this record should be reopened.

### C. Keep the prefixes as a recommendation rather than a requirement

- **Against — a *should* that nobody follows is the same drift in softer words.** An agent
  reading the document still cannot tell whether to comply. Rejected for that reason.

## Decision

**A for authority, B for branches.**

1. **The root `CONTRIBUTING.md` of `github.com/ai-mpathyminds/yaagents` governs contribution to
   all eight repositories.** Where any other contribution document in any of the eight
   disagrees with it, the root wins.
2. **A submodule's `CONTRIBUTING.md` is a pointer.** It names the root by absolute URL and
   states no rule of its own. Build and test instructions for one component belong in that
   component's `README.md`. The 9-line stub is the target shape; the `client-ts` copy and the
   18-line `gateway` and `sdk-go` documents are to be replaced by it.
3. **Issues for every component go to the meta-repo tracker**, as root lines 12–15 say. The
   "file bugs on this repo" instruction in `gateway` and `sdk-go` no longer holds.
4. **Until those replacements land,** a contributor or agent who finds a submodule contribution
   document reads it only for what the root does not address, and follows the root wherever the
   two conflict.
5. **Branch names have no mandated prefix.** A branch is short, lowercase and kebab-case, and it
   says what the change is. The `wakalix/` namespace is reserved for branches the automated
   workflow creates. Root line 81 is to be rewritten to say this.
6. **Nothing else in the root changes because of this record.** Conventional Commits and the DCO
   sign-off are just as unfollowed as the branch prefixes, but this record deliberately does not
   drop them; see *Found, not acted on*.

## Consequences

- **What it costs is one click from a component repository,** and it only works if each stub
  links the root by absolute URL.
- **Four follow-up changes are needed, none of them made here.** Each change in a submodule is
  its own pull request in that repository, followed by a pointer bump in the meta-repo:
  1. Replace `client-ts/CONTRIBUTING.md`, `gateway/CONTRIBUTING.md` and
     `sdk-go/CONTRIBUTING.md` with the stub. Move anything specific to the component into its
     `README.md` first.
  2. Check that the four existing stubs link the root by absolute URL.
  3. Rewrite root line 81 as clause 5 says.
  4. Bring the PR template, `adapter-request.yml` and `bug.yml` into line with the root; see
     below.
- **Issues will keep arriving in `gateway` and `sdk-go` for as long as their trackers are
  open.** Closing them, or leaving them open with a pointer, is a repository setting and a
  maintainer's call. This record does not make it.
- **The site's copy of the root can lag.** `pages.yml` redeploys on a push to `main` only when
  `docs/**`, `spec/**`, `schemas/**`, `openapi/**` or the workflow itself changes, and its
  pull-request build watches only `docs/**` and the workflow. A change to the root
  `CONTRIBUTING.md` alone therefore neither rebuilds the published *Contributing* page nor
  test-renders it before merge, and the site stays stale until an unrelated docs change or a
  manual dispatch. The fix is to add `'CONTRIBUTING.md'` to both `paths` lists. Not done here.
- **The site page adds its own text.** `contributing.mdx` puts a legal caution above the
  rendered root. That caution repeats root lines 5–8, so the two agree today, but it is a
  second copy of one paragraph.
- **The branch rule changes for no one in practice.** It records what every merged branch
  already did.

## Found, not acted on

- **The root's other rules are not followed either.** Of the last 100 commits, 0 carry a
  `Signed-off-by:` trailer, and 0 of the last 100 non-merge subjects follow Conventional
  Commits. Root line 109 says an unsigned pull request "will fail the DCO check", but no
  workflow under `.github/workflows/` runs one. A GitHub App could, and it would not show in
  the tree, but 14 unsigned merges suggest nothing is enforcing it. This is the same gap
  between document and practice as the branch rule. I did not resolve it the same way because
  the DCO is a legal commitment (root lines 5–8, ADR PI2-yaa-0003) and not an engineering
  convention. **A person needs to decide it.**
- **The PR template links a heading that no longer exists.** `.github/PULL_REQUEST_TEMPLATE.md:20`
  links `CONTRIBUTING.md#developer-certificate-of-origin-dco`. Commit `684240f` removed that
  heading from the root, so the anchor is dead.
- **The adapter-request form says contributions are closed; the root no longer does.**
  `.github/ISSUE_TEMPLATE/adapter-request.yml` lines 52–53 and 78 say "contributions are not
  currently accepted (legal review pending — see CONTRIBUTING.md)". The root said so at
  `2fb028c`; `cbd6df3` removed it, and the root now describes a full PR process under a
  disclaimer. Under this record the root wins, so the form is stale. **Whether contributions
  are in fact open is a legal and policy fact, though,** and I did not treat the root's
  wording as settling it.
- **`bug.yml`'s *Component* dropdown has no entry for `sdk-go` or `client-go`,** yet this record
  sends their bugs to that form. Root line 30 lists all eight.
- **Root line 161 names a retired planning cycle** in a form that
  `bin/yaagents-public-mirror-verify.sh` does not allowlist. By my reading of the script's
  pattern and allowlist, that is a Check B hit. Check B runs as a blocking step of
  `supply-chain-audit.yml` on every push to `main`. I did not run the script, so this is a
  reading of the script, not an observed failure. It matters more now that the root is the
  only governing document.

## What I read

| Source | Id or path |
|---|---|
| The work packet | `01a08c2f-01ff-7194-8c32-dd73dd3b7f47`, for work item `01a089f7-38ce-7148-86ad-afcf678e82b9`. `prior.artifacts` and `prior.attempts` were both empty. |
| wakalix record | `artifact_read` listed one artifact, `01a089fd-b3a3-7310-b610-8cb401882677` (ADR 0001), which I read in full. Also `catalog_read("architect")` and `registry_read`. |
| Repository topology | `.gitmodules`, and `git submodule status` for the seven pins |
| Contribution documents | Root `CONTRIBUTING.md` at `ebf8a71` (206 lines) and its history: `2fb028c`, `cbd6df3`, `684240f`, `a361ec7` |
| Site | `docs/src/components/ContributingContent.astro`, `docs/src/content/docs/contributing.mdx`, `docs/astro.config.mjs`, and the links in `community.mdx`, `plugins.mdx`, `roadmap.mdx` |
| Front door | `README.md` |
| Templates | `.github/PULL_REQUEST_TEMPLATE.md`, `.github/ISSUE_TEMPLATE/bug.yml`, `.github/ISSUE_TEMPLATE/adapter-request.yml` |
| Workflows and gates | Triggers in every file under `.github/workflows/`, `pages.yml` in detail, Check B and Check C in `supply-chain-audit.yml`, and `bin/yaagents-public-mirror-verify.sh` |
| Branch practice | `git log --merges` (all 14 PR merges), the remote branch list, and the last 100 commits for sign-off and commit style |

## What I did not do

- **I did not read any of the seven submodule `CONTRIBUTING.md` files.** The submodules are
  not checked out in this worktree. Reading their pinned blobs from the main checkout's git
  store, and fetching them from GitHub, both needed approvals this session did not have. Every
  statement above about the `client-ts`, `gateway`, `sdk-go` and stub documents therefore
  comes from the packet's own description. The decision relies on no more than that
  description: the two documents that route issues to their own repositories, the one
  near-copy, and the four stubs that point at the root. **Before the follow-ups replace the
  `gateway` and `sdk-go` documents, someone must read them** for component-specific guidance
  worth moving to their READMEs.
- **I did not read branch or merge history in any submodule.** The branch finding covers the
  meta-repo only.
- **I changed nothing but this file.** The root `CONTRIBUTING.md`, the templates, `pages.yml`
  and every submodule are untouched. The changes in *Consequences* are follow-ups.
- **I did not run `bin/yaagents-public-mirror-verify.sh`.** I checked this file by hand against
  the script's five content patterns and its allowlist: the only retired id here is in the
  allowlisted `ADR PI2-yaa-0003` form, and the superseded licence's name does not appear.
- **I did not commit, push or open a pull request.** Staging and committing needed an approval
  this session did not have, so this record is in the working tree of
  `wakalix/wi-01a089f738ce` and not yet in a commit.
- **Whoever commits this record decides whether to add a DCO `Signed-off-by:` trailer.** A
  sign-off is a legal attestation made on behalf of the committing identity, and it is not
  mine to make.

## What was missing or ambiguous in what I was given

- **`tail.acceptance_criteria` was empty, `tail.input` was null, and `context` was empty.**
  The ask arrived as one paragraph of `tail.intent`, and no requirement set came with it, so
  there were no NFRs to respond to.
- **The packet was more precise than I could check.** It gives the line counts and contents of
  the three kinds of submodule document. I relied on those without verifying them; see *What
  I did not do*.
- **The exit check fixes one path for two questions,** while ADR 0001 sets "one decision per
  file". I treated the two as one decision, for the reason given under *Context*. A reviewer
  who disagrees can split the branch clause into its own record without changing either
  answer.
- **This record depends on a record that has not merged.** ADR 0001 sits on
  `wakalix/wi-01a089f738cb`, not on `main`. If it is rejected, this file's folder and number
  go with it.
- **"Eight repositories" is the packet's count.** I confirmed it as seven submodules plus the
  meta-repo.
- **Whether contributions are open at all** is contradicted within the meta-repo itself (see
  *Found, not acted on*), and nothing I was given settles it.

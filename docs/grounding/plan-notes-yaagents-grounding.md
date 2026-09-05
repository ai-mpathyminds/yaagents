# Plan notes — grounding cycle for YAAgents

**Companion to a `plan-proposal`.** Produced by scrum-master, skill `draft-plan`.
**Packet:** `01a0723f-694d-7c53-b6e5-4bf241e15108` · **Work item:** `01a071f9-c034-732f-8122-29d72e5176ed`
**Product:** `yaagents` · **Repository:** `github.com/ai-mpathyminds/yaagents`
**Branch:** `wakalix/wi-01a071f9c034` (base `main`) · **Date:** 2026-09-05

> **The plan of record is not in this file.** The entries — their ids, agents, skills,
> dependencies and exit checks — travel in the result envelope and are held by wakalix.
> This note deliberately does not restate them, so that there is only one copy and no
> second copy that can drift out of agreement with it. What is here is the reasoning a
> person approving that plan needs and the entries cannot carry: what was read, what this
> plan decided and why, what it left open, and what it did not do.

---

## 1. What I read

| Source | Id / path | What it gave |
|---|---|---|
| `intake-brief` | `01a07195-ddec-736f-9c83-50fd2e8136ae` (`docs/grounding/intake-brief-yaagents.md`, commit `94ce402`) | The whole basis of this plan: the eight-repository scope, the pinned submodule commits, the five open items in its §6, and the blocker in its §7. |
| `catalog_read()` | — | The eight agents this organisation has. |
| `catalog_read(role=…)` | `scrum-master`, `onboarding-agent`, `product-manager`, `architect`, `developer`, `reviewer`, `wakalix-helper` | The skills each agent actually has. Every `agent_role`/`skill_name` pair in the plan was taken from these replies and none was guessed. |
| `registry_read()` | — | The twelve registered artifact types, and the definition of `grounding-manifest` the plan's entries are meant to produce. |
| Packet `prior` | — | Two earlier dispatches of this same ask, both `expired`, and the one artifact above. Nothing below re-derives that artifact; it is read and cited. |
| Workspace | `wakalix/wi-01a071f9c034` at `ebf8a71` | Confirmed directly: `gateway/`, `sdk-go/`, `sdk-fastapi/`, `client-python/`, `client-ts/`, `client-go/`, `cli/` each hold **zero entries** on this branch too, and `docs/grounding/` did not exist here before this note. The intake brief's §7 blocker is still live in the workspace this plan will be run from. |
| Workspace | `.github/workflows/supply-chain-audit.yml` lines 175–217 | Check C-1 verified at source: it greps a fixed licence-name literal across `*.go`, `*.py`, `*.ts`, `*.md`, `*.toml` in all eight repositories and **exits 1 on any hit**. Every document this cycle writes is `*.md` and is in that scan's scope. The literal is not reproduced anywhere in this note, for that reason; it is the superseded source-available licence name that ADR PI2-yaa-0003 replaced with Apache-2.0. |

---

## 2. The four decisions this plan makes

Each one closes something the intake brief left open. Each is stated here so the approver
can reject it as a decision rather than discover it as an assumption.

### 2.1 One product with per-component rows, not eight products

Intake brief §6.1 offered two shapes and did not choose. This plan takes **(a)**: a single
product `yaagents` whose grounding carries per-component rows for eight repositories.

The ground for it is in the dispatch, not in my preference: `tail.product` is the single
value `yaagents` and `delivery.repository` is the single meta repository. §6.1 itself notes
that this is the dispatcher's framing rather than the requester's stated intent. It remains
the only framing anything in the record actually asserts. **If the requester meant eight
separately grounded products, this plan is the wrong shape and should be rejected here, at
the approval gate, rather than amended later.**

### 2.2 All eight are read as sources; seven are also written to

Intake brief §6.2 asked whether the submodules are in scope as *sources* or as *targets*.
This plan treats them as sources — the point is to describe the product — but the mechanics
of §2.3 mean each component's manifest is written **into that component's own repository**,
on a managed branch, as a pull request a person reviews (`delivery.review_mode` is
`pr-human`). That is a real, outward side effect on seven repositories that this cycle was
not explicitly asked to touch, and it is the main thing worth objecting to in this plan.

### 2.3 Dispatch per repository, rather than populating submodules in one workspace

This is the load-bearing choice, and it is the one the intake brief's §7 explicitly deferred
to the plan.

- **Populating the submodules in a single grounding workspace** was not chosen. `.git` here
  is a file, not a directory — this is a `git worktree`, and `git worktree add` does not
  populate submodules. Doing it needs network access to seven GitHub repositories and
  credentials; §9 of the intake brief records that the previous session was not given them,
  and I did not probe for them either (see §4 below). A plan whose first entry may simply be
  unable to run is not a plan.
- **One entry per repository** was chosen. The plan schema carries a `repository` on every
  entry, which is the engine's own way of saying an entry may run somewhere other than the
  dispatching repository. Each component entry therefore reads a repository that its own
  runner checks out, and no entry depends on the empty directories in this worktree.

The cost is §2.2, and a second cost: the eight scans do not share one workspace, so nothing
mechanically guarantees they are cut the same way. §2.4 is how the plan answers that.

### 2.4 The meta scan fixes the dimension set, and every component scan waits for it

Intake brief §6.3 records that no canonical dimension set for a `grounding-manifest` exists
in the record, that `catalog_read` does not enumerate one, and that it did not invent one.
That gap does not go away by being restated, so the plan assigns it: the first entry grounds
the meta repository *and* thereby fixes the dimension set the other seven are cut against.

This is why the seven component entries depend on it. **The dependency is on the dimension
set, not on convenience** — seven manifests each organised by whatever their own repository
volunteered would not compose into a product-level view, and composing them is the last
entry's whole job. Nothing else in the plan is sequenced by preference: the seven component
scans have no dependency on each other and are free to run at once.

### 2.5 The pinned commits are grounded; drift is reported, not followed

Intake brief §6.4 left pins-versus-branch-heads open, and noted the two will diverge
(published artifacts at v0.3.0, v0.4 in progress). This plan grounds **the commit the meta
repository pins at `ebf8a71`**, because that is the state the product actually declares and
it is reproducible, and asks each component entry to **record where that component's default
branch head has moved past its pin** rather than silently grounding the head. Neither answer
is discarded; one becomes the manifest and the other becomes a row in it.

---

## 3. What this plan does not cover

- **It stops at grounding.** No requirement set, no design note, no code. Intake brief §6.5
  records that nothing in the record says why grounding is being done, for whom, or what it
  is meant to enable. Planning work past the manifest would mean inventing that purpose, so
  the plan ends where the evidence does.
- **It plans no remediation.** Nothing found by the eight scans is scheduled to be fixed.
  Grounding reads; what it finds is input to the cycle after this one.
- **It adds no mechanical gate entries.** `wakalix-helper` has `validate-artifact` and
  `scrum-master` has `check-plan-coherence`, and neither appears as an entry. Nothing in the
  intake asked for them, and the record validates artifacts against their registered type on
  the way in regardless.

## 4. What I did not do

- **I did not test whether the submodule remotes are reachable from this runner.** The
  `git ls-remote` probe was not permitted in this session, so §2.3's premise — that
  populating submodules here needs credentials this environment may not have — rests on the
  intake brief's §9 and on the worktree fact I did confirm (all seven directories empty),
  not on a live network check. If the runner *can* clone submodules, the alternative in §2.3
  becomes available and this plan is more conservative than it needed to be.
- **I did not change the workspace beyond this note.** No submodule initialised, no
  configuration touched.
- **I did not write the plan entries into this repository.** The result schema is explicit
  that a plan travels in its envelope and that a copy in git is the copy that disagrees.
- **I did not resolve intake brief §6.5**, and no entry in the plan closes it either.

## 5. What was missing or ambiguous in what I was given

- **`tail.acceptance_criteria` was empty and `tail.input` and `tail.intake` were null.** The
  entire ask is the sentence *"Open this grounding cycle: run draft-plan"*. There is no
  stated test for whether this plan is the right plan, so §2 is my judgement offered for
  rejection, not a checklist I was handed.
- **`context.paths` and `context.excerpts` were both empty**, as they were for the intake.
  The intake brief was reached through `artifact_read`, by the id in `prior.artifacts`.
- **The packet carried no requirement set and no design note**, which are the two inputs
  `draft-plan` is described as sequencing from — `scrum-master.accepts` lists
  `requirement-set`, `design-note`, `review-verdict`, `exit-check-result`, `handoff-note`,
  and an `intake-brief` is none of them. This plan is therefore sequenced from an intake one
  step earlier in the chain than the skill expects. That is the right thing to do for a
  grounding cycle, which exists precisely before requirements are agreeable, but it is worth
  a reader knowing: **no requirement is covered by these entries because no requirement
  exists yet**, and `check-plan-coherence`'s "a requirement no entry covers" test has
  nothing to test against.
- **`grounding-manifest` has no schema I could read** — `registry_read` gives its media type
  and a one-line description, not its fields. The exit checks in the plan can therefore only
  assert that a manifest *was written to a named path*, not that it carries the dimensions
  §2.4 asks for. That is the weakest link in the plan's measurement, stated rather than
  papered over.
- **`scan-workspace` is used for the final composition entry**, which composes eight
  manifests rather than scanning a workspace. It is the only skill in the catalog that
  produces a `grounding-manifest`, so there was no better-fitting line to write. If the
  organisation wants composition to be its own skill, this plan is where that shows.
- **Two earlier dispatches of this ask expired.** Nothing in the record says why, and no
  partial output from either survives to be read, so this run started from the intake brief
  rather than from anything they left.

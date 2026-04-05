# Research: Simplify Bootstrap / Upgrade Path

- Date: 2026-04-04
- Base Hash: db1ffa7
- Participants: Agent (Claude) / Human (Kelvin)

## 1. Problem Statement

`scripts/bootstrap.sh` (645 lines, v1.2.0) has a "decorate" mode intended to add TGS workflow files to existing repos. When run against a repo with an old `tgs/` directory, it:

1. Copies `.tmpl`-suffixed files as-is (never renders them), creating duplicates alongside existing files
2. Misses critical v2 additions: `tgs/adapters/`, `tgs/agentops/prompts/`, `.claude/` setup
3. Writes a `tgs.mk` with the known `sed -E` portability bug (fixed in Makefile at commit 8974d42)
4. Clones the entire tgsflow repo (~14 thought dirs, Go source, tests) just to copy a handful of scaffold files
5. Has no upgrade semantics — `safe_copy` skips existing files, and `--force` overwrites but still with wrong filenames

Meanwhile, `tgs init` (the Go CLI at `src/cmd/init.go`) already does everything decorate tries to do, but correctly: it renders `.tmpl` files, is idempotent, handles Makefile injection, supports vendor decoration, and supports remote template sources.

**Goal:** Replace the 645-line bootstrap.sh with a thin script that installs the `tgs` CLI and runs `tgs init`, and fix gaps in `tgs init` for the upgrade scenario.

## 2. Current State

**bootstrap.sh** (`scripts/bootstrap.sh`):
- Two modes: new-project creation (clone + template overlay) and decorate (add TGS to existing repo)
- Decorate path: `scripts/bootstrap.sh:388-456` — clones repo, walks `src/templates/data/tgs/`, copies files via `safe_copy`
- `safe_copy` (`scripts/bootstrap.sh:68-87`): skips existing, no rename of `.tmpl` suffix
- `write_tgs_mk` (`scripts/bootstrap.sh:105-137`): hardcodes Makefile snippet with stale `sed -E` pattern
- Template selection menu (`scripts/bootstrap.sh:204-335`): 130 lines of interactive menu for react/python/go/cli/none

**tgs init** (`src/cmd/init.go`):
- Renders `.tmpl` → strips suffix, substitutes Go template variables (`src/templates/render.go:60-88`)
- Ensures Makefile has `new-thought` target (`src/cmd/init.go:448-494`)
- Vendor decoration: `tgs init claude` / `tgs init gemini` (`src/cmd/init.go:496-512`)
- Remote templates: archive URL, git repo, local dir (`src/cmd/init.go:67-115`)
- **Gaps found:**
  - No `--force` flag — cannot update existing files during upgrade
  - `ensureAdaptersExecutable()` is commented out (`src/cmd/init.go:158-161`)
  - `makefileNewThoughtBlock()` (`src/cmd/init.go:473-493`) has same `sed -E` bug
  - Does not set up `.claude/rules/` or `.claude/hooks.json`

**install.sh** (`scripts/install.sh`): Already exists — downloads the `tgs` binary from GitHub releases. 30 lines, works well.

## 3. Constraints & Assumptions

- **Platform:** macOS and Linux (NFR-002). `sed -E` is macOS-specific; POSIX `tr+sed` required.
- **No breaking changes:** Existing users who call `bootstrap.sh` via `curl | bash` must not break. The new script should accept the same flags and degrade gracefully.
- **Idempotency:** SR-017 requires preserving existing files by default. Upgrade requires an explicit opt-in (`--force` or `--upgrade`).
- **No internet required for tgs init:** Embedded templates mean `tgs init` works offline once CLI is installed. Bootstrap still needs network for the install step.
- **install.sh dependency:** `scripts/install.sh` requires a published GitHub release. During development, local builds must still work.

## 4. Risks & Impact

- **Upgrade data loss:** A `--force` flag on `tgs init` could overwrite user-customized design docs or AGENTOPS.md. Mitigation: only force-update known-stale files, or prompt/warn.
- **Bootstrap URL breakage:** If anyone has `curl ... | bash` pointing to the old bootstrap.sh, changing its behavior could confuse them. Mitigation: keep the same entry point, just simplify internals.
- **Template divergence:** The `.tmpl` files in `src/templates/data/tgs/` are slightly stale compared to the actual `tgs/` files in this repo (e.g., research.md template says "5-10 lines max" but the `.tmpl` version doesn't). Need to reconcile.
- **Missing `.claude/` in init:** The v2 workflow depends on `.claude/rules/`, CLAUDE.md, and skills — none of which `tgs init` currently handles. This is scope creep but important for the upgrade story.

## 5. Alternatives Considered

| Option | Pros | Cons | Effort |
|--------|------|------|--------|
| **A: Thin bootstrap -> install.sh + tgs init** | Eliminates 600+ lines of shell duplication; single source of truth for scaffolding; `tgs init` already handles rendering, idempotency, remote templates | Need to fix `tgs init` gaps (--force, adapters, sed bug, .claude/); new-project template overlay feature lost | Medium |
| **B: Fix bootstrap.sh in-place** | Minimal change; preserves new-project creation flow | Maintains two parallel implementations of the same logic; `.tmpl` rendering in bash is fragile; every fix needs to happen in two places | Medium |
| **C: Remove bootstrap.sh entirely, document `tgs init`** | Simplest; zero maintenance; single path | Loses the `curl pipe bash` onboarding story; users need to install tgs CLI first (two steps instead of one) | Low |
| **D: Bootstrap installs CLI + runs init, keep new-project mode** | Best of both worlds; thin shell + full CLI power; preserves `curl pipe bash` UX; new-project creation via `tgs init --scaffold react` | Largest scope; need to add template overlay to `tgs init` | High |

## 6. Perspectives Considered

- [x] **End-user impact** — A user upgrading an old repo currently gets a broken result with no error. The fix makes upgrade actually work. New users get a simpler onboarding (install + init vs. a 645-line script).
- [x] **Security adversarial view** — bootstrap.sh runs `git clone` from a hardcoded URL and copies files into the working directory. The thin version reduces attack surface by delegating to a signed binary. The `--force` flag needs to be explicit to prevent accidental overwrites.
- [x] **Privacy impact** — No personal data involved. Network calls only for installing the CLI binary.
- [x] **Operator/admin impact** — Teams using the bootstrap in CI/automation get a simpler, more predictable script. The `tgs init` path has proper exit codes and structured logging.
- [x] **Future-self debt** — Maintaining two parallel scaffold implementations (bash + Go) is the current debt. Converging on one (Go) eliminates it permanently.

## 7. Ethics Quick-Check

- [x] Does this collect or process personal data? **No.**
- [x] Does this automate a decision that affects people? **No.**
- [x] Does this reduce human oversight of a critical path? **No.** The bootstrap is a developer tool, not a production system.
- [x] Could this be used to discriminate, exclude, or harm? **No.**

## 8. Recommendation

**Option A: Thin bootstrap that calls install.sh + tgs init**, with targeted fixes to `tgs init`:

1. **Rewrite `bootstrap.sh`** to ~80 lines: detect platform, install `tgs` CLI via `install.sh`, run `tgs init` with appropriate flags. Keep `--decorate`, `--dry-run`, `--force` for backward compatibility.
2. **Add `--force` flag to `tgs init`** so it can overwrite existing files during upgrades.
3. **Fix `sed -E` bug** in `makefileNewThoughtBlock()` (`src/cmd/init.go:481`).
4. **Uncomment `ensureAdaptersExecutable()`** or implement adapter copy in the template tree.
5. **Reconcile stale `.tmpl` files** with current `tgs/` content.
6. **Drop new-project template overlay** from bootstrap (react/python/go/cli templates stay in `templates/` for reference but aren't auto-applied by bootstrap). Users who want these can use `tgs init --templates`.

This eliminates the duplication, fixes the upgrade path, and keeps the `curl | bash` onboarding story. The new-project overlay is a niche feature that's better served by `tgs init --templates` than 130 lines of interactive bash menu.

**Why Option A over D:** Option D adds template overlay to the Go CLI, which is scope creep for this thought. The templates are already in the repo for reference. We can add `tgs init --scaffold <type>` in a future thought if demand exists.

## 9. References & Links

- Current bootstrap: `scripts/bootstrap.sh`
- CLI installer: `scripts/install.sh`
- Go init command: `src/cmd/init.go`
- Template renderer: `src/templates/render.go`
- Template data: `src/templates/data/tgs/`
- sed -E fix commit: 8974d42
- Needs: N-005, N-015, N-016, N-017, N-018, N-024, N-025, N-026
- Requirements: SR-005, SR-015, SR-016, SR-017, SR-028, SR-029, SR-030

---
Approval checkpoint: Please review this research and reply one of:
- APPROVE research
- REQUEST CHANGES: <notes>

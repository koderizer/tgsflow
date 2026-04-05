# Plan: Simplify Bootstrap / Upgrade Path

See [research.md](./research.md) — approved 2026-04-05.

> **Revision (2026-04-05):** Pivoted away from Go `text/template` rendering entirely.
> The scaffold is now a tree of **plain files** copied as-is. No `.tmpl` suffixes, no
> variable substitution, no `text/template` dependency. Source of truth is a dedicated
> `scaffold/` directory in the repo that mirrors the exact layout the target repo
> should end up with (`scaffold/.claude/`, `scaffold/tgs/`, `scaffold/CLAUDE.md`, etc.).

## 1. Objectives

- Eliminate Go templating from the scaffolding path. Scaffold files are plain `.md`, `.sh`, `.json`, `.yml` — no `.tmpl`, no `{{.Var}}` placeholders.
- Reduce `scripts/bootstrap.sh` to ~60 lines: fetch scaffold tarball → extract into target repo → done.
- Ship the "modern" `.claude/` setup (rules, commands, hooks) as part of the scaffold so fresh repos get slash commands (`/tgs-research`, `/tgs-plan`, `/tgs-close`) and path-scoped rules out of the box.
- Make upgrade ergonomic: `--force` overwrites TGS-managed scaffold files but never touches user content (`tgs/thoughts/`, user-edited `tgs.yml`).
- Simplify `tgs init` (Go CLI) to a plain embedded-file walker. Delete `src/templates/render.go` template rendering logic.
- Keep a single source of truth: `scaffold/` at repo root. `src/templates/data/tgs/` is deleted.

## 2. Scope / Non-goals

**In-scope:**
- Create `scaffold/` directory at repo root with the full "modern" layout (see §5).
- Delete `src/templates/data/tgs/` and all `.tmpl` files.
- Rewrite `scripts/bootstrap.sh` as a thin fetch+extract script (~60 lines).
- Simplify `src/cmd/init.go` and `src/templates/render.go` — remove `text/template` imports, just walk embedded `scaffold/` and copy files (no rendering).
- Add `--force` behavior so upgrades work.
- Add CI release asset: publish `scaffold.tar.gz` on each release (or fetch directly from git tag).
- Update tests.

**Out-of-scope:**
- Any per-project variable substitution. Users edit `tgs.yml` after install if they want a different project name.
- Keeping react/python/go/cli template overlays in bootstrap — removed (templates stay in `templates/` for reference only).
- Remote custom template sources (`--templates URL`, `--templates-ref`, `--templates-subdir`) — removed. Orgs that want custom scaffolds can fork the repo.
- Changes to `scripts/install.sh` or release plumbing.
- Changes to `tgs/design/*` or the workflow rules themselves.

**Explicit deletions:**
- `src/templates/data/` (entire tree) → replaced by root-level `scaffold/`.
- `text/template` usage in `src/templates/render.go`.
- `downloadZip` / `downloadTarGz` / `cloneGitRepo` / `findTemplatesRoot` / `dirHasTgsTemplates` / `unzipFile` in `src/cmd/init.go`.
- `--templates`, `--templates-ref`, `--templates-subdir` flags on `tgs init`.
- Interactive template menu and new-project creation flow in `bootstrap.sh`.

## 3. Acceptance Criteria

1. `ls scaffold/` contains: `.claude/`, `tgs/`, `CLAUDE.md`, `Makefile.tgs.mk`, `tgs.yml` (generic defaults). No `.tmpl` files anywhere in the repo.
2. `grep -r "text/template" src/` returns nothing.
3. `grep -r "\.tmpl" src/ scripts/` returns nothing.
4. `scripts/bootstrap.sh` is ≤80 lines (target: ~60 excluding comments).
5. Running `./scripts/bootstrap.sh` in an empty directory:
   - Fetches the scaffold (tarball or git archive).
   - Extracts `.claude/`, `tgs/`, `CLAUDE.md`, `tgs.yml`, Makefile snippet into the current directory.
   - Is idempotent: re-running without `--force` preserves existing files.
6. Running `./scripts/bootstrap.sh --force` overwrites TGS-managed files but leaves `tgs/thoughts/` and any user-added files untouched.
7. After bootstrap, `.claude/commands/tgs-research.md`, `.claude/commands/tgs-plan.md`, `.claude/commands/tgs-close.md`, `.claude/rules/README.md`, `.claude/hooks.json`, and `CLAUDE.md` all exist.
8. `tgs init` (Go CLI) produces the same result as `bootstrap.sh` when run offline against an empty dir (uses `embed.FS` against `scaffold/`).
9. `make new-thought title="x"` works after bootstrap on macOS and Linux (POSIX `tr+sed`, no `sed -E`).
10. `go test ./...` passes. Tests for templated rendering are deleted; new tests cover the plain-copy path and `--force` behavior.
11. On an existing repo with a stale old `tgs/` dir, `./scripts/bootstrap.sh --force` produces a clean upgrade: old files replaced, `tgs/thoughts/` preserved.

## 4. Phases & Tasks

- **Phase 1: Build the scaffold tree**
  - [ ] Create `scaffold/` at repo root.
  - [ ] Copy current `.claude/rules/`, `.claude/commands/`, `.claude/hooks.json` into `scaffold/.claude/`.
  - [ ] Copy current `CLAUDE.md` into `scaffold/CLAUDE.md` (review for any tgsflow-specific bits; keep it generic).
  - [ ] Copy current authoritative `tgs/agentops/`, `tgs/design/`, `tgs/adapters/`, `tgs/context/` into `scaffold/tgs/`. Strip `tgs/thoughts/` (project-specific).
  - [ ] Create `scaffold/tgs/tgs.yml` from current `tgs/tgs.yml` with `project: my-project` as default.
  - [ ] Create `scaffold/Makefile.tgs.mk` (the `new-thought` target, POSIX-compliant — no `sed -E`).
  - [ ] Verify adapter scripts under `scaffold/tgs/adapters/` retain executable bit (use `git update-index --chmod=+x` if needed).

- **Phase 2: Rewrite bootstrap.sh**
  - [ ] New implementation: parse flags, detect `tar`/`curl`, fetch scaffold tarball from latest release OR via `git archive` from main, extract into CWD.
  - [ ] Implement `--force` (pass through to `cp` / `tar --overwrite`).
  - [ ] Implement `--dry-run` (list what would be copied).
  - [ ] Preserve `--help`. Remove `--decorate` (no longer needed — decoration IS the default), `--with-templates` (deleted).
  - [ ] After extract, append/create `Makefile` with `include Makefile.tgs.mk` (or equivalent), only if not already present.
  - [ ] Ensure `tgs/thoughts/` directory exists (empty) so `make new-thought` works immediately.

- **Phase 3: Simplify tgs init (Go CLI)**
  - [ ] Rewrite `src/cmd/init.go`: remove `downloadZip`, `downloadTarGz`, `cloneGitRepo`, `findTemplatesRoot`, `unzipFile`, `isArchiveURL`, `isLikelyGitURL`, `execCommand`, `--templates*` flags.
  - [ ] Replace `RenderTGSTree` with a plain `CopyScaffoldTree` that walks `embed.FS` for `scaffold/` and copies files verbatim.
  - [ ] Rewrite `src/templates/render.go`: delete `text/template` imports; keep only a minimal `CopyScaffoldTreeFromFS(source fs.FS, destRoot string, force bool) error` helper.
  - [ ] Add `--force` flag to `tgs init`.
  - [ ] Update `//go:embed` directive to point at the new `scaffold/` path (may need to move `scaffold/` under `src/` or use `//go:embed all:../../scaffold` — research in Phase 3).
  - [ ] Fix `makefileNewThoughtBlock()` or delete it entirely (since the Makefile snippet is now a real file in `scaffold/`).
  - [ ] Keep `decorateVendorReadme` for `tgs init claude` / `tgs init gemini` — still useful.

- **Phase 4: Tests and cleanup**
  - [ ] Delete `src/templates/render_test.go` tests for `.tmpl` rendering.
  - [ ] Add `TestCopyScaffold_IdempotentByDefault` — pre-create a file, run copy, assert unchanged.
  - [ ] Add `TestCopyScaffold_ForceOverwrites` — pre-create stale file, run with force, assert replaced.
  - [ ] Add `TestCopyScaffold_IncludesClaudeDir` — assert `.claude/rules/`, `.claude/commands/`, `.claude/hooks.json`, `CLAUDE.md` all present.
  - [ ] Integration test in `scripts/test-bootstrap.sh` (or new file): create temp dir with stale fake `tgs/`, run bootstrap `--force`, assert clean state.
  - [ ] Update `Makefile` targets (`bootstrap`, `test-bootstrap`) to match new flow.
  - [ ] Update `README.md` install/decorate section.
  - [ ] Delete obsolete files: `src/templates/data/tgs/`, any `.tmpl` files elsewhere.

## 5. File/Module Changes

**Add (new `scaffold/` tree at repo root):**
- `scaffold/CLAUDE.md` — generic version of current `CLAUDE.md`
- `scaffold/Makefile.tgs.mk` — POSIX-compliant `new-thought` target
- `scaffold/tgs.yml` — generic config with `project: my-project`
- `scaffold/.claude/hooks.json`
- `scaffold/.claude/rules/README.md`
- `scaffold/.claude/rules/example-data-privacy.md`
- `scaffold/.claude/rules/example-conventions.md`
- `scaffold/.claude/commands/tgs-research.md`
- `scaffold/.claude/commands/tgs-plan.md`
- `scaffold/.claude/commands/tgs-close.md`
- `scaffold/tgs/README.md`
- `scaffold/tgs/agentops/AGENTOPS.md`
- `scaffold/tgs/agentops/tgs/research.md`
- `scaffold/tgs/agentops/tgs/plan.md`
- `scaffold/tgs/agentops/tgs/implementation.md`
- `scaffold/tgs/agentops/prompts/*.md` (5 files)
- `scaffold/tgs/adapters/claude-code.sh` (+x)
- `scaffold/tgs/adapters/gemini-code.sh` (+x)
- `scaffold/tgs/design/*.md` (6 files)
- `scaffold/tgs/context/*.md` (2 files)
- `scaffold/tgs/thoughts/.gitkeep`

**Modify:**
- `scripts/bootstrap.sh` — full rewrite to ~60 lines.
- `src/cmd/init.go` — drop templating paths; add `--force`; update embed directive.
- `src/templates/render.go` — drastically simplified or renamed to `src/templates/copy.go`. Remove `text/template`.
- `src/templates/render_test.go` → `src/templates/copy_test.go`.
- `Makefile` — update `bootstrap` / `test-bootstrap` targets.
- `README.md` — update install section.

**Delete:**
- `src/templates/data/tgs/` (entire subtree — all `.tmpl` files)
- `src/templates/data/ci/` and `src/templates/data/thought/` if they contain `.tmpl` files (inspect first; may be safe to keep as plain files if already plain)

**Do NOT touch:**
- `scripts/install.sh` (CLI installer — orthogonal)
- `tgs/thoughts/` (project-specific)
- `src/core/ears/`, `src/core/config/`, etc. — unrelated to init/scaffolding
- `templates/react/`, `templates/python/`, `templates/go/`, `templates/cli/` — separate feature, out of scope (but no longer referenced by bootstrap)

## 6. Test Plan

**Unit tests (Go):**
- `TestCopyScaffoldTree_Default` — copies all files, skips existing.
- `TestCopyScaffoldTree_Force` — overwrites existing files when `force=true`.
- `TestCopyScaffoldTree_PreservesThoughts` — ensures `tgs/thoughts/` user content is never overwritten even with `--force`.
- `TestCmdInit_Force` — CLI-level `tgs init --force` exercise.
- `TestCmdInit_NoForce_Idempotent` — second run is a no-op.
- `TestScaffoldIncludesClaudeDir` — asserts `.claude/rules/README.md` and `.claude/commands/tgs-research.md` are present in the embedded FS.

**Integration (shell):**
- `scripts/test-bootstrap.sh` — new or updated:
  - Create `/tmp/test-$$`, run `bootstrap.sh`, assert files exist.
  - Run `bootstrap.sh` again (idempotency), assert no changes.
  - Add a stale file, run `bootstrap.sh --force`, assert replaced.
  - Add a user file under `tgs/thoughts/foo/bar.md`, run `bootstrap.sh --force`, assert preserved.
- Manual check on macOS and Linux: `make new-thought title="smoke"` works.

**Regression:**
- `go test ./...` green after removing template tests.
- `make build` succeeds with new `//go:embed` directive.
- `go vet ./...` clean.

## 7. Rollout & Rollback

**Rollout:**
1. Land the PR on `main`. All scaffold files committed in one PR so reviewers see the full picture.
2. Cut a new release via goreleaser. Ensure release assets include a `scaffold.tar.gz` (or ensure bootstrap.sh can fetch via `git archive` from the tag).
3. Verify `curl -sSL https://raw.githubusercontent.com/akelv/tgsflow/main/scripts/bootstrap.sh | bash` works on a clean temp dir.
4. Update any docs that reference the old flags (`--decorate`, `--with-templates`).

**Rollback:**
- `git revert` the PR. Previous bootstrap.sh and `src/templates/data/tgs/` tree are restored.
- Old releases remain pinnable via `install.sh TAG=vX.Y.Z`.

## 8. Estimates & Risks

**Risks:**
- **Embed path:** Go's `//go:embed` cannot reference parent directories. `scaffold/` at repo root is not reachable from `src/templates/render.go`. Options: (a) put `scaffold/` under `src/cmd/scaffold/`, (b) put it under `src/templates/scaffold/`, (c) keep it at repo root and have a build-time `go generate` copy step. Recommendation: **place scaffold at `src/templates/scaffold/`** (or rename `src/templates/data/` → `src/templates/scaffold/`) for zero friction. Update Phase 1 to use this path.
- **Executable bit on adapter scripts:** `embed.FS` does not preserve mode. Solution: after copying `*.sh` files, Go code chmods them to `0755`. Bootstrap.sh uses `tar xzf` which does preserve mode.
- **Scaffold drift:** If someone edits the live `tgs/` in tgsflow but forgets to update `scaffold/tgs/`, the scaffold becomes stale. Mitigation: add a `make sync-scaffold` target (not this thought) or a CI check that diffs key files. For now, document the maintenance requirement in `scaffold/README.md`.
- **Bootstrap fetch source:** Fetching `scaffold.tar.gz` from GitHub releases adds a dependency on release tagging. Alternative: `git archive --remote=...` — works without releases but requires git 2.10+. Recommendation: prefer `curl` + `github.com/akelv/tgsflow/archive/main.tar.gz` (always works, no release needed), then extract only the `scaffold/` subdirectory.
- **`--force` blast radius:** A user running `--force` on a customized repo could lose edits. Mitigation: explicit warning in output; only copies files that exist in `scaffold/` (never touches files outside the scaffold tree); hard-skip anything under `tgs/thoughts/`.

**Unknowns to resolve in Phase 1:**
- Confirm the exact path for `scaffold/` (repo root vs under `src/`). Decision gate before Phase 3.
- Confirm whether `.claude/commands/tgs-*.md` in the scaffold need any content changes to be generic (vs tgsflow-specific references).
- Confirm if `scripts/install.sh` is still needed at all. If bootstrap.sh does everything, install.sh remains only for users who want the `tgs` CLI separately. Decision: **keep install.sh as-is**, it's orthogonal.

---
Approval checkpoint: Please review this plan and reply one of:
- APPROVE plan
- REQUEST CHANGES: <notes>

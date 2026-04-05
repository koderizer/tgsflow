# Implementation Summary: Simplify Bootstrap / Upgrade Path

- Base Hash: db1ffa7
- Date: 2026-04-05
- Research: [research.md](./research.md) — approved 2026-04-05
- Plan: [plan.md](./plan.md) — approved 2026-04-05 (v2, simplified)

## 1. Overview (What & Why)

Eliminated Go `text/template` rendering from the scaffolding pipeline and collapsed `bootstrap.sh` from **645 lines to 103 lines**. The scaffold is now a tree of **plain files** embedded in the binary and copied verbatim to target repos. New users get a working TGS setup — including the modern `.claude/` slash commands (`/tgs-research`, `/tgs-plan`, `/tgs-close`), rules, hooks, and `CLAUDE.md` — with one command.

Upgrades finally work: `bootstrap.sh --force` (or `tgs init --force`) overwrites TGS-managed scaffold files while guaranteeing `tgs/thoughts/` (user work) is never touched.

## 2. File Changes

**Added — `src/templates/scaffold/` (31 files):**
- `CLAUDE.md`, `Makefile.tgs.mk`
- `.claude/hooks.json`
- `.claude/rules/{README,example-data-privacy,example-conventions}.md`
- `.claude/commands/{tgs-research,tgs-plan,tgs-close}.md`
- `tgs/{README.md,tgs.yml}` (generic `project: my-project` default)
- `tgs/adapters/{claude-code,gemini-code}.sh` (executable bit preserved)
- `tgs/agentops/AGENTOPS.md`
- `tgs/agentops/tgs/{research,plan,implementation}.md`
- `tgs/agentops/prompts/{context_brief,context_search,impl,plan,review}.md`
- `tgs/context/{brief_template,search_prompt}.md`
- `tgs/design/{00_context,10_needs,20_requirements,30_architecture,40_vnv,50_decisions}.md`
- `tgs/thoughts/.gitkeep`

**Added — Go code:**
- `src/templates/copy.go` — single file, ~60 lines. Replaces `render.go`/`write.go`. One exported function `CopyScaffold(destRoot, force)` that walks `embed.FS` and copies files. `//go:embed all:scaffold` (the `all:` prefix includes dotfiles).

**Rewritten:**
- `scripts/bootstrap.sh` — 645 → **103 lines**. Fetches `codeload.github.com/${REPO}/tar.gz/refs/heads/${BRANCH}` via `curl`, extracts to `mktemp`, finds `src/templates/scaffold/` inside, copies files into CWD. Honors `--force` (overwrite) and `--dry-run`. Skips `tgs/thoughts/*` always. Ensures root `Makefile` includes `Makefile.tgs.mk`.
- `src/cmd/init.go` — 513 → **93 lines**. Removed: archive/git/subdir template sourcing, `--templates*` flags, `--ci-template`, `--interactive`, `--decorate` flag, `claude|gemini` vendor subcommand, `sed -E` bug, all the zip/tar/git-clone helpers. Kept: `CmdInit`, cobra constructor, Makefile include helper. Added `--force` flag.
- `src/cmd/init_test.go` — new tests: `TestInitSeedsFiles` (verifies 22+ files incl. `.claude/`, adapter `+x`), `TestInit_IdempotentByDefault`, `TestInit_ForceOverwrites`, `TestInit_ForcePreservesThoughts`.

**Deleted:**
- `src/templates/render.go` (126 lines of `text/template` rendering)
- `src/templates/render_test.go`
- `src/templates/write.go` (`WriteIfMissing` was unreferenced)
- `src/templates/data/` — entire subtree (`tgs/*.tmpl`, `ci/*.tmpl`, `thought/*.tmpl`)
- `src/core/config/factory.go` (180 lines — `TemplateData`, `RenderConfigYAML`, `EnsureConfigFile`, `PromptInteractive`)
- `src/core/config/factory_test.go`

**Modified:**
- `README.md` — Quick Start updated to describe the new scaffold behavior (mentions `.claude/` slash commands, `--force` upgrade path).
- `Makefile` — removed stale `bootstrap`, `test-bootstrap`, `clean-templates` targets; added `ears-gen` to help.

## 3. Commands & Migrations

No runtime migrations. Fresh install:

```bash
curl -sSL https://raw.githubusercontent.com/akelv/tgsflow/main/scripts/bootstrap.sh | bash
```

Upgrade an existing repo that has a stale `tgs/` from an older version:

```bash
curl -sSL https://raw.githubusercontent.com/akelv/tgsflow/main/scripts/bootstrap.sh | bash -s -- --force
```

Or, once `tgs` CLI is installed, equivalent commands:

```bash
tgs init              # idempotent, preserves existing files
tgs init --force      # overwrite scaffold files (tgs/thoughts/ preserved)
```

## 4. How to Test

**Build & unit tests:**
```bash
make build
go test ./...
```
Expected: `ok  src/cmd`, `ok  src/core/brain`, `ok  src/core/config`, `ok  src/core/ears`. (Pre-existing `go vet` warnings in ANTLR generated code and `core/brain/brain.go` struct tags are unrelated.)

**Go CLI end-to-end:**
```bash
T=$(mktemp -d) && cd "$T"
~/github/tgsflow/bin/tgs init
find . -type f | wc -l         # 32 files (31 scaffold + Makefile)
test -x tgs/adapters/claude-code.sh && echo "executable OK"
~/github/tgsflow/bin/tgs init  # re-run: idempotent, no changes
echo STALE > CLAUDE.md
~/github/tgsflow/bin/tgs init --force
head -1 CLAUDE.md              # should show the scaffold content, not "STALE"
```

**bootstrap.sh end-to-end** (tested locally via `file://` URL during implementation):
- Fresh run: 31 files written, Makefile created with `include Makefile.tgs.mk`.
- Idempotent re-run: 30 skips (all present).
- `--force`: 30 files re-written.
- User file at `tgs/thoughts/myhash-feature/notes.md` preserved through `--force` run.
- `tgs/adapters/*.sh` retains executable bit after extraction via `cp -p`.

**Unit test coverage for force/preservation:**
```bash
go test ./src/cmd/... -run TestInit -v
```

## 5. Integration Steps

1. Review and merge the PR on `main`.
2. The `codeload.github.com/akelv/tgsflow/tar.gz/refs/heads/main` URL becomes authoritative once merged.
3. No release tagging is required for bootstrap to work — it pulls from the branch archive. A new `tgs` CLI release is still useful for users who want the offline-capable binary; `scripts/install.sh` is unchanged.

## 6. Rollback

Single `git revert` of the PR restores the prior bootstrap.sh, templating stack, and config factory. No runtime state or external services to reconcile.

## 7. Follow-ups & Next Steps

- **Scaffold sync automation:** Add a CI check or `make sync-scaffold` target that diffs `src/templates/scaffold/` against the authoritative live copies under `tgs/`, `.claude/`, `CLAUDE.md`. Not urgent — the plan explicitly deferred this.
- **Pre-existing `go vet` warnings:** `core/brain/brain.go:17,20,23` has malformed struct tags (`json:"role","content"`) and the generated ANTLR parser has unreachable code. Neither introduced by this PR, but worth a follow-up thought.
- **Release:** Cut a `tgs` CLI release so `install.sh` + `tgs init --force` become a valid upgrade path for users who prefer the binary over the shell script.
- **Onboarding doc:** Consider a short "What's new in v2" section in README once a release is tagged.

## 8. Links

- Research: `tgs/thoughts/db1ffa7-simplify-bootstrap-upgrade-path/research.md`
- Plan: `tgs/thoughts/db1ffa7-simplify-bootstrap-upgrade-path/plan.md`
- Key files:
  - `scripts/bootstrap.sh`
  - `src/templates/copy.go`
  - `src/templates/scaffold/` (31 files)
  - `src/cmd/init.go`
  - `src/cmd/init_test.go`

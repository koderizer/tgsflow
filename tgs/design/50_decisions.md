# Architecture Decisions & Change Log

## ADR-0001: Adopt Cobra/Viper for CLI and configuration

- **Context:** The CLI needed a structured command tree, consistent help UX, and environment/config support while remaining extensible.
- **Decision:** Refactor to Cobra for command parsing and Viper for optional config/env loading. Keep root command terse with subcommands for `init`, `context`, `verify`, and `agent`.
- **Status:** accepted
- **Consequences:**
  - Pros: Better UX, composability, standard flags, easier subcommand growth.
  - Cons: Additional dependencies; need to keep Viper optional to avoid brittle config failures.
- **Reference:** `tgs/b48976e-refactor-cli-to-cobra-viper/`

## ADR-0002: Integrate EARS grammar and linter via ANTLR4

- **Context:** Requirements must be unambiguous and lintable against EARS patterns.
- **Decision:** Use an ANTLR4 grammar (`src/core/ears/ears.g4`) to generate a Go parser and provide lint helpers integrated with `tgs verify` when enabled in `tgs.yaml`.
- **Status:** accepted
- **Consequences:**
  - Pros: Formal parsing with clear shapes; deterministic linting.
  - Cons: Java/ANTLR toolchain required for regeneration; CI needs caching.
- **Reference:** `tgs/f71f872-ears-linter-core-and-verify-integration/`

## ADR-0003: Thought directories for approval-gated workflow

- **Context:** Changes must be traceable from intent to implementation with human approvals.
- **Decision:** Use `tgs/<BASE_HASH>-<kebab-title>/` per thought, containing `README.md`, `research.md`, `plan.md`, and `implementation.md`. Implement only after approvals.
- **Status:** accepted
- **Consequences:**
  - Pros: Strong traceability; clearer handoff points for AI agents; audit-friendly.
  - Cons: Additional documentation burden; requires team discipline.
- **Reference:** `tgs/b4552ea-standardize-agentops-intake-to-pr-and-enrich-new-thought/`

## ADR-0004: Release automation with GoReleaser and Homebrew

- **Context:** Distribute `tgs` binaries reliably across macOS/Linux and provide easy installation.
- **Decision:** Use GoReleaser to build and publish releases, with a Homebrew tap and a curl installer script.
- **Status:** accepted
- **Consequences:**
  - Pros: Simple upgrades; reproducible builds; wider adoption.
  - Cons: Requires release hygiene and CI maintenance.
- **Reference:** `tgs/43ec077-automate-releases-with-goreleaser-and-homebrew/`

## ADR-0005: Repository decoration via `tgs init`

- **Context:** Teams need a fast way to introduce TGS structure and CI templates to existing repos.
- **Decision:** Provide `tgs init --decorate --ci-template {github|gitlab|none}` to idempotently create `tgs/` skeleton and optional CI workflows.
- **Status:** accepted
- **Consequences:**
  - Pros: Low-friction onboarding; consistent structure; idempotent.
  - Cons: Template drift risk; requires periodic updates.
- **Reference:** `tgs/612a57f-decorate-existing-software-project-repository/`

## ADR-0006: Agent integration via adapter process

- **Context:** Need a flexible way to integrate with external AI agents/tools without coupling CLI to vendor APIs.
- **Decision:** Implement `tgs agent exec` that shells out to an adapter (`tgs/adapters/claude-code.sh`) with prompt/context via flags and env, returning patches or text.
- **Status:** accepted
- **Consequences:**
  - Pros: Vendor-agnostic, replaceable adapters, controlled IO and timeouts.
  - Cons: Shell process management; error handling relies on adapter conventions.
- **Reference:** `tgs/f0d3f9a-add-agent-parent-command/`

## ADR-0007: Plain-file scaffold, no templating (supersedes part of ADR-0005)

- **Context:** The original decorate path rendered Go `text/template` files (`.tmpl`) with variable substitution, supported multiple template sources (local dir, archive URL, git repo, subdirectory), and reimplemented similar logic in `bootstrap.sh`. This produced a 645-line bootstrap script, a 513-line `init.go`, a 180-line config factory just to substitute a few variables in `tgs.yml`, and a broken upgrade path for existing repos (`.tmpl` files were copied raw into target dirs, adapters were missing, `sed -E` was non-portable).
- **Decision:** Ship the scaffold as a tree of **plain files** at `src/templates/scaffold/`, embedded into the binary via `//go:embed all:scaffold`. `bootstrap.sh` downloads `codeload.github.com/<repo>/tar.gz/refs/heads/<branch>`, extracts it, finds the scaffold directory, and copies files verbatim into the current directory. `tgs init` does the same offline from the embedded FS. No `text/template`, no `.tmpl` suffixes, no variable substitution. `tgs.yml` ships with generic defaults; users edit the `project:` field after install. Files under `tgs/thoughts/` are never overwritten — even with `--force`.
- **Status:** accepted
- **Consequences:**
  - Pros: `bootstrap.sh` collapses from 645 → 103 lines; `init.go` from 513 → 93 lines; templating stack (`render.go`, `write.go`, `factory.go`) deleted entirely. Upgrade path (`--force`) finally works. Scaffold includes the modern `.claude/` setup (rules, slash commands, hooks, `CLAUDE.md`) out of the box. Single source of truth for scaffolding.
  - Cons: Per-project variable customization (project name, AI provider, etc.) is no longer automatic — users edit `tgs/tgs.yml` post-install. Scaffold drift requires manual sync between the live `tgs/`/`.claude/` and `src/templates/scaffold/` (follow-up to automate).
- **Reference:** `tgs/thoughts/db1ffa7-simplify-bootstrap-upgrade-path/`

---

## Change Log
| Date | Change | Impact | Owner |
|------|--------|--------|-------|
| 2025-09-17 | ADR-0002 Integrate EARS grammar and linter | Formal requirements linting enabled | tgs/f71f872-ears-linter-core-and-verify-integration |
| 2025-09-17 | ADR-0004 Release automation with GoReleaser/Homebrew | Reliable distribution channels established | tgs/43ec077-automate-releases-with-goreleaser-and-homebrew |
| 2025-09-14 | ADR-0001 Adopt Cobra/Viper for CLI/config | CLI extensibility and UX improved | tgs/b48976e-refactor-cli-to-cobra-viper |
| 2025-09-14 | ADR-0003 Thought directories for approval-gated workflow | End-to-end traceability instituted | tgs/b4552ea-standardize-agentops-intake-to-pr-and-enrich-new-thought |
| 2025-09-14 | ADR-0005 Repository decoration via tgs init | One-command onboarding for repos | tgs/612a57f-decorate-existing-software-project-repository |
| 2025-09-14 | ADR-0006 Agent integration via adapter process | Vendor-agnostic AI agent execution | tgs/f0d3f9a-add-agent-parent-command |
| 2026-04-05 | ADR-0007 Plain-file scaffold, no templating | bootstrap.sh 645→103 lines; modern `.claude/` setup ships with scaffold; working upgrade path | tgs/thoughts/db1ffa7-simplify-bootstrap-upgrade-path |

---

### Checklist
- [x] Every major trade-off documented as an ADR  
- [x] Consequences captured (pros/cons, risks)  
- [x] Change log updated when requirements/architecture shift  

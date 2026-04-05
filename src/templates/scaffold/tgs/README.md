# Thoughts (TGS) Directory

This `tgs/` directory holds the thinking and governance system (TGS) docs for your repo. It is intentionally minimal and repository-agnostic.

## Structure

- `design/` — Long-lived system-level design docs
  - `00_context.md`
  - `10_needs.md`
  - `20_requirements.md`
  - `30_architecture.md`
  - `40_vnv.md`
  - `50_decisions.md`
- `agentops/` — AgentOps workflow and per-thought scaffolding templates
  - `AGENTOPS.md` — Canonical workflow (system prompt)
  - `tgs/` — Templates used by `make new-thought`
  - `prompts/` — Prompt files consumed by TGS commands
- `adapters/` — Shell adapters for AI vendor CLIs (claude, gemini)
- `context/` — Context packing templates
- `thoughts/` — Per-thought working directories (created by `make new-thought`)

## Usage

Start a new thought:

```bash
make new-thought title="My Feature" spec="One-line idea"
```

Follow `tgs/agentops/AGENTOPS.md` for the approval-gated workflow.

## Notes

- Production code must not live under `tgs/`.
- `tgs/tgs.yml` is the TGS configuration file — edit the `project:` field and any AI provider settings to taste.

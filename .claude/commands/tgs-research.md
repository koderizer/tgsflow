# TGS Research Phase

You are starting the research phase of a TGS thought. Follow these steps exactly.

## Steps

### 1. Create Thought Directory
```bash
make new-thought title="<title>" spec="<spec>"
```

### 2. Scan Design Docs
- Read `tgs/design/10_needs.md` — any relevant existing needs?
- Read `tgs/design/20_requirements.md` — any relevant requirements?
- Update these docs if the task introduces new needs or requirements.

### 3. Author research.md
Use the template at `tgs/agentops/tgs/research.md`. Fill every section including:
- **Perspectives Considered** (end-user, security, privacy, operator, future-debt)
- **Ethics Quick-Check** (data collection, automated decisions, oversight, harm)

### 4. Right-Size the Document
| Complexity | Target |
|-----------|--------|
| Bug fix | ~100 lines |
| Single-module | ~250 lines |
| Multi-module | ~400 lines |
| New architecture | ~1,000 lines |

### 5. Commit, Push, and Ask for Review
Push to GitHub — the human reviews there, not locally.

Tell the human: "Please review `research.md` in `tgs/<dir>/`. Reply: APPROVE research | REQUEST CHANGES: <notes>."

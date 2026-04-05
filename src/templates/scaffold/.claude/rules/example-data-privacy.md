---
description: "Data privacy routing — global hard rule example"
---

## Data Privacy Routing (HARD RULE)

User's personal data MUST stay on the user's device (kernel/local storage), NEVER routed to cloud services.

**Why:** This is the fundamental trust contract. Cloud handles platform services (auth, billing, external APIs). User data (conversations, memory, preferences) stays local. Violating this compromises the entire product trust model.

## Human Craft Zones

These areas require human authorship or mandatory expert review:
- Authentication and authorization flows
- Cryptographic implementations
- Payment processing
- Privacy boundary decisions
- Security policy changes

**Why:** Industry data shows AI-generated code has 2.74x more vulnerabilities (Veracode 2025). Critical paths need human judgment, not just AI generation.

# db1ffa7 - simplify-bootstrap-upgrade-path

- Base Hash: `db1ffa7`

## Quick Links
- [research.md](./research.md)
- [plan.md](./plan.md)
- [implementation.md](./implementation.md)

## Idea Spec
Redesign bootstrap.sh so it provides a clean, simple upgrade path for repos with old tgsflow/tgs directories. Current decorate mode copies .tmpl files as-is, misses .claude/ setup, has stale tgs.mk, and clones the full repo unnecessarily.

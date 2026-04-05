.PHONY: new-thought

new-thought:
	@if ! command -v git >/dev/null; then echo "git not found in PATH"; exit 2; fi; \
	if [ -z "$(title)" ]; then echo "Usage: make new-thought title=\"short title\" [spec=\"idea\"]"; exit 1; fi; \
	if [ ! -d "tgs/agentops/tgs" ]; then echo "Templates missing at tgs/agentops/tgs"; exit 2; fi; \
	HASH=$$(git rev-parse --short HEAD); \
	SLUG=$$(printf "%s" "$(title)" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9' '-' | sed 's/^-*//' | sed 's/-*$$//'); \
	DIR="tgs/thoughts/$$HASH-$$SLUG"; \
	mkdir -p "$$DIR"; \
	for f in tgs/agentops/tgs/*; do bn=$$(basename "$$f"); if [ ! -e "$$DIR/$$bn" ]; then cp "$$f" "$$DIR/"; fi; done; \
	if [ ! -f "$$DIR/README.md" ]; then \
		{ \
			printf "# %s - %s\n\n" "$$HASH" "$(title)"; \
			printf -- "- Base Hash: \`%s\`\n\n" "$$HASH"; \
			printf "## Quick Links\n- [research.md](./research.md)\n- [plan.md](./plan.md)\n- [implementation.md](./implementation.md)\n\n"; \
			if [ -n "$(spec)" ]; then printf "## Idea Spec\n%s\n" "$(spec)"; fi; \
		} > "$$DIR/README.md"; \
	fi; \
	echo "Created $$DIR"

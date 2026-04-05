#!/usr/bin/env bash
# TGSFlow bootstrap — fetch the modern scaffold and decorate the current directory.
# See https://github.com/akelv/tgsflow
set -euo pipefail

REPO="${TGS_REPO:-akelv/tgsflow}"
BRANCH="${TGS_BRANCH:-main}"
ARCHIVE_URL="https://codeload.github.com/${REPO}/tar.gz/refs/heads/${BRANCH}"

FORCE=0
DRY_RUN=0

usage() {
  cat >&2 <<EOF
Usage: bootstrap.sh [--force] [--dry-run] [--help]

Downloads the TGS scaffold from github.com/${REPO} and copies it into the
current directory. Files under tgs/thoughts/ are never touched.

Flags:
  --force     Overwrite TGS-managed files (default: preserve existing).
  --dry-run   List what would be copied without making changes.
  -h, --help  Show this help.

Environment:
  TGS_REPO    Override source repo (default: akelv/tgsflow)
  TGS_BRANCH  Override branch (default: main)
EOF
}

log() { printf '[tgs] %s\n' "$*"; }
die() { printf '[tgs] ERROR: %s\n' "$*" >&2; exit 1; }

for arg in "$@"; do
  case "$arg" in
    --force)   FORCE=1 ;;
    --dry-run) DRY_RUN=1 ;;
    -h|--help) usage; exit 0 ;;
    *) printf '[tgs] unknown flag: %s\n' "$arg" >&2; usage; exit 2 ;;
  esac
done

command -v curl >/dev/null || die "curl not found in PATH"
command -v tar  >/dev/null || die "tar not found in PATH"

TMPDIR=$(mktemp -d) || die "mktemp failed"
trap 'rm -rf "$TMPDIR"' EXIT

log "fetching $ARCHIVE_URL"
curl -fsSL "$ARCHIVE_URL" -o "$TMPDIR/archive.tgz" || die "download failed"

log "extracting archive"
tar -xzf "$TMPDIR/archive.tgz" -C "$TMPDIR"

# Locate the scaffold directory inside the extracted tree.
SRC=$(find "$TMPDIR" -type d -path '*/src/templates/scaffold' -print 2>/dev/null | head -1)
[ -n "$SRC" ] && [ -d "$SRC" ] || die "scaffold not found in archive"

copy_one() {
  local src="$1" dst="$2"
  if [ -e "$dst" ] && [ "$FORCE" -eq 0 ]; then
    log "skip (exists): $dst"
    return 0
  fi
  if [ "$DRY_RUN" -eq 1 ]; then
    log "would write: $dst"
    return 0
  fi
  mkdir -p "$(dirname "$dst")"
  cp -p "$src" "$dst"
  log "wrote: $dst"
}

# Walk scaffold and copy each file. Never touch tgs/thoughts/* (user content).
(cd "$SRC" && find . -type f ! -path './tgs/thoughts/*') | while IFS= read -r rel; do
  rel="${rel#./}"
  copy_one "$SRC/$rel" "./$rel"
done

# Ensure thoughts dir exists for `make new-thought`.
if [ "$DRY_RUN" -eq 0 ]; then
  mkdir -p ./tgs/thoughts
fi

# Ensure Makefile includes Makefile.tgs.mk (skip if new-thought already present).
inc="include Makefile.tgs.mk"
if [ ! -f Makefile ]; then
  if [ "$DRY_RUN" -eq 1 ]; then
    log "would create Makefile with '$inc'"
  else
    printf '%s\n' "$inc" > Makefile
    log "created Makefile"
  fi
elif ! grep -Eq '^new-thought:|Makefile\.tgs\.mk' Makefile 2>/dev/null; then
  if [ "$DRY_RUN" -eq 1 ]; then
    log "would append '$inc' to Makefile"
  else
    printf '\n%s\n' "$inc" >> Makefile
    log "appended include to Makefile"
  fi
fi

log "done"

#!/usr/bin/env bash
# Behavior test for install.sh at its external seam:
# FBH_SKILLS_DIR override in, symlinks out. Asserts idempotency.
set -u

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SANDBOX="$(mktemp -d)"
trap 'rm -rf "$SANDBOX"' EXIT

fail() { echo "FAIL: $1" >&2; exit 1; }

# 1. First install: every skill dir in skills/ is linked into the target.
FBH_SKILLS_DIR="$SANDBOX/skills" bash "$REPO_DIR/install.sh" >/dev/null 2>&1 \
    || fail "first install exited non-zero"

for skill_dir in "$REPO_DIR"/skills/*/; do
    name="$(basename "$skill_dir")"
    link="$SANDBOX/skills/$name"
    [ -L "$link" ] || fail "$name not linked"
    [ "$(readlink "$link")" = "${skill_dir%/}" ] || fail "$name links to wrong target: $(readlink "$link")"
done

# 2. Idempotency: re-run exits zero and leaves exactly one link per skill.
FBH_SKILLS_DIR="$SANDBOX/skills" bash "$REPO_DIR/install.sh" >/dev/null 2>&1 \
    || fail "second install exited non-zero"

want="$(ls -1d "$REPO_DIR"/skills/*/ | wc -l | tr -d ' ')"
got="$(ls -1 "$SANDBOX/skills" | wc -l | tr -d ' ')"
[ "$got" = "$want" ] || fail "expected $want entries after re-install, got $got"

echo "PASS: install.sh links skills and is idempotent"

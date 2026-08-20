#!/usr/bin/env bash
# Installs the harness: symlinks every skill under skills/ into the
# local Claude Code skills directory, and builds the fbh CLI.
# Idempotent: safe to re-run after every `git pull`.
#
# FBH_SKILLS_DIR overrides the target directory (used by tests).
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET_DIR="${FBH_SKILLS_DIR:-$HOME/.claude/skills}"

mkdir -p "$TARGET_DIR"

for skill_dir in "$REPO_DIR"/skills/*/; do
    name="$(basename "$skill_dir")"
    link="$TARGET_DIR/$name"
    src="${skill_dir%/}"
    if [ -L "$link" ]; then
        if [ "$(readlink "$link")" = "$src" ]; then
            echo "ok: $name already linked"
            continue
        fi
        echo "error: $link points elsewhere: $(readlink "$link")" >&2
        exit 1
    fi
    if [ -e "$link" ]; then
        echo "error: $link exists and is not a symlink" >&2
        exit 1
    fi
    ln -s "$src" "$link"
    echo "linked: $name -> $src"
done

if command -v go >/dev/null 2>&1; then
    (cd "$REPO_DIR" && go build -o bin/fbh ./cmd/fbh)
    echo "built: bin/fbh ($("$REPO_DIR/bin/fbh" --version))"
else
    echo "warn: go not found, skipped building fbh" >&2
fi

echo "install complete"

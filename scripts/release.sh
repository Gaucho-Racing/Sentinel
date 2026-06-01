#!/usr/bin/env bash
set -euo pipefail

# Cuts a Sentinel release across all services.
#
#   ./scripts/release.sh 5.1.0
#       Bumps the Version constant in core/config, oauth/config, and
#       discord/config to 5.1.0, commits, pushes, and creates a v5.1.0 GH
#       release. The per-service workflows (core.yml/oauth.yml/discord.yml)
#       observe the tag and publish images to ghcr.io.
#
# Run from any subdirectory; the script cd's to the repo root.

usage() {
    cat <<EOF
Usage: $0 <version>

Examples:
  $0 5.1.0
  $0                       # prompts for version, shows current release
EOF
}

while getopts ":h" opt; do
    case $opt in
        h) usage; exit 0 ;;
        *) usage; exit 1 ;;
    esac
done
shift $((OPTIND - 1))

INPUT="${1:-}"

for cmd in gh git; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "Error: $cmd is required"
        exit 1
    fi
done

BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$BRANCH" != "main" ]]; then
    echo "Error: must be on main branch (currently on $BRANCH)"
    exit 1
fi

git fetch origin main --tags --quiet
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/main)
if [[ "$LOCAL" != "$REMOTE" ]]; then
    echo "Error: local main is not up to date with origin/main"
    echo "  local:  $LOCAL"
    echo "  remote: $REMOTE"
    exit 1
fi

PREV=$(git tag -l 'v*' | sort -V | tail -n1)

if [[ -z "$INPUT" ]]; then
    echo ""
    if [[ -n "$PREV" ]]; then
        echo "Current release: ${PREV}"
    else
        echo "Current release: (none)"
    fi
    echo ""
    read -rp "Enter new version: " INPUT
fi

if [[ -z "$INPUT" ]]; then
    echo "Error: version cannot be empty"
    exit 1
fi
INPUT="${INPUT#v}"
if [[ ! "$INPUT" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: version must be a valid semver (e.g. 5.1.0)"
    exit 1
fi
SEMVER="$INPUT"
VERSION="v${INPUT}"
TAG="$VERSION"

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT"

if git tag -l "$TAG" | grep -q "^${TAG}$"; then
    echo "Error: tag $TAG already exists"
    exit 1
fi

SERVICES=("core" "oauth" "discord")

echo ""
echo "=== Release Summary ==="
echo "  Version: ${VERSION}"
echo "  Tag:     ${TAG}"
echo "  Commit:  $(git rev-parse --short HEAD)"
echo "  Branch:  main"
echo ""
echo "  Files to update:"
for svc in "${SERVICES[@]}"; do
    echo "    ${svc}/config/config.go"
done
echo ""
echo "  Docker images that will be tagged:"
for svc in "${SERVICES[@]}"; do
    echo "    ghcr.io/gaucho-racing/sentinel-${svc}:${SEMVER}"
done
echo ""
read -rp "Proceed? (y/N) " CONFIRM
if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
    echo "Aborted."
    exit 0
fi

# Bump the Version field in each service's rincon.Service struct literal.
# Matches lines like:  Version:     "5.0.0",
for svc in "${SERVICES[@]}"; do
    sed -i '' "s/Version:.*\".*\"/Version:     \"${SEMVER}\"/" "${REPO_ROOT}/${svc}/config/config.go"
done

FILES=()
for svc in "${SERVICES[@]}"; do
    FILES+=("${svc}/config/config.go")
done
git add "${FILES[@]}"
# --allow-empty so the release still cuts when the Version constants are
# already at target (no-op sed). The release commit is the anchor for the
# tag; whether it changes files isn't important.
git commit --allow-empty -m "release: sentinel ${VERSION}"
git push origin main

gh release create "$TAG" \
    --target main \
    --title "${VERSION}" \
    --generate-notes

echo ""
echo "Done. ${TAG} released. Per-service workflows will publish images shortly."

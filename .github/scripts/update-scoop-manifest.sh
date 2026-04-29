#!/usr/bin/env bash
set -euo pipefail

: "${TAG_NAME:?TAG_NAME is required}"
: "${GITHUB_REPOSITORY_OWNER:?GITHUB_REPOSITORY_OWNER is required}"
: "${SCOOP_BUCKET_PAT:?SCOOP_BUCKET_PAT is required}"

VERSION="${TAG_NAME#v}"
REPO_OWNER="${GITHUB_REPOSITORY_OWNER}"
BUCKET_REPO="scoop-bucket"
WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

WIN_ASSET="broask_${VERSION}_windows_amd64.zip"
ASSET_URL="https://github.com/${REPO_OWNER}/broask/releases/download/${TAG_NAME}/${WIN_ASSET}"
CHECKSUM_URL="https://github.com/${REPO_OWNER}/broask/releases/download/${TAG_NAME}/checksums.txt"

CHECKSUM="$(curl -fsSL "$CHECKSUM_URL" | awk -v file="$WIN_ASSET" '$2 == file { print $1 }')"
if [[ -z "$CHECKSUM" ]]; then
  echo "Could not find checksum for $WIN_ASSET in checksums.txt" >&2
  exit 1
fi

git clone "https://x-access-token:${SCOOP_BUCKET_PAT}@github.com/${REPO_OWNER}/${BUCKET_REPO}.git" "$WORKDIR/${BUCKET_REPO}"
cd "$WORKDIR/${BUCKET_REPO}"
mkdir -p bucket

cat > bucket/broask.json <<EOF
{
  "version": "${VERSION}",
  "description": "Play a sound when CLI tools ask for confirmation",
  "homepage": "https://github.com/${REPO_OWNER}/broask",
  "license": "MIT",
  "architecture": {
    "64bit": {
      "url": "${ASSET_URL}",
      "hash": "${CHECKSUM}"
    }
  },
  "bin": "broask.exe"
}
EOF

git add bucket/broask.json
if git diff --cached --quiet -- bucket/broask.json; then
  echo "No scoop manifest changes to commit"
  exit 0
fi
git -c user.name='Mulham' -c user.email='mulhamna@gmail.com' commit -m "chore: update broask scoop manifest to ${VERSION}"
git push origin HEAD:main

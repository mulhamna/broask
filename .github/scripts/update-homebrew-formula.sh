#!/usr/bin/env bash
set -euo pipefail

: "${TAG_NAME:?TAG_NAME is required}"
: "${HOMEBREW_TAP_TOKEN:?HOMEBREW_TAP_TOKEN is required}"

VERSION="${TAG_NAME#v}"
WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

CHECKSUMS_URL="https://github.com/mulhamna/broask/releases/download/${TAG_NAME}/checksums.txt"
sha_darwin_arm64="$(curl -fsSL "$CHECKSUMS_URL" | awk -v file="broask_${VERSION}_darwin_arm64.tar.gz" '$2 == file { print $1 }')"
sha_darwin_amd64="$(curl -fsSL "$CHECKSUMS_URL" | awk -v file="broask_${VERSION}_darwin_amd64.tar.gz" '$2 == file { print $1 }')"
sha_linux_arm64="$(curl -fsSL "$CHECKSUMS_URL" | awk -v file="broask_${VERSION}_linux_arm64.tar.gz" '$2 == file { print $1 }')"
sha_linux_amd64="$(curl -fsSL "$CHECKSUMS_URL" | awk -v file="broask_${VERSION}_linux_amd64.tar.gz" '$2 == file { print $1 }')"

for value in "$sha_darwin_arm64" "$sha_darwin_amd64" "$sha_linux_arm64" "$sha_linux_amd64"; do
  test -n "$value"
done

git clone "https://x-access-token:${HOMEBREW_TAP_TOKEN}@github.com/mulhamna/homebrew-tap.git" "$WORKDIR/homebrew-tap"
cd "$WORKDIR/homebrew-tap"
test -f Formula/broask.rb

python3 - <<PY
from pathlib import Path
import os
import re

path = Path('Formula/broask.rb')
content = path.read_text()
version = os.environ['VERSION']
replacements = {
    'REPLACE_DARWIN_ARM64_SHA': os.environ['SHA_DARWIN_ARM64'],
    'REPLACE_DARWIN_AMD64_SHA': os.environ['SHA_DARWIN_AMD64'],
    'REPLACE_LINUX_ARM64_SHA': os.environ['SHA_LINUX_ARM64'],
    'REPLACE_LINUX_AMD64_SHA': os.environ['SHA_LINUX_AMD64'],
}

content = re.sub(r'version "[^"]+"', f'version "{version}"', content)
for old, new in replacements.items():
    content = content.replace(old, new)

for arch, sha in [
    ('darwin_arm64', os.environ['SHA_DARWIN_ARM64']),
    ('darwin_amd64', os.environ['SHA_DARWIN_AMD64']),
    ('linux_arm64', os.environ['SHA_LINUX_ARM64']),
    ('linux_amd64', os.environ['SHA_LINUX_AMD64']),
]:
    content = re.sub(
        rf'(broask/releases/download/v)#\{{version\}}(/broask_)\#\{{version\}}_{arch}(\.tar\.gz"\n\s+sha256 ")[^"]+"',
        rf'\g<1>#{{version}}\2#{{version}}_{arch}\3{sha}"',
        content,
    )

path.write_text(content)
PY

export VERSION
export SHA_DARWIN_ARM64="$sha_darwin_arm64"
export SHA_DARWIN_AMD64="$sha_darwin_amd64"
export SHA_LINUX_ARM64="$sha_linux_arm64"
export SHA_LINUX_AMD64="$sha_linux_amd64"

cd "$WORKDIR/homebrew-tap"
git config user.name 'github-actions[bot]'
git config user.email 'github-actions[bot]@users.noreply.github.com'
git add Formula/broask.rb
git diff --staged --quiet && echo "No changes to commit" && exit 0
git commit -m "chore: bump broask to v${VERSION}"
git push origin HEAD:main

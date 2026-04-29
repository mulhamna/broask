#!/usr/bin/env bash
set -euo pipefail

: "${TAG_NAME:?TAG_NAME is required}"
: "${HOMEBREW_TAP_TOKEN:?HOMEBREW_TAP_TOKEN is required}"

VERSION="${TAG_NAME#v}"
WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

CHECKSUMS_URL="https://github.com/mulhamna/broask/releases/download/${TAG_NAME}/checksums.txt"
SHA_DARWIN_AMD64="$(curl -fsSL "$CHECKSUMS_URL" | awk -v file="broask_${VERSION}_darwin_amd64.tar.gz" '$2 == file { print $1 }')"
SHA_DARWIN_ARM64="$(curl -fsSL "$CHECKSUMS_URL" | awk -v file="broask_${VERSION}_darwin_arm64.tar.gz" '$2 == file { print $1 }')"
SHA_LINUX_AMD64="$(curl -fsSL "$CHECKSUMS_URL" | awk -v file="broask_${VERSION}_linux_amd64.tar.gz" '$2 == file { print $1 }')"
SHA_LINUX_ARM64="$(curl -fsSL "$CHECKSUMS_URL" | awk -v file="broask_${VERSION}_linux_arm64.tar.gz" '$2 == file { print $1 }')"

for value in "$SHA_DARWIN_AMD64" "$SHA_DARWIN_ARM64" "$SHA_LINUX_AMD64" "$SHA_LINUX_ARM64"; do
  test -n "$value"
done

git clone "https://x-access-token:${HOMEBREW_TAP_TOKEN}@github.com/mulhamna/homebrew-tap.git" "$WORKDIR/homebrew-tap"
cd "$WORKDIR/homebrew-tap"
test -f Formula/broask.rb

cat > Formula/broask.rb <<EOF
class Broask < Formula
  desc "Play a sound when CLI tools ask for confirmation"
  homepage "https://github.com/mulhamna/broask"
  version "${VERSION}"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/mulhamna/broask/releases/download/v#{version}/broask_#{version}_darwin_arm64.tar.gz"
      sha256 "${SHA_DARWIN_ARM64}"
    end

    on_intel do
      url "https://github.com/mulhamna/broask/releases/download/v#{version}/broask_#{version}_darwin_amd64.tar.gz"
      sha256 "${SHA_DARWIN_AMD64}"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/mulhamna/broask/releases/download/v#{version}/broask_#{version}_linux_arm64.tar.gz"
      sha256 "${SHA_LINUX_ARM64}"
    end

    on_intel do
      url "https://github.com/mulhamna/broask/releases/download/v#{version}/broask_#{version}_linux_amd64.tar.gz"
      sha256 "${SHA_LINUX_AMD64}"
    end
  end

  def install
    bin.install "broask"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/broask version")
  end
end
EOF

git config user.name 'github-actions[bot]'
git config user.email 'github-actions[bot]@users.noreply.github.com'
git add Formula/broask.rb
git diff --staged --quiet && echo "No changes to commit" && exit 0
git commit -m "chore: bump broask to v${VERSION}"
git push origin HEAD:main

# Contributing to broask

Thanks for contributing to broask.

## Development setup

Requirements:
- Go 1.22+

Clone the repo and run locally:

```bash
git clone https://github.com/mulhamna/broask.git
cd broask
go run . -- --help
```

Build the binary:

```bash
go build -o broask .
./broask --help
```

## Before opening a PR

Please keep changes focused and easy to review.

Checklist:
- use clear commit messages
- update docs when behavior changes
- test the CLI flow you touched
- avoid unrelated refactors in the same PR

## Release notes

This repo uses Conventional Commits for automated releases.

Examples:
- `feat: add new config option`
- `fix: avoid duplicate sound trigger`
- `docs: clarify install steps`

## Pull requests

When opening a PR:
- explain what changed
- explain why it changed
- include any user-facing behavior change
- mention platform-specific caveats if relevant

## Scope

Good contributions include:
- prompt detection improvements
- audio backend fixes
- CLI usability improvements
- docs and install improvements
- packaging and release automation fixes

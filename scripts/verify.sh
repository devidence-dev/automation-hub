#!/bin/sh
# Local gate of this repo. The shared checks (gofmt, golangci-lint, go test -race with the coverage
# minimum, gocognit, biome ci) run through gate (https://github.com/devidence-dev/dev-gate), which
# holds every pin. This file adds the two checks specific to this repo: the drift of the generated
# assets and openspec validate, which runs through the host's openspec wrapper (its version is pinned
# there, in one place).
#
# Usage: scripts/verify.sh [--fix]
#   --fix  regenerates the assets, then runs gate go --fix (gofmt -w, biome check --write) before the checks.
set -eu
cd "$(dirname "$0")/.."
command -v gate >/dev/null 2>&1 || {
	echo "FAIL: gate is not on PATH (install: https://github.com/devidence-dev/dev-gate)" >&2
	exit 1
}

# check_drift regenerates the assets and fails when a generated file differs from the index or when the
# generation creates an untracked generated file. git diff does not see new files, so both are checked.
# The check is limited to the generated paths, so unrelated unstaged work does not fail the gate.
check_drift() {
	echo "==> generated assets (scripts/generate-depreport-assets.sh)"
	sh scripts/generate-depreport-assets.sh
	if ! git diff --exit-code -- '*_templ.go' internal/services/depreport/assets/depreport.min.css; then
		echo "FAIL: generated assets differ from the index (run scripts/verify.sh --fix and stage them)" >&2
		exit 1
	fi
	untracked="$(git ls-files --others --exclude-standard -- '*_templ.go' internal/services/depreport/assets/depreport.min.css)"
	if [ -n "$untracked" ]; then
		echo "$untracked"
		echo "FAIL: generation created untracked files (stage them)" >&2
		exit 1
	fi
	echo "ok: generated assets match the index"
}

# check_openspec validates every change and spec with the host's openspec command. Its version is
# pinned in one place only: the ~/.local/bin/openspec wrapper (OPENSPEC_VERSION).
check_openspec() {
	if ! command -v openspec >/dev/null 2>&1; then
		echo "FAIL: openspec is not on PATH (expected the ~/.local/bin/openspec wrapper)"
		exit 1
	fi
	echo "==> openspec validate ($(openspec --version), host wrapper)"
	openspec validate --all --strict
	echo "ok: openspec validate passes"
}

fix=no
case "$*" in
"") ;;
--fix) fix=yes ;;
*)
	echo "usage: scripts/verify.sh [--fix]" >&2
	exit 2
	;;
esac

if [ "$fix" = yes ]; then
	sh scripts/generate-depreport-assets.sh
	gate go --fix --gofmt --cognit --biome
else
	gate go --gofmt --cognit --biome
fi
check_drift
check_openspec
echo "==> verify passed"

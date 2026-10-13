#!/bin/sh
# Local gate of this repo (openspec design D15), modelled on voraz-bot's scripts/validate.sh. It runs
# the checks of the shared CI (devidence-dev/workflows ci-go.yml) plus Biome and OpenSpec. In Docker:
# gofmt, golangci-lint, go test -race with the coverage minimum, gocognit, the drift of the generated
# assets and biome ci. openspec validate runs through the host's openspec wrapper, the one place its
# version is pinned. Caches live in ~/.cache, shared by every project.
#
# Usage: scripts/verify.sh [--fix]
#   --fix  runs gofmt -w, biome check --write and the asset generation first, then the checks.
#
# Set VERIFY_SOURCE_ONLY=1 to load the functions without running the gate (used by the checks).
set -eu

GO_IMAGE="golang:$(cat .go-version)"
LINT_IMAGE="golangci/golangci-lint:v2.14.0"
# Same linter set as CI (devidence-dev/workflows ci-go.yml); the defaults alone miss gosec and misspell.
LINTERS="errcheck,govet,staticcheck,unused,misspell,gosec"
GOCOGNIT_PKG="github.com/uudashr/gocognit/cmd/gocognit@v1.2.0"
COGNITIVE_MAX=15
COGNITIVE_IGNORE='_templ\.go$'
MIN_COVERAGE="${MIN_COVERAGE:-80}"
# Biome 2.5.15 standalone binaries, pinned by SHA-256 (openspec design D16). The release publishes no
# checksum file; these hashes equal the GitHub asset digests. A download that does not match stops the gate.
BIOME_VERSION="2.5.15"
BIOME_BASE="https://github.com/biomejs/biome/releases/download/@biomejs/biome@$BIOME_VERSION"
BIOME_SHA256_X64="5d867a0b2ccea1755b7e508b01d836a8c64e31a1e4d11ec24b43b6b4eec8b3b6"
BIOME_SHA256_ARM64="a56bcd73ddbfdb0f57a82bb67b7ea698eed4027aff602b95aa18fafe919d59e0"
USER_ID="$(id -u):$(id -g)"

# prepare_caches creates the caches as the current user; otherwise Docker creates them as root.
prepare_caches() {
	mkdir -p "$HOME/.cache/go-build" "$HOME/.cache/go-mod" "$HOME/.cache/golangci-lint" \
		"$HOME/.cache/biome" "$HOME/.cache/npm"
}

# fix runs the formatters and regenerates the assets.
fix() {
	echo "==> fix: gofmt -w, biome check --write, asset generation"
	docker run --rm --user "$USER_ID" -e HOME=/tmp -v "$PWD":/app -w /app "$GO_IMAGE" gofmt -w .
	run_biome check --write .
	sh scripts/generate-depreport-assets.sh
}

# check_gofmt fails when gofmt lists a file.
check_gofmt() {
	echo "==> gofmt -l ($GO_IMAGE)"
	unformatted="$(docker run --rm --user "$USER_ID" -e HOME=/tmp -v "$PWD":/app -w /app "$GO_IMAGE" gofmt -l .)"
	if [ -n "$unformatted" ]; then
		echo "$unformatted"
		echo "FAIL: gofmt lists the files above (run scripts/verify.sh --fix)" >&2
		exit 1
	fi
	echo "ok: gofmt lists no file"
}

# check_lint runs golangci-lint with the CI linters.
check_lint() {
	echo "==> golangci-lint ($LINT_IMAGE, linters: $LINTERS)"
	docker run --rm --user "$USER_ID" -e HOME=/tmp -e GOPATH=/tmp/gopath -e GOLANGCI_LINT_CACHE=/lintcache \
		-v "$HOME/.cache/golangci-lint":/lintcache -v "$HOME/.cache/go-mod":/go/pkg/mod \
		-v "$PWD":/app -w /app "$LINT_IMAGE" golangci-lint run --timeout=5m --enable="$LINTERS" ./...
	echo "ok: golangci-lint exits 0"
}

# check_tests runs the tests with the race detector and fails below MIN_COVERAGE. The coverage ignores
# the files of sonar.coverage.exclusions, as CI does; the module path comes from go.mod.
check_tests() {
	module="$(sed -n 's/^module //p' go.mod)"
	excluded="$(sed -n 's/^sonar.coverage.exclusions=//p' sonar-project.properties | sed 's/\./\\./g; s/,/|/g')"
	echo "==> go test -race with coverage ($GO_IMAGE, minimum ${MIN_COVERAGE}%)"
	docker run --rm -i --user "$USER_ID" -e HOME=/tmp -e GOPATH=/tmp/gopath -e GOCACHE=/gocache -e GOMODCACHE=/gomod \
		-e MODULE="$module" -e EXCLUDED="$excluded" -e MIN_COVERAGE="$MIN_COVERAGE" \
		-v "$HOME/.cache/go-build":/gocache -v "$HOME/.cache/go-mod":/gomod \
		-v "$PWD":/app -w /app "$GO_IMAGE" sh -s <<'EOF'
set -eu
go test -race -covermode=atomic ./... -coverprofile=/tmp/all.out
grep -v -E "^${MODULE}/(${EXCLUDED}):" /tmp/all.out > /tmp/cover.out
pct="$(go tool cover -func=/tmp/cover.out | awk 'END { gsub("%", "", $3); print $3 }')"
echo "coverage: ${pct}% (minimum ${MIN_COVERAGE}%, sonar exclusions applied)"
awk -v p="$pct" -v m="$MIN_COVERAGE" 'BEGIN { exit (p + 0 >= m + 0) ? 0 : 1 }' \
	|| { echo "FAIL: coverage ${pct}% is below ${MIN_COVERAGE}%" >&2; exit 1; }
EOF
	echo "ok: tests pass with the coverage minimum"
}

# check_cognitive fails on any function above COGNITIVE_MAX, tests included. The generated *_templ.go
# files are skipped with COGNITIVE_IGNORE.
check_cognitive() {
	echo "==> gocognit -over $COGNITIVE_MAX -ignore '$COGNITIVE_IGNORE' ($GOCOGNIT_PKG)"
	docker run --rm -i --user "$USER_ID" -e HOME=/tmp -e GOPATH=/tmp/gopath -e GOCACHE=/gocache -e GOMODCACHE=/gomod \
		-e COGNITIVE_MAX="$COGNITIVE_MAX" -e COGNITIVE_IGNORE="$COGNITIVE_IGNORE" -e GOCOGNIT_PKG="$GOCOGNIT_PKG" \
		-v "$HOME/.cache/go-build":/gocache -v "$HOME/.cache/go-mod":/gomod \
		-v "$PWD":/app -w /app "$GO_IMAGE" sh -s <<'EOF'
set -eu
status=0
findings="$(go run "$GOCOGNIT_PKG" -test -over "$COGNITIVE_MAX" -ignore "$COGNITIVE_IGNORE" .)" || status=$?
if [ -n "$findings" ]; then
	echo "$findings"
	echo "FAIL: functions above complexity $COGNITIVE_MAX" >&2
	exit 1
fi
if [ "$status" -ne 0 ]; then
	echo "FAIL: gocognit did not run (exit $status)" >&2
	exit 1
fi
EOF
	echo "ok: no function above $COGNITIVE_MAX"
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

# run_biome runs the pinned Biome binary with the given arguments. The binary is downloaded into the
# cache once; its SHA-256 is checked before every run, and a mismatch stops the gate before Biome runs.
run_biome() {
	case "$(uname -m)" in
	x86_64) biome_asset="linux-x64" biome_sha256="$BIOME_SHA256_X64" ;;
	aarch64 | arm64) biome_asset="linux-arm64" biome_sha256="$BIOME_SHA256_ARM64" ;;
	*)
		echo "unsupported architecture: $(uname -m)" >&2
		exit 1
		;;
	esac
	docker run --rm -i --user "$USER_ID" -e HOME=/tmp \
		-e BIOME_URL="$BIOME_BASE/biome-$biome_asset" -e BIOME_SHA256="$biome_sha256" \
		-e BIOME_BIN="/cache/biome-$BIOME_VERSION-$biome_asset" \
		-v "$HOME/.cache/biome":/cache -v "$PWD":/app -w /app "$GO_IMAGE" sh -s -- "$@" <<'EOF'
set -eu
if ! echo "$BIOME_SHA256  $BIOME_BIN" | sha256sum -c - >/dev/null 2>&1; then
	rm -f "$BIOME_BIN"
	curl -fsSL -o "$BIOME_BIN.tmp" "$BIOME_URL"
	mv "$BIOME_BIN.tmp" "$BIOME_BIN"
fi
if ! echo "$BIOME_SHA256  $BIOME_BIN" | sha256sum -c -; then
	rm -f "$BIOME_BIN"
	echo "FAIL: the Biome binary does not match its pinned SHA-256" >&2
	exit 1
fi
cp "$BIOME_BIN" /tmp/biome
chmod +x /tmp/biome
/tmp/biome "$@"
EOF
}

# check_biome runs biome ci on the CSS and JS sources (biome.json limits the files).
check_biome() {
	echo "==> biome ci ($BIOME_VERSION, pinned by SHA-256)"
	run_biome ci .
	echo "ok: biome ci passes"
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

main() {
	cd "$(dirname "$0")/.."
	prepare_caches
	if [ "${1:-}" = "--fix" ]; then
		fix
	fi
	check_gofmt
	check_lint
	check_tests
	check_cognitive
	check_drift
	check_biome
	check_openspec
	echo "==> verify passed"
}

if [ -z "${VERIFY_SOURCE_ONLY:-}" ]; then
	main "$@"
fi

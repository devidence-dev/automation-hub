#!/bin/sh
# Generates the committed assets of the dependency report: the templ components (*_templ.go)
# and the minified Tailwind CSS (assets/depreport.min.css). It runs in an ephemeral golang
# container, so nothing is installed on the host. A second run changes nothing.
#
# Pinned tools and their checksum sources (verified 2026-10-12):
#   templ v0.3.1070: checksums.txt in the GitHub release, lines "<sha256>  templ_Linux_x86_64.tar.gz"
#                    (the archive holds the templ binary).
#   tailwindcss v4.3.3: sha256sums.txt in the GitHub release, lines "<sha256>  ./tailwindcss-linux-x64"
#                    (the standalone binary).
# Each download is checked against its line in the release's checksum file before it runs.
# A missing line or a hash mismatch stops the script.
#
# Usage: scripts/generate-depreport-assets.sh   (needs docker; run from anywhere in the repo)
# Set DEPREPORT_SOURCE_ONLY=1 to load the functions without running them (used by the checks).

set -eu

TEMPL_VERSION=v0.3.1070
TEMPL_BASE=https://github.com/a-h/templ/releases/download/$TEMPL_VERSION
TAILWIND_VERSION=v4.3.3
TAILWIND_BASE=https://github.com/tailwindlabs/tailwindcss/releases/download/$TAILWIND_VERSION
CSS_INPUT=internal/services/depreport/assets/input.css
CSS_OUTPUT=internal/services/depreport/assets/depreport.min.css

# verify_sha256 FILE SUMS NAME checks FILE against the line NAME of the checksum file SUMS.
verify_sha256() {
	want=$(awk -v name="$3" '{ sub(/^\.\//, "", $2) } $2 == name { print $1 }' "$2")
	if [ -z "$want" ]; then
		echo "no checksum line for $3 in $2" >&2
		return 1
	fi
	got=$(sha256sum "$1" | awk '{ print $1 }')
	if [ "$got" != "$want" ]; then
		echo "checksum mismatch for $3: got $got, want $want" >&2
		return 1
	fi
}

# fetch URL SUMS_URL NAME OUT downloads an asset and its checksum file, and keeps the asset at OUT
# only after it verifies against NAME.
fetch() {
	curl -fsSL -o "$4" "$1"
	curl -fsSL -o "$4.sums" "$2"
	verify_sha256 "$4" "$4.sums" "$3"
}

# arch sets the release names of the container architecture.
arch() {
	case $(uname -m) in
	x86_64) TEMPL_ARCH=Linux_x86_64 TAILWIND_ARCH=linux-x64 ;;
	aarch64 | arm64) TEMPL_ARCH=Linux_arm64 TAILWIND_ARCH=linux-arm64 ;;
	*)
		echo "unsupported architecture: $(uname -m)" >&2
		return 1
		;;
	esac
}

# generate runs inside the container: it fetches the verified tools and builds the assets.
generate() {
	arch
	tools=$(mktemp -d)
	fetch "$TEMPL_BASE/templ_$TEMPL_ARCH.tar.gz" "$TEMPL_BASE/checksums.txt" "templ_$TEMPL_ARCH.tar.gz" "$tools/templ.tar.gz"
	tar -xzf "$tools/templ.tar.gz" -C "$tools" templ
	fetch "$TAILWIND_BASE/tailwindcss-$TAILWIND_ARCH" "$TAILWIND_BASE/sha256sums.txt" "tailwindcss-$TAILWIND_ARCH" "$tools/tailwindcss"
	chmod +x "$tools/templ" "$tools/tailwindcss"
	"$tools/templ" generate
	"$tools/tailwindcss" -i "$CSS_INPUT" -o "$CSS_OUTPUT" --minify
}

# main runs generate inside the container, or starts that container when called on the host.
main() {
	if [ -n "${DEPREPORT_IN_CONTAINER:-}" ]; then
		generate
		return
	fi
	cd "$(dirname "$0")/.."
	if [ ! -f "$CSS_INPUT" ]; then
		echo "missing Tailwind CSS input: $CSS_INPUT" >&2
		exit 1
	fi
	exec docker run --rm --user "$(id -u):$(id -g)" -e HOME=/tmp -e DEPREPORT_IN_CONTAINER=1 \
		-v "$PWD":/app -w /app "golang:$(cat .go-version)" sh /app/scripts/generate-depreport-assets.sh
}

if [ -z "${DEPREPORT_SOURCE_ONLY:-}" ]; then
	main "$@"
fi

#!/bin/sh
# Generates the committed assets of the dependency report: the templ components (*_templ.go) and the
# minified Tailwind CSS (assets/depreport.min.css). The pinned templ and Tailwind run through gate
# (https://github.com/devidence-dev/dev-gate); this file adds only the input check. A second run
# changes nothing.
# Usage: scripts/generate-depreport-assets.sh
set -eu
cd "$(dirname "$0")/.."
command -v gate >/dev/null 2>&1 || {
	echo "FAIL: gate is not on PATH (install: https://github.com/devidence-dev/dev-gate)" >&2
	exit 1
}
[ $# -eq 0 ] || {
	echo "usage: scripts/generate-depreport-assets.sh" >&2
	exit 2
}

CSS_INPUT=internal/services/depreport/assets/input.css
CSS_OUTPUT=internal/services/depreport/assets/depreport.min.css

if [ ! -f "$CSS_INPUT" ]; then
	echo "missing Tailwind CSS input: $CSS_INPUT" >&2
	exit 1
fi
gate templ generate
gate tailwind -i "$CSS_INPUT" -o "$CSS_OUTPUT" --minify

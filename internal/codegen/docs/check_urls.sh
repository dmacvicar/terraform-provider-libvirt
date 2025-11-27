#!/usr/bin/env bash
set -euo pipefail

# Check all reference URLs in the YAML docs for reachability (HTTP) and anchor presence (when specified).
# Usage: ./check_urls.sh

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

# Collect unique URLs from YAML files (reference fields only), forgiving malformed YAML
readarray -t URLS < <(python3 - <<'PY'
import pathlib
import re

docs_dir = pathlib.Path(__file__).resolve().parent
seen = set()
ref_re = re.compile(r'^\s*reference:\s*("?)([^"#\s][^"]*?)\1\s*$')

for path in docs_dir.glob("*.yaml"):
    try:
        for line in path.read_text().splitlines():
            m = ref_re.match(line)
            if m:
                ref = m.group(2).strip()
                if ref:
                    seen.add(ref)
    except Exception as exc:
        print(f"Skipping {path} due to error: {exc}", file=sys.stderr)

for url in sorted(seen):
    print(url)
PY
)

if [[ ${#URLS[@]} -eq 0 ]]; then
    echo "No URLs found in YAML docs."
    exit 0
fi

echo "Found ${#URLS[@]} unique URLs"
failures=0

for url in "${URLS[@]}"; do
    base="${url%%#*}"
    anchor=""
    if [[ "$url" == *"#"* ]]; then
        anchor="${url#*#}"
    fi

    echo "Checking $url"

    # Fetch the page (follow redirects). Store body for anchor checks.
    body_file="$tmpdir/body"
    if ! curl -fsSL "$base" -o "$body_file" -w '%{http_code}' >/dev/null; then
        echo "  ERROR: failed to fetch $base"
        failures=$((failures + 1))
        continue
    fi

    if [[ -n "$anchor" ]]; then
        if ! grep -Eq "id=[\"']${anchor}[\"']|name=[\"']${anchor}[\"']" "$body_file"; then
            echo "  ERROR: anchor #$anchor not found in $base"
            failures=$((failures + 1))
        fi
    fi
done

if [[ $failures -gt 0 ]]; then
    echo "Completed with $failures failures"
    exit 1
fi

echo "All URLs reachable and anchors present where specified."

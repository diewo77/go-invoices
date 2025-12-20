#!/usr/bin/env sh
set -eu

INPUT=./assets/css/input.css
OUTDIR=./static
OUTFILE=tailwind.css

if [ ! -f package.json ]; then
  echo "No package.json found; skipping CSS build" >&2
  exit 0
fi

# Warn if DaisyUI is referenced in tailwind.config.js but not installed.
if grep -q 'daisyui' tailwind.config.js 2>/dev/null; then
  if [ ! -d node_modules/daisyui ]; then
    echo "WARNING: DaisyUI plugin referenced but not installed (node_modules/daisyui missing). Run 'npm install' to include component styles." >&2
  fi
fi

echo "Building Tailwind CSS..." >&2
npx tailwindcss -i "$INPUT" -o "$OUTDIR/$OUTFILE" ${TAILWIND_EXTRA_FLAGS:-}

if [ ! -f "$OUTDIR/$OUTFILE" ]; then
  echo "Build failed: $OUTDIR/$OUTFILE missing" >&2
  exit 1
fi

HASH=$(sha1sum "$OUTDIR/$OUTFILE" | cut -c1-16 || shasum "$OUTDIR/$OUTFILE" | cut -c1-16)
HASHED="tailwind.$HASH.css"
cp "$OUTDIR/$OUTFILE" "$OUTDIR/$HASHED"

cat > "$OUTDIR/manifest.json" <<EOF
{
  "tailwind.css": "$HASHED"
}
EOF

echo "Generated manifest with $HASHED" >&2

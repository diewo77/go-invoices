#!/usr/bin/env sh
set -eu
[ -d node_modules ] || npm install --no-audit --no-fund
css_watch() {
  echo "[css-watch] initial build"; ./build-assets.sh || true
  if command -v inotifywait >/dev/null 2>&1; then
    while inotifywait -e modify,create,delete -r assets/css tailwind.config.js postcss.config.js >/dev/null 2>&1; do
      echo "[css-watch] change detected"; ./build-assets.sh || true
    done
  else
    echo "[css-watch] using polling (inotifywait not found)"
    while sleep 2; do
      changed=$(find assets/css -type f -newer static/manifest.json -print -quit 2>/dev/null || true)
      cfg_changed=$(find . -maxdepth 1 -name tailwind.config.js -o -name postcss.config.js -newer static/manifest.json -print -quit 2>/dev/null || true)
      if [ -n "$changed$cfg_changed" ]; then
        echo "[css-watch] polling rebuild"; ./build-assets.sh || true
      fi
    done
  fi
}
css_watch &

# Optionally run SQL migrations (golang-migrate) on startup when MIGRATIONS=1
if [ "${MIGRATIONS:-}" = "1" ] || [ "${MIGRATIONS:-}" = "true" ] || [ "${MIGRATIONS:-}" = "yes" ]; then
  echo "[dev] Running SQL migrations at startup..."
  # The app runs migrations internally too when MIGRATIONS=1, but this early run helps fail fast.
  go run ./cmd/server --migrate-only || true
fi

reflex -r "(cmd|internal|templates)/.*\.(go|html)$" -s -- go run ./cmd/server

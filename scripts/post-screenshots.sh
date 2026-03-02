#!/usr/bin/env bash
# Post visual review screenshots to #product Discord channel
set -euo pipefail

SCREENSHOT_DIR="${1:-e2e/screenshots}"
CHANNEL="${DISCORD_CHANNEL:-1477475122490245193}"

if [ ! -d "$SCREENSHOT_DIR" ]; then
  echo "Error: screenshot directory not found: $SCREENSHOT_DIR"
  exit 1
fi

count=0
for img in "$SCREENSHOT_DIR"/*.png; do
  [ -f "$img" ] || continue
  name=$(basename "$img" .png)
  openclaw message send \
    --channel discord \
    --target "$CHANNEL" \
    --media "$img" \
    --message "📸 ${name} page"
  count=$((count + 1))
done

if [ "$count" -eq 0 ]; then
  echo "No screenshots found in $SCREENSHOT_DIR"
  exit 1
fi

echo "✓ Posted $count screenshot(s) to Discord"

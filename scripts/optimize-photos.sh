#!/bin/bash
set -euo pipefail

# Generates web-optimized WebP images from original photo exports.
# Usage: ./scripts/optimize-photos.sh photos/2025-sicily
#        ./scripts/optimize-photos.sh photos/2025-sicily --html

ALBUM_DIR="${1:?Usage: optimize-photos.sh <album-directory> [--html]}"
GENERATE_HTML="${2:-}"

THUMB_SIZE=600
FULL_SIZE=2000
THUMB_QUALITY=80
FULL_QUALITY=85

THUMB_DIR="$ALBUM_DIR/web/thumb"
FULL_DIR="$ALBUM_DIR/web/full"
mkdir -p "$THUMB_DIR" "$FULL_DIR"

# Derive the album URL path from the directory name
ALBUM_SLUG="$(basename "$ALBUM_DIR")"

counter=1
for src in "$ALBUM_DIR"/*.jpeg "$ALBUM_DIR"/*.JPEG "$ALBUM_DIR"/*.png "$ALBUM_DIR"/*.PNG; do
    [ -f "$src" ] || continue
    padded=$(printf "%03d" "$counter")
    echo "[$padded] $src"

    # Thumbnail (600px longest edge)
    magick "$src" -resize "${THUMB_SIZE}x${THUMB_SIZE}>" \
        -quality "$THUMB_QUALITY" "$THUMB_DIR/${padded}.webp"

    # Full size (2000px longest edge)
    magick "$src" -resize "${FULL_SIZE}x${FULL_SIZE}>" \
        -quality "$FULL_QUALITY" "$FULL_DIR/${padded}.webp"

    counter=$((counter + 1))
done

total=$((counter - 1))
echo "Done. Processed $total images."

# Optionally generate HTML fragment for the album page
if [ "$GENERATE_HTML" = "--html" ]; then
    HTML_FILE="$ALBUM_DIR/fragment.html"
    echo "Generating HTML fragment -> $HTML_FILE"
    > "$HTML_FILE"
    for i in $(seq 1 "$total"); do
        padded=$(printf "%03d" "$i")
        idx=$((i - 1))
        cat >> "$HTML_FILE" <<EOF
                            <figure class="photo-item" role="listitem">
                                <a href="/photos/${ALBUM_SLUG}/web/full/${padded}.webp" class="photo-link" data-index="${idx}">
                                    <img src="/photos/${ALBUM_SLUG}/web/thumb/${padded}.webp" alt="" loading="lazy" class="photo-thumb" />
                                </a>
                            </figure>
EOF
    done
    echo "HTML fragment written ($total figures)."
fi

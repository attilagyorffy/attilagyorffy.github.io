#!/bin/bash
# Generate OG social sharing images for all pages.
# Uses headless Chrome to screenshot HTML templates at 1200x630.
# Run from the repository root: bash tools/generate-og-images.sh

set -euo pipefail

CHROME="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
SITE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PORT=8791
TEMPLATE="$SITE_DIR/tools/og-template.html"
TEMPLATE_DEFAULT="$SITE_DIR/tools/og-template-default.html"
TMP_HTML="/tmp/og-card-render.html"

# Start local server for font loading
python3 -m http.server "$PORT" -d "$SITE_DIR" &>/dev/null &
SERVER_PID=$!
trap "kill $SERVER_PID 2>/dev/null" EXIT
sleep 0.5

screenshot() {
    local url="$1"
    local output="$2"
    "$CHROME" --headless=new \
        --screenshot="$output" \
        --window-size=1200,630 \
        --hide-scrollbars \
        --disable-gpu \
        --no-sandbox \
        "$url" 2>/dev/null
}

escape_for_sed() {
    printf '%s' "$1" | sed 's/[&/\]/\\&/g'
}

generate_post_image() {
    local slug="$1"
    local title="$2"
    local subtitle="$3"
    local output="$SITE_DIR/blog/$slug/og.png"

    local escaped_title escaped_subtitle
    escaped_title=$(escape_for_sed "$title")
    escaped_subtitle=$(escape_for_sed "$subtitle")

    sed "s|TAGLINE_PLACEHOLDER|$escaped_subtitle|g; s|TITLE_PLACEHOLDER|$escaped_title|g" "$TEMPLATE" > "$TMP_HTML"

    cp "$TMP_HTML" "$SITE_DIR/_og-render.html"
    screenshot "http://localhost:$PORT/_og-render.html" "$output"
    echo "  Generated: blog/$slug/og.png"
}

echo "Generating OG images..."
echo ""

# Default image
cp "$TEMPLATE_DEFAULT" "$SITE_DIR/_og-render.html"
screenshot "http://localhost:$PORT/_og-render.html" "$SITE_DIR/images/og-default.png"
echo "  Generated: images/og-default.png"

# Blog posts
generate_post_image "open-source-is-just-plumbing-now" \
    "Open Source Is Just Plumbing Now" \
    "Nobody cares about plumbing until shit comes up through the floor."

generate_post_image "stop-naming-go-constructors-new" \
    "Stop Naming Your Go Constructors New" \
    "A grown man loses it over three letters in a function name."

generate_post_image "the-day-viktor-orban-messaged-me" \
    "The Day &#x201c;Viktor Orbán&#x201d; Messaged Me" \
    "The marketing is automated. The GDPR violations are artisanal."

generate_post_image "when-one-company-breaks-half-the-internet-goes-with-it" \
    "When One Company Breaks, Half the Internet Goes with It" \
    "Cloudflare sneezed and half the internet called in sick."

generate_post_image "the-quiet-revolt-of-the-digital-self" \
    "The Quiet Revolt of the Digital Self" \
    "One billionaire bought a platform and we all remembered we own nothing."

generate_post_image "stop-shipping-desktop-images-to-phones" \
    "Stop Shipping Desktop Images to Phones" \
    "Sending a 4000-pixel hero to a phone on 3G is not a personality trait."

generate_post_image "ansible-hangs-on-freebsd-and-nobody-tells-you" \
    "Ansible Hangs on FreeBSD and Nobody Tells You" \
    "One environment variable. That&#x2019;s the whole fix. You&#x2019;re welcome."

generate_post_image "stop-typing-full-git-urls" \
    "Stop Typing Full Git URLs" \
    "Three lines in .gitconfig and you can stop living like a caveman."

generate_post_image "replacing-openssl-with-libressl-on-freebsd" \
    "Replacing OpenSSL with LibreSSL on FreeBSD" \
    "OpenSSL hid its own bugs from the tools designed to catch them."

generate_post_image "postgres-refuses-your-not-null-migration" \
    "Postgres Refuses Your NOT NULL Migration" \
    "Postgres isn&#x2019;t wrong. You are."

generate_post_image "every-ember-phoenix-tutorial-was-wrong" \
    "Every Ember-Phoenix Tutorial Was Wrong" \
    "Every tutorial was either broken or lying. So here&#x2019;s one that isn&#x2019;t."

generate_post_image "npm-progress-bar-halves-your-install-speed" \
    "npm&#x2019;s Progress Bar Halves Your Install Speed" \
    "A progress bar that makes things slower. Peak JavaScript."

generate_post_image "your-freebsd-files-are-readable-by-everyone" \
    "Your FreeBSD Files Are Readable by Everyone" \
    "Your API keys are world-readable by default. Sleep well."

generate_post_image "run-the-w3c-validator-locally-with-docker" \
    "Run the W3C Validator Locally with Docker" \
    "The validator needs the internet. The internet is the problem."

generate_post_image "stop-naming-all-your-callbacks-cb" \
    "Stop Naming All Your Callbacks cb" \
    "Everyone does it. Everyone spends forty minutes debugging. Nobody learns."

generate_post_image "testing-activemodel-serializer-with-rspec" \
    "Testing ActiveModel::Serializer with RSpec" \
    "They rewrote the gem and broke how you test it. Nobody warned you."

generate_post_image "zero-config-mobile-testing-with-pow-and-bonjour" \
    "Zero-Config Mobile Testing with Pow and Bonjour" \
    "I made mDNS do my job because looking up IP addresses is beneath me."

generate_post_image "debugging-ember-apps-in-end-to-end-tests" \
    "Debugging Ember Apps in End-to-End Tests" \
    "Your test fails. The browser closes. All you&#x2019;ve got is &#x201c;expected true, got false.&#x201d;"

generate_post_image "why-your-ember-promises-fail-silently" \
    "Why Your Ember Promises Fail Silently" \
    "Something breaks. Console&#x2019;s clean. Welcome to promise hell."

generate_post_image "when-brew-install-node-required-5-5gb-of-xcode" \
    "When brew install node required 5.5GB of Xcode" \
    "Apple broke the compiler. Installing JavaScript now required an entire IDE."

# Clean up temp file
rm -f "$SITE_DIR/_og-render.html"

echo ""
echo "Done. Generated 21 OG images."

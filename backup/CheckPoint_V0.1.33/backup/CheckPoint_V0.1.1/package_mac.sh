#!/bin/bash

# Configuration
APP_NAME="CheckPoint"
VERSION=$1
BUNDLE_ID="io.samphoerna.github.checkpoint"
VOL_NAME="CheckPoint – Secure Device Checker"
OUTPUT_DMG="CheckPoint_${VERSION}.dmg"

if [ -z "$VERSION" ]; then
    echo "Usage: ./package_mac.sh <version>"
    echo "Example: ./package_mac.sh v0.1.1"
    exit 1
fi

export PATH=$PATH:$(go env GOPATH)/bin:$HOME/go/bin

echo "🚀 Packaging $APP_NAME for macOS ($VERSION)..."

# 1. Update Icon (Just to be sure)
cp frontend/src/assets/logo.png build/appicon.png

# 2. Build via Wails (Production, Universal)
# Note: universal might require both toolchains. 
# If failures occur, we default to host arch or specify explicit -platform darwin/arm64
echo "🔨 Building Application Bundle..."
wails build -platform darwin/universal -ldflags "-X main.AppVersion=$VERSION" -clean

if [ $? -ne 0 ]; then
    echo "❌ Wails build failed!"
    exit 1
fi

# 3. Prepare DMG Staging
echo "📂 Preparing DMG folder..."
rm -rf dist_dmg
mkdir dist_dmg

# Copy App
cp -R "build/bin/$APP_NAME.app" dist_dmg/

# Create /Applications Symlink
ln -s /Applications dist_dmg/Applications

# 4. Create DMG using hdiutil
echo "📦 Creating .dmg..."
rm -f "$OUTPUT_DMG"

hdiutil create -volname "$VOL_NAME" \
    -srcfolder dist_dmg \
    -ov -format UDZO \
    "$OUTPUT_DMG"

if [ $? -eq 0 ]; then
    echo "✅ Success! Installer created:"
    echo "   Running: open $OUTPUT_DMG"
else
    echo "❌ DMG creation failed."
    exit 1
fi

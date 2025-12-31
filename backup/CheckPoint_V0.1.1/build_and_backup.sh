#!/bin/bash

# Configuration
APP_NAME="Checkpoint"
BACKUP_DIR="backups"
VERSION_TAG=$1

if [ -z "$VERSION_TAG" ]; then
    echo "Usage: ./build_and_backup.sh <version_tag>"
    echo "Example: ./build_and_backup.sh v0.1.0"
    exit 1
fi

echo "🚀 Building $APP_NAME version $VERSION_TAG..."

# Ensure backup directory exists
mkdir -p "$BACKUP_DIR"

# 1. Build with Version Injection
# -ldflags "-X main.AppVersion=..." injects the variable in main.go
wails build -ldflags "-X main.AppVersion=$VERSION_TAG"

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"

# 2. Create Backup Archive
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
ARCHIVE_NAME="${APP_NAME}_${VERSION_TAG}_${TIMESTAMP}.zip"
APP_PATH="build/bin/$APP_NAME.app" 

if [ ! -d "$APP_PATH" ]; then
    echo "⚠️  App bundle not found at $APP_PATH. Checking for binary..."
    # Fallback for non-bundle builds if any
fi

echo "📦 Creating backup archive: $BACKUP_DIR/$ARCHIVE_NAME..."
zip -r "$BACKUP_DIR/$ARCHIVE_NAME" "$APP_PATH"

echo "🎉 Done! Version $VERSION_TAG is built and backed up."
echo "   Release it with: git tag -a $VERSION_TAG -m \"Release $VERSION_TAG\" && git push origin $VERSION_TAG"

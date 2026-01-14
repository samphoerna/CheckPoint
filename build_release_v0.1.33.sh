#!/bin/bash
set -e

VERSION="V0.1.33"
APP_NAME="CheckPoint"

echo "🚀 Starting Release Build for $VERSION..."

# Ensure we are in the project root
export PATH=$PATH:$(go env GOPATH)/bin

# 1. Windows Build
echo "🪟 Building for Windows (Portable EXE)..."
wails build -platform windows/amd64 -ldflags "-X main.AppVersion=$VERSION" -clean
mv "build/bin/${APP_NAME}.exe" "build/bin/${APP_NAME}_${VERSION}.exe"
echo "✅ Windows build available at: build/bin/${APP_NAME}_${VERSION}.exe"

# 2. macOS Build
echo "🍎 Building for macOS (App Bundle)..."
wails build -platform darwin/universal -ldflags "-X main.AppVersion=$VERSION"
mv "build/bin/${APP_NAME}.app" "build/bin/${APP_NAME}_${VERSION}.app"
echo "✅ macOS build available at: build/bin/${APP_NAME}_${VERSION}.app"

echo "🎉 Build Process Complete!"

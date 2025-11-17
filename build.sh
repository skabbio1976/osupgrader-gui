#!/bin/bash
# Build script for osupgrader-gui

set -e

echo "Building osupgrader-gui..."

# Build for Linux
echo "Building for Linux..."
go build -ldflags="-s -w" -o osupgrader-gui ./cmd/osupgrader-gui
echo "✓ Linux binary: osupgrader-gui ($(du -h osupgrader-gui | cut -f1))"

# Build for Windows using fyne-cross
echo "Building for Windows (using fyne-cross with Docker)..."
~/go/bin/fyne-cross windows -arch=amd64 -app-id com.example.osupgrader ./cmd/osupgrader-gui

# Extract Windows executable
cd fyne-cross/dist/windows-amd64
unzip -o "OS Upgrader GUI.exe.zip"
cd ../../..

# Copy to root and rename
cp "fyne-cross/dist/windows-amd64/OS Upgrader GUI.exe" osupgrader-gui.exe
echo "✓ Windows binary: osupgrader-gui.exe ($(du -h osupgrader-gui.exe | cut -f1))"

echo ""
echo "Build complete!"
echo "  Linux:   ./osupgrader-gui"
echo "  Windows: ./osupgrader-gui.exe"

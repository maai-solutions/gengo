#!/bin/bash
# Build script for GenGo with Bubble Tea, Cobra, and Viper
# This script sets the correct GOROOT for the Homebrew Go installation

GOROOT="/home/linuxbrew/.linuxbrew/Cellar/go/1.24.5/libexec"
export GOROOT

echo "Building GenGo with GOROOT=$GOROOT"
echo "Dependencies: Bubble Tea, Cobra, Viper"

# Check if first argument is "test"
if [ "$1" = "test" ]; then
    echo "Running tests..."
    shift  # Remove "test" from arguments
    go test "$@"
    exit $?
fi

# Normal build
go build -o gengo

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Usage:"
    echo "  ./gengo           - Start Bubble Tea interactive CLI"
    echo "  ./gengo version   - Show version (0.0.0)"
    echo "  ./gengo --help    - Show help"
else
    echo "Build failed!"
    exit 1
fi

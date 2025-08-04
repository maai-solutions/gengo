#!/bin/bash
# Build script for GenGo with Bubble Tea, Cobra, and Viper
# This script sets the correct GOROOT for the Homebrew Go installation

# Detect platform and set GOROOT accordingly
if command -v brew >/dev/null 2>&1; then
    # Homebrew is available, get the Go prefix
    BREW_GO_PREFIX=$(brew --prefix go 2>/dev/null)
    if [ -n "$BREW_GO_PREFIX" ] && [ -d "$BREW_GO_PREFIX/libexec" ]; then
        GOROOT="$BREW_GO_PREFIX/libexec"
    else
        echo "Warning: Homebrew Go installation not found, using system Go"
        unset GOROOT
    fi
else
    echo "Warning: Homebrew not found, using system Go"
    unset GOROOT
fi

if [ -n "$GOROOT" ]; then
    export GOROOT
    echo "Building GenGo with GOROOT=$GOROOT"
else
    echo "Building GenGo with system Go"
fi

# Check for whisper.cpp dependency
if command -v brew >/dev/null 2>&1; then
    # Check if whisper.cpp source is available for Go bindings
    WHISPER_SOURCE_DIR="/tmp/whisper.cpp"
    
    if [ ! -d "$WHISPER_SOURCE_DIR" ]; then
        echo "Cloning whisper.cpp source for Go bindings..."
        git clone https://github.com/ggerganov/whisper.cpp.git "$WHISPER_SOURCE_DIR"
    fi
    
    if [ -d "$WHISPER_SOURCE_DIR" ]; then
        echo "Building whisper.cpp from source..."
        cd "$WHISPER_SOURCE_DIR"
        make -j$(nproc)
        
        # Set CGO flags to use the source build
        export CGO_CPPFLAGS="-I${WHISPER_SOURCE_DIR}/include -I${WHISPER_SOURCE_DIR}/ggml/include"
        export CGO_LDFLAGS="-L${WHISPER_SOURCE_DIR}/build/src -L${WHISPER_SOURCE_DIR}/build/ggml/src -lwhisper -lggml -lggml-base -lggml-cpu"
        export LD_LIBRARY_PATH="${WHISPER_SOURCE_DIR}/build/src:${WHISPER_SOURCE_DIR}/build/ggml/src:${LD_LIBRARY_PATH}"
        export CGO_ENABLED=1
        
        cd /home/udg/projects/git/gengo
    else
        echo "Warning: Could not set up whisper.cpp source. ASR functionality may not work."
    fi
fi
echo "Dependencies: Bubble Tea, Cobra, Viper"

# Check if first argument is "test"
if [ "$1" = "test" ]; then
    echo "Running tests..."
    shift  # Remove "test" from arguments
    go test "$@"
    exit $?
fi

# Normal build
if [ -n "$LD_LIBRARY_PATH" ]; then
    export LD_LIBRARY_PATH
fi
go build -o gengo

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Usage:"
    echo "  ./run_gengo.sh        - Start Bubble Tea interactive CLI (recommended)"
    echo "  ./gengo version       - Show version (0.0.0)"
    echo "  ./gengo --help        - Show help"
    echo ""
    echo "Note: Use run_gengo.sh wrapper script to automatically set library paths"
    echo "Or manually set LD_LIBRARY_PATH before running ./gengo:"
    echo "  export LD_LIBRARY_PATH=\"/tmp/whisper.cpp/build/src:/tmp/whisper.cpp/build/ggml/src:\${LD_LIBRARY_PATH}\""
else
    echo "Build failed!"
    exit 1
fi

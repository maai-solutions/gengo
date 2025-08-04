#!/bin/bash
# Wrapper script to run GenGo with proper library paths

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
WHISPER_LIB_PATH="/tmp/whisper.cpp/build/src:/tmp/whisper.cpp/build/ggml/src"

# Set library path and run gengo
export LD_LIBRARY_PATH="${WHISPER_LIB_PATH}:${LD_LIBRARY_PATH}"
exec "${SCRIPT_DIR}/gengo" "$@"

#!/bin/bash

# Test script for Bubble Tea interactive mode
echo "Testing Bubble Tea interactive mode..."
echo ""

# Create a temporary expect-like script using here document
echo "Starting interactive mode test..."

# Use printf to send keystrokes to the application
# First type "hello", then press enter, then type "/exit" and press enter
(
    sleep 1
    printf "hello\n"
    sleep 1
    printf "/exit\n"
    sleep 1
) | ./gengo

package cmd

import (
	"testing"
)

func TestIsValidYouTubeURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"https://youtu.be/dQw4w9WgXcQ", true},
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"https://m.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"https://youtube.com/embed/dQw4w9WgXcQ", true},
		{"https://youtube.com/v/dQw4w9WgXcQ", true},
		{"invalid-url", false},
		{"https://example.com", false},
		{"", false},
	}

	for _, test := range tests {
		result := isValidYouTubeURL(test.url)
		if result != test.expected {
			t.Errorf("isValidYouTubeURL(%q) = %v, expected %v", test.url, result, test.expected)
		}
	}
}

func TestExtractVideoID(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"https://youtube.com/watch?v=dQw4w9WgXcQ&t=30s", "dQw4w9WgXcQ"},
		{"https://youtu.be/dQw4w9WgXcQ?t=30", "dQw4w9WgXcQ"},
		{"invalid-url", ""},
		{"", ""},
	}

	for _, test := range tests {
		result := extractVideoID(test.url)
		if result != test.expected {
			t.Errorf("extractVideoID(%q) = %q, expected %q", test.url, result, test.expected)
		}
	}
}

func TestGenerateTranscriptFilename(t *testing.T) {
	// Test with valid YouTube URL
	filename := generateTranscriptFilename("https://youtube.com/watch?v=dQw4w9WgXcQ")
	if !contains(filename, "dQw4w9WgXcQ") {
		t.Errorf("Expected filename to contain video ID, got: %s", filename)
	}
	if !contains(filename, ".md") {
		t.Errorf("Expected filename to have .md extension, got: %s", filename)
	}

	// Test with invalid URL
	filename = generateTranscriptFilename("invalid-url")
	if !contains(filename, "transcript") {
		t.Errorf("Expected filename to contain 'transcript' for invalid URL, got: %s", filename)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "llo wo", true},
		{"hello world", "xyz", false},
		{"", "test", false},
		{"test", "", true}, // Empty substring should match
		{"test", "test", true},
		{"test", "testing", false},
	}

	for _, test := range tests {
		result := contains(test.s, test.substr)
		if result != test.expected {
			t.Errorf("contains(%q, %q) = %v, expected %v", test.s, test.substr, result, test.expected)
		}
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected int
	}{
		{"hello world", "world", 6},
		{"hello world", "hello", 0},
		{"hello world", "llo", 2},
		{"hello world", "xyz", -1},
		{"", "test", -1},
		{"test", "test", 0},
		{"testing", "test", 0},
	}

	for _, test := range tests {
		result := indexOf(test.s, test.substr)
		if result != test.expected {
			t.Errorf("indexOf(%q, %q) = %d, expected %d", test.s, test.substr, result, test.expected)
		}
	}
}

package pdf

import (
	"fmt"
	"strings"
	"testing"
)

// Mock test to verify the structure works
func TestTextExtractorCreation(t *testing.T) {
	extractor := NewTextExtractor()
	if extractor == nil {
		t.Error("Expected non-nil extractor")
	}
	if extractor.Config == nil {
		t.Error("Expected non-nil config")
	}
}

func TestCleanText(t *testing.T) {
	extractor := NewTextExtractor()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "   Some   text   with   spaces   \n\n\n   More text   \n",
			expected: "Some   text   with   spaces\nMore text",
		},
		{
			input:    "\n\n\nHello\n\n\nWorld\n\n\n",
			expected: "Hello\nWorld",
		},
		{
			input:    "",
			expected: "",
		},
		{
			input:    "   \n   \n   ",
			expected: "",
		},
	}

	for i, test := range tests {
		result := extractor.CleanText(test.input)
		if result != test.expected {
			t.Errorf("Test %d failed. Expected %q, got %q", i+1, test.expected, result)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	extractor := NewTextExtractor()

	// Test with non-existent file
	_, err := extractor.ExtractFromFile("non-existent-file.pdf")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "file does not exist") {
		t.Errorf("Expected 'file does not exist' error, got: %s", err.Error())
	}

	// Test with empty byte array
	_, err = extractor.ExtractFromBytes([]byte{})
	if err == nil {
		t.Error("Expected error for empty byte array")
	}
	if !strings.Contains(err.Error(), "empty byte array") {
		t.Errorf("Expected 'empty byte array' error, got: %s", err.Error())
	}

	// Test with nil reader
	_, err = extractor.ExtractFromReader(nil)
	if err == nil {
		t.Error("Expected error for nil reader")
	}
	if !strings.Contains(err.Error(), "nil reader") {
		t.Errorf("Expected 'nil reader' error, got: %s", err.Error())
	}

	// Test with nil writer
	err = extractor.ExtractFromFileToWriter("test.pdf", nil)
	if err == nil {
		t.Error("Expected error for nil writer")
	}
	if !strings.Contains(err.Error(), "nil writer") {
		t.Errorf("Expected 'nil writer' error, got: %s", err.Error())
	}
}

// Example usage that can be run manually
func ExampleTextExtractor() {
	// Create a new text extractor
	extractor := NewTextExtractor()

	// Example: Clean extracted text
	dirtyText := "   Some   text   with   lots   of   spaces   \n\n\n   More text   \n"
	cleanText := extractor.CleanText(dirtyText)
	fmt.Println("Cleaned text:", cleanText)
	// Output: Cleaned text: Some   text   with   lots   of   spaces
	// More text
}

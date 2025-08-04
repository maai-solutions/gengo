package extractors

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewContentExtractor(t *testing.T) {
	extractor := NewContentExtractor()

	if extractor == nil {
		t.Error("Expected non-nil extractor")
	}

	if extractor.skipTags == nil {
		t.Error("Expected non-nil skipTags map")
	}

	// Check that skip tags are properly initialized
	expectedSkipTags := []string{"script", "style", "nav", "header", "footer", "aside"}
	for _, tag := range expectedSkipTags {
		if !extractor.skipTags[tag] {
			t.Errorf("Expected skip tag %s to be true", tag)
		}
	}

	if extractor.inSkip == nil {
		t.Error("Expected non-nil inSkip map")
	}
}

func TestIsContentTag(t *testing.T) {
	tests := []struct {
		tag      string
		expected bool
	}{
		{"p", true},
		{"h1", true},
		{"h2", true},
		{"h3", true},
		{"h4", true},
		{"h5", true},
		{"h6", true},
		{"article", true},
		{"section", true},
		{"main", true},
		{"div", false},
		{"span", false},
		{"script", false},
		{"style", false},
	}

	for _, test := range tests {
		result := isContentTag(test.tag)
		if result != test.expected {
			t.Errorf("isContentTag(%s) = %v, expected %v", test.tag, result, test.expected)
		}
	}
}

func TestIsHeaderTag(t *testing.T) {
	tests := []struct {
		tag      string
		expected bool
	}{
		{"h1", true},
		{"h2", true},
		{"h3", true},
		{"h4", true},
		{"h5", true},
		{"h6", true},
		{"h7", false}, // Not a valid header
		{"p", false},
		{"header", false}, // Not a header tag (h1-h6)
		{"", false},
	}

	for _, test := range tests {
		result := isHeaderTag(test.tag)
		if result != test.expected {
			t.Errorf("isHeaderTag(%s) = %v, expected %v", test.tag, result, test.expected)
		}
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "Hello World"},
		{"File<Name>", "File-Name-"},
		{"Test:File", "Test-File"},
		{"Path/To/File", "Path-To-File"},
		{"File|Name", "File-Name"},
		{"Question?Mark", "Question-Mark"},
		{"Asterisk*File", "Asterisk-File"},
		{"   Spaced   ", "Spaced"},
		{"", ""},
	}

	for _, test := range tests {
		result := sanitizeFilename(test.input)
		if result != test.expected {
			t.Errorf("sanitizeFilename(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestExtractFromHTML(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		url           string
		expectTitle   string
		expectContent bool // Whether content should contain certain elements
	}{
		{
			name: "Simple HTML with title and paragraph",
			html: `
<!DOCTYPE html>
<html>
<head>
    <title>Test Page Title</title>
</head>
<body>
    <h1>Main Heading</h1>
    <p>This is a test paragraph.</p>
</body>
</html>`,
			url:           "https://example.com",
			expectTitle:   "Test Page Title",
			expectContent: true,
		},
		{
			name: "HTML without title",
			html: `
<!DOCTYPE html>
<html>
<body>
    <p>Content without title</p>
</body>
</html>`,
			url:           "https://example.com",
			expectTitle:   "Untitled",
			expectContent: true,
		},
		{
			name: "HTML with skip tags",
			html: `
<!DOCTYPE html>
<html>
<head>
    <title>Page with Scripts</title>
</head>
<body>
    <script>console.log("should be skipped");</script>
    <p>Visible content</p>
    <style>.hidden { display: none; }</style>
</body>
</html>`,
			url:           "https://example.com",
			expectTitle:   "Page with Scripts",
			expectContent: true,
		},
		{
			name: "HTML with headers",
			html: `
<!DOCTYPE html>
<html>
<head>
    <title>Headers Test</title>
</head>
<body>
    <h1>Header 1</h1>
    <h2>Header 2</h2>
    <p>Regular paragraph</p>
</body>
</html>`,
			url:           "https://example.com",
			expectTitle:   "Headers Test",
			expectContent: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			title, content := ExtractFromHTML(test.html, test.url)

			if title != test.expectTitle {
				t.Errorf("Expected title %q, got %q", test.expectTitle, title)
			}

			if test.expectContent {
				if content == "" {
					t.Error("Expected non-empty content")
				}
				if !strings.Contains(content, test.url) {
					t.Error("Expected content to contain source URL")
				}
				if !strings.Contains(content, "Source:") {
					t.Error("Expected content to contain 'Source:' reference")
				}
			}
		})
	}
}

func TestDownloadAndExtract(t *testing.T) {
	// Create a test server
	testHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>Test Server Page</title>
</head>
<body>
    <h1>Welcome to Test Server</h1>
    <p>This is a test page served by httptest.</p>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHTML))
	}))
	defer server.Close()

	title, content, err := DownloadAndExtract(server.URL)

	if err != nil {
		t.Fatalf("DownloadAndExtract failed: %v", err)
	}

	if title != "Test Server Page" {
		t.Errorf("Expected title 'Test Server Page', got %q", title)
	}

	if content == "" {
		t.Error("Expected non-empty content")
	}

	if !strings.Contains(content, server.URL) {
		t.Error("Expected content to contain source URL")
	}

	if !strings.Contains(content, "Welcome to Test Server") {
		t.Error("Expected content to contain extracted text")
	}
}

func TestDownloadAndExtractInvalidURL(t *testing.T) {
	_, _, err := DownloadAndExtract("http://invalid-url-that-should-not-exist.local")

	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	if !strings.Contains(err.Error(), "failed to fetch URL") {
		t.Errorf("Expected 'failed to fetch URL' error, got: %v", err)
	}
}

func TestSaveToProject(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	// Change to temp directory
	os.Chdir(tempDir)

	testTitle := "Test Document"
	testContent := "# Test Document\n\nThis is test content."
	testProject := "test-project"

	err := SaveToProject(testTitle, testContent, testProject)
	if err != nil {
		t.Fatalf("SaveToProject failed: %v", err)
	}

	// Check if project directory was created
	projectPath := filepath.Join(tempDir, testProject)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Error("Project directory was not created")
	}

	// Check if file was created with correct content
	expectedFile := filepath.Join(projectPath, "Test Document.md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Error("Markdown file was not created")
	}

	// Read and verify file content
	savedContent, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(savedContent) != testContent {
		t.Errorf("File content mismatch. Expected %q, got %q", testContent, string(savedContent))
	}
}

func TestSaveToProjectInvalidPath(t *testing.T) {
	// Try to save to a path that should fail (using invalid characters in different OS)
	err := SaveToProject("test", "content", "/invalid\x00path")

	if err == nil {
		t.Error("Expected error for invalid project path")
	}
}

func TestContentExtractorHandleData(t *testing.T) {
	extractor := NewContentExtractor()

	// Test empty/whitespace data
	extractor.handleData("")
	extractor.handleData("   ")
	extractor.handleData("\n\t")

	if len(extractor.Content) > 0 || extractor.Title != "" {
		t.Error("Expected no content for empty/whitespace data")
	}

	// Test title handling
	extractor.inTitle = true
	extractor.handleData("Test Title")
	if extractor.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got %q", extractor.Title)
	}

	// Test body content handling
	extractor.inTitle = false
	extractor.inBody = true
	extractor.currTag = "p"
	extractor.handleData("Test content")

	if len(extractor.Content) == 0 {
		t.Error("Expected content to be added")
	}

	// Test header handling
	extractor.currTag = "h1"
	extractor.handleData("Header Text")

	found := false
	for _, content := range extractor.Content {
		if strings.Contains(content, "# Header Text") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected header formatting in content")
	}
}

func TestContentExtractorIsInAnySkipTag(t *testing.T) {
	extractor := NewContentExtractor()

	// Initially should not be in any skip tag
	if extractor.isInAnySkipTag() {
		t.Error("Expected not to be in any skip tag initially")
	}

	// Set one skip tag to true
	extractor.inSkip["script"] = true
	if !extractor.isInAnySkipTag() {
		t.Error("Expected to be in a skip tag")
	}

	// Set it back to false
	extractor.inSkip["script"] = false
	if extractor.isInAnySkipTag() {
		t.Error("Expected not to be in any skip tag after resetting")
	}
}

// Example usage that can be run manually
func ExampleExtractFromHTML() {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Example Page</title>
</head>
<body>
    <h1>Welcome</h1>
    <p>This is an example page.</p>
</body>
</html>`

	title, content := ExtractFromHTML(html, "https://example.com")

	// Print title (sanitized for filename use)
	_ = title // Example Page

	// Content will include markdown formatting
	_ = content
	// Output format: markdown with title, source, and extracted content
}

func ExampleDownloadAndExtract() {
	// This example would work with a real URL
	// title, content, err := DownloadAndExtract("https://example.com")
	// if err != nil {
	//     log.Fatal(err)
	// }
	// fmt.Printf("Title: %s\nContent: %s\n", title, content)
}

func ExampleSaveToProject() {
	title := "My Document"
	content := "# My Document\n\nThis is the content."
	projectName := "my-project"

	err := SaveToProject(title, content, projectName)
	if err != nil {
		// handle error
		_ = err
	}
	// Creates: ./my-project/My Document.md
}

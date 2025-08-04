package extractor

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type ContentExtractor struct {
	Title     string
	Content   []string
	inTitle   bool
	inBody    bool
	inSkip    map[string]bool
	currTag   string
	skipTags  map[string]bool
}

func NewContentExtractor() *ContentExtractor {
	return &ContentExtractor{
		skipTags: map[string]bool{
			"script": true,
			"style":  true,
			"nav":    true,
			"header": true,
			"footer": true,
			"aside":  true,
		},
		inSkip: make(map[string]bool),
	}
}

func (ce *ContentExtractor) traverse(n *html.Node) {
	switch n.Type {
	case html.ElementNode:
		ce.currTag = n.Data
		if n.Data == "title" {
			ce.inTitle = true
		}
		if ce.skipTags[n.Data] {
			ce.inSkip[n.Data] = true
		}
		if isContentTag(n.Data) {
			ce.inBody = true
		}
	case html.TextNode:
		ce.handleData(n.Data)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ce.traverse(c)
	}

	// Handle end tags
	switch n.Type {
	case html.ElementNode:
		if n.Data == "title" {
			ce.inTitle = false
		}
		if ce.skipTags[n.Data] {
			ce.inSkip[n.Data] = false
		}
		if isContentTag(n.Data) {
			if ce.inBody {
				ce.Content = append(ce.Content, "\n")
			}
			ce.inBody = false
		}
	}
}

func (ce *ContentExtractor) handleData(data string) {
	cleaned := strings.TrimSpace(data)
	if cleaned == "" {
		return
	}

	if ce.inTitle {
		ce.Title += cleaned
	} else if ce.inBody && !ce.isInAnySkipTag() {
		if isHeaderTag(ce.currTag) {
			level := ce.currTag[1:] // h1, h2, etc.
			ce.Content = append(ce.Content, fmt.Sprintf("\n%s %s\n", strings.Repeat("#", int(level[0]-'0')), cleaned))
		} else {
			ce.Content = append(ce.Content, cleaned+" ")
		}
	}
}

func (ce *ContentExtractor) isInAnySkipTag() bool {
	for _, in := range ce.inSkip {
		if in {
			return true
		}
	}
	return false
}

func isContentTag(tag string) bool {
	switch tag {
	case "p", "h1", "h2", "h3", "h4", "h5", "h6", "article", "section", "main":
		return true
	default:
		return false
	}
}

func isHeaderTag(tag string) bool {
	return strings.HasPrefix(tag, "h") && len(tag) == 2 && tag[1] >= '1' && tag[1] <= '6'
}

func sanitizeFilename(name string) string {
	re := regexp.MustCompile(`[<>:"/\\|?*]`)
	return strings.TrimSpace(re.ReplaceAllString(name, "-"))
}

// ExtractFromHTML extracts content from HTML string
func ExtractFromHTML(htmlContent string, url string) (string, string) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", ""
	}

	parser := NewContentExtractor()
	parser.traverse(doc)

	title := parser.Title
	if title == "" {
		title = "Untitled"
	}
	sanitizedTitle := sanitizeFilename(title)

	content := strings.Join(parser.Content, "")
	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")

	markdown := fmt.Sprintf("# %s\n\nSource: %s\n\n---\n\n%s", title, url, content)

	return sanitizedTitle, markdown
}

// DownloadAndExtract downloads a webpage and extracts its content
func DownloadAndExtract(url string) (string, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response body: %v", err)
	}

	title, content := ExtractFromHTML(string(htmlContent), url)
	return title, content, nil
}

// SaveToProject saves content to a project folder structure
func SaveToProject(title, content, projectName string) error {
	projectDir := filepath.Join(".", projectName)
	
	// Create project directory if it doesn't exist
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %v", err)
	}

	// Create filename from title
	filename := fmt.Sprintf("%s.md", title)
	filepath := filepath.Join(projectDir, filename)

	// Write content to file
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

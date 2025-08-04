package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	extractors "maai.solutions/gengo/internal/extractors/web"
)

var (
	webOutputFile  string
	webOutputDir   string
	webProjectName string
	webVerbose     bool
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Extract content from web pages",
	Long: `Extract and save content from web pages in a clean, readable format.

The web extractor parses HTML content and extracts the main text content,
removing navigation, scripts, styles, and other non-content elements.

Examples:
  gengo web extract https://example.com                     # Extract to stdout
  gengo web extract https://example.com --output page.md    # Save to file
  gengo web extract https://example.com --project my-proj   # Save to project folder
  gengo web extract https://example.com --dir ./web-content # Save to custom directory`,
}

// webExtractCmd represents the extract subcommand
var webExtractCmd = &cobra.Command{
	Use:   "extract [url]",
	Short: "Extract content from a web page",
	Long: `Extract content from a web page and output as clean markdown.

The command downloads the webpage, parses the HTML, and extracts the main
content while removing navigation, advertisements, and other non-content elements.

The output includes:
- Page title as heading
- Source URL
- Clean text content formatted as markdown

Options:
- Save to specific file with --output
- Save to project folder with --project
- Save to custom directory with --dir
- Verbose output with --verbose`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]

		// Validate URL (basic check)
		if !isValidURL(url) {
			fmt.Printf("Error: Invalid URL: %s\n", url)
			fmt.Println("Please provide a valid URL (e.g., https://example.com)")
			os.Exit(1)
		}

		if webVerbose {
			fmt.Printf("Extracting content from: %s\n", url)
		}

		// Extract content from web page
		title, content, err := extractors.DownloadAndExtract(url)
		if err != nil {
			fmt.Printf("Error extracting content: %v\n", err)
			os.Exit(1)
		}

		if webVerbose {
			fmt.Printf("Page title: %s\n", title)
			fmt.Printf("Content length: %d characters\n", len(content))
		}

		// Handle output based on specified options
		if webProjectName != "" {
			// Save to project structure
			err := extractors.SaveToProject(title, content, webProjectName)
			if err != nil {
				fmt.Printf("Error saving to project: %v\n", err)
				os.Exit(1)
			}

			projectPath := filepath.Join(".", webProjectName, fmt.Sprintf("%s.md", title))
			fmt.Printf("✅ Content extracted and saved to project!\n")
			fmt.Printf("File: %s\n", projectPath)

		} else if webOutputFile != "" {
			// Save to specific file
			err := os.WriteFile(webOutputFile, []byte(content), 0644)
			if err != nil {
				fmt.Printf("Error writing to file %s: %v\n", webOutputFile, err)
				os.Exit(1)
			}
			fmt.Printf("✅ Content extracted and saved to: %s\n", webOutputFile)

		} else if webOutputDir != "" {
			// Save to custom directory
			if err := os.MkdirAll(webOutputDir, 0755); err != nil {
				fmt.Printf("Error creating output directory: %v\n", err)
				os.Exit(1)
			}

			filename := fmt.Sprintf("%s.md", title)
			outputPath := filepath.Join(webOutputDir, filename)

			err := os.WriteFile(outputPath, []byte(content), 0644)
			if err != nil {
				fmt.Printf("Error writing to file %s: %v\n", outputPath, err)
				os.Exit(1)
			}
			fmt.Printf("✅ Content extracted and saved to: %s\n", outputPath)

		} else {
			// Output to stdout
			fmt.Print(content)
		}
	},
}

// isValidURL performs basic URL validation
func isValidURL(url string) bool {
	url = strings.TrimSpace(url)
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func init() {
	// Add web command to root
	rootCmd.AddCommand(webCmd)

	// Add subcommands to web
	webCmd.AddCommand(webExtractCmd)

	// Add flags to extract command
	webExtractCmd.Flags().StringVarP(&webOutputFile, "output", "o", "", "Output file path (default: stdout)")
	webExtractCmd.Flags().StringVarP(&webOutputDir, "dir", "d", "", "Output directory path")
	webExtractCmd.Flags().StringVarP(&webProjectName, "project", "p", "", "Project name (creates project folder structure)")
	webExtractCmd.Flags().BoolVarP(&webVerbose, "verbose", "v", false, "Verbose output")
}

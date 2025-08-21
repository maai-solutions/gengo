package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	extractors "maai.solutions/gengo/internal/extractors/pdf"
)

var (
	outputFile  string
	pages       []int
	cleanText   bool
	projectName string
)

// pdfCmd represents the pdf command
var pdfCmd = &cobra.Command{
	Use:   "pdf",
	Short: "Extract text from PDF files",
	Long: `Extract text from PDF files using various input and output formats.
	
Examples:
  gengo pdf extract file.pdf                    # Extract all text to stdout
  gengo pdf extract file.pdf --output text.txt  # Extract all text to file
  gengo pdf extract file.pdf --pages 1,3,5      # Extract specific pages
  gengo pdf extract file.pdf --clean            # Extract and clean text
  gengo pdf info file.pdf                       # Get PDF information`,
}

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract [pdf-file]",
	Short: "Extract text from a PDF file",
	Long: `Extract text from a PDF file and output to stdout or a file.
	
The command supports various options:
- Extract all pages or specific pages
- Output to stdout or save to file
- Clean extracted text by removing excessive whitespace`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pdfFile := args[0]

		// Check if file exists
		if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
			fmt.Printf("Error: File does not exist: %s\n", pdfFile)
			os.Exit(1)
		}

		// Create PDF extractor
		extractor := extractors.NewTextExtractor()

		var text string
		var err error

		// Extract text
		if len(pages) > 0 {
			text, err = extractor.ExtractPages(pdfFile, pages)
			if err != nil {
				fmt.Printf("Error extracting pages %v from PDF: %v\n", pages, err)
				os.Exit(1)
			}
		} else {
			text, err = extractor.ExtractFromFile(pdfFile)
			if err != nil {
				fmt.Printf("Error extracting text from PDF: %v\n", err)
				os.Exit(1)
			}
		}

		// Clean text if requested
		if cleanText {
			text = extractor.CleanText(text)
		}

		// Create output configuration
		outputConfig := NewOutputConfig(projectName, outputFile)
		
		// Generate default filename from PDF file
		defaultFilename := filepath.Base(pdfFile) + ".txt"
		
		// Output text
		if projectName != "" || outputFile != "" {
			outputPath, err := outputConfig.SaveToProject([]byte(text), defaultFilename)
			if err != nil {
				fmt.Printf("Error saving text: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Text extracted and saved to: %s\n", outputPath)
		} else {
			fmt.Print(text)
		}
	},
}

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info [pdf-file]",
	Short: "Get information about a PDF file",
	Long:  `Get information about a PDF file such as page count.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pdfFile := args[0]

		// Check if file exists
		if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
			fmt.Printf("Error: File does not exist: %s\n", pdfFile)
			os.Exit(1)
		}

		// Create PDF extractor
		extractor := extractors.NewTextExtractor()

		// Get page count
		pageCount, err := extractor.GetPageCount(pdfFile)
		if err != nil {
			fmt.Printf("Error getting PDF info: %v\n", err)
			os.Exit(1)
		}

		// Get file info
		fileInfo, _ := os.Stat(pdfFile)

		fmt.Printf("PDF Information:\n")
		fmt.Printf("  File: %s\n", filepath.Base(pdfFile))
		fmt.Printf("  Path: %s\n", pdfFile)
		fmt.Printf("  Size: %d bytes\n", fileInfo.Size())
		fmt.Printf("  Pages: %d\n", pageCount)
	},
}

func init() {
	// Add pdf command to root
	rootCmd.AddCommand(pdfCmd)

	// Add subcommands to pdf
	pdfCmd.AddCommand(extractCmd)
	pdfCmd.AddCommand(infoCmd)

	// Add flags to extract command
	extractCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output filename (saved in PROJECTS or project subdirectory)")
	extractCmd.Flags().StringVarP(&projectName, "project", "p", "", "Project name (creates subdirectory in PROJECTS)")
	extractCmd.Flags().IntSliceVarP(&pages, "pages", "", []int{}, "Specific pages to extract (e.g., --pages 1,3,5)")
	extractCmd.Flags().BoolVarP(&cleanText, "clean", "c", false, "Clean extracted text by removing excessive whitespace")
}

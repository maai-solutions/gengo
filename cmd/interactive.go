package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"maai.solutions/gengo/internal/extractors/asr"
	extractors "maai.solutions/gengo/internal/extractors/pdf"
	webextractors "maai.solutions/gengo/internal/extractors/web"
	"maai.solutions/gengo/internal/extractors/ytaudio"
)

type model struct {
	input   string
	cursor  int
	history []string
}

func initialModel() model {
	return model{
		input:   "",
		cursor:  0,
		history: []string{},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			command := strings.TrimSpace(m.input)
			if command != "" {
				m.history = append(m.history, fmt.Sprintf("> %s", command))

				// Handle commands
				if command == "/exit" || command == "/quit" {
					return m, tea.Quit
				}

				response := m.handleCommand(command)
				if response != "" {
					m.history = append(m.history, response)
				}
			}
			m.input = ""
			m.cursor = 0
		case "backspace":
			if m.cursor > 0 {
				m.input = m.input[:m.cursor-1] + m.input[m.cursor:]
				m.cursor--
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.cursor < len(m.input) {
				m.cursor++
			}
		default:
			// Insert character at cursor position
			if len(msg.String()) == 1 {
				m.input = m.input[:m.cursor] + msg.String() + m.input[m.cursor:]
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString("GenGo Interactive CLI\n")
	s.WriteString("Type '/help' for commands, '/exit' to quit or 'Ctrl+C'\n")
	s.WriteString(strings.Repeat("─", 50) + "\n\n")

	// Show command history
	for _, line := range m.history {
		s.WriteString(line + "\n")
	}

	// Show startup help if no history
	if len(m.history) == 0 {
		s.WriteString("Welcome! Try these commands:\n")
		s.WriteString("  /help                           - Show all commands\n")
		s.WriteString("  ytaudio check                   - Check YouTube transcription setup\n")
		s.WriteString("  pdf info <file.pdf>             - Get PDF information\n")
		s.WriteString("  web extract <url>               - Extract web page content\n\n")
	}

	// Show current input with cursor
	s.WriteString("> ")
	for i, r := range m.input {
		if i == m.cursor {
			s.WriteString("│")
		}
		s.WriteString(string(r))
	}
	if m.cursor >= len(m.input) {
		s.WriteString("│")
	}

	return s.String()
} // handleCommand processes interactive commands and returns response
func (m model) handleCommand(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "/help":
		return m.getHelpText()
	case "ytaudio":
		return m.handleYtAudioCommand(args)
	case "pdf":
		return m.handlePdfCommand(args)
	case "web":
		return m.handleWebCommand(args)
	default:
		return fmt.Sprintf("Unknown command: %s\nType '/help' for available commands", cmd)
	}
}

// getHelpText returns help information for available commands
func (m model) getHelpText() string {
	return `Available commands:
  /help                                    - Show this help
  /exit, /quit                            - Exit the interactive mode
  
  ytaudio transcribe <youtube-url>        - Transcribe YouTube video
  ytaudio check                           - Check ytaudio dependencies
  
  pdf extract <file.pdf>                  - Extract text from PDF
  pdf info <file.pdf>                     - Get PDF information
  
  web extract <url>                       - Extract content from web page
  
Examples:
  ytaudio transcribe https://youtube.com/watch?v=abc123
  pdf extract document.pdf
  web extract https://example.com/article
  pdf info document.pdf`
}

// handleYtAudioCommand processes ytaudio subcommands
func (m model) handleYtAudioCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: ytaudio [transcribe|check] [args...]"
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "transcribe":
		return m.handleYtAudioTranscribe(subArgs)
	case "check":
		return m.handleYtAudioCheck()
	default:
		return fmt.Sprintf("Unknown ytaudio subcommand: %s\nAvailable: transcribe, check", subCmd)
	}
}

// handleYtAudioTranscribe handles YouTube transcription
func (m model) handleYtAudioTranscribe(args []string) string {
	if len(args) == 0 {
		return "Usage: ytaudio transcribe <youtube-url>"
	}

	videoURL := args[0]

	// Validate YouTube URL (basic check)
	if !isValidYouTubeURL(videoURL) {
		return fmt.Sprintf("Error: Invalid YouTube URL: %s\nPlease provide a valid YouTube URL", videoURL)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Configure ASR with default settings
	asrConfig := asr.DefaultConfig()

	// Configure YouTube transcription service
	outputDir := "./transcripts"
	config := &ytaudio.Config{
		OutputDir:    outputDir,
		ASRConfig:    asrConfig,
		CleanupFiles: false, // Keep files by default in interactive mode
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Sprintf("Error creating output directory: %v", err)
	}

	// Create service and transcribe
	service := ytaudio.NewService(config)
	result, err := service.TranscribeYouTubeVideo(ctx, videoURL)
	if err != nil {
		return fmt.Sprintf("Error transcribing video: %v", err)
	}

	// Generate filename and save transcript
	filename := generateTranscriptFilename(videoURL)
	transcriptPath := filepath.Join(outputDir, filename)

	// Create markdown content with metadata
	content := formatTranscriptMarkdown(videoURL, result)

	if err := os.WriteFile(transcriptPath, []byte(content), 0644); err != nil {
		return fmt.Sprintf("Error saving transcript: %v", err)
	}

	return fmt.Sprintf("✅ Transcription completed!\nSaved to: %s\nDuration: %.2f seconds",
		transcriptPath, result.Duration.Seconds())
}

// handleYtAudioCheck checks ytaudio dependencies
func (m model) handleYtAudioCheck() string {
	result := `YouTube Audio Transcription Dependencies:

Required tools:
- yt-dlp or youtube-dl (for downloading)
- whisper or whisper.cpp (for transcription)

Installation instructions:
- Install yt-dlp: pip install yt-dlp
- Install whisper: pip install openai-whisper
- Or install whisper.cpp: https://github.com/ggerganov/whisper.cpp

Please install whisper models:
  pip install openai-whisper
  whisper --model base "test"  # This downloads the base model
  
For whisper.cpp:
  Download models from https://huggingface.co/ggerganov/whisper.cpp`

	return result
}

// handlePdfCommand processes pdf subcommands
func (m model) handlePdfCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: pdf [extract|info] <file.pdf> [options]"
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "extract":
		return m.handlePdfExtract(subArgs)
	case "info":
		return m.handlePdfInfo(subArgs)
	default:
		return fmt.Sprintf("Unknown pdf subcommand: %s\nAvailable: extract, info", subCmd)
	}
}

// handlePdfExtract handles PDF text extraction
func (m model) handlePdfExtract(args []string) string {
	if len(args) == 0 {
		return "Usage: pdf extract <file.pdf> [--pages 1,2,3] [--output output.txt] [--clean]"
	}

	pdfFile := args[0]

	// Check if file exists
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		return fmt.Sprintf("Error: File does not exist: %s", pdfFile)
	}

	// Parse additional arguments
	var outputFile string
	var pages []int
	var cleanText bool

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "--clean":
			cleanText = true
		case "--pages":
			if i+1 < len(args) {
				pageStr := args[i+1]
				pageNums := strings.Split(pageStr, ",")
				for _, p := range pageNums {
					if num, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
						pages = append(pages, num)
					}
				}
				i++
			}
		}
	}

	// Create PDF extractor
	extractor := extractors.NewTextExtractor()

	var text string
	var err error

	// Extract text
	if len(pages) > 0 {
		text, err = extractor.ExtractPages(pdfFile, pages)
		if err != nil {
			return fmt.Sprintf("Error extracting pages %v from PDF: %v", pages, err)
		}
	} else {
		text, err = extractor.ExtractFromFile(pdfFile)
		if err != nil {
			return fmt.Sprintf("Error extracting text from PDF: %v", err)
		}
	}

	// Clean text if requested
	if cleanText {
		text = extractor.CleanText(text)
	}

	// Output text
	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(text), 0644)
		if err != nil {
			return fmt.Sprintf("Error writing to file %s: %v", outputFile, err)
		}
		return fmt.Sprintf("✅ Text extracted and saved to: %s", outputFile)
	} else {
		// For interactive mode, show first 500 characters
		if len(text) > 500 {
			return fmt.Sprintf("✅ Text extracted (showing first 500 chars):\n\n%s...\n\n[Total length: %d characters]",
				text[:500], len(text))
		}
		return fmt.Sprintf("✅ Text extracted:\n\n%s", text)
	}
}

// handlePdfInfo handles PDF information retrieval
func (m model) handlePdfInfo(args []string) string {
	if len(args) == 0 {
		return "Usage: pdf info <file.pdf>"
	}

	pdfFile := args[0]

	// Check if file exists
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		return fmt.Sprintf("Error: File does not exist: %s", pdfFile)
	}

	// Create PDF extractor and get info
	extractor := extractors.NewTextExtractor()
	pageCount, err := extractor.GetPageCount(pdfFile)
	if err != nil {
		return fmt.Sprintf("Error getting PDF info: %v", err)
	}

	// Get file info
	fileInfo, _ := os.Stat(pdfFile)

	return fmt.Sprintf("✅ PDF Information:\n  File: %s\n  Path: %s\n  Size: %d bytes\n  Pages: %d",
		filepath.Base(pdfFile), pdfFile, fileInfo.Size(), pageCount)
}

// handleWebCommand processes web subcommands
func (m model) handleWebCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: web [extract] <url> [options]"
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "extract":
		return m.handleWebExtract(subArgs)
	default:
		return fmt.Sprintf("Unknown web subcommand: %s\nAvailable: extract", subCmd)
	}
}

// handleWebExtract handles web page content extraction
func (m model) handleWebExtract(args []string) string {
	if len(args) == 0 {
		return "Usage: web extract <url> [--output output.md] [--project project-name]"
	}

	url := args[0]

	// Validate URL (basic check)
	if !isValidURL(url) {
		return fmt.Sprintf("Error: Invalid URL: %s\nPlease provide a valid URL (e.g., https://example.com)", url)
	}

	// Parse additional arguments
	var outputFile string
	var projectName string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "--project":
			if i+1 < len(args) {
				projectName = args[i+1]
				i++
			}
		}
	}

	// Extract content from web page
	title, content, err := webextractors.DownloadAndExtract(url)
	if err != nil {
		return fmt.Sprintf("Error extracting content: %v", err)
	}

	// Handle output based on specified options
	if projectName != "" {
		// Save to project structure
		err := webextractors.SaveToProject(title, content, projectName)
		if err != nil {
			return fmt.Sprintf("Error saving to project: %v", err)
		}

		projectPath := filepath.Join(".", projectName, fmt.Sprintf("%s.md", title))
		return fmt.Sprintf("✅ Content extracted and saved to project!\nFile: %s\nTitle: %s", projectPath, title)

	} else if outputFile != "" {
		// Save to specific file
		err := os.WriteFile(outputFile, []byte(content), 0644)
		if err != nil {
			return fmt.Sprintf("Error writing to file %s: %v", outputFile, err)
		}
		return fmt.Sprintf("✅ Content extracted and saved to: %s\nTitle: %s", outputFile, title)

	} else {
		// For interactive mode, show first 800 characters
		if len(content) > 800 {
			return fmt.Sprintf("✅ Content extracted from: %s\nTitle: %s\n\n(Showing first 800 chars):\n\n%s...\n\n[Total length: %d characters]\n\nTip: Use --output or --project to save the full content",
				url, title, content[:800], len(content))
		}
		return fmt.Sprintf("✅ Content extracted from: %s\nTitle: %s\n\n%s", url, title, content)
	}
}

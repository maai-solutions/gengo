package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"maai.solutions/gengo/internal/extractors/asr"
	"maai.solutions/gengo/internal/extractors/ytaudio"
)

var (
	ytOutputDir   string
	ytModel       string
	ytVerbose     bool
	ytKeepFiles   bool
	ytTimeout     time.Duration
	ytProjectName string
)

// ytaudioCmd represents the ytaudio command
var ytaudioCmd = &cobra.Command{
	Use:   "ytaudio",
	Short: "Extract and transcribe audio from YouTube videos",
	Long: `Extract audio from YouTube videos and transcribe them using Whisper ASR.
	
Examples:
  gengo ytaudio transcribe https://youtube.com/watch?v=example    # Basic transcription
  gengo ytaudio transcribe url --project my-project              # Save to project folder
  gengo ytaudio transcribe url --model large --verbose           # Use large model with verbose output
  gengo ytaudio transcribe url --keep --output ./transcripts     # Keep downloaded files
  gengo ytaudio check                                             # Check dependencies`,
}

// transcribeCmd represents the transcribe command
var transcribeCmd = &cobra.Command{
	Use:   "transcribe [youtube-url]",
	Short: "Transcribe audio from a YouTube video",
	Long: `Download audio from a YouTube video and transcribe it using Whisper ASR.
	
The command supports various options:
- Specify Whisper model (tiny, base, small, medium, large)
- Save transcription to project folder or custom output directory
- Keep or cleanup downloaded files
- Verbose output for detailed progress`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		videoURL := args[0]

		// Validate YouTube URL (basic check)
		if !isValidYouTubeURL(videoURL) {
			fmt.Printf("Error: Invalid YouTube URL: %s\n", videoURL)
			fmt.Println("Please provide a valid YouTube URL (e.g., https://youtube.com/watch?v=...)")
			os.Exit(1)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), ytTimeout)
		defer cancel()

		// Configure ASR
		asrConfig := asr.DefaultConfig()
		if ytModel != "" {
			modelPath := ytaudio.FindWhisperModel(ytModel)
			if modelPath == "" {
				fmt.Printf("Error: Whisper model '%s' not found\n", ytModel)
				fmt.Println("Available models: tiny, base, small, medium, large")
				fmt.Println("Make sure the model is installed and in a standard location")
				os.Exit(1)
			}
			asrConfig.WhisperModel = modelPath
		}

		// Configure YouTube transcription service
		config := &ytaudio.Config{
			OutputDir:    ytOutputDir,
			ASRConfig:    asrConfig,
			CleanupFiles: !ytKeepFiles,
		}

		// Ensure output directory exists
		if err := os.MkdirAll(ytOutputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		if ytVerbose {
			fmt.Printf("Starting transcription of: %s\n", videoURL)
			fmt.Printf("Output directory: %s\n", ytOutputDir)
			fmt.Printf("Whisper model: %s\n", ytModel)
			fmt.Printf("Keep files: %t\n", ytKeepFiles)
		}

		// Create service and transcribe
		service := ytaudio.NewService(config)
		result, err := service.TranscribeYouTubeVideo(ctx, videoURL)
		if err != nil {
			fmt.Printf("Error transcribing video: %v\n", err)
			os.Exit(1)
		}

		// Handle output based on project name or direct output
		if ytProjectName != "" {
			// Save to project structure
			projectDir := filepath.Join(ytOutputDir, ytProjectName)
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				fmt.Printf("Error creating project directory: %v\n", err)
				os.Exit(1)
			}

			// Generate filename from video URL/ID
			filename := generateTranscriptFilename(videoURL)
			transcriptPath := filepath.Join(projectDir, filename)

			// Create markdown content with metadata
			content := formatTranscriptMarkdown(videoURL, result)

			if err := os.WriteFile(transcriptPath, []byte(content), 0644); err != nil {
				fmt.Printf("Error writing transcript file: %v\n", err)
				os.Exit(1)
			}

			if ytVerbose {
				fmt.Printf("Transcription completed in %v\n", result.Duration)
			}
			fmt.Printf("Transcript saved to: %s\n", transcriptPath)
		} else {
			// Output to stdout
			if ytVerbose {
				fmt.Printf("Transcription completed in %v\n", result.Duration)
				fmt.Println("--- Transcript ---")
			}
			fmt.Println(result.Text)
		}
	},
}

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if required dependencies are available",
	Long: `Check if all required dependencies for YouTube audio transcription are available.
	
This includes:
- ffmpeg (for audio conversion)
- whisper or whisper.cpp (for transcription)
- Required Python packages (if using OpenAI Whisper)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking YouTube audio transcription dependencies...")

		if err := ytaudio.CheckDependencies(); err != nil {
			fmt.Printf("❌ Dependency check failed: %v\n", err)
			fmt.Println("\nTo fix this, please install the missing dependencies:")
			fmt.Println("- Install ffmpeg: https://ffmpeg.org/download.html")
			fmt.Println("- Install whisper: pip install openai-whisper")
			fmt.Println("- Or install whisper.cpp: https://github.com/ggerganov/whisper.cpp")
			os.Exit(1)
		}

		fmt.Println("✅ All dependencies are available!")

		// Show available models
		fmt.Println("\nAvailable Whisper models:")
		models := []string{"tiny", "base", "small", "medium", "large"}
		for _, model := range models {
			modelPath := ytaudio.FindWhisperModel(model)
			if modelPath != "" {
				fmt.Printf("  ✅ %s: %s\n", model, modelPath)
			} else {
				fmt.Printf("  ❌ %s: not found\n", model)
			}
		}
	},
}

// modelsCmd represents the models command
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available Whisper models",
	Long:  `List all available Whisper models and their locations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available Whisper models:")

		models := []string{"tiny", "base", "small", "medium", "large"}
		foundAny := false

		for _, model := range models {
			modelPath := ytaudio.FindWhisperModel(model)
			if modelPath != "" {
				fmt.Printf("  ✅ %s: %s\n", model, modelPath)
				foundAny = true
			} else {
				fmt.Printf("  ❌ %s: not found\n", model)
			}
		}

		if !foundAny {
			fmt.Println("\nNo Whisper models found!")
			fmt.Println("Please install whisper models:")
			fmt.Println("  pip install openai-whisper")
			fmt.Println("  # Models will be downloaded automatically on first use")
		}
	},
}

func init() {
	// Add ytaudio command to root
	rootCmd.AddCommand(ytaudioCmd)

	// Add subcommands to ytaudio
	ytaudioCmd.AddCommand(transcribeCmd)
	ytaudioCmd.AddCommand(checkCmd)
	ytaudioCmd.AddCommand(modelsCmd)

	// Add flags to transcribe command
	transcribeCmd.Flags().StringVarP(&ytOutputDir, "output", "o", "./ytaudio_output", "Output directory for transcripts and temporary files")
	transcribeCmd.Flags().StringVarP(&ytModel, "model", "m", "base", "Whisper model to use (tiny, base, small, medium, large)")
	transcribeCmd.Flags().BoolVarP(&ytVerbose, "verbose", "v", false, "Enable verbose output")
	transcribeCmd.Flags().BoolVarP(&ytKeepFiles, "keep", "k", false, "Keep downloaded audio files")
	transcribeCmd.Flags().DurationVarP(&ytTimeout, "timeout", "t", 30*time.Minute, "Timeout for the entire operation")
	transcribeCmd.Flags().StringVarP(&ytProjectName, "project", "p", "", "Save transcript to a project folder (creates organized structure)")
}

// isValidYouTubeURL performs basic validation of YouTube URLs
func isValidYouTubeURL(url string) bool {
	// Basic YouTube URL patterns
	patterns := []string{
		"youtube.com/watch",
		"youtu.be/",
		"youtube.com/embed/",
		"youtube.com/v/",
		"m.youtube.com/watch",
	}

	for _, pattern := range patterns {
		if contains(url, pattern) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOf(s, substr) >= 0))
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// generateTranscriptFilename creates a filename from a YouTube URL
func generateTranscriptFilename(videoURL string) string {
	// Extract video ID from various YouTube URL formats
	videoID := extractVideoID(videoURL)
	if videoID == "" {
		videoID = "transcript"
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return fmt.Sprintf("%s_%s.md", videoID, timestamp)
}

// extractVideoID extracts the video ID from a YouTube URL
func extractVideoID(url string) string {
	// Handle youtube.com/watch?v=ID format
	if idx := indexOf(url, "v="); idx >= 0 {
		start := idx + 2
		end := start
		for end < len(url) && url[end] != '&' && url[end] != '#' {
			end++
		}
		if end > start {
			return url[start:end]
		}
	}

	// Handle youtu.be/ID format
	if idx := indexOf(url, "youtu.be/"); idx >= 0 {
		start := idx + 9
		end := start
		for end < len(url) && url[end] != '?' && url[end] != '#' {
			end++
		}
		if end > start {
			return url[start:end]
		}
	}

	return ""
}

// formatTranscriptMarkdown formats the transcription result as markdown
func formatTranscriptMarkdown(videoURL string, result *ytaudio.TranscriptionResult) string {
	videoID := extractVideoID(videoURL)
	title := "YouTube Video Transcript"
	if videoID != "" {
		title = fmt.Sprintf("YouTube Video Transcript (%s)", videoID)
	}

	content := fmt.Sprintf(`# %s

**Source:** %s  
**Transcribed:** %s  
**Duration:** %v  

---

## Transcript

%s
`, title, videoURL, time.Now().Format("2006-01-02 15:04:05"), result.Duration, result.Text)

	return content
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"maai.solutions/gengo/internal/extractors/asr"
)

var (
	audioModel       string
	audioVerbose     bool
	audioKeepFiles   bool
	audioTimeout     time.Duration
	audioProjectName string
	audioOutputFile  string
)

// audioCmd represents the audio command
var audioCmd = &cobra.Command{
	Use:   "audio",
	Short: "Transcribe audio files using Whisper ASR",
	Long: `Transcribe audio files (MP3, WAV, etc.) using Whisper ASR.
	
Examples:
  gengo audio transcribe audio.mp3                              # Basic transcription
  gengo audio transcribe audio.mp3 --project my-project         # Save to project folder
  gengo audio transcribe audio.mp3 --model large --verbose      # Use large model with verbose output
  gengo audio transcribe audio.mp3 --output ./transcripts       # Save to custom output
  gengo audio check                                             # Check dependencies`,
}

// audioTranscribeCmd represents the transcribe command for audio files
var audioTranscribeCmd = &cobra.Command{
	Use:   "transcribe [audio-file]",
	Short: "Transcribe an audio file",
	Long: `Transcribe an audio file (MP3, WAV, etc.) using Whisper ASR.
	
The command supports various options:
- Specify Whisper model (tiny, base, small, medium, large)
- Save transcription to project folder or custom output directory
- Keep or cleanup temporary files
- Verbose output for detailed progress`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Debug: Print all flag values at the very start
		if os.Getenv("GENGO_DEBUG") != "" {
			fmt.Printf("DEBUG FLAGS: project=%q, output=%q, verbose=%v\n", audioProjectName, audioOutputFile, audioVerbose)
		}
		
		audioPath := args[0]

		// Check if file exists
		if _, err := os.Stat(audioPath); err != nil {
			fmt.Printf("Error: Audio file not found: %s\n", audioPath)
			os.Exit(1)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), audioTimeout)
		defer cancel()

		// Configure ASR
		asrConfig := asr.DefaultConfig()
		if audioModel != "" {
			modelPath := asr.FindWhisperModel(audioModel)
			if modelPath == "" {
				fmt.Printf("Error: Whisper model '%s' not found\n", audioModel)
				fmt.Println("Available models: tiny, base, small, medium, large")
				fmt.Println("Make sure the model is installed and in a standard location")
				os.Exit(1)
			}
			asrConfig.WhisperModel = modelPath
		}

		// Create output configuration
		outputConfig := NewOutputConfig(audioProjectName, audioOutputFile)
		
		// Debug output
		if os.Getenv("GENGO_DEBUG") != "" {
			fmt.Printf("DEBUG audio: ProjectName=%q, OutputFile=%q\n", audioProjectName, audioOutputFile)
		}
		
		// Create temp directory for processing
		tempDir := filepath.Join(os.TempDir(), "gengo-audio")
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			fmt.Printf("Error creating temp directory: %v\n", err)
			os.Exit(1)
		}
		
		// Clean up temp directory if not keeping files
		if !audioKeepFiles {
			defer os.RemoveAll(tempDir)
		}

		if audioVerbose {
			fmt.Printf("Starting transcription of: %s\n", audioPath)
			if audioProjectName != "" {
				fmt.Printf("Project: %s\n", audioProjectName)
			}
			if audioOutputFile != "" {
				fmt.Printf("Output file: %s\n", audioOutputFile)
			}
			fmt.Printf("Whisper model: %s\n", audioModel)
			fmt.Printf("Keep temp files: %t\n", audioKeepFiles)
		}

		// Create ASR service and transcribe
		service := asr.NewService(asrConfig)
		
		// Measure transcription time
		startTime := time.Now()
		
		// Transcribe the audio file
		result, err := service.TranscribeAudio(ctx, audioPath, tempDir)
		if err != nil {
			fmt.Printf("Error transcribing audio: %v\n", err)
			os.Exit(1)
		}
		
		duration := time.Since(startTime)

		// Generate default filename from audio file
		defaultFilename := generateAudioTranscriptFilename(audioPath)
		
		// Create markdown content with metadata
		content := formatAudioTranscriptMarkdown(audioPath, result, duration)
		
		// Save to project structure
		if audioProjectName != "" || audioOutputFile != "" {
			// More debug
			if os.Getenv("GENGO_DEBUG") != "" {
				fmt.Printf("DEBUG: Before save - ProjectName=%q, OutputFile=%q, DefaultFile=%q\n", 
					audioProjectName, audioOutputFile, defaultFilename)
			}
			
			outputPath, err := outputConfig.SaveToProject([]byte(content), defaultFilename)
			if err != nil {
				fmt.Printf("Error saving transcript: %v\n", err)
				os.Exit(1)
			}
			
			if audioVerbose {
				fmt.Printf("Transcription completed in %v\n", duration)
			}
			fmt.Printf("Transcript saved to: %s\n", outputPath)
		} else {
			// Output to stdout if no output specified
			if audioVerbose {
				fmt.Printf("Transcription completed in %v\n", duration)
				fmt.Println("--- Transcript ---")
			}
			fmt.Println(result.Text)
		}
	},
}

// audioCheckCmd represents the check command for audio transcription
var audioCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if required dependencies are available",
	Long: `Check if all required dependencies for audio transcription are available.
	
This includes:
- ffmpeg (for audio conversion)
- whisper or whisper.cpp (for transcription)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking audio transcription dependencies...")

		if err := asr.CheckDependencies(); err != nil {
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
			modelPath := asr.FindWhisperModel(model)
			if modelPath != "" {
				fmt.Printf("  ✅ %s: %s\n", model, modelPath)
			} else {
				fmt.Printf("  ❌ %s: not found\n", model)
			}
		}
	},
}

// audioModelsCmd represents the models command for audio transcription
var audioModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available Whisper models",
	Long:  `List all available Whisper models and their locations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available Whisper models:")

		models := []string{"tiny", "base", "small", "medium", "large"}
		foundAny := false

		for _, model := range models {
			modelPath := asr.FindWhisperModel(model)
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
	// Add audio command to root
	rootCmd.AddCommand(audioCmd)

	// Add subcommands to audio
	audioCmd.AddCommand(audioTranscribeCmd)
	audioCmd.AddCommand(audioCheckCmd)
	audioCmd.AddCommand(audioModelsCmd)

	// Add flags to transcribe command
	audioTranscribeCmd.Flags().StringVarP(&audioOutputFile, "output", "o", "", "Output filename (saved in PROJECTS or project subdirectory)")
	audioTranscribeCmd.Flags().StringVarP(&audioModel, "model", "m", "base", "Whisper model to use (tiny, base, small, medium, large)")
	audioTranscribeCmd.Flags().BoolVarP(&audioVerbose, "verbose", "v", false, "Enable verbose output")
	audioTranscribeCmd.Flags().BoolVarP(&audioKeepFiles, "keep", "k", false, "Keep temporary files")
	audioTranscribeCmd.Flags().DurationVarP(&audioTimeout, "timeout", "t", 30*time.Minute, "Timeout for the entire operation")
	audioTranscribeCmd.Flags().StringVarP(&audioProjectName, "project", "p", "", "Project name (creates subdirectory in PROJECTS)")
}

// generateAudioTranscriptFilename creates a filename from an audio file path
func generateAudioTranscriptFilename(audioPath string) string {
	// Get base filename without extension
	baseName := filepath.Base(audioPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	
	// Sanitize filename
	nameWithoutExt = sanitizeAudioFilename(nameWithoutExt)
	if nameWithoutExt == "" {
		nameWithoutExt = "transcript"
	}
	
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return fmt.Sprintf("%s_%s.md", nameWithoutExt, timestamp)
}

// sanitizeAudioFilename removes or replaces characters that are not safe for filenames
func sanitizeAudioFilename(filename string) string {
	// Replace spaces with underscores
	filename = strings.ReplaceAll(filename, " ", "_")
	
	// Remove or replace unsafe characters
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t"}
	for _, char := range unsafe {
		filename = strings.ReplaceAll(filename, char, "")
	}
	
	// Replace multiple underscores with single underscore
	for strings.Contains(filename, "__") {
		filename = strings.ReplaceAll(filename, "__", "_")
	}
	
	// Trim underscores from start and end
	filename = strings.Trim(filename, "_")
	
	return filename
}

// formatAudioTranscriptMarkdown formats the transcription result as markdown
func formatAudioTranscriptMarkdown(audioPath string, result *asr.Result, duration time.Duration) string {
	// Get absolute path for the source file
	absPath, _ := filepath.Abs(audioPath)
	fileName := filepath.Base(audioPath)
	
	content := fmt.Sprintf(`# Audio Transcript: %s

**Source File:** %s  
**Full Path:** %s  
**Transcribed:** %s  
**Processing Time:** %v  

---

## Transcript

%s
`, fileName, fileName, absPath, time.Now().Format("2006-01-02 15:04:05"), duration, result.Text)

	if result.Language != "" {
		content = strings.Replace(content, 
			"**Processing Time:**", 
			fmt.Sprintf("**Language:** %s  \n**Processing Time:**", result.Language), 
			1)
	}

	return content
}
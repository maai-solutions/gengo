package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"maai.solutions/gengo/internal/extractors/asr"
	"maai.solutions/gengo/internal/extractors/podcast"
)

var (
	podcastModel       string
	podcastVerbose     bool
	podcastKeepFiles   bool
	podcastTimeout     time.Duration
	podcastProjectName string
	podcastOutputFile  string
)

// podcastCmd represents the podcast command
var podcastCmd = &cobra.Command{
	Use:   "podcast",
	Short: "Download and transcribe podcast episodes",
	Long: `Download podcast episodes and transcribe them using Whisper ASR.

Examples:
  gengo podcast list https://feeds.example.com/podcast.rss           # List episodes
  gengo podcast transcribe https://feeds.example.com/episode.mp3    # Direct episode URL
  gengo podcast transcribe https://feeds.example.com/rss 5          # Episode 5 from feed
  gengo podcast transcribe url --project my-project                 # Save to project folder
  gengo podcast transcribe url --model large --verbose              # Use large model with verbose output`,
}

// listCmd represents the list command
var podcastListCmd = &cobra.Command{
	Use:   "list [rss-url]",
	Short: "List episodes from a podcast RSS feed",
	Long: `Download and display a list of podcast episodes from an RSS feed.

This command fetches the podcast RSS feed and displays information about all available episodes including:
- Episode number (sorted by date, newest first)
- Episode title
- Audio file URL (for use with transcribe command)
- Publication date (with --verbose flag)`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rssURL := args[0]

		// Validate RSS URL (basic check)
		if !isValidURL(rssURL) {
			fmt.Printf("Error: Invalid RSS URL: %s\n", rssURL)
			fmt.Println("Please provide a valid RSS URL (e.g., https://example.com/feed.rss)")
			os.Exit(1)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Configure podcast service
		tempDir := filepath.Join(os.TempDir(), "gengo-podcast")
		config := &podcast.Config{
			OutputDir:    tempDir,
			ASRConfig:    asr.DefaultConfig(),
			CleanupFiles: true,
		}

		if podcastVerbose {
			fmt.Printf("Fetching episodes from: %s\n", rssURL)
		}

		// Create service and get episodes
		service := podcast.NewService(config)
		episodes, err := service.ListEpisodes(ctx, rssURL)
		if err != nil {
			fmt.Printf("Error fetching episodes: %v\n", err)
			os.Exit(1)
		}

		if len(episodes) == 0 {
			fmt.Println("No episodes found in the podcast feed.")
			return
		}

		// Display results
		fmt.Printf("Podcast: %s\n", episodes[0].PodcastTitle)
		fmt.Printf("Episodes: %d\n\n", len(episodes))

		for _, episode := range episodes {
			fmt.Printf("Episode %d: %s\n", episode.Number, episode.Title)
			fmt.Printf("  URL: %s\n", episode.FileURL)
			if podcastVerbose {
				fmt.Printf("  Date: %s\n", episode.Date)
			}
			fmt.Println()
		}
	},
}

// transcribeCmd represents the transcribe command
var podcastTranscribeCmd = &cobra.Command{
	Use:   "transcribe [rss-url-or-episode-url] [episode-number]",
	Short: "Transcribe a podcast episode",
	Long: `Download and transcribe a podcast episode using Whisper ASR.

Usage modes:
1. Direct episode URL:
   gengo podcast transcribe https://example.com/episode.mp3

2. Episode from RSS feed:
   gengo podcast transcribe https://example.com/feed.rss 5

The command supports various options:
- Specify Whisper model (tiny, base, small, medium, large)
- Save transcription to project folder or custom output directory
- Keep or cleanup downloaded files
- Verbose output for detailed progress`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		// Debug: Print all flag values at the very start
		if os.Getenv("GENGO_DEBUG") != "" {
			fmt.Printf("DEBUG FLAGS: project=%q, output=%q, verbose=%v\n", podcastProjectName, podcastOutputFile, podcastVerbose)
		}

		url := args[0]
		var episodeNumber int
		var err error

		// Check if this is episode number mode (2 arguments)
		if len(args) == 2 {
			episodeNumber, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Printf("Error: Invalid episode number: %s\n", args[1])
				fmt.Println("Episode number must be a valid integer")
				os.Exit(1)
			}
		}

		// Validate URL (basic check)
		if !isValidURL(url) {
			fmt.Printf("Error: Invalid URL: %s\n", url)
			fmt.Println("Please provide a valid URL")
			os.Exit(1)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), podcastTimeout)
		defer cancel()

		// Configure ASR
		asrConfig := asr.DefaultConfig()
		if podcastModel != "" {
			modelPath := findWhisperModel(podcastModel)
			if modelPath == "" {
				fmt.Printf("Error: Whisper model '%s' not found\n", podcastModel)
				fmt.Println("Available models: tiny, base, small, medium, large")
				fmt.Println("Make sure the model is installed and in a standard location")
				os.Exit(1)
			}
			asrConfig.WhisperModel = modelPath
		}

		// Create output configuration
		outputConfig := NewOutputConfig(podcastProjectName, podcastOutputFile)

		// Debug output
		if os.Getenv("GENGO_DEBUG") != "" {
			fmt.Printf("DEBUG podcast: ProjectName=%q, OutputFile=%q\n", podcastProjectName, podcastOutputFile)
		}

		// Configure podcast transcription service with temp directory
		tempDir := filepath.Join(os.TempDir(), "gengo-podcast")
		config := &podcast.Config{
			OutputDir:    tempDir,
			ASRConfig:    asrConfig,
			CleanupFiles: !podcastKeepFiles,
		}

		// Ensure temp directory exists
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			fmt.Printf("Error creating temp directory: %v\n", err)
			os.Exit(1)
		}

		if podcastVerbose {
			fmt.Printf("Starting transcription of: %s\n", url)
			if episodeNumber > 0 {
				fmt.Printf("Episode number: %d\n", episodeNumber)
			}
			if podcastProjectName != "" {
				fmt.Printf("Project: %s\n", podcastProjectName)
			}
			if podcastOutputFile != "" {
				fmt.Printf("Output file: %s\n", podcastOutputFile)
			}
			fmt.Printf("Whisper model: %s\n", podcastModel)
			fmt.Printf("Keep temp files: %t\n", podcastKeepFiles)
		}

		// Create service and transcribe
		service := podcast.NewService(config)
		var result *podcast.TranscriptionResult

		if episodeNumber > 0 {
			// Transcribe specific episode from RSS feed
			result, err = service.TranscribeEpisodeFromFeed(ctx, url, episodeNumber)
		} else {
			// Transcribe direct episode URL
			result, err = service.TranscribePodcastEpisode(ctx, url)
		}

		if err != nil {
			fmt.Printf("Error transcribing episode: %v\n", err)
			os.Exit(1)
		}

		// Generate default filename from episode info
		defaultFilename := generatePodcastFilename(result)

		// Create markdown content with metadata
		content := formatPodcastMarkdown(url, result)

		// Save to project structure
		if podcastProjectName != "" || podcastOutputFile != "" {
			// More debug
			if os.Getenv("GENGO_DEBUG") != "" {
				fmt.Printf("DEBUG: Before save - ProjectName=%q, OutputFile=%q, DefaultFile=%q\n",
					podcastProjectName, podcastOutputFile, defaultFilename)
			}

			outputPath, err := outputConfig.SaveToProject([]byte(content), defaultFilename)
			if err != nil {
				fmt.Printf("Error saving transcript: %v\n", err)
				os.Exit(1)
			}

			if podcastVerbose {
				fmt.Printf("Transcription completed in %v\n", result.Duration)
			}
			fmt.Printf("Transcript saved to: %s\n", outputPath)
		} else {
			// Output to stdout if no output specified
			if podcastVerbose {
				fmt.Printf("Transcription completed in %v\n", result.Duration)
				fmt.Println("--- Transcript ---")
			}
			fmt.Println(result.Text)
		}
	},
}

// checkCmd represents the check command
var podcastCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if required dependencies are available",
	Long: `Check if all required dependencies for podcast transcription are available.

This includes:
- ffmpeg (for audio conversion)
- whisper or whisper.cpp (for transcription)
- Required Python packages (if using OpenAI Whisper)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking podcast transcription dependencies...")

		if err := podcast.CheckDependencies(); err != nil {
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
			modelPath := findWhisperModel(model)
			if modelPath != "" {
				fmt.Printf("  ✅ %s: %s\n", model, modelPath)
			} else {
				fmt.Printf("  ❌ %s: not found\n", model)
			}
		}
	},
}

func init() {
	// Add podcast command to root
	rootCmd.AddCommand(podcastCmd)

	// Add subcommands to podcast
	podcastCmd.AddCommand(podcastListCmd)
	podcastCmd.AddCommand(podcastTranscribeCmd)
	podcastCmd.AddCommand(podcastCheckCmd)

	// Add flags to transcribe command
	podcastTranscribeCmd.Flags().StringVarP(&podcastOutputFile, "output", "o", "", "Output filename (saved in PROJECTS or project subdirectory)")
	podcastTranscribeCmd.Flags().StringVarP(&podcastModel, "model", "m", "base", "Whisper model to use (tiny, base, small, medium, large)")
	podcastTranscribeCmd.Flags().BoolVarP(&podcastVerbose, "verbose", "v", false, "Enable verbose output")
	podcastTranscribeCmd.Flags().BoolVarP(&podcastKeepFiles, "keep", "k", false, "Keep downloaded audio files")
	podcastTranscribeCmd.Flags().DurationVarP(&podcastTimeout, "timeout", "t", 60*time.Minute, "Timeout for the entire operation")
	podcastTranscribeCmd.Flags().StringVarP(&podcastProjectName, "project", "p", "", "Project name (creates subdirectory in PROJECTS)")

	// Add flags to list command
	podcastListCmd.Flags().BoolVarP(&podcastVerbose, "verbose", "v", false, "Enable verbose output")
}

// findWhisperModel tries to find the whisper model in common locations
func findWhisperModel(modelName string) string {
	return asr.FindWhisperModel(modelName)
}

// generatePodcastFilename creates a filename from podcast episode information
func generatePodcastFilename(result *podcast.TranscriptionResult) string {
	// Sanitize episode title for filename
	title := sanitizeFilename(result.EpisodeTitle)
	if title == "" {
		title = "episode"
	}

	// Limit title length to avoid filesystem issues
	if len(title) > 50 {
		title = title[:50]
	}

	// Sanitize podcast title
	podcastTitle := sanitizeFilename(result.PodcastTitle)
	if podcastTitle == "" {
		podcastTitle = "podcast"
	}
	if len(podcastTitle) > 30 {
		podcastTitle = podcastTitle[:30]
	}

	return fmt.Sprintf("%s_%s.md", podcastTitle, title)
}

// formatPodcastMarkdown formats the transcription result as markdown
func formatPodcastMarkdown(episodeURL string, result *podcast.TranscriptionResult) string {
	// Use episode title as the document title
	title := result.EpisodeTitle
	if title == "" {
		title = "Podcast Episode Transcript"
	}

	content := fmt.Sprintf(`# %s

**Podcast:** %s  
**Episode:** %s  
**Source:** %s  
**Transcribed:** %s  
**Duration:** %v  

---

## Transcript

%s
`, title, result.PodcastTitle, result.EpisodeTitle, episodeURL, time.Now().Format("2006-01-02 15:04:05"), result.Duration, result.Text)

	return content
}
package podcast

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"maai.solutions/gengo/internal/extractors/asr"
)

// Config holds configuration for the podcast transcription service
type Config struct {
	OutputDir    string
	ASRConfig    *asr.Config // ASR configuration
	CleanupFiles bool        // whether to delete temporary files
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		OutputDir:    "/tmp/podcast",
		ASRConfig:    asr.DefaultConfig(),
		CleanupFiles: true,
	}
}

// TranscriptionResult holds the result of podcast transcription
type TranscriptionResult struct {
	Text         string
	Duration     time.Duration
	EpisodeTitle string
	PodcastTitle string
	EpisodeURL   string
	Error        error
}

// Service handles podcast transcription
type Service struct {
	config     *Config
	asrService *asr.Service
}

// NewService creates a new podcast transcription service
func NewService(config *Config) *Service {
	if config == nil {
		config = DefaultConfig()
	}
	return &Service{
		config:     config,
		asrService: asr.NewService(config.ASRConfig),
	}
}

// ListEpisodes fetches and returns a list of episodes from a podcast RSS URL
func (s *Service) ListEpisodes(ctx context.Context, rssURL string) ([]Episode, error) {
	feed, err := GetPodcastFeed(rssURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch podcast feed: %w", err)
	}

	episodes := GetEpisodes(feed)
	return episodes, nil
}

// TranscribePodcastEpisode downloads a podcast episode and transcribes it
func (s *Service) TranscribePodcastEpisode(ctx context.Context, episodeURL string) (*TranscriptionResult, error) {
	start := time.Now()

	// Ensure output directory exists
	if err := os.MkdirAll(s.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// For direct episode URL transcription, we need to create a minimal episode
	// In a real implementation, you might want to fetch metadata from the RSS
	episode := Episode{
		PodcastTitle: "Unknown Podcast",
		Title:        "Episode",
		FileURL:      episodeURL,
		Date:         time.Now().Format(time.RFC1123Z),
		Number:       1,
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	baseFilename := fmt.Sprintf("episode_%d", timestamp)
	audioPath := filepath.Join(s.config.OutputDir, baseFilename+".mp3")

	// Download episode audio
	if err := s.downloadEpisode(ctx, episode, audioPath); err != nil {
		return nil, fmt.Errorf("failed to download episode: %w", err)
	}

	// Transcribe audio using ASR service
	result, err := s.asrService.TranscribeAudio(ctx, audioPath, s.config.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Cleanup temporary files if configured
	if s.config.CleanupFiles {
		os.Remove(audioPath)
	}

	duration := time.Since(start)
	return &TranscriptionResult{
		Text:         strings.TrimSpace(result.Text),
		Duration:     duration,
		EpisodeTitle: episode.Title,
		PodcastTitle: episode.PodcastTitle,
		EpisodeURL:   episodeURL,
	}, nil
}

// TranscribeEpisodeFromFeed transcribes a specific episode from a podcast feed
func (s *Service) TranscribeEpisodeFromFeed(ctx context.Context, rssURL string, episodeNumber int) (*TranscriptionResult, error) {
	start := time.Now()

	// Get episodes list
	episodes, err := s.ListEpisodes(ctx, rssURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", err)
	}

	// Find the specified episode
	var targetEpisode *Episode
	for _, ep := range episodes {
		if ep.Number == episodeNumber {
			targetEpisode = &ep
			break
		}
	}

	if targetEpisode == nil {
		return nil, fmt.Errorf("episode %d not found", episodeNumber)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(s.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate filename based on episode
	filename := s.generateEpisodeFilename(*targetEpisode)
	audioPath := filepath.Join(s.config.OutputDir, filename)

	// Download episode audio
	if err := s.downloadEpisode(ctx, *targetEpisode, audioPath); err != nil {
		return nil, fmt.Errorf("failed to download episode: %w", err)
	}

	// Transcribe audio using ASR service
	result, err := s.asrService.TranscribeAudio(ctx, audioPath, s.config.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Cleanup temporary files if configured
	if s.config.CleanupFiles {
		os.Remove(audioPath)
	}

	duration := time.Since(start)
	return &TranscriptionResult{
		Text:         strings.TrimSpace(result.Text),
		Duration:     duration,
		EpisodeTitle: targetEpisode.Title,
		PodcastTitle: targetEpisode.PodcastTitle,
		EpisodeURL:   targetEpisode.FileURL,
	}, nil
}

// downloadEpisode downloads a podcast episode to the specified path
func (s *Service) downloadEpisode(ctx context.Context, episode Episode, outputPath string) error {
	// Use the existing DownloadFile logic but with custom path
	// We'll create a temporary episode with the desired path
	tempEpisode := episode
	
	// Save the original download logic but direct to our specified path
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Download file directly to our specified path
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Download the audio file
	client := &http.Client{
		Timeout: 30 * time.Minute, // Long timeout for large podcast files
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", tempEpisode.FileURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download episode: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save episode: %w", err)
	}

	return nil
}

// generateEpisodeFilename creates a filename from episode information
func (s *Service) generateEpisodeFilename(episode Episode) string {
	// Sanitize title for filename
	title := sanitizeFilename(episode.Title)
	if title == "" {
		title = "episode"
	}
	
	// Limit title length
	if len(title) > 50 {
		title = title[:50]
	}
	
	return fmt.Sprintf("%d_%s.mp3", episode.Number, title)
}

// sanitizeFilename removes or replaces characters that are not safe for filenames
func sanitizeFilename(filename string) string {
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

// CheckDependencies verifies that required external tools are available
func CheckDependencies() error {
	return asr.CheckDependencies()
}
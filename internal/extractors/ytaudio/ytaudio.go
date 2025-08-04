package ytaudio

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
	"maai.solutions/gengo/internal/extractors/asr"
)

// Config holds configuration for the YouTube transcription service
type Config struct {
	OutputDir    string
	ASRConfig    *asr.Config // ASR configuration
	CleanupFiles bool        // whether to delete temporary files
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		OutputDir:    "/tmp/ytaudio",
		ASRConfig:    asr.DefaultConfig(),
		CleanupFiles: true,
	}
}

// TranscriptionResult holds the result of transcription
type TranscriptionResult struct {
	Text     string
	Duration time.Duration
	Error    error
}

// Service handles YouTube audio transcription
type Service struct {
	config     *Config
	asrService *asr.Service
}

// NewService creates a new transcription service
func NewService(config *Config) *Service {
	if config == nil {
		config = DefaultConfig()
	}
	return &Service{
		config:     config,
		asrService: asr.NewService(config.ASRConfig),
	}
}

// TranscribeYouTubeVideo downloads a YouTube video, extracts audio, and transcribes it
func (s *Service) TranscribeYouTubeVideo(ctx context.Context, videoURL string) (*TranscriptionResult, error) {
	start := time.Now()

	// Ensure output directory exists
	if err := os.MkdirAll(s.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	baseFilename := fmt.Sprintf("video_%d", timestamp)
	videoPath := filepath.Join(s.config.OutputDir, baseFilename+".mp4") // Default to mp4

	// Download video using github.com/kkdai/youtube
	if err := s.downloadVideo(ctx, videoURL, videoPath); err != nil {
		return nil, fmt.Errorf("failed to download video: %w", err)
	}

	// Transcribe audio using ASR service (handles conversion automatically)
	result, err := s.asrService.TranscribeAudio(ctx, videoPath, s.config.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Cleanup temporary files if configured
	if s.config.CleanupFiles {
		os.Remove(videoPath)
	}

	duration := time.Since(start)
	return &TranscriptionResult{
		Text:     strings.TrimSpace(result.Text),
		Duration: duration,
	}, nil
}

// downloadVideo downloads a YouTube video using github.com/kkdai/youtube library
func (s *Service) downloadVideo(ctx context.Context, videoURL, outputPath string) error {
	client := youtube.Client{}

	video, err := client.GetVideo(videoURL)
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	// Find the best audio format
	formats := video.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return fmt.Errorf("no audio formats found for video")
	}

	// Select the best audio format (prefer highest bitrate)
	var bestFormat *youtube.Format
	for i := range formats {
		format := &formats[i]
		if bestFormat == nil || format.Bitrate > bestFormat.Bitrate {
			bestFormat = format
		}
	}

	if bestFormat == nil {
		return fmt.Errorf("no suitable audio format found")
	}

	// Download the video/audio stream
	stream, _, err := client.GetStream(video, bestFormat)
	if err != nil {
		return fmt.Errorf("failed to get video stream: %w", err)
	}
	defer stream.Close()

	// Create the output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Copy the stream to the file
	_, err = io.Copy(file, stream)
	if err != nil {
		return fmt.Errorf("failed to copy video: %w", err)
	}

	return nil
}

// FindWhisperModel tries to find the whisper model in common locations
func FindWhisperModel(modelName string) string {
	return asr.FindWhisperModel(modelName)
}

// CheckDependencies verifies that required external tools are available
func CheckDependencies() error {
	return asr.CheckDependencies()
}

// TranscribeURL is a convenience function that creates a service and transcribes a URL
func TranscribeURL(ctx context.Context, videoURL string, config *Config) (string, error) {
	service := NewService(config)
	result, err := service.TranscribeYouTubeVideo(ctx, videoURL)
	if err != nil {
		return "", err
	}
	return result.Text, nil
}

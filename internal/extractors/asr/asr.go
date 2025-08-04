package asr

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// Config holds configuration for the ASR service
type Config struct {
	WhisperModel string // path to the whisper model file (e.g., ggml-base.bin)
	Language     string // optional: auto-detect if empty
}

// DefaultConfig returns a default ASR configuration
func DefaultConfig() *Config {
	return &Config{
		WhisperModel: "models/ggml-base.bin", // path to model file
		Language:     "",                     // auto-detect
	}
}

// Result holds the result of ASR transcription
type Result struct {
	Text     string
	Language string // detected or specified language
}

// Service handles automatic speech recognition
type Service struct {
	config *Config
}

// NewService creates a new ASR service
func NewService(config *Config) *Service {
	if config == nil {
		config = DefaultConfig()
	}
	return &Service{config: config}
}

// TranscribeFile transcribes audio from a WAV file
func (s *Service) TranscribeFile(ctx context.Context, audioPath string) (*Result, error) {
	// Check if model file exists
	if _, err := os.Stat(s.config.WhisperModel); err != nil {
		return nil, fmt.Errorf("whisper model file not found: %s", s.config.WhisperModel)
	}

	// Initialize whisper model
	model, err := whisper.New(s.config.WhisperModel)
	if err != nil {
		return nil, fmt.Errorf("failed to load whisper model: %w", err)
	}
	defer model.Close()

	// Create context for processing
	context, err := model.NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create whisper context: %w", err)
	}

	// Set language if specified
	if s.config.Language != "" {
		if err := context.SetLanguage(s.config.Language); err != nil {
			return nil, fmt.Errorf("failed to set language: %w", err)
		}
	}

	// Load and process audio data
	data, err := loadAudioData(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load audio data: %w", err)
	}

	// Process the audio data
	err = context.Process(data, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to process audio: %w", err)
	}

	// Collect all segments
	var text strings.Builder
	for {
		segment, err := context.NextSegment()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get segment: %w", err)
		}
		text.WriteString(segment.Text)
		text.WriteString("\n")
	}

	return &Result{
		Text:     strings.TrimSpace(text.String()),
		Language: s.config.Language, // TODO: get detected language from whisper
	}, nil
}

// TranscribeAudio transcribes audio from any supported format by first converting to WAV
func (s *Service) TranscribeAudio(ctx context.Context, inputPath, tempDir string) (*Result, error) {
	// Generate temporary WAV file path
	wavPath := filepath.Join(tempDir, "temp_audio.wav")

	// Convert audio to WAV format suitable for Whisper
	if err := convertToWAV(ctx, inputPath, wavPath); err != nil {
		return nil, fmt.Errorf("failed to convert audio to WAV: %w", err)
	}
	defer os.Remove(wavPath) // Clean up temp file

	// Transcribe the WAV file
	return s.TranscribeFile(ctx, wavPath)
}

// convertToWAV converts any audio file to 16kHz mono 16-bit WAV using FFmpeg
func convertToWAV(ctx context.Context, inputPath, outputPath string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", inputPath,           // Input file
		"-acodec", "pcm_s16le",    // Output codec: 16-bit PCM
		"-ar", "16000",            // Sample rate: 16kHz (required by whisper)
		"-ac", "1",                // Channels: mono
		"-y",                      // Overwrite output file
		outputPath,                // Output file
	)
	
	// Capture stderr for error reporting
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}
	
	return nil
}

// FindWhisperModel tries to find the whisper model in common locations
func FindWhisperModel(modelName string) string {
	modelFilename := "ggml-" + modelName + ".bin"

	// Common locations where whisper models might be stored
	commonPaths := []string{
		"./models/",
		"./whisper-models/",
		"/usr/local/share/whisper/",
		"/opt/whisper/models/",
		filepath.Join(os.Getenv("HOME"), ".cache/whisper/"),
		filepath.Join(os.Getenv("HOME"), ".local/share/whisper/"),
	}

	for _, basePath := range commonPaths {
		fullPath := filepath.Join(basePath, modelFilename)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	return ""
}

// CheckDependencies verifies that required external tools are available
func CheckDependencies() error {
	// Check if FFmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %w\nPlease install FFmpeg (https://ffmpeg.org/download.html)", err)
	}
	return nil
}
package ytaudio

import (
	"testing"
	"maai.solutions/gengo/internal/extractors/asr"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.OutputDir != "/tmp/ytaudio" {
		t.Errorf("Expected default OutputDir to be '/tmp/ytaudio', got '%s'", config.OutputDir)
	}

	if config.ASRConfig.WhisperModel != "models/ggml-base.bin" {
		t.Errorf("Expected default WhisperModel to be 'models/ggml-base.bin', got '%s'", config.ASRConfig.WhisperModel)
	}

	if config.ASRConfig.Language != "" {
		t.Errorf("Expected default Language to be empty, got '%s'", config.ASRConfig.Language)
	}

	if !config.CleanupFiles {
		t.Error("Expected default CleanupFiles to be true")
	}
}

func TestNewService(t *testing.T) {
	// Test with nil config
	service := NewService(nil)
	if service == nil {
		t.Error("Expected service to not be nil")
	}
	if service.config == nil {
		t.Error("Expected service config to not be nil")
	}

	// Test with custom config
	customConfig := &Config{
		OutputDir: "/custom/path",
		ASRConfig: &asr.Config{
			WhisperModel: "large",
			Language:     "en",
		},
		CleanupFiles: false,
	}

	service2 := NewService(customConfig)
	if service2.config.OutputDir != "/custom/path" {
		t.Errorf("Expected custom OutputDir, got '%s'", service2.config.OutputDir)
	}
}

func TestCheckDependencies(t *testing.T) {
	// This test will pass or fail depending on what's installed on the system
	// In a real scenario, you might want to mock the exec.LookPath function
	err := CheckDependencies()

	// We can't assert much here without knowing the system state
	// But we can at least ensure the function doesn't panic
	if err != nil {
		t.Logf("Dependencies not available: %v", err)
		t.Skip("Skipping test due to missing dependencies")
	} else {
		t.Log("All dependencies are available")
	}
}

// Example of how to test the transcription with a mock or test video
// This is commented out since it requires actual dependencies and network access
/*
func TestTranscribeYouTubeVideo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check dependencies first
	if err := CheckDependencies(); err != nil {
		t.Skipf("Skipping test due to missing dependencies: %v", err)
	}

	service := NewService(&Config{
		OutputDir:    "/tmp/ytaudio_test",
		WhisperModel: "tiny", // Use fastest model for testing
		CleanupFiles: true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Use a very short test video URL
	testURL := "https://www.youtube.com/watch?v=test" // Replace with actual short test video

	result, err := service.TranscribeYouTubeVideo(ctx, testURL)
	if err != nil {
		t.Fatalf("Transcription failed: %v", err)
	}

	if result.Text == "" {
		t.Error("Expected non-empty transcription text")
	}

	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
}
*/

// Benchmark for performance testing (also requires dependencies)
/*
func BenchmarkTranscribeURL(b *testing.B) {
	if err := CheckDependencies(); err != nil {
		b.Skipf("Skipping benchmark due to missing dependencies: %v", err)
	}

	config := &Config{
		OutputDir:    "/tmp/ytaudio_bench",
		WhisperModel: "tiny",
		CleanupFiles: true,
	}

	testURL := "https://www.youtube.com/watch?v=shorttestvideo"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		_, err := TranscribeURL(ctx, testURL, config)
		cancel()

		if err != nil {
			b.Fatalf("Transcription failed: %v", err)
		}
	}
}
*/

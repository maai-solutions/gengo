package asr

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// loadAudioData loads WAV audio file and converts it to float32 samples
func loadAudioData(audioPath string) ([]float32, error) {
	// Open the WAV file
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Skip WAV header (44 bytes)
	header := make([]byte, 44)
	if _, err := file.Read(header); err != nil {
		return nil, fmt.Errorf("failed to read WAV header: %w", err)
	}

	// Verify it's a valid WAV file
	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return nil, fmt.Errorf("invalid WAV file format")
	}

	// Get audio parameters from header
	channels := binary.LittleEndian.Uint16(header[22:24])
	sampleRate := binary.LittleEndian.Uint32(header[24:28])
	bitsPerSample := binary.LittleEndian.Uint16(header[34:36])

	// Verify expected format (16kHz mono 16-bit)
	if channels != 1 || sampleRate != 16000 || bitsPerSample != 16 {
		return nil, fmt.Errorf("unexpected audio format: %d channels, %d Hz, %d bits", channels, sampleRate, bitsPerSample)
	}

	// Read the rest of the file as audio data
	audioData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	// Convert 16-bit samples to float32
	if len(audioData)%2 != 0 {
		return nil, fmt.Errorf("invalid audio data length for 16-bit samples")
	}

	samples := make([]float32, len(audioData)/2)
	for i := 0; i < len(samples); i++ {
		// Convert int16 to float32 normalized to [-1, 1]
		sample := int16(binary.LittleEndian.Uint16(audioData[i*2 : i*2+2]))
		samples[i] = float32(sample) / 32768.0
	}

	return samples, nil
}
// Package utils provides utility functions for AnaStreamer
package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// WAVHeader represents a WAV file header
type WAVHeader struct {
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32
	Format        [4]byte // "WAVE"
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32
}

// PCMToWAV converts raw PCM data to WAV format
func PCMToWAV(pcmData []byte, sampleRate int, channels int, bitsPerSample int) ([]byte, error) {
	dataSize := uint32(len(pcmData))

	header := WAVHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     36 + dataSize,
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, // PCM
		NumChannels:   uint16(channels),
		SampleRate:    uint32(sampleRate),
		ByteRate:      uint32(sampleRate * channels * bitsPerSample / 8),
		BlockAlign:    uint16(channels * bitsPerSample / 8),
		BitsPerSample: uint16(bitsPerSample),
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: dataSize,
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, header); err != nil {
		return nil, fmt.Errorf("failed to write WAV header: %w", err)
	}
	buf.Write(pcmData)

	return buf.Bytes(), nil
}

// SaveWAV saves PCM data as a WAV file
func SaveWAV(filename string, pcmData []byte, sampleRate int, channels int, bitsPerSample int) error {
	wavData, err := PCMToWAV(pcmData, sampleRate, channels, bitsPerSample)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, wavData, 0644)
}

// CalculateRMS calculates the Root Mean Square of audio samples
func CalculateRMS(samples []int16) float64 {
	if len(samples) == 0 {
		return 0
	}

	var sum float64
	for _, sample := range samples {
		sum += float64(sample) * float64(sample)
	}

	return sum / float64(len(samples))
}

// BytesToInt16 converts byte slice to int16 slice (little endian)
func BytesToInt16(data []byte) []int16 {
	if len(data) == 0 {
		return []int16{}
	}

	// Ensure even number of bytes for proper int16 conversion
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}

	samples := make([]int16, len(data)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
	}
	return samples
}

// Int16ToBytes converts int16 slice to byte slice (little endian)
func Int16ToBytes(samples []int16) []byte {
	data := make([]byte, len(samples)*2)
	for i, sample := range samples {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(sample))
	}
	return data
}

// NormalizeAudio normalizes audio samples to a target peak level
func NormalizeAudio(samples []int16, targetPeak int16) []int16 {
	if len(samples) == 0 {
		return samples
	}

	// Find current peak
	var maxPeak int16
	for _, s := range samples {
		if s > maxPeak {
			maxPeak = s
		}
		if -s > maxPeak {
			maxPeak = -s
		}
	}

	if maxPeak == 0 {
		return samples
	}

	// Calculate scale factor
	scale := float64(targetPeak) / float64(maxPeak)

	// Apply normalization
	normalized := make([]int16, len(samples))
	for i, s := range samples {
		normalized[i] = int16(float64(s) * scale)
	}

	return normalized
}

// GetBinaryPath returns the full path to a binary, adding .exe on Windows
func GetBinaryPath(basePath string) string {
	if runtime.GOOS == "windows" && filepath.Ext(basePath) == "" {
		return basePath + ".exe"
	}
	return basePath
}

// BinaryExists checks if a binary exists and is executable
func BinaryExists(path string) bool {
	path = GetBinaryPath(path)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// RunCommand executes a command and returns its output
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunCommandWithInput executes a command with stdin input
func RunCommandWithInput(input []byte, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = bytes.NewReader(input)

	return cmd.Output()
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// GetTempFilePath returns a path for a temporary file
func GetTempFilePath(prefix, suffix string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d%s", prefix, os.Getpid(), suffix))
}

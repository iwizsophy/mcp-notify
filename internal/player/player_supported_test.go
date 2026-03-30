//go:build windows || darwin || linux

package player

import (
	"path/filepath"
	"testing"

	"github.com/ebitengine/oto/v3"
)

func TestDecodeAudioFileWAV(t *testing.T) {
	t.Parallel()

	soundPath := filepath.Join("..", "..", "sounds", "complete.wav")
	clip, err := decodeAudioFile(soundPath)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(clip.data) == 0 {
		t.Fatalf("expected decoded PCM data")
	}
	if clip.config.sampleRate <= 0 {
		t.Fatalf("expected positive sample rate, got %d", clip.config.sampleRate)
	}
	if clip.config.channelCount != 1 && clip.config.channelCount != 2 {
		t.Fatalf("expected mono or stereo, got %d", clip.config.channelCount)
	}
	if clip.config.format != oto.FormatSignedInt16LE {
		t.Fatalf("expected signed 16-bit PCM, got %v", clip.config.format)
	}
}

func TestDecodeAudioFileMP3(t *testing.T) {
	t.Parallel()

	soundPath := filepath.Join("..", "..", "sounds", "alerts", "sample.mp3")
	clip, err := decodeAudioFile(soundPath)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(clip.data) == 0 {
		t.Fatalf("expected decoded PCM data")
	}
	if clip.config.sampleRate <= 0 {
		t.Fatalf("expected positive sample rate, got %d", clip.config.sampleRate)
	}
	if clip.config.channelCount != 2 {
		t.Fatalf("expected stereo MP3 output, got %d", clip.config.channelCount)
	}
	if clip.config.format != oto.FormatSignedInt16LE {
		t.Fatalf("expected signed 16-bit PCM, got %v", clip.config.format)
	}
}

func TestConvertSampleToSignedInt16(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		sample         int
		sourceBitDepth int
		want           int16
	}{
		{name: "unsigned 8-bit silence", sample: 128, sourceBitDepth: 8, want: 0},
		{name: "unsigned 8-bit minimum", sample: 0, sourceBitDepth: 8, want: -32768},
		{name: "signed 16-bit passthrough", sample: -12345, sourceBitDepth: 16, want: -12345},
		{name: "signed 24-bit downsample", sample: 8388352, sourceBitDepth: 24, want: 32767},
		{name: "signed 32-bit downsample", sample: -2147483648, sourceBitDepth: 32, want: -32768},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := convertSampleToSignedInt16(tc.sample, tc.sourceBitDepth)
			if got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

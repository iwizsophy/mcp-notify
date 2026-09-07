//go:build windows || darwin || linux

package player

import (
	"encoding/binary"
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

func TestNormalizeDecodedAudioResamplesMonoToStereo(t *testing.T) {
	t.Parallel()

	source := make([]byte, 4)
	binary.LittleEndian.PutUint16(source[0:], uint16(int16(1000)))
	binary.LittleEndian.PutUint16(source[2:], uint16(int16(2000)))

	clip, err := normalizeDecodedAudio(&decodedAudio{
		data: source,
		config: audioConfig{
			sampleRate:   24000,
			channelCount: 1,
			format:       oto.FormatSignedInt16LE,
		},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if clip.config != playbackConfig {
		t.Fatalf("expected playback config, got %+v", clip.config)
	}
	if len(clip.data) != 16 {
		t.Fatalf("expected four stereo frames, got %d bytes", len(clip.data))
	}
	for frame := 0; frame < 4; frame++ {
		left := readPCM16Sample(clip.data, frame, 0, 2)
		right := readPCM16Sample(clip.data, frame, 1, 2)
		if left != right {
			t.Fatalf("expected duplicated mono sample at frame %d, got %d and %d", frame, left, right)
		}
	}
}

func TestNormalizeDecodedAudioRejectsUnsafeSampleRate(t *testing.T) {
	t.Parallel()

	_, err := normalizeDecodedAudio(&decodedAudio{
		data: []byte{0, 0},
		config: audioConfig{
			sampleRate:   1,
			channelCount: 1,
			format:       oto.FormatSignedInt16LE,
		},
	})
	if err == nil {
		t.Fatalf("expected unsafe sample rate to be rejected")
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

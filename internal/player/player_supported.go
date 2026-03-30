//go:build windows || darwin || linux

package player

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"

	"mcp-notify/internal/validation"
)

type Player struct {
	mu         sync.Mutex
	context    *oto.Context
	ready      chan struct{}
	config     audioConfig
	loadedPath string
	clip       *decodedAudio
}

type audioConfig struct {
	sampleRate   int
	channelCount int
	format       oto.Format
}

type decodedAudio struct {
	data   []byte
	config audioConfig
}

type bufferedAudioReader struct {
	reader *bytes.Reader
	done   chan struct{}
	once   sync.Once
}

func New() *Player {
	return &Player{}
}

func (p *Player) Play(ctx context.Context, soundPath string, wait bool) *validation.AppError {
	if !wait {
		return spawnDetachedPlayback(soundPath)
	}

	clip, err := p.loadDecodedAudio(soundPath)
	if err != nil {
		return err
	}

	audioContext, ready, err := p.ensureContext(clip.config)
	if err != nil {
		return err
	}
	if err := waitForContextReady(ctx, ready); err != nil {
		return err
	}

	reader := newBufferedAudioReader(clip.data)
	audioPlayer := audioContext.NewPlayer(reader)
	audioPlayer.Play()

	return waitForPlayback(ctx, audioPlayer, reader.done)
}

func (p *Player) loadDecodedAudio(soundPath string) (*decodedAudio, *validation.AppError) {
	p.mu.Lock()
	if p.clip != nil {
		if p.loadedPath != soundPath {
			p.mu.Unlock()
			return nil, validation.NewAppError(
				"configured sound path changed after audio initialization",
				"restart the MCP server after changing the configured sound path",
			)
		}
		clip := p.clip
		p.mu.Unlock()
		return clip, nil
	}
	p.mu.Unlock()

	clip, appErr := decodeAudioFile(soundPath)
	if appErr != nil {
		return nil, appErr
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.clip != nil {
		if p.loadedPath != soundPath {
			return nil, validation.NewAppError(
				"configured sound path changed after audio initialization",
				"restart the MCP server after changing the configured sound path",
			)
		}
		return p.clip, nil
	}
	p.loadedPath = soundPath
	p.clip = clip
	return clip, nil
}

func (p *Player) ensureContext(config audioConfig) (*oto.Context, chan struct{}, *validation.AppError) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.context != nil {
		if p.config != config {
			return nil, nil, validation.NewAppError(
				"configured sound format changed after audio initialization",
				"restart the MCP server after replacing the configured sound with a file that uses a different sample rate, channel count, or encoding",
			)
		}
		return p.context, p.ready, nil
	}

	contextOptions := &oto.NewContextOptions{
		SampleRate:   config.sampleRate,
		ChannelCount: config.channelCount,
		Format:       config.format,
	}

	audioContext, ready, err := oto.NewContext(contextOptions)
	if err != nil {
		return nil, nil, validation.NewAppError(
			"failed to initialize audio output",
			err.Error(),
		)
	}

	p.context = audioContext
	p.ready = ready
	p.config = config
	return p.context, p.ready, nil
}

func waitForContextReady(ctx context.Context, ready <-chan struct{}) *validation.AppError {
	select {
	case <-ready:
		return nil
	case <-ctx.Done():
		return validation.NewAppError(
			"audio playback was cancelled",
			ctx.Err().Error(),
		)
	}
}

func waitForPlayback(ctx context.Context, player *oto.Player, done <-chan struct{}) *validation.AppError {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	readerExhausted := false
	for {
		if err := player.Err(); err != nil {
			return validation.NewAppError(
				"failed to play the requested sound",
				err.Error(),
			)
		}

		select {
		case <-done:
			readerExhausted = true
		default:
		}

		if readerExhausted && !player.IsPlaying() && player.BufferedSize() == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			player.Pause()
			return validation.NewAppError(
				"audio playback was cancelled",
				ctx.Err().Error(),
			)
		case <-ticker.C:
		}
	}
}

func decodeAudioFile(soundPath string) (*decodedAudio, *validation.AppError) {
	switch extension := filepath.Ext(soundPath); extension {
	case ".mp3":
		return decodeMP3File(soundPath)
	case ".wav":
		return decodeWAVFile(soundPath)
	default:
		return nil, validation.NewAppError(
			fmt.Sprintf("unsupported audio format: %s", extension),
			"currently supported audio formats are .wav and .mp3",
		)
	}
}

func decodeMP3File(soundPath string) (*decodedAudio, *validation.AppError) {
	file, err := os.Open(soundPath)
	if err != nil {
		return nil, validation.NewAppError(
			"failed to open configured sound file",
			err.Error(),
		)
	}
	defer file.Close()

	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return nil, validation.NewAppError(
			"failed to decode the configured MP3 file",
			err.Error(),
		)
	}

	data, err := io.ReadAll(decoder)
	if err != nil {
		return nil, validation.NewAppError(
			"failed to decode the configured MP3 file",
			err.Error(),
		)
	}
	if len(data) == 0 {
		return nil, validation.NewAppError(
			"configured MP3 file decoded to an empty audio stream",
			fmt.Sprintf("resolved path: %s", soundPath),
		)
	}

	return &decodedAudio{
		data: data,
		config: audioConfig{
			sampleRate:   decoder.SampleRate(),
			channelCount: 2,
			format:       oto.FormatSignedInt16LE,
		},
	}, nil
}

func decodeWAVFile(soundPath string) (*decodedAudio, *validation.AppError) {
	file, err := os.Open(soundPath)
	if err != nil {
		return nil, validation.NewAppError(
			"failed to open configured sound file",
			err.Error(),
		)
	}
	defer file.Close()

	decoder := wav.NewDecoder(file)
	decoder.ReadInfo()
	if err := decoder.Err(); err != nil {
		return nil, validation.NewAppError(
			"failed to decode the configured WAV file",
			err.Error(),
		)
	}

	if decoder.WavAudioFormat != 1 {
		return nil, validation.NewAppError(
			"unsupported WAV encoding",
			fmt.Sprintf("resolved path: %s, audio format: %d (only PCM WAV is supported)", soundPath, decoder.WavAudioFormat),
		)
	}

	if decoder.NumChans != 1 && decoder.NumChans != 2 {
		return nil, validation.NewAppError(
			"unsupported WAV channel count",
			fmt.Sprintf("resolved path: %s, channel count: %d (only mono and stereo WAV files are supported)", soundPath, decoder.NumChans),
		)
	}

	pcmBuffer, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, validation.NewAppError(
			"failed to decode the configured WAV file",
			err.Error(),
		)
	}

	data, appErr := intBufferToSignedInt16LE(pcmBuffer)
	if appErr != nil {
		return nil, appErr
	}
	if len(data) == 0 {
		return nil, validation.NewAppError(
			"configured WAV file decoded to an empty audio stream",
			fmt.Sprintf("resolved path: %s", soundPath),
		)
	}

	return &decodedAudio{
		data: data,
		config: audioConfig{
			sampleRate:   int(decoder.SampleRate),
			channelCount: int(decoder.NumChans),
			format:       oto.FormatSignedInt16LE,
		},
	}, nil
}

func intBufferToSignedInt16LE(buffer *audio.IntBuffer) ([]byte, *validation.AppError) {
	if buffer == nil || buffer.Format == nil {
		return nil, validation.NewAppError(
			"failed to decode the configured WAV file",
			"decoder returned an empty PCM buffer",
		)
	}

	if buffer.Format.NumChannels != 1 && buffer.Format.NumChannels != 2 {
		return nil, validation.NewAppError(
			"unsupported WAV channel count",
			fmt.Sprintf("channel count: %d (only mono and stereo WAV files are supported)", buffer.Format.NumChannels),
		)
	}

	output := make([]byte, len(buffer.Data)*2)
	for i, sample := range buffer.Data {
		value := convertSampleToSignedInt16(sample, buffer.SourceBitDepth)
		binary.LittleEndian.PutUint16(output[i*2:], uint16(value))
	}
	return output, nil
}

func convertSampleToSignedInt16(sample, sourceBitDepth int) int16 {
	switch {
	case sourceBitDepth <= 0:
		return 0
	case sourceBitDepth == 8:
		return clampInt16((sample - 128) << 8)
	case sourceBitDepth > 16:
		return clampInt16(sample >> (sourceBitDepth - 16))
	case sourceBitDepth < 16:
		return clampInt16(sample << (16 - sourceBitDepth))
	default:
		return clampInt16(sample)
	}
}

func clampInt16(value int) int16 {
	if value > 32767 {
		return 32767
	}
	if value < -32768 {
		return -32768
	}
	return int16(value)
}

func newBufferedAudioReader(data []byte) *bufferedAudioReader {
	return &bufferedAudioReader{
		reader: bytes.NewReader(data),
		done:   make(chan struct{}),
	}
}

func (r *bufferedAudioReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if err == io.EOF {
		r.once.Do(func() {
			close(r.done)
		})
	}
	return n, err
}

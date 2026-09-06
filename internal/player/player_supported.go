//go:build windows || darwin || linux

package player

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
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

var playbackConfig = audioConfig{
	sampleRate:   48000,
	channelCount: 2,
	format:       oto.FormatSignedInt16LE,
}

const (
	minSupportedSampleRate = 8000
	maxSupportedSampleRate = 192000
	maxNormalizedPCMBytes  = 256 << 20
)

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
	return p.playDecoded(ctx, clip)
}

func (p *Player) PlayWAV(ctx context.Context, wavData []byte) *validation.AppError {
	if len(wavData) == 0 {
		return validation.NewAppError(
			"synthesized WAV data is empty",
			"the speech synthesizer returned no audio data",
		)
	}

	clip, err := decodeWAV(bytes.NewReader(wavData), "synthesized WAV response")
	if err != nil {
		return err
	}
	clip, err = normalizeDecodedAudio(clip)
	if err != nil {
		return err
	}

	return p.playDecoded(ctx, clip)
}

func (p *Player) playDecoded(ctx context.Context, clip *decodedAudio) *validation.AppError {

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
		// The player keeps one decoded clip and one oto context alive for the
		// server lifetime, so hot-swapping the configured file requires restart.
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
	clip, appErr = normalizeDecodedAudio(clip)
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

		// Wait until the reader is exhausted and oto drains its internal buffer.
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

	return decodeWAV(file, soundPath)
}

func decodeWAV(reader io.ReadSeeker, source string) (*decodedAudio, *validation.AppError) {
	decoder := wav.NewDecoder(reader)
	decoder.ReadInfo()
	if err := decoder.Err(); err != nil {
		return nil, validation.NewAppError(
			"failed to decode WAV audio",
			err.Error(),
		)
	}

	if decoder.WavAudioFormat != 1 {
		return nil, validation.NewAppError(
			"unsupported WAV encoding",
			fmt.Sprintf("source: %s, audio format: %d (only PCM WAV is supported)", source, decoder.WavAudioFormat),
		)
	}

	if decoder.NumChans != 1 && decoder.NumChans != 2 {
		return nil, validation.NewAppError(
			"unsupported WAV channel count",
			fmt.Sprintf("source: %s, channel count: %d (only mono and stereo WAV files are supported)", source, decoder.NumChans),
		)
	}

	pcmBuffer, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, validation.NewAppError(
			"failed to decode WAV audio",
			err.Error(),
		)
	}

	data, appErr := intBufferToSignedInt16LE(pcmBuffer)
	if appErr != nil {
		return nil, appErr
	}
	if len(data) == 0 {
		return nil, validation.NewAppError(
			"WAV audio decoded to an empty stream",
			fmt.Sprintf("source: %s", source),
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

func normalizeDecodedAudio(clip *decodedAudio) (*decodedAudio, *validation.AppError) {
	if clip == nil {
		return nil, validation.NewAppError(
			"failed to normalize audio",
			"decoder returned no audio clip",
		)
	}
	if clip.config.format != oto.FormatSignedInt16LE {
		return nil, validation.NewAppError(
			"failed to normalize audio",
			"only signed 16-bit little-endian PCM can be normalized",
		)
	}
	if clip.config.sampleRate < minSupportedSampleRate || clip.config.sampleRate > maxSupportedSampleRate ||
		(clip.config.channelCount != 1 && clip.config.channelCount != 2) {
		return nil, validation.NewAppError(
			"failed to normalize audio",
			fmt.Sprintf("invalid source format: %d Hz, %d channels", clip.config.sampleRate, clip.config.channelCount),
		)
	}
	frameSize := clip.config.channelCount * 2
	if len(clip.data)%frameSize != 0 {
		return nil, validation.NewAppError(
			"failed to normalize audio",
			"PCM data does not contain a whole number of audio frames",
		)
	}

	sourceFrames := len(clip.data) / frameSize
	if sourceFrames == 0 {
		return nil, validation.NewAppError(
			"failed to normalize audio",
			"PCM data contains no audio frames",
		)
	}
	targetFrames64 := (int64(sourceFrames)*int64(playbackConfig.sampleRate) + int64(clip.config.sampleRate) - 1) / int64(clip.config.sampleRate)
	targetBytes64 := targetFrames64 * int64(playbackConfig.channelCount) * 2
	if targetFrames64 <= 0 || targetBytes64 > maxNormalizedPCMBytes || targetBytes64 > int64(^uint(0)>>1) {
		return nil, validation.NewAppError(
			"failed to normalize audio",
			fmt.Sprintf("normalized PCM data would exceed the %d byte limit", maxNormalizedPCMBytes),
		)
	}
	targetFrames := int(targetFrames64)
	targetData := make([]byte, targetFrames*playbackConfig.channelCount*2)

	for targetFrame := 0; targetFrame < targetFrames; targetFrame++ {
		sourcePosition := float64(targetFrame) * float64(clip.config.sampleRate) / float64(playbackConfig.sampleRate)
		firstFrame := int(sourcePosition)
		if firstFrame >= sourceFrames {
			firstFrame = sourceFrames - 1
		}
		secondFrame := firstFrame + 1
		if secondFrame >= sourceFrames {
			secondFrame = firstFrame
		}
		fraction := sourcePosition - float64(firstFrame)

		for targetChannel := 0; targetChannel < playbackConfig.channelCount; targetChannel++ {
			sourceChannel := targetChannel
			if clip.config.channelCount == 1 {
				sourceChannel = 0
			}
			firstSample := readPCM16Sample(clip.data, firstFrame, sourceChannel, clip.config.channelCount)
			secondSample := readPCM16Sample(clip.data, secondFrame, sourceChannel, clip.config.channelCount)
			interpolated := float64(firstSample) + (float64(secondSample)-float64(firstSample))*fraction
			targetOffset := (targetFrame*playbackConfig.channelCount + targetChannel) * 2
			binary.LittleEndian.PutUint16(targetData[targetOffset:], uint16(int16(math.Round(interpolated))))
		}
	}

	return &decodedAudio{
		data:   targetData,
		config: playbackConfig,
	}, nil
}

func readPCM16Sample(data []byte, frame, channel, channelCount int) int16 {
	offset := (frame*channelCount + channel) * 2
	return int16(binary.LittleEndian.Uint16(data[offset:]))
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
		// 8-bit PCM is unsigned, so it must be centered around zero first.
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

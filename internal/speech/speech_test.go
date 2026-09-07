package speech

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"mcp-notify/internal/validation"
)

type fakeSynthesizer struct {
	data []byte
	err  *validation.AppError
}

func (f *fakeSynthesizer) Synthesize(_ context.Context, _ Request) ([]byte, *validation.AppError) {
	return f.data, f.err
}

type fakeWAVPlayer struct {
	played chan []byte
	err    *validation.AppError
}

func (f *fakeWAVPlayer) PlayWAV(_ context.Context, data []byte) *validation.AppError {
	if f.played != nil {
		f.played <- data
	}
	return f.err
}

func TestServiceSpeakWaitsForPlayback(t *testing.T) {
	t.Parallel()

	played := make(chan []byte, 1)
	service := NewService(
		&fakeSynthesizer{data: []byte("wav")},
		&fakeWAVPlayer{played: played},
		log.New(io.Discard, "", 0),
	)

	if err := service.Speak(context.Background(), Request{Text: "hello", Speaker: 3}, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := <-played; string(got) != "wav" {
		t.Fatalf("unexpected audio data: %q", got)
	}
}

func TestServiceSpeakStartsPlaybackAsynchronously(t *testing.T) {
	t.Parallel()

	played := make(chan []byte, 1)
	service := NewService(
		&fakeSynthesizer{data: []byte("wav")},
		&fakeWAVPlayer{played: played},
		log.New(io.Discard, "", 0),
	)

	if err := service.Speak(context.Background(), Request{Text: "hello", Speaker: 3}, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case got := <-played:
		if string(got) != "wav" {
			t.Fatalf("unexpected audio data: %q", got)
		}
	case <-time.After(time.Second):
		t.Fatalf("asynchronous playback was not started")
	}
}

func TestServiceSpeakReturnsSynthesisError(t *testing.T) {
	t.Parallel()

	want := validation.NewAppError("synthesis failed", "test")
	service := NewService(&fakeSynthesizer{err: want}, &fakeWAVPlayer{}, log.New(io.Discard, "", 0))
	if got := service.Speak(context.Background(), Request{}, true); got != want {
		t.Fatalf("expected synthesis error, got %v", got)
	}
}

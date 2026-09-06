package speech

import (
	"context"
	"log"

	"mcp-notify/internal/validation"
)

type Request struct {
	Text            string
	Speaker         int
	SpeedScale      *float64
	PitchScale      *float64
	IntonationScale *float64
	VolumeScale     *float64
}

type Synthesizer interface {
	Synthesize(ctx context.Context, request Request) ([]byte, *validation.AppError)
}

type WAVPlayer interface {
	PlayWAV(ctx context.Context, wavData []byte) *validation.AppError
}

type Service struct {
	synthesizer Synthesizer
	player      WAVPlayer
	logger      *log.Logger
}

func NewService(synthesizer Synthesizer, player WAVPlayer, logger *log.Logger) *Service {
	return &Service{
		synthesizer: synthesizer,
		player:      player,
		logger:      logger,
	}
}

func (s *Service) Speak(ctx context.Context, request Request, wait bool) *validation.AppError {
	wavData, err := s.synthesizer.Synthesize(ctx, request)
	if err != nil {
		return err
	}

	if wait {
		return s.player.PlayWAV(ctx, wavData)
	}

	go func() {
		if err := s.player.PlayWAV(context.Background(), wavData); err != nil && s.logger != nil {
			s.logger.Printf("asynchronous speech playback failed: %s (%s)", err.Message, err.Details)
		}
	}()

	return nil
}

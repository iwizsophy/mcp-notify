//go:build !windows && !darwin && !linux

package player

import (
	"context"
	"runtime"

	"mcp-notify/internal/validation"
)

type Player struct{}

func New() *Player {
	return &Player{}
}

func (p *Player) Play(_ context.Context, _ string, _ bool) *validation.AppError {
	return validation.NewAppError(
		"unsupported operating system",
		"this build currently supports Windows, macOS, and Linux; detected GOOS="+runtime.GOOS,
	)
}

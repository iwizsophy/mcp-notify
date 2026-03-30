//go:build windows || darwin || linux

package player

import (
	"os"
	"os/exec"
	"path/filepath"

	"mcp-notify/internal/validation"
)

func spawnDetachedPlayback(soundPath string) *validation.AppError {
	executablePath, err := os.Executable()
	if err != nil {
		return validation.NewAppError(
			"failed to locate the playback executable",
			err.Error(),
		)
	}

	cmd := buildDetachedPlaybackCommand(executablePath, soundPath)
	if err := startDetachedCommand(cmd); err != nil {
		return validation.NewAppError(
			"failed to start detached audio playback",
			err.Error(),
		)
	}

	return nil
}

func buildDetachedPlaybackCommand(executablePath, soundPath string) *exec.Cmd {
	cmd := exec.Command(executablePath, "--play-once-path", soundPath)
	cmd.Dir = filepath.Dir(executablePath)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = detachedProcessAttributes()
	return cmd
}

func startDetachedCommand(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

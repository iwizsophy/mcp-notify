//go:build windows || darwin || linux

package player

import (
	"path/filepath"
	"testing"
)

func TestBuildDetachedPlaybackCommand(t *testing.T) {
	t.Parallel()

	executablePath := filepath.Join("C:\\", "mcp", "mcp-notify", "mcp-notify.exe")
	soundPath := filepath.Join("C:\\", "mcp", "mcp-notify", "sounds", "継続作業.wav")

	cmd := buildDetachedPlaybackCommand(executablePath, soundPath)

	if got := cmd.Path; got != executablePath {
		t.Fatalf("expected executable path %q, got %q", executablePath, got)
	}
	if got, want := cmd.Args, []string{executablePath, "--play-once-path", soundPath}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("unexpected command args: %#v", got)
	}
	if got := cmd.Dir; got != filepath.Dir(executablePath) {
		t.Fatalf("expected working directory %q, got %q", filepath.Dir(executablePath), got)
	}
	if cmd.SysProcAttr == nil {
		t.Fatalf("expected detached process attributes to be configured")
	}
}

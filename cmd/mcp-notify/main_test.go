package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigAcceptsSplitSoundFlag(t *testing.T) {
	t.Parallel()

	cfg, err := parseConfig([]string{"--sound", "作業終了.wav", "--wait=false"})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if cfg.soundPath != "作業終了.wav" {
		t.Fatalf("expected soundPath to be preserved, got %q", cfg.soundPath)
	}
	if cfg.wait {
		t.Fatalf("expected wait=false")
	}
}

func TestParseConfigRejectsCombinedSoundFlagToken(t *testing.T) {
	t.Parallel()

	_, err := parseConfig([]string{"--sound 作業終了.wav"})
	if err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestParseConfigAcceptsPlayOncePath(t *testing.T) {
	t.Parallel()

	cfg, err := parseConfig([]string{"--play-once-path", `C:\mcp\mcp-notify\sounds\継続作業.wav`})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if cfg.playOncePath != `C:\mcp\mcp-notify\sounds\継続作業.wav` {
		t.Fatalf("expected playOncePath to be preserved, got %q", cfg.playOncePath)
	}
}

func TestParseConfigAcceptsPlayOnce(t *testing.T) {
	t.Parallel()

	cfg, err := parseConfig([]string{"--play-once", "complete.wav", "--wait=false"})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if cfg.playOnceSound != "complete.wav" {
		t.Fatalf("expected playOnceSound to be preserved, got %q", cfg.playOnceSound)
	}
	if cfg.wait {
		t.Fatalf("expected wait=false")
	}
}

func TestParseConfigAcceptsServerNameAndToolPrefix(t *testing.T) {
	t.Parallel()

	cfg, err := parseConfig([]string{"--server-name", "notify-complete", "--tool-prefix", "complete_"})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if cfg.serverName != "notify-complete" {
		t.Fatalf("expected serverName to be preserved, got %q", cfg.serverName)
	}
	if cfg.toolPrefix != "complete_" {
		t.Fatalf("expected toolPrefix to be preserved, got %q", cfg.toolPrefix)
	}
}

func TestParseConfigRejectsPlayOnceAndSoundTogether(t *testing.T) {
	t.Parallel()

	_, err := parseConfig([]string{"--play-once", "complete.wav", "--sound", "other.wav"})
	if err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestResolveBaseDirFromExecutablePrefersExecutableDirSounds(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	exeDir := filepath.Join(tmpDir, "app")
	if err := os.MkdirAll(filepath.Join(exeDir, "sounds"), 0o755); err != nil {
		t.Fatalf("mkdir sounds: %v", err)
	}

	baseDir, err := resolveBaseDirFromExecutable(filepath.Join(tmpDir, "work"), filepath.Join(exeDir, "mcp-notify.exe"))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if baseDir != exeDir {
		t.Fatalf("expected executable dir %q, got %q", exeDir, baseDir)
	}
}

func TestResolveBaseDirFromExecutableFallsBackToExecutableParentDirSounds(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	exeDir := filepath.Join(tmpDir, "bin")
	parentDir := filepath.Dir(exeDir)
	if err := os.MkdirAll(filepath.Join(parentDir, "sounds"), 0o755); err != nil {
		t.Fatalf("mkdir sounds: %v", err)
	}

	baseDir, err := resolveBaseDirFromExecutable(filepath.Join(tmpDir, "work"), filepath.Join(exeDir, "mcp-notify.exe"))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if baseDir != parentDir {
		t.Fatalf("expected parent dir %q, got %q", parentDir, baseDir)
	}
}

func TestResolveBaseDirFromExecutableFallsBackToWorkingDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	workingDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workingDir, 0o755); err != nil {
		t.Fatalf("mkdir work: %v", err)
	}

	baseDir, err := resolveBaseDirFromExecutable(workingDir, filepath.Join(tmpDir, "bin", "mcp-notify.exe"))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if baseDir != workingDir {
		t.Fatalf("expected working dir %q, got %q", workingDir, baseDir)
	}
}

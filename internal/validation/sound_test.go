package validation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSoundValidatorValidateConfiguredPathResolvesRelativePath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	soundsDir := filepath.Join(tmpDir, "sounds")
	if err := os.MkdirAll(soundsDir, 0o755); err != nil {
		t.Fatalf("mkdir sounds: %v", err)
	}
	target := filepath.Join(soundsDir, "notice.wav")
	if err := os.WriteFile(target, []byte("wav"), 0o644); err != nil {
		t.Fatalf("write sound file: %v", err)
	}

	validator := NewSoundValidator(tmpDir)
	resolved, err := validator.ValidateConfiguredPath("notice.wav")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if resolved.ResolvedPath != target {
		t.Fatalf("expected %q, got %q", target, resolved.ResolvedPath)
	}
}

func TestSoundValidatorValidateConfiguredPathRejectsEmptyPath(t *testing.T) {
	t.Parallel()

	validator := NewSoundValidator(t.TempDir())
	_, err := validator.ValidateConfiguredPath("")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Message != "configured soundPath must not be empty" {
		t.Fatalf("unexpected error: %s", err.Message)
	}
}

func TestSoundValidatorValidateConfiguredPathRejectsMissingFile(t *testing.T) {
	t.Parallel()

	validator := NewSoundValidator(t.TempDir())
	_, err := validator.ValidateConfiguredPath("missing.wav")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Message != "configured sound file does not exist" {
		t.Fatalf("unexpected error: %s", err.Message)
	}
}

func TestSoundValidatorValidateConfiguredPathRejectsUnsupportedExtension(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	soundsDir := filepath.Join(tmpDir, "sounds")
	if err := os.MkdirAll(soundsDir, 0o755); err != nil {
		t.Fatalf("mkdir sounds: %v", err)
	}
	target := filepath.Join(soundsDir, "notice.txt")
	if err := os.WriteFile(target, []byte("txt"), 0o644); err != nil {
		t.Fatalf("write sound file: %v", err)
	}

	validator := NewSoundValidator(tmpDir)
	_, err := validator.ValidateConfiguredPath("notice.txt")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Message != "unsupported audio format: .txt" {
		t.Fatalf("unexpected error: %s", err.Message)
	}
}

func TestSoundValidatorValidateConfiguredPathAcceptsMP3(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	soundsDir := filepath.Join(tmpDir, "sounds")
	if err := os.MkdirAll(soundsDir, 0o755); err != nil {
		t.Fatalf("mkdir sounds: %v", err)
	}
	target := filepath.Join(soundsDir, "notice.mp3")
	if err := os.WriteFile(target, []byte("mp3"), 0o644); err != nil {
		t.Fatalf("write sound file: %v", err)
	}

	validator := NewSoundValidator(tmpDir)
	resolved, err := validator.ValidateConfiguredPath("notice.mp3")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if resolved.ResolvedPath != target {
		t.Fatalf("expected %q, got %q", target, resolved.ResolvedPath)
	}
}

func TestSoundValidatorValidateConfiguredPathRejectsAbsolutePath(t *testing.T) {
	t.Parallel()

	validator := NewSoundValidator(t.TempDir())
	absolutePath := filepath.Join(t.TempDir(), "notice.wav")
	_, err := validator.ValidateConfiguredPath(absolutePath)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Message != "absolute paths are not allowed" {
		t.Fatalf("unexpected error: %s", err.Message)
	}
}

func TestSoundValidatorValidateConfiguredPathRejectsPathTraversal(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	soundsDir := filepath.Join(tmpDir, "sounds")
	if err := os.MkdirAll(soundsDir, 0o755); err != nil {
		t.Fatalf("mkdir sounds: %v", err)
	}
	outside := filepath.Join(tmpDir, "escape.wav")
	if err := os.WriteFile(outside, []byte("wav"), 0o644); err != nil {
		t.Fatalf("write outside sound file: %v", err)
	}

	validator := NewSoundValidator(tmpDir)
	_, err := validator.ValidateConfiguredPath(filepath.Join("..", "escape.wav"))
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Message != "configured soundPath must stay within the sounds directory" {
		t.Fatalf("unexpected error: %s", err.Message)
	}
}

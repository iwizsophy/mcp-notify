package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"mcp-notify/internal/validation"
)

type fakePlayer struct {
	lastPath string
	lastWait bool
	err      *validation.AppError
}

func (f *fakePlayer) Play(_ context.Context, soundPath string, wait bool) *validation.AppError {
	f.lastPath = soundPath
	f.lastWait = wait
	return f.err
}

func TestPlayNotificationSoundToolCallSuccessUsesConfiguredDefaults(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	soundsDir := filepath.Join(tmpDir, "sounds")
	if err := os.MkdirAll(soundsDir, 0o755); err != nil {
		t.Fatalf("mkdir sounds: %v", err)
	}
	soundPath := filepath.Join(soundsDir, "complete.wav")
	if err := os.WriteFile(soundPath, []byte("wav"), 0o644); err != nil {
		t.Fatalf("write sound file: %v", err)
	}

	player := &fakePlayer{}
	tool := NewPlayNotificationSoundTool(validation.NewSoundValidator(tmpDir), player, "complete.wav", true, "")

	result, rpcErr := tool.Call(context.Background(), nil)
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %+v", rpcErr)
	}
	if result.IsError {
		t.Fatalf("expected success result, got error")
	}
	if player.lastPath != soundPath {
		t.Fatalf("expected resolved path %q, got %q", soundPath, player.lastPath)
	}
	if !player.lastWait {
		t.Fatalf("expected configured wait=true")
	}
}

func TestPlayNotificationSoundToolCallValidationFailure(t *testing.T) {
	t.Parallel()

	tool := NewPlayNotificationSoundTool(validation.NewSoundValidator(t.TempDir()), &fakePlayer{}, "missing.wav", true, "")

	result, rpcErr := tool.Call(context.Background(), nil)
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %+v", rpcErr)
	}
	if !result.IsError {
		t.Fatalf("expected validation error result")
	}
}

func TestPlayNotificationSoundToolCallRejectsArguments(t *testing.T) {
	t.Parallel()

	tool := NewPlayNotificationSoundTool(validation.NewSoundValidator(t.TempDir()), &fakePlayer{}, "complete.wav", true, "")

	args, err := json.Marshal(map[string]any{"unknown": "ignored.wav"})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	_, rpcErr := tool.Call(context.Background(), args)
	if rpcErr == nil {
		t.Fatalf("expected rpc error")
	}
	if rpcErr.Code != errCodeInvalidParams {
		t.Fatalf("expected invalid params, got %d", rpcErr.Code)
	}
}

func TestPlayNotificationSoundToolCallUsesRuntimeOverrides(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	soundsDir := filepath.Join(tmpDir, "sounds")
	if err := os.MkdirAll(filepath.Join(soundsDir, "alerts"), 0o755); err != nil {
		t.Fatalf("mkdir sounds: %v", err)
	}
	soundPath := filepath.Join(soundsDir, "alerts", "sample.mp3")
	if err := os.WriteFile(soundPath, []byte("mp3"), 0o644); err != nil {
		t.Fatalf("write sound file: %v", err)
	}

	player := &fakePlayer{}
	tool := NewPlayNotificationSoundTool(validation.NewSoundValidator(tmpDir), player, "", true, "")

	overrideWait := false
	args, err := json.Marshal(map[string]any{
		"soundPath": "alerts/sample.mp3",
		"wait":      overrideWait,
	})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	result, rpcErr := tool.Call(context.Background(), args)
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %+v", rpcErr)
	}
	if result.IsError {
		t.Fatalf("expected success result, got error")
	}
	if player.lastPath != soundPath {
		t.Fatalf("expected resolved path %q, got %q", soundPath, player.lastPath)
	}
	if player.lastWait {
		t.Fatalf("expected runtime wait=false override")
	}
}

func TestPlayNotificationSoundToolCallFailsWithoutConfiguredOrRequestedSound(t *testing.T) {
	t.Parallel()

	tool := NewPlayNotificationSoundTool(validation.NewSoundValidator(t.TempDir()), &fakePlayer{}, "", true, "")

	result, rpcErr := tool.Call(context.Background(), nil)
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %+v", rpcErr)
	}
	if !result.IsError {
		t.Fatalf("expected error result")
	}
}

func TestPlayNotificationSoundToolDefinitionAppliesPrefix(t *testing.T) {
	t.Parallel()

	tool := NewPlayNotificationSoundTool(validation.NewSoundValidator(t.TempDir()), &fakePlayer{}, "", true, "complete_")
	definition := tool.Definition()

	if definition.Name != "complete_play_mcp_notification_sound" {
		t.Fatalf("expected prefixed tool name, got %q", definition.Name)
	}
}

func TestPlayNotificationSoundToolErrorMentionsPrefixedSchemaName(t *testing.T) {
	t.Parallel()

	tool := NewPlayNotificationSoundTool(validation.NewSoundValidator(t.TempDir()), &fakePlayer{}, "complete.wav", true, "complete_")

	args, err := json.Marshal(map[string]any{"unknown": "ignored.wav"})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	_, rpcErr := tool.Call(context.Background(), args)
	if rpcErr == nil {
		t.Fatalf("expected rpc error")
	}
	if rpcErr.Message != "arguments must match the complete_play_mcp_notification_sound schema" {
		t.Fatalf("unexpected rpc error message: %q", rpcErr.Message)
	}
}

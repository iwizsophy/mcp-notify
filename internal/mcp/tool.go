package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	"mcp-notify/internal/validation"
)

type soundPlayer interface {
	Play(ctx context.Context, soundPath string, wait bool) *validation.AppError
}

type PlayNotificationSoundTool struct {
	validator *validation.SoundValidator
	player    soundPlayer
	soundPath string
	wait      bool
	name      string
}

type playSoundResponse struct {
	Success   bool   `json:"success"`
	SoundPath string `json:"soundPath,omitempty"`
	Mode      string `json:"mode,omitempty"`
	Error     string `json:"error,omitempty"`
	Details   string `json:"details,omitempty"`
}

type playSoundArguments struct {
	SoundPath string `json:"soundPath,omitempty"`
	Wait      *bool  `json:"wait,omitempty"`
}

func NewPlayNotificationSoundTool(validator *validation.SoundValidator, player soundPlayer, soundPath string, wait bool, toolPrefix string) *PlayNotificationSoundTool {
	return &PlayNotificationSoundTool{
		validator: validator,
		player:    player,
		soundPath: soundPath,
		wait:      wait,
		name:      toolPrefix + "play_mcp_notification_sound",
	}
}

func (t *PlayNotificationSoundTool) Definition() toolDefinition {
	return toolDefinition{
		Name:        t.name,
		Title:       "Play MCP Notification Sound",
		Description: "Play a local notification sound file on the current machine.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"soundPath": map[string]any{
					"type":        "string",
					"description": "Optional relative path under sounds/. If omitted, the server's startup --sound value is used.",
				},
				"wait": map[string]any{
					"type":        "boolean",
					"description": "Optional playback mode override. true waits for completion; false returns after spawning detached playback.",
				},
			},
			"additionalProperties": false,
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"success": map[string]any{"type": "boolean"},
				"soundPath": map[string]any{
					"type": "string",
				},
				"mode": map[string]any{
					"type": "string",
					"enum": []string{"sync", "async"},
				},
				"error": map[string]any{
					"type": "string",
				},
				"details": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"success"},
		},
	}
}

func (t *PlayNotificationSoundTool) Call(ctx context.Context, arguments json.RawMessage) (toolResult, *responseError) {
	effectivePath := t.soundPath
	effectiveWait := t.wait

	if trimmed := bytes.TrimSpace(arguments); len(trimmed) > 0 && string(trimmed) != "null" {
		var provided playSoundArguments
		decoder := json.NewDecoder(bytes.NewReader(trimmed))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&provided); err != nil {
			return toolResult{}, &responseError{
				Code:    errCodeInvalidParams,
				Message: "arguments must match the " + t.name + " schema",
			}
		}
		if err := decoder.Decode(new(struct{})); err != io.EOF {
			return toolResult{}, &responseError{
				Code:    errCodeInvalidParams,
				Message: "arguments must match the " + t.name + " schema",
			}
		}
		if trimmedPath := strings.TrimSpace(provided.SoundPath); trimmedPath != "" {
			effectivePath = trimmedPath
		}
		if provided.Wait != nil {
			effectiveWait = *provided.Wait
		}
	}

	resolved, err := t.validator.ValidateRequestedPath(effectivePath)
	if err != nil {
		result := playSoundResponse{
			Success: false,
			Error:   err.Message,
			Details: err.Details,
		}
		return toolResult{
			Content:           marshalTextContent(result),
			StructuredContent: result,
			IsError:           true,
		}, nil
	}

	if err := t.player.Play(ctx, resolved.ResolvedPath, effectiveWait); err != nil {
		result := playSoundResponse{
			Success: false,
			Error:   err.Message,
			Details: err.Details,
		}
		return toolResult{
			Content:           marshalTextContent(result),
			StructuredContent: result,
			IsError:           true,
		}, nil
	}

	mode := "sync"
	if !effectiveWait {
		mode = "async"
	}

	result := playSoundResponse{
		Success:   true,
		SoundPath: resolved.ResolvedPath,
		Mode:      mode,
	}

	return toolResult{
		Content:           marshalTextContent(result),
		StructuredContent: result,
	}, nil
}

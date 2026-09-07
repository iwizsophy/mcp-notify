package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"

	"mcp-notify/internal/speech"
	"mcp-notify/internal/validation"
)

const maxSpeechTextRunes = 1000

type textSpeaker interface {
	Speak(ctx context.Context, request speech.Request, wait bool) *validation.AppError
}

type SpeakTextTool struct {
	speaker        textSpeaker
	defaultSpeaker int
	wait           bool
	name           string
}

type speakTextArguments struct {
	Text            string   `json:"text"`
	Speaker         *int     `json:"speaker,omitempty"`
	Wait            *bool    `json:"wait,omitempty"`
	SpeedScale      *float64 `json:"speedScale,omitempty"`
	PitchScale      *float64 `json:"pitchScale,omitempty"`
	IntonationScale *float64 `json:"intonationScale,omitempty"`
	VolumeScale     *float64 `json:"volumeScale,omitempty"`
}

type speakTextResponse struct {
	Success  bool   `json:"success"`
	Provider string `json:"provider,omitempty"`
	Speaker  *int   `json:"speaker,omitempty"`
	Mode     string `json:"mode,omitempty"`
	Error    string `json:"error,omitempty"`
	Details  string `json:"details,omitempty"`
}

func NewSpeakTextTool(speaker textSpeaker, defaultSpeaker int, wait bool, toolPrefix string) *SpeakTextTool {
	return &SpeakTextTool{
		speaker:        speaker,
		defaultSpeaker: defaultSpeaker,
		wait:           wait,
		name:           toolPrefix + "speak_text",
	}
}

func (t *SpeakTextTool) Definition() toolDefinition {
	return toolDefinition{
		Name:        t.name,
		Title:       "Speak Text",
		Description: "Synthesize text with VOICEVOX and play it on the current machine.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text": map[string]any{
					"type":        "string",
					"description": "Text to synthesize and speak.",
					"minLength":   1,
					"maxLength":   maxSpeechTextRunes,
				},
				"speaker": map[string]any{
					"type":        "integer",
					"minimum":     0,
					"description": "Optional VOICEVOX speaker/style ID override.",
				},
				"wait": map[string]any{
					"type":        "boolean",
					"description": "Optional playback mode override. true waits for playback; false returns after synthesis and starts playback asynchronously.",
				},
				"speedScale":      scaleDefinition("Speaking speed multiplier.", 0.5, 2.0),
				"pitchScale":      scaleDefinition("Voice pitch adjustment.", -0.15, 0.15),
				"intonationScale": scaleDefinition("Intonation multiplier.", 0.0, 2.0),
				"volumeScale":     scaleDefinition("Volume multiplier.", 0.0, 2.0),
			},
			"required":             []string{"text"},
			"additionalProperties": false,
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"success":  map[string]any{"type": "boolean"},
				"provider": map[string]any{"type": "string", "enum": []string{"voicevox"}},
				"speaker":  map[string]any{"type": "integer"},
				"mode":     map[string]any{"type": "string", "enum": []string{"sync", "async"}},
				"error":    map[string]any{"type": "string"},
				"details":  map[string]any{"type": "string"},
			},
			"required": []string{"success"},
		},
	}
}

func (t *SpeakTextTool) Call(ctx context.Context, arguments json.RawMessage) (toolResult, *responseError) {
	var provided speakTextArguments
	trimmed := bytes.TrimSpace(arguments)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return toolResult{}, t.invalidArguments()
	}

	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&provided); err != nil {
		return toolResult{}, t.invalidArguments()
	}
	if err := decoder.Decode(new(struct{})); err != io.EOF {
		return toolResult{}, t.invalidArguments()
	}

	provided.Text = strings.TrimSpace(provided.Text)
	if provided.Text == "" || !utf8.ValidString(provided.Text) || utf8.RuneCountInString(provided.Text) > maxSpeechTextRunes {
		return toolResult{}, t.invalidArguments()
	}

	effectiveSpeaker := t.defaultSpeaker
	if provided.Speaker != nil {
		effectiveSpeaker = *provided.Speaker
	}
	if effectiveSpeaker < 0 || !validScale(provided.SpeedScale, 0.5, 2.0) ||
		!validScale(provided.PitchScale, -0.15, 0.15) ||
		!validScale(provided.IntonationScale, 0.0, 2.0) ||
		!validScale(provided.VolumeScale, 0.0, 2.0) {
		return toolResult{}, t.invalidArguments()
	}

	effectiveWait := t.wait
	if provided.Wait != nil {
		effectiveWait = *provided.Wait
	}

	request := speech.Request{
		Text:            provided.Text,
		Speaker:         effectiveSpeaker,
		SpeedScale:      provided.SpeedScale,
		PitchScale:      provided.PitchScale,
		IntonationScale: provided.IntonationScale,
		VolumeScale:     provided.VolumeScale,
	}
	if err := t.speaker.Speak(ctx, request, effectiveWait); err != nil {
		result := speakTextResponse{
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
	result := speakTextResponse{
		Success:  true,
		Provider: "voicevox",
		Speaker:  &effectiveSpeaker,
		Mode:     mode,
	}
	return toolResult{
		Content:           marshalTextContent(result),
		StructuredContent: result,
	}, nil
}

func (t *SpeakTextTool) invalidArguments() *responseError {
	return &responseError{
		Code:    errCodeInvalidParams,
		Message: "arguments must match the " + t.name + " schema",
	}
}

func scaleDefinition(description string, minimum, maximum float64) map[string]any {
	return map[string]any{
		"type":        "number",
		"minimum":     minimum,
		"maximum":     maximum,
		"description": description,
	}
}

func validScale(value *float64, minimum, maximum float64) bool {
	return value == nil || (*value >= minimum && *value <= maximum)
}

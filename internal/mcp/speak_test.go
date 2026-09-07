package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"mcp-notify/internal/speech"
	"mcp-notify/internal/validation"
)

type fakeTextSpeaker struct {
	lastRequest speech.Request
	lastWait    bool
	err         *validation.AppError
}

func (f *fakeTextSpeaker) Speak(_ context.Context, request speech.Request, wait bool) *validation.AppError {
	f.lastRequest = request
	f.lastWait = wait
	return f.err
}

func TestSpeakTextToolCallUsesDefaults(t *testing.T) {
	t.Parallel()

	speaker := &fakeTextSpeaker{}
	tool := NewSpeakTextTool(speaker, 3, true, "")
	result, rpcErr := tool.Call(context.Background(), json.RawMessage(`{"text":"作業が完了しました"}`))

	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %+v", rpcErr)
	}
	if result.IsError {
		t.Fatalf("expected success result")
	}
	if speaker.lastRequest.Text != "作業が完了しました" {
		t.Fatalf("unexpected text: %q", speaker.lastRequest.Text)
	}
	if speaker.lastRequest.Speaker != 3 {
		t.Fatalf("expected default speaker 3, got %d", speaker.lastRequest.Speaker)
	}
	if !speaker.lastWait {
		t.Fatalf("expected wait=true")
	}
	response, ok := result.StructuredContent.(speakTextResponse)
	if !ok || response.Speaker == nil || *response.Speaker != 3 {
		t.Fatalf("expected speaker 3 in response, got %+v", result.StructuredContent)
	}
}

func TestSpeakTextToolResponsePreservesSpeakerZero(t *testing.T) {
	t.Parallel()

	tool := NewSpeakTextTool(&fakeTextSpeaker{}, 0, true, "")
	result, rpcErr := tool.Call(context.Background(), json.RawMessage(`{"text":"hello"}`))
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %+v", rpcErr)
	}

	encoded, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if string(encoded) != `{"success":true,"provider":"voicevox","speaker":0,"mode":"sync"}` {
		t.Fatalf("speaker zero was not preserved: %s", encoded)
	}
}

func TestSpeakTextToolCallUsesOverrides(t *testing.T) {
	t.Parallel()

	speaker := &fakeTextSpeaker{}
	tool := NewSpeakTextTool(speaker, 3, true, "")
	result, rpcErr := tool.Call(context.Background(), json.RawMessage(`{
		"text":" hello ",
		"speaker":2,
		"wait":false,
		"speedScale":1.25,
		"pitchScale":0.05,
		"intonationScale":1.4,
		"volumeScale":0.8
	}`))

	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %+v", rpcErr)
	}
	if result.IsError {
		t.Fatalf("expected success result")
	}
	if speaker.lastRequest.Text != "hello" || speaker.lastRequest.Speaker != 2 || speaker.lastWait {
		t.Fatalf("unexpected request: %+v, wait=%v", speaker.lastRequest, speaker.lastWait)
	}
	if speaker.lastRequest.SpeedScale == nil || *speaker.lastRequest.SpeedScale != 1.25 {
		t.Fatalf("expected speedScale override")
	}
}

func TestSpeakTextToolCallRejectsInvalidArguments(t *testing.T) {
	t.Parallel()
	tooLong, err := json.Marshal(map[string]string{"text": strings.Repeat("a", maxSpeechTextRunes+1)})
	if err != nil {
		t.Fatalf("marshal overlong input: %v", err)
	}

	testCases := []struct {
		name string
		args json.RawMessage
	}{
		{name: "missing arguments", args: nil},
		{name: "missing text", args: json.RawMessage(`{}`)},
		{name: "blank text", args: json.RawMessage(`{"text":"   "}`)},
		{name: "text too long", args: tooLong},
		{name: "unknown property", args: json.RawMessage(`{"text":"hello","unknown":true}`)},
		{name: "negative speaker", args: json.RawMessage(`{"text":"hello","speaker":-1}`)},
		{name: "speed too low", args: json.RawMessage(`{"text":"hello","speedScale":0.49}`)},
		{name: "pitch too high", args: json.RawMessage(`{"text":"hello","pitchScale":0.16}`)},
		{name: "intonation too high", args: json.RawMessage(`{"text":"hello","intonationScale":2.01}`)},
		{name: "volume too high", args: json.RawMessage(`{"text":"hello","volumeScale":2.01}`)},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tool := NewSpeakTextTool(&fakeTextSpeaker{}, 3, true, "")
			_, rpcErr := tool.Call(context.Background(), tc.args)
			if rpcErr == nil || rpcErr.Code != errCodeInvalidParams {
				t.Fatalf("expected invalid params error, got %+v", rpcErr)
			}
		})
	}
}

func TestSpeakTextToolCallReturnsBackendError(t *testing.T) {
	t.Parallel()

	tool := NewSpeakTextTool(&fakeTextSpeaker{err: validation.NewAppError("VOICEVOX unavailable", "connection refused")}, 3, true, "")
	result, rpcErr := tool.Call(context.Background(), json.RawMessage(`{"text":"hello"}`))

	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %+v", rpcErr)
	}
	if !result.IsError {
		t.Fatalf("expected tool error result")
	}
}

func TestSpeakTextToolDefinitionAppliesPrefix(t *testing.T) {
	t.Parallel()

	tool := NewSpeakTextTool(&fakeTextSpeaker{}, 3, true, "notify_")
	if tool.Definition().Name != "notify_speak_text" {
		t.Fatalf("unexpected tool name: %q", tool.Definition().Name)
	}
}

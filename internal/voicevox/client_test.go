package voicevox

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mcp-notify/internal/speech"
)

func TestClientSynthesize(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/audio_query", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Query().Get("text") != "こんにちは" || request.URL.Query().Get("speaker") != "3" {
			t.Errorf("unexpected audio query request: %s %s", request.Method, request.URL.String())
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"speedScale":1,"pitchScale":0,"intonationScale":1,"volumeScale":1}`))
	})
	mux.HandleFunc("/synthesis", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Query().Get("speaker") != "3" {
			t.Errorf("unexpected synthesis request: %s %s", request.Method, request.URL.String())
		}
		var query map[string]any
		if err := json.NewDecoder(request.Body).Decode(&query); err != nil {
			t.Errorf("decode synthesis body: %v", err)
		}
		if query["speedScale"] != 1.25 || query["pitchScale"] != 0.05 || query["intonationScale"] != 1.4 || query["volumeScale"] != 0.8 {
			t.Errorf("speech options were not applied: %+v", query)
		}
		writer.Header().Set("Content-Type", "audio/wav")
		_, _ = writer.Write([]byte("wav-data"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	speed, pitch, intonation, volume := 1.25, 0.05, 1.4, 0.8
	wavData, appErr := client.Synthesize(context.Background(), speech.Request{
		Text:            "こんにちは",
		Speaker:         3,
		SpeedScale:      &speed,
		PitchScale:      &pitch,
		IntonationScale: &intonation,
		VolumeScale:     &volume,
	})
	if appErr != nil {
		t.Fatalf("unexpected app error: %v", appErr)
	}
	if string(wavData) != "wav-data" {
		t.Fatalf("unexpected WAV data: %q", wavData)
	}
}

func TestClientSynthesizeReturnsHTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.Error(writer, "speaker not found", http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, appErr := client.Synthesize(context.Background(), speech.Request{Text: "hello", Speaker: 999})
	if appErr == nil || appErr.Message != "VOICEVOX audio query request failed" {
		t.Fatalf("expected HTTP error, got %v", appErr)
	}
}

func TestClientSynthesizeRejectsInvalidAudioQuery(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`not-json`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, appErr := client.Synthesize(context.Background(), speech.Request{Text: "hello", Speaker: 3})
	if appErr == nil || appErr.Message != "VOICEVOX returned an invalid audio query" {
		t.Fatalf("expected invalid query error, got %v", appErr)
	}
}

func TestClientSynthesizeRejectsEmptyAudio(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/audio_query", func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{}`))
	})
	mux.HandleFunc("/synthesis", func(_ http.ResponseWriter, _ *http.Request) {})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, appErr := client.Synthesize(context.Background(), speech.Request{Text: "hello", Speaker: 3})
	if appErr == nil || appErr.Message != "VOICEVOX returned empty synthesized audio" {
		t.Fatalf("expected empty audio error, got %v", appErr)
	}
}

func TestReadResponseEnforcesSizeLimit(t *testing.T) {
	t.Parallel()

	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("123456")),
	}
	_, appErr := readResponse(response, 5, "VOICEVOX audio query")
	if appErr == nil || appErr.Message != "VOICEVOX audio query response is too large" {
		t.Fatalf("expected size-limit error, got %v", appErr)
	}
}

func TestNewClientRejectsInvalidEndpoint(t *testing.T) {
	t.Parallel()

	for _, endpoint := range []string{"", "127.0.0.1:50021", "file:///tmp/voicevox", "http://localhost:50021?test=1"} {
		if _, err := NewClient(endpoint, nil); err == nil {
			t.Errorf("expected endpoint %q to be rejected", endpoint)
		}
	}
}

func TestClientPreservesEndpointBasePath(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:50021/api/", nil)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if got := client.apiURL("audio_query").Path; got != "/api/audio_query" {
		t.Fatalf("unexpected API path: %q", got)
	}
}

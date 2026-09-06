package voicevox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mcp-notify/internal/speech"
	"mcp-notify/internal/validation"
)

const (
	maxAudioQueryBytes = 4 << 20
	maxAudioBytes      = 32 << 20
)

type Client struct {
	endpoint   *url.URL
	httpClient *http.Client
}

func NewClient(endpoint string, httpClient *http.Client) (*Client, error) {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return nil, fmt.Errorf("parse VOICEVOX URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("VOICEVOX URL must use http or https")
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("VOICEVOX URL must include a host")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("VOICEVOX URL must not include a query or fragment")
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		endpoint:   parsed,
		httpClient: httpClient,
	}, nil
}

func (c *Client) Synthesize(ctx context.Context, request speech.Request) ([]byte, *validation.AppError) {
	audioQueryURL := c.apiURL("audio_query")
	query := audioQueryURL.Query()
	query.Set("text", request.Text)
	query.Set("speaker", fmt.Sprintf("%d", request.Speaker))
	audioQueryURL.RawQuery = query.Encode()

	audioQueryRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, audioQueryURL.String(), nil)
	if err != nil {
		return nil, requestError("failed to create VOICEVOX audio query request", err)
	}
	audioQueryRequest.Header.Set("Accept", "application/json")

	audioQueryResponse, err := c.httpClient.Do(audioQueryRequest)
	if err != nil {
		return nil, requestError("failed to request a VOICEVOX audio query", err)
	}
	audioQueryBody, appErr := readResponse(audioQueryResponse, maxAudioQueryBytes, "VOICEVOX audio query")
	if appErr != nil {
		return nil, appErr
	}

	var audioQuery map[string]any
	if err := json.Unmarshal(audioQueryBody, &audioQuery); err != nil {
		return nil, requestError("VOICEVOX returned an invalid audio query", err)
	}
	applySpeechOptions(audioQuery, request)

	synthesisBody, err := json.Marshal(audioQuery)
	if err != nil {
		return nil, requestError("failed to encode the VOICEVOX synthesis request", err)
	}

	synthesisURL := c.apiURL("synthesis")
	query = synthesisURL.Query()
	query.Set("speaker", fmt.Sprintf("%d", request.Speaker))
	synthesisURL.RawQuery = query.Encode()

	synthesisRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, synthesisURL.String(), bytes.NewReader(synthesisBody))
	if err != nil {
		return nil, requestError("failed to create VOICEVOX synthesis request", err)
	}
	synthesisRequest.Header.Set("Accept", "audio/wav")
	synthesisRequest.Header.Set("Content-Type", "application/json")

	synthesisResponse, err := c.httpClient.Do(synthesisRequest)
	if err != nil {
		return nil, requestError("failed to request VOICEVOX speech synthesis", err)
	}
	wavData, appErr := readResponse(synthesisResponse, maxAudioBytes, "VOICEVOX synthesis")
	if appErr != nil {
		return nil, appErr
	}
	if len(wavData) == 0 {
		return nil, validation.NewAppError(
			"VOICEVOX returned empty synthesized audio",
			"the synthesis endpoint returned an empty response body",
		)
	}

	return wavData, nil
}

func (c *Client) apiURL(name string) *url.URL {
	result := *c.endpoint
	result.Path = strings.TrimRight(result.Path, "/") + "/" + name
	result.RawPath = ""
	result.RawQuery = ""
	result.Fragment = ""
	return &result
}

func applySpeechOptions(audioQuery map[string]any, request speech.Request) {
	if request.SpeedScale != nil {
		audioQuery["speedScale"] = *request.SpeedScale
	}
	if request.PitchScale != nil {
		audioQuery["pitchScale"] = *request.PitchScale
	}
	if request.IntonationScale != nil {
		audioQuery["intonationScale"] = *request.IntonationScale
	}
	if request.VolumeScale != nil {
		audioQuery["volumeScale"] = *request.VolumeScale
	}
}

func readResponse(response *http.Response, limit int64, operation string) ([]byte, *validation.AppError) {
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, requestError("failed to read "+operation+" response", err)
	}
	if int64(len(body)) > limit {
		return nil, validation.NewAppError(
			operation+" response is too large",
			fmt.Sprintf("response exceeded the %d byte limit", limit),
		)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		details := strings.TrimSpace(string(body))
		if len(details) > 1024 {
			details = details[:1024]
		}
		return nil, validation.NewAppError(
			operation+" request failed",
			fmt.Sprintf("HTTP %d: %s", response.StatusCode, details),
		)
	}

	return body, nil
}

func requestError(message string, err error) *validation.AppError {
	return validation.NewAppError(message, err.Error())
}

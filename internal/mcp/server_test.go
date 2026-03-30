package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"strconv"
	"strings"
	"testing"
)

type stubTool struct {
	definition    toolDefinition
	result        toolResult
	rpcErr        *responseError
	lastArguments json.RawMessage
	callCount     int
}

func (t *stubTool) Definition() toolDefinition {
	return t.definition
}

func (t *stubTool) Call(_ context.Context, arguments json.RawMessage) (toolResult, *responseError) {
	t.callCount++
	t.lastArguments = append(json.RawMessage(nil), arguments...)
	return t.result, t.rpcErr
}

func TestInitializeReturnsConfiguredErrorWithData(t *testing.T) {
	t.Parallel()

	server := NewServer("mcp-notify", "1.0.0", log.New(io.Discard, "", 0))
	server.SetInitializeCheck(func() *ResponseError {
		return &ResponseError{
			Code:    errCodeInvalidParams,
			Message: "invalid startup sound configuration",
			Data: map[string]any{
				"error":   "configured soundPath must not be empty",
				"details": "set the server startup argument --sound to a file name under the sounds directory",
			},
		}
	})

	req := request{
		JSONRPC: jsonRPCVersion,
		ID:      json.RawMessage("1"),
		Method:  "initialize",
		Params:  initializeParamsJSON("2025-11-25"),
	}

	resp := server.handleLine(context.Background(), mustJSON(t, req))
	if resp == nil || resp.Error == nil {
		t.Fatalf("expected initialize error response")
	}
	if resp.Error.Message != "invalid startup sound configuration" {
		t.Fatalf("unexpected message: %s", resp.Error.Message)
	}
	if resp.Error.Data == nil {
		t.Fatalf("expected error data")
	}
}

func TestInitializeSuccessMarksServerInitialized(t *testing.T) {
	t.Parallel()

	server := NewServer("mcp-notify", "1.0.0", log.New(io.Discard, "", 0))
	req := initializeRequest("2025-11-25")

	resp := server.handleLine(context.Background(), mustJSON(t, req))
	result := mustInitializeResult(t, resp)
	if !server.initialized {
		t.Fatalf("expected server to be marked initialized")
	}
	if result.ProtocolVersion != "2025-11-25" {
		t.Fatalf("expected negotiated protocol version %q, got %q", "2025-11-25", result.ProtocolVersion)
	}
}

func TestInitializeReturnsConfiguredServerName(t *testing.T) {
	t.Parallel()

	server := NewServer("notify-complete", "1.0.0", log.New(io.Discard, "", 0))
	req := initializeRequest("2025-11-25")

	resp := server.handleLine(context.Background(), mustJSON(t, req))
	result := mustInitializeResult(t, resp)

	if result.ServerInfo.Name != "notify-complete" {
		t.Fatalf("expected configured server name, got %q", result.ServerInfo.Name)
	}
}

func TestInitializeSupportsPreviousProtocolVersion(t *testing.T) {
	t.Parallel()

	server := NewServer("mcp-notify", "1.0.0", log.New(io.Discard, "", 0))
	resp := server.handleLine(context.Background(), mustJSON(t, initializeRequest("2024-11-05")))
	result := mustInitializeResult(t, resp)

	if result.ProtocolVersion != "2024-11-05" {
		t.Fatalf("expected negotiated protocol version %q, got %q", "2024-11-05", result.ProtocolVersion)
	}
	if !server.initialized {
		t.Fatalf("expected server to be marked initialized")
	}
}

func TestInitializeNegotiatesUnsupportedProtocolVersion(t *testing.T) {
	t.Parallel()

	server := NewServer("mcp-notify", "1.0.0", log.New(io.Discard, "", 0))
	resp := server.handleLine(context.Background(), mustJSON(t, initializeRequest("2026-02-01")))
	result := mustInitializeResult(t, resp)

	if result.ProtocolVersion != protocolVersion {
		t.Fatalf("expected negotiated protocol version %q, got %q", protocolVersion, result.ProtocolVersion)
	}
}

func TestServeSupportsLineDelimitedTransport(t *testing.T) {
	t.Parallel()

	server := NewServer("mcp-notify", "1.0.0", log.New(io.Discard, "", 0))
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}` + "\n")
	var output bytes.Buffer

	if err := server.Serve(context.Background(), input, &output); err != nil {
		t.Fatalf("serve failed: %v", err)
	}

	if !strings.Contains(output.String(), `"protocolVersion":"2024-11-05"`) {
		t.Fatalf("expected initialize result in output, got %q", output.String())
	}
}

func TestServeSupportsContentLengthTransport(t *testing.T) {
	t.Parallel()

	server := NewServer("mcp-notify", "1.0.0", log.New(io.Discard, "", 0))
	payload := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}`
	input := strings.NewReader("Content-Length: " + mustStringLen(payload) + "\r\n\r\n" + payload)
	var output bytes.Buffer

	if err := server.Serve(context.Background(), input, &output); err != nil {
		t.Fatalf("serve failed: %v", err)
	}

	if !strings.HasPrefix(output.String(), "Content-Length: ") {
		t.Fatalf("expected Content-Length response, got %q", output.String())
	}
	if !strings.Contains(output.String(), `"protocolVersion":"2025-11-25"`) {
		t.Fatalf("expected initialize result in output, got %q", output.String())
	}
}

func TestToolsListAndCallRemainAvailableAfterInitialize(t *testing.T) {
	t.Parallel()

	server := NewServer("mcp-notify", "1.0.0", log.New(io.Discard, "", 0))
	tool := &stubTool{
		definition: toolDefinition{
			Name:        "test-tool",
			Description: "test tool",
			InputSchema: map[string]any{"type": "object"},
		},
		result: toolResult{
			Content: []toolContent{{
				Type: "text",
				Text: `{"ok":true}`,
			}},
		},
	}
	server.RegisterTool(tool)

	initializeResp := server.handleLine(context.Background(), mustJSON(t, initializeRequest("2024-11-05")))
	mustInitializeResult(t, initializeResp)

	listResp := server.handleLine(context.Background(), mustJSON(t, request{
		JSONRPC: jsonRPCVersion,
		ID:      json.RawMessage("2"),
		Method:  "tools/list",
	}))
	if listResp == nil || listResp.Error != nil {
		t.Fatalf("expected tools/list success response, got %+v", listResp)
	}
	listResult, ok := listResp.Result.(toolsListResult)
	if !ok {
		t.Fatalf("expected toolsListResult, got %T", listResp.Result)
	}
	if len(listResult.Tools) != 1 || listResult.Tools[0].Name != "test-tool" {
		t.Fatalf("unexpected tools/list result: %+v", listResult.Tools)
	}

	callResp := server.handleLine(context.Background(), mustJSON(t, request{
		JSONRPC: jsonRPCVersion,
		ID:      json.RawMessage("3"),
		Method:  "tools/call",
		Params: json.RawMessage(`{
			"name":"test-tool",
			"arguments":{"message":"hello"}
		}`),
	}))
	if callResp == nil || callResp.Error != nil {
		t.Fatalf("expected tools/call success response, got %+v", callResp)
	}
	callResult, ok := callResp.Result.(toolResult)
	if !ok {
		t.Fatalf("expected toolResult, got %T", callResp.Result)
	}
	if len(callResult.Content) != 1 || callResult.Content[0].Text != `{"ok":true}` {
		t.Fatalf("unexpected tools/call result: %+v", callResult)
	}
	if tool.callCount != 1 {
		t.Fatalf("expected tool to be called once, got %d", tool.callCount)
	}
	if string(tool.lastArguments) != `{"message":"hello"}` {
		t.Fatalf("unexpected tool arguments: %s", string(tool.lastArguments))
	}
}

func initializeRequest(version string) request {
	return request{
		JSONRPC: jsonRPCVersion,
		ID:      json.RawMessage("1"),
		Method:  "initialize",
		Params:  initializeParamsJSON(version),
	}
}

func initializeParamsJSON(version string) json.RawMessage {
	return json.RawMessage(`{
		"protocolVersion":"` + version + `",
		"capabilities":{},
		"clientInfo":{"name":"test","version":"1.0.0"}
	}`)
}

func mustInitializeResult(t *testing.T, resp *response) initializeResult {
	t.Helper()

	if resp == nil || resp.Error != nil {
		t.Fatalf("expected initialize success response, got %+v", resp)
	}

	result, ok := resp.Result.(initializeResult)
	if !ok {
		t.Fatalf("expected initializeResult, got %T", resp.Result)
	}

	return result
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	body, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return body
}

func mustStringLen(value string) string {
	return strconv.Itoa(len(value))
}

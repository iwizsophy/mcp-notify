package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

type Tool interface {
	Definition() toolDefinition
	Call(ctx context.Context, arguments json.RawMessage) (toolResult, *responseError)
}

type Server struct {
	name        string
	version     string
	logger      *log.Logger
	tools       map[string]Tool
	initialized bool
	initCheck   func() *ResponseError
}

func NewServer(name, version string, logger *log.Logger) *Server {
	return &Server{
		name:    name,
		version: version,
		logger:  logger,
		tools:   make(map[string]Tool),
	}
}

func (s *Server) RegisterTool(tool Tool) {
	s.tools[tool.Definition().Name] = tool
}

func (s *Server) SetInitializeCheck(check func() *ResponseError) {
	s.initCheck = check
}

func (s *Server) Serve(ctx context.Context, input io.Reader, output io.Writer) error {
	reader := bufio.NewReader(input)
	writer := newTransportWriter(output)

	for {
		payload, mode, err := readMessage(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read request: %w", err)
		}

		if len(payload) == 0 {
			continue
		}

		if writer.mode == transportModeUnset {
			writer.mode = mode
		}

		reply := s.handleLine(ctx, payload)
		if reply == nil {
			continue
		}

		if err := writer.write(reply); err != nil {
			return fmt.Errorf("write response: %w", err)
		}
	}
}

func (s *Server) handleLine(ctx context.Context, line []byte) *response {
	var req request
	if err := json.Unmarshal(line, &req); err != nil {
		return &response{
			JSONRPC: jsonRPCVersion,
			ID:      json.RawMessage("null"),
			Error: &responseError{
				Code:    errCodeParseError,
				Message: "invalid JSON-RPC message",
			},
		}
	}

	if req.JSONRPC != "" && req.JSONRPC != jsonRPCVersion {
		return s.errorResponse(req.ID, errCodeInvalidReq, "jsonrpc must be \"2.0\"")
	}

	if req.Method == "" {
		return s.errorResponse(req.ID, errCodeInvalidReq, "method is required")
	}

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "notifications/initialized":
		s.initialized = true
		return nil
	case "ping":
		return s.resultResponse(req.ID, map[string]any{})
	case "tools/list":
		if resp := s.requireInitialized(req); resp != nil {
			return resp
		}
		return s.handleToolsList(req)
	case "tools/call":
		if resp := s.requireInitialized(req); resp != nil {
			return resp
		}
		return s.handleToolsCall(ctx, req)
	case "shutdown":
		return s.resultResponse(req.ID, map[string]any{})
	case "notifications/cancelled", "notifications/exit":
		return nil
	default:
		return s.errorResponse(req.ID, errCodeMethodMissing, "method not found")
	}
}

func (s *Server) requireInitialized(req request) *response {
	if s.initialized {
		return nil
	}
	return s.errorResponse(req.ID, errCodeInvalidReq, "server has not been initialized")
}

func (s *Server) handleInitialize(req request) *response {
	if req.hasID() == false {
		return s.errorResponse(json.RawMessage("null"), errCodeInvalidReq, "initialize must be a request")
	}

	var params initializeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.errorResponse(req.ID, errCodeInvalidParams, "initialize params are invalid")
	}

	if params.ProtocolVersion == "" {
		return s.errorResponse(req.ID, errCodeInvalidParams, "protocolVersion is required")
	}

	negotiatedVersion := negotiateProtocolVersion(params.ProtocolVersion)

	if s.initCheck != nil {
		if initErr := s.initCheck(); initErr != nil {
			return s.errorResponseWithData(req.ID, initErr.Code, initErr.Message, initErr.Data)
		}
	}

	s.initialized = true

	return s.resultResponse(req.ID, initializeResult{
		ProtocolVersion: negotiatedVersion,
		Capabilities: map[string]any{
			"tools": map[string]any{
				"listChanged": false,
			},
		},
		ServerInfo: peerInfo{
			Name:    s.name,
			Version: s.version,
		},
	})
}

func (s *Server) handleToolsList(req request) *response {
	tools := make([]toolDefinition, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool.Definition())
	}

	return s.resultResponse(req.ID, toolsListResult{Tools: tools})
}

func (s *Server) handleToolsCall(ctx context.Context, req request) *response {
	var params toolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.errorResponse(req.ID, errCodeInvalidParams, "tool call params are invalid")
	}

	tool, ok := s.tools[params.Name]
	if !ok {
		return s.errorResponse(req.ID, errCodeInvalidParams, fmt.Sprintf("unknown tool: %s", params.Name))
	}

	result, rpcErr := tool.Call(ctx, params.Arguments)
	if rpcErr != nil {
		return s.errorResponse(req.ID, rpcErr.Code, rpcErr.Message)
	}

	return s.resultResponse(req.ID, result)
}

func (s *Server) resultResponse(id json.RawMessage, result any) *response {
	if len(id) == 0 {
		id = json.RawMessage("null")
	}
	return &response{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Result:  result,
	}
}

func (s *Server) errorResponse(id json.RawMessage, code int, message string) *response {
	return s.errorResponseWithData(id, code, message, nil)
}

func (s *Server) errorResponseWithData(id json.RawMessage, code int, message string, data any) *response {
	if len(id) == 0 {
		id = json.RawMessage("null")
	}
	return &response{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Error: &responseError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

func marshalTextContent(payload any) []toolContent {
	body, err := json.Marshal(payload)
	if err != nil {
		body = []byte(fmt.Sprintf(`{"error":"failed to serialize result: %s"}`, err.Error()))
	}

	return []toolContent{{
		Type: "text",
		Text: string(body),
	}}
}

type transportMode int

const (
	transportModeUnset transportMode = iota
	transportModeLine
	transportModeContentLength
)

type transportWriter struct {
	writer io.Writer
	mode   transportMode
}

func newTransportWriter(writer io.Writer) *transportWriter {
	return &transportWriter{writer: writer}
}

func (w *transportWriter) write(resp *response) error {
	body, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	switch w.mode {
	case transportModeContentLength:
		if _, err := fmt.Fprintf(w.writer, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
			return err
		}
		_, err = w.writer.Write(body)
		return err
	default:
		_, err := w.writer.Write(append(body, '\n'))
		return err
	}
}

func readMessage(reader *bufio.Reader) ([]byte, transportMode, error) {
	firstBytes, err := reader.Peek(1)
	if err != nil {
		return nil, transportModeUnset, err
	}

	// Accept newline-delimited JSON for local tests and Content-Length framed
	// messages for stdio clients that speak the MCP transport literally.
	if firstBytes[0] == '{' || firstBytes[0] == '[' {
		line, err := reader.ReadBytes('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, transportModeUnset, err
		}
		return bytes.TrimSpace(line), transportModeLine, nil
	}

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, transportModeUnset, err
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			break
		}
		key, value, found := strings.Cut(trimmed, ":")
		if !found {
			return nil, transportModeUnset, fmt.Errorf("invalid header line: %q", trimmed)
		}
		headers[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}

	lengthHeader, ok := headers["content-length"]
	if !ok {
		return nil, transportModeUnset, fmt.Errorf("missing Content-Length header")
	}
	length, err := strconv.Atoi(lengthHeader)
	if err != nil || length < 0 {
		return nil, transportModeUnset, fmt.Errorf("invalid Content-Length header: %q", lengthHeader)
	}

	body := make([]byte, length)
	if _, err := io.ReadFull(reader, body); err != nil {
		return nil, transportModeUnset, err
	}

	return body, transportModeContentLength, nil
}

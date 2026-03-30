package mcp

import "encoding/json"

const (
	jsonRPCVersion       = "2.0"
	protocolVersion      = "2025-11-25"
	errCodeParseError    = -32700
	errCodeInvalidReq    = -32600
	errCodeMethodMissing = -32601
	errCodeInvalidParams = -32602
	errCodeInternal      = -32603
)

var supportedProtocolVersions = []string{
	"2024-11-05",
	protocolVersion,
}

func negotiateProtocolVersion(requested string) string {
	for _, supported := range supportedProtocolVersions {
		if requested == supported {
			return requested
		}
	}

	return protocolVersion
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func (r request) hasID() bool {
	return len(r.ID) > 0
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *responseError  `json:"error,omitempty"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ResponseError = responseError

type initializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities,omitempty"`
	ClientInfo      peerInfo       `json:"clientInfo"`
}

type initializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ServerInfo      peerInfo       `json:"serverInfo"`
}

type peerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type toolsListResult struct {
	Tools []toolDefinition `json:"tools"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type toolDefinition struct {
	Name         string         `json:"name"`
	Title        string         `json:"title,omitempty"`
	Description  string         `json:"description,omitempty"`
	InputSchema  map[string]any `json:"inputSchema"`
	OutputSchema map[string]any `json:"outputSchema,omitempty"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolResult struct {
	Content           []toolContent `json:"content"`
	StructuredContent any           `json:"structuredContent,omitempty"`
	IsError           bool          `json:"isError,omitempty"`
}

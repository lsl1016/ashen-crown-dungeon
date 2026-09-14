package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ashen-crown-dungeon/internal/agentgateway"
)

const mcpProtocolVersion = "2026-07-28"

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (s *Server) mcp(w http.ResponseWriter, r *http.Request) {
	if !s.requireAgentAuth(w, r) {
		return
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.UseNumber()
	var req mcpRequest
	if err := dec.Decode(&req); err != nil {
		s.writeMCPError(w, nil, http.StatusBadRequest, -32700, "Parse error", err.Error())
		return
	}
	if req.JSONRPC != "2.0" || len(req.ID) == 0 || req.Method == "" {
		s.writeMCPError(w, req.ID, http.StatusBadRequest, -32600, "Invalid Request", nil)
		return
	}
	params, err := decodeMCPParams(req.Params)
	if err != nil {
		s.writeMCPError(w, req.ID, http.StatusBadRequest, -32602, "Invalid params", err.Error())
		return
	}
	if err := validateMCPEnvelope(r, req.Method, params); err != nil {
		s.writeMCPError(w, req.ID, http.StatusBadRequest, -32020, "HeaderMismatch", err.Error())
		return
	}

	serverMeta := map[string]any{
		"io.modelcontextprotocol/serverInfo": map[string]any{"name": "ashen-crown-gm", "version": "0.7.0-ai-gm"},
	}
	switch req.Method {
	case "server/discover":
		s.writeMCPResult(w, req.ID, map[string]any{
			"resultType":        "complete",
			"supportedVersions": []string{mcpProtocolVersion},
			"capabilities":      map[string]any{"tools": map[string]any{}},
			"serverInfo":        map[string]any{"name": "ashen-crown-gm", "version": "0.7.0-ai-gm"},
			"instructions":      "Operate a persisted Ashen Crown run through tools. Always inspect before mutation. Every tool call carries runId explicitly. Use dryRun for high-risk world mutations.",
			"ttlMs":             60000,
			"cacheScope":        "private",
			"_meta":             serverMeta,
		})
	case "tools/list":
		if cursor, _ := params["cursor"].(string); cursor != "" {
			s.writeMCPError(w, req.ID, http.StatusBadRequest, -32602, "Invalid params", "pagination cursor is not supported because the tool list fits in one page")
			return
		}
		s.writeMCPResult(w, req.ID, map[string]any{
			"resultType": "complete",
			"tools":      s.agent.Tools(),
			"ttlMs":      60000,
			"cacheScope": "private",
			"_meta":      serverMeta,
		})
	case "tools/call":
		name, _ := params["name"].(string)
		name = strings.TrimSpace(name)
		if name == "" {
			s.writeMCPError(w, req.ID, http.StatusBadRequest, -32602, "Invalid params", "tools/call.params.name is required")
			return
		}
		if _, ok := s.agent.Definition(name); !ok {
			s.writeMCPError(w, req.ID, http.StatusBadRequest, -32602, "Invalid params", "unknown tool: "+name)
			return
		}
		arguments := map[string]any{}
		if raw, ok := params["arguments"]; ok {
			m, ok := raw.(map[string]any)
			if !ok {
				s.writeMCPError(w, req.ID, http.StatusBadRequest, -32602, "Invalid params", "arguments must be an object")
				return
			}
			arguments = m
		}
		if err := validateMCPRunHeader(r, arguments); err != nil {
			s.writeMCPError(w, req.ID, http.StatusBadRequest, -32020, "HeaderMismatch", err.Error())
			return
		}
		caller := callerFromHTTP(r, "mcp")
		if meta, ok := params["_meta"].(map[string]any); ok {
			if ci, ok := meta["io.modelcontextprotocol/clientInfo"].(map[string]any); ok {
				n, _ := ci["name"].(string)
				v, _ := ci["version"].(string)
				caller.ClientName = strings.TrimSpace(n + " " + v)
			}
			if trace, ok := meta["traceparent"].(string); ok && caller.TraceID == "" {
				caller.TraceID = trace
			}
		}
		result, execErr := s.agent.Execute(r.Context(), agentgateway.ExecuteRequest{Tool: name, Arguments: arguments}, caller)
		if execErr != nil {
			s.writeMCPResult(w, req.ID, map[string]any{
				"resultType": "complete",
				"content":    []map[string]any{{"type": "text", "text": execErr.Error()}},
				"isError":    true,
				"_meta":      serverMeta,
			})
			return
		}
		s.writeMCPResult(w, req.ID, map[string]any{
			"resultType":        "complete",
			"content":           []map[string]any{{"type": "text", "text": result.Summary}},
			"structuredContent": result,
			"isError":           false,
			"_meta":             serverMeta,
		})
	default:
		s.writeMCPError(w, req.ID, http.StatusNotFound, -32601, "Method not found", req.Method)
	}
}

func decodeMCPParams(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]any{}, nil
	}
	var params map[string]any
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err := dec.Decode(&params); err != nil {
		return nil, err
	}
	if params == nil {
		params = map[string]any{}
	}
	return params, nil
}

func validateMCPEnvelope(r *http.Request, method string, params map[string]any) error {
	headerVersion := strings.TrimSpace(r.Header.Get("MCP-Protocol-Version"))
	if headerVersion != mcpProtocolVersion {
		return fmt.Errorf("MCP-Protocol-Version must be %s", mcpProtocolVersion)
	}
	if h := strings.TrimSpace(r.Header.Get("Mcp-Method")); h == "" || h != method {
		return fmt.Errorf("Mcp-Method header must match JSON-RPC method")
	}
	meta, ok := params["_meta"].(map[string]any)
	if !ok {
		return fmt.Errorf("params._meta is required for MCP %s", mcpProtocolVersion)
	}
	pv, _ := meta["io.modelcontextprotocol/protocolVersion"].(string)
	if pv != mcpProtocolVersion || pv != headerVersion {
		return fmt.Errorf("protocolVersion metadata/header mismatch")
	}
	if _, ok := meta["io.modelcontextprotocol/clientCapabilities"].(map[string]any); !ok {
		return fmt.Errorf("params._meta.io.modelcontextprotocol/clientCapabilities is required")
	}
	if method == "tools/call" {
		name, _ := params["name"].(string)
		if h := strings.TrimSpace(r.Header.Get("Mcp-Name")); h == "" || h != name {
			return fmt.Errorf("Mcp-Name header must match tools/call params.name")
		}
	}
	return nil
}

func validateMCPRunHeader(r *http.Request, arguments map[string]any) error {
	runID, _ := arguments["runId"].(string)
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return fmt.Errorf("runId argument is required")
	}
	raw := strings.TrimSpace(r.Header.Get("Mcp-Param-Run-Id"))
	if raw == "" {
		return fmt.Errorf("Mcp-Param-Run-Id is required because runId uses x-mcp-header")
	}
	headerRunID, err := decodeMCPParamHeader(raw)
	if err != nil {
		return fmt.Errorf("invalid Mcp-Param-Run-Id: %w", err)
	}
	if headerRunID != runID {
		return fmt.Errorf("Mcp-Param-Run-Id does not match arguments.runId")
	}
	return nil
}

func decodeMCPParamHeader(raw string) (string, error) {
	if strings.HasPrefix(raw, "=?base64?") && strings.HasSuffix(raw, "?=") {
		encoded := strings.TrimSuffix(strings.TrimPrefix(raw, "=?base64?"), "?=")
		b, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return raw, nil
}

func (s *Server) writeMCPResult(w http.ResponseWriter, id json.RawMessage, result any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(mcpResponse{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *Server) writeMCPError(w http.ResponseWriter, id json.RawMessage, httpStatus, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(mcpResponse{JSONRPC: "2.0", ID: id, Error: &mcpError{Code: code, Message: message, Data: data}})
}

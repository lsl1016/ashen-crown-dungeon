package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ashen-crown-dungeon/internal/game"
)

func newMCPTestServer(t *testing.T) (*Server, *game.Run) {
	t.Helper()
	t.Setenv("ASHEN_AGENT_TOKEN", "")
	store := game.NewStore(t.TempDir())
	engine := game.NewEngine()
	run, err := engine.NewRun("MCPTest", "warden", 8080)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveRun(run); err != nil {
		t.Fatal(err)
	}
	return New(engine, store, t.TempDir()), run
}

func mcpMeta() map[string]any {
	return map[string]any{
		"io.modelcontextprotocol/protocolVersion":    mcpProtocolVersion,
		"io.modelcontextprotocol/clientCapabilities": map[string]any{},
		"io.modelcontextprotocol/clientInfo":         map[string]any{"name": "test-client", "version": "1.0"},
	}
}

func doMCP(t *testing.T, s *Server, method, name string, params map[string]any, runID string) *httptest.ResponseRecorder {
	t.Helper()
	params["_meta"] = mcpMeta()
	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("MCP-Protocol-Version", mcpProtocolVersion)
	req.Header.Set("Mcp-Method", method)
	if name != "" {
		req.Header.Set("Mcp-Name", name)
	}
	if runID != "" {
		req.Header.Set("Mcp-Param-Run-Id", runID)
	}
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	return rr
}

func TestMCPDiscover(t *testing.T) {
	s, _ := newMCPTestServer(t)
	rr := doMCP(t, s, "server/discover", "", map[string]any{}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	result := out["result"].(map[string]any)
	if result["resultType"] != "complete" {
		t.Fatalf("unexpected result: %v", result)
	}
}

func TestMCPToolsList(t *testing.T) {
	s, _ := newMCPTestServer(t)
	rr := doMCP(t, s, "tools/list", "", map[string]any{}, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	result := out["result"].(map[string]any)
	tools := result["tools"].([]any)
	if len(tools) < 10 {
		t.Fatalf("expected tools, got %d", len(tools))
	}
	found := false
	for _, raw := range tools {
		tool := raw.(map[string]any)
		if tool["name"] == "inspect_world" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("inspect_world missing")
	}
}

func TestMCPToolCall(t *testing.T) {
	s, run := newMCPTestServer(t)
	args := map[string]any{"runId": run.ID, "visibility": "gm"}
	rr := doMCP(t, s, "tools/call", "inspect_world", map[string]any{"name": "inspect_world", "arguments": args}, run.ID)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	result := out["result"].(map[string]any)
	if result["isError"] == true {
		t.Fatalf("tool error: %v", result)
	}
	structured := result["structuredContent"].(map[string]any)
	if structured["runId"] != run.ID {
		t.Fatalf("wrong run id: %v", structured)
	}
}

func TestMCPRunHeaderMismatch(t *testing.T) {
	s, run := newMCPTestServer(t)
	args := map[string]any{"runId": run.ID}
	rr := doMCP(t, s, "tools/call", "inspect_world", map[string]any{"name": "inspect_world", "arguments": args}, "wrong-run")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	errObj := out["error"].(map[string]any)
	if int(errObj["code"].(float64)) != -32020 {
		t.Fatalf("unexpected error: %v", errObj)
	}
}

func TestAgentGatewayAuth(t *testing.T) {
	t.Setenv("ASHEN_AGENT_TOKEN", "secret-token")
	store := game.NewStore(t.TempDir())
	engine := game.NewEngine()
	s := New(engine, store, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/agent/tools", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/agent/tools", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAgentToolCallFacade(t *testing.T) {
	s, run := newMCPTestServer(t)
	body, _ := json.Marshal(map[string]any{
		"runId":      run.ID,
		"visibility": "gm",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/agent/tools/call/inspect_world", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["success"] != true {
		t.Fatalf("unexpected facade response: %v", out)
	}
	if out["tool"] != "inspect_world" || out["runId"] != run.ID {
		t.Fatalf("unexpected facade response: %v", out)
	}
	if _, wrapped := out["result"]; wrapped {
		t.Fatalf("facade must return flattened result, got: %v", out)
	}
}

func TestAgentToolCallFacadeBusinessErrorUsesHTTP200(t *testing.T) {
	s, run := newMCPTestServer(t)
	body, _ := json.Marshal(map[string]any{"runId": run.ID})
	req := httptest.NewRequest(http.MethodPost, "/api/agent/tools/call/not_a_tool", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 business error, got %d body=%s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["success"] != false {
		t.Fatalf("expected success=false, got %v", out)
	}
	if out["message"] == "" {
		t.Fatalf("expected model-readable error message, got %v", out)
	}
}

package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"ashen-crown-dungeon/internal/agentgateway"
	"ashen-crown-dungeon/internal/game"
)

func newAgentGateway(engine *game.Engine, store *game.Store) *agentgateway.Gateway {
	requireReason := true
	if v := strings.TrimSpace(os.Getenv("ASHEN_AGENT_REQUIRE_REASON")); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			requireReason = parsed
		}
	}
	return agentgateway.New(engine, store, agentgateway.Config{
		Token:                    strings.TrimSpace(os.Getenv("ASHEN_AGENT_TOKEN")),
		RequireReasonForHighRisk: requireReason,
	})
}

func (s *Server) agentTools(w http.ResponseWriter, r *http.Request) {
	if !s.requireAgentAuth(w, r) {
		return
	}
	jsonOut(w, http.StatusOK, map[string]any{
		"gatewayVersion": "1.0",
		"gameVersion":    "0.7.0",
		"tools":          s.agent.Tools(),
	})
}

func (s *Server) agentToolDefinition(w http.ResponseWriter, r *http.Request) {
	if !s.requireAgentAuth(w, r) {
		return
	}
	def, ok := s.agent.Definition(r.PathValue("name"))
	if !ok {
		errOut(w, http.StatusNotFound, "Agent Tool 不存在")
		return
	}
	jsonOut(w, http.StatusOK, def)
}

func (s *Server) agentExecute(w http.ResponseWriter, r *http.Request) {
	if !s.requireAgentAuth(w, r) {
		return
	}
	var in agentgateway.ExecuteRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.UseNumber()
	if err := dec.Decode(&in); err != nil {
		errOut(w, http.StatusBadRequest, "请求格式错误: "+err.Error())
		return
	}
	result, err := s.agent.Execute(r.Context(), in, callerFromHTTP(r, "http-gateway"))
	if err != nil {
		errOut(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOut(w, http.StatusOK, agentgateway.ExecuteResponse{Result: result})
}

// agentToolCall is the HTTP facade used by generic API-to-MCP gateways.
//
// The tool name comes from the URL and the request body is the tool arguments
// directly, so an upstream gateway can register every GM action as an
// independent MCP tool without wrapping the payload as {"tool":...,"arguments":...}.
//
// Business errors intentionally use HTTP 200 with success=false. This lets the
// generic mcp-server preserve the upstream message instead of collapsing it to
// a generic "upstream HTTP status 400" error. Authentication and malformed HTTP
// requests still use normal HTTP error status codes.
func (s *Server) agentToolCall(w http.ResponseWriter, r *http.Request) {
	if !s.requireAgentAuth(w, r) {
		return
	}

	name := strings.TrimSpace(r.PathValue("name"))
	if name == "" {
		errOut(w, http.StatusBadRequest, "Agent Tool 名称不能为空")
		return
	}

	args := map[string]any{}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.UseNumber()
	if err := dec.Decode(&args); err != nil {
		errOut(w, http.StatusBadRequest, "请求格式错误: "+err.Error())
		return
	}

	result, err := s.agent.Execute(r.Context(), agentgateway.ExecuteRequest{
		Tool:      name,
		Arguments: args,
	}, callerFromHTTP(r, "http-tool-facade"))
	if err != nil {
		runID, _ := args["runId"].(string)
		jsonOut(w, http.StatusOK, map[string]any{
			"success": false,
			"code":    4000,
			"message": err.Error(),
			"tool":    name,
			"runId":   strings.TrimSpace(runID),
			"changed": false,
			"summary": "工具执行失败",
		})
		return
	}

	out := map[string]any{
		"success": true,
		"code":    0,
		"message": result.Summary,
		"tool":    result.Tool,
		"runId":   result.RunID,
		"changed": result.Changed,
		"summary": result.Summary,
	}
	if result.DryRun {
		out["dryRun"] = true
	}
	if result.Data != nil {
		out["data"] = result.Data
	}
	if len(result.Warnings) > 0 {
		out["warnings"] = result.Warnings
	}
	if result.AuditID != "" {
		out["auditId"] = result.AuditID
	}
	jsonOut(w, http.StatusOK, out)
}

func (s *Server) agentAudit(w http.ResponseWriter, r *http.Request) {
	if !s.requireAgentAuth(w, r) {
		return
	}
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	records, err := s.agent.RecentAudit(strings.TrimSpace(r.URL.Query().Get("runId")), limit)
	if err != nil {
		errOut(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOut(w, http.StatusOK, map[string]any{"records": records})
}

func (s *Server) requireAgentAuth(w http.ResponseWriter, r *http.Request) bool {
	token := s.agent.Token()
	if token == "" {
		return true
	}
	if r.Header.Get("Authorization") != "Bearer "+token {
		w.Header().Set("WWW-Authenticate", `Bearer realm="ashen-crown-agent"`)
		errOut(w, http.StatusUnauthorized, "Agent Gateway 未授权")
		return false
	}
	return true
}

func callerFromHTTP(r *http.Request, transport string) agentgateway.Caller {
	return agentgateway.Caller{
		Name:       strings.TrimSpace(r.Header.Get("X-Agent-Name")),
		Transport:  transport,
		ClientName: strings.TrimSpace(r.Header.Get("User-Agent")),
		TraceID:    strings.TrimSpace(r.Header.Get("traceparent")),
	}
}

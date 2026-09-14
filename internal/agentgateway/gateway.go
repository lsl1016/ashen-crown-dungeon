package agentgateway

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ashen-crown-dungeon/internal/game"
)

type Config struct {
	Token                    string
	RequireReasonForHighRisk bool
	AuditDir                 string
}

type Caller struct {
	Name       string `json:"name,omitempty"`
	Transport  string `json:"transport,omitempty"`
	ClientName string `json:"clientName,omitempty"`
	TraceID    string `json:"traceId,omitempty"`
}

type ExecuteRequest struct {
	Tool      string         `json:"tool"`
	Arguments map[string]any `json:"arguments"`
}

type ExecuteResponse struct {
	Result game.AgentToolResult `json:"result"`
}

type AuditRecord struct {
	ID         string         `json:"id"`
	Timestamp  time.Time      `json:"timestamp"`
	DurationMs int64          `json:"durationMs"`
	Success    bool           `json:"success"`
	Tool       string         `json:"tool"`
	RunID      string         `json:"runId,omitempty"`
	Risk       string         `json:"risk"`
	DryRun     bool           `json:"dryRun"`
	Changed    bool           `json:"changed"`
	Reason     string         `json:"reason,omitempty"`
	Caller     Caller         `json:"caller,omitempty"`
	Arguments  map[string]any `json:"arguments,omitempty"`
	Summary    string         `json:"summary,omitempty"`
	Error      string         `json:"error,omitempty"`
}

type Gateway struct {
	engine *game.Engine
	store  *game.Store
	cfg    Config

	lockMu  sync.Mutex
	locks   map[string]*sync.Mutex
	auditMu sync.Mutex
}

func New(engine *game.Engine, store *game.Store, cfg Config) *Gateway {
	if cfg.AuditDir == "" {
		cfg.AuditDir = filepath.Join(store.Root(), "audit")
	}
	_ = os.MkdirAll(cfg.AuditDir, 0755)
	return &Gateway{engine: engine, store: store, cfg: cfg, locks: map[string]*sync.Mutex{}}
}

func (g *Gateway) Token() string                     { return g.cfg.Token }
func (g *Gateway) Tools() []game.AgentToolDefinition { return game.AgentToolDefinitions() }
func (g *Gateway) Definition(name string) (game.AgentToolDefinition, bool) {
	return game.AgentToolDefinitionByName(name)
}

func (g *Gateway) Execute(ctx context.Context, req ExecuteRequest, caller Caller) (game.AgentToolResult, error) {
	started := time.Now()
	req.Tool = strings.TrimSpace(req.Tool)
	if req.Arguments == nil {
		req.Arguments = map[string]any{}
	}
	def, ok := game.AgentToolDefinitionByName(req.Tool)
	if !ok {
		return game.AgentToolResult{}, fmt.Errorf("未知 Agent Tool: %s", req.Tool)
	}
	if err := game.ValidateAgentArguments(def, req.Arguments); err != nil {
		return game.AgentToolResult{}, err
	}
	runID, _ := req.Arguments["runId"].(string)
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return game.AgentToolResult{}, errors.New("runId 必填")
	}
	dryRun, _ := req.Arguments["dryRun"].(bool)
	reason, _ := req.Arguments["reason"].(string)
	reason = strings.TrimSpace(reason)
	risk := game.AgentToolRisk(req.Tool)
	if g.cfg.RequireReasonForHighRisk && risk == "high" && !game.AgentToolReadOnly(req.Tool) && reason == "" {
		return game.AgentToolResult{}, errors.New("高风险 Tool 必须提供 reason")
	}

	mu := g.runLock(runID)
	mu.Lock()
	defer mu.Unlock()

	run, err := g.store.LoadRun(runID)
	if err != nil {
		g.writeAudit(AuditRecord{ID: newAuditID(), Timestamp: time.Now(), DurationMs: time.Since(started).Milliseconds(), Success: false, Tool: req.Tool, RunID: runID, Risk: risk, DryRun: dryRun, Reason: reason, Caller: caller, Arguments: sanitizedArgs(req.Arguments), Error: "Run 不存在"})
		return game.AgentToolResult{}, fmt.Errorf("Run 不存在: %s", runID)
	}
	target := run
	if dryRun {
		target, err = cloneRun(run)
		if err != nil {
			return game.AgentToolResult{}, err
		}
	}
	result, execErr := g.engine.ExecuteAgentTool(target, req.Tool, req.Arguments)
	auditID := newAuditID()
	result.DryRun = dryRun
	result.Risk = risk
	result.AuditID = auditID
	if execErr == nil && result.Changed && !dryRun {
		g.engine.Prepare(target)
		execErr = g.store.SaveRun(target)
	}
	record := AuditRecord{
		ID: auditID, Timestamp: time.Now(), DurationMs: time.Since(started).Milliseconds(), Success: execErr == nil,
		Tool: req.Tool, RunID: runID, Risk: risk, DryRun: dryRun, Changed: result.Changed,
		Reason: reason, Caller: caller, Arguments: sanitizedArgs(req.Arguments), Summary: result.Summary,
	}
	if execErr != nil {
		record.Error = execErr.Error()
	}
	g.writeAudit(record)
	if execErr != nil {
		return result, execErr
	}
	select {
	case <-ctx.Done():
		return result, ctx.Err()
	default:
		return result, nil
	}
}

func (g *Gateway) RecentAudit(runID string, limit int) ([]AuditRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	path := filepath.Join(g.cfg.AuditDir, "agent-tools.jsonl")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return []AuditRecord{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	all := []AuditRecord{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		var rec AuditRecord
		if json.Unmarshal(scanner.Bytes(), &rec) != nil {
			continue
		}
		if runID == "" || rec.RunID == runID {
			all = append(all, rec)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(all) > limit {
		all = all[len(all)-limit:]
	}
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	return all, nil
}

func (g *Gateway) runLock(runID string) *sync.Mutex {
	g.lockMu.Lock()
	defer g.lockMu.Unlock()
	if mu, ok := g.locks[runID]; ok {
		return mu
	}
	mu := &sync.Mutex{}
	g.locks[runID] = mu
	return mu
}

func (g *Gateway) writeAudit(rec AuditRecord) {
	g.auditMu.Lock()
	defer g.auditMu.Unlock()
	_ = os.MkdirAll(g.cfg.AuditDir, 0755)
	f, err := os.OpenFile(filepath.Join(g.cfg.AuditDir, "agent-tools.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	b, err := json.Marshal(rec)
	if err != nil {
		return
	}
	_, _ = f.Write(append(b, '\n'))
}

func cloneRun(src *game.Run) (*game.Run, error) {
	b, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	var dst game.Run
	if err := json.Unmarshal(b, &dst); err != nil {
		return nil, err
	}
	return &dst, nil
}

func sanitizedArgs(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		if strings.EqualFold(k, "authorization") || strings.Contains(strings.ToLower(k), "token") {
			out[k] = "[redacted]"
			continue
		}
		out[k] = v
	}
	return out
}

func newAuditID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err == nil {
		return "audit_" + hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("audit_%d", time.Now().UnixNano())
}

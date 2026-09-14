package agentgateway

import (
	"context"
	"path/filepath"
	"testing"

	"ashen-crown-dungeon/internal/game"
)

func newTestGateway(t *testing.T) (*Gateway, *game.Store, *game.Run) {
	t.Helper()
	root := t.TempDir()
	store := game.NewStore(root)
	engine := game.NewEngine()
	run, err := engine.NewRun("AgentTest", "warden", 7007)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveRun(run); err != nil {
		t.Fatal(err)
	}
	gw := New(engine, store, Config{RequireReasonForHighRisk: true, AuditDir: filepath.Join(root, "audit")})
	return gw, store, run
}

func TestDryRunDoesNotPersist(t *testing.T) {
	gw, store, run := newTestGateway(t)
	before := run.NPCLocations["iven"]
	target := "room_01"
	if before == target {
		target = "room_02"
	}
	result, err := gw.Execute(context.Background(), ExecuteRequest{Tool: "move_npc", Arguments: map[string]any{
		"runId": run.ID, "npcId": "iven", "roomId": target, "dryRun": true, "reason": "preview movement",
	}}, Caller{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.DryRun || !result.Changed {
		t.Fatalf("expected changed dry-run, got %+v", result)
	}
	persisted, err := store.LoadRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.NPCLocations["iven"] != before {
		t.Fatalf("dry-run persisted NPC move: %s -> %s", before, persisted.NPCLocations["iven"])
	}
}

func TestMutationPersistsAndAudits(t *testing.T) {
	gw, store, run := newTestGateway(t)
	before := run.NPCLocations["iven"]
	target := "room_01"
	if before == target {
		target = "room_02"
	}
	result, err := gw.Execute(context.Background(), ExecuteRequest{Tool: "move_npc", Arguments: map[string]any{
		"runId": run.ID, "npcId": "iven", "roomId": target, "reason": "story relocation",
	}}, Caller{Name: "test-agent"})
	if err != nil {
		t.Fatal(err)
	}
	if result.AuditID == "" {
		t.Fatal("audit id missing")
	}
	persisted, err := store.LoadRun(run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.NPCLocations["iven"] != target {
		t.Fatalf("expected %s, got %s", target, persisted.NPCLocations["iven"])
	}
	records, err := gw.RecentAudit(run.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) == 0 || records[0].Tool != "move_npc" || !records[0].Success {
		t.Fatalf("unexpected audit records: %+v", records)
	}
}

func TestHighRiskRequiresReason(t *testing.T) {
	gw, _, run := newTestGateway(t)
	_, err := gw.Execute(context.Background(), ExecuteRequest{Tool: "advance_bell", Arguments: map[string]any{
		"runId": run.ID, "steps": 1,
	}}, Caller{Name: "test"})
	if err == nil {
		t.Fatal("expected high-risk reason error")
	}
	if _, err := gw.Execute(context.Background(), ExecuteRequest{Tool: "advance_bell", Arguments: map[string]any{
		"runId": run.ID, "steps": 1, "dryRun": true, "reason": "preview world time",
	}}, Caller{Name: "test"}); err != nil {
		t.Fatalf("dry-run with reason failed: %v", err)
	}
}

func TestSchemaValidationRejectsUnknownField(t *testing.T) {
	gw, _, run := newTestGateway(t)
	_, err := gw.Execute(context.Background(), ExecuteRequest{Tool: "inspect_world", Arguments: map[string]any{
		"runId": run.ID, "bogus": true,
	}}, Caller{Name: "test"})
	if err == nil {
		t.Fatal("expected schema validation error")
	}
}

package game

import "testing"

func TestDeterministicDungeon(t *testing.T) {
	a, ae := GenerateDungeon(42)
	b, be := GenerateDungeon(42)
	if len(a) != len(b) || len(ae) != len(be) {
		t.Fatal("same seed should generate same size")
	}
	for id, ra := range a {
		rb := b[id]
		if rb == nil || ra.Type != rb.Type || ra.Name != rb.Name || ra.Scene != rb.Scene {
			t.Fatalf("room mismatch %s", id)
		}
		if len(ra.Elements) != len(rb.Elements) {
			t.Fatalf("element mismatch %s", id)
		}
	}
}

func TestBasicFlowAndInteraction(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Tester", "warden", 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Rooms) != 32 {
		t.Fatalf("want 32 rooms, got %d", len(r.Rooms))
	}
	if len(r.Rooms["room_01"].Elements) < 2 {
		t.Fatal("entrance should expose interactive elements")
	}
	if err := e.Interact(r, "traveler_pack"); err != nil {
		t.Fatal(err)
	}
	if !hasItem(r.Player.Inventory, "bandage") {
		t.Fatal("interaction should grant bandage")
	}
	if err := e.Interact(r, "traveler_pack"); err == nil {
		t.Fatal("one-shot element should not be reusable")
	}

	var next string
	for _, ed := range r.Edges {
		if ed.From == r.CurrentRoomID {
			next = ed.To
			break
		}
		if ed.To == r.CurrentRoomID {
			next = ed.From
			break
		}
	}
	if next == "" {
		t.Fatal("no adjacent room")
	}
	if err := e.Move(r, next); err != nil {
		t.Fatal(err)
	}
}

func TestSecretRoomStartsLocked(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Tester", "seer", 99)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"room_24", "room_31", "room_32"} {
		secret := r.Rooms[id]
		if secret == nil || !secret.Locked || secret.Discovered {
			t.Fatalf("%s should start locked and undiscovered", id)
		}
	}
}

func TestAmbientInteractionWritesLog(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Tester", "ranger", 4242)
	if err != nil {
		t.Fatal(err)
	}
	var ambientID string
	for _, el := range r.Rooms[r.CurrentRoomID].Elements {
		if el.Action == "message" {
			ambientID = el.ID
			break
		}
	}
	if ambientID == "" {
		t.Fatal("every room should have an ambient interaction")
	}
	before := len(r.Log)
	if err := e.Interact(r, ambientID); err != nil {
		t.Fatal(err)
	}
	if len(r.Log) <= before {
		t.Fatal("ambient interaction should append an exploration log")
	}
}

func TestEventCombatClearsEventPanel(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Tester", "warden", 7)
	if err != nil {
		t.Fatal(err)
	}
	// Directly open an event with a deterministic combat-producing failure choice.
	def := Events["drowned_bell"]
	r.ActiveEvent = &ActiveEvent{ID: def.ID, Title: def.Title, Description: def.Description, Choices: append([]ChoiceDef(nil), def.Choices...)}
	// Make the failure path deterministic without depending on a roll by injecting
	// a temporary no-check choice that starts combat through the same effect pipeline.
	r.ActiveEvent.Choices = []ChoiceDef{{ID: "fight", Text: "触发战斗", OnSuccess: []Effect{{Type: "combat", Target: "drowned_bellman"}}}}
	if err := e.ResolveChoice(r, "fight"); err != nil {
		t.Fatal(err)
	}
	if r.Combat == nil {
		t.Fatal("event should start combat")
	}
	if r.ActiveEvent != nil {
		t.Fatal("event panel must be cleared once combat starts")
	}
}

func TestTalentAndEquipmentBuild(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Builder", "warden", 101)
	if err != nil {
		t.Fatal(err)
	}
	if r.Player.TalentPoints != 1 {
		t.Fatalf("expected 1 starting talent point, got %d", r.Player.TalentPoints)
	}
	beforeHP := r.Player.MaxHP
	if err := e.UnlockTalent(r, "warden_vitality"); err != nil {
		t.Fatal(err)
	}
	if r.Player.MaxHP != beforeHP+8 || r.Player.TalentPoints != 0 {
		t.Fatal("vitality talent should raise max hp and consume talent point")
	}
	if r.Player.Equipment["armor"] != "chainmail" || r.Player.Defense < 14 {
		t.Fatal("warden should start with equipped chainmail and its defense bonus")
	}
	r.Player.Inventory = append(r.Player.Inventory, "night_badge")
	beforeDefense := r.Player.Defense
	if err := e.ItemAction(r, "night_badge", "equip"); err != nil {
		t.Fatal(err)
	}
	if r.Player.Equipment["trinket"] != "night_badge" || r.Player.Defense != beforeDefense+1 {
		t.Fatal("trinket equipment bonus should apply")
	}
}

func TestCombatIntentAndBossPhase(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Fighter", "seer", 202)
	if err != nil {
		t.Fatal(err)
	}
	e.startCombat(r, "crown_bearer")
	if r.Combat == nil || r.Combat.Intent.ID == "" || r.Combat.BossPhase != 1 {
		t.Fatal("combat should expose an enemy intent immediately")
	}
	r.Combat.EnemyHP = r.Combat.EnemyMaxHP / 2
	messages := []string{}
	e.updateBossPhase(r, Enemies["crown_bearer"], &messages)
	if r.Combat.BossPhase != 2 {
		t.Fatalf("expected boss phase 2, got %d", r.Combat.BossPhase)
	}
}

func TestConditionalSceneInteraction(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Explorer", "ranger", 303)
	if err != nil {
		t.Fatal(err)
	}
	r.CurrentRoomID = "room_26" // 折冠门楼
	r.Rooms["room_26"].Discovered = true
	r.Rooms["room_26"].Visited = true
	if err := e.Interact(r, "watch_locker"); err == nil {
		t.Fatal("locker should require the old night badge")
	}
	r.Player.Inventory = append(r.Player.Inventory, "night_badge")
	if err := e.Interact(r, "watch_locker"); err != nil {
		t.Fatal(err)
	}
	if !hasItem(r.Player.Inventory, "chainmail") {
		t.Fatal("conditional interaction should grant chainmail")
	}
}

func TestWorldClockAdvances(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Walker", "ranger", 404)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		e.advanceClock(r)
	}
	if r.Clock.Bell != 2 || r.Clock.Steps != 5 {
		t.Fatalf("expected bell 2 after five steps, got bell=%d steps=%d", r.Clock.Bell, r.Clock.Steps)
	}
}

func TestCombatConsumableConsumesTurn(t *testing.T) {
	e := NewEngine()
	run, err := e.NewRun("test", "ranger", 5150)
	if err != nil {
		t.Fatal(err)
	}
	run.Player.Inventory = append(run.Player.Inventory, "frost_salt")
	e.startCombat(run, "bone_thrall")
	if run.Combat == nil {
		t.Fatal("combat should start")
	}
	round := run.Combat.Round
	if err := e.ItemAction(run, "frost_salt", "use"); err != nil {
		t.Fatal(err)
	}
	if run.Combat != nil && run.Combat.Round != round+1 {
		t.Fatalf("combat consumable should consume one round: got %d want %d", run.Combat.Round, round+1)
	}
	if hasItem(run.Player.Inventory, "frost_salt") {
		t.Fatal("combat consumable should be consumed")
	}
}

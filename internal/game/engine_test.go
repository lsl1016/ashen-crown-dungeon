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
	if len(r.Rooms) != 44 {
		t.Fatalf("want 44 rooms, got %d", len(r.Rooms))
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

func TestV04ActTwoStartsLocked(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("ActTester", "warden", 8080)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Rooms) != 44 {
		t.Fatalf("want 44 rooms in V0.4, got %d", len(r.Rooms))
	}
	if room := r.Rooms["room_33"]; room == nil || !room.Locked || room.Discovered {
		t.Fatal("act two frontier should start locked and undiscovered")
	}
	if room := r.Rooms["room_44"]; room == nil || !room.Locked || room.Discovered {
		t.Fatal("final heart chamber should start locked and undiscovered")
	}
}

func TestV04CrownBearerUnlocksActTwoWithoutEndingRun(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("ActTester", "warden", 8181)
	if err != nil {
		t.Fatal(err)
	}
	r.CurrentRoomID = "room_09"
	e.startCombat(r, "crown_bearer")
	e.winCombat(r, Enemies["crown_bearer"])
	if r.GameOver || r.Victory {
		t.Fatal("first boss must not end V0.4 campaign")
	}
	if !r.Flags["act2_unlocked"] {
		t.Fatal("act two flag should be set")
	}
	if room := r.Rooms["room_33"]; room == nil || room.Locked || !room.Discovered {
		t.Fatal("frontier should unlock after first boss")
	}
	if q := r.Quests["beyond_mist"]; q == nil || q.Status != "active" {
		t.Fatal("act two quest should become active")
	}
}

func TestV04FinalBossEndsCampaign(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("FinalTester", "seer", 8282)
	if err != nil {
		t.Fatal(err)
	}
	r.CurrentRoomID = "room_44"
	e.startCombat(r, "gate_heart")
	e.winCombat(r, Enemies["gate_heart"])
	if !r.GameOver || !r.Victory || !r.Flags["defeated_gate_heart"] {
		t.Fatal("final boss should finish campaign")
	}
}

func TestV04OuterGateRequiresCompassAndCanUnlockFinalRoom(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("GateTester", "seer", 8383)
	if err != nil {
		t.Fatal(err)
	}
	r.CurrentRoomID = "room_43"
	r.Rooms["room_43"].Discovered = true
	r.Rooms["room_43"].Visited = true
	if err := e.Interact(r, "outer_gate_line"); err == nil {
		t.Fatal("outer gate should require the void compass")
	}
	r.Player.Inventory = append(r.Player.Inventory, "void_compass")
	r.Player.Attributes.Will = 100 // make the DC check deterministic for this progression test
	if err := e.Interact(r, "outer_gate_line"); err != nil {
		t.Fatal(err)
	}
	if room := r.Rooms["room_44"]; room == nil || room.Locked || !room.Discovered {
		t.Fatal("final heart chamber should unlock after a successful gate check")
	}
	if !r.Flags["opened_outer_gate"] {
		t.Fatal("successful gate interaction should persist its world flag")
	}
}

func TestV05DialogueQuestAndRelationship(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Speaker", "ranger", 5050)
	if err != nil {
		t.Fatal(err)
	}
	r.CurrentRoomID = "room_13"
	r.Rooms["room_13"].Discovered = true
	r.Rooms["room_13"].Visited = true
	if err := e.Interact(r, "npc_iven"); err != nil {
		t.Fatal(err)
	}
	if r.ActiveDialogue == nil || r.ActiveDialogue.NPCID != "iven" {
		t.Fatal("Iven dialogue should become active")
	}
	if err := e.DialogueChoice(r, "patrol"); err != nil {
		t.Fatal(err)
	}
	if q := r.Quests["lost_patrol"]; q == nil || q.Status != "active" {
		t.Fatal("dialogue choice should accept lost patrol quest")
	}
	if r.NPCRelations["iven"] != 2 {
		t.Fatalf("expected Iven relationship +2, got %d", r.NPCRelations["iven"])
	}
	if r.ActiveDialogue == nil || r.ActiveDialogue.NodeID != "accept" {
		t.Fatal("dialogue should advance to accept node")
	}
}

func TestV05ShopBuyAndSell(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Trader", "warden", 5051)
	if err != nil {
		t.Fatal(err)
	}
	r.Player.Gold = 200
	r.CurrentRoomID = "room_14"
	r.Rooms["room_14"].Discovered = true
	r.Rooms["room_14"].Visited = true
	if err := e.StartDialogue(r, "veil_broker"); err != nil {
		t.Fatal(err)
	}
	if err := e.DialogueChoice(r, "trade"); err != nil {
		t.Fatal(err)
	}
	if r.ActiveShop != "veil_market" {
		t.Fatal("merchant dialogue should open veil market")
	}
	beforeStock := r.ShopStock["veil_market:healing_draught"]
	beforeGold := r.Player.Gold
	if err := e.ShopAction(r, "veil_market", "buy", "healing_draught"); err != nil {
		t.Fatal(err)
	}
	if !hasItem(r.Player.Inventory, "healing_draught") || r.ShopStock["veil_market:healing_draught"] != beforeStock-1 || r.Player.Gold >= beforeGold {
		t.Fatal("buy should update inventory, stock, and gold")
	}
	goldAfterBuy := r.Player.Gold
	if err := e.ShopAction(r, "veil_market", "sell", "healing_draught"); err != nil {
		t.Fatal(err)
	}
	if r.Player.Gold <= goldAfterBuy {
		t.Fatal("sell should award gold")
	}
	if r.ShopStock["veil_market:healing_draught"] != beforeStock {
		t.Fatal("sold item should return to merchant stock")
	}
}

func TestV05GrowthPointsAndMastery(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Grower", "seer", 5052)
	if err != nil {
		t.Fatal(err)
	}
	if r.Player.AttributePoints != 1 || r.Player.MasteryPoints != 1 {
		t.Fatalf("new run should start with growth points, got attr=%d mastery=%d", r.Player.AttributePoints, r.Player.MasteryPoints)
	}
	oldWill := r.Player.Attributes.Will
	if err := e.UpgradeAttribute(r, "will"); err != nil {
		t.Fatal(err)
	}
	if r.Player.Attributes.Will != oldWill+1 || r.Player.AttributePoints != 0 {
		t.Fatal("attribute upgrade should consume point and increase stat")
	}
	oldHP := r.Player.MaxHP
	if err := e.UpgradeGrowth(r, "survivor_instinct"); err != nil {
		t.Fatal(err)
	}
	if r.Player.Growth["survivor_instinct"] != 1 || r.Player.MaxHP != oldHP+3 || r.Player.MasteryPoints != 0 {
		t.Fatal("survivor mastery should persist rank and increase max hp")
	}
}

func TestV05SignatureMasteryReducesSkillCostAtRankThree(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Master", "seer", 5053)
	if err != nil {
		t.Fatal(err)
	}
	r.Player.Growth["signature_mastery"] = 3
	r.Player.Energy = r.Player.MaxEnergy
	r.Player.Attributes.Will = 100
	e.startCombat(r, "bone_thrall")
	before := r.Player.Energy
	if err := e.CombatAction(r, "skill"); err != nil {
		t.Fatal(err)
	}
	if r.Player.Energy != before-2 {
		t.Fatalf("rank-three signature mastery should cost 2 energy: got %d from %d", r.Player.Energy, before)
	}
}

func TestV06NPCScheduleAndRegionState(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Watcher", "ranger", 6060)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.NPCLocations["iven"]; got != "room_13" {
		t.Fatalf("Iven should start at patrol camp, got %s", got)
	}
	r.Clock.Bell = 7
	e.onBellChanged(r)
	if got := r.NPCLocations["iven"]; got != "room_25" {
		t.Fatalf("Iven should migrate to infirmary at bell 7, got %s", got)
	}
	if got := r.RegionStates["外墓区"]; got != "灰雾加深 · 巡夜线失联" {
		t.Fatalf("unexpected outer cemetery state: %s", got)
	}
	r.Flags["lost_patrol_reported"] = true
	e.syncRegionStates(r)
	e.syncNPCLocations(r, false)
	if got := r.NPCLocations["iven"]; got != "room_01" {
		t.Fatalf("reported patrol should move Iven to entrance, got %s", got)
	}
	if got := r.RegionStates["外墓区"]; got != "巡夜团重新建立哨线" {
		t.Fatalf("reported patrol should stabilize region, got %s", got)
	}
}

func TestV06ShopRefreshesAfterThreeBells(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Restocker", "warden", 6061)
	if err != nil {
		t.Fatal(err)
	}
	key := "veil_market:healing_draught"
	base := r.ShopStock[key]
	if base < 1 {
		t.Fatal("expected healing draught stock")
	}
	r.ShopStock[key] = 0
	r.ShopRefresh["veil_market"] = 1
	r.Clock.Bell = 4
	e.refreshShops(r)
	if r.ShopStock[key] <= 0 {
		t.Fatal("merchant should restock consumables after three bells")
	}
	if r.ShopRefresh["veil_market"] != 4 {
		t.Fatal("shop refresh bell should persist")
	}
}

func TestV06TalentPrerequisites(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Tree", "warden", 6062)
	if err != nil {
		t.Fatal(err)
	}
	r.Player.Level = 4
	r.Player.TalentPoints = 4
	if err := e.UnlockTalent(r, "warden_execution"); err == nil {
		t.Fatal("tier-two talent should require its prerequisite")
	}
	if err := e.UnlockTalent(r, "warden_bulwark"); err != nil {
		t.Fatal(err)
	}
	if err := e.UnlockTalent(r, "warden_execution"); err != nil {
		t.Fatal(err)
	}
	if err := e.UnlockTalent(r, "warden_march"); err != nil {
		t.Fatal(err)
	}
	if !e.hasTalent(r, "warden_march") {
		t.Fatal("capstone talent should unlock after prerequisites")
	}
}

func TestV06AffixAppliesWhenEquipped(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Affixed", "ranger", 6063)
	if err != nil {
		t.Fatal(err)
	}
	r.Player.Inventory = append(r.Player.Inventory, "graveglass_edge")
	r.ItemAffixes["graveglass_edge"] = []AffixState{{ID: "keen", Name: "锋锐", Power: 1, Attributes: AttributeSet{Dexterity: 1}}}
	beforeDex := r.Player.Attributes.Dexterity
	if err := e.equipItem(r, "graveglass_edge"); err != nil {
		t.Fatal(err)
	}
	if r.Player.Attributes.Dexterity != beforeDex+1 {
		t.Fatalf("affix attribute should apply on equip: before=%d after=%d", beforeDex, r.Player.Attributes.Dexterity)
	}
	if got := e.weaponPower(r); got < Items["graveglass_edge"].Power+1 {
		t.Fatalf("weapon affix power should participate in damage power, got %d", got)
	}
}

func TestV06CombatDistanceActions(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Spacer", "ranger", 6064)
	if err != nil {
		t.Fatal(err)
	}
	r.Player.HP = 999
	r.Player.MaxHP = 999
	e.startCombat(r, "bone_thrall")
	if r.Combat.Distance != 2 {
		t.Fatalf("ranger should start at middle distance, got %d", r.Combat.Distance)
	}
	// Force a non-melee enemy intent so the enemy cannot immediately undo the retreat.
	r.Combat.Intent = EnemyIntent{ID: "defend", Kind: "defend", Label: "防御架势"}
	if err := e.CombatAction(r, "retreat"); err != nil {
		t.Fatal(err)
	}
	if r.Combat == nil || r.Combat.Distance != 3 {
		t.Fatalf("retreat should reach far distance, got %+v", r.Combat)
	}
	r.Combat.Intent = EnemyIntent{ID: "defend", Kind: "defend", Label: "防御架势"}
	if err := e.CombatAction(r, "advance"); err != nil {
		t.Fatal(err)
	}
	if r.Combat == nil || r.Combat.Distance != 2 {
		t.Fatalf("advance should return to middle distance, got %+v", r.Combat)
	}
}

func TestV06QuestTurnInPersistsOutcomeAndMovesNPC(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Reporter", "ranger", 6065)
	if err != nil {
		t.Fatal(err)
	}
	r.Quests["lost_patrol"] = &QuestState{ID: "lost_patrol", Title: "失踪的第三巡夜队", Status: "completed", Stage: "return", Progress: 1, Goal: 1}
	r.CurrentRoomID = "room_13"
	if err := e.StartDialogue(r, "iven"); err != nil {
		t.Fatal(err)
	}
	if r.ActiveDialogue == nil || r.ActiveDialogue.NodeID != "return" {
		t.Fatalf("completed quest should start at return node: %+v", r.ActiveDialogue)
	}
	if err := e.DialogueChoice(r, "report"); err != nil {
		t.Fatal(err)
	}
	e.Prepare(r)
	q := r.Quests["lost_patrol"]
	if q.Outcome != "reported" || q.Stage != "resolved" {
		t.Fatalf("quest outcome should persist, got %+v", q)
	}
	if r.NPCLocations["iven"] != "room_01" {
		t.Fatalf("turn-in should move Iven to entrance, got %s", r.NPCLocations["iven"])
	}
	if !hasItem(r.Player.Inventory, "patrol_charm") {
		t.Fatal("turn-in should grant patrol charm")
	}
}

func TestV06PrepareNormalizesOldSave(t *testing.T) {
	e := NewEngine()
	r, err := e.NewRun("Legacy", "seer", 6066)
	if err != nil {
		t.Fatal(err)
	}
	r.NPCLocations = nil
	r.RegionStates = nil
	r.ShopRefresh = nil
	r.ItemAffixes = nil
	e.Prepare(r)
	if len(r.NPCLocations) == 0 || len(r.RegionStates) == 0 || len(r.ShopRefresh) == 0 || r.ItemAffixes == nil {
		t.Fatal("Prepare should normalize all V0.6 persistent maps")
	}
}

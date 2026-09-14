package game

import (
	"fmt"
	"sort"
)

func initialNPCLocations() map[string]string {
	return map[string]string{
		"iven":              "room_13",
		"veil_broker":       "room_14",
		"nameless_prisoner": "room_07",
		"orrin":             "room_41",
		"mara":              "room_42",
	}
}

func initialRegionStates() map[string]string {
	return map[string]string{
		"外墓区":  "灰雾封锁",
		"旧城区":  "残响不稳定",
		"王城中环": "封印重排",
		"内廷":   "王庭警戒",
		"烬冠核心": "黑火脉动",
		"雾外荒原": "尚未抵达",
		"荒原东界": "黑门封闭",
	}
}

func initialShopRefresh(bell int) map[string]int {
	m := map[string]int{}
	for id := range Shops {
		m[id] = bell
	}
	return m
}

func (e *Engine) ensureV06State(run *Run) {
	e.ensureV05State(run)
	if run.NPCLocations == nil {
		run.NPCLocations = initialNPCLocations()
	}
	if run.RegionStates == nil {
		run.RegionStates = initialRegionStates()
	}
	if run.ShopRefresh == nil {
		run.ShopRefresh = initialShopRefresh(max(1, run.Clock.Bell))
	}
	if run.ItemAffixes == nil {
		run.ItemAffixes = map[string][]AffixState{}
	}
	for _, q := range run.Quests {
		if q == nil {
			continue
		}
		if q.Stage == "" {
			if q.Status == "completed" {
				q.Stage = "return"
			} else {
				q.Stage = "active"
			}
		}
	}
	e.syncRegionStates(run)
	e.syncNPCLocations(run, false)
}

func (e *Engine) syncRegionStates(run *Run) {
	if run.RegionStates == nil {
		run.RegionStates = initialRegionStates()
	}
	if run.Flags["lost_patrol_reported"] {
		run.RegionStates["外墓区"] = "巡夜团重新建立哨线"
	} else if run.Clock.Bell >= 7 {
		run.RegionStates["外墓区"] = "灰雾加深 · 巡夜线失联"
	}
	if run.Flags["prisoner_freed"] {
		run.RegionStates["旧城区"] = "被抹去的名字开始回响"
	} else if run.Clock.Bell >= 7 {
		run.RegionStates["旧城区"] = "残响暴增"
	}
	if run.Flags["defeated_crown_bearer"] {
		run.RegionStates["烬冠核心"] = "王冠断裂 · 锁链失衡"
	}
	if run.Flags["act2_unlocked"] {
		run.RegionStates["雾外荒原"] = "灰风高危"
	}
	if run.Flags["mara_hunt_reported"] {
		run.RegionStates["雾外荒原"] = "逐风者掌握风暴巡逻路线"
	}
	if run.Flags["opened_outer_gate"] {
		run.RegionStates["荒原东界"] = "外封印破口"
	}
	if run.Flags["defeated_gate_heart"] {
		run.RegionStates["荒原东界"] = "门后心跳停止"
	}
}

func desiredNPCLocation(run *Run, npcID string) string {
	// Plot consequences override ordinary daily schedules.
	switch npcID {
	case "iven":
		if run.Flags["lost_patrol_reported"] {
			return "room_01"
		}
	case "nameless_prisoner":
		if run.Flags["prisoner_freed"] {
			return "room_13"
		}
	case "orrin":
		if run.Flags["orrin_oath"] || run.Flags["opened_outer_gate"] {
			return "room_43"
		}
	case "mara":
		if run.Flags["mara_hunt_reported"] {
			return "room_33"
		}
	}

	// V0.7 schedules are actual movement data, not just descriptive metadata.
	if schedule, ok := NPCSchedules[npcID]; ok {
		bell := max(1, run.Clock.Bell)
		to := ""
		for _, row := range schedule.Entries {
			if bell < row.FromBell || bell > row.ToBell {
				continue
			}
			if row.RequireFlag != "" && !run.Flags[row.RequireFlag] {
				continue
			}
			if run.Rooms[row.RoomID] != nil {
				to = row.RoomID // later matching rows can specialize an earlier fallback row
			}
		}
		if to != "" {
			return to
		}
	}

	return initialNPCLocations()[npcID]
}

func (e *Engine) syncNPCLocations(run *Run, logMoves bool) {
	if run.NPCLocations == nil {
		run.NPCLocations = initialNPCLocations()
	}
	ids := make([]string, 0, len(NPCs))
	for id := range NPCs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		to := desiredNPCLocation(run, id)
		if manual := run.NPCOverrides[id]; manual != "" && run.Rooms[manual] != nil {
			to = manual
		}
		if to == "" {
			continue
		}
		from := run.NPCLocations[id]
		if from == to {
			continue
		}
		run.NPCLocations[id] = to
		if logMoves && from != "" {
			npc := NPCs[id]
			room := run.Rooms[to]
			if room != nil {
				e.log(run, "world", fmt.Sprintf("%s离开原来的位置，前往「%s」。", npc.Name, room.Name))
			}
		}
	}
}

func (e *Engine) dialogueStartNode(run *Run, npcID, fallback string) string {
	switch npcID {
	case "iven":
		if run.Flags["lost_patrol_reported"] || run.Flags["lost_patrol_sealed"] {
			return "after"
		}
		if q := run.Quests["lost_patrol"]; q != nil && q.Status == "completed" {
			return "return"
		}
	case "nameless_prisoner":
		if run.Flags["prisoner_freed"] || run.Flags["prisoner_name_bound"] {
			return "after"
		}
		if q := run.Quests["nameless_prisoner"]; q != nil && q.Status == "completed" {
			return "return"
		}
	case "mara":
		if run.Flags["mara_hunt_reported"] || run.Flags["mara_route_withheld"] {
			return "after"
		}
		if q := run.Quests["storm_hunt"]; q != nil && q.Status == "completed" {
			return "return"
		}
	}
	return fallback
}

func (e *Engine) onBellChanged(run *Run) {
	e.ensureV06State(run)
	e.refreshShops(run)
	e.syncRegionStates(run)
	e.syncNPCLocations(run, true)
}

func (e *Engine) refreshShops(run *Run) {
	bell := max(1, run.Clock.Bell)
	for shopID, shop := range Shops {
		last := run.ShopRefresh[shopID]
		if last == 0 {
			last = 1
		}
		if bell-last < 3 {
			continue
		}
		changed := 0
		for _, row := range shop.Items {
			key := shopID + ":" + row.ItemID
			cur := run.ShopStock[key]
			cap := row.Stock
			if cur < cap {
				step := 1
				if item := Items[row.ItemID]; item.Type == "consumable" {
					step = 2
				}
				next := min(cap, cur+step)
				changed += next - cur
				run.ShopStock[key] = next
			}
		}
		run.ShopRefresh[shopID] = bell
		if changed > 0 {
			e.log(run, "shop", fmt.Sprintf("第 %d 钟后，%s补充了 %d 件货物。", bell, shop.Name, changed))
		}
	}
}

func (e *Engine) affixForItem(run *Run, itemID, salt string) []AffixState {
	item, ok := Items[itemID]
	if !ok || item.Slot == "" {
		return nil
	}
	candidates := make([]AffixDef, 0, len(Affixes))
	for _, a := range Affixes {
		if item.Slot != "weapon" && a.Power > 0 {
			continue
		}
		candidates = append(candidates, a)
	}
	if len(candidates) == 0 {
		return nil
	}
	idx := int(hash64(fmt.Sprintf("affix:%d:%s:%s", run.Seed, itemID, salt)) % uint64(len(candidates)))
	a := candidates[idx]
	return []AffixState{{ID: a.ID, Name: a.Name, Description: a.Description, Power: a.Power, Defense: a.Defense, MaxHP: a.MaxHP, MaxEnergy: a.MaxEnergy, Attributes: a.Attributes}}
}

func (e *Engine) ensureItemAffix(run *Run, itemID, salt string, force bool) []AffixState {
	e.ensureV06State(run)
	if existing, ok := run.ItemAffixes[itemID]; ok && len(existing) > 0 {
		return existing
	}
	item, ok := Items[itemID]
	if !ok || item.Slot == "" {
		return nil
	}
	if !force {
		chance := 35
		if item.Rarity == "rare" || item.Rarity == "legendary" {
			chance = 60
		}
		if int(hash64(fmt.Sprintf("affix-chance:%d:%s:%s", run.Seed, itemID, salt))%100) >= chance {
			return nil
		}
	}
	aff := e.affixForItem(run, itemID, salt)
	if len(aff) > 0 {
		run.ItemAffixes[itemID] = aff
	}
	return aff
}

func (e *Engine) grantAffixedLoot(run *Run, itemID, source string, force bool) {
	run.Player.Inventory = append(run.Player.Inventory, itemID)
	aff := e.ensureItemAffix(run, itemID, fmt.Sprintf("%s:%d", source, run.Turn), force)
	item := Items[itemID]
	if len(aff) > 0 {
		e.log(run, "loot", fmt.Sprintf("获得战利品：「%s · %s」。", aff[0].Name, item.Name))
	} else {
		e.log(run, "loot", "获得战利品：「"+item.Name+"」。")
	}
}

func (e *Engine) affixPower(run *Run, itemID string) int {
	power := 0
	for _, a := range run.ItemAffixes[itemID] {
		power += a.Power
	}
	return power
}

func (e *Engine) applyAffixBonuses(run *Run, itemID string, sign int) {
	if run == nil || itemID == "" {
		return
	}
	for _, a := range run.ItemAffixes[itemID] {
		run.Player.Defense += a.Defense * sign
		run.Player.MaxHP += a.MaxHP * sign
		run.Player.MaxEnergy += a.MaxEnergy * sign
		run.Player.Attributes.Strength += a.Attributes.Strength * sign
		run.Player.Attributes.Dexterity += a.Attributes.Dexterity * sign
		run.Player.Attributes.Perception += a.Attributes.Perception * sign
		run.Player.Attributes.Will += a.Attributes.Will * sign
	}
	run.Player.MaxHP = max(1, run.Player.MaxHP)
	run.Player.MaxEnergy = max(1, run.Player.MaxEnergy)
	run.Player.HP = clamp(run.Player.HP, 0, run.Player.MaxHP)
	run.Player.Energy = clamp(run.Player.Energy, 0, run.Player.MaxEnergy)
}

func startingDistance(classID string) int {
	switch classID {
	case "warden":
		return 1
	case "ranger":
		return 2
	default:
		return 3
	}
}

func distanceName(distance int) string {
	switch distance {
	case 1:
		return "近距"
	case 2:
		return "中距"
	default:
		return "远距"
	}
}

func (e *Engine) setQuestOutcome(run *Run, questID, outcome string) {
	if q := run.Quests[questID]; q != nil {
		q.Status = "completed"
		q.Stage = "resolved"
		q.Outcome = outcome
		q.Progress = q.Goal
	}
}

// Prepare normalizes older saves before they are exposed through the HTTP API.
func (e *Engine) Prepare(run *Run) {
	if run == nil {
		return
	}
	e.ensureV07State(run)
	if run.Combat != nil && run.Combat.Distance <= 0 {
		run.Combat.Distance = startingDistance(run.Player.Class)
	}
}

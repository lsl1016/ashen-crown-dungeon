package game

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

func startingSkillsForClass(classID string) []string {
	switch classID {
	case "warden":
		return []string{"warden_smite", "warden_chainbreaker"}
	case "ranger":
		return []string{"ranger_pierce", "ranger_smoke_arrow"}
	default:
		return []string{"seer_burst", "seer_warding_flare"}
	}
}

func signatureSkillForClass(classID string) string {
	s := startingSkillsForClass(classID)
	if len(s) == 0 {
		return ""
	}
	return s[0]
}

func (e *Engine) ensureV07State(run *Run) {
	e.ensureV06State(run)
	if len(run.Player.Skills) == 0 {
		run.Player.Skills = startingSkillsForClass(run.Player.Class)
	}
	if run.WorldEvents == nil {
		run.WorldEvents = map[string]*WorldEventState{}
	}
	if run.NPCWorld == nil {
		run.NPCWorld = map[string]*NPCWorldState{}
	}
	if run.SceneStates == nil {
		run.SceneStates = map[string]*SceneState{}
	}
	if run.QuestDecisions == nil {
		run.QuestDecisions = map[string]string{}
	}
	if run.NPCOverrides == nil {
		run.NPCOverrides = map[string]string{}
	}
	if run.RegionOverrides == nil {
		run.RegionOverrides = map[string]string{}
	}
	e.applyLivingWorldOverrides(run)
	e.syncWorldEvents(run, false)
	e.syncLivingWorldRegions(run)
	e.applyLivingWorldOverrides(run)
	e.syncNPCWorld(run)
	e.syncSceneStates(run)
	if run.Combat != nil {
		if run.Combat.Distance <= 0 {
			run.Combat.Distance = startingDistance(run.Player.Class)
		}
		if run.Combat.Terrain == "" {
			e.configureCombatTerrain(run, run.Combat)
		}
	}
}

func (e *Engine) onBellChangedV07(run *Run) {
	if run.WorldEvents == nil {
		run.WorldEvents = map[string]*WorldEventState{}
	}
	e.syncWorldEvents(run, true)
	e.syncLivingWorldRegions(run)
	e.applyLivingWorldOverrides(run)
	e.syncNPCWorld(run)
	e.syncSceneStates(run)
}

func (e *Engine) applyLivingWorldOverrides(run *Run) {
	for npcID, roomID := range run.NPCOverrides {
		if _, ok := NPCs[npcID]; ok && run.Rooms[roomID] != nil {
			run.NPCLocations[npcID] = roomID
		}
	}
	for region, value := range run.RegionOverrides {
		if strings.TrimSpace(value) != "" {
			run.RegionStates[region] = value
		}
	}
}

func worldEventResolved(run *Run, def WorldEventDef) bool {
	if def.ResolveFlag != "" && run.Flags[def.ResolveFlag] {
		return true
	}
	switch def.ID {
	case "grey_plague_tide":
		return run.Flags["lost_patrol_sealed"]
	case "nameless_procession":
		return run.Flags["prisoner_name_bound"]
	case "ashstorm_front":
		return run.Flags["mara_route_withheld"]
	}
	return false
}

func (e *Engine) syncWorldEvents(run *Run, notify bool) {
	if run.WorldEvents == nil {
		run.WorldEvents = map[string]*WorldEventState{}
	}
	ids := make([]string, 0, len(LivingWorldEvents))
	for id := range LivingWorldEvents {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	bell := max(1, run.Clock.Bell)
	for _, id := range ids {
		def := LivingWorldEvents[id]
		if def.RequireFlag != "" && !run.Flags[def.RequireFlag] {
			continue
		}
		st := run.WorldEvents[id]
		resolved := worldEventResolved(run, def)
		if st == nil && bell >= def.TriggerBell {
			st = &WorldEventState{ID: id, Title: def.Title, Region: def.Region, Stage: "emerging", Status: "active", StartedBell: bell, UpdatedBell: bell, Severity: def.Severity, Description: def.Description}
			run.WorldEvents[id] = st
			if notify {
				e.log(run, "world_event", fmt.Sprintf("世界事件开始：%s。%s", def.Title, def.Description))
			}
		}
		if st == nil {
			continue
		}
		if resolved {
			if st.Status != "resolved" {
				st.Status = "resolved"
				st.Stage = "resolved"
				st.UpdatedBell = bell
				st.Description = worldEventResolutionText(run, def)
				if notify {
					e.log(run, "world_event", fmt.Sprintf("世界事件结束：%s。%s", def.Title, st.Description))
				}
			}
			continue
		}
		if st.Status == "resolved" {
			continue
		}
		if def.EscalateBell > 0 && bell >= def.EscalateBell && st.Stage != "escalated" {
			st.Stage = "escalated"
			st.Severity = min(4, def.Severity+1)
			st.UpdatedBell = bell
			st.Description = def.Description + " 事件已经恶化，并开始改变附近场景与 NPC 行动。"
			if notify {
				e.log(run, "world_event", fmt.Sprintf("世界事件升级：%s（严重度 %d）。", def.Title, st.Severity))
			}
		}
	}
}

func worldEventResolutionText(run *Run, def WorldEventDef) string {
	switch def.ID {
	case "grey_plague_tide":
		if run.Flags["lost_patrol_sealed"] {
			return "巡夜团暂时压住灰疫，却把赛勒的记录一并封存；哨线稳定，但真相没有被记住。"
		}
		return "第三巡夜队的真相公开后，新的巡夜灯沿外墓区重新点亮。"
	case "nameless_procession":
		if run.Flags["prisoner_name_bound"] {
			return "真名重新封进锁链，无名游行停止，但旧城区再次失去一个本该被记住的人。"
		}
		return "囚徒重新拥有姓名后，游行者开始一个接一个想起自己的名字。"
	case "ashstorm_front":
		if run.Flags["mara_route_withheld"] {
			return "商路勉强避开灰暴，但逐风者没有拿到完整路线。"
		}
		return "逐风者掌握空钟骑士的巡逻路线，灰暴锋线被拆成数条可穿越的风道。"
	case "crown_name_hunt":
		return "赫里昂倒下后，王庭残响失去寻找活人姓名的中心。"
	}
	return "事件的影响正在消退。"
}

func (e *Engine) syncLivingWorldRegions(run *Run) {
	if run.RegionStates == nil {
		run.RegionStates = initialRegionStates()
	}
	// Explicit branch outcomes override the broad V0.6 region summary.
	if run.Flags["lost_patrol_sealed"] {
		run.RegionStates["外墓区"] = "哨线维持 · 真相被封存"
	}
	if run.Flags["prisoner_name_bound"] {
		run.RegionStates["旧城区"] = "残响沉寂 · 真名重新被封"
	}
	if run.Flags["mara_route_withheld"] {
		run.RegionStates["雾外荒原"] = "商路开放 · 逐风者仍缺完整路线"
	}
	for _, st := range run.WorldEvents {
		if st == nil || st.Status != "active" {
			continue
		}
		suffix := "事件活跃"
		if st.Stage == "escalated" {
			suffix = "事件恶化"
		}
		run.RegionStates[st.Region] = fmt.Sprintf("%s · %s", st.Title, suffix)
	}
}

func scheduleActivity(run *Run, npcID, location string) string {
	def, ok := NPCSchedules[npcID]
	if !ok {
		return "在墓城中等待局势变化"
	}
	bell := max(1, run.Clock.Bell)
	for _, row := range def.Entries {
		if bell < row.FromBell || bell > row.ToBell {
			continue
		}
		if row.RequireFlag != "" && !run.Flags[row.RequireFlag] {
			continue
		}
		if row.RoomID == location {
			return row.Activity
		}
	}
	// Dynamic flag-driven moves are still represented as activities.
	switch npcID {
	case "iven":
		if run.Flags["lost_patrol_reported"] {
			return "在墓城入口重建第三巡夜队的值守名册"
		}
	case "veil_broker":
		if location == "room_33" {
			return "在雾外界碑搭建临时黑市"
		}
	case "nameless_prisoner":
		if run.Flags["prisoner_freed"] {
			return "第一次以自己的姓名帮助巡夜者辨认旧城档案"
		}
	case "mara":
		if location == "room_33" {
			return "把安全风道标记在界碑背面"
		}
	case "orrin":
		if location == "room_43" {
			return "守在外封印黑门前兑现最后的骑士誓言"
		}
	}
	return "观察附近的世界异动"
}

func (e *Engine) syncNPCWorld(run *Run) {
	if run.NPCWorld == nil {
		run.NPCWorld = map[string]*NPCWorldState{}
	}
	for id, npc := range NPCs {
		loc := run.NPCLocations[id]
		rel := run.NPCRelations[id]
		mood := "谨慎"
		if rel >= 5 {
			mood = "信任"
		} else if rel >= 2 {
			mood = "友善"
		} else if rel <= -2 {
			mood = "戒备"
		}
		health := "稳定"
		if st := run.WorldEvents["grey_plague_tide"]; st != nil && st.Status == "active" && id == "iven" {
			health = "疲惫"
		}
		knowledge := []string{npc.Faction}
		if run.Flags["knows_crown_truth"] {
			knowledge = append(knowledge, "王冠真相正在扩散")
		}
		run.NPCWorld[id] = &NPCWorldState{Location: loc, Activity: scheduleActivity(run, id, loc), Health: health, Mood: mood, Knowledge: knowledge}
	}
}

func (e *Engine) syncSceneStates(run *Run) {
	if run.SceneStates == nil {
		run.SceneStates = map[string]*SceneState{}
	}
	activeByRegion := map[string]*WorldEventState{}
	for _, ev := range run.WorldEvents {
		if ev != nil && ev.Status == "active" {
			if cur := activeByRegion[ev.Region]; cur == nil || ev.Severity > cur.Severity {
				activeByRegion[ev.Region] = ev
			}
		}
	}
	for id, room := range run.Rooms {
		st := &SceneState{Kind: "stable", Label: "暂时稳定", Description: run.RegionStates[room.Zone]}
		if ev := activeByRegion[room.Zone]; ev != nil {
			def := LivingWorldEvents[ev.ID]
			st = &SceneState{Kind: def.Overlay, Label: ev.Title, Description: ev.Description, Overlay: def.Overlay, Severity: ev.Severity}
		} else if strings.Contains(run.RegionStates[room.Zone], "重建") || strings.Contains(run.RegionStates[room.Zone], "掌握") {
			st = &SceneState{Kind: "secured", Label: "区域已稳定", Description: run.RegionStates[room.Zone], Overlay: "safe"}
		} else if strings.Contains(run.RegionStates[room.Zone], "破口") || strings.Contains(run.RegionStates[room.Zone], "高危") {
			st = &SceneState{Kind: "danger", Label: "高危区域", Description: run.RegionStates[room.Zone], Overlay: "danger", Severity: 2}
		}
		run.SceneStates[id] = st
	}
}

func (e *Engine) configureCombatTerrain(run *Run, combat *CombatState) {
	if combat == nil {
		return
	}
	room := run.Rooms[run.CurrentRoomID]
	if room == nil {
		return
	}
	combat.Hazards = nil
	switch room.Type {
	case "bridge", "gatehouse":
		combat.Terrain = "broken_cover"
		combat.TerrainHint = "断柱与门垛：处于远距时防御 +2。"
	case "flooded", "aqueduct":
		combat.Terrain = "blackwater"
		combat.TerrainHint = "黑水带位于近距；停留在那里会在回合末受到 2 点侵蚀伤害。"
		combat.Hazards = []CombatHazard{{ID: "blackwater", Label: "黑水带", Description: "腐蚀性黑水漫过脚踝。", Distance: 1, Damage: 2}}
	case "forge":
		combat.Terrain = "cold_forge"
		combat.TerrainHint = "冷炉裂口位于中距；停留会受到 3 点灼灰伤害。"
		combat.Hazards = []CombatHazard{{ID: "forge裂口", Label: "冷炉裂口", Description: "炉底仍有灰白热流。", Distance: 2, Damage: 3}}
	case "ashfield", "ruins", "crater", "meteor":
		combat.Terrain = "ashstorm"
		combat.TerrainHint = "灰暴切过中距；中距回合末受到 2 点风蚀伤害。"
		combat.Hazards = []CombatHazard{{ID: "ashstorm", Label: "灰暴带", Description: "横向灰暴遮蔽战场。", Distance: 2, Damage: 2}}
	default:
		combat.Terrain = "open"
		combat.TerrainHint = "开阔战场，没有额外地形修正。"
	}
	if st := run.SceneStates[run.CurrentRoomID]; st != nil {
		switch st.Kind {
		case "plague":
			combat.Hazards = append(combat.Hazards, CombatHazard{ID: "plague_spores", Label: "灰疫孢雾", Description: "孢雾聚集在远距。", Distance: 3, Damage: 2})
			combat.TerrainHint += " 灰疫孢雾额外占据远距。"
		case "echo":
			combat.Hazards = append(combat.Hazards, CombatHazard{ID: "name_echo", Label: "失名残响", Description: "重复姓名的耳语聚集在中距。", Distance: 2, Damage: 1})
			combat.TerrainHint += " 无名游行的残响占据中距。"
		case "blackfire":
			combat.Hazards = append(combat.Hazards, CombatHazard{ID: "blackfire", Label: "王庭黑火", Description: "黑火沿近距地面追逐活人的名字。", Distance: 1, Damage: 3})
			combat.TerrainHint += " 王庭黑火封锁近距。"
		case "storm":
			combat.Hazards = append(combat.Hazards, CombatHazard{ID: "storm_front", Label: "灰暴锋线", Description: "锋线扫过远距，砂砾像刀片一样切割。", Distance: 3, Damage: 2})
			combat.TerrainHint += " 灰暴锋线额外扫过远距。"
		}
	}
}

func (e *Engine) applyCombatHazards(run *Run, messages *[]string) {
	if run.Combat == nil || len(run.Combat.Hazards) == 0 {
		return
	}
	for _, h := range run.Combat.Hazards {
		if h.Distance != run.Combat.Distance || h.Damage <= 0 {
			continue
		}
		dmg := h.Damage
		if run.Combat.Guarded {
			dmg = max(1, dmg-1)
		}
		run.Player.HP -= dmg
		*messages = append(*messages, fmt.Sprintf("地形「%s」造成 %d 点伤害。", h.Label, dmg))
		if run.Player.HP <= 0 {
			run.Player.HP = 0
			run.GameOver = true
			*messages = append(*messages, "你倒在危险地形中，灰雾覆盖了最后的脚印。")
			return
		}
	}
}

func (e *Engine) useDataSkill(run *Run, enemy EnemyDef, skillID string, enemyDefense, weaken int, messages *[]string) (bool, error) {
	combat := run.Combat
	if combat == nil {
		return false, errors.New("当前不在战斗中")
	}
	skill, ok := Skills[skillID]
	if !ok || skill.Class != run.Player.Class {
		return false, errors.New("当前职业无法使用这个技能")
	}
	owned := false
	for _, id := range run.Player.Skills {
		if id == skillID {
			owned = true
			break
		}
	}
	if !owned {
		return false, errors.New("尚未掌握这个技能")
	}
	if combat.Distance < max(1, skill.MinDistance) || (skill.MaxDistance > 0 && combat.Distance > skill.MaxDistance) {
		return false, fmt.Errorf("%s无法在%s使用", skill.Name, distanceName(combat.Distance))
	}
	cdKey := "skill:" + skillID
	if combat.Cooldowns[cdKey] > 0 {
		return false, fmt.Errorf("%s还需要 %d 回合冷却", skill.Name, combat.Cooldowns[cdKey])
	}
	cost := skill.Cost
	if skillID == signatureSkillForClass(run.Player.Class) && e.growthRank(run, "signature_mastery") >= 3 {
		cost = max(1, cost-1)
	}
	if run.Player.Energy < cost {
		return false, errors.New("能量不足")
	}
	run.Player.Energy -= cost
	combat.Cooldowns[cdKey] = skill.Cooldown
	if skillID == signatureSkillForClass(run.Player.Class) {
		combat.Cooldowns["skill"] = skill.Cooldown // old saves / old UI compatibility.
	}
	bonus := e.attribute(run.Player.Attributes, skill.HitAttribute) + skill.HitBonus - weaken*2
	if skillID == "ranger_pierce" {
		bonus += run.Player.Attributes.Perception / 2
	}
	mastery := 0
	if skillID == signatureSkillForClass(run.Player.Class) {
		mastery = e.growthRank(run, "signature_mastery") * 2
	}
	targetDefense := enemyDefense + skill.DefenseModifier
	roll := e.roll(run, "skill:"+skillID, 20)
	critAt := 20
	if e.hasTalent(run, "ranger_predator") {
		critAt = 19
	}
	critical := roll >= critAt
	total := roll + bonus
	success := roll != 1 && (critical || total >= targetDefense)
	run.LastRoll = &RollResult{Kind: "skill", Label: skill.Name, Roll: roll, Bonus: bonus, Total: total, DC: targetDefense, Success: success, Critical: critical}
	interrupted := false
	if !success {
		*messages = append(*messages, skill.Name+"被敌人躲开。")
		return false, nil
	}
	for _, ef := range skill.Effects {
		switch ef.Type {
		case "damage":
			dmg := ef.Value + bonus + mastery
			if ef.Dice > 0 {
				dmg += e.roll(run, "skill-damage:"+skillID, ef.Dice)
			}
			if skillID == "warden_smite" && e.hasTalent(run, "warden_execution") {
				dmg += 5
				if combat.Intent.Kind == "heavy" || combat.Intent.Kind == "charge" {
					interrupted = true
				}
			}
			if skillID == "seer_burst" && combat.Distance == 3 && e.hasTalent(run, "seer_overchannel") {
				dmg += 4
			}
			if run.Player.Equipment["weapon"] == "scribe_wand" && strings.HasPrefix(skillID, "seer_") {
				dmg += 2
			}
			if critical {
				dmg += max(2, ef.Value)
			}
			combat.EnemyHP -= max(1, dmg)
			*messages = append(*messages, fmt.Sprintf("%s命中，造成 %d 点伤害。", skill.Name, max(1, dmg)))
		case "guard":
			combat.Guarded = true
		case "enemy_status":
			name, desc := combatStatusMeta(ef.Target)
			combat.EnemyStatuses = addStatus(combat.EnemyStatuses, StatusState{ID: ef.Target, Name: name, Description: desc, Rounds: max(1, ef.Rounds), Stacks: max(1, ef.Stacks)})
		case "set_distance":
			combat.Distance = clamp(ef.Value, 1, 3)
		case "retreat":
			combat.Distance = clamp(combat.Distance+max(1, ef.Value), 1, 3)
		case "shield":
			combat.Counters["seer_ward"] = max(1, ef.Value/5)
		case "energy":
			run.Player.Energy = clamp(run.Player.Energy+ef.Value, 0, run.Player.MaxEnergy)
		case "heal":
			run.Player.HP = clamp(run.Player.HP+ef.Value, 0, run.Player.MaxHP)
		}
	}
	if skillID == "ranger_pierce" && e.hasTalent(run, "ranger_bleed") {
		combat.EnemyStatuses = addStatus(combat.EnemyStatuses, StatusState{ID: "bleed", Name: "流血", Description: "每回合受到 2 点伤害。", Rounds: 2, Stacks: 1})
	}
	if skillID == "seer_burst" && e.hasTalent(run, "seer_cinder") {
		combat.EnemyStatuses = addStatus(combat.EnemyStatuses, StatusState{ID: "burn", Name: "燃烧", Description: "余烬持续灼烧。", Rounds: 3, Stacks: 1})
	}
	if skillID == "seer_burst" && e.hasTalent(run, "seer_siphon") {
		run.Player.HP = clamp(run.Player.HP+3, 0, run.Player.MaxHP)
		*messages = append(*messages, "残响汲取恢复 3 点生命。")
	}
	return interrupted, nil
}

func combatStatusMeta(id string) (string, string) {
	switch id {
	case "weakened":
		return "弱化", "攻击节奏被打乱，命中降低。"
	case "burn":
		return "燃烧", "每回合受到余烬伤害。"
	case "bleed":
		return "流血", "每回合受到伤害。"
	default:
		return id, id
	}
}

// GM helpers are deliberately explicit. The editor calls these instead of writing save JSON directly.
func (e *Engine) GMSetRegion(run *Run, region, value string) error {
	if strings.TrimSpace(region) == "" || strings.TrimSpace(value) == "" {
		return errors.New("区域与状态不能为空")
	}
	e.ensureV07State(run)
	if value == "@auto" {
		delete(run.RegionOverrides, region)
		e.ensureV07State(run)
		e.log(run, "gm", fmt.Sprintf("GM 将区域「%s」恢复为自动模拟。", region))
		return nil
	}
	run.RegionOverrides[region] = value
	run.RegionStates[region] = value
	e.syncSceneStates(run)
	e.log(run, "gm", fmt.Sprintf("GM 将区域「%s」状态固定为「%s」。", region, value))
	return nil
}

func (e *Engine) GMMoveNPC(run *Run, npcID, roomID string) error {
	if _, ok := NPCs[npcID]; !ok {
		return errors.New("NPC 不存在")
	}
	e.ensureV07State(run)
	if roomID == "@schedule" {
		delete(run.NPCOverrides, npcID)
		e.syncNPCLocations(run, false)
		e.syncNPCWorld(run)
		e.log(run, "gm", fmt.Sprintf("GM 将 %s 恢复为自动日程。", NPCs[npcID].Name))
		return nil
	}
	if run.Rooms[roomID] == nil {
		return errors.New("房间不存在")
	}
	run.NPCOverrides[npcID] = roomID
	run.NPCLocations[npcID] = roomID
	e.syncNPCWorld(run)
	e.log(run, "gm", fmt.Sprintf("GM 将 %s 固定移动到「%s」。", NPCs[npcID].Name, run.Rooms[roomID].Name))
	return nil
}

func (e *Engine) GMSetFlag(run *Run, flag string, value bool) error {
	if strings.TrimSpace(flag) == "" {
		return errors.New("Flag 不能为空")
	}
	e.ensureV07State(run)
	run.Flags[flag] = value
	e.ensureV07State(run)
	e.log(run, "gm", fmt.Sprintf("GM 设置 Flag %s=%t。", flag, value))
	return nil
}

func (e *Engine) GMTriggerWorldEvent(run *Run, eventID string) error {
	def, ok := LivingWorldEvents[eventID]
	if !ok {
		return errors.New("世界事件不存在")
	}
	e.ensureV07State(run)
	run.WorldEvents[eventID] = &WorldEventState{ID: def.ID, Title: def.Title, Region: def.Region, Stage: "emerging", Status: "active", StartedBell: max(1, run.Clock.Bell), UpdatedBell: max(1, run.Clock.Bell), Severity: def.Severity, Description: def.Description}
	e.syncLivingWorldRegions(run)
	e.syncSceneStates(run)
	e.log(run, "gm", "GM 强制触发世界事件："+def.Title+"。")
	return nil
}

func (e *Engine) GMAdvanceBell(run *Run) error {
	e.ensureV07State(run)
	if run.Clock.Bell >= 13 {
		return errors.New("第十三钟之后无法继续推进")
	}
	run.Clock.Bell++
	run.Clock.Steps = run.Clock.Bell * 5
	run.Clock.Threat = max(0, (run.Clock.Bell-1)/3)
	run.Clock.LastChange = fmt.Sprintf("GM 将世界时间推进到第 %d 钟。", run.Clock.Bell)
	e.log(run, "gm", run.Clock.LastChange)
	e.onBellChanged(run)
	e.onBellChangedV07(run)
	return nil
}

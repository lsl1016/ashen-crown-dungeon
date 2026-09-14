package game

import (
	"errors"
	"fmt"
	"strings"
)

func (e *Engine) hasTalent(run *Run, id string) bool {
	return run.Player.Talents != nil && run.Player.Talents[id] > 0
}

func (e *Engine) UnlockTalent(run *Run, talentID string) error {
	if run.GameOver {
		return errors.New("冒险已经结束")
	}
	if run.Combat != nil {
		return errors.New("战斗中不能学习天赋")
	}
	if run.Player.TalentPoints <= 0 {
		return errors.New("没有可用天赋点")
	}
	var def *TalentDef
	for i := range Talents {
		if Talents[i].ID == talentID {
			def = &Talents[i]
			break
		}
	}
	if def == nil {
		return errors.New("未知天赋")
	}
	if def.Class != run.Player.Class {
		return errors.New("该天赋不属于当前职业")
	}
	if run.Player.Talents == nil {
		run.Player.Talents = map[string]int{}
	}
	if def.RequiredLevel > 0 && run.Player.Level < def.RequiredLevel {
		return fmt.Errorf("需要达到 %d 级", def.RequiredLevel)
	}
	for _, req := range def.Requires {
		if run.Player.Talents[req] <= 0 {
			for _, td := range Talents {
				if td.ID == req {
					return fmt.Errorf("需要先学习「%s」", td.Name)
				}
			}
			return errors.New("缺少前置天赋")
		}
	}
	if run.Player.Talents[talentID] >= def.MaxRank {
		return errors.New("该天赋已经达到最高等级")
	}
	run.Player.Talents[talentID]++
	run.Player.TalentPoints--
	switch talentID {
	case "warden_vitality":
		run.Player.MaxHP += 8
		run.Player.HP = clamp(run.Player.HP+8, 0, run.Player.MaxHP)
	case "seer_reservoir":
		run.Player.MaxEnergy += 3
		run.Player.Energy = clamp(run.Player.Energy+3, 0, run.Player.MaxEnergy)
	}
	e.log(run, "talent", fmt.Sprintf("已学习天赋「%s」：%s", def.Name, def.Description))
	return nil
}

func (e *Engine) equipItem(run *Run, itemID string) error {
	item, ok := Items[itemID]
	if !ok || item.Slot == "" {
		return errors.New("该物品不能装备")
	}
	if !hasItem(run.Player.Inventory, itemID) {
		return errors.New("背包中没有该物品")
	}
	if run.Player.Equipment == nil {
		run.Player.Equipment = map[string]string{}
	}
	oldID := run.Player.Equipment[item.Slot]
	if oldID == itemID {
		return nil
	}
	if old, ok := Items[oldID]; ok {
		e.applyItemBonuses(&run.Player, old, -1)
		e.applyAffixBonuses(run, oldID, -1)
	}
	run.Player.Equipment[item.Slot] = itemID
	e.applyItemBonuses(&run.Player, item, 1)
	e.applyAffixBonuses(run, itemID, 1)
	e.log(run, "item", fmt.Sprintf("已装备%s：「%s」。", slotName(item.Slot), item.Name))
	return nil
}

func (e *Engine) applyItemBonuses(p *Player, item ItemDef, sign int) {
	p.Defense += item.Defense * sign
	p.MaxHP += item.MaxHP * sign
	p.MaxEnergy += item.MaxEnergy * sign
	p.Attributes.Strength += item.Attributes.Strength * sign
	p.Attributes.Dexterity += item.Attributes.Dexterity * sign
	p.Attributes.Perception += item.Attributes.Perception * sign
	p.Attributes.Will += item.Attributes.Will * sign
	if p.MaxHP < 1 {
		p.MaxHP = 1
	}
	if p.MaxEnergy < 1 {
		p.MaxEnergy = 1
	}
	p.HP = clamp(p.HP, 0, p.MaxHP)
	p.Energy = clamp(p.Energy, 0, p.MaxEnergy)
}

func slotName(slot string) string {
	switch slot {
	case "weapon":
		return "武器"
	case "armor":
		return "护甲"
	case "trinket":
		return "饰品"
	default:
		return "装备"
	}
}

func (e *Engine) advanceClock(run *Run) {
	if run.Clock.Bell <= 0 {
		run.Clock.Bell = 1
	}
	run.Clock.Steps++
	if run.Clock.Steps%5 != 0 || run.Clock.Bell >= 13 {
		return
	}
	run.Clock.Bell++
	run.Clock.Threat = max(0, (run.Clock.Bell-1)/3)
	switch run.Clock.Bell {
	case 4:
		run.Clock.LastChange = "第四声钟后，外墓区的道路发生第一次明显重排。"
		run.Flags["world_shift_1"] = true
	case 7:
		run.Clock.LastChange = "第七声钟后，王庭残响开始主动寻找仍活着的名字。"
		run.Flags["world_shift_2"] = true
	case 10:
		run.Clock.LastChange = "第十声钟后，王座方向传来黑火燃烧的声音。"
		run.Flags["king_stirring"] = true
	case 13:
		run.Clock.LastChange = "第十三声钟本应响起，但整座墓城突然陷入绝对寂静。"
		run.Flags["thirteenth_silence"] = true
	default:
		run.Clock.LastChange = fmt.Sprintf("第%d声钟穿过石壁，墓城威胁正在上升。", run.Clock.Bell)
	}
	e.log(run, "world", run.Clock.LastChange)
	e.onBellChanged(run)
	e.onBellChangedV07(run)
}

func statusStacks(statuses []StatusState, id string) int {
	for _, s := range statuses {
		if s.ID == id {
			return max(1, s.Stacks)
		}
	}
	return 0
}

func addStatus(statuses []StatusState, next StatusState) []StatusState {
	if next.Stacks <= 0 {
		next.Stacks = 1
	}
	for i := range statuses {
		if statuses[i].ID == next.ID {
			statuses[i].Rounds = max(statuses[i].Rounds, next.Rounds)
			statuses[i].Stacks = clamp(statuses[i].Stacks+next.Stacks, 1, 5)
			return statuses
		}
	}
	return append(statuses, next)
}

func (e *Engine) tickEnemyStatuses(run *Run, messages *[]string) {
	if run.Combat == nil {
		return
	}
	out := make([]StatusState, 0, len(run.Combat.EnemyStatuses))
	for _, st := range run.Combat.EnemyStatuses {
		dmg := 0
		switch st.ID {
		case "burn":
			dmg = 3 * max(1, st.Stacks)
			if run.Player.Equipment["trinket"] == "ember_shard" {
				dmg += max(1, st.Stacks)
			}
		case "bleed":
			dmg = 2 * max(1, st.Stacks)
		}
		if dmg > 0 {
			run.Combat.EnemyHP -= dmg
			*messages = append(*messages, fmt.Sprintf("%s造成 %d 点持续伤害。", st.Name, dmg))
		}
		st.Rounds--
		if st.Rounds > 0 {
			out = append(out, st)
		}
	}
	run.Combat.EnemyStatuses = out
}

func (e *Engine) tickPlayerStatuses(combat *CombatState) {
	out := make([]StatusState, 0, len(combat.PlayerStatuses))
	for _, st := range combat.PlayerStatuses {
		st.Rounds--
		if st.Rounds > 0 {
			out = append(out, st)
		}
	}
	combat.PlayerStatuses = out
}

func (e *Engine) tickCooldowns(combat *CombatState) {
	for k, v := range combat.Cooldowns {
		if v > 0 {
			combat.Cooldowns[k] = v - 1
		}
	}
}

func (e *Engine) updateBossPhase(run *Run, enemy EnemyDef, messages *[]string) {
	if !enemy.Boss || run.Combat == nil {
		return
	}
	combat := run.Combat
	ratio := float64(combat.EnemyHP) / float64(max(1, combat.EnemyMaxHP))
	phase := combat.BossPhase
	if ratio <= 0.30 {
		phase = 3
	} else if ratio <= 0.60 {
		phase = 2
	}
	if phase <= combat.BossPhase {
		return
	}
	combat.BossPhase = phase
	switch phase {
	case 2:
		combat.EnemyDefense++
		if enemy.ID == "gate_heart" {
			*messages = append(*messages, "【最终 Boss 二阶段】黑色心脏裂开一圈星光，整个心室开始随它收缩。防御提升。")
			e.log(run, "boss", "门后之心进入第二阶段：封印空间开始主动挤压入侵者。")
		} else {
			*messages = append(*messages, "【Boss 二阶段】王冠的外环裂开，赫里昂拔出第二道锁链，防御提升。")
			e.log(run, "boss", "赫里昂进入第二阶段：王冠开始直接操纵墓城残响。")
		}
	case 3:
		combat.EnemyDefense++
		name := "王名压迫"
		desc := "攻击检定 -2。"
		if enemy.ID == "gate_heart" {
			name = "门后凝视"
			desc = "现实边界正在变薄，攻击检定 -2。"
		}
		combat.PlayerStatuses = addStatus(combat.PlayerStatuses, StatusState{ID: "weakened", Name: name, Description: desc, Rounds: 2, Stacks: 1})
		if enemy.ID == "gate_heart" {
			*messages = append(*messages, "【最终 Boss 三阶段】心室失去上下方向，黑门完全睁开。")
			e.log(run, "boss", "门后之心进入第三阶段：现实与门后的空间开始重叠。")
		} else {
			*messages = append(*messages, "【Boss 三阶段】王座后的十三道锁链全部绷紧，黑火吞没整个大厅。")
			e.log(run, "boss", "赫里昂进入第三阶段：第十三道锁链显现。")
		}
	}
}

func (e *Engine) chooseEnemyIntent(run *Run, enemy EnemyDef, combat *CombatState) EnemyIntent {
	profile := enemy.IntentProfile
	if profile == "" {
		profile = "brute"
	}
	pools := map[string][]string{
		"brute":      {"attack", "heavy", "attack", "defend"},
		"skirmisher": {"attack", "multi", "hex", "attack"},
		"predator":   {"multi", "attack", "heavy", "attack"},
		"guardian":   {"defend", "attack", "heavy", "defend"},
		"trickster":  {"hex", "attack", "multi", "defend"},
		"hexer":      {"hex", "attack", "drain", "hex"},
		"duelist":    {"attack", "defend", "heavy", "multi"},
	}
	pool := pools[profile]
	if enemy.Boss {
		switch combat.BossPhase {
		case 1:
			pool = []string{"attack", "defend", "heavy", "attack"}
		case 2:
			pool = []string{"hex", "multi", "heavy", "drain"}
		default:
			pool = []string{"charge", "multi", "drain", "heavy"}
		}
	}
	if combat.Distance > 1 && (profile == "brute" || profile == "guardian" || profile == "duelist" || profile == "predator") && combat.Round%2 == 0 {
		return intentDef("advance", enemy)
	}
	if len(pool) == 0 {
		pool = []string{"attack"}
	}
	idx := int(hash64(fmt.Sprintf("intent:%d:%s:%d:%d:%d", run.Seed, enemy.ID, combat.Round+1, combat.BossPhase, run.Turn)) % uint64(len(pool)))
	return intentDef(pool[idx], enemy)
}

func intentDef(id string, enemy EnemyDef) EnemyIntent {
	switch id {
	case "advance":
		return EnemyIntent{ID: id, Kind: "advance", Label: "逼近", Description: "敌人试图缩短站位距离。", Icon: "→", Telegraph: "后撤可以继续保持距离，近战职业也可以迎上去"}
	case "heavy":
		return EnemyIntent{ID: id, Kind: "heavy", Label: "蓄力重击", Description: "下一击伤害更高，但动作明显。", Icon: "☄", Power: 4, Telegraph: "适合防御，铁誓守卫可尝试打断"}
	case "multi":
		return EnemyIntent{ID: id, Kind: "multi", Label: "连续突袭", Description: "连续发动两次较轻攻击。", Icon: "≋", Power: 2, Telegraph: "高防御比单次减伤更有效"}
	case "defend":
		return EnemyIntent{ID: id, Kind: "defend", Label: "防御架势", Description: "敌人准备强化下一轮防御。", Icon: "⛨", Telegraph: "可以趁机治疗、蓄能或使用道具"}
	case "hex":
		return EnemyIntent{ID: id, Kind: "hex", Label: "记忆侵蚀", Description: "削弱你的攻击检定并造成少量伤害。", Icon: "◈", Power: 2, Telegraph: "意图不会被普通防御完全抵消"}
	case "drain":
		return EnemyIntent{ID: id, Kind: "drain", Label: "残响汲取", Description: "造成伤害并回复自身生命。", Icon: "☽", Power: 2, Telegraph: "优先压低敌人生命可减少回复收益"}
	case "charge":
		return EnemyIntent{ID: id, Kind: "charge", Label: "王冠黑火", Description: "Boss 正在聚集一次高威胁攻击。", Icon: "♛", Power: 6, Telegraph: "强烈建议防御、打断或使用烟弹"}
	default:
		return EnemyIntent{ID: "attack", Kind: "attack", Label: "试探攻击", Description: "一次标准攻击。", Icon: "⚔", Telegraph: enemy.Weakness}
	}
}

func (e *Engine) resolveEnemyIntent(run *Run, enemy EnemyDef, messages *[]string) {
	combat := run.Combat
	if combat == nil {
		return
	}
	intent := combat.Intent
	switch intent.Kind {
	case "advance":
		if combat.Distance > 1 {
			combat.Distance--
		}
		*messages = append(*messages, fmt.Sprintf("%s向前逼近，距离变为%s。", enemy.Name, distanceName(combat.Distance)))
		return
	case "defend":
		combat.Counters["enemy_guard"] = 1
		*messages = append(*messages, enemy.Name+"进入防御架势，下一轮防御 +3。")
		return
	case "hex":
		dmg := e.enemyAttackOnce(run, enemy, -2, -2, messages, "记忆侵蚀")
		if dmg >= 0 {
			combat.PlayerStatuses = addStatus(combat.PlayerStatuses, StatusState{ID: "weakened", Name: "记忆模糊", Description: "攻击检定 -2。", Rounds: 2, Stacks: 1})
			*messages = append(*messages, "你的记忆出现短暂空白，攻击检定受到削弱。")
		}
	case "multi":
		e.enemyAttackOnce(run, enemy, 0, -2, messages, "第一击")
		if !run.GameOver {
			e.enemyAttackOnce(run, enemy, -1, -2, messages, "第二击")
		}
	case "heavy":
		e.enemyAttackOnce(run, enemy, -1, 4, messages, "蓄力重击")
	case "drain":
		dmg := e.enemyAttackOnce(run, enemy, 0, 2, messages, "残响汲取")
		if dmg > 0 {
			heal := max(2, dmg/2)
			combat.EnemyHP = clamp(combat.EnemyHP+heal, 0, combat.EnemyMaxHP)
			*messages = append(*messages, fmt.Sprintf("%s从你的记忆中恢复 %d 点生命。", enemy.Name, heal))
		}
	case "charge":
		combat.Distance = 1
		e.enemyAttackOnce(run, enemy, -2, 6, messages, "王冠黑火")
	default:
		e.enemyAttackOnce(run, enemy, 0, 0, messages, "攻击")
	}
}

func (e *Engine) enemyAttackOnce(run *Run, enemy EnemyDef, attackMod, damageMod int, messages *[]string, label string) int {
	combat := run.Combat
	if combat == nil {
		return -1
	}
	attackRoll := e.roll(run, "enemy-"+combat.Intent.Kind+"-"+label, 20) + enemy.Attack + attackMod + run.Clock.Threat/2 - statusStacks(combat.EnemyStatuses, "weakened")*2
	if combat.Distance == 3 && (enemy.IntentProfile == "brute" || enemy.IntentProfile == "guardian" || enemy.IntentProfile == "duelist" || enemy.IntentProfile == "predator") {
		attackRoll -= 2
	}
	defense := run.Player.Defense
	if combat.Guarded {
		defense += 4
		if e.hasTalent(run, "warden_anchor") && combat.Distance == 1 {
			defense++
		}
		if e.hasTalent(run, "warden_bulwark") {
			defense += 2
		}
	}
	if e.hasTalent(run, "ranger_quickstep") && combat.Round <= 2 {
		defense += 2
	}
	if attackRoll < defense {
		*messages = append(*messages, fmt.Sprintf("%s的%s落空。", enemy.Name, label))
		return 0
	}
	dmgRange := max(1, enemy.DamageMax-enemy.DamageMin+1)
	dmg := enemy.DamageMin + e.roll(run, "enemy-damage-"+label, dmgRange) - 1 + damageMod + run.Clock.Threat/2
	if combat.Distance == 3 && (enemy.IntentProfile == "brute" || enemy.IntentProfile == "guardian" || enemy.IntentProfile == "duelist" || enemy.IntentProfile == "predator") {
		dmg--
	}
	if combat.Guarded {
		dmg = (dmg + 1) / 2
	}
	if run.Flags["frost_salt_active"] {
		dmg = max(1, dmg-2)
		delete(run.Flags, "frost_salt_active")
	}
	if combat.Counters["seer_ward"] > 0 {
		dmg = max(0, dmg-5)
		combat.Counters["seer_ward"] = 0
		*messages = append(*messages, "符文护幕吸收了 5 点伤害。")
	}
	dmg = max(0, dmg)
	run.Player.HP -= dmg
	if combat.Guarded && e.hasTalent(run, "warden_retaliate") && dmg > 0 {
		combat.EnemyHP -= 3
		*messages = append(*messages, "铁誓反击造成 3 点伤害。")
	}
	*messages = append(*messages, fmt.Sprintf("%s的%s命中，造成 %d 点伤害。", enemy.Name, label, dmg))
	if run.Player.HP <= 0 {
		if run.Player.Equipment["trinket"] == "black_rose" && !run.Flags["black_rose_saved"] {
			run.Flags["black_rose_saved"] = true
			run.Player.HP = 1
			*messages = append(*messages, "灰玫瑰在胸前化为温热灰烬，你从致死伤害中保留了 1 点生命。")
		} else {
			run.Player.HP = 0
			run.GameOver = true
			*messages = append(*messages, "你倒在墓城里。雾重新覆盖了你的足迹。")
		}
	}
	return dmg
}

func (e *Engine) executeElementAction(run *Run, room *Room, el *SceneElement) error {
	switch el.Action {
	case "dialogue":
		return e.StartDialogue(run, el.Target)
	case "event":
		if _, ok := Events[el.Target]; !ok {
			return errors.New("事件不存在")
		}
		e.startEvent(run, el.Target, room)
	case "lore":
		entry, ok := Lore[el.Target]
		if !ok {
			return errors.New("记录不存在")
		}
		e.addLore(run, entry)
		e.log(run, "lore", "发现记录：「"+entry.Title+"」。")
	case "loot":
		item, ok := Items[el.Target]
		if !ok {
			return errors.New("物品不存在")
		}
		run.Player.Inventory = append(run.Player.Inventory, el.Target)
		e.log(run, "loot", "从「"+el.Label+"」获得物品：「"+item.Name+"」。")
	case "gold":
		run.Player.Gold += max(0, el.Value)
		e.log(run, "loot", fmt.Sprintf("你在「%s」里找到 %d 枚古金币。", el.Label, max(0, el.Value)))
	case "rest":
		heal := el.Value
		if heal <= 0 {
			heal = 12
		}
		run.Player.HP = clamp(run.Player.HP+heal, 0, run.Player.MaxHP)
		run.Player.Energy = run.Player.MaxEnergy
		e.log(run, "rest", "你在不灭余火旁休整，生命与能量得到恢复。")
	case "unlock":
		target := run.Rooms[el.Target]
		if target == nil {
			return errors.New("目标地点不存在")
		}
		target.Locked = false
		target.Discovered = true
		e.log(run, "world", "机关启动，一条隐藏道路通向「"+target.Name+"」。")
	case "combat":
		if _, ok := Enemies[el.Target]; !ok {
			return errors.New("敌人不存在")
		}
		e.startCombat(run, el.Target)
	case "flag":
		if strings.TrimSpace(el.Target) == "" {
			return errors.New("缺少世界状态标记")
		}
		run.Flags[el.Target] = true
		e.log(run, "world", "世界状态发生变化："+el.Label+"。")
	case "message":
		text := strings.TrimSpace(el.Target)
		if text == "" {
			text = el.Description
		}
		e.log(run, "explore", el.Label+"："+text)
	case "", "check":
		// 纯检定元素的效果已经在 Interact 中处理。
	default:
		return errors.New("该元素没有可执行交互")
	}
	return nil
}

// spendCombatItemTurn lets tactical consumables participate in the same turn
// economy as attacks and skills. Using a combat item is powerful, but it never
// grants a free action before the enemy reacts.
func (e *Engine) spendCombatItemTurn(run *Run, itemName string) {
	if run.Combat == nil || run.GameOver {
		return
	}
	combat := run.Combat
	enemy := Enemies[combat.EnemyID]
	combat.Round++
	messages := []string{fmt.Sprintf("你使用了%s。", itemName)}

	e.tickEnemyStatuses(run, &messages)
	if combat.EnemyHP <= 0 {
		e.winCombat(run, enemy)
		return
	}

	e.updateBossPhase(run, enemy, &messages)
	e.resolveEnemyIntent(run, enemy, &messages)
	if run.GameOver || run.Combat == nil {
		return
	}

	e.tickPlayerStatuses(combat)
	e.tickCooldowns(combat)
	combat.Intent = e.chooseEnemyIntent(run, enemy, combat)
	combat.LastMessage = strings.Join(messages, " ")
	combat.Guarded = false
	e.log(run, "combat", combat.LastMessage)
}

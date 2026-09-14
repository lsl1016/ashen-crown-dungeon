package game

import (
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"time"
)

type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

func (e *Engine) NewRun(name, classID string, seed int64) (*Run, error) {
	var class *ClassDef
	for i := range Classes {
		if Classes[i].ID == classID {
			class = &Classes[i]
			break
		}
	}
	if class == nil {
		return nil, errors.New("未知职业")
	}
	if strings.TrimSpace(name) == "" {
		name = "无名旅者"
	}
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rooms, edges := GenerateDungeon(seed)
	player := Player{
		Name: name, Class: class.ID, Level: 1, XP: 0,
		HP: class.HP, MaxHP: class.HP, Energy: class.Energy, MaxEnergy: class.Energy,
		Defense: class.Defense, Gold: 24, Attributes: class.Attributes,
		Inventory: []string{class.StarterItem, "healing_draught", "bandage"},
		Equipment: map[string]string{"weapon": class.StarterItem}, TalentPoints: 1, Talents: map[string]int{},
	}
	run := &Run{
		ID:   fmt.Sprintf("run_%x", hash64(fmt.Sprintf("%d-%s-%d", seed, name, time.Now().UnixNano()))),
		Seed: seed, CreatedAt: time.Now(), UpdatedAt: time.Now(), Turn: 1,
		Player: player, Rooms: rooms, Edges: edges, CurrentRoomID: "room_01",
		Flags: map[string]bool{}, Quests: map[string]*QuestState{}, Lore: []LoreEntry{}, Log: []LogEntry{},
		Clock: WorldClock{Bell: 1, Steps: 0, Threat: 0, LastChange: "第一声钟已经结束。墓城开始重新排列道路。"},
	}
	// 每个职业都带一件能展示 V0.3 装备 Build 的起始副装备。
	switch class.ID {
	case "warden":
		run.Player.Inventory = append(run.Player.Inventory, "chainmail")
		_ = e.equipItem(run, "chainmail")
	case "ranger":
		run.Player.Inventory = append(run.Player.Inventory, "night_lantern")
		_ = e.equipItem(run, "night_lantern")
	case "seer":
		run.Player.Inventory = append(run.Player.Inventory, "ashen_charm")
		_ = e.equipItem(run, "ashen_charm")
	}
	run.Quests["ashen_crown"] = &QuestState{ID: "ashen_crown", Title: "主线 · 灰烬王冠", Description: "深入烬冠王庭，决定赫里昂与王冠的命运。", Status: "active", Progress: 0, Goal: 1}
	e.revealNeighbors(run, "room_01")
	e.log(run, "story", "你进入了沉眠墓城。第一声钟鸣已经结束。场景中的物件、门、机关和道路现在都会真实改变世界状态。")
	e.log(run, "level", "你拥有 1 点初始天赋点。可以从角色面板选择第一项 Build 能力。")
	return run, nil
}

func (e *Engine) Move(run *Run, target string) error {
	if run.GameOver {
		return errors.New("本次冒险已经结束")
	}
	if run.Combat != nil {
		return errors.New("战斗中无法移动")
	}
	if run.ActiveEvent != nil {
		return errors.New("必须先处理当前事件")
	}
	room := run.Rooms[target]
	if room == nil {
		return errors.New("目标房间不存在")
	}
	if !e.adjacent(run, run.CurrentRoomID, target) {
		return errors.New("该房间与当前位置不相邻")
	}
	if room.Locked {
		return errors.New("通路被封锁")
	}
	run.CurrentRoomID = target
	room.Discovered = true
	firstVisit := !room.Visited
	room.Visited = true
	run.Turn++
	e.advanceClock(run)
	e.revealNeighbors(run, target)
	e.log(run, "move", fmt.Sprintf("进入「%s」。", room.Name))
	if firstVisit && !room.Resolved {
		e.triggerRoom(run, room)
	}
	run.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) ResolveChoice(run *Run, choiceID string) error {
	if run.ActiveEvent == nil {
		return errors.New("当前没有可处理事件")
	}
	var choice *ChoiceDef
	for i := range run.ActiveEvent.Choices {
		if run.ActiveEvent.Choices[i].ID == choiceID {
			choice = &run.ActiveEvent.Choices[i]
			break
		}
	}
	if choice == nil {
		return errors.New("未知选项")
	}
	room := run.Rooms[run.CurrentRoomID]
	eventID := run.ActiveEvent.ID
	resultText := ""
	run.LastRoll = nil
	if choice.Check != nil {
		roll := e.roll(run, "event:"+run.ActiveEvent.ID+":"+choice.ID, 20)
		bonus := e.attribute(run.Player.Attributes, choice.Check.Attribute)
		if choice.Check.Attribute == "perception" && e.hasTalent(run, "ranger_scavenger") {
			bonus++
		}
		total := roll + bonus
		success := total >= choice.Check.DC
		run.LastRoll = &RollResult{Kind: "check", Label: choice.Text, Roll: roll, Bonus: bonus, Total: total, DC: choice.Check.DC, Success: success, Critical: roll == 20}
		if success {
			if err := e.validateEffects(run, choice.OnSuccess); err != nil {
				return err
			}
			resultText = fmt.Sprintf("检定成功：D20 %d + %d = %d，对抗 DC %d。", roll, bonus, total, choice.Check.DC)
			e.applyEffects(run, choice.OnSuccess)
		} else {
			if err := e.validateEffects(run, choice.OnFailure); err != nil {
				return err
			}
			resultText = fmt.Sprintf("检定失败：D20 %d + %d = %d，对抗 DC %d。", roll, bonus, total, choice.Check.DC)
			e.applyEffects(run, choice.OnFailure)
		}
	} else {
		if err := e.validateEffects(run, choice.OnSuccess); err != nil {
			return err
		}
		e.applyEffects(run, choice.OnSuccess)
		resultText = "你做出了选择。"
	}
	e.log(run, "event", resultText)
	// 事件若引发战斗，立刻退出事件界面；战斗结果负责继续推进房间状态。
	if run.Combat != nil {
		run.ActiveEvent = nil
	} else if eventID == "merchant" && choiceID != "leave" {
		// 商人允许连续交易，只有主动离开时才退出交易界面。
		def := Events["merchant"]
		run.ActiveEvent = &ActiveEvent{ID: def.ID, Title: def.Title, Description: def.Description, Choices: append([]ChoiceDef(nil), def.Choices...)}
	} else {
		run.ActiveEvent = nil
		if room != nil {
			room.Resolved = true
		}
	}
	run.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) CombatAction(run *Run, action string) error {
	if run.Combat == nil {
		return errors.New("当前不在战斗中")
	}
	if run.GameOver {
		return errors.New("冒险已经结束")
	}
	combat := run.Combat
	enemy := Enemies[combat.EnemyID]
	if combat.Cooldowns == nil {
		combat.Cooldowns = map[string]int{}
	}
	if combat.Counters == nil {
		combat.Counters = map[string]int{}
	}
	combat.Round++
	combat.Guarded = false
	run.LastRoll = nil
	messages := []string{}
	interrupted := false

	// 上一轮敌人的防御姿态只影响本轮玩家行动。
	enemyDefense := combat.EnemyDefense
	if combat.Counters["enemy_guard"] > 0 {
		enemyDefense += 3
	}
	weaken := statusStacks(combat.PlayerStatuses, "weakened")

	switch action {
	case "attack":
		roll := e.roll(run, "player-attack", 20)
		bonus := run.Player.Attributes.Strength
		if run.Player.Class == "ranger" {
			bonus = run.Player.Attributes.Dexterity
		}
		bonus -= weaken * 2
		critAt := 20
		if e.hasTalent(run, "ranger_predator") {
			critAt = 19
		}
		total := roll + bonus
		critical := roll >= critAt
		success := roll != 1 && (critical || total >= enemyDefense)
		run.LastRoll = &RollResult{Kind: "attack", Label: "普通攻击", Roll: roll, Bonus: bonus, Total: total, DC: enemyDefense, Success: success, Critical: critical}
		if success {
			dmg := 2 + bonus + e.roll(run, "player-damage", 6) + e.weaponPower(run)
			if critical {
				dmg += 4 + e.weaponPower(run)
				messages = append(messages, "暴击！")
				if run.Player.Equipment["weapon"] == "glass_knife" {
					combat.EnemyStatuses = addStatus(combat.EnemyStatuses, StatusState{ID: "bleed", Name: "流血", Description: "每回合受到伤害。", Rounds: 2, Stacks: 1})
				}
			}
			if run.Player.Equipment["weapon"] == "royal_spear" && (enemy.IntentProfile == "guardian" || enemy.Archetype == "守卫") {
				dmg += 2
			}
			if run.Player.Equipment["weapon"] == "sunken_mace" && (combat.Intent.Kind == "heavy" || combat.Intent.Kind == "charge") {
				dmg += 3
			}
			if run.Player.Equipment["weapon"] == "starsteel_blade" && (combat.Intent.Kind == "heavy" || combat.Intent.Kind == "charge" || enemy.Archetype == "构装") {
				dmg += 3
			}
			combat.EnemyHP -= max(1, dmg)
			messages = append(messages, fmt.Sprintf("你命中%s，造成 %d 点伤害。", combat.EnemyName, max(1, dmg)))
		} else {
			messages = append(messages, "你的攻击擦过敌人的防御。")
		}
	case "skill":
		if combat.Cooldowns["skill"] > 0 {
			return fmt.Errorf("职业技还需要 %d 回合冷却", combat.Cooldowns["skill"])
		}
		if run.Player.Energy < 3 {
			return errors.New("能量不足")
		}
		run.Player.Energy -= 3
		combat.Cooldowns["skill"] = 2
		roll := e.roll(run, "skill-attack", 20)
		bonus := run.Player.Attributes.Will + run.Player.Attributes.Perception/2 - weaken*2
		targetDefense := enemyDefense - 1
		baseDamage := 7
		skillName := "职业技"
		switch run.Player.Class {
		case "warden":
			bonus = run.Player.Attributes.Strength + 2 - weaken*2
			targetDefense = enemyDefense
			baseDamage = 6
			combat.Guarded = true
			skillName = "铁誓猛击"
			if e.hasTalent(run, "warden_execution") {
				baseDamage += 5
				if combat.Intent.Kind == "heavy" || combat.Intent.Kind == "charge" {
					interrupted = true
				}
			}
		case "ranger":
			bonus = run.Player.Attributes.Dexterity + run.Player.Attributes.Perception/2 - weaken*2
			targetDefense = enemyDefense - 3
			baseDamage = 7
			skillName = "弱点穿刺"
		case "seer":
			bonus = run.Player.Attributes.Will + 2 - weaken*2
			targetDefense = enemyDefense - 2
			baseDamage = 9
			if run.Player.Equipment["weapon"] == "scribe_wand" {
				baseDamage += 2
			}
			skillName = "余烬爆裂"
		}
		critAt := 20
		if e.hasTalent(run, "ranger_predator") {
			critAt = 19
		}
		total := roll + bonus
		critical := roll >= critAt
		success := roll != 1 && (critical || total >= targetDefense)
		run.LastRoll = &RollResult{Kind: "skill", Label: skillName, Roll: roll, Bonus: bonus, Total: total, DC: targetDefense, Success: success, Critical: critical}
		if success {
			dmg := baseDamage + bonus + e.roll(run, "skill-damage", 8)
			if critical {
				dmg += baseDamage
				messages = append(messages, "技能暴击！")
			}
			combat.EnemyHP -= max(1, dmg)
			messages = append(messages, fmt.Sprintf("%s命中，造成 %d 点伤害。", skillName, max(1, dmg)))
			if run.Player.Class == "ranger" && e.hasTalent(run, "ranger_bleed") {
				combat.EnemyStatuses = addStatus(combat.EnemyStatuses, StatusState{ID: "bleed", Name: "流血", Description: "每回合受到 2 点伤害。", Rounds: 2, Stacks: 1})
				messages = append(messages, "敌人进入流血状态。")
			}
			if run.Player.Class == "seer" && e.hasTalent(run, "seer_cinder") {
				combat.EnemyStatuses = addStatus(combat.EnemyStatuses, StatusState{ID: "burn", Name: "燃烧", Description: "余烬持续灼烧。", Rounds: 3, Stacks: 1})
				messages = append(messages, "余烬附着在敌人身上持续燃烧。")
			}
			if run.Player.Class == "seer" && e.hasTalent(run, "seer_siphon") {
				run.Player.HP = clamp(run.Player.HP+3, 0, run.Player.MaxHP)
				messages = append(messages, "你从残响中汲取力量，恢复 3 点生命。")
			}
		} else {
			messages = append(messages, skillName+"被敌人躲开。")
		}
	case "guard":
		combat.Guarded = true
		run.Player.Energy = clamp(run.Player.Energy+1, 0, run.Player.MaxEnergy)
		messages = append(messages, "你收紧架势，本轮防御提升并恢复 1 点能量。")
	case "potion":
		if !removeItem(&run.Player, "healing_draught") {
			return errors.New("背包里没有红叶药剂")
		}
		heal := Items["healing_draught"].Power
		run.Player.HP = clamp(run.Player.HP+heal, 0, run.Player.MaxHP)
		messages = append(messages, fmt.Sprintf("你饮下药剂，恢复 %d 点生命。", heal))
	case "flee":
		if enemy.Boss {
			return errors.New("王庭的大门已经封闭，无法逃跑")
		}
		raw := e.roll(run, "flee", 20)
		bonus := run.Player.Attributes.Dexterity
		dc := 13
		if e.hasTalent(run, "ranger_quickstep") {
			dc -= 2
		}
		total := raw + bonus
		success := total >= dc
		run.LastRoll = &RollResult{Kind: "check", Label: "撤退", Roll: raw, Bonus: bonus, Total: total, DC: dc, Success: success, Critical: raw == 20}
		if success {
			run.Combat = nil
			messages = append(messages, "你成功脱离战斗。这个房间仍然危险。")
			e.log(run, "combat", strings.Join(messages, " "))
			return nil
		}
		messages = append(messages, "你没能摆脱敌人。")
	default:
		return errors.New("未知战斗行动")
	}

	// 敌人的上一轮防御在玩家完成行动后失效。
	combat.Counters["enemy_guard"] = 0
	if combat.EnemyHP <= 0 {
		e.winCombat(run, enemy)
		return nil
	}

	// 持续伤害在敌人行动前结算，因此可以真正打断即将发生的意图。
	e.tickEnemyStatuses(run, &messages)
	if combat.EnemyHP <= 0 {
		e.winCombat(run, enemy)
		return nil
	}

	e.updateBossPhase(run, enemy, &messages)
	if interrupted {
		messages = append(messages, "你的重击打断了敌人的蓄力意图！")
	} else {
		e.resolveEnemyIntent(run, enemy, &messages)
	}
	if run.GameOver {
		combat.LastMessage = strings.Join(messages, " ")
		e.log(run, "combat", combat.LastMessage)
		return nil
	}

	e.tickPlayerStatuses(combat)
	e.tickCooldowns(combat)
	combat.Intent = e.chooseEnemyIntent(run, enemy, combat)
	combat.LastMessage = strings.Join(messages, " ")
	e.log(run, "combat", combat.LastMessage)
	run.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) UseItem(run *Run, itemID string) error {
	return e.ItemAction(run, itemID, "use")
}

func (e *Engine) ItemAction(run *Run, itemID, action string) error {
	if !hasItem(run.Player.Inventory, itemID) {
		return errors.New("背包中没有该物品")
	}
	item, ok := Items[itemID]
	if !ok {
		return errors.New("未知物品")
	}
	if run.Player.Equipment == nil {
		run.Player.Equipment = map[string]string{}
	}
	if action == "equip" {
		if run.Combat != nil {
			return errors.New("战斗中不能更换装备")
		}
		if item.Slot == "" {
			return errors.New("该物品不能装备")
		}
		return e.equipItem(run, itemID)
	}
	if action != "use" && action != "" {
		return errors.New("未知物品操作")
	}
	switch itemID {
	case "healing_draught", "bandage", "star_salve":
		if run.Combat != nil {
			return errors.New("战斗中请使用战斗行动栏的药剂动作")
		}
		if run.Player.HP >= run.Player.MaxHP {
			return errors.New("当前生命已经满了")
		}
		removeItem(&run.Player, itemID)
		run.Player.HP = clamp(run.Player.HP+item.Power, 0, run.Player.MaxHP)
		e.log(run, "item", fmt.Sprintf("你使用%s，恢复 %d 点生命。", item.Name, item.Power))
	case "luminous_tonic":
		if run.Combat != nil {
			return errors.New("战斗中无法从背包饮用辉光药剂")
		}
		if run.Player.Energy >= run.Player.MaxEnergy {
			return errors.New("当前能量已经满了")
		}
		removeItem(&run.Player, itemID)
		run.Player.Energy = clamp(run.Player.Energy+item.Power, 0, run.Player.MaxEnergy)
		e.log(run, "item", fmt.Sprintf("你饮下%s，恢复 %d 点能量。", item.Name, item.Power))
	case "frost_salt":
		if run.Combat == nil {
			return errors.New("霜盐只能在战斗中使用")
		}
		removeItem(&run.Player, itemID)
		run.Flags["frost_salt_active"] = true
		e.log(run, "item", "你把霜盐撒向敌人，它下一次命中的伤害会降低。")
		e.spendCombatItemTurn(run, item.Name)
	case "smoke_bomb":
		if run.Combat == nil {
			return errors.New("烟弹只能在战斗中使用")
		}
		removeItem(&run.Player, itemID)
		run.Combat.Guarded = true
		run.Combat.Intent = e.chooseEnemyIntent(run, Enemies[run.Combat.EnemyID], run.Combat)
		e.log(run, "item", "黑灰烟雾遮住战场，敌人原本的攻击节奏被打乱。")
		e.spendCombatItemTurn(run, item.Name)
	case "ember_flask":
		if run.Combat == nil {
			return errors.New("灼灰瓶只能在战斗中使用")
		}
		removeItem(&run.Player, itemID)
		run.Combat.EnemyStatuses = addStatus(run.Combat.EnemyStatuses, StatusState{ID: "burn", Name: "燃烧", Description: "灼灰持续燃烧。", Rounds: 3, Stacks: 1})
		e.log(run, "item", "灼灰瓶在敌人身上炸开，施加了 3 回合燃烧。")
		e.spendCombatItemTurn(run, item.Name)
	case "storm_phial":
		if run.Combat == nil {
			return errors.New("瓶装灰暴只能在战斗中使用")
		}
		removeItem(&run.Player, itemID)
		run.Combat.EnemyStatuses = addStatus(run.Combat.EnemyStatuses, StatusState{ID: "weakened", Name: "灰暴弱化", Description: "攻击节奏被灰暴打乱。", Rounds: 2, Stacks: 1})
		run.Combat.Guarded = true
		e.log(run, "item", "你释放瓶装灰暴，敌人被弱化，你也借风势进入防御姿态。")
		e.spendCombatItemTurn(run, item.Name)
	default:
		if item.Slot != "" {
			return e.equipItem(run, itemID)
		}
		return errors.New("该物品无法主动使用")
	}
	run.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) Interact(run *Run, elementID string) error {
	if run.GameOver {
		return errors.New("本次冒险已经结束")
	}
	if run.Combat != nil {
		return errors.New("战斗中无法调查场景")
	}
	if run.ActiveEvent != nil {
		return errors.New("请先处理当前事件")
	}
	room := run.Rooms[run.CurrentRoomID]
	if room == nil {
		return errors.New("当前房间不存在")
	}
	var el *SceneElement
	for i := range room.Elements {
		if room.Elements[i].ID == elementID {
			el = &room.Elements[i]
			break
		}
	}
	if el == nil {
		return errors.New("该场景元素不存在")
	}
	if el.HiddenUnlessFlag != "" && !run.Flags[el.HiddenUnlessFlag] {
		return errors.New("你还没有发现这个目标")
	}
	if el.RequiresFlag != "" && !run.Flags[el.RequiresFlag] {
		return errors.New("缺少触发这个机关所需的线索或世界状态")
	}
	if el.RequiresItem != "" && !hasItem(run.Player.Inventory, el.RequiresItem) {
		name := el.RequiresItem
		if item, ok := Items[el.RequiresItem]; ok {
			name = item.Name
		}
		return errors.New("需要物品：「" + name + "」")
	}
	flag := "interacted:" + room.ID + ":" + el.ID
	if el.OneShot && run.Flags[flag] {
		return errors.New("这里已经调查过了")
	}
	run.LastRoll = nil

	if el.Check != nil {
		roll := e.roll(run, "interact:"+room.ID+":"+el.ID, 20)
		bonus := e.attribute(run.Player.Attributes, el.Check.Attribute)
		if e.hasTalent(run, "ranger_scavenger") && el.Check.Attribute == "perception" {
			bonus++
		}
		total := roll + bonus
		success := total >= el.Check.DC
		run.LastRoll = &RollResult{Kind: "check", Label: el.Label, Roll: roll, Bonus: bonus, Total: total, DC: el.Check.DC, Success: success, Critical: roll == 20}
		if !success {
			e.applyEffects(run, el.OnFailure)
			e.log(run, "explore", fmt.Sprintf("调查「%s」失败：D20 %d + %d = %d / DC %d。", el.Label, roll, bonus, total, el.Check.DC))
			run.UpdatedAt = time.Now()
			return nil
		}
		e.applyEffects(run, el.OnSuccess)
		e.log(run, "explore", fmt.Sprintf("调查「%s」成功：D20 %d + %d = %d / DC %d。", el.Label, roll, bonus, total, el.Check.DC))
	}

	if el.OneShot {
		run.Flags[flag] = true
	}
	if el.ConsumesItem && el.RequiresItem != "" {
		removeItem(&run.Player, el.RequiresItem)
	}
	if err := e.executeElementAction(run, room, el); err != nil {
		if el.OneShot {
			delete(run.Flags, flag)
		}
		return err
	}
	run.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) triggerRoom(run *Run, room *Room) {
	switch room.Type {
	case "combat":
		ids := []string{"bone_thrall", "grave_spider", "ash_hound", "royal_guard", "tomb_crow", "ash_cultist", "stone_sentinel", "chain_wraith", "ink_wraith", "memory_knight"}
		id := ids[int(hash64(fmt.Sprintf("%d-%s", run.Seed, room.ID))%uint64(len(ids)))]
		e.startCombat(run, id)
	case "boss":
		e.startCombat(run, "crown_bearer")
	case "ashfield", "ruins":
		ids := []string{"ash_raider", "glass_walker", "dune_wraith", "sky_leech", "storm_knight"}
		id := ids[int(hash64(fmt.Sprintf("v04-combat:%d:%s", run.Seed, room.ID))%uint64(len(ids)))]
		e.startCombat(run, id)
	case "finalboss":
		e.startCombat(run, "gate_heart")
	case "npc":
		if !run.Flags["met_scout"] {
			run.Flags["met_scout"] = true
			e.startEvent(run, "wounded_scout", room)
		} else if q := run.Quests["lost_patrol"]; q != nil && q.Status == "active" && !run.Flags["captain_found"] {
			run.Flags["captain_found"] = true
			e.startEvent(run, "forgotten_captain", room)
		} else {
			room.Resolved = true
			e.log(run, "story", "营地里没有新的幸存者，但留下了可以调查的补给和笔记。")
		}
	case "event":
		pool := []string{"sealed_reliquary", "memory_mural", "black_well", "echo_child", "sealed_door", "cultist_cache", "royal_confessor", "crow_nest", "memory_mirror"}
		id := pool[int(hash64(fmt.Sprintf("event-%d-%s", run.Seed, room.ID))%uint64(len(pool)))]
		for i := 0; i < len(pool); i++ {
			if !run.Flags["seen_"+id] {
				break
			}
			id = pool[(indexOf(pool, id)+1)%len(pool)]
		}
		run.Flags["seen_"+id] = true
		e.startEvent(run, id, room)
	case "chapel", "library", "flooded", "prison", "garden", "ossuary", "forge", "observatory", "banquet", "merchant", "rest", "shrine", "treasure", "secret", "infirmary", "gatehouse", "aqueduct", "bridge", "court", "belltower", "reliquary", "mausoleum", "frontier", "caravan", "glassmarsh", "crater", "starwatch", "windshrine", "meteor", "windcamp", "outergate":
		// 特色地点不强制弹事件；玩家直接点击场景中的热点触发内容。
		room.Resolved = true
		e.log(run, "story", "这里有多个可调查目标。场景中发光的标记可以直接点击。")
	default:
		room.Resolved = true
	}
}

func (e *Engine) startEvent(run *Run, id string, room *Room) {
	def := Events[id]
	room.EventID = id
	copyChoices := append([]ChoiceDef(nil), def.Choices...)
	run.ActiveEvent = &ActiveEvent{ID: def.ID, Title: def.Title, Description: def.Description, Choices: copyChoices}
	e.log(run, "event", fmt.Sprintf("触发事件：%s。", def.Title))
}

func (e *Engine) startCombat(run *Run, enemyID string) {
	enemy := Enemies[enemyID]
	hpBonus := run.Clock.Threat * 2
	if enemy.Boss {
		hpBonus = run.Clock.Threat * 3
	}
	combat := &CombatState{
		EnemyID: enemy.ID, EnemyName: enemy.Name, EnemyHP: enemy.HP + hpBonus, EnemyMaxHP: enemy.HP + hpBonus,
		EnemyDefense: enemy.Defense + run.Clock.Threat/3, Round: 0, LastMessage: enemy.Description,
		BossPhase: 1, Cooldowns: map[string]int{}, Counters: map[string]int{},
	}
	if e.hasTalent(run, "seer_reservoir") {
		run.Player.Energy = clamp(run.Player.Energy+2, 0, run.Player.MaxEnergy)
	}
	if e.hasTalent(run, "seer_ward") {
		combat.Counters["seer_ward"] = 1
	}
	combat.Intent = e.chooseEnemyIntent(run, enemy, combat)
	run.Combat = combat
	e.log(run, "combat", fmt.Sprintf("%s出现了。%s 敌人意图：%s。", enemy.Name, enemy.Description, combat.Intent.Label))
}

func (e *Engine) winCombat(run *Run, enemy EnemyDef) {
	room := run.Rooms[run.CurrentRoomID]
	if room != nil {
		room.Resolved = true
	}
	gold := enemy.GoldMin
	if enemy.GoldMax > enemy.GoldMin {
		gold += e.roll(run, "loot-gold", enemy.GoldMax-enemy.GoldMin+1) - 1
	}
	run.Player.Gold += gold
	run.Player.XP += enemy.XP
	msg := fmt.Sprintf("你击败了%s，获得 %d 金币与 %d 经验。", enemy.Name, gold, enemy.XP)

	if !enemy.Boss {
		lootPool := []string{"bandage", "healing_draught", "frost_salt", "luminous_tonic"}
		chance := e.roll(run, "loot-drop", 100)
		dropThreshold := 36
		if e.hasTalent(run, "ranger_scavenger") {
			dropThreshold = 52
		}
		if chance <= dropThreshold {
			itemID := lootPool[int(hash64(fmt.Sprintf("drop:%d:%s:%d", run.Seed, enemy.ID, run.Turn))%uint64(len(lootPool)))]
			run.Player.Inventory = append(run.Player.Inventory, itemID)
			msg += " 敌人还掉落了「" + Items[itemID].Name + "」。"
		}
	}

	if enemy.Boss {
		if enemy.ID == "crown_bearer" {
			run.Flags["defeated_crown_bearer"] = true
			if q := run.Quests["ashen_crown"]; q != nil {
				q.Progress = 1
				q.Status = "completed"
			}
			if room := run.Rooms["room_33"]; room != nil {
				room.Locked = false
				room.Discovered = true
			}
			if run.Quests["beyond_mist"] == nil {
				run.Quests["beyond_mist"] = &QuestState{ID: "beyond_mist", Title: "第二幕 · 雾外荒原", Description: "赫里昂倒下后，王座后的石门打开。沿灰风向东，找到真正被王冠压住的那扇门。", Status: "active", Goal: 1}
			}
			run.Flags["act2_unlocked"] = true
			msg += " 赫里昂没有化成灰。他把最后一道王冠锁链扯断，王座后的石墙随之裂开。冷风第一次从墓城之外吹入——真正的封印还在雾外荒原。第二幕已经开启。"
			e.log(run, "act", "第二幕开启：雾外荒原。地图东侧出现一条通往地表的道路。")
		} else if enemy.ID == "gate_heart" {
			run.Flags["defeated_gate_heart"] = true
			if q := run.Quests["beyond_mist"]; q != nil {
				q.Progress = 1
				q.Status = "completed"
			}
			run.Victory = true
			run.GameOver = true
			ending := "门后之心停止跳动。灰雾第一次被真正的风吹散，你看见维尔之外仍有无数封印遗迹。你没有结束这个世界的危险，但至少终止了这一次苏醒。"
			if run.Flags["accepted_wind_oath"] && run.Flags["heard_true_name"] && hasItem(run.Player.Inventory, "void_compass") {
				ending = "你用赫里昂的真名固定最后一条锁链，再以风誓把门后的心跳转移到无人记得的星图中。黎明照进荒原，墓城与灰雾同时开始退去。维尔第一次从历史里重新拥有名字。"
			} else if run.Flags["touched_star_wound"] && run.Flags["knows_last_price"] {
				ending = "你以坠星之力切断封印的遗忘循环。门关闭了，但代价是从此只有你记得赫里昂、墓城和这场战争。荒原上的人会称你为一个没有来历的旅者。"
			}
			msg += " " + ending
		}
	}
	run.Combat = nil
	e.log(run, "combat", msg)
	e.levelUp(run)
	run.UpdatedAt = time.Now()
}

func (e *Engine) applyEffects(run *Run, effects []Effect) {
	for _, ef := range effects {
		switch ef.Type {
		case "gold":
			run.Player.Gold = max(0, run.Player.Gold+ef.Value)
		case "hp":
			delta := ef.Value
			if delta < 0 && run.Player.Equipment["trinket"] == "ashen_charm" {
				delta = min(0, delta+1)
			}
			run.Player.HP = clamp(run.Player.HP+delta, 0, run.Player.MaxHP)
		case "heal":
			run.Player.HP = clamp(run.Player.HP+ef.Value, 0, run.Player.MaxHP)
		case "max_hp":
			run.Player.MaxHP += ef.Value
			run.Player.HP += ef.Value
		case "energy":
			if ef.Value > 50 {
				run.Player.Energy = run.Player.MaxEnergy
			} else {
				run.Player.Energy = clamp(run.Player.Energy+ef.Value, 0, run.Player.MaxEnergy)
			}
		case "item":
			run.Player.Inventory = append(run.Player.Inventory, ef.Target)
			e.log(run, "loot", "获得物品：「"+Items[ef.Target].Name+"」。")
		case "consume_item":
			removeItem(&run.Player, ef.Target)
		case "flag":
			run.Flags[ef.Target] = true
		case "xp":
			run.Player.XP += ef.Value
			e.levelUp(run)
		case "quest":
			if run.Quests[ef.Target] == nil {
				switch ef.Target {
				case "lost_patrol":
					run.Quests[ef.Target] = &QuestState{ID: ef.Target, Title: "支线 · 失踪巡夜队", Description: "寻找伊文失踪的队长，并查明巡夜队在王庭门前遭遇了什么。", Status: "active", Goal: 1}
				case "nameless_prisoner":
					run.Quests[ef.Target] = &QuestState{ID: ef.Target, Title: "支线 · 被抹去的名字", Description: "寻找能够恢复囚徒姓名的王族记录。王名密室也许保存着答案。", Status: "active", Goal: 1}
				}
			}
		case "quest_complete":
			if q := run.Quests[ef.Target]; q != nil {
				q.Status = "completed"
				q.Progress = q.Goal
			}
		case "lore":
			if entry, ok := Lore[ef.Target]; ok {
				e.addLore(run, entry)
			}
		case "unlock":
			if room := run.Rooms[ef.Target]; room != nil {
				room.Locked = false
				room.Discovered = true
				e.log(run, "world", "隐藏道路已经打开："+room.Name+"。")
			}
		case "combat":
			if _, ok := Enemies[ef.Target]; ok && run.Combat == nil {
				e.startCombat(run, ef.Target)
			}
		case "message":
			e.log(run, "story", ef.Text)
		}
	}
}

func (e *Engine) validateEffects(run *Run, effects []Effect) error {
	for _, ef := range effects {
		if ef.Type == "gold" && ef.Value < 0 && run.Player.Gold < -ef.Value {
			return errors.New("金币不足")
		}
		if ef.Type == "consume_item" && !hasItem(run.Player.Inventory, ef.Target) {
			return errors.New("缺少所需物品：" + Items[ef.Target].Name)
		}
	}
	return nil
}

func (e *Engine) revealNeighbors(run *Run, id string) {
	for _, ed := range run.Edges {
		if ed.From == id {
			if r := run.Rooms[ed.To]; r != nil && !r.Locked {
				r.Discovered = true
			}
		}
		if ed.To == id {
			if r := run.Rooms[ed.From]; r != nil && !r.Locked {
				r.Discovered = true
			}
		}
	}
}

func (e *Engine) adjacent(run *Run, a, b string) bool {
	for _, ed := range run.Edges {
		if (ed.From == a && ed.To == b) || (ed.From == b && ed.To == a) {
			return true
		}
	}
	return false
}

func (e *Engine) attribute(a AttributeSet, key string) int {
	switch key {
	case "strength":
		return a.Strength
	case "dexterity":
		return a.Dexterity
	case "perception":
		return a.Perception
	case "will":
		return a.Will
	}
	return 0
}

func (e *Engine) roll(run *Run, label string, sides int) int {
	if sides <= 1 {
		return 1
	}
	v := hash64(fmt.Sprintf("%d:%d:%s", run.Seed, run.Turn, label))
	run.Turn++
	return int(v%uint64(sides)) + 1
}

func (e *Engine) weaponPower(run *Run) int {
	if run.Player.Equipment != nil {
		if id := run.Player.Equipment["weapon"]; id != "" {
			if item, ok := Items[id]; ok && item.Type == "weapon" && hasItem(run.Player.Inventory, id) {
				return item.Power
			}
		}
	}
	for _, id := range run.Player.Inventory {
		if item, ok := Items[id]; ok && item.Type == "weapon" {
			return item.Power
		}
	}
	return 0
}

func (e *Engine) levelUp(run *Run) {
	for run.Player.XP >= run.Player.Level*40 {
		run.Player.XP -= run.Player.Level * 40
		run.Player.Level++
		run.Player.MaxHP += 4
		run.Player.HP = run.Player.MaxHP
		run.Player.MaxEnergy++
		run.Player.Energy = run.Player.MaxEnergy
		run.Player.TalentPoints++
		e.log(run, "level", fmt.Sprintf("你升到了 %d 级，并获得 1 点天赋点。生命与能量得到恢复。", run.Player.Level))
	}
}

func (e *Engine) addLore(run *Run, entry LoreEntry) {
	if entry.ID == "" {
		return
	}
	for _, existing := range run.Lore {
		if existing.ID == entry.ID {
			return
		}
	}
	run.Lore = append(run.Lore, entry)
}

func (e *Engine) log(run *Run, typ, msg string) {
	run.Log = append(run.Log, LogEntry{Turn: run.Turn, Type: typ, Message: msg, CreatedAt: time.Now()})
	if len(run.Log) > 120 {
		run.Log = run.Log[len(run.Log)-120:]
	}
}

func hash64(s string) uint64 { h := fnv.New64a(); _, _ = h.Write([]byte(s)); return h.Sum64() }
func clamp(v, minv, maxv int) int {
	if v < minv {
		return minv
	}
	if v > maxv {
		return maxv
	}
	return v
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func hasItem(items []string, id string) bool {
	for _, v := range items {
		if v == id {
			return true
		}
	}
	return false
}
func removeItem(p *Player, id string) bool {
	for i, v := range p.Inventory {
		if v == id {
			p.Inventory = append(p.Inventory[:i], p.Inventory[i+1:]...)
			return true
		}
	}
	return false
}
func indexOf(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return 0
}

func SortedRooms(run *Run) []*Room {
	out := make([]*Room, 0, len(run.Rooms))
	for _, r := range run.Rooms {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

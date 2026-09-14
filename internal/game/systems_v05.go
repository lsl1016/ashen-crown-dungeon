package game

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

func initialShopStock() map[string]int {
	stock := map[string]int{}
	for shopID, shop := range Shops {
		for _, entry := range shop.Items {
			stock[shopID+":"+entry.ItemID] = entry.Stock
		}
	}
	return stock
}

func (e *Engine) ensureV05State(run *Run) {
	if run.Player.Growth == nil {
		run.Player.Growth = map[string]int{}
	}
	if run.NPCRelations == nil {
		run.NPCRelations = map[string]int{}
	}
	if run.ShopStock == nil {
		run.ShopStock = initialShopStock()
	}
	if run.Flags == nil {
		run.Flags = map[string]bool{}
	}
	if run.Quests == nil {
		run.Quests = map[string]*QuestState{}
	}
}

func (e *Engine) StartDialogue(run *Run, npcID string) error {
	e.ensureV06State(run)
	if run.GameOver {
		return errors.New("冒险已经结束")
	}
	if run.Combat != nil || run.ActiveEvent != nil {
		return errors.New("当前状态无法开始对话")
	}
	if run.ActiveShop != "" {
		return errors.New("请先结束交易")
	}
	npc, ok := NPCs[npcID]
	if !ok {
		return errors.New("NPC 不存在")
	}
	if loc := run.NPCLocations[npcID]; loc != "" && loc != run.CurrentRoomID {
		if room := run.Rooms[loc]; room != nil {
			return fmt.Errorf("%s现在不在这里；最近有人在「%s」见过%s", npc.Name, room.Name, npc.Name)
		}
		return errors.New("这个 NPC 现在不在这里")
	}
	dlg, ok := Dialogues[npc.DialogueID]
	if !ok {
		return errors.New("对话树不存在")
	}
	start := e.dialogueStartNode(run, npcID, dlg.StartNode)
	if _, ok := dlg.Nodes[start]; !ok {
		start = dlg.StartNode
	}
	run.ActiveDialogue = &ActiveDialogue{NPCID: npcID, DialogueID: dlg.ID, NodeID: start}
	e.log(run, "dialogue", fmt.Sprintf("你与%s交谈。", npc.Name))
	run.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) DialogueChoice(run *Run, choiceID string) error {
	e.ensureV06State(run)
	if run.ActiveDialogue == nil {
		return errors.New("当前没有进行中的对话")
	}
	active := run.ActiveDialogue
	dlg, ok := Dialogues[active.DialogueID]
	if !ok {
		return errors.New("对话树不存在")
	}
	node, ok := dlg.Nodes[active.NodeID]
	if !ok {
		return errors.New("当前对话节点不存在")
	}
	var choice *DialogueChoiceDef
	for i := range node.Choices {
		if node.Choices[i].ID == choiceID {
			choice = &node.Choices[i]
			break
		}
	}
	if choice == nil {
		return errors.New("未知对话选项")
	}
	if choice.RequiresFlag != "" && !run.Flags[choice.RequiresFlag] {
		return errors.New("当前还无法选择这句话")
	}
	if choice.RequiresItem != "" && !hasItem(run.Player.Inventory, choice.RequiresItem) {
		return errors.New("缺少对话所需物品")
	}
	if err := e.validateEffects(run, choice.Effects); err != nil {
		return err
	}
	e.applyEffects(run, choice.Effects)
	npc := NPCs[active.NPCID]
	e.log(run, "dialogue", fmt.Sprintf("你对%s说：「%s」", npc.Name, choice.Text))

	if choice.OpenShop != "" {
		if _, ok := Shops[choice.OpenShop]; !ok {
			return errors.New("商店不存在")
		}
		run.ActiveShop = choice.OpenShop
		run.ActiveDialogue = nil
		e.log(run, "shop", "开始交易："+Shops[choice.OpenShop].Name+"。")
	} else if choice.End {
		run.ActiveDialogue = nil
	} else if choice.NextNode != "" {
		if _, ok := dlg.Nodes[choice.NextNode]; !ok {
			return errors.New("下一个对话节点不存在")
		}
		run.ActiveDialogue.NodeID = choice.NextNode
	} else {
		run.ActiveDialogue = nil
	}
	run.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) CloseDialogue(run *Run) error {
	if run.ActiveDialogue == nil {
		return nil
	}
	run.ActiveDialogue = nil
	run.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) ShopAction(run *Run, shopID, action, itemID string) error {
	e.ensureV06State(run)
	shop, ok := Shops[shopID]
	if !ok {
		return errors.New("商店不存在")
	}
	if run.Combat != nil || run.ActiveEvent != nil {
		return errors.New("当前无法交易")
	}
	if action == "open" {
		run.ActiveDialogue = nil
		run.ActiveShop = shopID
		return nil
	}
	if action == "close" {
		run.ActiveShop = ""
		e.log(run, "shop", "结束交易。")
		return nil
	}
	if run.ActiveShop != shopID {
		return errors.New("请先与对应商人开始交易")
	}
	item, ok := Items[itemID]
	if !ok {
		return errors.New("物品不存在")
	}
	key := shopID + ":" + itemID
	switch action {
	case "buy":
		price := -1
		for _, entry := range shop.Items {
			if entry.ItemID == itemID {
				price = entry.Price
				break
			}
		}
		if price < 0 {
			return errors.New("该商店不出售这件物品")
		}
		if run.ShopStock[key] <= 0 {
			return errors.New("已经售罄")
		}
		if run.Player.Gold < price {
			return errors.New("金币不足")
		}
		run.Player.Gold -= price
		run.Player.Inventory = append(run.Player.Inventory, itemID)
		if item.Slot != "" {
			e.ensureItemAffix(run, itemID, "shop:"+shopID, item.Rarity == "rare" || item.Rarity == "legendary")
		}
		run.ShopStock[key]--
		e.log(run, "shop", fmt.Sprintf("购买「%s」，花费 %d 枚古金币。", item.Name, price))
	case "sell":
		if !hasItem(run.Player.Inventory, itemID) {
			return errors.New("背包里没有这件物品")
		}
		for _, equippedID := range run.Player.Equipment {
			if equippedID == itemID {
				return errors.New("已装备物品不能直接出售")
			}
		}
		switch item.Type {
		case "weapon", "armor", "trinket", "consumable":
		default:
			return errors.New("剧情或关键物品不能出售")
		}
		value := int(math.Round(float64(max(1, item.Value)) * shop.Buyback))
		value = max(1, value)
		removeItem(&run.Player, itemID)
		run.Player.Gold += value
		run.ShopStock[key]++
		e.log(run, "shop", fmt.Sprintf("出售「%s」，获得 %d 枚古金币。", item.Name, value))
	default:
		return errors.New("未知交易操作")
	}
	run.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) UpgradeAttribute(run *Run, attribute string) error {
	e.ensureV06State(run)
	if run.Combat != nil || run.ActiveEvent != nil || run.ActiveDialogue != nil || run.ActiveShop != "" {
		return errors.New("当前状态不能分配属性点")
	}
	if run.Player.AttributePoints <= 0 {
		return errors.New("没有可用属性点")
	}
	switch attribute {
	case "strength":
		run.Player.Attributes.Strength++
	case "dexterity":
		run.Player.Attributes.Dexterity++
	case "perception":
		run.Player.Attributes.Perception++
	case "will":
		run.Player.Attributes.Will++
	default:
		return errors.New("未知属性")
	}
	run.Player.AttributePoints--
	e.log(run, "growth", fmt.Sprintf("投入 1 点属性：%s。", attributeCN(attribute)))
	run.UpdatedAt = time.Now()
	return nil
}

func attributeCN(id string) string {
	switch id {
	case "strength":
		return "力量"
	case "dexterity":
		return "敏捷"
	case "perception":
		return "感知"
	case "will":
		return "意志"
	default:
		return id
	}
}

func (e *Engine) UpgradeGrowth(run *Run, growthID string) error {
	e.ensureV06State(run)
	if run.Combat != nil || run.ActiveEvent != nil || run.ActiveDialogue != nil || run.ActiveShop != "" {
		return errors.New("当前状态不能进行精通训练")
	}
	if run.Player.MasteryPoints <= 0 {
		return errors.New("没有可用精通点")
	}
	var def *GrowthDef
	for i := range Growth {
		if Growth[i].ID == growthID {
			def = &Growth[i]
			break
		}
	}
	if def == nil {
		return errors.New("未知精通")
	}
	rank := run.Player.Growth[growthID]
	if rank >= def.MaxRank {
		return errors.New("该精通已达到最高等级")
	}
	rank++
	run.Player.Growth[growthID] = rank
	run.Player.MasteryPoints--
	if growthID == "survivor_instinct" {
		run.Player.MaxHP += 3
		run.Player.HP = clamp(run.Player.HP+3, 0, run.Player.MaxHP)
		if rank == 3 {
			run.Player.Defense++
		}
	}
	e.log(run, "growth", fmt.Sprintf("精通「%s」提升到 %d 级。", def.Name, rank))
	run.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) growthRank(run *Run, id string) int {
	if run.Player.Growth == nil {
		return 0
	}
	return run.Player.Growth[id]
}

func (e *Engine) dialogueNode(run *Run) (NPCDef, DialogueNodeDef, bool) {
	if run.ActiveDialogue == nil {
		return NPCDef{}, DialogueNodeDef{}, false
	}
	npc, ok := NPCs[run.ActiveDialogue.NPCID]
	if !ok {
		return NPCDef{}, DialogueNodeDef{}, false
	}
	dlg, ok := Dialogues[run.ActiveDialogue.DialogueID]
	if !ok {
		return NPCDef{}, DialogueNodeDef{}, false
	}
	node, ok := dlg.Nodes[run.ActiveDialogue.NodeID]
	return npc, node, ok
}

func relationLabel(v int) string {
	switch {
	case v >= 5:
		return "信任"
	case v >= 2:
		return "友善"
	case v <= -2:
		return "戒备"
	default:
		return "中立"
	}
}

func (e *Engine) applyDialogueEffect(run *Run, ef Effect) bool {
	e.ensureV06State(run)
	switch ef.Type {
	case "relation":
		run.NPCRelations[ef.Target] = clamp(run.NPCRelations[ef.Target]+ef.Value, -10, 10)
		if npc, ok := NPCs[ef.Target]; ok {
			e.log(run, "relation", fmt.Sprintf("%s 对你的态度：%s（%+d）。", npc.Name, relationLabel(run.NPCRelations[ef.Target]), ef.Value))
		}
		return true
	case "item_once":
		flag := "unique_item:" + ef.Target
		if !run.Flags[flag] && !hasItem(run.Player.Inventory, ef.Target) {
			run.Player.Inventory = append(run.Player.Inventory, ef.Target)
			run.Flags[flag] = true
			if it, ok := Items[ef.Target]; ok {
				e.log(run, "loot", "获得物品：「"+it.Name+"」。")
			}
		}
		return true
	}
	return false
}

func v05Quest(run *Run, id string) *QuestState {
	switch id {
	case "storm_hunt":
		return &QuestState{ID: id, Title: "支线 · 空钟猎人", Description: "击败一名风暴骑士，带回其胸甲上的空钟纹路情报。", Status: "active", Stage: "hunt_storm_knight", Goal: 1}
	}
	return nil
}

func isBlank(s string) bool { return strings.TrimSpace(s) == "" }

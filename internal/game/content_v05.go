package game

// V0.5 focuses on RPG-facing systems: persistent NPCs, dialogue trees,
// real shop inventory and deeper character growth. These are still data-driven
// and intentionally separate from any future AI runtime.

var Growth = []GrowthDef{
	{ID: "weapon_training", Name: "武器训练", Description: "普通攻击伤害每级 +1；第 3 级额外获得 +1 命中。", Icon: "⚔", MaxRank: 3},
	{ID: "signature_mastery", Name: "职业技精通", Description: "职业技基础伤害每级 +2；第 3 级能量消耗从 3 降为 2。", Icon: "✦", MaxRank: 3},
	{ID: "survivor_instinct", Name: "幸存者本能", Description: "每级最大生命 +3；第 3 级额外获得 +1 防御。", Icon: "⛨", MaxRank: 3},
}

var NPCs = map[string]NPCDef{
	"iven": {
		ID: "iven", Name: "伊文", Title: "巡夜团记录员", Faction: "巡夜团",
		Description: "年轻的巡夜记录员，左臂缠着刚换过的绷带。他把每个失踪者的名字都写在袖口。",
		Portrait:    "/assets/portraits/npc_iven.svg", DialogueID: "dlg_iven",
	},
	"veil_broker": {
		ID: "veil_broker", Name: "帷面客", Title: "无脸商会行商", Faction: "无脸商会",
		Description: "银白面具没有开孔，但他总能准确看向你手里的金币。",
		Portrait:    "/assets/portraits/npc_broker.svg", DialogueID: "dlg_broker", ShopID: "veil_market",
	},
	"nameless_prisoner": {
		ID: "nameless_prisoner", Name: "无名囚徒", Title: "被王史抹去的人", Faction: "旧王残响",
		Description: "他仍然活着，或者至少还记得如何呼吸。锁链上的每一节都刻着一个被划掉的名字。",
		Portrait:    "/assets/portraits/npc_prisoner.svg", DialogueID: "dlg_prisoner",
	},
	"mara": {
		ID: "mara", Name: "玛拉", Title: "逐风斥候", Faction: "巡夜团",
		Description: "来自雾外的斥候，护目镜被镜砂划出细密白痕。她声称自己已经三次走到黑门前，又三次被风送回。",
		Portrait:    "/assets/portraits/npc_mara.svg", DialogueID: "dlg_mara", ShopID: "wind_supply",
	},
	"orrin": {
		ID: "orrin", Name: "奥林", Title: "最后的无旗骑士", Faction: "旧王残响",
		Description: "盔甲里没有肉身，只有一团维持姿势的灰。他拒绝再举王旗，却仍守着通往东界的道路。",
		Portrait:    "/assets/portraits/npc_orrin.svg", DialogueID: "dlg_orrin",
	},
}

var Shops = map[string]ShopDef{
	"veil_market": {
		ID: "veil_market", Name: "无脸集市 · 帷面客", NPCID: "veil_broker", Buyback: 0.45,
		Items: []ShopItemDef{
			{ItemID: "healing_draught", Price: 15, Stock: 5}, {ItemID: "luminous_tonic", Price: 22, Stock: 3},
			{ItemID: "smoke_bomb", Price: 24, Stock: 3}, {ItemID: "frost_salt", Price: 18, Stock: 3},
			{ItemID: "chainmail", Price: 48, Stock: 1}, {ItemID: "glass_knife", Price: 62, Stock: 1},
			{ItemID: "night_lantern", Price: 50, Stock: 1}, {ItemID: "ember_flask", Price: 36, Stock: 2},
			{ItemID: "veil_thread", Price: 44, Stock: 1},
		},
	},
	"wind_supply": {
		ID: "wind_supply", Name: "逐风补给 · 玛拉", NPCID: "mara", Buyback: 0.50,
		Items: []ShopItemDef{
			{ItemID: "star_salve", Price: 36, Stock: 3}, {ItemID: "storm_phial", Price: 42, Stock: 2},
			{ItemID: "wasteland_cloak", Price: 78, Stock: 1}, {ItemID: "moon_dagger", Price: 38, Stock: 1},
			{ItemID: "windglass_charm", Price: 66, Stock: 1},
		},
	},
}

var Dialogues = map[string]DialogueDef{
	"dlg_iven": {
		ID: "dlg_iven", NPCID: "iven", StartNode: "start", Nodes: map[string]DialogueNodeDef{
			"start": {ID: "start", Text: "你终于来了。第三巡夜队昨夜进了内环，只回来一盏灯。我知道规矩要求我等增援，但第十三声钟不会等。", Choices: []DialogueChoiceDef{
				{ID: "patrol", Text: "我去找第三巡夜队。", NextNode: "accept", Effects: []Effect{{Type: "quest", Target: "lost_patrol"}, {Type: "flag", Target: "iven_patrol_accepted"}, {Type: "relation", Target: "iven", Value: 2}}},
				{ID: "crown", Text: "先告诉我你知道的王冠。", NextNode: "crown"},
				{ID: "decline", Text: "我只为王冠而来。", NextNode: "decline", Effects: []Effect{{Type: "flag", Target: "iven_patrol_declined"}, {Type: "relation", Target: "iven", Value: -1}}},
				{ID: "leave", Text: "结束谈话。", End: true},
			}},
			"accept":  {ID: "accept", Text: "谢谢。队长叫赛勒。他有一枚黑铁哨，如果你找到它——哪怕只找到哨子——也带回来。", Choices: []DialogueChoiceDef{{ID: "done", Text: "我会记住。", End: true}}},
			"crown":   {ID: "crown", Text: "巡夜团旧档把王冠叫作‘门闩’。赫里昂不是戴着它统治，他是在用自己的名字压住什么。", Choices: []DialogueChoiceDef{{ID: "more", Text: "第三巡夜队也在查这个？", NextNode: "accept"}, {ID: "back", Text: "我知道了。", End: true, Effects: []Effect{{Type: "lore", Target: "crown_chains"}}}}},
			"decline": {ID: "decline", Text: "那就别在他们的尸体旁说这句话。墓城会记住。", Choices: []DialogueChoiceDef{{ID: "leave", Text: "离开。", End: true}}},
		},
	},
	"dlg_broker": {
		ID: "dlg_broker", NPCID: "veil_broker", StartNode: "start", Nodes: map[string]DialogueNodeDef{
			"start": {ID: "start", Text: "欢迎，仍有名字的客人。这里不问你从哪来，只问你愿意用什么换明天。", Choices: []DialogueChoiceDef{
				{ID: "trade", Text: "看看货物。", OpenShop: "veil_market", Effects: []Effect{{Type: "relation", Target: "veil_broker", Value: 1}}},
				{ID: "rumor", Text: "你听过第十三声钟吗？", NextNode: "rumor"},
				{ID: "face", Text: "你的脸在哪里？", NextNode: "face"},
				{ID: "leave", Text: "离开。", End: true},
			}},
			"rumor": {ID: "rumor", Text: "听过。每次都有人忘记自己听过。生意人最大的优势，就是把账写在纸上。", Choices: []DialogueChoiceDef{{ID: "mark", Text: "给我看看账。", End: true, Effects: []Effect{{Type: "lore", Target: "merchant_crate"}, {Type: "relation", Target: "veil_broker", Value: 1}}}, {ID: "trade", Text: "回到交易。", OpenShop: "veil_market"}}},
			"face":  {ID: "face", Text: "卖掉了。价格很好。你也会有一些以后不再需要的东西。", Choices: []DialogueChoiceDef{{ID: "leave", Text: "……算了。", End: true}}},
		},
	},
	"dlg_prisoner": {
		ID: "dlg_prisoner", NPCID: "nameless_prisoner", StartNode: "start", Nodes: map[string]DialogueNodeDef{
			"start": {ID: "start", Text: "先别问我是谁。这个问题他们问了三百年。问一点我还能回答的。", Choices: []DialogueChoiceDef{
				{ID: "name", Text: "我帮你找回名字。", NextNode: "name", Effects: []Effect{{Type: "quest", Target: "nameless_prisoner"}, {Type: "flag", Target: "prisoner_quest_accepted"}, {Type: "relation", Target: "nameless_prisoner", Value: 2}}},
				{ID: "crown", Text: "王冠在哪里？", NextNode: "crown"},
				{ID: "leave", Text: "保持距离。", End: true},
			}},
			"name":  {ID: "name", Text: "王族把名字锁在镜子和墓志里。拿着这把总钥，它也许能打开通往答案的第一扇门。", Choices: []DialogueChoiceDef{{ID: "take", Text: "接过钥匙。", End: true, Effects: []Effect{{Type: "item_once", Target: "prison_key"}}}}},
			"crown": {ID: "crown", Text: "王座上的东西是影子。真正的王冠是十三道锁链互相记得彼此。国王只是第十三个结。", Choices: []DialogueChoiceDef{{ID: "remember", Text: "我会记住。", End: true, Effects: []Effect{{Type: "flag", Target: "prisoner_warning"}, {Type: "relation", Target: "nameless_prisoner", Value: 1}}}}},
		},
	},
	"dlg_mara": {
		ID: "dlg_mara", NPCID: "mara", StartNode: "start", Nodes: map[string]DialogueNodeDef{
			"start": {ID: "start", Text: "如果你从墓城出来，那你已经比我想的麻烦。东边黑门附近有风暴骑士，我缺一个敢把它们从路上挪开的人。", Choices: []DialogueChoiceDef{
				{ID: "hunt", Text: "我接这个活。", NextNode: "hunt", Effects: []Effect{{Type: "quest", Target: "storm_hunt"}, {Type: "flag", Target: "mara_hunt_accepted"}, {Type: "relation", Target: "mara", Value: 2}}},
				{ID: "trade", Text: "先看看荒原补给。", OpenShop: "wind_supply"},
				{ID: "gate", Text: "你真的到过黑门？", NextNode: "gate"},
				{ID: "leave", Text: "暂时不用。", End: true},
			}},
			"hunt": {ID: "hunt", Text: "别数你杀了几只。找到它们胸甲上的空钟纹，回来告诉我纹路朝哪边。那才是我需要的答案。", Choices: []DialogueChoiceDef{{ID: "done", Text: "明白。", End: true}}},
			"gate": {ID: "gate", Text: "到过。第一次罗盘疯转，第二次影子比我先回来，第三次我听见门里有人用我的声音叫我。之后我学会了别回答。", Choices: []DialogueChoiceDef{{ID: "note", Text: "这条情报很值钱。", End: true, Effects: []Effect{{Type: "flag", Target: "mara_gate_warning"}, {Type: "relation", Target: "mara", Value: 1}}}}},
		},
	},
	"dlg_orrin": {
		ID: "dlg_orrin", NPCID: "orrin", StartNode: "start", Nodes: map[string]DialogueNodeDef{
			"start": {ID: "start", Text: "旗已经烧了。誓言没有。你若还想举起维尔的名字，我会拦你；你若只是想把门重新关上，我会告诉你最后一条旧路。", Choices: []DialogueChoiceDef{
				{ID: "oath", Text: "我不替王朝复辟，只替活人关门。", NextNode: "oath", Effects: []Effect{{Type: "flag", Target: "orrin_oath"}, {Type: "relation", Target: "orrin", Value: 3}, {Type: "xp", Value: 16}}},
				{ID: "king", Text: "赫里昂值得被记住。", NextNode: "king", Effects: []Effect{{Type: "relation", Target: "orrin", Value: 1}}},
				{ID: "leave", Text: "不做承诺。", End: true},
			}},
			"oath": {ID: "oath", Text: "那就收下这个回答：真正的骑士不守国王，守的是那些国王看不见的人。", Choices: []DialogueChoiceDef{{ID: "done", Text: "向他致意。", End: true, Effects: []Effect{{Type: "item_once", Target: "oath_ring"}}}}},
			"king": {ID: "king", Text: "记住他，但别替他洗去选择。守门三百年很伟大，把整座城市拖进门里也是真的。", Choices: []DialogueChoiceDef{{ID: "done", Text: "我会同时记住两件事。", End: true, Effects: []Effect{{Type: "flag", Target: "balanced_helios_memory"}}}}},
		},
	},
}

func init() {
	Items["veil_thread"] = ItemDef{ID: "veil_thread", Name: "帷面丝线", Type: "trinket", Slot: "trinket", Rarity: "rare", Description: "无脸商会用来缝合破损记忆的银灰细线。敏捷 +1，最大能量 +1。", Value: 44, Icon: "⌁", MaxEnergy: 1, Attributes: AttributeSet{Dexterity: 1}, Special: "撤退检定 +1"}
	Items["windglass_charm"] = ItemDef{ID: "windglass_charm", Name: "风玻璃护符", Type: "trinket", Slot: "trinket", Rarity: "rare", Description: "由镜砂与风骨封成的薄片。感知 +1，意志 +1。", Value: 66, Icon: "◇", Attributes: AttributeSet{Perception: 1, Will: 1}, Special: "荒原调查更稳定"}
	Items["oath_ring"] = ItemDef{ID: "oath_ring", Name: "无旗誓戒", Type: "trinket", Slot: "trinket", Rarity: "legendary", Description: "没有王徽的旧骑士戒。防御 +1，最大生命 +4。", Value: 120, Icon: "○", Defense: 1, MaxHP: 4, Special: "见证无旗骑士最后的誓言"}

	// Assign local portrait artwork to the full combat roster. Several enemies
	// intentionally share an archetype portrait while bosses use bespoke art.
	portraitByEnemy := map[string]string{
		"crown_bearer": "/assets/portraits/enemy_crown_bearer.svg", "gate_heart": "/assets/portraits/enemy_gate_heart.svg",
		"bone_thrall": "/assets/portraits/enemy_skeleton.svg", "memory_knight": "/assets/portraits/enemy_knight.svg",
		"storm_knight": "/assets/portraits/enemy_storm_knight.svg", "ash_raider": "/assets/portraits/enemy_raider.svg",
		"ash_cultist": "/assets/portraits/enemy_cultist.svg", "name_eater": "/assets/portraits/enemy_name_eater.svg",
		"chain_wraith": "/assets/portraits/enemy_chain_wraith.svg", "rose_keeper": "/assets/portraits/enemy_rose.svg",
		"forge_automaton": "/assets/portraits/enemy_automaton.svg", "sky_leech": "/assets/portraits/enemy_leech.svg",
		"drowned_bellman": "/assets/portraits/enemy_drowned.svg", "tomb_crow": "/assets/portraits/enemy_crow.svg",
		"ash_hound": "/assets/portraits/enemy_ash_hound.svg", "grave_spider": "/assets/portraits/enemy_grave_spider.svg",
		"ink_wraith": "/assets/portraits/enemy_ink_wraith.svg", "glass_walker": "/assets/portraits/enemy_glass_walker.svg",
		"dune_wraith": "/assets/portraits/enemy_dune_wraith.svg", "royal_guard": "/assets/portraits/enemy_royal_guard.svg",
		"stone_sentinel": "/assets/portraits/enemy_stone_sentinel.svg",
	}
	for id, enemy := range Enemies {
		if p := portraitByEnemy[id]; p != "" {
			enemy.Portrait = p
		} else if enemy.Portrait == "" {
			enemy.Portrait = "/assets/portraits/enemy_skeleton.svg"
		}
		Enemies[id] = enemy
	}
}

// applyV05Actors adds persistent NPCs to existing rooms without changing the
// deterministic map topology or room IDs used by older saves/tests.
func applyV05Actors(rooms map[string]*Room) {
	appendElement := func(roomID string, el SceneElement) {
		if room := rooms[roomID]; room != nil {
			for i := range room.Elements {
				if room.Elements[i].ID == el.ID {
					room.Elements[i] = el
					return
				}
			}
			room.Elements = append(room.Elements, el)
		}
	}
	appendElement("room_13", SceneElement{ID: "npc_iven", Kind: "npc", Label: "巡夜记录员 · 伊文", Description: "他在整理失踪巡夜队的名单。", Icon: "♟", X: 55, Y: 47, Action: "dialogue", Target: "iven", OneShot: false})
	appendElement("room_14", SceneElement{ID: "masked_merchant", Kind: "npc", Label: "帷面客", Description: "无脸商会的行商正在擦拭银白面具。", Icon: "☻", X: 55, Y: 47, Action: "dialogue", Target: "veil_broker", OneShot: false})
	appendElement("room_07", SceneElement{ID: "chained_shadow", Kind: "npc", Label: "无名囚徒", Description: "锁链后的男人抬起头。他似乎仍然活着。", Icon: "♟", X: 63, Y: 45, Action: "dialogue", Target: "nameless_prisoner", OneShot: false})
	appendElement("room_41", SceneElement{ID: "npc_orrin", Kind: "npc", Label: "无旗骑士 · 奥林", Description: "一套空盔甲坐在熄灭营火旁。里面传出缓慢呼吸。", Icon: "♞", X: 58, Y: 48, Action: "dialogue", Target: "orrin", OneShot: false})
	appendElement("room_42", SceneElement{ID: "npc_mara", Kind: "npc", Label: "逐风斥候 · 玛拉", Description: "她正用一块黑玻璃磨短刀。", Icon: "♟", X: 66, Y: 47, Action: "dialogue", Target: "mara", OneShot: false})
}

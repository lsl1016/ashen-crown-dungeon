package game

// V0.7 turns the existing persistent state into a small living-world simulation.
// All definitions in this file are data. Runtime code interprets them; future AI
// and the local GM editor operate on the same structures instead of editing Go.

var Skills = map[string]SkillDef{
	"warden_smite": {
		ID: "warden_smite", Class: "warden", Name: "铁誓猛击", Icon: "⚔", Cost: 3, Cooldown: 2, MinDistance: 1, MaxDistance: 2,
		HitAttribute: "strength", HitBonus: 2, Description: "沉肩推进并用誓刃重击；命中后进入防御姿态。",
		Effects: []Effect{{Type: "damage", Value: 6, Dice: 8}, {Type: "guard", Value: 1}}, Tags: []string{"melee", "interrupt"},
	},
	"warden_chainbreaker": {
		ID: "warden_chainbreaker", Class: "warden", Name: "断链冲锋", Icon: "▰", Cost: 2, Cooldown: 3, MinDistance: 1, MaxDistance: 3,
		HitAttribute: "strength", HitBonus: 1, Description: "强行拉近到近距，击碎敌人节奏并施加弱化。",
		Effects: []Effect{{Type: "set_distance", Value: 1}, {Type: "damage", Value: 4, Dice: 6}, {Type: "enemy_status", Target: "weakened", Rounds: 2, Stacks: 1}}, Tags: []string{"gap_close", "control"},
	},
	"ranger_pierce": {
		ID: "ranger_pierce", Class: "ranger", Name: "弱点穿刺", Icon: "➵", Cost: 3, Cooldown: 2, MinDistance: 1, MaxDistance: 3,
		HitAttribute: "dexterity", HitBonus: 2, DefenseModifier: -3, Description: "利用感知锁定护甲缝隙，几乎不受高防御影响。",
		Effects: []Effect{{Type: "damage", Value: 7, Dice: 8}}, Tags: []string{"precision", "ranged"},
	},
	"ranger_smoke_arrow": {
		ID: "ranger_smoke_arrow", Class: "ranger", Name: "灰幕箭", Icon: "⌁", Cost: 2, Cooldown: 3, MinDistance: 2, MaxDistance: 3,
		HitAttribute: "dexterity", HitBonus: 1, DefenseModifier: -1, Description: "引爆一支灰幕箭，削弱敌人并借烟势后撤。",
		Effects: []Effect{{Type: "damage", Value: 3, Dice: 4}, {Type: "enemy_status", Target: "weakened", Rounds: 2, Stacks: 1}, {Type: "retreat", Value: 1}}, Tags: []string{"ranged", "control"},
	},
	"seer_burst": {
		ID: "seer_burst", Class: "seer", Name: "余烬爆裂", Icon: "✹", Cost: 3, Cooldown: 2, MinDistance: 1, MaxDistance: 3,
		HitAttribute: "will", HitBonus: 2, DefenseModifier: -2, Description: "引爆目标附近残留的旧王符文。",
		Effects: []Effect{{Type: "damage", Value: 9, Dice: 8}}, Tags: []string{"arcane", "burst"},
	},
	"seer_warding_flare": {
		ID: "seer_warding_flare", Class: "seer", Name: "护幕星火", Icon: "◌", Cost: 2, Cooldown: 3, MinDistance: 1, MaxDistance: 3,
		HitAttribute: "will", HitBonus: 2, DefenseModifier: -1, Description: "以较轻的星火轰击目标，并立即重建符文护幕。",
		Effects: []Effect{{Type: "damage", Value: 3, Dice: 4}, {Type: "shield", Value: 5}, {Type: "energy", Value: 1}}, Tags: []string{"arcane", "defense"},
	},
}

var LivingWorldEvents = map[string]WorldEventDef{
	"grey_plague_tide": {
		ID: "grey_plague_tide", Title: "灰疫潮", Region: "外墓区", TriggerBell: 4, EscalateBell: 7, ResolveFlag: "lost_patrol_reported", Severity: 2, Overlay: "plague",
		Description: "灰疫医馆外开始出现忘记自己姓名的巡夜者。若无人处理，灰雾会吞没外墓区的哨线。",
	},
	"nameless_procession": {
		ID: "nameless_procession", Title: "无名游行", Region: "旧城区", TriggerBell: 7, EscalateBell: 10, ResolveFlag: "prisoner_freed", Severity: 2, Overlay: "echo",
		Description: "没有脸与姓名的旧城居民沿着三百年前的路线游行，反复寻找被抹去的名字。",
	},
	"crown_name_hunt": {
		ID: "crown_name_hunt", Title: "王庭猎名", Region: "内廷", TriggerBell: 10, EscalateBell: 13, ResolveFlag: "defeated_crown_bearer", Severity: 3, Overlay: "blackfire",
		Description: "王庭残响开始主动搜索仍活着的名字，守卫与黑火在内廷重新集结。",
	},
	"ashstorm_front": {
		ID: "ashstorm_front", Title: "灰暴锋线", Region: "雾外荒原", TriggerBell: 8, EscalateBell: 11, ResolveFlag: "mara_hunt_reported", RequireFlag: "act2_unlocked", Severity: 3, Overlay: "storm",
		Description: "空钟骑士把灰暴驱赶成一道移动锋线，商路与逐风营地随时可能被切断。",
	},
}

var NPCSchedules = map[string]NPCScheduleDef{
	"iven": {NPCID: "iven", Entries: []NPCScheduleEntry{
		{FromBell: 1, ToBell: 3, RoomID: "room_13", Activity: "在断桥营火整理巡夜名册"},
		{FromBell: 4, ToBell: 6, RoomID: "room_13", Activity: "组织巡夜者调查灰疫潮"},
		{FromBell: 7, ToBell: 13, RoomID: "room_25", Activity: "在灰疫医馆照看失去姓名的伤员"},
	}},
	"veil_broker": {NPCID: "veil_broker", Entries: []NPCScheduleEntry{
		{FromBell: 1, ToBell: 7, RoomID: "room_14", Activity: "在墓城黑市清点旧王遗物"},
		{FromBell: 8, ToBell: 13, RoomID: "room_14", Activity: "等待墓城东侧石门真正开启"},
		{FromBell: 8, ToBell: 13, RoomID: "room_33", Activity: "把商摊搬到雾外界碑附近", RequireFlag: "act2_unlocked"},
	}},
	"nameless_prisoner": {NPCID: "nameless_prisoner", Entries: []NPCScheduleEntry{{FromBell: 1, ToBell: 13, RoomID: "room_07", Activity: "在旧王囚室等待自己的名字"}}},
	"orrin":             {NPCID: "orrin", Entries: []NPCScheduleEntry{{FromBell: 1, ToBell: 13, RoomID: "room_41", Activity: "守着无旗骑士营最后一面烧毁的旗"}}},
	"mara":              {NPCID: "mara", Entries: []NPCScheduleEntry{{FromBell: 1, ToBell: 13, RoomID: "room_42", Activity: "在逐风营火绘制灰暴变化路线"}}},
}

var QuestGraphs = map[string]QuestGraphDef{
	"lost_patrol": {
		ID: "lost_patrol", Title: "失踪的第三巡夜队",
		Nodes: []QuestGraphNodeDef{
			{ID: "accept", Label: "接受伊文委托", Kind: "start"},
			{ID: "find_captain", Label: "寻找赛勒队长", Kind: "objective"},
			{ID: "return", Label: "带回最后记录", Kind: "turnin"},
			{ID: "reported", Label: "公开真相 · 重建哨线", Kind: "ending"},
			{ID: "sealed", Label: "封存记录 · 维持表面秩序", Kind: "ending"},
		},
		Edges: []QuestGraphEdgeDef{{From: "accept", To: "find_captain"}, {From: "find_captain", To: "return"}, {From: "return", To: "reported", Label: "把真相交给巡夜团"}, {From: "return", To: "sealed", Label: "封存赛勒的记录"}},
	},
	"nameless_prisoner": {
		ID: "nameless_prisoner", Title: "被抹去的名字",
		Nodes: []QuestGraphNodeDef{{ID: "accept", Label: "答应寻找真名", Kind: "start"}, {ID: "recover_name", Label: "从王名密室找回真名", Kind: "objective"}, {ID: "return", Label: "返回旧王囚室", Kind: "turnin"}, {ID: "restored", Label: "归还名字", Kind: "ending"}, {ID: "bound", Label: "把名字重新封进锁链", Kind: "ending"}},
		Edges: []QuestGraphEdgeDef{{From: "accept", To: "recover_name"}, {From: "recover_name", To: "return"}, {From: "return", To: "restored", Label: "归还"}, {From: "return", To: "bound", Label: "重新封印"}},
	},
	"storm_hunt": {
		ID: "storm_hunt", Title: "空钟猎人",
		Nodes: []QuestGraphNodeDef{{ID: "accept", Label: "接受玛拉委托", Kind: "start"}, {ID: "hunt_storm_knight", Label: "调查空钟纹路", Kind: "objective"}, {ID: "return", Label: "返回逐风营火", Kind: "turnin"}, {ID: "route_mapped", Label: "共享安全路线", Kind: "ending"}, {ID: "sold", Label: "把路线卖给帷面客", Kind: "ending"}},
		Edges: []QuestGraphEdgeDef{{From: "accept", To: "hunt_storm_knight"}, {From: "hunt_storm_knight", To: "return"}, {From: "return", To: "route_mapped", Label: "交给逐风者"}, {From: "return", To: "sold", Label: "卖给黑市"}},
	},
}

func init() {
	// Turn the three existing side-quest turn-ins into visible branching quest-graph choices.
	// Also tag the original V0.6 endings with a decision so both branches use
	// the same quest-graph state model.
	if d, ok := Dialogues["dlg_iven"]; ok {
		n := d.Nodes["return"]
		for i := range n.Choices {
			if n.Choices[i].ID == "report" {
				n.Choices[i].Effects = append(n.Choices[i].Effects, Effect{Type: "quest_decision", Target: "lost_patrol", Text: "reported"})
			}
		}
		d.Nodes["return"] = n
		Dialogues["dlg_iven"] = d
	}
	if d, ok := Dialogues["dlg_prisoner"]; ok {
		n := d.Nodes["return"]
		for i := range n.Choices {
			if n.Choices[i].ID == "restore" {
				n.Choices[i].Effects = append(n.Choices[i].Effects, Effect{Type: "quest_decision", Target: "nameless_prisoner", Text: "restored"})
			}
		}
		d.Nodes["return"] = n
		Dialogues["dlg_prisoner"] = d
	}
	if d, ok := Dialogues["dlg_mara"]; ok {
		n := d.Nodes["return"]
		for i := range n.Choices {
			if n.Choices[i].ID == "report" {
				n.Choices[i].Effects = append(n.Choices[i].Effects, Effect{Type: "quest_decision", Target: "storm_hunt", Text: "route_mapped"})
			}
		}
		d.Nodes["return"] = n
		Dialogues["dlg_mara"] = d
	}
	if d, ok := Dialogues["dlg_iven"]; ok {
		n := d.Nodes["return"]
		n.Choices = append(n.Choices, DialogueChoiceDef{ID: "seal", Text: "封存记录，告诉巡夜团赛勒仍然失踪。", End: true, Effects: []Effect{{Type: "flag", Target: "lost_patrol_sealed"}, {Type: "relation", Target: "iven", Value: -2}, {Type: "gold", Value: 50}, {Type: "quest_outcome", Target: "lost_patrol", Text: "sealed"}, {Type: "quest_decision", Target: "lost_patrol", Text: "sealed"}, {Type: "region_state", Target: "外墓区", Text: "哨线维持 · 真相被封存"}}})
		d.Nodes["return"] = n
		Dialogues["dlg_iven"] = d
	}
	if d, ok := Dialogues["dlg_prisoner"]; ok {
		n := d.Nodes["return"]
		n.Choices = append(n.Choices, DialogueChoiceDef{ID: "bind", Text: "把真名重新封入锁链，阻止旧城残响继续扩散。", End: true, Effects: []Effect{{Type: "flag", Target: "prisoner_name_bound"}, {Type: "relation", Target: "nameless_prisoner", Value: -3}, {Type: "item_once", Target: "name_seal"}, {Type: "quest_outcome", Target: "nameless_prisoner", Text: "bound"}, {Type: "quest_decision", Target: "nameless_prisoner", Text: "bound"}, {Type: "region_state", Target: "旧城区", Text: "残响沉寂 · 真名重新被封"}}})
		d.Nodes["return"] = n
		Dialogues["dlg_prisoner"] = d
	}
	if d, ok := Dialogues["dlg_mara"]; ok {
		n := d.Nodes["return"]
		n.Choices = append(n.Choices, DialogueChoiceDef{ID: "keep", Text: "只交一半路线，保留最安全的一条给自己。", End: true, Effects: []Effect{{Type: "flag", Target: "mara_route_withheld"}, {Type: "relation", Target: "mara", Value: -1}, {Type: "gold", Value: 60}, {Type: "quest_outcome", Target: "storm_hunt", Text: "sold"}, {Type: "quest_decision", Target: "storm_hunt", Text: "sold"}, {Type: "region_state", Target: "雾外荒原", Text: "商路开放 · 逐风者仍缺完整路线"}}})
		d.Nodes["return"] = n
		Dialogues["dlg_mara"] = d
	}
}

package game

// V0.6 adds a persistent world simulation layer on top of the V0.5 RPG systems.
// The important rule stays unchanged: code defines capabilities, Run state defines
// the actual world, and future AI may only act through those capabilities.

var Affixes = []AffixDef{
	{ID: "keen", Name: "锋锐", Description: "武器威力 +1。", Power: 1},
	{ID: "graveforged", Name: "墓锻", Description: "武器威力 +2，但只会出现在较稀有战利品上。", Power: 2},
	{ID: "stout", Name: "坚固", Description: "防御 +1。", Defense: 1},
	{ID: "vital", Name: "强韧", Description: "最大生命 +4。", MaxHP: 4},
	{ID: "resonant", Name: "共鸣", Description: "最大能量 +2。", MaxEnergy: 2},
	{ID: "watchful", Name: "守望", Description: "感知 +1。", Attributes: AttributeSet{Perception: 1}},
	{ID: "unyielding", Name: "不屈", Description: "意志 +1。", Attributes: AttributeSet{Will: 1}},
	{ID: "fleet", Name: "逐风", Description: "敏捷 +1。", Attributes: AttributeSet{Dexterity: 1}},
}

func init() {
	// V0.6 reward / drop gear.
	Items["patrol_charm"] = ItemDef{ID: "patrol_charm", Name: "第三巡夜哨坠", Type: "trinket", Slot: "trinket", Rarity: "rare", Description: "赛勒队长的黑铁哨被重新穿成护坠。防御 +1，感知 +1。", Value: 70, Icon: "♩", Defense: 1, Attributes: AttributeSet{Perception: 1}, Special: "证明第三巡夜队并未被遗忘"}
	Items["name_seal"] = ItemDef{ID: "name_seal", Name: "复名银印", Type: "trinket", Slot: "trinket", Rarity: "legendary", Description: "无名囚徒重新获得姓名后留下的银印。意志 +1，最大能量 +2。", Value: 130, Icon: "印", MaxEnergy: 2, Attributes: AttributeSet{Will: 1}, Special: "对残响类敌人的意志检定更稳定"}
	Items["storm_clasp"] = ItemDef{ID: "storm_clasp", Name: "空钟披扣", Type: "armor", Slot: "armor", Rarity: "rare", Description: "风暴骑士胸甲上的空钟纹被改造成披扣。防御 +1，最大生命 +5。", Value: 92, Icon: "♢", Defense: 1, MaxHP: 5, Special: "荒原风暴无法轻易把你推离站位"}
	Items["graveglass_edge"] = ItemDef{ID: "graveglass_edge", Name: "墓玻璃长刃", Type: "weapon", Slot: "weapon", Rarity: "rare", Description: "黑玻璃与墓银焊成的长刃。攻击 +6。", Power: 6, Value: 88, Icon: "╱", Special: "适合作为带随机词条的精英战利品"}
	Items["echo_mail"] = ItemDef{ID: "echo_mail", Name: "残响鳞衣", Type: "armor", Slot: "armor", Rarity: "rare", Description: "每片鳞甲都刻着一个残缺姓名。防御 +2，最大能量 +1。", Value: 96, Icon: "▦", Defense: 2, MaxEnergy: 1}
	Items["star_lens"] = ItemDef{ID: "star_lens", Name: "坠星透镜", Type: "trinket", Slot: "trinket", Rarity: "rare", Description: "从反向观测台拆下的透镜。感知 +1，最大生命 +2。", Value: 86, Icon: "◉", MaxHP: 2, Attributes: AttributeSet{Perception: 1}}

	veil := Shops["veil_market"]
	veil.Items = append(veil.Items, ShopItemDef{ItemID: "graveglass_edge", Price: 96, Stock: 1}, ShopItemDef{ItemID: "echo_mail", Price: 102, Stock: 1})
	Shops["veil_market"] = veil
	wind := Shops["wind_supply"]
	wind.Items = append(wind.Items, ShopItemDef{ItemID: "star_lens", Price: 92, Stock: 1})
	Shops["wind_supply"] = wind

	// Turn the old flat list into a real three-tier class tree and add two capstone
	// choices for every class. Existing saves remain valid because talent IDs are stable.
	for i := range Talents {
		switch Talents[i].ID {
		case "warden_bulwark", "warden_vitality", "ranger_predator", "ranger_quickstep", "ranger_scavenger", "seer_cinder", "seer_reservoir", "seer_ward":
			Talents[i].Tier = 1
		case "warden_execution":
			Talents[i].Tier, Talents[i].RequiredLevel, Talents[i].Requires = 2, 2, []string{"warden_bulwark"}
		case "warden_retaliate":
			Talents[i].Tier, Talents[i].RequiredLevel, Talents[i].Requires = 2, 2, []string{"warden_vitality"}
		case "ranger_bleed":
			Talents[i].Tier, Talents[i].RequiredLevel, Talents[i].Requires = 2, 2, []string{"ranger_predator"}
		case "seer_siphon":
			Talents[i].Tier, Talents[i].RequiredLevel, Talents[i].Requires = 2, 2, []string{"seer_reservoir"}
		}
	}
	Talents = append(Talents,
		TalentDef{ID: "warden_anchor", Class: "warden", Name: "不动锚誓", Description: "防御时额外恢复 2 点生命；位于近距时再获得 +1 防御。", Icon: "⌑", MaxRank: 1, Tier: 3, RequiredLevel: 4, Requires: []string{"warden_vitality", "warden_bulwark"}},
		TalentDef{ID: "warden_march", Class: "warden", Name: "铁墙推进", Description: "推进站位时同时进入防御姿态并恢复 1 点能量。", Icon: "▰", MaxRank: 1, Tier: 3, RequiredLevel: 4, Requires: []string{"warden_execution"}},
		TalentDef{ID: "ranger_longshot", Class: "ranger", Name: "雾外长射", Description: "远距普通攻击命中 +1，伤害 +3。", Icon: "➵", MaxRank: 1, Tier: 2, RequiredLevel: 2, Requires: []string{"ranger_predator"}},
		TalentDef{ID: "ranger_ghostwalk", Class: "ranger", Name: "无痕撤步", Description: "后撤时进入防御姿态；到达远距后额外恢复 1 点能量。", Icon: "⌁", MaxRank: 1, Tier: 3, RequiredLevel: 4, Requires: []string{"ranger_quickstep"}},
		TalentDef{ID: "seer_overchannel", Class: "seer", Name: "星火过载", Description: "远距施放职业技时基础伤害 +4。", Icon: "✷", MaxRank: 1, Tier: 2, RequiredLevel: 2, Requires: []string{"seer_cinder"}},
		TalentDef{ID: "seer_echo_step", Class: "seer", Name: "残响移步", Description: "后撤时恢复 2 点能量，并重新生成一次 3 点符文护幕。", Icon: "◌", MaxRank: 1, Tier: 3, RequiredLevel: 4, Requires: []string{"seer_ward"}},
	)

	// Turn-in nodes make side quests proper chains instead of ending when the objective
	// flips to completed. StartDialogue chooses these nodes from Run state.
	d := Dialogues["dlg_iven"]
	d.Nodes["return"] = DialogueNodeDef{ID: "return", Text: "你带回来的不是赛勒，只是他留下的事实。对巡夜团来说，这已经足够让一个名字从失踪名单上划掉。", Choices: []DialogueChoiceDef{{ID: "report", Text: "把队长最后的记录交给伊文。", End: true, Effects: []Effect{{Type: "flag", Target: "lost_patrol_reported"}, {Type: "relation", Target: "iven", Value: 2}, {Type: "gold", Value: 30}, {Type: "item_once", Target: "patrol_charm"}, {Type: "quest_outcome", Target: "lost_patrol", Text: "reported"}, {Type: "region_state", Target: "外墓区", Text: "巡夜团重新建立哨线"}}}}}
	d.Nodes["after"] = DialogueNodeDef{ID: "after", Text: "外墓区重新亮起两盏巡夜灯。伊文不再把赛勒列在失踪者里，而是列在‘完成值守者’下面。", Choices: []DialogueChoiceDef{{ID: "leave", Text: "点头离开。", End: true}}}
	Dialogues["dlg_iven"] = d

	d = Dialogues["dlg_prisoner"]
	d.Nodes["return"] = DialogueNodeDef{ID: "return", Text: "当你念出从王名密室找到的真名，锁链第一次不是断裂，而是主动松开。他低声重复那个名字，像在重新学会自己。", Choices: []DialogueChoiceDef{{ID: "restore", Text: "把名字还给他。", End: true, Effects: []Effect{{Type: "flag", Target: "prisoner_freed"}, {Type: "relation", Target: "nameless_prisoner", Value: 3}, {Type: "item_once", Target: "name_seal"}, {Type: "quest_outcome", Target: "nameless_prisoner", Text: "restored"}, {Type: "region_state", Target: "旧城区", Text: "被抹去的名字开始回响"}}}}}
	d.Nodes["after"] = DialogueNodeDef{ID: "after", Text: "囚室已经空了。你后来在巡夜营地见过他一次；没人知道他去了哪里，只知道他现在拥有一个能够被记住的名字。", Choices: []DialogueChoiceDef{{ID: "leave", Text: "让他保留自己的新生活。", End: true}}}
	Dialogues["dlg_prisoner"] = d

	d = Dialogues["dlg_mara"]
	d.Nodes["return"] = DialogueNodeDef{ID: "return", Text: "玛拉把你记下的空钟纹路在沙地上重画三遍。第三遍时，她终于笑了一下：这不是骑士徽记，是一条巡逻路线。", Choices: []DialogueChoiceDef{{ID: "report", Text: "把风暴骑士的路线交给她。", End: true, Effects: []Effect{{Type: "flag", Target: "mara_hunt_reported"}, {Type: "relation", Target: "mara", Value: 2}, {Type: "gold", Value: 45}, {Type: "item_once", Target: "storm_clasp"}, {Type: "quest_outcome", Target: "storm_hunt", Text: "route_mapped"}, {Type: "region_state", Target: "雾外荒原", Text: "逐风者掌握风暴巡逻路线"}}}}}
	d.Nodes["after"] = DialogueNodeDef{ID: "after", Text: "逐风营地的地图上已经出现了几条安全路线。玛拉说它们也许明天就会变，但至少今天有人能活着走过去。", Choices: []DialogueChoiceDef{{ID: "trade", Text: "看看今天的补给。", OpenShop: "wind_supply"}, {ID: "leave", Text: "继续赶路。", End: true}}}
	Dialogues["dlg_mara"] = d
}

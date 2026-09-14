package game

var Classes = []ClassDef{
	{ID: "warden", Name: "铁誓守卫", Description: "被逐出圣堂的守墓人。生命与力量突出，擅长正面压制。", HP: 42, Energy: 7, Defense: 13, Attributes: AttributeSet{Strength: 4, Dexterity: 1, Perception: 2, Will: 2}, StarterItem: "iron_blade", SkillName: "铁誓猛击", SkillDescription: "造成重击，并在本轮获得防御姿态。", Icon: "🛡️"},
	{ID: "ranger", Name: "暮影游侠", Description: "在灰雾边境长大的遗迹猎人。敏捷与感知突出，擅长发现陷阱。", HP: 34, Energy: 9, Defense: 14, Attributes: AttributeSet{Strength: 2, Dexterity: 4, Perception: 4, Will: 1}, StarterItem: "moon_dagger", SkillName: "弱点穿刺", SkillDescription: "利用感知寻找弱点，更容易命中高防御敌人。", Icon: "🏹"},
	{ID: "seer", Name: "余烬术士", Description: "能够听见旧王朝残响的符文研究者。意志强大，技能爆发高。", HP: 30, Energy: 12, Defense: 12, Attributes: AttributeSet{Strength: 1, Dexterity: 2, Perception: 3, Will: 5}, StarterItem: "ember_focus", SkillName: "余烬爆裂", SkillDescription: "引爆符文余烬，造成高额意志伤害。", Icon: "🔥"},
}

var Talents = []TalentDef{
	{ID: "warden_bulwark", Class: "warden", Name: "城墙誓言", Description: "防御行动额外获得 2 防御；成功格挡后恢复 1 点能量。", Icon: "▰", MaxRank: 1},
	{ID: "warden_execution", Class: "warden", Name: "处刑重势", Description: "铁誓猛击伤害 +5，并更容易打断敌人的蓄力。", Icon: "⚔", MaxRank: 1},
	{ID: "warden_vitality", Class: "warden", Name: "墓守体魄", Description: "最大生命 +8，解锁时立即恢复 8 点生命。", Icon: "♥", MaxRank: 1},
	{ID: "warden_retaliate", Class: "warden", Name: "铁誓反击", Description: "防御时若敌人近战命中，自动反击 3 点伤害。", Icon: "↶", MaxRank: 1},
	{ID: "ranger_predator", Class: "ranger", Name: "猎杀本能", Description: "普通攻击与职业技在自然骰 19-20 时暴击。", Icon: "◎", MaxRank: 1},
	{ID: "ranger_bleed", Class: "ranger", Name: "裂伤箭路", Description: "弱点穿刺命中后施加 2 回合流血。", Icon: "滴", MaxRank: 1},
	{ID: "ranger_quickstep", Class: "ranger", Name: "暮影步", Description: "每场战斗前两回合防御 +2，撤退 DC -2。", Icon: "➶", MaxRank: 1},
	{ID: "ranger_scavenger", Class: "ranger", Name: "遗迹拾荒者", Description: "战斗掉落率提升，调查类检定获得 +1。", Icon: "◇", MaxRank: 1},
	{ID: "seer_cinder", Class: "seer", Name: "余烬蔓延", Description: "余烬爆裂命中后施加 3 回合燃烧。", Icon: "✹", MaxRank: 1},
	{ID: "seer_reservoir", Class: "seer", Name: "灰潮储能", Description: "最大能量 +3；每场战斗开始时额外恢复 2 点能量。", Icon: "◌", MaxRank: 1},
	{ID: "seer_siphon", Class: "seer", Name: "残响汲取", Description: "职业技命中后恢复 3 点生命。", Icon: "☽", MaxRank: 1},
	{ID: "seer_ward", Class: "seer", Name: "符文护幕", Description: "每场战斗第一次受到伤害时减免 5 点。", Icon: "◈", MaxRank: 1},
}

var Items = map[string]ItemDef{
	"iron_blade":      {ID: "iron_blade", Name: "铁誓长刃", Type: "weapon", Slot: "weapon", Rarity: "common", Description: "守墓人使用的旧式长刃，护手上刻着失落王庭的徽记。攻击 +4。", Power: 4, Value: 25, Icon: "⚔️"},
	"moon_dagger":     {ID: "moon_dagger", Name: "月影短刃", Type: "weapon", Slot: "weapon", Rarity: "common", Description: "刀面几乎不反光。攻击 +3。", Power: 3, Value: 28, Icon: "🗡️"},
	"ember_focus":     {ID: "ember_focus", Name: "余烬法器", Type: "weapon", Slot: "weapon", Rarity: "common", Description: "仍有温度的黑曜石核心。攻击 +3。", Power: 3, Value: 30, Icon: "🔮"},
	"grave_saber":     {ID: "grave_saber", Name: "墓银军刀", Type: "weapon", Slot: "weapon", Rarity: "uncommon", Description: "冷炉旁尚未开刃的墓银刀胚。攻击 +5。", Power: 5, Value: 42, Icon: "⚔"},
	"royal_spear":     {ID: "royal_spear", Name: "王庭短枪", Type: "weapon", Slot: "weapon", Rarity: "rare", Description: "王庭近卫的短枪。攻击 +6。", Power: 6, Value: 58, Icon: "🔱", Special: "对守卫类敌人额外造成 2 点伤害"},
	"glass_knife":     {ID: "glass_knife", Name: "黑玻璃匕首", Type: "weapon", Slot: "weapon", Rarity: "uncommon", Description: "灰烬教团的仪式短刀。攻击 +4。", Power: 4, Value: 36, Icon: "🔪", Special: "暴击时附加流血"},
	"sunken_mace":     {ID: "sunken_mace", Name: "沉钟锤", Type: "weapon", Slot: "weapon", Rarity: "rare", Description: "从沉水钟厅拖出的钟槌。攻击 +7，但非常沉重。", Power: 7, Value: 72, Icon: "🔨", Special: "对蓄力中的敌人伤害 +3"},
	"scribe_wand":     {ID: "scribe_wand", Name: "抄写员灰杖", Type: "weapon", Slot: "weapon", Rarity: "rare", Description: "银线缠绕的短杖。攻击 +5。", Power: 5, Value: 64, Icon: "🪄", Special: "余烬术士职业技伤害 +2"},
	"chainmail":       {ID: "chainmail", Name: "巡夜锁甲", Type: "armor", Slot: "armor", Rarity: "common", Description: "修补过很多次的巡夜团锁甲。防御 +1，最大生命 +4。", Value: 34, Icon: "⛓️", Defense: 1, MaxHP: 4},
	"royal_plate":     {ID: "royal_plate", Name: "王庭残甲", Type: "armor", Slot: "armor", Rarity: "rare", Description: "仍残留誓约刻纹的胸甲。防御 +2，最大生命 +7。", Value: 80, Icon: "🛡", Defense: 2, MaxHP: 7},
	"scribe_coat":     {ID: "scribe_coat", Name: "银线抄写袍", Type: "armor", Slot: "armor", Rarity: "uncommon", Description: "档案官的防护外袍。最大能量 +2，意志 +1。", Value: 52, Icon: "🥼", MaxEnergy: 2, Attributes: AttributeSet{Will: 1}},
	"healing_draught": {ID: "healing_draught", Name: "红叶药剂", Type: "consumable", Rarity: "common", Description: "恢复 14 点生命。", Power: 14, Value: 16, Icon: "🧪"},
	"bandage":         {ID: "bandage", Name: "巡夜绷带", Type: "consumable", Rarity: "common", Description: "恢复 7 点生命。", Power: 7, Value: 8, Icon: "➰"},
	"luminous_tonic":  {ID: "luminous_tonic", Name: "磷光补剂", Type: "consumable", Rarity: "uncommon", Description: "恢复 5 点能量。", Power: 5, Value: 18, Icon: "🧫"},
	"frost_salt":      {ID: "frost_salt", Name: "霜盐", Type: "consumable", Rarity: "uncommon", Description: "战斗中使用，使敌人的下一次伤害降低。", Power: 2, Value: 14, Icon: "❄️"},
	"smoke_bomb":      {ID: "smoke_bomb", Name: "黑灰烟弹", Type: "consumable", Rarity: "uncommon", Description: "战斗中打乱敌人意图，并提高本轮防御。", Power: 3, Value: 20, Icon: "💨"},
	"ember_flask":     {ID: "ember_flask", Name: "灼灰瓶", Type: "consumable", Rarity: "rare", Description: "战斗中施加 3 回合燃烧。", Power: 3, Value: 28, Icon: "🔥"},
	"ashen_charm":     {ID: "ashen_charm", Name: "灰烬护符", Type: "trinket", Slot: "trinket", Rarity: "uncommon", Description: "被烧焦的银质护符。最大生命 +3。", Power: 4, Value: 32, Icon: "📿", MaxHP: 3, Special: "事件伤害 -1"},
	"night_lantern":   {ID: "night_lantern", Name: "巡夜提灯", Type: "trinket", Slot: "trinket", Rarity: "uncommon", Description: "冷蓝磷火照出墙上的细小痕迹。感知 +1。", Value: 38, Icon: "🏮", Attributes: AttributeSet{Perception: 1}},
	"night_badge":     {ID: "night_badge", Name: "旧巡夜徽章", Type: "trinket", Slot: "trinket", Rarity: "rare", Description: "钟声之后仍须守望。防御 +1。", Value: 45, Icon: "✥", Defense: 1},
	"ember_shard":     {ID: "ember_shard", Name: "余烬碎片", Type: "trinket", Slot: "trinket", Rarity: "rare", Description: "来自王冠封印的碎片。最大能量 +1。", Power: 1, Value: 48, Icon: "♦️", MaxEnergy: 1, Special: "燃烧伤害 +1"},
	"silver_goblet":   {ID: "silver_goblet", Name: "巡夜银杯", Type: "trinket", Slot: "trinket", Rarity: "uncommon", Description: "宴厅中唯一没有盛酒的杯子。最大生命 +2。", Value: 30, Icon: "🥂", MaxHP: 2},
	"black_rose":      {ID: "black_rose", Name: "灰玫瑰", Type: "trinket", Slot: "trinket", Rarity: "rare", Description: "花瓣仍有体温。意志 +1。", Value: 42, Icon: "🥀", Attributes: AttributeSet{Will: 1}, Special: "第一次致死伤害后保留 1 HP"},
	"crown_fragment":  {ID: "crown_fragment", Name: "王冠残片", Type: "quest", Rarity: "legendary", Description: "从历史上被删除的封印碎片。", Value: 0, Icon: "♛"},
	"royal_signet":    {ID: "royal_signet", Name: "失落王印", Type: "quest", Rarity: "legendary", Description: "沉眠王朝的王印，也是王冠封印的钥匙之一。", Value: 0, Icon: "💍"},
	"bone_key":        {ID: "bone_key", Name: "骨制钥匙", Type: "quest", Description: "由指骨拼成的古老钥匙。", Value: 0, Icon: "🗝️"},
	"crow_feather":    {ID: "crow_feather", Name: "黑羽信物", Type: "quest", Description: "灰烬教团使用的联络信物。", Value: 8, Icon: "🪶"},
	"old_map":         {ID: "old_map", Name: "褪色墓城图", Type: "quest", Description: "标着一条已经被石匠封死的王室捷径。", Value: 12, Icon: "🗺️"},
	"prison_key":      {ID: "prison_key", Name: "囚室总钥", Type: "quest", Description: "更像身份凭证的总钥。", Value: 15, Icon: "🔑"},
	"star_plate":      {ID: "star_plate", Name: "坠星铜片", Type: "quest", Description: "刻有王族密室坐标。", Value: 24, Icon: "✶"},
	"scribe_seal":     {ID: "scribe_seal", Name: "抄写员火漆章", Type: "quest", Description: "王庭档案官的火漆印章。", Value: 16, Icon: "🔻"},
}

var Enemies = map[string]EnemyDef{
	"bone_thrall":     {ID: "bone_thrall", Name: "骨骸仆从", Description: "由王庭陪葬者的骸骨拼接而成。", HP: 18, Attack: 3, Defense: 11, DamageMin: 3, DamageMax: 7, XP: 14, GoldMin: 4, GoldMax: 9, Icon: "💀", Archetype: "亡灵", Weakness: "重击", IntentProfile: "brute"},
	"grave_spider":    {ID: "grave_spider", Name: "墓穴巨蛛", Description: "腹部泛着惨绿色荧光。", HP: 16, Attack: 4, Defense: 13, DamageMin: 2, DamageMax: 6, XP: 13, GoldMin: 2, GoldMax: 5, Icon: "🕷️", Archetype: "野兽", Weakness: "火焰", IntentProfile: "skirmisher"},
	"ash_hound":       {ID: "ash_hound", Name: "灰烬猎犬", Description: "皮毛下像藏着未熄灭的火星。", HP: 22, Attack: 4, Defense: 12, DamageMin: 4, DamageMax: 8, XP: 17, GoldMin: 5, GoldMax: 10, Icon: "🐺", Archetype: "野兽", Weakness: "流血", IntentProfile: "predator"},
	"royal_guard":     {ID: "royal_guard", Name: "无眠近卫", Description: "王朝最后一夜仍在站岗的重甲亡魂。", HP: 28, Attack: 5, Defense: 14, DamageMin: 5, DamageMax: 9, XP: 24, GoldMin: 8, GoldMax: 14, Icon: "🛡️", Archetype: "守卫", Weakness: "破甲", IntentProfile: "guardian"},
	"tomb_crow":       {ID: "tomb_crow", Name: "噬忆鸦群", Description: "它们啄食尸体最后留下的记忆。", HP: 14, Attack: 4, Defense: 14, DamageMin: 2, DamageMax: 6, XP: 12, GoldMin: 1, GoldMax: 4, Icon: "🐦‍⬛", Archetype: "群体", Weakness: "范围攻击", IntentProfile: "skirmisher"},
	"ash_cultist":     {ID: "ash_cultist", Name: "灰烬教团盗墓者", Description: "仍然活着的人类，比亡者更难预测。", HP: 24, Attack: 5, Defense: 13, DamageMin: 4, DamageMax: 8, XP: 20, GoldMin: 9, GoldMax: 18, Icon: "🥷", Archetype: "人类", Weakness: "控制", IntentProfile: "trickster"},
	"stone_sentinel":  {ID: "stone_sentinel", Name: "王庭石卫", Description: "沉重石像被古代誓约重新驱动。", HP: 34, Attack: 5, Defense: 15, DamageMin: 5, DamageMax: 10, XP: 28, GoldMin: 5, GoldMax: 9, Icon: "🗿", Archetype: "构装", Weakness: "破甲", IntentProfile: "guardian"},
	"chain_wraith":    {ID: "chain_wraith", Name: "锁链怨灵", Description: "被抹去姓名的亡魂。", HP: 25, Attack: 6, Defense: 13, DamageMin: 4, DamageMax: 9, XP: 23, GoldMin: 3, GoldMax: 8, Icon: "⛓️", Archetype: "亡灵", Weakness: "意志", IntentProfile: "hexer"},
	"drowned_bellman": {ID: "drowned_bellman", Name: "溺亡钟守", Description: "浸透黑水的钟守仍抱着铜槌。", HP: 27, Attack: 5, Defense: 12, DamageMin: 5, DamageMax: 10, XP: 25, GoldMin: 5, GoldMax: 11, Icon: "🔔", Archetype: "亡灵", Weakness: "敏捷", IntentProfile: "brute"},
	"ink_wraith":      {ID: "ink_wraith", Name: "墨痕幽灵", Description: "从焚毁档案中爬出的字迹。", HP: 20, Attack: 6, Defense: 14, DamageMin: 3, DamageMax: 8, XP: 21, GoldMin: 2, GoldMax: 6, Icon: "🖋️", Archetype: "残响", Weakness: "火焰", IntentProfile: "hexer"},
	"rose_keeper":     {ID: "rose_keeper", Name: "灰园守枝人", Description: "藤蔓、盔甲与人骨纠缠成的园丁。", HP: 32, Attack: 6, Defense: 14, DamageMin: 5, DamageMax: 10, XP: 29, GoldMin: 6, GoldMax: 12, Icon: "🌿", Archetype: "植物", Weakness: "火焰", IntentProfile: "predator"},
	"forge_automaton": {ID: "forge_automaton", Name: "冷炉铸偶", Description: "停不下来的自动锻造傀儡。", HP: 36, Attack: 6, Defense: 15, DamageMin: 5, DamageMax: 11, XP: 31, GoldMin: 8, GoldMax: 13, Icon: "⚙️", Archetype: "构装", Weakness: "破甲", IntentProfile: "guardian"},
	"memory_knight":   {ID: "memory_knight", Name: "记忆骑士", Description: "挥剑前会重复三百年前战死的最后一句话。", HP: 30, Attack: 7, Defense: 15, DamageMin: 5, DamageMax: 10, XP: 30, GoldMin: 7, GoldMax: 14, Icon: "♞", Archetype: "精英", Weakness: "格挡反击", IntentProfile: "duelist"},
	"name_eater":      {ID: "name_eater", Name: "噬名者", Description: "靠吞食被抹去的名字维持形体。", HP: 38, Attack: 7, Defense: 15, DamageMin: 6, DamageMax: 11, XP: 38, GoldMin: 10, GoldMax: 18, Icon: "◼️", Archetype: "残响", Weakness: "真名", IntentProfile: "hexer"},
	"crown_bearer":    {ID: "crown_bearer", Name: "烬冠执掌者 · 赫里昂", Description: "旧王赫里昂与王冠融为一体，胸腔里燃着永不熄灭的黑火。", HP: 92, Attack: 8, Defense: 16, DamageMin: 6, DamageMax: 13, XP: 120, GoldMin: 35, GoldMax: 50, Icon: "👑", Boss: true, Archetype: "Boss", Weakness: "真名 / 王印", IntentProfile: "boss"},
}

var Lore = map[string]LoreEntry{
	"stele_oath":     {ID: "stele_oath", Title: "断碑上的守门誓", Text: "‘王死则门开，王不死则国亡。愿最后戴冠者以自己的名字，替所有无名者守住门。’末尾签名被凿去。"},
	"scribe_note":    {ID: "scribe_note", Title: "抄写员便签", Text: "封印完成后必须从公开史册中删除王族真名，否则‘门后的东西’能够沿名字寻找封印者。"},
	"prison_names":   {ID: "prison_names", Title: "囚墙姓名", Text: "被刮去的姓名大多属于王庭工匠。最深的一行写着：‘我们没有造王冠，我们造的是锁。’"},
	"queen_epitaph":  {ID: "queen_epitaph", Title: "灰园王后墓铭", Text: "‘若赫里昂忘了为何戴冠，让花提醒他仍曾是人。’"},
	"merchant_crate": {ID: "merchant_crate", Title: "会敲门的木箱", Text: "木箱标签来自三百年前，收件人却写着今天的日期。"},
	"patrol_blanket": {ID: "patrol_blanket", Title: "巡夜团毛毯", Text: "‘不要相信第十三声钟。它从来没有真正响过。’"},
	"camp_notes":     {ID: "camp_notes", Title: "巡夜团近日报告", Text: "过去七天，墓城每夜都会多出一条地图上不存在的路。"},
	"crown_chains":   {ID: "crown_chains", Title: "十三道王冠锁链", Text: "每条锁链都对应一种牺牲：姓名、记忆、血统、王权、睡眠、饥饿、时间、后代、故乡、爱、死亡、历史，以及没有刻出的第十三种。"},
	"wall_marks":     {ID: "wall_marks", Title: "撤退刻痕", Text: "其中一个箭头指向墙内：‘路会在第二次经过时出现。’"},
	"bell_manual":    {ID: "bell_manual", Title: "钟塔值守手册", Text: "十三钟并非计时装置。每一次钟鸣都代表一层封印把记忆转移给下一层。"},
	"physician_note": {ID: "physician_note", Title: "灰疫医师记录", Text: "所谓灰疫没有传染性。患者共同症状是忘记亲友姓名，随后忘记自己的名字。"},
	"aqueduct_marks": {ID: "aqueduct_marks", Title: "引水渠工匠刻痕", Text: "封城前夜，工匠故意把一条引水渠改成通向王庭地下。那不是排水道，而是逃生路。"},
	"court_verdict":  {ID: "court_verdict", Title: "最后一次审判", Text: "维尔最后一个公开判决不是处死叛徒，而是命令所有史官烧掉国王的完整姓名。"},
}
var Events = map[string]EventDef{
	"first_bell": {
		ID: "first_bell", Title: "没有钟的钟声", Description: "你握住断裂钟绳。它明明不再连接任何铜钟，头顶却传来第一声沉重钟鸣。灰雾中短暂出现一支三百年前的送葬队伍。",
		Choices: []ChoiceDef{
			{ID: "follow", Text: "跟随送葬队伍的影子", Check: &CheckDef{Attribute: "will", DC: 11}, OnSuccess: []Effect{{Type: "xp", Value: 8}, {Type: "lore", Target: "stele_oath"}, {Type: "message", Text: "你看见队伍抬着的不是棺材，而是一只巨大的锁匣。"}}, OnFailure: []Effect{{Type: "hp", Value: -3}, {Type: "message", Text: "影子穿过你的身体，带走了一点温度。"}}},
			{ID: "release", Text: "松开钟绳", OnSuccess: []Effect{{Type: "message", Text: "钟声立刻停止，但远处似乎有人回应了一声。"}}},
		},
	},
	"sealed_reliquary": {
		ID: "sealed_reliquary", Title: "封蜡圣匣", Description: "石台上摆着一只被黑色封蜡包裹的银匣。蜡面刻有王庭禁制：未经允许者，将以鲜血付价。",
		Choices: []ChoiceDef{
			{ID: "inspect", Text: "检查禁制的薄弱处", Check: &CheckDef{Attribute: "perception", DC: 12}, OnSuccess: []Effect{{Type: "gold", Value: 18}, {Type: "item", Target: "bone_key"}, {Type: "message", Text: "你在蜡层下找到一枚骨制钥匙与旧金币。"}}, OnFailure: []Effect{{Type: "hp", Value: -6}, {Type: "message", Text: "封蜡突然裂开，细针刺穿手掌。"}}},
			{ID: "force", Text: "直接撬开圣匣", Check: &CheckDef{Attribute: "strength", DC: 13}, OnSuccess: []Effect{{Type: "item", Target: "healing_draught"}, {Type: "gold", Value: 10}}, OnFailure: []Effect{{Type: "hp", Value: -5}}},
			{ID: "leave", Text: "保持封印", OnSuccess: []Effect{{Type: "message", Text: "你压下贪念，把圣匣留在寂静里。"}}},
		},
	},
	"whispering_altar": {
		ID: "whispering_altar", Title: "低语祭坛", Description: "祭坛裂缝里有暗红微光。一个不属于你的声音承诺：献出记忆，换取力量。",
		Choices: []ChoiceDef{
			{ID: "resist", Text: "拒绝低语并解读符文", Check: &CheckDef{Attribute: "will", DC: 12}, OnSuccess: []Effect{{Type: "max_hp", Value: 3}, {Type: "heal", Value: 8}, {Type: "message", Text: "你反向利用祭坛残余力量，身体变得更坚韧。"}}, OnFailure: []Effect{{Type: "hp", Value: -5}}},
			{ID: "offer", Text: "献上 12 枚金币", OnSuccess: []Effect{{Type: "gold", Value: -12}, {Type: "max_hp", Value: 5}, {Type: "heal", Value: 10}}},
			{ID: "leave", Text: "远离祭坛", OnSuccess: []Effect{}},
		},
	},
	"wounded_scout": {
		ID: "wounded_scout", Title: "失踪的斥候", Description: "一名披着巡夜团斗篷的年轻人靠在墙边，腿上的伤已经发黑。他自称伊文，并说王墓深处还有同伴。",
		Choices: []ChoiceDef{
			{ID: "aid", Text: "分给他一瓶药剂", OnSuccess: []Effect{{Type: "consume_item", Target: "healing_draught"}, {Type: "quest", Target: "lost_patrol"}, {Type: "gold", Value: 12}, {Type: "message", Text: "伊文恢复了一些力气，把巡夜团密令交给你。"}}},
			{ID: "bandage", Text: "尝试现场处理伤口", Check: &CheckDef{Attribute: "perception", DC: 11}, OnSuccess: []Effect{{Type: "quest", Target: "lost_patrol"}, {Type: "item", Target: "frost_salt"}}, OnFailure: []Effect{{Type: "quest", Target: "lost_patrol"}}},
			{ID: "leave", Text: "先离开", OnSuccess: []Effect{}},
		},
	},
	"memory_mural": {
		ID: "memory_mural", Title: "记忆壁画", Description: "壁画描绘旧王赫里昂将一顶燃烧的王冠戴上头颅。最后一格被人刻意刮去，只留下三个词：王印、誓约、名字。",
		Choices: []ChoiceDef{
			{ID: "study", Text: "研究残缺文字", Check: &CheckDef{Attribute: "will", DC: 11}, OnSuccess: []Effect{{Type: "flag", Target: "knows_crown_truth"}, {Type: "xp", Value: 12}, {Type: "message", Text: "你意识到王冠并非武器，而是一座用国王灵魂维持的封印。"}}, OnFailure: []Effect{{Type: "message", Text: "符文在视线里互相覆盖，你只感到一阵头痛。"}}},
			{ID: "copy", Text: "拓印壁画", OnSuccess: []Effect{{Type: "xp", Value: 6}, {Type: "flag", Target: "copied_mural"}}},
		},
	},
	"black_well": {
		ID: "black_well", Title: "无底黑井", Description: "井中没有水声。你扔下一粒碎石，许久之后却听见自己的声音从下面说：别让他醒来。",
		Choices: []ChoiceDef{
			{ID: "listen", Text: "继续倾听", Check: &CheckDef{Attribute: "will", DC: 13}, OnSuccess: []Effect{{Type: "xp", Value: 15}, {Type: "flag", Target: "heard_below"}}, OnFailure: []Effect{{Type: "hp", Value: -4}}},
			{ID: "coin", Text: "投入一枚金币", OnSuccess: []Effect{{Type: "gold", Value: -1}, {Type: "heal", Value: 6}}},
			{ID: "leave", Text: "封住井口", OnSuccess: []Effect{}},
		},
	},
	"forgotten_captain": {
		ID: "forgotten_captain", Title: "巡夜团长", Description: "一具尚未完全僵硬的尸体靠在王庭门边，手里紧握着巡夜团徽章。旁边写着：赫里昂不是敌人，王冠才是。",
		Choices: []ChoiceDef{
			{ID: "search", Text: "搜索遗物", OnSuccess: []Effect{{Type: "item", Target: "royal_signet"}, {Type: "quest_complete", Target: "lost_patrol"}, {Type: "xp", Value: 20}}},
			{ID: "salute", Text: "为死者整理遗容", OnSuccess: []Effect{{Type: "max_hp", Value: 2}, {Type: "quest_complete", Target: "lost_patrol"}, {Type: "item", Target: "royal_signet"}}},
		},
	},
	"treasure_cache": {
		ID: "treasure_cache", Title: "陪葬宝库", Description: "一排铜箱仍整齐堆放，只有中央箱没有积灰，仿佛不久前才被碰过。",
		Choices: []ChoiceDef{
			{ID: "careful", Text: "先检查机关", Check: &CheckDef{Attribute: "perception", DC: 12}, OnSuccess: []Effect{{Type: "gold", Value: 26}, {Type: "item", Target: "healing_draught"}}, OnFailure: []Effect{{Type: "hp", Value: -7}, {Type: "gold", Value: 10}}},
			{ID: "grab", Text: "拿走最显眼的金币", OnSuccess: []Effect{{Type: "gold", Value: 16}}},
		},
	},
	"rest_fire": {
		ID: "rest_fire", Title: "不灭余火", Description: "石盆里燃着一簇不会熄灭的蓝白火焰。靠近时，你第一次感到这座墓城仍记得温暖。",
		Choices: []ChoiceDef{
			{ID: "rest", Text: "坐下休整", OnSuccess: []Effect{{Type: "heal", Value: 18}, {Type: "energy", Value: 99}}},
			{ID: "meditate", Text: "凝视旧日影像", Check: &CheckDef{Attribute: "will", DC: 10}, OnSuccess: []Effect{{Type: "xp", Value: 10}, {Type: "energy", Value: 99}}, OnFailure: []Effect{{Type: "heal", Value: 8}}},
		},
	},
	"merchant": {
		ID: "merchant", Title: "无脸商人", Description: "戴瓷白面具的人坐在黑布后。他说自己只向活人收钱，因为死人总会回来付款。",
		Choices: []ChoiceDef{
			{ID: "potion", Text: "购买红叶药剂（16 金）", OnSuccess: []Effect{{Type: "gold", Value: -16}, {Type: "item", Target: "healing_draught"}}},
			{ID: "salt", Text: "购买霜盐（14 金）", OnSuccess: []Effect{{Type: "gold", Value: -14}, {Type: "item", Target: "frost_salt"}}},
			{ID: "charm", Text: "购买灰烬护符（32 金）", OnSuccess: []Effect{{Type: "gold", Value: -32}, {Type: "item", Target: "ashen_charm"}, {Type: "max_hp", Value: 4}}},
			{ID: "rumor", Text: "询问王冠传闻", OnSuccess: []Effect{{Type: "flag", Target: "merchant_rumor"}, {Type: "message", Text: "商人说：‘王冠不会杀人。它只是让戴上它的人永远不能死。’"}}},
			{ID: "leave", Text: "结束交易", OnSuccess: []Effect{}},
		},
	},
	"echo_child": {
		ID: "echo_child", Title: "回声里的孩子", Description: "小女孩的影子从柱后探头。她没有脚步声，只反复问：今天是庆典吗？父王为什么让大家躲起来？",
		Choices: []ChoiceDef{
			{ID: "truth", Text: "告诉她已经过去三百年", Check: &CheckDef{Attribute: "will", DC: 12}, OnSuccess: []Effect{{Type: "xp", Value: 14}, {Type: "item", Target: "ember_shard"}}, OnFailure: []Effect{{Type: "hp", Value: -3}}},
			{ID: "comfort", Text: "告诉她庆典很快开始", OnSuccess: []Effect{{Type: "heal", Value: 5}}},
		},
	},
	"sealed_door": {
		ID: "sealed_door", Title: "封死的王室侧门", Description: "石墙上有一道几乎看不出的门缝。铜牌写着：只有记得道路的人，才有资格走捷径。",
		Choices: []ChoiceDef{
			{ID: "decode", Text: "比对路线刻痕", Check: &CheckDef{Attribute: "perception", DC: 13}, OnSuccess: []Effect{{Type: "item", Target: "old_map"}, {Type: "xp", Value: 10}}, OnFailure: []Effect{}},
			{ID: "force", Text: "推动石门", Check: &CheckDef{Attribute: "strength", DC: 15}, OnSuccess: []Effect{{Type: "gold", Value: 8}}, OnFailure: []Effect{{Type: "hp", Value: -4}}},
		},
	},
	"ash_oracle": {
		ID: "ash_oracle", Title: "灰烬占卜", Description: "裂开的银盆里堆着细灰。你靠近时，灰烬自动排成你的名字，然后改写成另一个陌生名字。",
		Choices: []ChoiceDef{
			{ID: "touch", Text: "触碰陌生名字", Check: &CheckDef{Attribute: "will", DC: 14}, OnSuccess: []Effect{{Type: "flag", Target: "heard_true_name"}, {Type: "xp", Value: 18}}, OnFailure: []Effect{{Type: "hp", Value: -6}}},
			{ID: "scatter", Text: "吹散灰烬", OnSuccess: []Effect{}},
		},
	},
	"royal_banquet": {
		ID: "royal_banquet", Title: "最后的宴席", Description: "长桌食物没有腐烂。半透明宾客举杯庆祝，完全看不见你。",
		Choices: []ChoiceDef{
			{ID: "observe", Text: "观察宾客对话", Check: &CheckDef{Attribute: "perception", DC: 11}, OnSuccess: []Effect{{Type: "gold", Value: 7}, {Type: "flag", Target: "saw_last_feast"}}, OnFailure: []Effect{}},
			{ID: "toast", Text: "举起空杯加入宴席", Check: &CheckDef{Attribute: "will", DC: 12}, OnSuccess: []Effect{{Type: "heal", Value: 10}}, OnFailure: []Effect{{Type: "hp", Value: -5}}},
		},
	},
	"mourning_statue": {
		ID: "mourning_statue", Title: "无面圣像", Description: "圣像双手捧着最后一支燃烧的蜡烛。蜡泪逆着烛身向上流。祭台后写着：‘悼念所有被迫忘记自己姓名的人。’",
		Choices: []ChoiceDef{
			{ID: "light", Text: "用蜡烛点亮其他烛台", Check: &CheckDef{Attribute: "will", DC: 11}, OnSuccess: []Effect{{Type: "heal", Value: 8}, {Type: "xp", Value: 10}, {Type: "message", Text: "数十支蜡烛同时亮起，礼拜堂里响起短暂而真实的祷词。"}}, OnFailure: []Effect{{Type: "hp", Value: -3}}},
			{ID: "search", Text: "检查圣像底座", Check: &CheckDef{Attribute: "perception", DC: 12}, OnSuccess: []Effect{{Type: "item", Target: "scribe_seal"}}, OnFailure: []Effect{}},
		},
	},
	"burned_ledger": {
		ID: "burned_ledger", Title: "焚书馆最后一册", Description: "银线装订的卷宗记录着王城封锁前七日。最后一页只有一句：‘不要让赫里昂知道第十三种代价。’",
		Choices: []ChoiceDef{
			{ID: "read", Text: "完整阅读卷宗", Check: &CheckDef{Attribute: "will", DC: 13}, OnSuccess: []Effect{{Type: "xp", Value: 15}, {Type: "lore", Target: "scribe_note"}, {Type: "flag", Target: "read_last_ledger"}}, OnFailure: []Effect{{Type: "hp", Value: -4}, {Type: "message", Text: "字迹像活物一样钻入视野，你不得不中断阅读。"}}},
			{ID: "seal", Text: "只取走档案官印章", OnSuccess: []Effect{{Type: "item", Target: "scribe_seal"}}},
		},
	},
	"drowned_bell": {
		ID: "drowned_bell", Title: "水下的第二声钟", Description: "黑水中的铜钟没有钟舌。你靠近时，水面却先于铜钟震动，像有什么东西在水下敲击它。",
		Choices: []ChoiceDef{
			{ID: "raise", Text: "把铜钟从水里拖起", Check: &CheckDef{Attribute: "strength", DC: 13}, OnSuccess: []Effect{{Type: "gold", Value: 12}, {Type: "xp", Value: 10}}, OnFailure: []Effect{{Type: "hp", Value: -5}, {Type: "combat", Target: "drowned_bellman"}}},
			{ID: "listen", Text: "把耳朵贴近水面", Check: &CheckDef{Attribute: "perception", DC: 12}, OnSuccess: []Effect{{Type: "flag", Target: "heard_second_bell"}, {Type: "message", Text: "你听见水下有人数到十二，然后故意跳过了十三。"}}, OnFailure: []Effect{{Type: "combat", Target: "drowned_bellman"}}},
		},
	},
	"chained_prisoner": {
		ID: "chained_prisoner", Title: "没有名字的囚徒", Description: "铁门后坐着一个仍会呼吸的人形。他说自己记不起名字，只记得三百年来一直有人每天来问同一个问题：王冠在哪里？",
		Choices: []ChoiceDef{
			{ID: "free", Text: "尝试打开锁链", Check: &CheckDef{Attribute: "dexterity", DC: 13}, OnSuccess: []Effect{{Type: "item", Target: "prison_key"}, {Type: "quest", Target: "nameless_prisoner"}, {Type: "message", Text: "囚徒没有离开，只把囚室总钥交给你：‘先找到我的名字。’"}}, OnFailure: []Effect{{Type: "combat", Target: "chain_wraith"}}},
			{ID: "question", Text: "追问王冠", Check: &CheckDef{Attribute: "will", DC: 12}, OnSuccess: []Effect{{Type: "flag", Target: "prisoner_warning"}, {Type: "message", Text: "囚徒说：‘王冠不在王座上。王座上的只是它想让你看见的形状。’"}}, OnFailure: []Effect{}},
		},
	},
	"ashen_garden": {
		ID: "ashen_garden", Title: "会记住体温的花", Description: "灰玫瑰在你靠近时转向你。花心浮现一张女性面孔，似乎正在等待一句迟到了三百年的答复。",
		Choices: []ChoiceDef{
			{ID: "pick", Text: "摘下一朵灰玫瑰", Check: &CheckDef{Attribute: "will", DC: 12}, OnSuccess: []Effect{{Type: "item", Target: "black_rose"}, {Type: "heal", Value: 6}}, OnFailure: []Effect{{Type: "combat", Target: "rose_keeper"}}},
			{ID: "listen", Text: "不碰花，只听它说话", OnSuccess: []Effect{{Type: "lore", Target: "queen_epitaph"}, {Type: "flag", Target: "queen_message"}}},
		},
	},
	"bone_oracle": {
		ID: "bone_oracle", Title: "千骨占卜", Description: "中央骨塔开始自行旋转。不同年份、不同身份的骨骼相互碰撞，竟拼出一句完整的话：‘活人，你缺少一枚印。’",
		Choices: []ChoiceDef{
			{ID: "ask_crown", Text: "询问王冠", Check: &CheckDef{Attribute: "perception", DC: 12}, OnSuccess: []Effect{{Type: "xp", Value: 12}, {Type: "flag", Target: "bones_warn_crown"}}, OnFailure: []Effect{{Type: "combat", Target: "bone_thrall"}}},
			{ID: "ask_exit", Text: "询问安全道路", OnSuccess: []Effect{{Type: "item", Target: "old_map"}}},
		},
	},
	"frozen_forge": {
		ID: "frozen_forge", Title: "最后一次锻造", Description: "冷炉在你靠近后自行喷出蓝火。铁砧上浮现一段工匠记忆：他们打造的不是王冠，而是十三道锁链。",
		Choices: []ChoiceDef{
			{ID: "rekindle", Text: "重新拉动风箱", Check: &CheckDef{Attribute: "strength", DC: 12}, OnSuccess: []Effect{{Type: "item", Target: "grave_saber"}, {Type: "xp", Value: 10}}, OnFailure: []Effect{{Type: "combat", Target: "forge_automaton"}}},
			{ID: "study", Text: "研究铁砧上的锁链模具", Check: &CheckDef{Attribute: "perception", DC: 13}, OnSuccess: []Effect{{Type: "flag", Target: "knows_thirteen_chains"}, {Type: "lore", Target: "crown_chains"}}, OnFailure: []Effect{}},
		},
	},
	"star_machine": {
		ID: "star_machine", Title: "地下观星者", Description: "星仪运转时，铜环投射出的不是星座，而是墓城地图。一个从未出现过的房间在最上层闪烁。",
		Choices: []ChoiceDef{
			{ID: "align", Text: "校准缺失星位", Check: &CheckDef{Attribute: "perception", DC: 13}, OnSuccess: []Effect{{Type: "item", Target: "star_plate"}, {Type: "unlock", Target: "room_24"}, {Type: "message", Text: "石墙深处传来齿轮移动声。地图上出现了‘王名密室’。"}}, OnFailure: []Effect{{Type: "hp", Value: -4}}},
			{ID: "watch", Text: "观察王座上方的星轨", Check: &CheckDef{Attribute: "will", DC: 12}, OnSuccess: []Effect{{Type: "flag", Target: "saw_false_crown"}, {Type: "message", Text: "星轨显示王座上方的王冠只是投影，真正封印位于更深处。"}}, OnFailure: []Effect{}},
		},
	},
	"name_mirror": {
		ID: "name_mirror", Title: "王名镜", Description: "镜面没有映出你的脸。它先写出你的名字，又在下面缓慢浮现‘赫里昂·维尔’。这是你第一次看到旧王完整真名。",
		Choices: []ChoiceDef{
			{ID: "speak", Text: "念出赫里昂的完整真名", Check: &CheckDef{Attribute: "will", DC: 14}, OnSuccess: []Effect{{Type: "flag", Target: "heard_true_name"}, {Type: "flag", Target: "knows_crown_truth"}, {Type: "quest_complete", Target: "nameless_prisoner"}, {Type: "xp", Value: 25}, {Type: "message", Text: "整个墓城短暂安静。某个困了三百年的意识第一次听见有人真正叫出自己。"}}, OnFailure: []Effect{{Type: "combat", Target: "name_eater"}}},
			{ID: "break", Text: "打碎镜子", Check: &CheckDef{Attribute: "strength", DC: 13}, OnSuccess: []Effect{{Type: "item", Target: "crown_fragment"}}, OnFailure: []Effect{{Type: "hp", Value: -6}}},
		},
	},
	"memory_mirror": {
		ID: "memory_mirror", Title: "残留的最后一刻", Description: "墙上浮现一段人的最后记忆：一名士兵把钥匙塞进砖缝，然后转身面对涌来的黑影。",
		Choices: []ChoiceDef{
			{ID: "search", Text: "按记忆位置检查墙缝", Check: &CheckDef{Attribute: "perception", DC: 11}, OnSuccess: []Effect{{Type: "item", Target: "bone_key"}, {Type: "gold", Value: 6}}, OnFailure: []Effect{}},
			{ID: "honor", Text: "向残影致意", OnSuccess: []Effect{{Type: "xp", Value: 6}}},
		},
	},
	"cultist_cache": {
		ID: "cultist_cache", Title: "教团临时据点", Description: "新鲜脚印通向一只黑布箱。里面有仪式刀、信件和一张标注王庭入口的草图。",
		Choices: []ChoiceDef{
			{ID: "search", Text: "搜查箱子", Check: &CheckDef{Attribute: "perception", DC: 11}, OnSuccess: []Effect{{Type: "item", Target: "glass_knife"}, {Type: "item", Target: "crow_feather"}, {Type: "gold", Value: 12}}, OnFailure: []Effect{{Type: "combat", Target: "ash_cultist"}}},
			{ID: "burn", Text: "烧掉教团物资", OnSuccess: []Effect{{Type: "xp", Value: 9}, {Type: "flag", Target: "burned_cult_cache"}}},
		},
	},
	"royal_confessor": {
		ID: "royal_confessor", Title: "告解室里的声音", Description: "封死的告解室里有人低声重复：‘我告诉国王，封印需要第十三种代价。我不该告诉他。’",
		Choices: []ChoiceDef{
			{ID: "ask", Text: "追问第十三种代价", Check: &CheckDef{Attribute: "will", DC: 14}, OnSuccess: []Effect{{Type: "flag", Target: "knows_last_price"}, {Type: "message", Text: "声音回答：‘遗忘。不是让别人忘记国王，而是让国王最终忘记自己为何守门。’"}}, OnFailure: []Effect{{Type: "hp", Value: -5}}},
			{ID: "leave", Text: "不再追问", OnSuccess: []Effect{}},
		},
	},
	"sealed_armory": {
		ID: "sealed_armory", Title: "王庭军械封条", Description: "武器架上挂着一把仍保持镜面光泽的短枪。封条写着：仅供仍记得誓言者取用。",
		Choices: []ChoiceDef{
			{ID: "oath", Text: "复述断碑上的守门誓", Check: &CheckDef{Attribute: "will", DC: 11}, OnSuccess: []Effect{{Type: "item", Target: "royal_spear"}}, OnFailure: []Effect{{Type: "combat", Target: "memory_knight"}}},
			{ID: "leave", Text: "保留封条", OnSuccess: []Effect{}},
		},
	},
	"crow_nest": {
		ID: "crow_nest", Title: "噬忆鸦巢", Description: "天花板裂缝里塞满黑羽与写过字的纸片。每张纸都只剩姓名，其他内容被啄得干干净净。",
		Choices: []ChoiceDef{
			{ID: "search", Text: "翻找纸片", Check: &CheckDef{Attribute: "dexterity", DC: 12}, OnSuccess: []Effect{{Type: "item", Target: "crow_feather"}, {Type: "xp", Value: 8}}, OnFailure: []Effect{{Type: "combat", Target: "tomb_crow"}}},
			{ID: "leave", Text: "不惊动鸦群", OnSuccess: []Effect{}},
		},
	},
}

func World() WorldMeta {
	return WorldMeta{
		Title:    "灰烬王冠：沉眠墓城",
		Subtitle: "Ashen Crown · Tomb of Veyr",
		Premise:  "三百年前，维尔王朝在一夜之间从史书中消失。末代国王赫里昂把自己的灵魂铸进灰烬王冠，让整座王都成为封印。如今封印开始重新排列墓城道路，你收到一封没有署名的信：在第十三次钟鸣之前进入王墓，查明王冠究竟是王权、牢笼，还是一把门闩。",
		Era:      "灰历 317 年，第十三钟前夜",
		Factions: []string{"巡夜团：监视墓城异动的边境守卫", "灰烬教团：相信王冠藏有永生秘密的结社", "无脸商会：在灾厄之间做生意的中立商人", "旧王残响：被王冠困在最后记忆中的王庭亡者"},
		Truths:   []string{"墓城亡者不是复活，而是记忆循环。", "赫里昂既是暴君，也是维持封印的人。", "王印、真名和十三道锁链会改变最终结局。", "墓城道路会主动重排，仿佛封印本身拥有意志。"},
		Classes:  Classes,
		Talents:  Talents,
		Items:    Items,
		ToolHint: "V0.3 继续保持代码定义能力、数据定义世界。AI 接入后应通过 Game Tool 调用现有交互、战斗、世界与内容能力，而不是修改源码。",
	}
}

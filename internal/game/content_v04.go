package game

// V0.4 内容包：在沉眠墓城之后加入第二幕“雾外荒原”。
// 仍然保持数据驱动：这里扩充物品、敌人、事件与地图模板，不改变 Runtime 的核心协议。
func init() {
	Items["void_compass"] = ItemDef{ID: "void_compass", Name: "虚空罗盘", Type: "quest", Rarity: "legendary", Description: "指针不朝北，只指向封印最薄的位置。", Icon: "✥"}
	Items["starsteel_blade"] = ItemDef{ID: "starsteel_blade", Name: "坠星刃", Type: "weapon", Slot: "weapon", Rarity: "legendary", Description: "由陨铁与墓银重新熔铸的长刃。攻击 +9。", Power: 9, Value: 120, Icon: "✦", Special: "攻击蓄力与构装敌人时额外造成 3 点伤害"}
	Items["wasteland_cloak"] = ItemDef{ID: "wasteland_cloak", Name: "灰风披衣", Type: "armor", Slot: "armor", Rarity: "rare", Description: "荒原旅队使用的重披衣。防御 +2，最大生命 +5，敏捷 +1。", Value: 86, Icon: "◩", Defense: 2, MaxHP: 5, Attributes: AttributeSet{Dexterity: 1}}
	Items["glass_charm"] = ItemDef{ID: "glass_charm", Name: "镜砂护符", Type: "trinket", Slot: "trinket", Rarity: "rare", Description: "由会记录声音的黑玻璃打磨而成。感知 +1，意志 +1。", Value: 74, Icon: "◇", Attributes: AttributeSet{Perception: 1, Will: 1}, Special: "事件检定失败时仍保留一部分线索"}
	Items["storm_phial"] = ItemDef{ID: "storm_phial", Name: "瓶装灰暴", Type: "consumable", Rarity: "rare", Description: "战斗中释放压缩灰暴，使敌人弱化 2 回合。", Power: 2, Value: 36, Icon: "☁"}
	Items["star_salve"] = ItemDef{ID: "star_salve", Name: "星屑药膏", Type: "consumable", Rarity: "rare", Description: "恢复 20 点生命。", Power: 20, Value: 34, Icon: "✚"}
	Items["wind_token"] = ItemDef{ID: "wind_token", Name: "逐风骨牌", Type: "quest", Rarity: "uncommon", Description: "旧荒原商队用来证明安全路线的骨牌。", Value: 12, Icon: "▱"}
	Items["gate_shard"] = ItemDef{ID: "gate_shard", Name: "门后黑晶", Type: "trinket", Slot: "trinket", Rarity: "legendary", Description: "最终封印脱落的黑晶。最大能量 +2，意志 +1。", Value: 140, Icon: "◆", MaxEnergy: 2, Attributes: AttributeSet{Will: 1}}

	Enemies["ash_raider"] = EnemyDef{ID: "ash_raider", Name: "灰原劫行者", Description: "披着风蚀甲片的荒原掠夺者，会根据你的架势改变攻击。", HP: 34, Attack: 7, Defense: 14, DamageMin: 5, DamageMax: 11, XP: 34, GoldMin: 12, GoldMax: 22, Icon: "⚔", Archetype: "人类", Weakness: "流血", IntentProfile: "duelist"}
	Enemies["glass_walker"] = EnemyDef{ID: "glass_walker", Name: "镜砂行尸", Description: "身体被黑玻璃穿透的旅人残骸，移动时会复制你的动作。", HP: 31, Attack: 7, Defense: 15, DamageMin: 5, DamageMax: 10, XP: 35, GoldMin: 6, GoldMax: 13, Icon: "◇", Archetype: "残响", Weakness: "重击", IntentProfile: "guardian"}
	Enemies["dune_wraith"] = EnemyDef{ID: "dune_wraith", Name: "骨风怨灵", Description: "只在风停的瞬间显形。", HP: 27, Attack: 8, Defense: 15, DamageMin: 4, DamageMax: 11, XP: 33, GoldMin: 5, GoldMax: 12, Icon: "〰", Archetype: "亡灵", Weakness: "意志", IntentProfile: "hexer"}
	Enemies["sky_leech"] = EnemyDef{ID: "sky_leech", Name: "坠星吸髓虫", Description: "从陨坑里爬出的巨大寄生体，背部闪烁着冷白星光。", HP: 39, Attack: 7, Defense: 13, DamageMin: 6, DamageMax: 12, XP: 40, GoldMin: 4, GoldMax: 9, Icon: "✹", Archetype: "异兽", Weakness: "火焰", IntentProfile: "predator"}
	Enemies["storm_knight"] = EnemyDef{ID: "storm_knight", Name: "无旗风骑士", Description: "最后一支护送王室撤离的骑士团残响。", HP: 48, Attack: 8, Defense: 16, DamageMin: 6, DamageMax: 13, XP: 52, GoldMin: 14, GoldMax: 24, Icon: "♞", Archetype: "精英", Weakness: "格挡反击", IntentProfile: "duelist"}
	Enemies["gate_heart"] = EnemyDef{ID: "gate_heart", Name: "门后之心 · 厄涅玛", Description: "它不是神，也不是怪物，而是三百年前所有封印都试图阻止进入世界的那一次心跳。", HP: 156, Attack: 10, Defense: 17, DamageMin: 8, DamageMax: 16, XP: 220, GoldMin: 60, GoldMax: 90, Icon: "◉", Boss: true, Archetype: "最终 Boss", Weakness: "真名 / 坠星 / 封印", IntentProfile: "boss"}

	Lore["waste_boundary"] = LoreEntry{ID: "waste_boundary", Title: "雾外界碑", Text: "界碑背面不是王朝文字，而是巡夜团百年前追加的警告：墓城并非封印本体，它只是一枚钉子。灰雾之外还有更多钉子。"}
	Lore["caravan_log"] = LoreEntry{ID: "caravan_log", Title: "最后一支荒原商队", Text: "商队记录写道：第十三钟之后，星星开始从错误的方向升起。所有罗盘都改为指向同一个裂谷。"}
	Lore["glass_memory"] = LoreEntry{ID: "glass_memory", Title: "镜砂里的声音", Text: "黑玻璃会记录经过它附近的最后一句话。数千块碎片都保存着同一句：‘门在呼吸。’"}
	Lore["star_map"] = LoreEntry{ID: "star_map", Title: "反向星图", Text: "星图不是观察天空，而是在记录门后之物投射到天空中的影子。投影越清晰，封印越薄。"}
	Lore["gate_prayer"] = LoreEntry{ID: "gate_prayer", Title: "无门者祷文", Text: "‘若王冠失效，沿灰风向东。若天空出现第二轮黑日，不要祈祷。那意味着它已经能听见。’"}

	Events["waste_beacon"] = EventDef{ID: "waste_beacon", Title: "熄灭的界火", Description: "墓城背后的石门开启后，你第一次看见真正天空。荒原界火已经熄灭，火盆里却留着一枚仍温热的巡夜骨牌。", Choices: []ChoiceDef{
		{ID: "rekindle", Text: "重新点燃界火", Check: &CheckDef{Attribute: "will", DC: 13}, OnSuccess: []Effect{{Type: "item", Target: "wind_token"}, {Type: "lore", Target: "waste_boundary"}, {Type: "xp", Value: 14}, {Type: "message", Text: "界火变成冷蓝色，远处三处风塔同时回应。"}}, OnFailure: []Effect{{Type: "hp", Value: -5}, {Type: "message", Text: "火焰短暂组成一只眼睛，然后在你掌心熄灭。"}}},
		{ID: "take", Text: "取走骨牌继续前进", OnSuccess: []Effect{{Type: "item", Target: "wind_token"}}},
	}}
	Events["fallen_caravan"] = EventDef{ID: "fallen_caravan", Title: "停止在昨天的商队", Description: "一支商队停在风里。篷布在动，人却全部保持着前一秒的姿势，仿佛时间只忘了继续作用在他们身上。", Choices: []ChoiceDef{
		{ID: "touch", Text: "触碰领队的肩膀", Check: &CheckDef{Attribute: "will", DC: 14}, OnSuccess: []Effect{{Type: "lore", Target: "caravan_log"}, {Type: "item", Target: "wasteland_cloak"}, {Type: "xp", Value: 18}}, OnFailure: []Effect{{Type: "combat", Target: "dune_wraith"}}},
		{ID: "search", Text: "只搜索货车", Check: &CheckDef{Attribute: "perception", DC: 12}, OnSuccess: []Effect{{Type: "item", Target: "star_salve"}, {Type: "gold", Value: 18}}, OnFailure: []Effect{{Type: "hp", Value: -4}}},
	}}
	Events["glass_echo"] = EventDef{ID: "glass_echo", Title: "会复述未来的镜砂", Description: "整片沼地铺着黑色玻璃。你尚未开口，脚边碎片已经用你的声音说出：‘不要走进裂谷。’", Choices: []ChoiceDef{
		{ID: "listen", Text: "寻找声音最清晰的碎片", Check: &CheckDef{Attribute: "perception", DC: 14}, OnSuccess: []Effect{{Type: "lore", Target: "glass_memory"}, {Type: "item", Target: "glass_charm"}, {Type: "xp", Value: 16}}, OnFailure: []Effect{{Type: "combat", Target: "glass_walker"}}},
		{ID: "break", Text: "砸碎发声的玻璃", Check: &CheckDef{Attribute: "strength", DC: 13}, OnSuccess: []Effect{{Type: "item", Target: "storm_phial"}}, OnFailure: []Effect{{Type: "hp", Value: -6}}},
	}}
	Events["star_wound"] = EventDef{ID: "star_wound", Title: "仍在坠落的陨星", Description: "陨坑中央悬着一块没有落地的陨铁。周围所有沙粒都在缓慢向上升。", Choices: []ChoiceDef{
		{ID: "forge", Text: "用王庭武器触碰陨铁", Check: &CheckDef{Attribute: "will", DC: 14}, OnSuccess: []Effect{{Type: "item", Target: "starsteel_blade"}, {Type: "xp", Value: 24}, {Type: "flag", Target: "touched_star_wound"}}, OnFailure: []Effect{{Type: "combat", Target: "sky_leech"}}},
		{ID: "observe", Text: "观察引力异常", Check: &CheckDef{Attribute: "perception", DC: 13}, OnSuccess: []Effect{{Type: "lore", Target: "star_map"}, {Type: "xp", Value: 12}}, OnFailure: []Effect{}},
	}}
	Events["reverse_observatory"] = EventDef{ID: "reverse_observatory", Title: "反向观测台", Description: "破碎透镜朝向地下。透过它，你看见一扇不存在于地表的巨大黑门，门缝里有东西在规律收缩。", Choices: []ChoiceDef{
		{ID: "align", Text: "用坠星铜片校准透镜", Check: &CheckDef{Attribute: "perception", DC: 14}, OnSuccess: []Effect{{Type: "item", Target: "void_compass"}, {Type: "lore", Target: "star_map"}, {Type: "xp", Value: 20}, {Type: "message", Text: "罗盘指针立刻转向荒原最东端。那里没有山，只有一道竖直的黑线。"}}, OnFailure: []Effect{{Type: "combat", Target: "storm_knight"}}},
		{ID: "mark", Text: "记录黑门位置", OnSuccess: []Effect{{Type: "flag", Target: "knows_gate_location"}, {Type: "xp", Value: 8}}},
	}}
	Events["wind_oracle"] = EventDef{ID: "wind_oracle", Title: "没有神像的风祠", Description: "祠堂中央只有十三根风骨。风吹过时，它们拼出一个问题：你愿意为了记住世界而被世界遗忘吗？", Choices: []ChoiceDef{
		{ID: "answer", Text: "回答：愿意", Check: &CheckDef{Attribute: "will", DC: 15}, OnSuccess: []Effect{{Type: "max_hp", Value: 4}, {Type: "energy", Value: 99}, {Type: "lore", Target: "gate_prayer"}, {Type: "flag", Target: "accepted_wind_oath"}}, OnFailure: []Effect{{Type: "hp", Value: -7}}},
		{ID: "leave", Text: "保持沉默", OnSuccess: []Effect{{Type: "heal", Value: 8}}},
	}}
	Events["black_gate"] = EventDef{ID: "black_gate", Title: "封印之外的门", Description: "黑门没有门扇。所谓入口只是空气里一道比黑暗更黑的竖线。虚空罗盘在这里停止转动。", Choices: []ChoiceDef{
		{ID: "open", Text: "让罗盘与封印共振", Check: &CheckDef{Attribute: "will", DC: 15}, OnSuccess: []Effect{{Type: "unlock", Target: "room_44"}, {Type: "flag", Target: "opened_outer_gate"}, {Type: "xp", Value: 30}, {Type: "message", Text: "竖线缓慢张开。你第一次听见门后的心跳。"}}, OnFailure: []Effect{{Type: "combat", Target: "storm_knight"}}},
		{ID: "listen", Text: "先听门后的声音", OnSuccess: []Effect{{Type: "lore", Target: "gate_prayer"}, {Type: "flag", Target: "heard_gate_heart"}}},
	}}

	roomLayout = append(roomLayout,
		roomTemplate{10, 3, "frontier"}, roomTemplate{11, 3, "ashfield"}, roomTemplate{12, 3, "caravan"}, roomTemplate{13, 3, "glassmarsh"},
		roomTemplate{14, 3, "crater"}, roomTemplate{15, 3, "starwatch"}, roomTemplate{12, 2, "windshrine"}, roomTemplate{13, 2, "meteor"},
		roomTemplate{14, 2, "ruins"}, roomTemplate{11, 4, "windcamp"}, roomTemplate{15, 2, "outergate"}, roomTemplate{16, 3, "finalboss"},
	)
	baseEdges = append(baseEdges,
		[2]int{8, 32}, [2]int{32, 33}, [2]int{33, 34}, [2]int{34, 35}, [2]int{35, 36}, [2]int{36, 37},
		[2]int{34, 38}, [2]int{38, 39}, [2]int{39, 35}, [2]int{33, 41}, [2]int{41, 34}, [2]int{36, 40},
		[2]int{40, 42}, [2]int{37, 42}, [2]int{42, 43},
	)
	roomNames["frontier"] = []string{"雾外界碑"}
	roomNames["ashfield"] = []string{"灰风平原"}
	roomNames["caravan"] = []string{"停滞商队"}
	roomNames["glassmarsh"] = []string{"镜砂洼地"}
	roomNames["crater"] = []string{"坠星裂谷"}
	roomNames["starwatch"] = []string{"反向观测台"}
	roomNames["windshrine"] = []string{"无像风祠"}
	roomNames["meteor"] = []string{"陨铁采掘场"}
	roomNames["ruins"] = []string{"无旗骑士营"}
	roomNames["windcamp"] = []string{"逐风营火"}
	roomNames["outergate"] = []string{"外封印黑门"}
	roomNames["finalboss"] = []string{"门后心室"}
}

var v04SceneByType = map[string]string{
	"frontier": "frontier", "ashfield": "ashfield", "caravan": "caravan", "glassmarsh": "glassmarsh",
	"crater": "crater", "starwatch": "starwatch", "windshrine": "windshrine", "meteor": "meteor",
	"ruins": "ruins", "windcamp": "windcamp", "outergate": "outergate", "finalboss": "voidheart",
}

var v04DescriptionByType = map[string]string{
	"frontier":   "墓城石门之后不是出口，而是一片终年刮着灰风的荒原。远处界火尽数熄灭，天空第一次真正出现在你头顶。",
	"ashfield":   "风把灰烬削成一层层波纹。地面偶尔露出王朝道路的石桩，证明这里曾有成百上千人从墓城向东逃亡。",
	"caravan":    "篷车、驮兽与旅人全部停在同一个动作里。只有篷布仍被风吹动，营火甚至保持着燃烧的形状。",
	"glassmarsh": "黑色玻璃像结冰湖面覆盖低地。每踩碎一片，都有陌生人的最后一句话从脚下响起。",
	"crater":     "巨大的陨坑撕开荒原。坑底陨铁仍悬在半空，碎石与灰尘违反重力向它缓慢上升。",
	"starwatch":  "废弃观测台的所有透镜都朝向地底。它们不是用来看星星，而是在监视一扇没有人愿意命名的门。",
	"windshrine": "十三根空心骨柱围成祠堂。这里没有神像，风本身就是唯一被供奉的东西。",
	"meteor":     "古老采掘架伸进陨铁层。工具停留在三百年前最后一次敲击的位置，金属切面仍散发冷光。",
	"ruins":      "折断的旗杆围着一片石垒。无旗骑士的盔甲整齐坐在营火旁，像只是等待下一次出发命令。",
	"windcamp":   "一圈白石挡住大部分灰风。营火低而稳定，是荒原上难得能重新整理装备与呼吸的安全地。",
	"outergate":  "荒原尽头竖着一道细长黑线。越靠近，周围声音越慢，最后只剩与你心跳不同步的另一声脉动。",
	"finalboss":  "黑门之后没有房间，只有由锁链、记忆与星光围出的巨大心室。中央那颗黑色心脏第一次因为你的到来加速跳动。",
}

func v04ElementsForRoom(t, roomID string) ([]SceneElement, bool) {
	var els []SceneElement
	switch t {
	case "frontier":
		els = []SceneElement{{ID: "waste_boundary_stone", Kind: "lore", Label: "雾外界碑", Description: "界碑背面有后人追加的文字。", Icon: "▤", X: 28, Y: 48, Action: "lore", Target: "waste_boundary", OneShot: true}, {ID: "dead_beacon", Kind: "event", Label: "熄灭界火", Description: "火盆里有一枚温热骨牌。", Icon: "✦", X: 68, Y: 42, Action: "event", Target: "waste_beacon", OneShot: true}}
	case "ashfield":
		els = []SceneElement{{ID: "ash_tracks", Kind: "object", Label: "逆风脚印", Description: "脚印朝着风吹来的方向延伸。", Icon: "⌁", X: 34, Y: 64, Action: "message", Target: "这些脚印没有主人。每当风停，它们就会向前多出现一步。", OneShot: true}, {ID: "raider_cache", Kind: "loot", Label: "劫行者藏包", Description: "压在路标石下面。", Icon: "⬡", X: 75, Y: 69, Action: "loot", Target: "storm_phial", Check: &CheckDef{Attribute: "perception", DC: 13}, OnFailure: []Effect{{Type: "combat", Target: "ash_raider"}}, OneShot: true}}
	case "caravan":
		els = []SceneElement{{ID: "frozen_caravan", Kind: "npc", Label: "停滞领队", Description: "他的眼睛仍能轻微转动。", Icon: "♟", X: 52, Y: 44, Action: "event", Target: "fallen_caravan", OneShot: true}, {ID: "route_case", Kind: "loot", Label: "路线匣", Description: "需要逐风骨牌打开。", Icon: "▣", X: 78, Y: 67, Action: "loot", Target: "star_salve", RequiresItem: "wind_token", OneShot: true}}
	case "glassmarsh":
		els = []SceneElement{{ID: "speaking_shard", Kind: "event", Label: "说话的玻璃", Description: "它正在用你的声音说话。", Icon: "◇", X: 48, Y: 42, Action: "event", Target: "glass_echo", OneShot: true}, {ID: "mirror_pool", Kind: "object", Label: "镜砂浅池", Description: "倒影比你的动作慢半拍。", Icon: "◌", X: 72, Y: 72, Action: "message", Target: "倒影最终停住，没有继续模仿你。它抬起手，指向东方裂谷。", OneShot: true}}
	case "crater":
		els = []SceneElement{{ID: "star_wound_core", Kind: "event", Label: "悬浮陨铁", Description: "周围碎石正在向上坠落。", Icon: "✹", X: 53, Y: 38, Action: "event", Target: "star_wound", OneShot: true}, {ID: "star_salve_cache", Kind: "loot", Label: "巡夜医箱", Description: "箱盖被陨石热量焊死。", Icon: "✚", X: 22, Y: 70, Action: "loot", Target: "star_salve", Check: &CheckDef{Attribute: "strength", DC: 13}, OneShot: true}}
	case "starwatch":
		els = []SceneElement{{ID: "reverse_scope", Kind: "event", Label: "反向主镜", Description: "透镜朝着地下。", Icon: "✺", X: 52, Y: 35, Action: "event", Target: "reverse_observatory", OneShot: true}, {ID: "compass_case", Kind: "loot", Label: "星盘仪器匣", Description: "锁舌会根据星光角度自行移动。", Icon: "✥", X: 75, Y: 66, Action: "loot", Target: "void_compass", Check: &CheckDef{Attribute: "perception", DC: 14}, OnFailure: []Effect{{Type: "hp", Value: -3}}, OneShot: true}, {ID: "observer_note", Kind: "lore", Label: "观测员刻痕", Description: "墙上重复刻着同一组坐标。", Icon: "〆", X: 24, Y: 62, Action: "lore", Target: "star_map", OneShot: true}}
	case "windshrine":
		els = []SceneElement{{ID: "wind_bones", Kind: "event", Label: "十三根风骨", Description: "风穿过骨孔时像有人说话。", Icon: "〰", X: 50, Y: 40, Action: "event", Target: "wind_oracle", OneShot: true}, {ID: "wind_rest", Kind: "rest", Label: "背风石龛", Description: "可以短暂休整。", Icon: "✦", X: 74, Y: 70, Action: "rest", Value: 16, OneShot: false}}
	case "meteor":
		els = []SceneElement{{ID: "starsteel_blank", Kind: "loot", Label: "陨铁刀胚", Description: "仍有足够材料再锻一件武器。", Icon: "╱", X: 56, Y: 52, Action: "loot", Target: "starsteel_blade", RequiresFlag: "touched_star_wound", OneShot: true}, {ID: "miner_cache", Kind: "loot", Label: "采掘者衣箱", Description: "里面叠着防灰风披衣。", Icon: "▥", X: 26, Y: 68, Action: "loot", Target: "wasteland_cloak", OneShot: true}}
	case "ruins":
		els = []SceneElement{{ID: "empty_banner", Kind: "lore", Label: "被割走的军旗", Description: "旗杆上仍系着王庭撤退命令。", Icon: "⚑", X: 25, Y: 43, Action: "message", Target: "命令最后一句写着：‘不要回王都。若国王成功，我们将再也不记得为何必须逃。’", OneShot: true}}
	case "windcamp":
		els = []SceneElement{{ID: "wind_fire", Kind: "rest", Label: "逐风营火", Description: "火焰紧贴地面却不会熄灭。", Icon: "✦", X: 52, Y: 63, Action: "rest", Value: 22, OneShot: false}, {ID: "storm_satchel", Kind: "loot", Label: "风暴行囊", Description: "里面还有一瓶压缩灰暴。", Icon: "⬡", X: 78, Y: 68, Action: "loot", Target: "storm_phial", OneShot: true}}
	case "outergate":
		els = []SceneElement{{ID: "outer_gate_line", Kind: "secret", Label: "黑门竖线", Description: "需要虚空罗盘与一次意志检定才能让门真正张开。", Icon: "┃", X: 52, Y: 38, Action: "unlock", Target: "room_44", RequiresItem: "void_compass", Check: &CheckDef{Attribute: "will", DC: 15}, OnSuccess: []Effect{{Type: "flag", Target: "opened_outer_gate"}, {Type: "xp", Value: 30}, {Type: "message", Text: "虚空罗盘停止转动，黑线缓慢张开。门后的心跳第一次清晰到盖过风声。"}}, OnFailure: []Effect{{Type: "combat", Target: "storm_knight"}}, OneShot: true}, {ID: "gate_whisper", Kind: "event", Label: "门缝低语", Description: "在真正开门前，也许可以先听听另一边。", Icon: "◉", X: 72, Y: 52, Action: "event", Target: "black_gate", RequiresItem: "void_compass", OneShot: true}, {ID: "gate_prayer_wall", Kind: "lore", Label: "风蚀祷文", Description: "石壁上只剩最后一段。", Icon: "▤", X: 24, Y: 67, Action: "lore", Target: "gate_prayer", OneShot: true}}
	case "finalboss":
		els = []SceneElement{{ID: "heart_chain", Kind: "lore", Label: "断裂封印链", Description: "这条链的另一端原本连接着赫里昂。", Icon: "⛓", X: 20, Y: 42, Action: "message", Target: "你终于理解：赫里昂并非在封印墓城，他只是替更古老的封印承担了其中一条锁链。", OneShot: true}}
	default:
		return nil, false
	}
	return els, true
}

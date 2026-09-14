package game

import (
	"fmt"
	"math/rand"
)

type roomTemplate struct {
	X    int
	Y    int
	Type string
}

// V0.3 在 V0.2 墓城骨架上扩展“上层废都”与条件道路。
// 主要地标稳定存在，房间内容与捷径仍由 Seed 决定，便于存档、回放和后续 Agent 操纵。
var roomLayout = []roomTemplate{
	{0, 3, "entrance"}, {1, 3, "event"}, {2, 3, "combat"}, {3, 3, "chapel"},
	{4, 3, "library"}, {5, 3, "flooded"}, {6, 3, "prison"}, {7, 3, "garden"}, {8, 3, "boss"},
	{2, 2, "ossuary"}, {2, 1, "treasure"}, {3, 4, "forge"}, {4, 4, "npc"}, {5, 4, "merchant"},
	{5, 2, "observatory"}, {5, 1, "shrine"}, {6, 2, "banquet"}, {6, 4, "rest"},
	{1, 4, "combat"}, {3, 2, "event"}, {4, 2, "combat"}, {7, 4, "npc"}, {7, 2, "treasure"}, {4, 0, "secret"},
	{0, 4, "infirmary"}, {0, 2, "gatehouse"}, {1, 2, "aqueduct"}, {1, 1, "bridge"},
	{6, 1, "court"}, {8, 2, "belltower"}, {8, 1, "reliquary"}, {8, 4, "mausoleum"},
}

var baseEdges = [][2]int{
	{0, 1}, {1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 6}, {6, 7}, {7, 8},
	{2, 9}, {9, 10}, {3, 11}, {11, 12}, {12, 13}, {5, 14}, {14, 15},
	{6, 16}, {16, 22}, {6, 17}, {1, 18}, {3, 19}, {19, 20}, {7, 21}, {7, 22}, {14, 23},
	{0, 24}, {24, 18}, {0, 25}, {25, 26}, {26, 9}, {26, 27}, {27, 10},
	{15, 28}, {28, 16}, {28, 22}, {22, 29}, {29, 30}, {30, 8}, {21, 31}, {31, 8},
}

var roomNames = map[string][]string{
	"entrance":    {"断碑入口"},
	"combat":      {"碎骨长廊", "无灯守卫厅", "殉葬者甬道", "裂甲回廊", "鸦羽阶梯", "沉钟岗"},
	"event":       {"回声室", "褪色壁画厅", "沉默井庭", "封蜡档案室", "镜厅残廊"},
	"treasure":    {"陪葬宝库", "银棺侧室", "王室军械库"},
	"npc":         {"巡夜团残营", "石门哨所", "断桥营火"},
	"merchant":    {"无脸集市"},
	"rest":        {"余火圣龛"},
	"shrine":      {"低语祭坛"},
	"chapel":      {"悼亡礼拜堂"},
	"library":     {"焚书馆"},
	"flooded":     {"沉水钟厅"},
	"prison":      {"铁誓囚室"},
	"garden":      {"灰玫瑰庭"},
	"ossuary":     {"千骨堂"},
	"forge":       {"冷炉工坊"},
	"observatory": {"坠星观测台"},
	"banquet":     {"无终宴厅"},
	"secret":      {"王名密室"},
	"infirmary":   {"灰疫医馆"},
	"gatehouse":   {"折冠门楼"},
	"aqueduct":    {"黑水引渠"},
	"bridge":      {"断月桥"},
	"court":       {"灰烬审判庭"},
	"belltower":   {"无钟之塔"},
	"reliquary":   {"王室封藏间"},
	"mausoleum":   {"白石王陵"},
	"boss":        {"烬冠王庭"},
}

func GenerateDungeon(seed int64) (map[string]*Room, []Edge) {
	r := rand.New(rand.NewSource(seed))
	rooms := map[string]*Room{}
	for i, p := range roomLayout {
		id := fmt.Sprintf("room_%02d", i+1)
		typ := p.Type
		names := roomNames[typ]
		name := names[r.Intn(len(names))]
		room := &Room{
			ID:          id,
			Name:        name,
			Type:        typ,
			Scene:       sceneForType(typ, r.Intn(4)),
			Description: roomDescription(typ),
			Zone:        zoneForX(p.X),
			X:           p.X,
			Y:           p.Y,
			Elements:    elementsForRoom(typ, id, seed),
		}
		rooms[id] = room
	}

	rooms["room_01"].Discovered = true
	rooms["room_01"].Visited = true
	rooms["room_01"].Resolved = true
	rooms["room_01"].Description = "你穿过被藤蔓覆盖的断碑，身后的灰雾立刻合拢。面前不再只是一条墓道，而是一座埋在地下的旧王都：钟声、流水、炉火与宴席的残响仍在不同城区里重复。"

	// 条件区域必须通过场景机关或关键物品开启。
	rooms["room_24"].Locked = true
	rooms["room_24"].Discovered = false
	rooms["room_31"].Locked = true
	rooms["room_31"].Discovered = false
	rooms["room_32"].Locked = true
	rooms["room_32"].Discovered = false
	// 第二幕在击败赫里昂之后开启；最终心室还需要在荒原中找到开启方式。
	rooms["room_33"].Locked = true
	rooms["room_33"].Discovered = false
	rooms["room_44"].Locked = true
	rooms["room_44"].Discovered = false

	edges := make([]Edge, 0, len(baseEdges)+3)
	for _, pair := range baseEdges {
		edges = append(edges, Edge{From: fmt.Sprintf("room_%02d", pair[0]+1), To: fmt.Sprintf("room_%02d", pair[1]+1)})
	}
	// Seed 决定额外捷径，同一 Seed 永远一致。
	if r.Intn(2) == 1 {
		edges = append(edges, Edge{From: "room_11", To: "room_20"})
	}
	if r.Intn(3) == 2 {
		edges = append(edges, Edge{From: "room_18", To: "room_22"})
	}
	if r.Intn(4) == 3 {
		edges = append(edges, Edge{From: "room_13", To: "room_18"})
	}
	return rooms, edges
}

func sceneForType(t string, variant int) string {
	if scene, ok := v04SceneByType[t]; ok {
		return scene
	}
	switch t {
	case "entrance":
		return "entrance"
	case "boss":
		return "throne"
	case "merchant":
		return "market"
	case "rest":
		return "ember"
	case "shrine":
		return "altar"
	case "treasure":
		if variant%2 == 0 {
			return "vault"
		}
		return "armory"
	case "npc":
		return "camp"
	case "chapel":
		return "chapel"
	case "library":
		return "library"
	case "flooded":
		return "flooded"
	case "prison":
		return "prison"
	case "garden":
		return "garden"
	case "ossuary":
		return "ossuary"
	case "forge":
		return "forge"
	case "observatory":
		return "observatory"
	case "banquet":
		return "banquet"
	case "secret":
		return "secret"
	case "infirmary":
		return "infirmary"
	case "gatehouse":
		return "gatehouse"
	case "aqueduct":
		return "aqueduct"
	case "bridge":
		return "bridge"
	case "court":
		return "court"
	case "belltower":
		return "belltower"
	case "reliquary":
		return "reliquary"
	case "mausoleum":
		return "mausoleum"
	case "combat":
		if variant%2 == 0 {
			return "hall"
		}
		return "crypt"
	default:
		if variant%3 == 0 {
			return "well"
		}
		if variant%2 == 0 {
			return "hall"
		}
		return "crypt"
	}
}

func zoneForX(x int) string {
	switch {
	case x >= 15:
		return "荒原东界"
	case x >= 10:
		return "雾外荒原"
	case x <= 1:
		return "外墓区"
	case x <= 3:
		return "旧城区"
	case x <= 5:
		return "王城中环"
	case x <= 7:
		return "内廷"
	default:
		return "烬冠核心"
	}
}

func roomDescription(t string) string {
	if desc, ok := v04DescriptionByType[t]; ok {
		return desc
	}
	switch t {
	case "combat":
		return "地上拖痕凌乱，墙角还有刚被碰落的灰尘。你能感觉到有什么东西正沿着柱影移动。"
	case "event":
		return "这间墓室保存着旧王朝最后几天的残片。文字、回声与物件都可能让三百年前的夜晚重新发生一次。"
	case "treasure":
		return "铜锁、陪葬箱和武器架在黑暗里反出暗淡光泽。维尔人的财富从来不会单独出现，陷阱也是陪葬品的一部分。"
	case "npc":
		return "这里有人类近期活动过。余温、绷带和被踩乱的灰尘说明墓城里不只有亡者。"
	case "merchant":
		return "黑布、铜秤、成排面具与一盏白灯组成了荒谬的小市场。商品看起来都是真的，商人是不是活人则很难判断。"
	case "rest":
		return "蓝白余火稳定燃烧，周围没有任何怪物敢靠近。这里是墓城少数允许人重新整理呼吸的地方。"
	case "shrine":
		return "祭坛裂缝中不断冒出暗红微光。只要站得足够近，你就会听见某个声音准确叫出你的名字。"
	case "chapel":
		return "倾倒的长椅面向一座没有面孔的圣像。数百支蜡烛早已熄灭，却仍有一支在祭台下微微发亮。"
	case "library":
		return "高耸书架像烧焦的树林。大部分卷册只剩灰，但一些用银线缝制的档案拒绝被火焰毁掉。"
	case "flooded":
		return "黑水没过脚踝，倒映着倒悬的钟架。每走一步，远处都会响起并非由你造成的第二次水声。"
	case "prison":
		return "锈蚀铁门把走廊切成一格格黑暗。墙上刻满姓名，其中很多名字被后来者用刀刮掉。"
	case "garden":
		return "地下庭院里生长着没有阳光也能盛开的灰色玫瑰。藤蔓缠住石像，花瓣落地时会化成温热的灰。"
	case "ossuary":
		return "成千上万块骨骼按照身份与年份码进壁龛。中央骨塔里偶尔传出像风铃一样清脆的碰撞声。"
	case "forge":
		return "巨大的锻炉已经冷却三百年，风箱却仍会自己起伏。铁砧边散落着未完成的王庭兵器。"
	case "observatory":
		return "一座青铜星仪占满大厅。这里看不见天空，仪器却仍准确追踪着墓城上方早已改变的群星。"
	case "banquet":
		return "宴桌上的食物没有腐烂，半透明宾客每隔几十秒就会重新举杯。没有人注意到庆典早已结束。"
	case "secret":
		return "没有地图记录这里。墙面刻满被抹去的王族真名，中央只有一张石桌和一面从未积灰的镜子。"
	case "infirmary":
		return "褪色帘幕把病床分成狭窄格子。所有病历最后一栏写的都不是死因，而是‘忘记姓名’。"
	case "gatehouse":
		return "巨大的落闸卡在半空，绞盘链条已经锈死。门楼两侧仍摆着最后一次封城时的武器箱。"
	case "aqueduct":
		return "黑水在石渠下缓慢流动，墙上维护刻痕指向一条本不该通向王庭的支渠。"
	case "bridge":
		return "月牙形石桥从中间断裂，下方不是河，而是看不见底的灰雾。另一侧挂着一根仍可使用的旧绳。"
	case "court":
		return "审判席和旁听席都覆盖着灰。中央石台上保留着维尔王朝最后一次公开判决。"
	case "belltower":
		return "塔里没有铜钟，只有十三组空荡钟架。最上层的机械仍随着远处钟声自行转动。"
	case "reliquary":
		return "这里封存的不是陪葬品，而是王族主动放弃的身份象征。每个封盒都有不同的钥匙规则。"
	case "mausoleum":
		return "白石棺椁排列得像一支沉默军队。最深处的王族石棺没有姓名，只有一枚钥匙孔。"
	case "boss":
		return "王座立在巨大圆厅尽头。灰烬王冠悬浮在王座上方，像一轮熄灭的太阳。十三道锁链从王冠延伸进地下。"
	default:
		return "墓城的石壁仍在低声回响。"
	}
}

func elementsForRoom(t, roomID string, seed int64) []SceneElement {
	// X/Y 是场景画布百分比坐标。元素本身只是数据，前端负责渲染热点。
	common := []SceneElement{}
	if extra, ok := v04ElementsForRoom(t, roomID); ok {
		common = extra
		ambientPick := int(hash64(fmt.Sprintf("ambient-v04:%d:%s", seed, roomID)) % 3)
		ambients := []SceneElement{
			{ID: "v04_ambient_wind", Kind: "object", Label: "灰风涡流", Description: "风在这里围着看不见的物体打转。", Icon: "〰", X: 15, Y: 62, Action: "message", Target: "你伸手探进涡流，短暂触到一片冰冷表面，像门的背面。", OneShot: true},
			{ID: "v04_ambient_star", Kind: "object", Label: "逆坠星屑", Description: "几粒星屑缓慢向天空坠落。", Icon: "·", X: 84, Y: 58, Action: "message", Target: "星屑越过你的指尖继续上升。荒原的重力正在被某个更大的东西拉扯。", OneShot: true},
			{ID: "v04_ambient_bone", Kind: "object", Label: "风蚀骨牌", Description: "刻着已经消失的商队编号。", Icon: "▱", X: 18, Y: 72, Action: "message", Target: "骨牌上的路线终点不是城市，而是一个被反复涂黑的圆。", OneShot: true},
		}
		common = append(common, ambients[ambientPick])
		return common
	}
	switch t {
	case "entrance":
		common = []SceneElement{
			{ID: "broken_stele", Kind: "lore", Label: "断裂王碑", Description: "碑文只剩最后四行。", Icon: "▤", X: 23, Y: 46, Action: "lore", Target: "stele_oath", OneShot: true},
			{ID: "bell_rope", Kind: "object", Label: "断钟绳", Description: "拉动它也许会惊醒什么。", Icon: "⌁", X: 68, Y: 32, Action: "event", Target: "first_bell", OneShot: true},
			{ID: "traveler_pack", Kind: "loot", Label: "遗失行囊", Description: "半埋在灰里的旅行者行囊。", Icon: "⬡", X: 77, Y: 70, Action: "loot", Target: "bandage", OneShot: true},
		}
	case "chapel":
		common = []SceneElement{
			{ID: "saint_statue", Kind: "event", Label: "无面圣像", Description: "圣像的手心放着一枚没有熄灭的蜡烛。", Icon: "✟", X: 50, Y: 30, Action: "event", Target: "mourning_statue", OneShot: true},
			{ID: "votive_box", Kind: "loot", Label: "奉献箱", Description: "锁已经锈断。", Icon: "▣", X: 25, Y: 69, Action: "gold", Value: 9, OneShot: true},
		}
	case "library":
		common = []SceneElement{
			{ID: "silver_ledger", Kind: "event", Label: "银线卷宗", Description: "唯一没有烧毁的档案。", Icon: "▤", X: 68, Y: 38, Action: "event", Target: "burned_ledger", OneShot: true},
			{ID: "charred_shelf", Kind: "lore", Label: "焦黑书架", Description: "夹层里藏着王庭抄写员的便签。", Icon: "⌑", X: 22, Y: 42, Action: "lore", Target: "scribe_note", OneShot: true},
		}
	case "flooded":
		common = []SceneElement{
			{ID: "sunken_bell", Kind: "event", Label: "沉水铜钟", Description: "钟体在水下轻微震动。", Icon: "◉", X: 58, Y: 64, Action: "event", Target: "drowned_bell", OneShot: true},
			{ID: "water_glint", Kind: "loot", Label: "水下微光", Description: "碎石缝里卡着一个小瓶。", Icon: "✦", X: 29, Y: 76, Action: "loot", Target: "luminous_tonic", OneShot: true},
		}
	case "prison":
		common = []SceneElement{
			{ID: "chained_shadow", Kind: "npc", Label: "锁链后的影子", Description: "它听见脚步后停止了呼吸。", Icon: "♟", X: 63, Y: 45, Action: "event", Target: "chained_prisoner", OneShot: true},
			{ID: "scratched_names", Kind: "lore", Label: "刮去的姓名", Description: "反复出现一个相同姓氏。", Icon: "〆", X: 25, Y: 40, Action: "lore", Target: "prison_names", OneShot: true},
		}
	case "garden":
		common = []SceneElement{
			{ID: "grey_rose", Kind: "event", Label: "灰玫瑰", Description: "花瓣落下时仍有体温。", Icon: "❀", X: 35, Y: 66, Action: "event", Target: "ashen_garden", OneShot: true},
			{ID: "stone_queen", Kind: "lore", Label: "王后石像", Description: "底座上的铭文被植物遮住。", Icon: "♕", X: 70, Y: 35, Action: "lore", Target: "queen_epitaph", OneShot: true},
		}
	case "ossuary":
		common = []SceneElement{
			{ID: "bone_tower", Kind: "event", Label: "中央骨塔", Description: "骨骼正在轻轻互相敲击。", Icon: "☷", X: 50, Y: 38, Action: "event", Target: "bone_oracle", OneShot: true},
			{ID: "officer_skull", Kind: "loot", Label: "军官壁龛", Description: "一枚徽章卡在颅骨下。", Icon: "◇", X: 22, Y: 57, Action: "loot", Target: "night_badge", OneShot: true},
		}
	case "forge":
		common = []SceneElement{
			{ID: "cold_furnace", Kind: "event", Label: "冷却王炉", Description: "风箱仍像肺一样起伏。", Icon: "♨", X: 66, Y: 48, Action: "event", Target: "frozen_forge", OneShot: true},
			{ID: "weapon_blank", Kind: "loot", Label: "未完成刀胚", Description: "材质仍然异常坚硬。", Icon: "╱", X: 30, Y: 67, Action: "loot", Target: "grave_saber", OneShot: true},
		}
	case "observatory":
		common = []SceneElement{
			{ID: "star_machine", Kind: "event", Label: "青铜星仪", Description: "星轨与墓城道路一一重合。", Icon: "✺", X: 50, Y: 39, Action: "event", Target: "star_machine", OneShot: true},
			{ID: "hidden_alignment", Kind: "secret", Label: "缺失的星位", Description: "调整星仪也许能显示一条不存在的路。", Icon: "✧", X: 73, Y: 61, Action: "unlock", Target: "room_24", OneShot: true},
		}
	case "banquet":
		common = []SceneElement{
			{ID: "empty_chair", Kind: "event", Label: "空着的主宾席", Description: "所有幽灵都会避开这张椅子。", Icon: "♜", X: 57, Y: 55, Action: "event", Target: "royal_banquet", OneShot: true},
			{ID: "silver_goblet", Kind: "loot", Label: "银杯", Description: "杯底刻着巡夜团的旧徽记。", Icon: "∪", X: 30, Y: 63, Action: "loot", Target: "silver_goblet", OneShot: true},
		}
	case "merchant":
		common = []SceneElement{
			{ID: "masked_merchant", Kind: "npc", Label: "无脸商人", Description: "他在等你开口。", Icon: "☻", X: 55, Y: 47, Action: "event", Target: "merchant", OneShot: false},
			{ID: "strange_crate", Kind: "object", Label: "封口木箱", Description: "箱子偶尔会从里面敲两下。", Icon: "▣", X: 25, Y: 68, Action: "lore", Target: "merchant_crate", OneShot: true},
		}
	case "rest":
		common = []SceneElement{
			{ID: "ember_fire", Kind: "rest", Label: "不灭余火", Description: "可以在这里恢复状态。", Icon: "✦", X: 51, Y: 62, Action: "rest", Value: 18, OneShot: false},
			{ID: "old_blanket", Kind: "lore", Label: "巡夜团毛毯", Description: "边角缝着一个名字。", Icon: "▱", X: 25, Y: 72, Action: "lore", Target: "patrol_blanket", OneShot: true},
		}
	case "shrine":
		common = []SceneElement{
			{ID: "red_altar", Kind: "event", Label: "低语祭坛", Description: "靠近后声音更清晰了。", Icon: "△", X: 50, Y: 42, Action: "event", Target: "whispering_altar", OneShot: true},
			{ID: "ash_bowl", Kind: "lore", Label: "银灰盆", Description: "灰烬会自动排列成文字。", Icon: "◒", X: 73, Y: 66, Action: "event", Target: "ash_oracle", OneShot: true},
		}
	case "treasure":
		common = []SceneElement{
			{ID: "central_chest", Kind: "loot", Label: "中央铜箱", Description: "灰尘最少，也最可疑。", Icon: "▰", X: 51, Y: 61, Action: "event", Target: "treasure_cache", OneShot: true},
			{ID: "weapon_rack", Kind: "loot", Label: "陪葬武器架", Description: "仍有一件武器没有彻底锈坏。", Icon: "⚔", X: 72, Y: 43, Action: "loot", Target: "royal_spear", OneShot: true},
		}
	case "secret":
		common = []SceneElement{
			{ID: "name_mirror", Kind: "event", Label: "真名镜", Description: "镜面不映照你的脸，只映照名字。", Icon: "◈", X: 50, Y: 38, Action: "event", Target: "name_mirror", OneShot: true},
			{ID: "sealed_drawer", Kind: "loot", Label: "王室暗格", Description: "里面放着从历史上被删除的物件。", Icon: "▤", X: 28, Y: 69, Action: "loot", Target: "crown_fragment", OneShot: true},
		}
	case "npc":
		common = []SceneElement{
			{ID: "camp_notes", Kind: "lore", Label: "营地笔记", Description: "最近几天的巡夜记录。", Icon: "▤", X: 27, Y: 66, Action: "lore", Target: "camp_notes", OneShot: true},
			{ID: "supply_box", Kind: "loot", Label: "补给箱", Description: "还剩一些能用的东西。", Icon: "▣", X: 72, Y: 67, Action: "loot", Target: "healing_draught", OneShot: true},
		}
	case "infirmary":
		common = []SceneElement{
			{ID: "physician_ledger", Kind: "lore", Label: "医师病历", Description: "病历把‘失忆’当成最危险的症状。", Icon: "✚", X: 30, Y: 40, Action: "lore", Target: "physician_note", OneShot: true},
			{ID: "sealed_medicine", Kind: "loot", Label: "档案官药柜", Description: "火漆锁只接受王庭抄写员印章。", Icon: "▣", X: 70, Y: 55, Action: "loot", Target: "luminous_tonic", RequiresItem: "scribe_seal", OneShot: true},
			{ID: "surgical_roll", Kind: "loot", Label: "医师工具卷", Description: "还有几卷干燥绷带。", Icon: "➰", X: 55, Y: 72, Action: "loot", Target: "bandage", OneShot: true},
		}
	case "gatehouse":
		common = []SceneElement{
			{ID: "rusted_winch", Kind: "object", Label: "落闸绞盘", Description: "锈死的绞盘也许还能被强行转动。", Icon: "⚙", X: 60, Y: 43, Action: "flag", Target: "gate_winch_open", Check: &CheckDef{Attribute: "strength", DC: 12}, OnSuccess: []Effect{{Type: "xp", Value: 8}}, OnFailure: []Effect{{Type: "hp", Value: -3}}, OneShot: true},
			{ID: "watch_locker", Kind: "loot", Label: "巡夜军械柜", Description: "徽章槽与旧巡夜徽章尺寸完全一致。", Icon: "▥", X: 26, Y: 64, Action: "loot", Target: "chainmail", RequiresItem: "night_badge", OneShot: true},
		}
	case "aqueduct":
		common = []SceneElement{
			{ID: "sluice_wheel", Kind: "object", Label: "引水闸轮", Description: "转动闸轮可以暂时降低水位。", Icon: "◉", X: 67, Y: 42, Action: "flag", Target: "aqueduct_drained", Check: &CheckDef{Attribute: "strength", DC: 12}, OnSuccess: []Effect{{Type: "xp", Value: 7}}, OnFailure: []Effect{{Type: "hp", Value: -2}}, OneShot: true},
			{ID: "submerged_satchel", Kind: "loot", Label: "水下皮包", Description: "只有排低水位后才能够到。", Icon: "⬡", X: 32, Y: 72, Action: "loot", Target: "smoke_bomb", RequiresFlag: "aqueduct_drained", OneShot: true},
			{ID: "mason_marks", Kind: "lore", Label: "工匠导流刻痕", Description: "这些刻痕故意避开官方图纸。", Icon: "〆", X: 18, Y: 46, Action: "lore", Target: "aqueduct_marks", OneShot: true},
		}
	case "bridge":
		common = []SceneElement{
			{ID: "broken_span", Kind: "object", Label: "断裂桥面", Description: "可以借残留绳索跳过缺口。", Icon: "⌁", X: 50, Y: 48, Action: "flag", Target: "bridge_crossed", Check: &CheckDef{Attribute: "dexterity", DC: 13}, OnSuccess: []Effect{{Type: "xp", Value: 10}}, OnFailure: []Effect{{Type: "hp", Value: -5}}, OneShot: true},
			{ID: "hanging_pack", Kind: "loot", Label: "悬挂行囊", Description: "卡在桥下铁环上。", Icon: "◇", X: 72, Y: 68, Action: "loot", Target: "frost_salt", RequiresFlag: "bridge_crossed", OneShot: true},
		}
	case "court":
		common = []SceneElement{
			{ID: "last_verdict", Kind: "lore", Label: "最后判决石板", Description: "判决要求史官烧掉国王的完整姓名。", Icon: "▤", X: 50, Y: 38, Action: "lore", Target: "court_verdict", OneShot: true},
			{ID: "judge_chest", Kind: "loot", Label: "审判官衣箱", Description: "锁扣需要细致观察才能避开毒针。", Icon: "▣", X: 25, Y: 68, Action: "loot", Target: "scribe_coat", Check: &CheckDef{Attribute: "perception", DC: 13}, OnFailure: []Effect{{Type: "hp", Value: -4}}, OneShot: true},
		}
	case "belltower":
		common = []SceneElement{
			{ID: "bell_manual", Kind: "lore", Label: "钟塔值守手册", Description: "十三次钟鸣的真正用途被写在夹页。", Icon: "▤", X: 27, Y: 55, Action: "lore", Target: "bell_manual", OneShot: true},
			{ID: "empty_bell_drive", Kind: "secret", Label: "第十三组钟架", Description: "齿轮排列和褪色墓城图上的王室记号一致。", Icon: "✧", X: 66, Y: 36, Action: "unlock", Target: "room_31", RequiresItem: "old_map", Check: &CheckDef{Attribute: "perception", DC: 14}, OnSuccess: []Effect{{Type: "xp", Value: 12}}, OneShot: true},
		}
	case "reliquary":
		common = []SceneElement{
			{ID: "royal_case", Kind: "loot", Label: "王族封盒", Description: "王印凹槽仍保留完整形状。", Icon: "♛", X: 52, Y: 42, Action: "loot", Target: "royal_plate", RequiresItem: "royal_signet", OneShot: true},
			{ID: "ember_vial", Kind: "loot", Label: "黑火样本", Description: "封在厚玻璃里的活性余烬。", Icon: "♦", X: 72, Y: 64, Action: "loot", Target: "ember_flask", OneShot: true},
		}
	case "mausoleum":
		common = []SceneElement{
			{ID: "nameless_sarcophagus", Kind: "loot", Label: "无名王族石棺", Description: "骨制钥匙与棺盖锁孔完全吻合。", Icon: "▰", X: 52, Y: 48, Action: "loot", Target: "sunken_mace", RequiresItem: "bone_key", OneShot: true},
			{ID: "crypt_gate", Kind: "secret", Label: "内陵石门", Description: "石门背面通往王庭侧廊。", Icon: "▥", X: 80, Y: 40, Action: "message", Target: "石门背后传来王庭黑火的热浪。你找到了一条避开主路的侧廊。", RequiresFlag: "world_shift_2", OneShot: true},
		}
	case "boss":
		common = []SceneElement{
			{ID: "crown_chain", Kind: "lore", Label: "王冠锁链", Description: "每一条锁链都刻着不同的誓约。", Icon: "⛓", X: 22, Y: 42, Action: "lore", Target: "crown_chains", OneShot: true},
		}
	default:
		// 普通战斗/事件房也至少有一个可点对象，避免清场后变成空房。
		pick := int(hash64(fmt.Sprintf("element:%d:%s", seed, roomID)) % 3)
		if pick == 0 {
			common = []SceneElement{{ID: "wall_marks", Kind: "lore", Label: "墙面刻痕", Description: "像是某种撤退路线标记。", Icon: "〆", X: 27, Y: 44, Action: "lore", Target: "wall_marks", OneShot: true}}
		} else if pick == 1 {
			common = []SceneElement{{ID: "broken_cache", Kind: "loot", Label: "破损补给箱", Description: "也许还有遗漏的东西。", Icon: "▣", X: 72, Y: 68, Action: "loot", Target: "bandage", OneShot: true}}
		} else {
			common = []SceneElement{{ID: "memory_echo", Kind: "event", Label: "记忆残响", Description: "一段半透明影像正在重复。", Icon: "◌", X: 57, Y: 43, Action: "event", Target: "memory_mirror", OneShot: true}}
		}
	}

	if roomID == "room_22" {
		common = append(common, SceneElement{ID: "royal_crypt_door", Kind: "secret", Label: "王陵侧门", Description: "门锁需要囚室总钥。", Icon: "▥", X: 86, Y: 50, Action: "unlock", Target: "room_32", RequiresItem: "prison_key", OneShot: true})
	}

	// 每个房间再附加一个轻量环境热点。它们主要用于增强“这个房间真的可以摸索”的感觉，
	// 不承担关键奖励，避免内容量增加后破坏数值平衡。
	ambientPick := int(hash64(fmt.Sprintf("ambient:%d:%s", seed, roomID)) % 5)
	ambients := []SceneElement{
		{ID: "ambient_relief", Kind: "object", Label: "褪色浮雕", Description: "浮雕上的人物都被刻意刮去了脸。", Icon: "▧", X: 14, Y: 55, Action: "message", Target: "你擦去石灰，发现所有人物的脸都在同一时期被凿掉。有人害怕的不是他们的身份，而是他们的名字。", OneShot: true},
		{ID: "ambient_tracks", Kind: "object", Label: "灰尘脚印", Description: "脚印到墙边后突然消失。", Icon: "⌁", X: 84, Y: 76, Action: "message", Target: "脚印很新，却在墙壁前凭空结束。墓城的道路似乎并不总与石墙保持一致。", OneShot: true},
		{ID: "ambient_candle", Kind: "object", Label: "蓝焰残烛", Description: "没有烛芯，却仍冒着一点冷光。", Icon: "˙", X: 16, Y: 72, Action: "message", Target: "你靠近残烛时，火焰短暂朝墓城深处倾斜，像在替某种看不见的风指路。", OneShot: true},
		{ID: "ambient_helmet", Kind: "object", Label: "裂开的头盔", Description: "头盔内侧刻着十二道短痕。", Icon: "◒", X: 82, Y: 58, Action: "message", Target: "十二道刻痕整齐排列，第十三道的位置只有一道很深的划伤。持有者似乎拒绝完成最后一次计数。", OneShot: true},
		{ID: "ambient_ash", Kind: "object", Label: "聚拢的灰烬", Description: "灰烬像被某种呼吸缓慢推开又聚拢。", Icon: "∴", X: 18, Y: 63, Action: "message", Target: "你停下来观察，灰烬的起伏与远处钟声完全同步。这里的尘埃似乎也属于封印的一部分。", OneShot: true},
	}
	common = append(common, ambients[ambientPick])
	return common
}

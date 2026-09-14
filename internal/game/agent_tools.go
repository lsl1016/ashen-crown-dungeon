package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// AgentToolDefinition is the canonical definition used by both the HTTP Tool Gateway
// and the MCP adapter. runId is explicit because MCP 2026-07-28 is stateless.
type AgentToolDefinition struct {
	Name         string         `json:"name"`
	Title        string         `json:"title,omitempty"`
	Description  string         `json:"description"`
	InputSchema  map[string]any `json:"inputSchema"`
	OutputSchema map[string]any `json:"outputSchema,omitempty"`
	Annotations  map[string]any `json:"annotations,omitempty"`
	Meta         map[string]any `json:"_meta,omitempty"`
}

type AgentToolResult struct {
	Tool     string   `json:"tool"`
	RunID    string   `json:"runId"`
	Changed  bool     `json:"changed"`
	DryRun   bool     `json:"dryRun,omitempty"`
	Risk     string   `json:"risk,omitempty"`
	AuditID  string   `json:"auditId,omitempty"`
	Summary  string   `json:"summary"`
	Data     any      `json:"data,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

func agentOutputSchema() map[string]any {
	return map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
		"properties": map[string]any{
			"tool":     map[string]any{"type": "string"},
			"runId":    map[string]any{"type": "string"},
			"changed":  map[string]any{"type": "boolean"},
			"dryRun":   map[string]any{"type": "boolean"},
			"risk":     map[string]any{"type": "string"},
			"auditId":  map[string]any{"type": "string"},
			"summary":  map[string]any{"type": "string"},
			"data":     map[string]any{},
			"warnings": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
		"required":             []string{"tool", "runId", "changed", "summary"},
		"additionalProperties": false,
	}
}

func agentObject(properties map[string]any, required ...string) map[string]any {
	properties["runId"] = map[string]any{
		"type":         "string",
		"minLength":    1,
		"description":  "正在操作的游戏 Run ID。MCP 无隐式会话，因此每次调用都必须显式携带。",
		"x-mcp-header": "Run-Id",
	}
	return map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"type":                 "object",
		"properties":           properties,
		"required":             append([]string{"runId"}, required...),
		"additionalProperties": false,
	}
}

func mutationSchema(properties map[string]any, required ...string) map[string]any {
	properties["dryRun"] = map[string]any{"type": "boolean", "default": false, "description": "仅在副本上执行并返回预览，不写入真实 Run。高影响操作应先 dryRun。"}
	properties["reason"] = map[string]any{"type": "string", "maxLength": 200, "description": "简短说明这次世界修改的剧情/规则原因，便于审计。"}
	return agentObject(properties, required...)
}

func agentDef(name, title, description, category, risk string, readOnly bool, input map[string]any) AgentToolDefinition {
	return AgentToolDefinition{
		Name: name, Title: title, Description: description,
		InputSchema:  input,
		OutputSchema: agentOutputSchema(),
		Annotations: map[string]any{
			"readOnlyHint":    readOnly,
			"destructiveHint": risk == "high",
			"idempotentHint":  readOnly,
			"openWorldHint":   false,
		},
		Meta: map[string]any{
			"ashen-crown/category": category,
			"ashen-crown/risk":     risk,
		},
	}
}

func AgentToolDefinitions() []AgentToolDefinition {
	defs := []AgentToolDefinition{
		agentDef("inspect_world", "Inspect World", "读取当前 Run 的精简世界事实。默认只返回玩家已知信息；visibility=gm 可读取隐藏 Flag、NPC 真实位置与所有区域状态。先读后写。", "inspect", "low", true,
			agentObject(map[string]any{"visibility": map[string]any{"type": "string", "enum": []string{"player", "gm"}, "default": "player"}})),
		agentDef("inspect_room", "Inspect Room", "读取一个房间、邻接道路、场景状态、可见交互物与在场 NPC。roomId 省略时使用当前房间。", "inspect", "low", true,
			agentObject(map[string]any{
				"roomId":     map[string]any{"type": "string"},
				"visibility": map[string]any{"type": "string", "enum": []string{"player", "gm"}, "default": "player"},
			})),
		agentDef("inspect_npc", "Inspect NPC", "读取 NPC 定义、当前地点、日程状态、关系值、健康/情绪和已知情报。", "inspect", "low", true,
			agentObject(map[string]any{"npcId": map[string]any{"type": "string", "minLength": 1}}, "npcId")),
		agentDef("inspect_quest", "Inspect Quest", "读取单个任务，或省略 questId 读取全部任务；包含阶段、结局和可用任务图。", "inspect", "low", true,
			agentObject(map[string]any{"questId": map[string]any{"type": "string"}})),
		agentDef("inspect_combat", "Inspect Combat", "读取当前战斗事实：敌人、意图、距离、地形危险、玩家状态和可用职业技能。", "inspect", "low", true,
			agentObject(map[string]any{})),
		agentDef("list_world_events", "List World Events", "读取当前世界事件状态以及对应事件定义。用于判断世界正在发生什么，而不是凭提示词猜测。", "inspect", "low", true,
			agentObject(map[string]any{})),
		agentDef("search_catalog", "Search Catalog", "按类型和关键词搜索合法的房间、NPC、敌人、物品、技能或世界事件 ID。调用 grant_item/spawn_enemy/move_npc 等工具前，不确定 ID 时先搜索。", "inspect", "low", true,
			agentObject(map[string]any{
				"kind":  map[string]any{"type": "string", "enum": []string{"room", "npc", "enemy", "item", "skill", "world_event"}},
				"query": map[string]any{"type": "string", "maxLength": 80},
				"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 30, "default": 10},
			}, "kind")),
		agentDef("recent_log", "Recent World Log", "读取最近的真实游戏日志，用于理解为什么世界变成现在这样。可按日志类型过滤，避免依赖模型记忆。", "inspect", "low", true,
			agentObject(map[string]any{
				"type":  map[string]any{"type": "string", "maxLength": 40},
				"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 50, "default": 12},
			})),
		agentDef("roll_check", "Roll Attribute Check", "执行由 Game Runtime 决定结果的 D20 属性检定。不要由模型自行掷骰或宣布成功。", "rule", "medium", false,
			mutationSchema(map[string]any{
				"attribute": map[string]any{"type": "string", "enum": []string{"strength", "dexterity", "perception", "will"}},
				"dc":        map[string]any{"type": "integer", "minimum": 2, "maximum": 30},
				"label":     map[string]any{"type": "string", "minLength": 1, "maxLength": 80},
			}, "attribute", "dc", "label")),
		agentDef("move_npc", "Move NPC", "把持久 NPC 移到已存在的房间，并写入 NPC Override。只在剧情有明确因果时使用。", "world_mutation", "medium", false,
			mutationSchema(map[string]any{
				"npcId":  map[string]any{"type": "string", "minLength": 1},
				"roomId": map[string]any{"type": "string", "minLength": 1},
			}, "npcId", "roomId")),
		agentDef("set_region_state", "Set Region State", "修改区域持续状态并同步场景表现。优先使用已有状态命名，不要随意创造无法被 UI 理解的值。", "world_mutation", "high", false,
			mutationSchema(map[string]any{
				"region": map[string]any{"type": "string", "minLength": 1},
				"state":  map[string]any{"type": "string", "minLength": 1, "maxLength": 64},
			}, "region", "state")),
		agentDef("trigger_world_event", "Trigger World Event", "触发一个已经在内容库定义的 Living World Event。适合重大剧情变化；应先 inspect_world/list_world_events。", "world_mutation", "high", false,
			mutationSchema(map[string]any{"eventId": map[string]any{"type": "string", "minLength": 1}}, "eventId")),
		agentDef("advance_bell", "Advance World Bell", "推进世界钟声 1-3 次，会驱动 NPC 日程、商店刷新和世界事件。高影响，默认先 dryRun。", "world_mutation", "high", false,
			mutationSchema(map[string]any{"steps": map[string]any{"type": "integer", "minimum": 1, "maximum": 3, "default": 1}})),
		agentDef("set_flag", "Set World Flag", "写入世界事实。已有 Flag 可更新；新 Flag 必须使用 ai. 前缀，避免污染引擎内部标记。", "world_mutation", "medium", false,
			mutationSchema(map[string]any{
				"key":   map[string]any{"type": "string", "minLength": 1, "maxLength": 120},
				"value": map[string]any{"type": "boolean"},
			}, "key", "value")),
		agentDef("reveal_room", "Reveal Room", "让一个已存在房间在玩家地图上被发现，不创建内容。", "map_mutation", "medium", false,
			mutationSchema(map[string]any{"roomId": map[string]any{"type": "string", "minLength": 1}}, "roomId")),
		agentDef("create_room", "Create Room", "创建一个新地图节点，可选立即连接到现有房间。只创建数据，不允许执行代码或自定义脚本。", "map_mutation", "high", false,
			mutationSchema(map[string]any{
				"id":          map[string]any{"type": "string", "pattern": "^[a-zA-Z0-9_.-]{1,64}$"},
				"name":        map[string]any{"type": "string", "minLength": 1, "maxLength": 80},
				"type":        map[string]any{"type": "string", "minLength": 1, "maxLength": 40},
				"zone":        map[string]any{"type": "string", "maxLength": 60},
				"description": map[string]any{"type": "string", "maxLength": 500},
				"x":           map[string]any{"type": "integer", "minimum": -100, "maximum": 100},
				"y":           map[string]any{"type": "integer", "minimum": -100, "maximum": 100},
				"connectTo":   map[string]any{"type": "string"},
				"discovered":  map[string]any{"type": "boolean", "default": false},
			}, "id", "name", "type", "x", "y")),
		agentDef("connect_rooms", "Connect Rooms", "连接两个已存在的地图节点。不会自动解锁或发现房间。", "map_mutation", "medium", false,
			mutationSchema(map[string]any{
				"from": map[string]any{"type": "string", "minLength": 1},
				"to":   map[string]any{"type": "string", "minLength": 1},
			}, "from", "to")),
		agentDef("spawn_enemy", "Spawn Enemy", "在当前房间生成一个预定义敌人并进入战斗。当前已有战斗时拒绝。", "encounter", "high", false,
			mutationSchema(map[string]any{"enemyId": map[string]any{"type": "string", "minLength": 1}}, "enemyId")),
		agentDef("grant_item", "Grant Item", "给予一个内容库中已存在的物品。不要用它绕过任务/商店规则；用于明确剧情奖励或 GM 修复。", "reward", "high", false,
			mutationSchema(map[string]any{"itemId": map[string]any{"type": "string", "minLength": 1}}, "itemId")),
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].Name < defs[j].Name })
	return defs
}

func AgentToolExists(name string) bool {
	for _, def := range AgentToolDefinitions() {
		if def.Name == name {
			return true
		}
	}
	return false
}

func (e *Engine) ExecuteAgentTool(run *Run, name string, args map[string]any) (AgentToolResult, error) {
	if run == nil {
		return AgentToolResult{}, errors.New("run 不能为空")
	}
	e.Prepare(run)
	base := AgentToolResult{Tool: name, RunID: run.ID}
	switch name {
	case "inspect_world":
		visibility := stringArg(args, "visibility", "player")
		data := e.agentWorldView(run, visibility == "gm")
		base.Summary = fmt.Sprintf("第 %d 钟；玩家位于 %s；世界事件 %d 个。", run.Clock.Bell, roomName(run, run.CurrentRoomID), len(run.WorldEvents))
		base.Data = data
		return base, nil
	case "inspect_room":
		roomID := stringArg(args, "roomId", run.CurrentRoomID)
		visibility := stringArg(args, "visibility", "player")
		data, err := e.agentRoomView(run, roomID, visibility == "gm")
		if err != nil {
			return base, err
		}
		base.Summary = fmt.Sprintf("已读取地点：%s。", roomName(run, roomID))
		base.Data = data
		return base, nil
	case "inspect_npc":
		npcID := stringArg(args, "npcId", "")
		npc, ok := NPCs[npcID]
		if !ok {
			return base, fmt.Errorf("未知 NPC: %s", npcID)
		}
		ws := run.NPCWorld[npcID]
		base.Summary = fmt.Sprintf("%s 当前位于 %s，关系值 %d。", npc.Name, roomName(run, run.NPCLocations[npcID]), run.NPCRelations[npcID])
		base.Data = map[string]any{"definition": npc, "world": ws, "location": run.NPCLocations[npcID], "relation": run.NPCRelations[npcID], "override": run.NPCOverrides[npcID]}
		return base, nil
	case "inspect_quest":
		questID := stringArg(args, "questId", "")
		if questID != "" {
			q := run.Quests[questID]
			if q == nil {
				return base, fmt.Errorf("未知任务: %s", questID)
			}
			base.Summary = fmt.Sprintf("任务「%s」：%s / %s。", q.Title, q.Status, q.Stage)
			base.Data = map[string]any{"quest": q, "decision": run.QuestDecisions[questID], "graph": QuestGraphs[questID]}
			return base, nil
		}
		ids := make([]string, 0, len(run.Quests))
		for id := range run.Quests {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		quests := make([]any, 0, len(ids))
		for _, id := range ids {
			quests = append(quests, map[string]any{"quest": run.Quests[id], "decision": run.QuestDecisions[id], "graph": QuestGraphs[id]})
		}
		base.Summary = fmt.Sprintf("共有 %d 个任务状态。", len(quests))
		base.Data = map[string]any{"quests": quests}
		return base, nil
	case "inspect_combat":
		if run.Combat == nil {
			base.Summary = "当前没有战斗。"
			base.Data = map[string]any{"active": false}
			return base, nil
		}
		enemy := Enemies[run.Combat.EnemyID]
		skills := []SkillDef{}
		for _, id := range run.Player.Skills {
			if s, ok := Skills[id]; ok {
				skills = append(skills, s)
			}
		}
		base.Summary = fmt.Sprintf("正在与 %s 战斗：%d/%d HP，距离 %d。", run.Combat.EnemyName, run.Combat.EnemyHP, run.Combat.EnemyMaxHP, run.Combat.Distance)
		base.Data = map[string]any{"active": true, "combat": run.Combat, "enemy": enemy, "player": map[string]any{"hp": run.Player.HP, "maxHp": run.Player.MaxHP, "energy": run.Player.Energy, "maxEnergy": run.Player.MaxEnergy, "defense": run.Player.Defense, "statuses": run.Player.Statuses}, "skills": skills}
		return base, nil
	case "list_world_events":
		ids := make([]string, 0, len(LivingWorldEvents))
		for id := range LivingWorldEvents {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		events := make([]any, 0, len(ids))
		for _, id := range ids {
			events = append(events, map[string]any{"definition": LivingWorldEvents[id], "state": run.WorldEvents[id]})
		}
		base.Summary = fmt.Sprintf("内容库定义 %d 个 Living World Event。", len(events))
		base.Data = map[string]any{"events": events}
		return base, nil
	case "search_catalog":
		kind := stringArg(args, "kind", "")
		query := strings.ToLower(strings.TrimSpace(stringArg(args, "query", "")))
		limit := intArg(args, "limit", 10)
		if limit < 1 {
			limit = 1
		}
		if limit > 30 {
			limit = 30
		}
		items, err := e.agentSearchCatalog(run, kind, query, limit)
		if err != nil {
			return base, err
		}
		base.Summary = fmt.Sprintf("%s 目录命中 %d 项。", kind, len(items))
		base.Data = map[string]any{"kind": kind, "query": query, "items": items}
		return base, nil
	case "recent_log":
		typ := strings.TrimSpace(stringArg(args, "type", ""))
		limit := intArg(args, "limit", 12)
		if limit < 1 {
			limit = 1
		}
		if limit > 50 {
			limit = 50
		}
		entries := make([]LogEntry, 0, limit)
		for i := len(run.Log) - 1; i >= 0 && len(entries) < limit; i-- {
			entry := run.Log[i]
			if typ != "" && entry.Type != typ {
				continue
			}
			entries = append(entries, entry)
		}
		base.Summary = fmt.Sprintf("返回最近 %d 条%s日志。", len(entries), func() string {
			if typ == "" {
				return ""
			}
			return " " + typ + " "
		}())
		base.Data = map[string]any{"entries": entries}
		return base, nil
	case "roll_check":
		attr := stringArg(args, "attribute", "")
		if attr != "strength" && attr != "dexterity" && attr != "perception" && attr != "will" {
			return base, errors.New("attribute 必须是 strength/dexterity/perception/will")
		}
		dc := intArg(args, "dc", 0)
		if dc < 2 || dc > 30 {
			return base, errors.New("dc 必须在 2-30")
		}
		label := strings.TrimSpace(stringArg(args, "label", ""))
		if label == "" || len([]rune(label)) > 80 {
			return base, errors.New("label 必须为 1-80 字符")
		}
		raw := e.roll(run, "agent-check:"+label, 20)
		bonus := e.attribute(run.Player.Attributes, attr)
		total := raw + bonus
		success := raw != 1 && (raw == 20 || total >= dc)
		rr := &RollResult{Kind: "agent_check", Label: label, Roll: raw, Bonus: bonus, Total: total, DC: dc, Success: success, Critical: raw == 20}
		run.LastRoll = rr
		e.log(run, "check", fmt.Sprintf("GM 检定「%s」：D20 %d + %d = %d / DC %d，成功=%t。", label, raw, bonus, total, dc, success))
		base.Changed = true
		base.Summary = fmt.Sprintf("%s：D20 %d + %d = %d / DC %d，成功=%t。", label, raw, bonus, total, dc, success)
		base.Data = rr
		return base, nil
	case "move_npc":
		npcID, roomID := stringArg(args, "npcId", ""), stringArg(args, "roomId", "")
		if err := e.GMMoveNPC(run, npcID, roomID); err != nil {
			return base, err
		}
		base.Changed = true
		base.Summary = fmt.Sprintf("NPC %s 已移动到 %s。", npcID, roomName(run, roomID))
		base.Data = map[string]any{"npcId": npcID, "roomId": roomID}
		return base, nil
	case "set_region_state":
		region, state := stringArg(args, "region", ""), stringArg(args, "state", "")
		if err := e.GMSetRegion(run, region, state); err != nil {
			return base, err
		}
		base.Changed = true
		base.Summary = fmt.Sprintf("区域 %s 状态已设为 %s。", region, state)
		base.Data = map[string]any{"region": region, "state": state}
		return base, nil
	case "trigger_world_event":
		eventID := stringArg(args, "eventId", "")
		if err := e.GMTriggerWorldEvent(run, eventID); err != nil {
			return base, err
		}
		base.Changed = true
		base.Summary = fmt.Sprintf("世界事件 %s 已触发。", eventID)
		base.Data = map[string]any{"eventId": eventID, "state": run.WorldEvents[eventID]}
		return base, nil
	case "advance_bell":
		steps := intArg(args, "steps", 1)
		if steps < 1 || steps > 3 {
			return base, errors.New("steps 必须在 1-3")
		}
		before := run.Clock.Bell
		for i := 0; i < steps; i++ {
			if err := e.GMAdvanceBell(run); err != nil {
				return base, err
			}
		}
		base.Changed = true
		base.Summary = fmt.Sprintf("世界钟从 %d 推进到 %d。", before, run.Clock.Bell)
		base.Data = map[string]any{"before": before, "after": run.Clock.Bell, "worldEvents": run.WorldEvents}
		return base, nil
	case "set_flag":
		key := strings.TrimSpace(stringArg(args, "key", ""))
		value := boolArg(args, "value", false)
		if key == "" {
			return base, errors.New("key 必填")
		}
		if _, exists := run.Flags[key]; !exists && !strings.HasPrefix(key, "ai.") {
			return base, errors.New("新 Flag 必须使用 ai. 前缀；已有引擎 Flag 可以更新")
		}
		if err := e.GMSetFlag(run, key, value); err != nil {
			return base, err
		}
		base.Changed = true
		base.Summary = fmt.Sprintf("Flag %s=%t。", key, value)
		base.Data = map[string]any{"key": key, "value": value}
		return base, nil
	case "reveal_room":
		roomID := stringArg(args, "roomId", "")
		if run.Rooms[roomID] == nil {
			return base, errors.New("房间不存在")
		}
		run.Rooms[roomID].Discovered = true
		e.log(run, "world", "发现地点：「"+run.Rooms[roomID].Name+"」。")
		base.Changed = true
		base.Summary = fmt.Sprintf("已发现地点 %s。", run.Rooms[roomID].Name)
		base.Data = run.Rooms[roomID]
		return base, nil
	case "create_room":
		id := strings.TrimSpace(stringArg(args, "id", ""))
		namev := strings.TrimSpace(stringArg(args, "name", ""))
		typ := strings.TrimSpace(stringArg(args, "type", ""))
		if id == "" || namev == "" || typ == "" {
			return base, errors.New("id/name/type 必填")
		}
		if run.Rooms[id] != nil {
			return base, errors.New("room id 已存在")
		}
		x, y := intArg(args, "x", 0), intArg(args, "y", 0)
		if x < -100 || x > 100 || y < -100 || y > 100 {
			return base, errors.New("x/y 必须在 -100..100")
		}
		zone := strings.TrimSpace(stringArg(args, "zone", "AI扩展区"))
		desc := strings.TrimSpace(stringArg(args, "description", "这是由 AI GM 通过受控工具创建的新地点。"))
		room := &Room{ID: id, Name: namev, Type: typ, Scene: sceneForType(typ, 0), Description: desc, Zone: zone, X: x, Y: y, Discovered: boolArg(args, "discovered", false)}
		run.Rooms[id] = room
		connectTo := strings.TrimSpace(stringArg(args, "connectTo", ""))
		if connectTo != "" {
			if run.Rooms[connectTo] == nil {
				delete(run.Rooms, id)
				return base, errors.New("connectTo 房间不存在")
			}
			run.Edges = append(run.Edges, Edge{From: connectTo, To: id})
		}
		e.log(run, "world", fmt.Sprintf("世界出现新地点：「%s」。", namev))
		base.Changed = true
		base.Summary = fmt.Sprintf("已创建地点 %s (%s)。", namev, id)
		base.Data = map[string]any{"room": room, "connectTo": connectTo}
		return base, nil
	case "connect_rooms":
		from, to := stringArg(args, "from", ""), stringArg(args, "to", "")
		if from == to {
			return base, errors.New("不能连接同一个房间")
		}
		if run.Rooms[from] == nil || run.Rooms[to] == nil {
			return base, errors.New("房间不存在")
		}
		if !e.adjacent(run, from, to) {
			run.Edges = append(run.Edges, Edge{From: from, To: to})
			e.log(run, "world", fmt.Sprintf("新道路连接 %s 与 %s。", run.Rooms[from].Name, run.Rooms[to].Name))
		}
		base.Changed = true
		base.Summary = fmt.Sprintf("已连接 %s ↔ %s。", roomName(run, from), roomName(run, to))
		base.Data = map[string]any{"from": from, "to": to}
		return base, nil
	case "spawn_enemy":
		enemyID := stringArg(args, "enemyId", "")
		if _, ok := Enemies[enemyID]; !ok {
			return base, fmt.Errorf("未知敌人: %s", enemyID)
		}
		if run.Combat != nil {
			return base, errors.New("当前已经在战斗中")
		}
		e.startCombat(run, enemyID)
		base.Changed = true
		base.Summary = fmt.Sprintf("遭遇已开始：%s。", Enemies[enemyID].Name)
		base.Data = run.Combat
		return base, nil
	case "grant_item":
		itemID := stringArg(args, "itemId", "")
		item, ok := Items[itemID]
		if !ok {
			return base, fmt.Errorf("未知物品: %s", itemID)
		}
		run.Player.Inventory = append(run.Player.Inventory, itemID)
		e.log(run, "loot", "GM 奖励物品：「"+item.Name+"」。")
		base.Changed = true
		base.Summary = fmt.Sprintf("已给予物品：%s。", item.Name)
		base.Data = map[string]any{"item": item, "inventoryCount": len(run.Player.Inventory)}
		return base, nil
	default:
		return base, fmt.Errorf("未知 Agent Tool: %s", name)
	}
}

func (e *Engine) agentWorldView(run *Run, gm bool) map[string]any {
	current := run.Rooms[run.CurrentRoomID]
	quests := make([]*QuestState, 0, len(run.Quests))
	ids := make([]string, 0, len(run.Quests))
	for id := range run.Quests {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		quests = append(quests, run.Quests[id])
	}
	data := map[string]any{
		"runId":          run.ID,
		"turn":           run.Turn,
		"clock":          run.Clock,
		"currentRoom":    current,
		"player":         map[string]any{"name": run.Player.Name, "class": run.Player.Class, "level": run.Player.Level, "hp": run.Player.HP, "maxHp": run.Player.MaxHP, "energy": run.Player.Energy, "maxEnergy": run.Player.MaxEnergy, "gold": run.Player.Gold, "attributes": run.Player.Attributes, "statuses": run.Player.Statuses, "equipment": run.Player.Equipment, "inventory": run.Player.Inventory},
		"quests":         quests,
		"activeEvent":    run.ActiveEvent,
		"activeDialogue": run.ActiveDialogue,
		"combat":         run.Combat,
		"worldEvents":    run.WorldEvents,
		"sceneState":     run.SceneStates[run.CurrentRoomID],
		"gameOver":       run.GameOver,
		"victory":        run.Victory,
	}
	if gm {
		data["flags"] = run.Flags
		data["npcLocations"] = run.NPCLocations
		data["npcWorld"] = run.NPCWorld
		data["regionStates"] = run.RegionStates
		data["questDecisions"] = run.QuestDecisions
	} else {
		npcs := []map[string]any{}
		for id, loc := range run.NPCLocations {
			if loc == run.CurrentRoomID {
				if n, ok := NPCs[id]; ok {
					npcs = append(npcs, map[string]any{"id": id, "name": n.Name, "title": n.Title, "relation": run.NPCRelations[id]})
				}
			}
		}
		data["npcsHere"] = npcs
	}
	return data
}

func (e *Engine) agentRoomView(run *Run, roomID string, gm bool) (map[string]any, error) {
	room := run.Rooms[roomID]
	if room == nil {
		return nil, errors.New("房间不存在")
	}
	adj := []map[string]any{}
	for _, edge := range run.Edges {
		other := ""
		if edge.From == roomID {
			other = edge.To
		} else if edge.To == roomID {
			other = edge.From
		}
		if other == "" || run.Rooms[other] == nil {
			continue
		}
		rr := run.Rooms[other]
		if gm || rr.Discovered {
			adj = append(adj, map[string]any{"id": rr.ID, "name": rr.Name, "locked": rr.Locked, "discovered": rr.Discovered})
		}
	}
	elems := []SceneElement{}
	for _, el := range room.Elements {
		if gm || el.HiddenUnlessFlag == "" || run.Flags[el.HiddenUnlessFlag] {
			elems = append(elems, el)
		}
	}
	npcs := []map[string]any{}
	for id, loc := range run.NPCLocations {
		if loc == roomID {
			if n, ok := NPCs[id]; ok {
				npcs = append(npcs, map[string]any{"id": id, "name": n.Name, "title": n.Title, "relation": run.NPCRelations[id], "world": run.NPCWorld[id]})
			}
		}
	}
	return map[string]any{"room": room, "adjacent": adj, "elements": elems, "npcs": npcs, "sceneState": run.SceneStates[roomID]}, nil
}

func (e *Engine) agentSearchCatalog(run *Run, kind, query string, limit int) ([]map[string]any, error) {
	matches := func(parts ...string) bool {
		if query == "" {
			return true
		}
		for _, part := range parts {
			if strings.Contains(strings.ToLower(part), query) {
				return true
			}
		}
		return false
	}
	out := []map[string]any{}
	appendIf := func(v map[string]any) {
		if len(out) < limit {
			out = append(out, v)
		}
	}
	switch kind {
	case "room":
		ids := make([]string, 0, len(run.Rooms))
		for id := range run.Rooms {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			r := run.Rooms[id]
			if r == nil || !matches(id, r.Name, r.Type, r.Zone, r.Description) {
				continue
			}
			appendIf(map[string]any{"id": id, "name": r.Name, "type": r.Type, "zone": r.Zone, "discovered": r.Discovered, "locked": r.Locked})
		}
	case "npc":
		ids := make([]string, 0, len(NPCs))
		for id := range NPCs {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			n := NPCs[id]
			if !matches(id, n.Name, n.Title, n.Description) {
				continue
			}
			appendIf(map[string]any{"id": id, "name": n.Name, "title": n.Title, "location": run.NPCLocations[id], "relation": run.NPCRelations[id]})
		}
	case "enemy":
		ids := make([]string, 0, len(Enemies))
		for id := range Enemies {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			en := Enemies[id]
			if !matches(id, en.Name, en.Description, en.Archetype, en.Weakness) {
				continue
			}
			appendIf(map[string]any{"id": id, "name": en.Name, "boss": en.Boss, "archetype": en.Archetype, "weakness": en.Weakness})
		}
	case "item":
		ids := make([]string, 0, len(Items))
		for id := range Items {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			it := Items[id]
			if !matches(id, it.Name, it.Type, it.Rarity, it.Description, it.Special) {
				continue
			}
			appendIf(map[string]any{"id": id, "name": it.Name, "type": it.Type, "slot": it.Slot, "rarity": it.Rarity, "value": it.Value})
		}
	case "skill":
		ids := make([]string, 0, len(Skills))
		for id := range Skills {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			sk := Skills[id]
			if !matches(id, sk.Name, sk.Class, sk.Description, strings.Join(sk.Tags, " ")) {
				continue
			}
			appendIf(map[string]any{"id": id, "name": sk.Name, "class": sk.Class, "cost": sk.Cost, "cooldown": sk.Cooldown, "tags": sk.Tags})
		}
	case "world_event":
		ids := make([]string, 0, len(LivingWorldEvents))
		for id := range LivingWorldEvents {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			ev := LivingWorldEvents[id]
			if !matches(id, ev.Title, ev.Description, ev.Region) {
				continue
			}
			appendIf(map[string]any{"id": id, "title": ev.Title, "region": ev.Region, "triggerBell": ev.TriggerBell, "severity": ev.Severity, "state": run.WorldEvents[id]})
		}
	default:
		return nil, fmt.Errorf("未知 catalog kind: %s", kind)
	}
	return out, nil
}

func roomName(run *Run, id string) string {
	if r := run.Rooms[id]; r != nil {
		return r.Name
	}
	if id == "" {
		return "未知地点"
	}
	return id
}

func stringArg(args map[string]any, key, fallback string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return fallback
}
func intArg(args map[string]any, key string, fallback int) int {
	switch v := args[key].(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return int(i)
		}
	}
	return fallback
}

func boolArg(args map[string]any, key string, fallback bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return fallback
}

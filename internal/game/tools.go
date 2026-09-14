package game

import (
	"errors"
	"fmt"
)

type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

func ToolDefinitions() []ToolDefinition {
	obj := func(props map[string]any, required ...string) map[string]any {
		return map[string]any{"type": "object", "properties": props, "required": required}
	}
	return []ToolDefinition{
		{Name: "create_room", Description: "在当前世界中创建一个新房间节点。AI 只能通过该工具创造地图内容。", InputSchema: obj(map[string]any{"id": map[string]any{"type": "string"}, "name": map[string]any{"type": "string"}, "type": map[string]any{"type": "string"}, "x": map[string]any{"type": "integer"}, "y": map[string]any{"type": "integer"}}, "id", "name", "type", "x", "y")},
		{Name: "connect_rooms", Description: "连接两个已存在的地图节点。", InputSchema: obj(map[string]any{"from": map[string]any{"type": "string"}, "to": map[string]any{"type": "string"}}, "from", "to")},
		{Name: "reveal_room", Description: "让一个房间在地图上被玩家发现。", InputSchema: obj(map[string]any{"roomId": map[string]any{"type": "string"}}, "roomId")},
		{Name: "set_flag", Description: "写入世界事实/剧情标记。", InputSchema: obj(map[string]any{"key": map[string]any{"type": "string"}, "value": map[string]any{"type": "boolean"}}, "key", "value")},
		{Name: "grant_item", Description: "由规则引擎校验后给予玩家物品。", InputSchema: obj(map[string]any{"itemId": map[string]any{"type": "string"}}, "itemId")},
		{Name: "spawn_enemy", Description: "在当前房间生成一个预定义敌人并开始战斗。", InputSchema: obj(map[string]any{"enemyId": map[string]any{"type": "string"}}, "enemyId")},
	}
}

func (e *Engine) ExecuteTool(run *Run, name string, args map[string]any) error {
	switch name {
	case "create_room":
		id, _ := args["id"].(string)
		namev, _ := args["name"].(string)
		typ, _ := args["type"].(string)
		x := toInt(args["x"])
		y := toInt(args["y"])
		if id == "" || namev == "" {
			return errors.New("id/name 必填")
		}
		if run.Rooms[id] != nil {
			return errors.New("room id 已存在")
		}
		run.Rooms[id] = &Room{ID: id, Name: namev, Type: typ, Scene: sceneForType(typ, 0), Description: "这是一个由世界生成工具创建的新地点。", X: x, Y: y}
		e.log(run, "world", fmt.Sprintf("世界发生变化：出现了新地点「%s」。", namev))
	case "connect_rooms":
		from, _ := args["from"].(string)
		to, _ := args["to"].(string)
		if run.Rooms[from] == nil || run.Rooms[to] == nil {
			return errors.New("房间不存在")
		}
		if !e.adjacent(run, from, to) {
			run.Edges = append(run.Edges, Edge{From: from, To: to})
		}
	case "reveal_room":
		id, _ := args["roomId"].(string)
		if run.Rooms[id] == nil {
			return errors.New("房间不存在")
		}
		run.Rooms[id].Discovered = true
	case "set_flag":
		key, _ := args["key"].(string)
		value, _ := args["value"].(bool)
		if key == "" {
			return errors.New("key 必填")
		}
		run.Flags[key] = value
	case "grant_item":
		id, _ := args["itemId"].(string)
		if _, ok := Items[id]; !ok {
			return errors.New("未知物品")
		}
		run.Player.Inventory = append(run.Player.Inventory, id)
		e.log(run, "loot", "获得物品：「"+Items[id].Name+"」。")
	case "spawn_enemy":
		id, _ := args["enemyId"].(string)
		if _, ok := Enemies[id]; !ok {
			return errors.New("未知敌人")
		}
		if run.Combat != nil {
			return errors.New("已经在战斗中")
		}
		e.startCombat(run, id)
	default:
		return errors.New("未知工具")
	}
	return nil
}

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	}
	return 0
}

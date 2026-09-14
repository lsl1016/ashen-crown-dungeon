package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ContentOverride struct {
	Kind  string          `json:"kind"`
	ID    string          `json:"id"`
	Value json.RawMessage `json:"value"`
}

type EditorOverrideFile struct {
	Overrides []ContentOverride `json:"overrides"`
}

func EditorContent() map[string]any {
	return map[string]any{
		"items":        Items,
		"enemies":      Enemies,
		"events":       Events,
		"npcs":         NPCs,
		"dialogues":    Dialogues,
		"shops":        Shops,
		"skills":       Skills,
		"worldEvents":  LivingWorldEvents,
		"npcSchedules": NPCSchedules,
		"questGraphs":  QuestGraphs,
		"talents":      Talents,
		"affixes":      Affixes,
	}
}

func ApplyContentOverride(ov ContentOverride) error {
	kind := strings.TrimSpace(ov.Kind)
	id := strings.TrimSpace(ov.ID)
	if kind == "" || id == "" || len(ov.Value) == 0 {
		return errors.New("kind、id、value 不能为空")
	}
	switch kind {
	case "items":
		var v ItemDef
		if err := json.Unmarshal(ov.Value, &v); err != nil {
			return err
		}
		v.ID = id
		Items[id] = v
	case "enemies":
		var v EnemyDef
		if err := json.Unmarshal(ov.Value, &v); err != nil {
			return err
		}
		v.ID = id
		Enemies[id] = v
	case "events":
		var v EventDef
		if err := json.Unmarshal(ov.Value, &v); err != nil {
			return err
		}
		v.ID = id
		Events[id] = v
	case "npcs":
		var v NPCDef
		if err := json.Unmarshal(ov.Value, &v); err != nil {
			return err
		}
		v.ID = id
		NPCs[id] = v
	case "dialogues":
		var v DialogueDef
		if err := json.Unmarshal(ov.Value, &v); err != nil {
			return err
		}
		v.ID = id
		Dialogues[id] = v
	case "shops":
		var v ShopDef
		if err := json.Unmarshal(ov.Value, &v); err != nil {
			return err
		}
		v.ID = id
		Shops[id] = v
	case "skills":
		var v SkillDef
		if err := json.Unmarshal(ov.Value, &v); err != nil {
			return err
		}
		v.ID = id
		Skills[id] = v
	case "worldEvents":
		var v WorldEventDef
		if err := json.Unmarshal(ov.Value, &v); err != nil {
			return err
		}
		v.ID = id
		LivingWorldEvents[id] = v
	case "npcSchedules":
		var v NPCScheduleDef
		if err := json.Unmarshal(ov.Value, &v); err != nil {
			return err
		}
		v.NPCID = id
		NPCSchedules[id] = v
	case "questGraphs":
		var v QuestGraphDef
		if err := json.Unmarshal(ov.Value, &v); err != nil {
			return err
		}
		v.ID = id
		QuestGraphs[id] = v
	default:
		return fmt.Errorf("不支持的内容类型：%s", kind)
	}
	return nil
}

func SortedEditorKinds() []string {
	out := []string{"items", "enemies", "events", "npcs", "dialogues", "shops", "skills", "worldEvents", "npcSchedules", "questGraphs"}
	sort.Strings(out)
	return out
}

func (e *Engine) GMUpsertRoom(run *Run, room Room, connectTo string) error {
	if strings.TrimSpace(room.ID) == "" || strings.TrimSpace(room.Name) == "" {
		return errors.New("房间 ID 与名称不能为空")
	}
	e.ensureV07State(run)
	if room.Zone == "" {
		room.Zone = "自定义区域"
	}
	if room.Type == "" {
		room.Type = "event"
	}
	if room.Scene == "" {
		room.Scene = "hall"
	}
	room.Discovered = true
	copyRoom := room
	run.Rooms[room.ID] = &copyRoom
	if connectTo != "" {
		if run.Rooms[connectTo] == nil {
			return errors.New("连接目标房间不存在")
		}
		found := false
		for _, ed := range run.Edges {
			if (ed.From == connectTo && ed.To == room.ID) || (ed.From == room.ID && ed.To == connectTo) {
				found = true
				break
			}
		}
		if !found {
			run.Edges = append(run.Edges, Edge{From: connectTo, To: room.ID})
		}
	}
	e.syncSceneStates(run)
	e.log(run, "gm", "GM 创建/更新地点：「"+room.Name+"」。")
	return nil
}

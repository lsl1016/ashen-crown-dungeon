# V0.7 Living World Runtime

V0.7 的目标不是继续把地图从 44 个房间扩成更多房间，而是让现有内容成为一个**会随时间运行、能产生后果、能被工具安全修改**的世界。

## 1. 状态层

Run 新增或正式纳入 V0.7 的世界事实：

```text
Player.Skills
WorldEvents
NPCWorld
SceneStates
QuestDecisions
NPCOverrides
RegionOverrides
Combat.Terrain
Combat.Hazards
```

这些字段都属于 Game Runtime 的事实，不由前端或未来 LLM 自行推断。

## 2. Living World Tick

V0.7 不引入第二套后台时钟，而是复用已有 `WorldClock`：

```text
Move
 ↓
advanceClock
 ↓ 每 5 step
Bell + 1
 ↓
onBellChanged        (V0.6 NPC / Shop / Region)
 ↓
onBellChangedV07     (WorldEvent / Scene / NPCWorld)
```

因此 Seed、移动、钟声、事件生命周期仍然可复现。

## 3. World Event

定义：

```go
type WorldEventDef struct {
    ID           string
    Title        string
    Region       string
    TriggerBell  int
    EscalateBell int
    ResolveFlag  string
    RequireFlag  string
    Severity     int
    Overlay      string
    Description  string
}
```

运行状态：

```text
nil
 ↓ TriggerBell
emerging / active
 ↓ EscalateBell
escalated / active
 ↓ ResolveFlag 或分支结果
resolved
```

事件影响：

```text
RegionStates
SceneStates
NPCWorld
Combat Hazards
World Event UI
Event Log
```

## 4. NPC Schedule 与 Override

自动日程由 `NPCScheduleDef` 定义。

GM / AI 调用 `move_npc` 时，不只是修改 `NPCLocations`，还写入：

```text
NPCOverrides[npcId] = roomId
```

之后即使执行 `Prepare()`、推进钟声或重新载入存档，位置仍然保持。

设置：

```text
roomId = @schedule
```

会移除 Override，重新回到自动日程。

这是未来 Agent 修改世界必须具备的语义：Tool 的效果不能在下一次派生状态刷新时凭空消失。

## 5. Region Override

与 NPC 相同，GM 手工修改区域状态会进入：

```text
RegionOverrides
```

派生顺序：

```text
V0.6 自动区域状态
 ↓
Living World Event 状态
 ↓
GM / AI Region Override
 ↓
SceneState
```

因此手工世界编辑拥有最后优先级。

## 6. QuestGraph

V0.7 先把三条支线结构化：

```go
type QuestGraphDef struct {
    ID    string
    Title string
    Nodes []QuestGraphNodeDef
    Edges []QuestGraphEdgeDef
}
```

Run 中保存：

```text
QuestState.Stage
QuestState.Outcome
QuestDecisions[questId]
```

QuestGraph 目前主要承担：

- 内容编辑与可视化语义
- 分支结局记录
- 世界事件 Resolve 条件
- 未来 AI 可读取的任务结构

后续可以继续把 Dialogue / Event 条件直接引用 QuestGraph Edge Condition。

## 7. Data-driven Skill

调用格式：

```text
POST /api/runs/{id}/combat
{
  "action": "skill:warden_chainbreaker"
}
```

运行时读取 `Skills[skillId]`：

```text
职业校验
技能拥有校验
Range 校验
Cooldown 校验
Energy 校验
命中检定
Effect interpreter
Cooldown 写回
敌人回合
Hazard 结算
```

当前 Effect：

```text
damage
guard
shield
energy
set_distance
retreat
enemy_status
```

新增技能只要现有 Effect 足够，就不需要新增 Go 分支。

## 8. Battlefield Terrain

CombatState：

```text
terrain
terrainHint
hazards[]
```

Hazard 结构：

```text
id
label
description
distance
damage
```

基础地形来自 Room Type；动态 Hazard 可以来自 SceneState / WorldEvent。

这使 Room、WorldEvent、Combat 三个系统第一次产生直接规则联动。

## 9. SceneState

`SceneState` 是世界事实到前端表现之间的稳定中间层：

```text
RegionStates + WorldEvents
          ↓
      SceneState
          ↓
CSS Overlay / Environment Layer / Battlefield Hazard
```

前端不需要理解“为什么灰疫发生”，只读取：

```json
{
  "kind": "plague",
  "label": "灰疫潮",
  "severity": 3
}
```

## 10. GM Editor

页面：

```text
/editor
```

内容 API：

```text
GET  /api/editor/content
GET  /api/editor/overrides
POST /api/editor/content
```

Live Run API：

```text
GET  /api/editor/runs
POST /api/editor/runs/{id}/world
POST /api/editor/runs/{id}/room
```

内容覆盖不会改 Go 源码，保存到：

```text
data/editor/content_overrides.json
```

进程启动时重新应用。

当前可覆盖类型：

```text
items
enemies
events
npcs
dialogues
shops
skills
worldEvents
npcSchedules
questGraphs
```

## 11. AI Tool Boundary

V0.7 Tool：

```text
create_room
connect_rooms
reveal_room
set_flag
grant_item
spawn_enemy
move_npc
set_region_state
trigger_world_event
```

未来 Agent 只负责：

```text
理解玩家意图
规划
选择 Tool
生成叙事
```

Game Runtime 负责：

```text
权限 / 条件校验
世界一致性
随机数
战斗
任务状态
NPC 位置
物品
存档
```

## 12. 下一步建议

V0.8 可以正式接入 AI DM，但仍然建议先以最小闭环接入：

```text
自然语言自由行动
 → Agent 读取 Room / NPC / WorldEvent / QuestGraph
 → 输出结构化 Tool Call
 → Game Runtime 执行
 → Agent 根据 Tool Result 叙述
```

不要让 LLM 直接返回一份替换整个 Run 的 JSON。

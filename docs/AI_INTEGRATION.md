# AI / DM Agent 接入预留

V0.7 **仍然没有接入 LLM**。这不是缺项，而是刻意让 Game Runtime 先成为独立、可玩的规则引擎。

当前已经提供：

- `GET /api/tools`：返回未来 Agent 可使用的 Tool Schema。
- `POST /api/runs/{id}/tools/execute`：无需模型即可测试 Tool 修改世界状态。

已有 Tool：

- `create_room`
- `connect_rooms`
- `reveal_room`
- `set_flag`
- `grant_item`
- `spawn_enemy`

未来推荐增加：

- `create_scene_element`
- `create_npc` / `move_npc`
- `create_quest` / `update_quest`
- `schedule_world_event`
- `create_item_from_effects`
- `change_room_scene`

推荐链路：

```text
玩家自然语言
    ↓
DM Agent：理解意图 + 规划
    ↓ structured tool call
Game Tool Layer
    ↓ validation
Go Game Runtime
    ↓
World State / Event Log / Snapshot
    ↓
前端按结构化状态重新渲染
```

例如玩家说“我敲击墙壁寻找密室”，Agent 最终应该调用已有规则能力，而不是直接输出“你发现了密室”并把它留在聊天文本里。

```text
inspect / check
   ↓ success
create_room
connect_rooms
reveal_room
```

世界因此真正改变，并会进入存档。

## 边界

AI 可以创造和组合**数据**，但不应该在游戏运行时：

- 修改 Go 源码
- 修改前端源码
- 直接覆盖完整存档 JSON
- 绕过 Game Runtime 发放物品或修改 HP

原则仍然是：**代码定义能力，数据定义世界，AI 通过受控 Tool 操纵能力和数据。**


## V0.5 可作为 AI 上下文的新状态

- `NPCRelations`：人物关系与对话选择后果。
- `ActiveDialogue`：当前结构化对话节点。
- `ActiveShop` / `ShopStock`：交易上下文与世界库存。
- `Quests`：由 NPC 主动接受的任务。
- `Player.Growth`：长期战斗精通。

未来 Agent 可以读取这些状态，但仍应通过规则层 API / Tool 修改它们。


## V0.7 进一步提供给 AI 的世界事实

V0.7 在既有 `npcLocations`、`regionStates`、`quest.stage/outcome`、`shopRefresh`、`itemAffixes` 和 `combat.distance` 之外，进一步新增 `worldEvents`、`npcWorld`、`sceneStates`、`questDecisions`、`player.skills`、`combat.terrain/hazards`。这些都应该进入未来 DM Agent 的只读上下文。

Agent 可以基于这些事实决定“说什么”和“调用什么 Tool”，但不能直接覆写事实。例如 NPC 已迁往灰疫医馆时，模型不能仅靠叙述让他继续站在断桥营火；必须通过未来受控的 NPC/World Tool 改变 Run。


## V0.7 Living World Tool

V0.7 新增 `move_npc`、`set_region_state`、`trigger_world_event`。GM Editor 也通过同一类 Runtime 方法修改 Run，而不是直接写存档。内容定义可通过 `/editor` 覆盖并持久化；未来 AI 仍不应拥有源码或文件系统写权限。

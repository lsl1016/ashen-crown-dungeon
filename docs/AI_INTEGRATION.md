# AI / DM Agent 接入预留

V0.3 **仍然没有接入 LLM**。这不是缺项，而是刻意让 Game Runtime 先成为独立、可玩的规则引擎。

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

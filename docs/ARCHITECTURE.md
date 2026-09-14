# V0.4 架构说明

## 总体结构

```text
Browser Game UI
  ├─ Scene Background
  ├─ Player / Hotspot Entity Layer
  ├─ VFX / Float Number / Cinematic Layer
  ├─ World Map (ACT I + ACT II)
  ├─ Intent Combat
  ├─ Build / Talent / Inventory
  └─ Journal / Codex
            │ HTTP
            ▼
Go HTTP API
            │
            ▼
Game Runtime
  ├─ Engine / Effect
  ├─ Combat Intent
  ├─ Status / Equipment / Talent
  ├─ Boss Phase
  ├─ World Clock
  ├─ SceneElement Conditions
  ├─ Two-Act Progression
  ├─ Map Generator
  ├─ Save / Event Log
  └─ Tool Boundary
            │
            ▼
JSON Run / Save persistence
```

## V0.4 的关键架构原则

### 后端仍是唯一事实来源

前端新增了实体移动、受击震屏、飘字、斩击、Boss 演出和 WebAudio，但这些都只是表现层。真实结果仍来自 API 返回的 `Run`：

```text
HP / Energy
Combat Intent
Boss Phase
Inventory
Flags
Room Locked / Discovered
Quest Status
Victory / GameOver
```

因此动画失败不会改变规则，刷新页面也不会丢失世界事实。

### 双幕不是两套游戏状态

ACT II 直接追加到同一个 `Rooms / Edges / Flags / Quests` 中。第一幕 Boss 只改变世界状态并解锁 `room_33`，最终黑门再解锁 `room_44`。这使后续 AI Agent 可以使用同一套 Tool 操纵整个 Campaign。

### 内容继续数据驱动

- `content.go`：基础内容。
- `content_v04.go`：第二幕物品、敌人、Lore、事件、地图模板、场景映射与 SceneElement。
- `generator.go`：把模板生成成确定性地图状态。
- `engine.go`：执行行为和世界规则。
- `systems_v03.go`：通用 Build / Intent / Status / Boss / Clock 系统，V0.4 继续复用。

## 为什么 V0.4 仍不用 React/npm

当前仍采用静态 HTML/CSS/JS + Go 单进程托管：

1. Windows 解压即可运行，不需要 Node/npm。
2. 规则、内容和玩法仍在快速迭代，减少构建链复杂度。
3. 当前实体 / VFX / Cinematic 已能验证 Steam 风格表现方向。
4. 后续切换 React + PixiJS + Tauri 时，HTTP API 与 Game Runtime 不需要推翻。

## AI 边界

未来 Agent 不能获得源码写权限：

```text
自然语言
→ Agent 解析 / 规划
→ Tool Call
→ Runtime 校验
→ World State
→ 前端播放结果
```

AI 只能通过受控能力改变数据；不能直接覆盖存档或伪造前端动画来制造不存在的世界事实。

# V0.3 架构说明

## 总体结构

```text
Browser Game UI
  ├─ Scene / Hotspot
  ├─ Map
  ├─ Intent Combat
  ├─ Build / Talent
  ├─ Inventory / Quest
  └─ Journal / Codex
            │ HTTP
            ▼
Go HTTP API
            │
            ▼
Game Runtime
  ├─ Engine              主流程
  ├─ Rule / Effect       检定与效果
  ├─ Combat Intent       敌人行为
  ├─ Status              Burn / Bleed / Weakened
  ├─ Boss Phase          三阶段状态
  ├─ Talent / Equipment  Build
  ├─ World Clock         持续世界演化
  ├─ SceneElement        条件场景交互
  ├─ Map Generator       Seed 地图
  ├─ Save / Event Log
  └─ Tool Boundary       未来 DM Agent
            │
            ▼
JSON Run / Save persistence
```

## 后端文件职责

- `engine.go`：创建 Run、移动、事件、战斗入口、物品、交互、胜负与 Effect。
- `systems_v03.go`：V0.3 的天赋、装备、状态、敌人 Intent、Boss Phase、WorldClock。
- `generator.go`：32 节点地图结构、场景元素和条件机关。
- `content.go`：职业、天赋、物品、敌人、事件和 Lore。
- `tools.go`：未来 Agent 允许调用的最小世界变更能力。
- `store.go`：自动 Run、手动 Save 与加载。

## 为什么暂时仍然不用 React/npm

V0.3 继续保持静态 HTML/CSS/JS + Go 单进程托管，是为了：

1. 用户解压即可运行，不需要 Node/npm。
2. 先验证 Game Runtime 和玩法，而不是把时间消耗在构建链。
3. Windows / Linux / macOS 可以直接分发单个后端二进制 + web 目录。
4. 后续升级 React + PixiJS + Tauri 时，HTTP API 和 Game Runtime 可以保持不变。

这不是最终 Steam 技术栈限制；只是当前原型阶段的可运行性优先策略。

## 状态归属

真正世界事实只存在于 `Run`：

```text
Player
Rooms / Edges
Flags
Quests
Lore
Combat
WorldClock
EventLog
RNG/Turn state
```

前端只显示后端状态，不自行决定骰点、奖励、Boss 阶段或地图解锁。

## AI 边界

未来 Agent 不能获得源码写权限。Agent 的职责是：

```text
理解玩家自然语言
→ 选择 / 组合 Tool
→ Game Runtime 校验
→ 保存真实世界状态
→ Agent 根据执行结果叙事
```

这使“AI 说发生了什么”和“世界实际发生了什么”保持一致。

# 灰烬王冠：沉眠墓城 v0.7.0 — Living World

一个不依赖 AI 也能完整运行的本地 Dungeon / TRPG RPG 原型。

V0.7 在 V0.6 的双幕战役、持续世界、NPC、任务、商店、装备词条和距离战斗之上，继续把游戏推进成一个**可模拟、可编辑、未来可由 Agent 操作的 Living World Runtime**：世界事件会随钟声出现和恶化，NPC 拥有结构化日程，支线任务出现真正的分叉结局，职业技能改成数据驱动，战斗场景拥有地形危险带，并加入本地 GM Editor。

> 当前版本仍然**没有接入任何 LLM**。世界事实、规则校验、随机数、任务状态、NPC 位置、战斗结果和内容持久化都由 Game Runtime 决定。未来 DM Agent 只能通过受控 Tool 操作这些能力。

## 直接运行

### Windows

解压后双击：

```bat
start-windows.bat
```

然后打开：

```text
http://localhost:8080
```

GM 编辑器：

```text
http://localhost:8080/editor
```

压缩包内已经包含 Windows amd64 可执行文件，不需要安装 Go、Node 或 npm。

### Linux

```bash
chmod +x start-linux.sh dist/ashen-crown-linux-amd64
./start-linux.sh
```

### macOS Apple Silicon

```bash
chmod +x start-macos-arm64.sh dist/ashen-crown-macos-arm64
./start-macos-arm64.sh
```

### 源码运行

```bash
go run ./cmd/server -web ./web -data ./data
```

---

# V0.7 核心变化

## 1. 世界事件调度器

世界不再只通过固定 Flag 变化。V0.7 新增结构化 `WorldEventDef / WorldEventState`：

```text
第 4 钟
  ↓
灰疫潮出现
  ↓
外墓区场景进入 plague 状态
  ↓
NPC 行动 / 区域描述 / 战斗危险带改变
  ↓
第 7 钟仍未解决
  ↓
事件升级
  ↓
玩家完成对应任务分支
  ↓
世界事件结束并留下持久后果
```

当前世界事件包括：

- 灰疫潮
- 无名游行
- 王庭猎名
- 灰暴锋线

事件拥有：

```text
triggerBell
escalateBell
requireFlag
resolveFlag
severity
overlay
region
```

同一份定义会同时被 Game Runtime、前端世界事件面板、GM Editor 和未来 Agent 使用。

## 2. NPC 结构化日程

NPC 不再只有一个“当前房间”。V0.7 新增：

```text
NPCScheduleDef
NPCWorldState
```

NPC 世界事实包括：

```text
location
activity
health
mood
knowledge
```

例如伊文会随着钟声在断桥营火、灰疫医馆之间行动；任务结果又可以改变其最终去向。

GM / 未来 AI 可以通过受控操作临时固定 NPC 到某个地点：

```text
NPCOverrides
```

该 Override 会真实进入存档，不会被下一次自动日程刷新覆盖；也可以恢复为自动日程。

## 3. 真正的支线分叉任务图

三条主要支线已经从单线：

```text
接受 → 目标 → 交付 → 奖励
```

升级成 `QuestGraph`：

```text
              ┌→ 公开赛勒记录 → 重建巡夜哨线
巡夜队任务 ──┤
              └→ 封存记录     → 表面秩序保留 / 真相消失
```

```text
              ┌→ 归还真名 → 无名游行开始恢复姓名
无名囚徒 ─────┤
              └→ 重新封名 → 游行停止 / 真名再次被抹去
```

```text
              ┌→ 共享路线 → 逐风者掌握安全风道
风暴骑士 ─────┤
              └→ 隐瞒路线 → 商路开放 / 阵营关系恶化
```

Run 新增：

```text
quest.stage
quest.outcome
questDecisions
```

不同结局会影响：

- NPC 关系
- NPC 行动
- 区域状态
- 世界事件是否解决
- 金币 / 独特物品
- 后续世界叙事

并且已经处理“分支完成后再次对话重复领奖”的问题。

## 4. 数据驱动 Skill / Effect

职业技能不再把每一个行为写死在 `switch` 里。

结构示例：

```json
{
  "id": "warden_chainbreaker",
  "class": "warden",
  "name": "断链冲锋",
  "cost": 2,
  "cooldown": 3,
  "minDistance": 1,
  "maxDistance": 3,
  "hitAttribute": "strength",
  "effects": [
    {"type": "set_distance", "value": 1},
    {"type": "damage", "value": 4, "dice": 6},
    {"type": "enemy_status", "target": "weakened", "rounds": 2}
  ]
}
```

V0.7 内置 6 个结构化技能：

**铁誓守卫**
- 铁誓猛击
- 断链冲锋

**暮影游侠**
- 弱点穿刺
- 灰幕箭

**余烬术士**
- 余烬爆裂
- 护幕星火

目前 Effect Runtime 已支持本版技能所需的：

```text
damage
guard
shield
energy
set_distance
retreat
enemy_status
```

后续可以继续扩展 `heal / poison / summon / teleport / push / pull` 等，而无需为每个新技能重新写一整段战斗流程。

## 5. 战场地形与危险带

V0.6 的近 / 中 / 远距离现在开始和场景发生联系。

例子：

```text
断桥 / 门楼
远距：利用断柱掩体，防御 +2
```

```text
黑水引渠
近距：黑水侵蚀，每回合末受伤
```

```text
冷炉
中距：炉底裂口造成灼灰伤害
```

```text
雾外荒原
中距：灰暴带造成风蚀伤害
```

世界事件还能叠加新的危险带：

- 灰疫潮 → 远距孢雾
- 无名游行 → 中距失名残响
- 王庭猎名 → 近距黑火
- 灰暴锋线 → 远距灰暴

因此推进 / 后撤第一次同时具备：

```text
攻击距离决策
+
敌人 Intent 决策
+
战场地形决策
```

## 6. SceneState：世界变化真正改变场景

每个房间现在都有派生 `SceneState`：

```text
stable
secured
danger
plague
echo
blackfire
storm
```

前端会根据真实场景状态改变：

- 环境 Overlay
- 粒子 / 光效
- 区域标签
- 世界事件展示
- 战斗 Hazard

因此“灰疫潮正在外墓区发生”不再只是一段日志文字。

## 7. 本地 GM Editor

访问：

```text
http://localhost:8080/editor
```

这是 V0.7 最重要的开发基础设施之一。

### 内容库

可以查看、搜索、新建、复制并覆盖：

```text
Item
Enemy
Event
NPC
Dialogue
Shop
Skill
WorldEvent
NPCSchedule
QuestGraph
```

覆盖内容保存到：

```text
data/editor/content_overrides.json
```

服务重启后会自动重新加载，不需要重新编译 Go。

QuestGraph 还带基础节点预览。

### Live Run

GM Editor 可以直接选择本地正在运行的 Run，并执行：

- 推进世界钟
- 强制触发世界事件
- 移动 NPC
- 恢复 NPC 自动日程
- 修改区域状态
- 设置世界 Flag
- 查看当前世界事件
- 查看 NPC 当前行为 / 地点 / 心情
- 查看区域状态
- 查看真实 Flag

### 地点编辑器

可以：

- 编辑现有 Room JSON
- 创建新地点
- 指定地图坐标
- 指定 Zone / Scene / Type
- 添加场景 Element
- 将新地点连接到已有地点
- 查看当前 Run 的实时小地图

GM 页面不会直接编辑存档文件，而是通过受控 HTTP API 调用 Game Runtime。

> GM Editor 当前没有身份认证，只设计给本地开发使用。不要把该服务直接暴露到公网。

## 8. V0.7 为未来 AI 增加的 Tool Boundary

除原有：

```text
create_room
connect_rooms
reveal_room
set_flag
grant_item
spawn_enemy
```

新增：

```text
move_npc
set_region_state
trigger_world_event
```

未来 DM Agent 依然不允许直接修改 Go / JS / 存档 JSON。

---

# 当前内容规模

```text
44   地图地点
133  场景交互元素
35   剧情 / 随机事件
21   类敌人 / Boss
48   种物品
18   个职业天赋
8    种装备词条
6    个数据驱动职业技能
5    个持久 NPC
5    棵 NPC 对话树
5    套 NPC 日程定义
3    棵分支 QuestGraph
4    个 Living World Event
2    套动态商店
7    个持续变化区域
3    档战斗距离
104  个本地 SVG 资源
```

V0.4–V0.6 已有的双幕 Campaign、Boss 三阶段、敌人 Intent、角色动画、NPC 立绘、商店、关系、成长、装备 Build、词条、存档和 World Seed 均保留。

---

# 操作方式

## 探索

- 点击场景热点调查。
- 点击 NPC 立绘开始对话。
- 底部行动栏提供显式操作入口。
- 右侧地图脉冲节点可移动。
- 部分互动需要物品、Flag、属性检定或前置世界状态。

## 战斗

战斗按钮现在会根据角色真正拥有的 `player.skills` 动态生成。

基础行动包括：

```text
普通攻击
职业技能 × 2
防御
推进
后撤
使用消耗品
尝试脱离
```

应综合：

```text
敌人 NEXT INTENT
当前近 / 中 / 远距离
Battlefield Terrain
当前 Hazard
自身 / 敌人状态
装备和词条
职业天赋
技能 Cost / Cooldown / Range
```

做决策。

---

# 项目结构

```text
ashen-crown-dungeon/
├── cmd/server/
│   └── main.go
├── internal/game/
│   ├── content.go
│   ├── content_v04.go
│   ├── content_v05.go
│   ├── content_v06.go
│   ├── content_v07.go         # Skill / WorldEvent / Schedule / QuestGraph
│   ├── editor.go              # 内容覆盖 + GM 地点操作
│   ├── engine.go
│   ├── generator.go
│   ├── systems_v03.go
│   ├── systems_v05.go
│   ├── systems_v06.go
│   ├── systems_v07.go         # Living World + Skill Runtime + Terrain + GM
│   ├── tools.go
│   ├── types.go
│   └── store.go
├── internal/httpapi/
│   └── server.go
├── web/
│   ├── index.html
│   ├── app.js
│   ├── styles.css
│   ├── editor.html
│   ├── editor.js
│   ├── editor.css
│   └── assets/
├── docs/
│   ├── V05_SYSTEMS.md
│   ├── V06_SYSTEMS.md
│   ├── V07_SYSTEMS.md
│   ├── AI_INTEGRATION.md
│   └── ...
├── data/
├── dist/
├── Dockerfile
├── docker-compose.yml
├── VERSION
├── start-windows.bat
├── start-linux.sh
└── start-macos-arm64.sh
```

---

# HTTP API

游戏接口：

```text
GET  /api/health
GET  /api/world
GET  /api/tools
POST /api/runs
GET  /api/runs/{id}
POST /api/runs/{id}/move
POST /api/runs/{id}/interact
POST /api/runs/{id}/action
POST /api/runs/{id}/combat
POST /api/runs/{id}/item
POST /api/runs/{id}/npc
POST /api/runs/{id}/dialogue
POST /api/runs/{id}/shop
POST /api/runs/{id}/growth
POST /api/runs/{id}/talent
POST /api/runs/{id}/save
POST /api/runs/{id}/tools/execute
GET  /api/saves
POST /api/saves/{id}/load
```

GM Editor：

```text
GET  /api/editor/content
GET  /api/editor/overrides
GET  /api/editor/runs
POST /api/editor/content
POST /api/editor/runs/{id}/world
POST /api/editor/runs/{id}/room
```

---

# 存档兼容

`Engine.Prepare(run)` 会继续兼容 V0.5 / V0.6 Run，并补齐 V0.7 字段：

```text
player.skills
worldEvents
npcWorld
sceneStates
questDecisions
npcOverrides
regionOverrides
combat.terrain
combat.hazards
```

旧存档可以继续加载，但建议新版本从新 Run 体验完整 Living World，因为过去已经发生过的历史事件无法完全逆推出新的世界事件阶段。

---

# AI 接入原则

正式接 AI 后保持：

```text
玩家自然语言
    ↓
DM Agent
    ↓
读取真实 Run / World State
    ↓
结构化意图 / Tool Call
    ↓
Game Runtime 校验
    ↓
Run / World State 改变
    ↓
前端表现
```

AI 不直接：

- 修改 Go / JavaScript
- 写存档 JSON
- 自己生成骰子结果
- 自己声明 NPC 已经移动
- 自己宣告玩家得到物品
- 自己决定不存在的世界事实

详见 `docs/AI_INTEGRATION.md` 与 `docs/V07_SYSTEMS.md`。

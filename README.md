# 灰烬王冠：沉眠墓城 v0.3.0

一个**完全不依赖 AI 也能运行和游玩的 Dungeon / TRPG 原型**。V0.3 基于 V0.2 增量升级，不推翻已有 Game Runtime；重点从“能点击、能探索”推进到“有战术、有 Build、有条件机关、有持续世界变化”。

> 当前版本仍未接入任何 LLM。AI DM 的 Tool Boundary 已保留，等游戏本身稳定后再接自然语言理解与内容编排。

## V0.3 新增的核心玩法

### 1. 意图制战斗

敌人不再每回合固定普攻。每个敌人都有行为模板，下一回合行动会提前展示：

- 普通攻击
- 蓄力重击
- 连续突袭
- 防御姿态
- 诅咒 / 汲取
- 冲锋

玩家可以看到 `NEXT INTENT` 再决定攻击、防御、技能或战斗道具。Boss「烬冠承载者」拥有三阶段状态变化，60% / 30% 生命会进入新的阶段并改变防御、状态与行动节奏。

### 2. 角色 Build

三个职业各有 4 个天赋，共 **12 个职业天赋**。创建角色获得 1 点初始天赋点，升级继续获得天赋点。

- 铁誓守卫：格挡、重击、生命、反击
- 暮影游侠：暴击、流血、闪避、搜刮
- 余烬术士：燃烧、能量、汲取、护幕

装备从单武器升级为：

```text
武器 + 护甲 + 饰品
```

装备会真实修改攻击、防御、最大生命、最大能量或属性，不只是 UI 文案。

### 3. 条件式场景交互

场景元素现在支持：

```text
需要物品
需要世界 Flag
隐藏直到满足条件
消耗钥匙
D20 属性检定
成功效果
失败效果
解锁地图
触发战斗
```

例如：

- 巡夜徽章才能打开门楼储物柜
- 抄写员印章才能打开医馆封存药柜
- 先排干黑水，才能搜寻水下背包
- 断月桥需要敏捷检定才能通过
- 地下星仪需要地图 + 感知检定才能开启王室封藏间
- 王印才能开启王室封藏箱
- 骨钥匙才能开启白石王陵石棺

### 4. 世界钟声

探索移动会推进墓城时间。每 5 次移动，钟声前进一步：

```text
I → IV → VII → X → XIII
```

钟声越靠后，世界威胁越高，后续敌人的生命、防御会被强化，并出现世界异动 Flag。UI 会持续显示当前钟声、威胁等级和最近一次世界变化。

## 当前内容规模

- 1 个完整 Campaign：《灰烬王冠：沉眠墓城》
- 3 个职业
- **32 个地图地点**
- **29 套本地 SVG 场景背景**
- **94 个场景交互元素**
- **28 个剧情 / 随机事件**
- **15 类敌人**，含 Boss
- **31 种物品**
- **12 个职业天赋**
- 武器 / 护甲 / 饰品三槽装备
- NPC、商人、余火休整点、支线任务、隐藏区域
- D20 属性检定、真实后端骰点
- 世界 Codex、冒险 Journal、Event Log
- 自动持久化 + 手动存档 / 读档
- World Seed 确定性生成
- 未来 AI Tool Schema

## 新区域

V0.3 在 V0.2 的礼拜堂、焚书馆、沉水钟厅、旧王囚室、灰玫瑰庭、千骨堂、冷炉工坊、观测台、宴厅等基础上新增：

- 灰疫医馆
- 折冠门楼
- 黑水引渠
- 断月桥
- 灰烬审判庭
- 无钟之塔
- 王室封藏间
- 白石王陵

每个新增区域都有独立场景 SVG 和自己的交互规则，不是单纯复制旧房间。

## 最简单运行方式

### Windows x64

无需 Go、Node 或 npm。解压后双击：

```text
start-windows.bat
```

浏览器访问：

```text
http://localhost:8080
```

### macOS Apple Silicon

```bash
chmod +x start-macos-arm64.sh
./start-macos-arm64.sh
```

### Linux x64

```bash
chmod +x start-linux.sh
./start-linux.sh
```

### 从源码运行

需要 Go 1.23+，不需要 npm：

```bash
go run ./cmd/server
```

## 操作说明

1. 创建角色并选择职业。
2. 开局先点右上角“天赋”，选择第一项 Build 能力。
3. 场景中发光热点可以直接点击；底部行动区也会列出同样的交互。
4. 条件未满足的交互会显示锁定原因，不需要猜谜式试错。
5. 地图上有脉冲的节点可以直接移动，右侧“附近道路”也可前往。
6. 战斗时先看敌人 `NEXT INTENT`，再决定行动。
7. 战斗消耗品会真正消耗一个回合，不能免费连续使用。
8. 数字键 `1~9` 可触发当前行动。
9. 右上角 `▣` 可手动存档，Run 本身也会自动持久化。

## 推荐 Seed

```text
20260913
260914
4242
31713
```

同一 Seed 下地图捷径、部分敌人、事件与骰点派生可以稳定复现，方便调试和后续 Agent Replay。

## 项目结构

```text
cmd/server/                 Go 启动入口
internal/game/
  engine.go                 主 Game Runtime
  systems_v03.go            天赋、装备、状态、敌人意图、Boss 阶段、世界钟声
  generator.go              32 地点地图与场景交互定义
  content.go                职业、天赋、物品、敌人、事件、Lore
  tools.go                  未来 AI Tool Boundary
  store.go                  Run / Save 持久化
internal/httpapi/           HTTP API + 静态资源托管
web/                        零 npm 依赖游戏前端
web/assets/scenes/          29 套本地 SVG 场景背景
docs/screenshots/           V0.3 UI 截图
data/runs/                  自动持久化 Run
data/saves/                 手动存档
docs/                       世界观、架构、游戏设计、AI 接入说明
```

## 关键 API

```text
POST /api/runs                         创建游戏
GET  /api/runs/{id}                    当前游戏状态
POST /api/runs/{id}/move               地图移动
POST /api/runs/{id}/interact           场景元素交互
POST /api/runs/{id}/action             事件选择
POST /api/runs/{id}/combat             战斗行动
POST /api/runs/{id}/item               使用 / 装备物品
POST /api/runs/{id}/talent             学习职业天赋
POST /api/runs/{id}/save               手动存档
GET  /api/saves                        存档列表
GET  /api/tools                        AI Tool Schema
POST /api/runs/{id}/tools/execute      Tool 调试执行
```

## AI 接入原则

后续仍坚持：

> **代码定义能力，数据定义世界，AI 通过 Tool 操纵能力与世界。**

```text
玩家自然语言
    ↓
DM Agent
    ↓
结构化 Tool Call
    ↓
Go Game Runtime 校验
    ↓
地图 / NPC / 战斗 / 任务 / 物品 / 世界 Flag
    ↓
持久化 World State
    ↓
前端重新渲染
```

AI 不需要、也不应该在运行时修改 Go / JavaScript 源码。

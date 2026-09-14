# 灰烬王冠：沉眠墓城 v0.6.0

一个不依赖 AI 也能完整运行的本地 Dungeon / TRPG RPG 原型。V0.6 在 V0.5 的双幕战役、NPC 对话、商店和成长系统之上，继续加入**持续世界模拟**：NPC 会迁移、区域会改变状态、任务有阶段与回报结局、商店随钟声补货、装备会出现确定性词条，战斗加入近 / 中 / 远三档站位。

> 当前版本刻意没有接入 LLM。Game Runtime 先负责事实、规则和持久化；未来 DM Agent 只通过受控 Tool 解释自然语言并操作这些能力，而不是运行时修改代码。

## 直接运行

### Windows

解压后双击：

```bat
start-windows.bat
```

然后访问：

```text
http://localhost:8080
```

压缩包内已经包含 Windows amd64 可执行文件，不需要 Go、Node 或 npm。

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

也可以从源码运行：

```bash
go run ./cmd/server -web ./web -data ./data
```

## V0.6 主要变化

### 1. NPC 日程和迁移

NPC 不再永远固定在出生房间。位置记录在 `Run.npcLocations`，会根据世界钟声和任务结果变化。例如：

- 伊文：断桥营火 → 第七钟后的灰疫医馆 → 巡夜任务交付后的墓城入口。
- 帷面客：无脸商铺 → 第二幕开放且钟声推进后迁往雾外界碑。
- 无名囚徒：旧王囚室 → 找回名字后前往巡夜营地。
- 奥林：无旗骑士营 → 宣誓或黑门开启后前往外封印黑门。
- 玛拉：逐风营火 → 风暴骑士任务回报后前往雾外界碑。

场景中的 NPC 立绘和底部行动按钮都从这份状态动态生成，因此 NPC 搬走以后旧位置不会留下一个“假 NPC”。

### 2. 区域状态演化

`Run.regionStates` 保存 7 个区域的当前状态，例如：

```text
外墓区    灰雾封锁
旧城区    残响不稳定
内廷      王庭警戒
雾外荒原  尚未抵达
荒原东界  黑门封闭
```

随着钟声、Boss、支线和黑门推进，这些状态会持久变化，例如：

```text
外墓区
灰雾封锁
  ↓ 第七钟
灰雾加深 · 巡夜线失联
  ↓ 完成并交付巡夜支线
巡夜团重新建立哨线
```

右侧世界脉冲、附近道路和世界信息面板都会读取真实区域状态。

### 3. 真正的任务链阶段和结局

任务不再只有 `active / completed`，还包含：

```text
stage
outcome
```

例如巡夜支线：

```text
接受任务
  ↓
寻找队长
  ↓
完成目标
  ↓
返回交付
  ↓
伊文确认记录
  ↓
结局：巡夜线重建
  ↓
NPC 迁移 + 区域状态变化 + 金币 + 独特饰品
```

目前带回报后果的支线包括巡夜队、无名囚徒和风暴骑士。

### 4. 商店随世界钟刷新

两套商店库存仍然持久化，但现在会按世界钟自动补货：

- 每推进 3 个钟声周期检查一次。
- 普通装备逐步补 1 件。
- 消耗品最多补 2 件。
- 第十三钟停止后不再自动刷新。
- 商店 UI 会显示最近补货和下一轮补货钟声。

商店位置还会随商人迁移而变化。

### 5. 装备随机词条

V0.6 新增 8 个词条：

- 锋锐：武器威力 +1
- 墓锻：武器威力 +2
- 坚固：防御 +1
- 强韧：最大生命 +4
- 共鸣：最大能量 +2
- 守望：感知 +1
- 不屈：意志 +1
- 逐风：敏捷 +1

精英战利品和部分购买装备会根据 `WorldSeed + Item + 来源` 生成确定性词条。同一个世界可重放，不会因读档变成另一件装备。

当前原型按 `itemId` 保存词条，因此同一基础物品的多个副本共享词条；如果后续进入正式装备刷取阶段，可以进一步升级成独立 `itemInstanceId`。

V0.6 新装备包括：

- 第三巡夜哨坠
- 复名银印
- 空钟披扣
- 墓玻璃长刃
- 残响鳞衣
- 坠星透镜

### 6. 近 / 中 / 远三档战斗站位

战斗增加：

```text
近距 ← 推进 / 后撤 → 中距 ← 推进 / 后撤 → 远距
```

职业默认距离：

- 铁誓守卫：近距
- 暮影游侠：中距
- 余烬术士：远距

站位不是纯 UI：

- 守卫远距离普通攻击会受到命中惩罚。
- 游侠可以通过远距天赋增加命中和伤害。
- 近战型敌人在远距攻击会降低命中和伤害，并可能使用“逼近”。
- Boss 的冲锋会直接压到近距。
- 推进和后撤消耗一个完整战斗回合，敌人仍会执行当前 Intent。

玩家实体会随着距离在场景内移动，距离 UI 也展示三档状态。

### 7. 三阶职业天赋树

V0.5 的 12 个职业天赋保留，同时新增 6 个高阶天赋，总计 18 个，并增加：

```text
Tier
RequiredLevel
Requires
```

部分新天赋：

**铁誓守卫**
- 不动锚誓：防御恢复生命，近距额外防御。
- 铁墙推进：推进时同时防御并恢复能量。

**暮影游侠**
- 雾外长射：远距普通攻击命中 +1、伤害 +3。
- 无痕撤步：后撤时保持防御，到远距恢复能量。

**余烬术士**
- 星火过载：远距职业技基础伤害 +4。
- 残响移步：后撤恢复能量并重建符文护幕。

UI 会明确显示 Tier、等级要求和前置天赋。

## 已保留的 V0.5 / V0.4 能力

- 双幕 44 地点完整 Campaign。
- 133 个场景交互。
- 35 个剧情 / 随机事件。
- 21 类敌人和 Boss，21 套独立战斗立绘。
- 48 种物品（包含 V0.6 新增装备与既有剧情/战斗物品）。
- 5 个持久 NPC、5 棵对话树、关系系统。
- 两套商店，买入 / 卖出 / 库存 / 金币。
- 3 个职业和多帧 Idle / Attack / Cast 动作。
- 敌人 Intent、异常状态、Boss 三阶段。
- 属性点、精通点、职业天赋点。
- 条件场景交互、隐藏房间和机关。
- Seed 确定性地图 / 事件 / 骰子 / 内容生成。
- 自动 Run 持久化、手动存档、读取存档。
- Event Log / Journal / Lore Codex。
- WebAudio 本地合成音效。
- AI Tool Boundary（仍未接模型）。

## 操作方式

### 探索

- 点击场景发光热点调查物件。
- 点击场景 NPC 直接开始对话。
- 底部行动栏提供相同的显式操作入口。
- 右侧地图中的脉冲节点可直接移动。
- 部分机关需要物品、Flag 或属性检定。

### 战斗

```text
1 普通攻击
2 职业技能
3 防御
4 推进
5 后撤
6 使用药剂
7 尝试脱离
```

应结合：

- 敌人下一步 Intent
- 当前近 / 中 / 远距离
- 自身状态
- 职业天赋
- 武器与词条

决定行动。

## 项目结构

```text
ashen-crown-dungeon/
├── cmd/server/
│   └── main.go
├── internal/game/
│   ├── content.go
│   ├── content_v04.go
│   ├── content_v05.go
│   ├── content_v06.go        # V0.6 词条、高阶天赋、任务回报
│   ├── engine.go
│   ├── generator.go
│   ├── systems_v03.go
│   ├── systems_v05.go
│   ├── systems_v06.go        # NPC 日程、区域、刷新、词条、距离系统
│   ├── tools.go
│   ├── types.go
│   └── store.go
├── internal/httpapi/
│   └── server.go
├── web/
│   ├── index.html
│   ├── app.js
│   ├── styles.css
│   └── assets/
│       ├── actors/
│       ├── portraits/
│       └── scenes/
├── docs/
│   ├── V05_SYSTEMS.md
│   ├── V06_SYSTEMS.md
│   ├── AI_INTEGRATION.md
│   └── ...
├── data/
├── dist/
├── Dockerfile
├── docker-compose.yml
├── start-windows.bat
├── start-linux.sh
└── start-macos-arm64.sh
```

## HTTP API

主要接口：

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

## 存档兼容

V0.6 对 V0.5 旧 Run 做运行时归一化：

```text
npcLocations
regionStates
shopRefresh
itemAffixes
quest.stage
quest.outcome
combat.distance
```

旧字段不存在时会自动补默认值。HTTP 修改请求会在写盘前再次归一化，确保由任务、钟声产生的 NPC 迁移和区域变化也真正进入存档。

建议发布版仍从新存档体验完整 V0.6，因为旧存档已经发生过的 V0.5 历史事件无法凭空还原成所有新阶段语义。

## AI 接入边界

正式接 AI 后仍建议保持：

```text
玩家自然语言
    ↓
DM Agent
    ↓
结构化意图 / Tool Call
    ↓
Game Runtime 校验
    ↓
修改 Run / World State
    ↓
前端表现
```

AI 不直接写 Go / JavaScript，不直接改数据库 JSON，也不自行宣告不存在的世界事实。

详见 `docs/AI_INTEGRATION.md` 和 `docs/V06_SYSTEMS.md`。

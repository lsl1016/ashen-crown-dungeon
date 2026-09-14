# Changelog

## 0.7.0-ai-gm-bridge-1.0

在 V0.7.0 Living World Runtime 上新增可供外部 Agent 使用的正式 GM 接入层。

- 新增 19 个 Agent Tool：世界/房间/NPC/任务/战斗检查、目录搜索、最近日志、D20 检定、NPC/区域/世界事件/世界钟/Flag/地图/遭遇/奖励操作。
- 新增统一 `AgentToolDefinitions()`，HTTP、MCP、JSON Schema 与参数校验共用同一事实源。
- 新增 `/api/agent/tools`、`/api/agent/tools/{name}`、`/api/agent/tools/execute`、`/api/agent/audit`。
- 新增真实 `dryRun`：在 Run 副本上执行，不写入持久化存档。
- 新增同 Run 串行锁，降低 Agent 并发修改世界造成竞态的风险。
- high-risk Tool 默认要求 `reason`，可通过 `ASHEN_AGENT_REQUIRE_REASON` 配置。
- 新增 `data/audit/agent-tools.jsonl` 审计日志。
- 新增 Bearer Service Token：`ASHEN_AGENT_TOKEN`。
- 新增 MCP 2026-07-28 stateless `/mcp`，实现 `server/discover`、`tools/list`、`tools/call`、现代 Header 校验和 `Mcp-Param-Run-Id`。
- 新增 JSON Schema 导出器 `go run ./cmd/schemaexport -out ./docs/schemas`，输出 Tool Catalog、Gateway Envelope 和每个 Tool Schema。
- 新增 `docs/AI_GM_GUIDE.md` 与 `docs/MCP_SERVER.md`。
- 新增三套可打包 Skill：Ashen Crown DM、World Director、Encounter Referee。
- 核心游戏版本仍为 V0.7.0；本层不绑定任何 LLM Provider。

## 0.7.0-living-world-editor

V0.7 把 V0.6 的持续状态推进成可调度、可分支、可编辑的 Living World Runtime。

- 新增 4 个结构化世界事件：灰疫潮、无名游行、王庭猎名、灰暴锋线；支持触发、升级、解决和场景 Overlay。
- 新增 `NPCScheduleDef / NPCWorldState`，5 个 NPC 拥有钟声日程、行为、健康、心情和知识状态。
- 三条支线升级为 QuestGraph，均拥有两个真实结局，并通过 `questDecisions` 持久化。
- 修复替代支线结局完成后再次对话可能重复领取奖励的问题。
- 新增 6 个 `SkillDef` 数据驱动技能，战斗支持 `skill:<id>`；技能 Cost、CD、Range、Effect 均来自内容定义。
- 战斗新增 Battlefield Terrain 与 Hazard；桥梁掩体、黑水、冷炉、灰暴及 Living World Event 均会影响回合。
- 新增 `SceneState`，区域与世界事件可以直接改变场景环境表现。
- 新增本地 `/editor` GM Editor：内容库、覆盖持久化、Run 选择、世界事件、NPC 移动、区域 / Flag 修改、Room 创建与连接、实时小地图。
- 新增内容覆盖文件 `data/editor/content_overrides.json`，重启服务后自动加载。
- 新增 GM / AI 持久 Override：`npcOverrides` 与 `regionOverrides`，避免手动世界修改被自动日程覆盖。
- Tool Boundary 新增 `move_npc`、`set_region_state`、`trigger_world_event`。
- 新增 V0.7 自动测试：世界事件生命周期、Quest 分支防重复领奖、数据技能、地形、旧存档补齐、GM 世界操作、内容覆盖。
- 仍未接入任何 LLM。

## 0.6.0-persistent-world-tactics

V0.6 从“RPG 交互完整”继续推进到“世界状态会持续演化”。

- 新增 `NPCLocations`：5 个持久 NPC 会根据钟声、任务结局和剧情 Flag 迁移。
- 新增 `RegionStates`：7 个区域拥有安全 / 高危 / 封锁 / 复苏等持久状态。
- Quest 新增 `Stage / Outcome`，三条支线拥有返回交付和世界后果。
- 两个商店每 3 个世界钟周期自动补货，并保存最近刷新钟声。
- 新增 8 个确定性装备词条和 6 件 V0.6 装备。
- Combat 新增近 / 中 / 远距离、推进 / 后撤、敌人逼近和冲锋压近。
- 职业天赋升级为三级树，新增 6 个依赖站位的高阶天赋，总计 18 个。
- UI 新增区域状态、距离轨、动态 NPC、词条展示、任务阶段中文化、商店刷新提示。
- HTTP 新增 `POST /api/runs/{id}/npc` 支持动态 NPC 直接对话。
- `Engine.Prepare` + HTTP 修改前后归一化，兼容 V0.5 旧存档字段缺失。
- 新增 V0.6 自动测试：NPC 日程、区域演化、商店补货、天赋前置、装备词条、战斗距离、任务回报、旧存档归一化。

## 0.5.0-rpg-characters-dialogue

V0.5 目标：在 V0.4 双幕战役上补齐 RPG 面向玩家的核心交互层，而不是继续单纯扩大地图数量。

### NPC / Dialogue / Quest

- 新增 5 个持久核心 NPC：伊文、帷面客、无名囚徒、玛拉、奥林。
- 新增 5 棵结构化多节点对话树。
- 新增 NPC 关系值 `NPCRelations`，玩家选择会真实改变关系。
- 对话可接受 / 拒绝任务、获取 Lore / Flag / 物品或进入商店。
- 场景中的 NPC 由 emoji 热点升级为可点击本地 SVG 立绘实体。

### Shop / Economy

- 新增 2 套持久商店：帷面集市、逐风补给。
- 商店库存进入 Run State。
- 支持真实购买、出售、库存扣减 / 回流、金币结算。
- 已装备物品与关键剧情物品不可直接出售。
- 新增帷面丝线、风玻璃护符、无旗誓戒，总物品数 42。

### Character Growth

- 新增属性点 `AttributePoints` 与精通点 `MasteryPoints`。
- 新增武器训练、职业技精通、幸存者本能 3 条通用精通路线，每条 3 级。
- 三级武器训练提供额外命中；三级职业技精通把技能能量消耗从 3 降为 2；幸存者本能提升生命并在三级增加防御。
- 升级除职业天赋外继续发放属性 / 精通成长资源。
- 右上角“天赋”改为统一“成长”面板。

### Presentation

- 新增 27 个本地 portrait SVG 文件，其中 5 个核心 NPC + 21 类敌人 / Boss 均有独立实际使用立绘，另保留 1 个通用 fallback。
- 新增 36 个职业动作帧：3 职业 × Idle / Attack / Cast × 4 帧。
- 战斗按行动切换真实帧素材，敌人立绘按 Intent 改变姿态反馈。
- 新增三职业专属技能 VFX。
- 场景加入远景 / 中景 / 前景环境层、粒子和轻量视差。
- 新增专用 NPC 对话舞台和商店 UI。

### Engineering / Verification

- 新增 `content_v05.go` 和 `systems_v05.go`，保持 V0.4 地图拓扑与旧系统增量兼容。
- 新增对话任务 / 关系、商店买卖、属性 / 精通、三级技能精通成本等单元测试。
- 实际 HTTP API 联调通过：NPC 对话 → 接任务 → 关系变化 → 商店买卖 → 成长分配。
- 前端通过 Node 语法检查和 HTML / JS 静态 DOM 选择器校验；容器 Chromium 因企业策略阻止访问本地站点，未将该环境限制误判为页面故障。
- 仍未接入任何 LLM。

## 0.4.0-two-act-cinematic

V0.4 目标：在 V0.3 的规则、Build 与条件交互上，补出更接近独立游戏的视觉反馈，并把 Campaign 从一幕墓城扩展成真正的双幕世界。

### Campaign / World

- 地图 32 → **44 个地点**。
- 新增第二幕「雾外荒原」12 个地点。
- 烬冠承载者赫里昂从最终 Boss 调整为第一幕 Boss；击败后不会结束游戏，而是解锁荒原主线。
- 新增最终 Boss「门后之心 · 厄涅玛」，拥有三阶段战斗和多条件结局。
- 新增 12 套荒原场景资产；本地 SVG 场景文件总数达到 41。
- 场景交互元素增加到 **130 个**。
- 新增 7 个荒原剧情 / 随机事件，事件总数 **35**。
- 新增荒原 Lore、世界事实、关键道具与第二幕任务。
- 反向观测台、虚空罗盘、外封印黑门增加可重试的主线推进路径，避免随机检定造成永久软锁。

### Combat / Build

- 新增灰风掠夺者、镜砂行尸、沙丘怨灵、天穹吸血虫、无旗风暴骑士与最终 Boss，共 **21 类敌人**。
- 新增虚空罗盘、星钢长刃、荒原斗篷、镜砂护符、瓶装灰暴、星辉药膏、风誓令牌、门后黑晶，共 **39 种物品**。
- 星钢长刃会对构装型敌人及重击 / 冲锋 Intent 产生额外效果。
- 瓶装灰暴施加弱化并真实降低敌人攻击，同时消耗一个战斗回合。
- 两名 Boss 分别使用独立的阶段转换叙事与状态反馈。

### Presentation

- 新增场景内玩家实体层。
- 点击场景热点时，玩家实体会移动到目标附近再执行交互。
- 新增攻击 / 技能 / 防御 / 受击动作反馈。
- 新增伤害、治疗、能量飘字。
- 新增斩击、爆发、防御 VFX 与受击震屏。
- 新增 Boss 入场、Boss 阶段转换、ACT II 开启、区域切换和最终胜利遮罩演出。
- 世界地图增加 ACT I / ACT II 分区和第二幕节点视觉区分。
- 新增 WebAudio 合成音效，不引入外部音频依赖；可在 UI 中关闭。

### 工程 / 验证

- V0.4 内容独立放入 `content_v04.go`，继续保持增量式架构。
- 新增第一幕 Boss 解锁第二幕、最终 Boss 结算、第二幕初始锁定、最终黑门条件解锁等单元测试。
- 保持 World State 为唯一事实来源，前端动画不自行决定伤害、掉落、检定或地图状态。
- 仍未接入任何 LLM。


## 0.3.0-build-and-world

V0.3 目标：在 V0.2 “可以真正交互”的基础上，让游戏开始具备战术决策、角色 Build、条件机关和持续世界演化。

### 战斗

- 新增敌人意图系统，敌人下一回合行动提前展示。
- 新增多类意图：攻击、重击、连击、防御、诅咒、汲取、冲锋。
- 新增燃烧、流血、弱化等持续状态。
- 职业技加入冷却，不再每回合无脑释放。
- 战斗消耗品进入正式回合经济，使用后敌人会正常行动。
- Boss「烬冠承载者」加入 60% / 30% 三阶段转换。
- 加入 Boss 阶段切换、敌人 / 玩家受击等前端反馈。

### Build

- 新增 12 个职业天赋。
- 新建角色获得 1 天赋点，升级继续获得天赋点。
- 新增 `POST /api/runs/{id}/talent`。
- 装备扩展为武器 / 护甲 / 饰品三槽。
- 物品支持稀有度、Defense、MaxHP、MaxEnergy、属性加成和特殊效果。
- 暮影游侠搜刮、暴击；守卫反击；术士燃烧/护幕等能力均进入后端结算。

### 场景与世界

- 地图 24 → **32 个地点**。
- 场景背景 21 → **29 套本地 SVG**。
- 场景交互元素增加到 **94 个**。
- 新增灰疫医馆、折冠门楼、黑水引渠、断月桥、灰烬审判庭、无钟之塔、王室封藏间、白石王陵。
- `SceneElement` 支持物品条件、Flag 条件、隐藏条件、钥匙消耗、D20 Check、成功 / 失败效果。
- 新增世界钟声 `WorldClock`：移动推进世界时间，威胁等级影响后续敌人强度。
- 新增多个世界异动 Flag 与隐藏区域开启条件。

### UI

- 顶栏增加钟声和天赋点。
- 左侧增加三槽 Build 展示、稀有度和装备效果。
- 敌人舞台增加 Archetype、状态、Boss Phase、NEXT INTENT。
- 地图扩展到 32 节点并明确条件封锁节点。
- 新增墓城异动 / 威胁面板。
- 条件交互在热点和行动栏中都会显示锁定原因与 DC。

### 工程与验证

- 保持 V0.2 Go Runtime 的增量升级路线。
- 新增 `systems_v03.go` 分离 Build / 状态 / Intent / Boss / Clock 规则。
- 增加天赋装备、条件交互、Boss 阶段、世界钟声单元测试。
- 补充真实 API 端到端链路验证。
- 前端执行 Node 语法检查，并使用 Chromium + Mock API 做 DOM / 渲染级检查。
- 仍不接入任何 LLM。

## 0.2.0-interactive

- 场景热点、行动区和地图节点可直接操作。
- 地图扩展到 24 地点、21 场景、67 个交互热点。
- 引入真实 D20、装备、消耗品、Codex、Journal、隐藏房间和支线。
- 保留 AI Tool Layer。

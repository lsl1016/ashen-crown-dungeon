# 灰烬王冠 V0.7.0：AI 操作 GM 完整指南

> 适用版本：Ashen Crown Dungeon V0.7.0 + AI GM Bridge 1.0  
> 目标：让任意 Agent Runtime 通过受控 Tool 操作真实 Game Runtime，而不是靠聊天文本“假装”世界发生变化。

## 1. 核心原则

AI Dungeon Master 不直接修改源码、不直接覆写 Run JSON，也不直接调用 `/editor` 的全部管理员接口。

推荐边界：

```text
玩家自然语言
    ↓
DM Agent
    ↓ 先读事实
AI Tool Gateway / MCP Server
    ↓ JSON Schema 校验 + 风险控制 + Run Lock + Audit
Game Runtime
    ↓
Run / World State / Event Log
    ↓
Web 游戏界面重新渲染
```

人类 GM 仍然可以使用：

```text
/editor
```

AI 使用：

```text
/api/agent/*
/mcp
```

这样可以把权限分开：

- `/editor`：管理员内容编辑、Run 操作。
- `/api/agent/*`：面向 Agent 的受控 HTTP Tool Gateway。
- `/mcp`：同一批 Tool 的 MCP 2026-07-28 暴露层。

原则：**代码定义能力，数据定义世界，AI 通过受控 Tool 操作世界。**

---

## 2. 目录与实现位置

核心实现：

```text
internal/game/agent_tools.go       Agent Tool 定义 + 执行逻辑
internal/game/agent_schema.go      JSON Schema 参数校验
internal/agentgateway/gateway.go   Run Lock / dryRun / Audit / Persistence
internal/httpapi/agent.go          HTTP Tool Gateway
internal/httpapi/mcp.go            MCP 2026-07-28 Streamable HTTP 子集
cmd/schemaexport/main.go           Schema 导出器

docs/schemas/
├── agent-tools.json
├── tool-gateway-execute.schema.json
└── tools/*.json
```

所有入口共享：

```go
game.AgentToolDefinitions()
```

这意味着：

```text
HTTP Schema
MCP tools/list
参数校验
文档 Schema
```

都来自同一事实源。

---

## 3. Agent 的标准 ReAct / GM 循环

每一轮推荐严格执行：

```text
1. Inspect
2. Understand
3. Plan
4. Rule Check（必要时）
5. dryRun（高影响修改）
6. Mutate
7. Re-inspect
8. Narrate
```

### 3.1 Inspect

至少先调用：

```text
inspect_world
inspect_room
```

需要更多信息再调用：

```text
inspect_npc
inspect_quest
inspect_combat
list_world_events
recent_log
search_catalog
```

不要凭模型记忆猜：

- NPC 当前在哪。
- 房间是否已发现。
- 世界事件是否已触发。
- 某个 Enemy / Item 的合法 ID。
- 玩家是否已经获得某件物品。

### 3.2 Rule Check

如果玩家做的是不确定行为：

```text
roll_check
```

骰子必须由 Game Runtime 产生。

错误：

```text
AI：你掷出了 18，成功。
```

正确：

```text
roll_check(attribute="perception", dc=14)
→ Runtime 返回真实 D20
→ AI 根据返回结果叙事
```

### 3.3 Mutation

修改真实世界时使用受控 Tool，例如：

```text
move_npc
set_region_state
trigger_world_event
advance_bell
set_flag
create_room
connect_rooms
reveal_room
spawn_enemy
grant_item
```

### 3.4 Re-inspect

高影响操作之后重新读取事实，例如：

```text
move_npc
→ inspect_npc
```

```text
trigger_world_event
→ list_world_events
→ inspect_world
```

只有确认 Runtime 真正执行成功之后，才向玩家叙述结果。

---

## 4. 工具分层

### Inspect：只读

| Tool | 用途 | Risk |
|---|---|---|
| `inspect_world` | 当前世界、玩家、任务、事件 | low |
| `inspect_room` | 房间、道路、元素、NPC | low |
| `inspect_npc` | NPC 地点、关系、状态、日程 | low |
| `inspect_quest` | 任务阶段 / outcome / QuestGraph | low |
| `inspect_combat` | 当前战斗、Intent、距离、Hazard | low |
| `list_world_events` | Living World Event 定义与状态 | low |
| `search_catalog` | 查询合法 room/npc/enemy/item/skill/world_event ID | low |
| `recent_log` | 最近真实 Event Log | low |

### Rule：规则裁决

| Tool | 用途 | Risk |
|---|---|---|
| `roll_check` | D20 属性检定 | medium |

### World Mutation：世界修改

| Tool | 用途 | Risk |
|---|---|---|
| `move_npc` | 移动 NPC + 写 Override | medium |
| `set_region_state` | 修改持续区域状态 | high |
| `trigger_world_event` | 强制触发 Living World Event | high |
| `advance_bell` | 推进世界钟 1-3 次 | high |
| `set_flag` | 设置已有 Flag 或 `ai.*` Flag | medium |

### Map Mutation

| Tool | 用途 | Risk |
|---|---|---|
| `reveal_room` | 发现已有房间 | medium |
| `create_room` | 创建真实新地点 | high |
| `connect_rooms` | 创建道路 | medium |

### Encounter / Reward

| Tool | 用途 | Risk |
|---|---|---|
| `spawn_enemy` | 触发预定义敌人遭遇 | high |
| `grant_item` | 发放预定义物品 | high |

---

## 5. HTTP AI Tool Gateway

### 5.1 启动

本地开发可以直接：

```bash
./dist/ashen-crown-linux-amd64
```

建议服务端配置 Token：

```bash
export ASHEN_AGENT_TOKEN="replace-with-secret"
export ASHEN_AGENT_REQUIRE_REASON="true"
./dist/ashen-crown-linux-amd64
```

如果 `ASHEN_AGENT_TOKEN` 为空，本地开发模式不会校验 Bearer Token。

公网环境不要留空。

### 5.2 获取工具

```http
GET /api/agent/tools
Authorization: Bearer <token>
```

返回 Tool Definition：

```json
{
  "gatewayVersion": "1.0",
  "gameVersion": "0.7.0",
  "tools": [
    {
      "name": "inspect_world",
      "description": "...",
      "inputSchema": {},
      "outputSchema": {},
      "annotations": {},
      "_meta": {}
    }
  ]
}
```

读取单个 Tool：

```http
GET /api/agent/tools/inspect_world
```

### 5.3 执行 Tool

```http
POST /api/agent/tools/execute
Authorization: Bearer <token>
X-Agent-Name: dm-agent
Content-Type: application/json
```

Body：

```json
{
  "tool": "inspect_world",
  "arguments": {
    "runId": "run_xxx",
    "visibility": "gm"
  }
}
```

结果：

```json
{
  "result": {
    "tool": "inspect_world",
    "runId": "run_xxx",
    "changed": false,
    "risk": "low",
    "auditId": "audit_xxx",
    "summary": "第 4 钟；玩家位于断桥营火；世界事件 1 个。",
    "data": {}
  }
}
```

---

## 6. dryRun

所有 Mutation Schema 都支持：

```json
{
  "dryRun": true,
  "reason": "为什么需要做这次修改"
}
```

`dryRun=true` 时：

```text
Load Run
↓
Clone Run
↓
Game Runtime 真执行
↓
返回结果
↓
不 SaveRun
```

所以 dryRun 不是前端假预览。

例如：

```json
{
  "tool": "advance_bell",
  "arguments": {
    "runId": "run_xxx",
    "steps": 2,
    "dryRun": true,
    "reason": "预览灰疫潮是否会在两个钟声后恶化"
  }
}
```

确认结果符合剧情后，再去掉 `dryRun` 执行。

默认配置下 high-risk Tool 必须提供 `reason`。

---

## 7. Audit

所有 Agent Tool Call 写入：

```text
data/audit/agent-tools.jsonl
```

字段：

```text
id
timestamp
durationMs
success
tool
runId
risk
dryRun
changed
reason
caller
arguments
summary
error
```

查询：

```http
GET /api/agent/audit?runId=run_xxx&limit=50
```

这使得后续可以回答：

> 为什么伊文会出现在灰疫医馆？

而不是依靠 LLM 自己回忆。

---

## 8. JSON Schema

Schema 为 JSON Schema Draft 2020-12。

生成：

```bash
go run ./cmd/schemaexport -out ./docs/schemas
```

输出：

```text
docs/schemas/agent-tools.json
docs/schemas/tool-gateway-execute.schema.json
docs/schemas/tools/<tool>.json
```

### runId

每个 Tool 的 Schema 都显式要求：

```json
{
  "runId": {
    "type": "string",
    "minLength": 1,
    "x-mcp-header": "Run-Id"
  }
}
```

这是刻意设计的。

Game Run 是应用状态句柄，而不是 MCP Session。

---

## 9. MCP Server

端点：

```text
POST /mcp
```

当前内置实现针对：

```text
MCP 2026-07-28
```

实现：

```text
server/discover
tools/list
tools/call
```

当前没有实现：

```text
prompts
resources
tasks
subscriptions
legacy initialize/session
```

这个游戏 Tool Call 都是短同步调用，因此当前不需要 Tasks。

### 9.1 为什么 MCP 自身无 Session

2026-07-28 协议已经取消 `initialize` / `initialized` 和 `Mcp-Session-Id`。

每个请求都自己携带：

```text
protocolVersion
clientCapabilities
clientInfo (optional)
```

游戏状态使用显式：

```text
runId
```

所以：

```text
MCP Stateless Transport
        ↓
runId
        ↓
Persistent Game Run
```

两者不冲突。

### 9.2 server/discover

请求头：

```text
MCP-Protocol-Version: 2026-07-28
Mcp-Method: server/discover
```

请求：

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "server/discover",
  "params": {
    "_meta": {
      "io.modelcontextprotocol/protocolVersion": "2026-07-28",
      "io.modelcontextprotocol/clientCapabilities": {},
      "io.modelcontextprotocol/clientInfo": {
        "name": "my-agent",
        "version": "1.0.0"
      }
    }
  }
}
```

### 9.3 tools/list

Header：

```text
MCP-Protocol-Version: 2026-07-28
Mcp-Method: tools/list
```

响应携带：

```text
resultType=complete
ttlMs=60000
cacheScope=private
```

工具顺序保持确定性，方便模型 Prompt Cache。

### 9.4 tools/call

例如：

```http
POST /mcp
MCP-Protocol-Version: 2026-07-28
Mcp-Method: tools/call
Mcp-Name: inspect_world
Mcp-Param-Run-Id: run_xxx
Authorization: Bearer <token>
```

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "inspect_world",
    "arguments": {
      "runId": "run_xxx",
      "visibility": "gm"
    },
    "_meta": {
      "io.modelcontextprotocol/protocolVersion": "2026-07-28",
      "io.modelcontextprotocol/clientCapabilities": {},
      "io.modelcontextprotocol/clientInfo": {
        "name": "dm-agent",
        "version": "1.0"
      }
    }
  }
}
```

`runId` 使用了 Schema：

```json
"x-mcp-header": "Run-Id"
```

因此现代 MCP Client 应同步发送：

```text
Mcp-Param-Run-Id
```

服务端会校验 Header 与 Body 是否一致。

### 9.5 Tool Result

成功：

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "resultType": "complete",
    "content": [
      {
        "type": "text",
        "text": "第 4 钟；玩家位于断桥营火；世界事件 1 个。"
      }
    ],
    "structuredContent": {
      "tool": "inspect_world",
      "runId": "run_xxx",
      "changed": false,
      "summary": "...",
      "data": {}
    },
    "isError": false
  }
}
```

规则层拒绝（例如错误 NPC ID）作为 Tool Error 返回：

```json
{
  "resultType": "complete",
  "content": [
    {
      "type": "text",
      "text": "未知 NPC: xxx"
    }
  ],
  "isError": true
}
```

协议错误、Header mismatch、未知 Tool 等使用 JSON-RPC Error。

---

## 10. 推荐 DM System Prompt

```text
你是灰烬王冠世界的 Dungeon Master。
Game Runtime 是世界事实的唯一来源。

规则：
1. 在决定世界事实前先 inspect。
2. 不得自行掷骰，使用 roll_check。
3. 不得声称世界发生了改变，除非对应 mutation Tool 已成功执行。
4. 不确定实体 ID 时使用 search_catalog。
5. 需要解释历史原因时使用 recent_log。
6. high-risk Tool 先 dryRun，再根据结果决定是否真正执行。
7. 不要使用 grant_item 绕过任务、战斗或商店规则。
8. 不要为了让剧情符合预期而篡改失败的骰子结果。
9. mutation 后重新 inspect，确认最终事实。
10. 最终回复面向玩家写叙事，不暴露内部 Tool 名和隐藏 Flag。
```

更完整版本见：

```text
examples/agent-system-prompt.md
```

---

## 11. 常见工作流

### 11.1 玩家寻找隐藏房间

玩家：

> 我敲击墙壁，听听后面是不是空的。

Agent：

```text
inspect_room
↓
roll_check(perception)
↓ success
create_room(dryRun=true)
↓
create_room
↓
connect_rooms（若 create_room 未直接连接）
↓
reveal_room
↓
inspect_room
↓
叙事
```

不要直接回复：

> 你发现一个秘密房间。

否则秘密房间不会真实存在。

### 11.2 NPC 避难

```text
inspect_npc(iven)
↓
inspect_room(plague_hospital)
↓
move_npc(dryRun=true)
↓
move_npc
↓
inspect_npc(iven)
```

### 11.3 触发重大世界事件

```text
inspect_world(visibility=gm)
↓
list_world_events
↓
trigger_world_event(dryRun=true)
↓
检查预览
↓
trigger_world_event
↓
inspect_world
```

### 11.4 生成遭遇

```text
inspect_room
↓
search_catalog(kind=enemy, query=亡灵)
↓
spawn_enemy(dryRun=true)
↓
spawn_enemy
↓
inspect_combat
```

---

## 12. Agent 不应该做的事情

禁止：

```text
修改 Go 源码
修改 React/JS 源码
直接编辑 data/runs/*.json
直接编辑完整 Run JSON
调用 /editor 任意覆写内容来解决普通剧情
自行生成骰子结果
自行声明 NPC 已移动
自行声明玩家获得物品
绕过 Game Runtime 修改 HP/Gold/Quest
```

如果当前 Tool 无法表达一个新机制：

```text
停止在 Runtime 层硬编
→ 记录“缺少底层能力”
→ 由开发者新增正式能力
```

---

## 13. 服务端部署建议

最少：

```text
Nginx / Caddy
      ↓ HTTPS
Go Game Server
      ├─ /
      ├─ /api/*
      ├─ /api/agent/*
      └─ /mcp
```

生产环境：

1. 必须设置 `ASHEN_AGENT_TOKEN` 或在反向代理实现 OAuth / Service Identity。
2. `/editor` 不要直接暴露公网，至少增加管理员认证。
3. `/mcp` 和 `/api/agent/*` 建议使用独立限流。
4. `data/` 必须持久化。
5. 备份 `data/runs`、`data/saves`、`data/editor`、`data/audit`。
6. 后续迁移 PostgreSQL 时保留 Gateway API，不需要改变 Agent Tool Contract。

---

## 14. 官方 MCP 规范说明

本实现针对 2026-07-28 MCP 现代生命周期：

- 不使用 `initialize` / `initialized`。
- 不使用 `Mcp-Session-Id`。
- 使用 `server/discover`。
- 每请求携带协议版本和客户端能力。
- Streamable HTTP 使用 `MCP-Protocol-Version`、`Mcp-Method`、`Mcp-Name`。
- `tools/list` 带 `ttlMs` / `cacheScope`。
- 每个现代结果带 `resultType`。

官方参考：

- https://modelcontextprotocol.io/
- https://blog.modelcontextprotocol.io/posts/2026-07-28/
- https://github.com/modelcontextprotocol/go-sdk

项目当前为了保持 V0.7.0 **离线即可编译**，MCP transport 使用 Go 标准库实现规范所需的 `server/discover`、`tools/list`、`tools/call` 子集，没有引入外部 SDK 依赖。

生产环境允许依赖下载时，可以把 `internal/httpapi/mcp.go` 替换成官方 Go SDK v1.7.0+ 的 Streamable HTTP Adapter；`internal/agentgateway` 与 `game.AgentToolDefinitions()` 无需改变。

---

## 15. Skills

项目附带三套 Skill：

```text
skills/ashen-crown-dm
skills/ashen-crown-world-director
skills/ashen-crown-encounter-referee
```

用途：

- `ashen-crown-dm`：总控 Dungeon Master，玩家自然语言 → inspect → Tool → 叙事。
- `ashen-crown-world-director`：NPC、区域、世界钟、Living World Event。
- `ashen-crown-encounter-referee`：检定、遭遇与战斗公平性。

Skill 的职责是告诉 Agent **什么时候用什么工具、以什么顺序用、什么不能做**。

Tool 才是真正改变世界的能力。

两者关系：

```text
Skill = SOP / 游戏导演规则
Tool  = 可执行能力
Runtime = 最终事实
```

# Ashen Crown V0.7.0 → mcp-server 工具注册指南

> 目标：使用 `lsl1016/mcp-server` 的“任意 HTTP JSON 接口注册为 MCP Tool”能力，把 Ashen Crown V0.7.0 AI-GM Bridge 中的 GM 能力注册到统一 MCP 网关。本文是对现有 `AI_GM_GUIDE.md` / `MCP_SERVER.md` 的增量说明，不要求把游戏规则搬到 MCP 层。

## 1. 推荐架构

```text
MCP Client / Agent Runtime
        │
        │ Authorization: Bearer <app_key>:<app_secret>
        ▼
lsl1016/mcp-server
POST /api/mcp
        │
        ├─ tools/list  ← mcp_tool + mcp_app_tool
        │
        └─ tools/call
             │
             │ HTTP JSON
             │ Authorization: Bearer <ASHEN_AGENT_TOKEN>
             ▼
Ashen Crown Game Server
POST /api/agent/tools/call/{name}
             │
             ▼
AgentGateway
  Schema Validation / Run Lock / dryRun / Audit
             │
             ▼
Game Runtime + Run State
```

客户端只持有 `mcp-server` 的应用凭证，不直接持有游戏 GM Token。Ashen Crown 的 `ASHEN_AGENT_TOKEN` 作为 mcp-server 到上游游戏服务的服务间凭证，存放在工具 `requestConfig.headers.Authorization` 中。

## 2. 为什么不能直接把现有 `/api/agent/tools/execute` 注册成 19 个独立 Tool

Ashen Crown 当前通用执行接口请求形态是：

```json
{
  "tool": "move_npc",
  "arguments": {
    "runId": "run_xxx",
    "npcId": "iven",
    "roomId": "room_25"
  }
}
```

而 `mcp-server` 当前 HTTP 通用分派规则是：

- GET / DELETE：MCP arguments 自动转 query string；
- POST / PUT / PATCH：MCP arguments **原样序列化成 JSON body**；
- `request_config` 只能配置 `method`、`timeout_ms`、静态 `headers`；
- 当前没有固定 body 字段注入 / body template / JSONPath 参数映射。

因此，如果把 `move_npc` 注册到 `/api/agent/tools/execute`，mcp-server 发送的是：

```json
{
  "runId": "run_xxx",
  "npcId": "iven",
  "roomId": "room_25"
}
```

而 Ashen Crown 需要外面再包一层 `tool + arguments`，二者不匹配。

### 2.1 零代码兼容方案

可以只注册 **一个** MCP Tool：`ashen_crown_execute`，让模型自己传：

```json
{
  "tool": "move_npc",
  "arguments": {
    "runId": "run_xxx",
    "npcId": "iven",
    "roomId": "room_25",
    "reason": "玩家成功说服伊文撤离"
  }
}
```

对应注册文件：

- `ashen-crown-single-tool.batch-create.example.json`

这个方案无需改 Ashen Crown，但不推荐正式使用：模型只看到一个“大总管”工具，失去了 MCP Tool 粒度、readOnly hint、独立 Schema 和按工具授权能力。

## 3. 推荐增量：增加 registration-friendly HTTP facade

在 Ashen Crown 只新增一个动态路由：

```text
POST /api/agent/tools/call/{name}
```

例如：

```text
POST /api/agent/tools/call/move_npc
```

请求 body **直接等于该 Tool 的 arguments**：

```json
{
  "runId": "run_xxx",
  "npcId": "iven",
  "roomId": "room_25",
  "reason": "玩家成功说服伊文撤离",
  "dryRun": true
}
```

内部只做：

```text
PathValue(name)
 + JSON body arguments
        ↓
AgentGateway.Execute(
  Tool=name,
  Arguments=body
)
```

不要复制 Game Runtime 规则，也不要绕过 AgentGateway。

### 3.1 推荐返回协议

为了适配 `mcp-server` 的上游成功/失败归一化，建议该 facade 对**业务错误**返回 HTTP 200，并用 `success/code/message` 表示执行结果；网络错误、服务不可用等真正的 transport error 再使用 4xx/5xx。

成功：

```json
{
  "success": true,
  "code": 0,
  "message": "ok",
  "tool": "move_npc",
  "runId": "run_xxx",
  "changed": true,
  "dryRun": false,
  "risk": "medium",
  "auditId": "audit_xxx",
  "summary": "NPC iven 已移动到灰疫医馆。",
  "data": {
    "npcId": "iven",
    "roomId": "room_25"
  },
  "warnings": []
}
```

业务失败：

```json
{
  "success": false,
  "code": 4004,
  "message": "房间不存在",
  "tool": "move_npc",
  "runId": "run_xxx",
  "changed": false
}
```

原因：`mcp-server` 会从 `success`、`errNo/errno/code` 和 `message/msg/errMsg` 等常见字段判断上游业务是否成功。当前它对非 2xx HTTP 响应只保留 HTTP 状态码，因此把可恢复的游戏规则错误建模成业务错误，模型能得到更有用的错误文案。

## 4. mcp-server Tool 注册字段映射

`/api/manage/tools/batchCreate` 的单工具字段与 Ashen Crown 的映射如下：

| mcp-server 字段 | Ashen Crown 来源/建议 |
|---|---|
| `bizTag` | 固定 `ashen-crown-gm` |
| `url` | `http://ashen-crown:8080/api/agent/tools/call/<toolName>` |
| `requestConfig` | POST + 10s timeout + 上游 Agent Token |
| `name` | Agent Tool `name` |
| `title` | Agent Tool `title` |
| `description` | Agent Tool `description` |
| `inputSchema` | Agent Tool `inputSchema`，建议去掉直连 MCP 专用的 `x-mcp-header` |
| `outputSchema` | 注册网关返回 Schema；建议开启 `x-output-projection` |
| `readOnly` | `readOnlyHint=true → 2`；否则 `1` |
| `isInternal` | `2`，Ashen Crown 是内部上游服务 |
| `status` | `2`，上线 |
| `owner` | 建议 `ashen-crown` |

`readOnly` / `status` 的数值不是布尔值：

```text
readOnly: 1 = 写操作
readOnly: 2 = 只读

status:   1 = 下线
status:   2 = 启用
```

## 5. requestConfig

推荐每个工具统一使用：

```json
{
  "method": "POST",
  "timeout_ms": 10000,
  "headers": {
    "Authorization": "Bearer <ASHEN_AGENT_TOKEN>",
    "X-Agent-Name": "mcp-server"
  }
}
```

mcp-server 会自动补齐 JSON 请求所需的 `Content-Type` / `Accept`。

生产环境建议：

- mcp-server 与 Ashen Crown 走内网地址；
- `ASHEN_AGENT_TOKEN` 只作为服务间凭证；
- MCP Client 只获得 mcp-server 的 `app_key:app_secret`；
- 不要把 Ashen Crown GM Token 下发到 Claude/Codex/浏览器客户端。

## 6. Output Schema 与结果整形

`mcp-server` 支持 `output_schema`，并可在 Schema 顶层增加：

```json
{
  "x-output-projection": true
}
```

开启后，服务端会：

1. 按 Schema 对上游整个 JSON 对象做投影；
2. 删除未声明的字段；
3. 校验投影后的 structured output；
4. 把字段 `description` 汇总成模型可读的字段说明；
5. 同时返回 `StructuredContent`。

这正好可以补足 Ashen Crown 当前 `data:any` 输出较宽的问题，而且不用让模型自己从巨大 Run JSON 中找关键字段。

推荐公共投影至少保留：

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "x-output-projection": true,
  "properties": {
    "success": {"type":"boolean"},
    "tool": {"type":"string"},
    "runId": {"type":"string"},
    "changed": {"type":"boolean"},
    "dryRun": {"type":"boolean"},
    "summary": {"type":"string"},
    "data": {"type":"object", "additionalProperties":true},
    "warnings": {"type":"array", "items":{"type":"string"}},
    "auditId": {"type":"string"}
  },
  "required": ["success", "tool", "runId", "changed", "summary"],
  "additionalProperties": false
}
```

后续建议再把每个 Tool 的 `data` 收紧，例如 `roll_check.data` 明确为：

```json
{
  "type": "object",
  "properties": {
    "kind": {"type":"string"},
    "label": {"type":"string"},
    "roll": {"type":"integer"},
    "bonus": {"type":"integer"},
    "total": {"type":"integer"},
    "dc": {"type":"integer"},
    "success": {"type":"boolean"},
    "critical": {"type":"boolean"}
  },
  "required": ["roll","bonus","total","dc","success"]
}
```

这部分可以逐 Tool 增量收紧，不影响注册机制。

## 7. 19 个工具注册矩阵

| Tool | Category | Risk | 类型 | 上游 URL |
|---|---|---|---|---|
| `advance_bell` | world_mutation | high | 写 | `http://ashen-crown:8080/api/agent/tools/call/advance_bell` |
| `connect_rooms` | map_mutation | medium | 写 | `http://ashen-crown:8080/api/agent/tools/call/connect_rooms` |
| `create_room` | map_mutation | high | 写 | `http://ashen-crown:8080/api/agent/tools/call/create_room` |
| `grant_item` | reward | high | 写 | `http://ashen-crown:8080/api/agent/tools/call/grant_item` |
| `inspect_combat` | inspect | low | 只读 | `http://ashen-crown:8080/api/agent/tools/call/inspect_combat` |
| `inspect_npc` | inspect | low | 只读 | `http://ashen-crown:8080/api/agent/tools/call/inspect_npc` |
| `inspect_quest` | inspect | low | 只读 | `http://ashen-crown:8080/api/agent/tools/call/inspect_quest` |
| `inspect_room` | inspect | low | 只读 | `http://ashen-crown:8080/api/agent/tools/call/inspect_room` |
| `inspect_world` | inspect | low | 只读 | `http://ashen-crown:8080/api/agent/tools/call/inspect_world` |
| `list_world_events` | inspect | low | 只读 | `http://ashen-crown:8080/api/agent/tools/call/list_world_events` |
| `move_npc` | world_mutation | medium | 写 | `http://ashen-crown:8080/api/agent/tools/call/move_npc` |
| `recent_log` | inspect | low | 只读 | `http://ashen-crown:8080/api/agent/tools/call/recent_log` |
| `reveal_room` | map_mutation | medium | 写 | `http://ashen-crown:8080/api/agent/tools/call/reveal_room` |
| `roll_check` | rule | medium | 写 | `http://ashen-crown:8080/api/agent/tools/call/roll_check` |
| `search_catalog` | inspect | low | 只读 | `http://ashen-crown:8080/api/agent/tools/call/search_catalog` |
| `set_flag` | world_mutation | medium | 写 | `http://ashen-crown:8080/api/agent/tools/call/set_flag` |
| `set_region_state` | world_mutation | high | 写 | `http://ashen-crown:8080/api/agent/tools/call/set_region_state` |
| `spawn_enemy` | encounter | high | 写 | `http://ashen-crown:8080/api/agent/tools/call/spawn_enemy` |
| `trigger_world_event` | world_mutation | high | 写 | `http://ashen-crown:8080/api/agent/tools/call/trigger_world_event` |

完整可导入的注册请求体见：

```text
ashen-crown-tools.batch-create.example.json
```

注意把：

```text
http://ashen-crown:8080
<ASHEN_AGENT_TOKEN>
```

替换成实际部署值。

## 8. 批量注册

假设：

```bash
MCP_SERVER=http://127.0.0.1:18080
ADMIN_TOKEN=demo-admin-token
```

执行：

```bash
curl -s \
  -H "X-Admin-Token: $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  --data-binary @ashen-crown-tools.batch-create.example.json \
  "$MCP_SERVER/api/manage/tools/batchCreate"
```

返回类似：

```json
{
  "data": {
    "list": [
      {"id": 21, "name": "inspect_world"},
      {"id": 22, "name": "inspect_room"}
    ]
  }
}
```

记录这些 `toolIds`，后续授权给 MCP App。

## 9. 创建 MCP App 并授权

创建应用：

```bash
curl -s \
  -H "X-Admin-Token: $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"appName":"ashen-crown-dm","owner":"ashen-crown"}' \
  "$MCP_SERVER/api/manage/app/create"
```

得到：

```text
appKey
appSecret
appId
```

给应用绑定刚注册的 19 个 Tool：

```bash
curl -s \
  -H "X-Admin-Token: $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"appId":1,"toolIds":[21,22,23]}' \
  "$MCP_SERVER/api/manage/app/grantTools"
```

正式使用时把全部 19 个真实 Tool ID 填入。

## 10. MCP 客户端最终配置

客户端不再直连 Ashen Crown `/mcp`，而是连接统一网关：

```json
{
  "mcpServers": {
    "ashen-crown": {
      "type": "http",
      "url": "https://mcp.example.com/api/mcp",
      "headers": {
        "Authorization": "Bearer <app_key>:<app_secret>"
      }
    }
  }
}
```

此时链路为：

```text
Agent
  ↓ tools/list
mcp-server
  ↓ 仅返回 app 已授权的 Ashen Crown Tools
Agent
  ↓ tools/call(move_npc)
mcp-server
  ↓ POST /api/agent/tools/call/move_npc
Ashen Crown AgentGateway
  ↓
Game Runtime
```

## 11. runId 怎么传

通过 mcp-server 注册后，`runId` 就是一个普通的 MCP Tool input 字段：

```json
{
  "runId": "run_01JXYZ",
  "npcId": "iven",
  "roomId": "room_25",
  "reason": "玩家成功说服伊文撤离",
  "dryRun": true
}
```

不需要客户端额外配置 `Mcp-Param-Run-Id`。注册 JSON 中已经移除了 Ashen Crown 直连 MCP Server 使用的 `x-mcp-header` 扩展；mcp-server 会把整个 arguments JSON 原样转给上游。

## 12. 推荐 Agent 调用规则

### 只读链路

```text
inspect_world
→ inspect_room / inspect_npc
→ 决策
```

### 改世界链路

```text
inspect
→ search_catalog（ID 不确定时）
→ mutation dryRun=true
→ 检查 summary/data
→ mutation dryRun=false
→ re-inspect
→ narrate
```

### 检定

```text
inspect
→ roll_check
→ 只接受 Runtime 的真实 roll/success
→ 根据真实结果继续调用 mutation 或叙事
```

## 13. 安全边界

不要注册以下人类 Admin/GM 能力给普通 DM Agent：

- 修改 Tool / Enemy / Item / Skill 模板；
- 删除内容；
- 修改任意存档 JSON；
- mcp-server 管理端 `/api/manage/**`；
- Ashen Crown `/editor` 管理权限。

Agent 应只获得本文件矩阵里的受控 Runtime Tool。

对于 `high` 风险工具：

- 保留 Ashen Crown `reason` 校验；
- Prompt / Skill 中要求先 `dryRun=true`；
- mcp-server `readOnly=1`，使 MCP Tool annotation 表达为写操作；
- 使用 `mcp_app_tool` 做按应用最小授权。

## 14. 注册自动同步建议

当前 Tool 真相源仍然是 Ashen Crown：

```text
GET /api/agent/tools
```

长期建议实现一个 `sync-ashen-tools` 小工具：

```text
GET Ashen Crown /api/agent/tools
        ↓
移除 x-mcp-header
        ↓
生成 mcp-server CreateToolInput
        ↓
按 name 查询已有 mcp_tool
        ↓
不存在 → batchCreate
存在且 Schema/描述变化 → batchUpdate
        ↓
按目标 App grantTools
```

这样以后 Ashen Crown 新增 Tool，只维护 `AgentToolDefinitions()`，MCP 注册信息自动同步，不需要手工维护 19 份配置。

## 15. 验收清单

注册完成后至少验证：

1. `tools/list` 只返回目标 App 已授权的工具；
2. `inspect_world` 能读取指定 `runId`；
3. `move_npc(dryRun=true)` 不修改存档；
4. `move_npc(dryRun=false)` 修改存档且产生 `auditId`；
5. 错误 roomId 能把“房间不存在”传回模型，而不是只看到 HTTP 400；
6. `roll_check` 的 `StructuredContent` 包含 roll/bonus/total/dc/success；
7. `x-output-projection` 能裁剪不需要的字段；
8. 未授权的 Tool 不出现在该 App 的 `tools/list`；
9. Ashen Crown Token 不暴露给 MCP Client；
10. Agent Audit 与 mcp-server `mcp_call_log` 能串起一次完整调用。

## 16. 与原 V0.7 AI-GM Bridge 的关系

建议最终只保留一份游戏规则实现：

```text
AgentToolDefinitions()      ← Tool 语义/Schema 真相源
AgentGateway               ← 鉴权后的规则入口
Game Runtime               ← 游戏事实与规则
```

MCP 有两种部署选择：

### 方案 A：Ashen Crown 内置 MCP

适合单服务、直接连接。

### 方案 B：统一 mcp-server 注册网关（推荐企业/多服务）

适合：

- 把多个业务 HTTP API 都注册成 MCP Tool；
- 统一应用授权；
- 统一限流；
- 统一调用审计；
- 统一结果投影；
- 多服务共用一个 MCP Endpoint。

使用方案 B 后，Ashen Crown 自带 `/mcp` 可以保留作本地开发/兼容入口，但生产 Agent 优先连接统一 `mcp-server /api/mcp`。

## 17. 本文基于的 mcp-server 实现点

检查的仓库：`lsl1016/mcp-server`，默认分支 `main`。

关键代码位置：

- `README.md`：总体架构、管理 API、注册示例、客户端配置；
- `service/manage/tools/tool.go`：`CreateToolInput`、批量新增/更新字段与 Schema 校验；
- `mcp/registry.go`：数据库 Tool → MCP SDK Tool 的动态转换，按 App 过滤；
- `mcp/dispatch.go`：HTTP 参数转发、请求方式与上游响应归一化；
- `mcp/toolconfig/config.go`：`request_config` 支持 method / timeout / headers；
- `mcp/output_projection.go`：`x-output-projection` 投影；
- `mcp/result.go`：StructuredContent 和 outputSchema 校验。

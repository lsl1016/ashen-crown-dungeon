# Ashen Crown 通用 MCP 网关 HTTP Facade

## 目标

让 `lsl1016/mcp-server` 把每个 Ashen Crown GM Tool 独立注册成 MCP Tool，而不是只注册一个总入口 `execute`。

新增接口：

```text
POST /api/agent/tools/call/{name}
```

请求 Body 直接等于该 Tool 的 arguments：

```json
{
  "runId": "run_xxx",
  "npcId": "iven",
  "roomId": "room_25",
  "reason": "玩家说服伊文撤离",
  "dryRun": true
}
```

服务端内部仍复用：

```text
AgentGateway.Execute
  -> ValidateAgentArguments
  -> Run Lock
  -> dryRun clone
  -> Game Engine ExecuteAgentTool
  -> SaveRun
  -> Audit
```

不复制任何游戏规则。

## 为什么不能只用 /api/agent/tools/execute

原接口要求：

```json
{
  "tool": "move_npc",
  "arguments": { ... }
}
```

而通用 `mcp-server` 对 POST 工具会把 MCP arguments 原样作为上游 JSON Body，因此一个独立的 `move_npc` 工具无法自动再包一层 `tool + arguments`。

Facade 把 Tool 名放到 URL：

```text
/api/agent/tools/call/move_npc
```

于是 MCP arguments 可以直接透传。

## 成功返回

```json
{
  "success": true,
  "code": 0,
  "message": "伊文已移动到灰疫医馆。",
  "tool": "move_npc",
  "runId": "run_xxx",
  "changed": true,
  "dryRun": false,
  "summary": "伊文已移动到灰疫医馆。",
  "data": {},
  "warnings": [],
  "auditId": "audit_xxx"
}
```

## 业务失败返回

业务失败使用 HTTP 200：

```json
{
  "success": false,
  "code": 4000,
  "message": "未知房间: room_999",
  "tool": "move_npc",
  "runId": "run_xxx",
  "changed": false,
  "summary": "工具执行失败"
}
```

这是为了兼容通用 `mcp-server` 的上游响应归一化逻辑：它能读取 `success=false / code / message` 并把真实业务错误返回给模型。

以下仍使用标准 HTTP 错误码：

- 401：Agent Token 不正确
- 400：HTTP Body 不是合法 JSON
- 413：Body 超过限制

## mcp-server 注册示例

```json
{
  "url": "http://ashen-crown:8080/api/agent/tools/call/move_npc",
  "requestConfig": "{\"method\":\"POST\",\"timeout_ms\":10000,\"headers\":{\"Authorization\":\"Bearer <ASHEN_AGENT_TOKEN>\",\"X-Agent-Name\":\"mcp-server\"}}",
  "name": "move_npc",
  "title": "Move NPC",
  "description": "移动一个持久 NPC 到指定地点。",
  "inputSchema": "{...}",
  "outputSchema": "{...}",
  "readOnly": 1,
  "isInternal": 2,
  "status": 2,
  "owner": "ashen-crown"
}
```

## 验证

```bash
curl -X POST http://127.0.0.1:8080/api/agent/tools/call/inspect_world \
  -H 'Authorization: Bearer ashen-dev-token' \
  -H 'Content-Type: application/json' \
  -H 'X-Agent-Name: mcp-server' \
  -d '{"runId":"run_xxx","visibility":"gm"}'
```

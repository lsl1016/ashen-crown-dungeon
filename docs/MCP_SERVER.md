# MCP Server 技术说明

## 协议

- MCP revision: `2026-07-28`
- HTTP endpoint: `POST /mcp`
- Transport model: stateless request/response
- Methods: `server/discover`, `tools/list`, `tools/call`
- Explicit game state handle: `runId`

## Header Contract

所有现代 MCP 请求：

```text
MCP-Protocol-Version: 2026-07-28
Mcp-Method: <json-rpc method>
```

`tools/call` 额外：

```text
Mcp-Name: <tool name>
Mcp-Param-Run-Id: <runId>
```

Header 与 JSON Body 不一致会返回 JSON-RPC `-32020 HeaderMismatch`。

## Authentication

如果设置：

```text
ASHEN_AGENT_TOKEN
```

则请求必须包含：

```text
Authorization: Bearer <token>
```

当前是轻量 Service Token，生产环境可以由反向代理替换成 OAuth / workload identity。

## Implementation Boundary

MCP Transport 不直接操作 `Run`。

```text
MCP
 ↓
AgentGateway
 ↓
JSON Schema Validation
 ↓
Run Lock
 ↓
Game Engine ExecuteAgentTool
 ↓
Store.SaveRun
 ↓
Audit
```

因此以后即使把 MCP 标准库实现替换成官方 Go SDK，游戏规则不会移动到 MCP 层。

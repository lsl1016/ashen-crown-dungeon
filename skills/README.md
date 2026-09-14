# Ashen Crown Skills

本目录包含三个面向 Agent 的 Skill 源码：

- `ashen-crown-dm`：总控 DM 工作流。
- `ashen-crown-world-director`：世界事件、NPC、区域、地图演化。
- `ashen-crown-encounter-referee`：检定、遭遇、奖励公平性。

已经验证并打包的安装包位于：

```text
skills-dist/ashen-crown-dm/skill.zip
skills-dist/ashen-crown-world-director/skill.zip
skills-dist/ashen-crown-encounter-referee/skill.zip
```

Skill 不直接修改游戏，它们定义 Agent 的 SOP。真实能力来自 `/api/agent/*` 或 `/mcp` 暴露的 Tool。

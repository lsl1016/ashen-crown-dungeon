# V0.5 RPG 系统说明

V0.5 不修改双幕 44 地点的核心地图拓扑，重点是在已有 Game Runtime 上增加可持续扩展的 RPG 交互层。

## NPC 与对话

NPC 使用 `NPCDef` 描述身份、阵营、立绘、对话树和可选商店。`DialogueDef` 是结构化节点图，每个 `DialogueChoiceDef` 可以：

- 跳转到下一个节点；
- 结束对话；
- 打开商店；
- 要求 Flag / Item；
- 应用 Quest / Flag / Lore / Item / Relation 等 Effect。

运行时只保存 `ActiveDialogue` 与 `NPCRelations`，世界定义仍保存在静态内容数据中。

## 商店

`ShopDef` 定义售卖清单、价格和回收比例；`Run.ShopStock` 保存本轮世界中的剩余库存。

买入流程：校验商店 → 校验库存 → 校验金币 → 扣金币 → 发物品 → 扣库存 → 写日志。

卖出流程：校验背包 → 禁止直接出售已装备物品 → 禁止出售剧情关键物品 → 按 Buyback 结算 → 回流库存 → 写日志。

## 角色成长

角色现在有三种成长资源：

1. `AttributePoints`：力量、敏捷、感知、意志。
2. `MasteryPoints`：三条通用精通路线。
3. `TalentPoints`：职业专属天赋。

精通不是展示字段，直接进入 Go 战斗结算。

## 表现层

- `web/assets/portraits/`：NPC 与敌人 / Boss 立绘。
- `web/assets/actors/`：三个职业三种动作、每种四帧。
- 玩家动作由后端行动结果触发，前端只播放表现，不自行计算伤害。
- 场景环境拆成远 / 中 / 前三个图层，并按照墓城 / 荒原主题呈现不同粒子。

## 后续 AI 接入

未来 DM Agent 可以使用 NPC 关系、任务、商店和成长状态作为上下文，但不能直接覆盖 Run JSON。推荐继续通过受控 Tool 修改：

- NPC / Quest / Relation
- Shop / Item
- World Flag
- Room / Scene Element

这样 AI 创造的剧情仍然受现有 Game Runtime 约束。

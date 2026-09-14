# V0.6 持续世界与战术系统

V0.6 的目标不是继续增加房间数量，而是让 V0.5 已经存在的 NPC、任务、经济、装备和战斗开始相互影响。

## 1. 持久世界状态

Run 新增四组核心字段：

```go
NPCLocations map[string]string
RegionStates map[string]string
ShopRefresh  map[string]int
ItemAffixes  map[string][]AffixState
```

它们与已有 Flags、Quests、NPCRelations、ShopStock 一起组成 V0.6 的世界事实。

### 设计原则

- NPC 当前在哪里，由 Run 决定，不由前端静态元素决定。
- 区域当前是否危险，由剧情事实和钟声决定。
- 商店补货发生过没有，要进入存档。
- 掉落词条在获得时确定，此后读档不能重新掷词条。

## 2. NPC 日程

`syncNPCLocations()` 根据 Run 事实计算 NPC 的期望位置。

例如伊文：

```text
默认：room_13 断桥营火
Bell >= 7：room_25 灰疫医馆
lost_patrol_reported：room_01 墓城入口
```

其中任务结局优先级高于纯时间条件。

前端不再假定房间 JSON 中有一个 NPC SceneElement 就代表 NPC 永远存在。它会读取：

```json
"npcLocations": {
  "iven": "room_25"
}
```

并动态生成场景人物与对话按钮。

## 3. 区域状态

区域状态用于表达长期后果，但不承担全部规则逻辑。

```text
Flags / Quest Outcome / Clock
          ↓
 syncRegionStates
          ↓
 RegionStates
          ↓
 UI + 后续规则读取
```

当前区域：

- 外墓区
- 旧城区
- 王城中环
- 内廷
- 烬冠核心
- 雾外荒原
- 荒原东界

## 4. 任务阶段

QuestState 新增：

```go
Stage   string
Outcome string
```

推荐生命周期：

```text
active objective
→ completed objective
→ return
→ resolved
```

`completed` 可以表示“目标已经做完但还没有交付”，`resolved + outcome` 表示这条任务已经产生世界后果。

这样 AI 以后询问任务状态时，不会把“杀掉目标”和“支线已经真正结束”混为一谈。

## 5. 商店刷新

每个商店保存：

```text
ShopStock
ShopRefresh[shopID]
```

世界钟每推进后执行刷新检查：

```text
bell - lastRefresh >= 3
```

消耗品每次最多补 2，其他商品最多补 1，且不超过内容定义里的库存上限。

## 6. 装备词条

AffixDef 是内容模板，AffixState 是写入 Run 的结果。

生成输入：

```text
WorldSeed
+ ItemID
+ Drop Source / Turn
```

因此具备：

- Seed 可复现
- Debug 可复现
- 存档可复现
- AI 无法通过重新描述把装备词条改掉

目前原型用 `ItemAffixes[itemID]` 保存，因此同一 ItemID 副本共享词条。后续正式刷装时升级到 `ItemInstance`：

```text
inventory[] -> instanceId
ItemInstance {
  instanceId
  itemId
  affixes[]
  durability
  bound
}
```

## 7. 战斗距离

CombatState：

```go
Distance int // 1 close, 2 middle, 3 far
```

距离是实际规则变量：

- 玩家 `advance`：Distance - 1
- 玩家 `retreat`：Distance + 1
- 近战敌人可以 Intent=`advance`
- Boss charge 会强制 Distance=1
- 远距会降低部分近战敌人的命中和伤害
- 部分职业天赋读取 Distance

距离操作本身消耗回合，因此不是免费位移。

## 8. 三阶职业树

TalentDef 新增：

```go
Tier          int
RequiredLevel int
Requires      []string
```

Engine 在学习时同时校验：

1. 职业匹配
2. 天赋点
3. 等级
4. 前置天赋
5. 最大 Rank

未来可以继续在这个结构上实现互斥分支：

```go
ExclusiveGroup string
```

而无需改变基本模型。

## 9. V0.5 旧存档

`Engine.Prepare(run)` 会补齐 V0.6 缺失字段，并由 HTTP 层在修改前后调用。

因此旧存档不会因为 nil map 或 Combat.Distance=0 直接崩溃。

## 10. AI 接入意义

V0.6 之后，未来 DM Agent 已经可以查询：

```text
这个 NPC 现在在哪？
这个区域现在安全吗？
这条任务只是目标完成了，还是已经交付？
这个商人什么时候补过货？
这把装备有什么固定词条？
玩家和敌人现在是什么距离？
```

这些答案来自 Run，而不是 LLM 记忆。

这正是“世界不是由模型临时描述，而是由 Agent 驱动并持续存在”的基础。

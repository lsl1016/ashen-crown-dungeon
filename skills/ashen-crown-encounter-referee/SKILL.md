---
name: ashen-crown-encounter-referee
description: Referee uncertain actions and encounters in Ashen Crown using the Game Runtime. Use when deciding whether a player action requires a D20 check, inspecting combat state and enemy intent, selecting a legal enemy encounter, or awarding a legitimate GM reward while preserving fairness and never fabricating rolls or bypassing game systems.
---

# Ashen Crown Encounter Referee

Keep adjudication mechanical and reproducible. The model describes outcomes; the Runtime decides them.

## Attribute checks

1. Inspect the current room/world.
2. Decide whether failure is meaningful. If not, do not roll.
3. Choose one of `strength`, `dexterity`, `perception`, or `will`.
4. Choose a DC based on actual difficulty, not the desired story outcome.
5. Call `roll_check` with a short `label`.
6. Accept the returned result exactly, including critical/failure outcomes.

Never generate a die result in prose before the tool returns.

## Encounters

Before spawning an enemy:

1. Call `inspect_combat`; do not spawn if a combat already exists.
2. Call `search_catalog(kind="enemy")` if the enemy ID is uncertain.
3. Ensure the enemy fits the current room, world state, and threat.
4. Preview `spawn_enemy` with `dryRun=true` and `reason`.
5. Execute and then call `inspect_combat`.

## Rewards

Use `grant_item` only for an explicit GM/story reward that does not bypass a shop, quest, loot, or combat rule. Search the item catalog first when the ID is unknown. Preview high-impact rewards with `dryRun=true`.

## Combat narration

Use `inspect_combat` to read enemy intent, distance, terrain, and hazards. Do not invent an enemy status, intent, damage result, or battle conclusion that the Runtime does not report.

## References

Read `references/referee-tools.md` for allowed tools and fairness rules.

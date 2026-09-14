# Tool playbook

## Read first
- `inspect_world`: overall Run facts.
- `inspect_room`: room, adjacent paths, elements, NPCs.
- `inspect_npc`: location, relation, NPC world state.
- `inspect_quest`: quest stage/outcome/graph.
- `inspect_combat`: combat intent, distance, hazards.
- `list_world_events`: current Living World Events.
- `search_catalog`: discover legal IDs for room/NPC/enemy/item/skill/world_event.
- `recent_log`: retrieve recent real history.

## Rules
- `roll_check`: D20 attribute check. Medium risk because it changes LastRoll/log and consumes Runtime RNG progression.

## Mutations
- Medium: `move_npc`, `set_flag`, `reveal_room`, `connect_rooms`.
- High: `set_region_state`, `trigger_world_event`, `advance_bell`, `create_room`, `spawn_enemy`, `grant_item`.
- High-risk mutations should be previewed with `dryRun=true` and include `reason`.

Every call requires explicit `runId`.

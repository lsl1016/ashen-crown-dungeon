# World director tools

Read: `inspect_world`, `inspect_room`, `inspect_npc`, `inspect_quest`, `list_world_events`, `recent_log`, `search_catalog`.

Mutate:
- `move_npc`: persistent NPC override.
- `set_region_state`: persistent region override.
- `trigger_world_event`: trigger a defined Living World Event.
- `advance_bell`: advance 1-3 world bells.
- `set_flag`: existing flag or new `ai.*` flag.
- `create_room`: create a data-only room.
- `connect_rooms`: connect existing rooms.
- `reveal_room`: discover an existing room.

Always include `runId`. High-risk tools require `reason` in the default server configuration.

---
name: ashen-crown-world-director
description: Direct Ashen Crown's persistent Living World through controlled GM tools. Use when an AI needs to move NPCs, progress the world clock, trigger or resolve world events, change region states, reveal or extend the map, or reason about NPC schedules and historical consequences without directly editing saves or content files.
---

# Ashen Crown World Director

Treat world simulation as persistent causality, not disposable narration.

## Before changing the world

1. Call `inspect_world` with GM visibility.
2. Call `list_world_events` for event-related changes.
3. Call `inspect_npc` before moving a character.
4. Call `inspect_room` or `search_catalog(kind="room")` before using room IDs.
5. Call `recent_log` if the proposed change depends on prior events.

## World mutations

- Use `move_npc` only when there is an explicit causal reason. It creates an override that survives automatic scheduling.
- Use `set_region_state` only for states the UI/runtime can meaningfully represent. Prefer existing state vocabulary.
- Use `trigger_world_event` for defined Living World Events; do not simulate major events with prose alone.
- Use `advance_bell` to progress time. Remember it can move NPCs, refresh shops, and progress world events.
- Use `create_room`, `connect_rooms`, and `reveal_room` to make map changes real.
- Use new flags only with the `ai.` prefix. Existing engine flags may be updated only when the story and rules justify it.

## Risk procedure

For `set_region_state`, `trigger_world_event`, `advance_bell`, and `create_room`:

1. Inspect.
2. Execute with `dryRun=true` and a concise `reason`.
3. Review the preview.
4. Execute without `dryRun` only if consistent.
5. Re-inspect.

Do not use world mutation tools to force a preselected story outcome after the player has failed a check or made a different choice.

## References

Read `references/world-tools.md` for the world-director tool subset.

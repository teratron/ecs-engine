# Example: ECS Core Patterns

This category contains examples validating the foundational Entity-Component-System (ECS) architecture. These examples serve as the primary verification for Phase 1 of the engine development.

## Planned Examples

- **`poc`**: A proof-of-concept example demonstrating basic entity spawning, component attachment, and simple system iteration.
- **`queries`**: Advanced query filtering (`With`, `Without`, `Added`, `Changed`) and tuple-based data access.
- **`system_chaining`**: Orchestrating system execution order using explicit dependencies and system sets.
- **`component_storage`**: Demonstrating the difference between `Table` (dense) and `SparseSet` storage strategies.
- **`lifecycle_hooks`**: Using `OnAdd`, `OnInsert`, and `OnRemove` hooks to respond to entity and component changes.

## Running the Example

```bash
go run ./examples/ecs
```

## Testing

```bash
go test ./examples/ecs/...
```

## Status

**Placeholder.** Implementation begins during Phase 1.

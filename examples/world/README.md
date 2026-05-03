# Example: World Systems

Examples in this category focus on high-level world management, including global resources, event bus communication, and entity hierarchies.

## Planned Examples

- **`resources`**: Managing global singletons and accessing them from systems.
- **`events`**: Using the double-buffered event bus to send and receive messages between systems.
- **`hierarchy`**: Creating parent-child relationships and propagating transforms or visibility.
- **`change_detection`**: Reacting to component mutations using the tick-based detection system.
- **`observers`**: Using reactive triggers that fire immediately when components are added or removed.

## Running the Example

```bash
go run ./examples/world
```

## Testing

```bash
go test ./examples/world/...
```

## Status

**Placeholder.** Implementation begins during Phase 1/2.

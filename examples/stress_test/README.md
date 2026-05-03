# Example: Stress Tests & Benchmarks

Examples in this category focus on pushing the engine to its limits to validate performance and scalability.

## Planned Examples

- **`massive_spawning`**: Spawning and despawning tens of thousands of entities per second.
- **`heavy_iteration`**: Iterating over millions of entities with complex component combinations.
- **`parallel_systems`**: Validating the efficiency of the parallel scheduler with high system counts.
- **`memory_usage`**: Tracking allocation patterns and heap pressure under heavy load.
- **`collision_stress`**: Simulating thousands of colliding bodies to test physics performance.

## Running the Example

```bash
go run ./examples/stress_test
```

## Testing

```bash
go test ./examples/stress_test/...
```

## Status

**Placeholder.** Implementation begins during Phase 2/3.

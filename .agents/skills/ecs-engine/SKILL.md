```markdown
# ecs-engine Development Patterns

> Auto-generated skill from repository analysis

## Overview

This skill teaches the core development patterns, coding conventions, and collaborative workflows used in the `ecs-engine` TypeScript codebase. The repository focuses on building an Entity-Component-System (ECS) engine, with a strong emphasis on design documentation, specification-driven development, and agent-assisted automation. No framework is enforced, and the project values clear, maintainable code and well-documented design processes.

---

## Coding Conventions

**File Naming**

- Use `kebab-case` for all file names.
  - Example: `entity-manager.ts`, `component-registry.test.ts`

**Import Style**

- Use **relative imports** for all modules.
  - Example:
    ```typescript
    import { Entity } from './entity';
    import { Component } from '../components/component';
    ```

**Export Style**

- Use **named exports** (not default).
  - Example:
    ```typescript
    // entity.ts
    export interface Entity { /* ... */ }
    export function createEntity() { /* ... */ }
    ```

**Commit Messages**

- Follow the **Conventional Commits** standard.
- Prefixes: `feat`, `docs`, `chore`
- Example:
  ```
  feat: add system scheduler for parallel execution
  docs: update physics system specification
  chore: refactor component registration logic
  ```

---

## Workflows

### Add or Update Design Specification
**Trigger:** When introducing or updating a system, feature, or architectural pattern specification.
**Command:** `/add-spec`

1. Create or update one or more files in `.design/main/specifications/`.
2. Update `.design/main/INDEX.md` to register the new or changed specification(s).
3. Optionally, update related cross-references in other spec files.

**Example:**
```bash
# Add a new rendering system spec
touch .design/main/specifications/rendering-system.md
# Edit INDEX.md to add a link to the new spec
nano .design/main/INDEX.md
```

---

### Add Draft and Design Specs in Parallel
**Trigger:** When documenting both early-stage drafts and formal design specs for new features or systems.
**Command:** `/add-draft-spec`

1. Create or update files in `.design/main/specifications/` for the formal spec.
2. Create or update corresponding draft documents in `.draft/` (e.g., `.draft/physics/`).
3. Update `.design/main/INDEX.md` and/or `.design/main/PLAN.md` to reference the new work.

**Example:**
```bash
# Add draft and spec for new physics system
touch .draft/physics/physics-research.md
touch .design/main/specifications/physics-system.md
nano .design/main/INDEX.md
nano .design/main/PLAN.md
```

---

### Enrich Multiple Specs with Patterns or Cross-References
**Trigger:** When propagating new architectural patterns or cross-references across multiple specs.
**Command:** `/enrich-specs`

1. Update several files in `.design/main/specifications/` to add new sections or patterns.
2. Update `.design/main/INDEX.md` to reflect the new version or changes.

**Example:**
```bash
# Add "Event Sourcing" pattern to multiple specs
nano .design/main/specifications/entity-system.md
nano .design/main/specifications/component-system.md
nano .design/main/INDEX.md
```

---

### Add New Go Implementation Specs
**Trigger:** When defining Go implementation details for ECS systems.
**Command:** `/add-go-specs`

1. Create or update files in `.design/main/specifications/*-go.md`.
2. Update `.design/main/INDEX.md` to register the new Go specs.

**Example:**
```bash
# Add Go implementation spec for the scheduler
touch .design/main/specifications/scheduler-go.md
nano .design/main/INDEX.md
```

---

### Add or Update Agent Skills or Rules
**Trigger:** When expanding or updating agent automation capabilities.
**Command:** `/add-agent-skill`

1. Create or update files in `.agents/skills/` or `.agents/rules/`.
2. Optionally, add or update workflow files in `.agents/workflows/`.
3. Optionally, update `AGENTS.md`.

**Example:**
```bash
# Add a new agent skill for code formatting
touch .agents/skills/formatting.md
nano AGENTS.md
```

---

## Testing Patterns

- **Test files** follow the pattern `*.test.*` (e.g., `entity-manager.test.ts`).
- **Testing framework** is not explicitly specified; check test files for setup.
- Place tests alongside or near the code they test.
- Example test file:
  ```typescript
  // entity-manager.test.ts
  import { createEntity } from './entity-manager';

  describe('createEntity', () => {
    it('should create a new entity with a unique ID', () => {
      const entity = createEntity();
      expect(entity.id).toBeDefined();
    });
  });
  ```

---

## Commands

| Command         | Purpose                                                                 |
|-----------------|-------------------------------------------------------------------------|
| /add-spec       | Add or update a design specification and register it in the index        |
| /add-draft-spec | Add both draft and design specification documents in parallel            |
| /enrich-specs   | Batch update multiple specs with new patterns or cross-references        |
| /add-go-specs   | Add or update Go implementation specs and register them in the index     |
| /add-agent-skill| Add or update agent skills or rules for development automation           |
```

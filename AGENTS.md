# Agent Rules

## 1. Language Policy

Consistency in communication and code is paramount.

### 1.1 Technical Content (English ONLY)

- **Codebase**: All identifiers (variables, functions, classes), comments, and docstrings.
- **Documentation**: Technical guides, READMEs, and implementation notes.
- **Process**: Commit messages, PR descriptions, and issue titles.
- **Environment**: Error messages, logs, and API definitions.

### 1.2 Communication (Russian)

- **Chat Interaction**: Discussions, explanations, and project planning.
- **Decision Making**: Strategic choices and high-level feature discussions.
- **Reviews**: Conversational feedback during pair programming.

## 2. Markdown Guidelines

- **Separators**: Avoid horizontal rules (`---`). Use them only in the footer if absolutely necessary.

## 3. Technology Stack

- **Language**: Go (latest stable version).
- **Dependencies**: Prefer standard library. Third-party packages require explicit justification.

## 4. Completion Protocol (Mandatory Checklist)

Before finishing any task, the agent MUST verify the following:

- [ ] **Technical Language**: All code, identifiers, comments, and technical docs are in English.
- [ ] **Communication Language**: All conversational responses and planning are in Russian.
- [ ] **Technology Stack**: Code is written in Go (latest stable), prioritizing the standard library.
- [ ] **SDD Integrity**:
  - [ ] No implementation code in `.design/` specifications (pseudo-code only).
  - [ ] `INDEX.md` and `PLAN.md` are updated following any spec changes.
  - [ ] Status transitions and versioning follow `RULES.md` protocol.
  - [ ] C26 Correlation: L1 specs link to corresponding `examples/` dir.
- [ ] **Cognitive Discipline**: No steps skipped, no assumptions made without asking.
- [ ] **Rule Synchronization**: New agent-facing conventions from `.design/RULES.md` are added to this checklist.
- [ ] **ECS Architecture Reference**: Skill loaded before any spec work; no external engine branding in project files.
- [ ] **Visual Excellence**: Web/UI components (if any) follow premium design guidelines.
- [ ] **Formatting**: No horizontal rules (---) used except in footers.

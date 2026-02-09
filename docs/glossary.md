# ORC Glossary

**Status**: Living document
**Last Updated**: 2026-02-09

A-Z definitions of ORC terminology. For schema details see [schema.md](schema.md). For lifecycle states see [shipment-lifecycle.md](shipment-lifecycle.md).

---

## Terms

**Approval**
A Goblin's sign-off on an IMP's implementation plan. Required before code changes.

**Commission**
A body of work being tracked. Top-level organizational unit. Contains shipments.

**El Presidente**
The human. Strategic decision maker and boss. Commands the forest.

**Factory**
A collection of workshops, typically representing a codebase or project area.

**Gatehouse**
The Goblin's workspace within a workshop. Coordination point for reviews and escalations.

**Goblin**
Workshop gatekeeper. Reviews plans, handles escalations, coordinates across workbenches. Does not write code.

**Handoff**
Session context snapshot for continuity between Claude sessions.

**IMP**
Implementation agent. Works in a workbench to implement features, fix bugs, complete tasks.

**Note**
Captured thought within a shipment. Types: idea, question, finding, decision, concern, spec.

**Plan**
C4-level implementation detail created by IMP. Specifies files and functions to edit.

**Receipt**
Proof of task completion. Created by `/imp-rec` after implementation.

**Shipment**
Unit of work with exploration → implementation lifecycle. Contains tasks and notes.

**Task**
Specific implementation work within a shipment. C2/C3 scope (what systems to touch).

**Tome**
Knowledge container at commission level. Holds notes for long-running reference.

**Watchdog**
IMP monitor. Tracks progress and reports anomalies. One per workbench.

**Workbench**
Git worktree where an IMP works. Isolated development environment with dedicated tmux window.

**Workshop**
Collection of workbenches for coordinated work. Has one gatehouse and many workbenches.

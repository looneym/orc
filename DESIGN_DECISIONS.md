# 🎯 ORC Design Decisions - Active Brainstorming

**Status**: In Design Phase
**Date**: 2026-01-13
**Purpose**: Track key design decisions as we iterate on Forest Factory architecture

---

## ✅ Decided Principles

### Decision 1: Expeditions as Coordination Primitive

**Problem**: Cross-repo work needs coordination (inspired by Gastown convoys)

**Solution**: **Forest Expeditions**

**Key Insight from El Presidente**:
> "Expeditions help coordinate WHO/HOW while work orders focus on WHAT"

**Model**: Simple 1:1:1 mapping
```
When work order "launched":
1 Work Order = 1 Expedition = 1 Grove

Work Order: "Implement OAuth refresh tokens" (WHAT)
Expedition: Coordinates WHO (IMP-Alpha) and HOW (which repos, groves)
Grove: Physical workspace where IMP works
```

**Why This Works**:
- **Clear separation of concerns**: Work order = requirements, Expedition = coordination
- **Simple model**: No complex many-to-many relationships initially
- **Scales naturally**: Can extend to multi-grove expeditions later if needed
- **Forest metaphor**: "Launch an expedition" = start work on an order

**Expedition Responsibilities**:
- WHO: Which IMP is assigned?
- HOW: Which repos/groves are involved?
- COORDINATION: How do discoveries flow between groves (future: multi-grove)?
- STATUS: Overall expedition health/progress

**Work Order Responsibilities**:
- WHAT: What needs to be done?
- WHY: Acceptance criteria, context
- DELIVERABLES: PRs, documentation, etc.

### Decision 2: Database Strategy - Graphiti First, No Rush

**Guidance from El Presidente**:
> "Be careful though as we are using Graphiti and I would prefer to keep a single DB if possible. Don't commit to a secondary DB plan just yet."

**Current Approach**:
- ✅ **Graphiti + Neo4j**: Semantic memory, decisions, discoveries, "why" knowledge
- ❓ **Structured data**: TBD - might fit in Graphiti, might need separate SQLite
- ⏸️ **Don't rush**: Explore Graphiti's capabilities before adding complexity

**Open Questions**:
1. Can Graphiti handle structured queries well enough? ("Show all in-progress work orders")
2. Do we need deterministic CRUD, or is semantic search sufficient?
3. Can expedition/grove/work-order relationships live in Graphiti graph?

**Next Steps**:
- Experiment with Graphiti for structured data
- See if episodes + entities + facts can handle expedition coordination
- Only add secondary DB if Graphiti proves insufficient

**Philosophy**: Start simple, add complexity only when necessary

### Decision 3: ORC CLI Exposed to Claude

**Agreement**: ORC CLI as primary interface, exposed to Claude

**El Presidente's guidance**:
> "Agree on a CLI that's exposed to Claude either via slash commands or skills"

**Integration Paths**:

**Option A: Slash Commands**
```
/orc expedition create "OAuth Implementation"
/orc grove create ml-auth-backend --repos intercom
/orc work create "Implement refresh tokens"
/orc status
```

**Option B: Skills**
```
Uses Skill tool with:
- skill: "orc"
- args: "expedition create 'OAuth Implementation'"
```

**Option C: MCP Server**
```
Claude uses MCP tools:
- mcp__orc__create_expedition
- mcp__orc__create_grove
- mcp__orc__list_work_orders
```

**Decision**: Explore all three, likely hybrid approach
- Slash commands for common operations
- Skills for complex workflows
- MCP for programmatic access from agents

**Key Requirement**: Must feel natural in Claude conversation flow

### Decision 4: Cross-Grove Discovery Sharing via ORC

**El Presidente's guidance**:
> "For mediating sharing we'll have to think about it but it should go through ORC and it uses the expedition as the coordination primitive"

**Principle**: ORC mediates, Expedition coordinates

**Model** (To be refined):
```
IMP-Alpha (Grove A) discovers: "Redis pub/sub pattern needed"
    ↓
    Stores in Graphiti with expedition context
    ↓
ORC notices discovery relevant to expedition
    ↓
When El Presidente switches to Grove B (IMP-Beta):
    ↓
/g-bootstrap surfaces: "IMP-Alpha discovered Redis pub/sub in Grove A"
```

**Key Points**:
- Discovery storage: Graphiti (semantic knowledge)
- Discovery routing: Via expedition metadata
- Discovery surfacing: During /g-bootstrap in related groves
- Mediator: ORC coordinates sharing, doesn't push unsolicited

**Open Question**: How does ORC determine "discovery X is relevant to grove Y"?
- Via expedition linking?
- Via semantic similarity in Graphiti?
- Explicit tagging by IMP?

---

## 🤔 Open Design Questions

### Question 1: Expedition Lifecycle Details

**When does expedition get created?**
- Option A: When work order moves from backlog → next?
- Option B: When El Presidente explicitly "launches" work?
- Option C: Automatically when grove created?

**Can expeditions span multiple work orders?**
- Current: 1:1 mapping (one work order = one expedition)
- Future: Could epic-level expedition coordinate multiple work orders?

**How does expedition "complete"?**
- Work order complete → expedition complete?
- Or expedition can outlive work order (for follow-up work)?

### Question 2: Graphiti as Primary Database?

**Can Graphiti handle**:
- Expedition → Grove relationships?
- Work order status tracking?
- IMP assignments?
- Fast queries like "show all in-progress work"?

**Testing needed**:
- Create expedition as Graphiti episode
- Link groves as entities with relationships
- Query: "What groves are part of expedition X?"
- Query: "What's the status of all expeditions?"

**If Graphiti insufficient**:
- Lightweight SQLite for structured data?
- Keep coordination metadata separate from semantic memory?

### Question 3: ORC CLI Implementation Path

**Build as**:
- Standalone binary (Go/Rust)?
- Python script using Graphiti SDK?
- Shell script wrapper around existing tools?
- MCP server with CLI frontend?

**Integration priority**:
- Start with slash commands (fastest to prototype)?
- Or build CLI first, then expose via slash commands?

### Question 4: Multi-Grove Expeditions (Future)

**Current**: 1 work order = 1 expedition = 1 grove (simple)

**Future scenario**:
```
Epic: "Implement OAuth 2.0"
    ├── WO-042: Backend token handling → Grove A (intercom)
    ├── WO-043: Infrastructure secrets → Grove B (infrastructure)
    └── WO-044: Frontend auth flow → Grove C (muster)

Should this be:
- 3 separate expeditions (current model)?
- 1 expedition with 3 groves (future model)?
```

**Design tension**:
- Simple: Keep 1:1:1 mapping always
- Flexible: Support 1 expedition : N groves for epics

**Lean towards**: Start simple (1:1:1), extend to 1:N only when needed

---

## 📝 Lessons from Gastown

### What We're Appropriating ✅

**1. Convoys → Expeditions**
- Cross-repo coordination container
- Good concept, adapted to forest metaphor

**2. TMux as UI Layer**
- Validation that our approach is sound
- Multi-window sessions work well

### What We're Rejecting ❌

**1. Distributed Agent Pool Model**
- Gastown: Autonomous agents pull from queue
- ORC: Collaborative focused sessions with El Presidente
- We're NOT building a swarm, we're building a partnership tool

**2. Scattered Configuration**
- Gastown: Config per project
- ORC: Centralized ~/.orc/ for everything
- "It's just for me and my laptop" - no distributed complexity

**3. Autonomous Task Completion**
- Gastown: Agent works, reports back, moves on
- ORC: Long-running sessions with real-time collaboration
- More like "pair programming" than "task delegation"

### What We're Learning From 🎓

**1. Bugginess Warning**
- Gastown is buggy and opinionated
- ORC should be MORE opinionated but LESS buggy
- Stability > feature richness

**2. Cross-Repo Coordination**
- Real need validated by Gastown's convoy concept
- But implement OUR way (expeditions via centralized coordination)

**3. Complexity Creep**
- Gastown's distributed model adds complexity
- ORC stays simple: centralized, single-user, local-first

---

## 🎯 Next Design Decisions Needed

### Priority 1: Database Architecture
**Question**: Can Graphiti handle structured expedition/grove/work-order data?
**Action**: Prototype expedition storage in Graphiti
**Timeline**: Before committing to secondary database

### Priority 2: ORC CLI Interface Design
**Question**: What does `orc` command look like?
**Action**: Draft CLI command structure and subcommands
**Timeline**: Before implementation

### Priority 3: Expedition Launch Workflow
**Question**: How does work order → expedition → grove creation flow?
**Action**: Define state machine and transitions
**Timeline**: Critical for implementation

### Priority 4: Discovery Sharing Mechanism
**Question**: How do cross-grove discoveries surface to relevant IMPs?
**Action**: Design ORC-mediated sharing protocol
**Timeline**: Can defer initially, but document approach

---

## 🌲 Design Principles (Reinforced)

### From North Star

1. **Personality Over Blandness**: Expeditions, not "projects"
2. **Simple First**: 1:1:1 mapping before complex relationships
3. **Centralized**: Single database, single CLI, single source of truth
4. **Claude-Native**: Everything exposed naturally to Claude
5. **Collaborative Not Autonomous**: Partnership tool, not swarm

### From Gastown Experience

6. **Avoid Distributed Complexity**: Local-first, single-user
7. **Stability Over Features**: Working simply > fancy but broken
8. **Long-Running Sessions**: Deep collaboration, not task delegation

---

## 📊 Current Architecture Sketch

```
┌─────────────────────────────────────────────────┐
│          🏛️ El Presidente                       │
│     (Supreme Commander, in conversation)        │
└────────────────┬────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────┐
│          🧙‍♂️ ORC (Orchestrator)                  │
│   • Coordinates expeditions                     │
│   • Manages grove lifecycle                     │
│   • Mediates cross-grove discoveries           │
│   • Exposed via CLI/Slash Commands/Skills       │
└────────────────┬────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────┐
│          🧠 Memory Layer                         │
│   • Graphiti + Neo4j (semantic, "why")         │
│   • Episodes: decisions, discoveries            │
│   • Entities: expeditions, groves, IMPs        │
│   • Facts: relationships, dependencies          │
│   (Secondary DB only if Graphiti insufficient)  │
└────────────────┬────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────┐
│      🎯 Expeditions (Coordination Layer)        │
│   Expedition: "OAuth Implementation"            │
│   ├─ WHO: IMP-Alpha                            │
│   ├─ HOW: Grove ml-auth-backend                │
│   ├─ WHAT: Work Order WO-042                   │
│   └─ DISCOVERIES: Cross-grove learnings        │
└────────────────┬────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────┐
│      🌲 Groves (Physical Workspaces)            │
│   Grove: ml-auth-backend                        │
│   ├─ Location: ~/src/worktrees/ml-auth-backend │
│   ├─ Repos: intercom                           │
│   ├─ TMux: Window with 3-pane layout           │
│   └─ IMP: IMP-Alpha (Claude session)           │
└────────────────┬────────────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────────────┐
│      👹 IMPs (Implementation Agents)            │
│   • Work in groves                              │
│   • Collaborate with El Presidente              │
│   • Store discoveries in Graphiti              │
│   • Context via /g-bootstrap                   │
└─────────────────────────────────────────────────┘
```

---

**Status**: Active brainstorming, evolving design
**Next**: Prototype expedition storage in Graphiti, validate data model
**Living Document**: Update as decisions crystallize

*Last Updated: 2026-01-13*

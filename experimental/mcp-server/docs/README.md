# ORC Task Management System

**Custom MCP-Based Task Coordination for ORC Ecosystem**

This directory contains the complete design and implementation plan for building a custom task management system specifically for El Presidente's ORC development workflow.

## Problem Statement

No existing task management tool provides native Claude Code CLI integration with worktree awareness, TMux session coordination, and MCP-based cross-agent communication.

## Solution Overview

Build a custom Rails-based MCP server using FastMCP gem that natively understands ORC's worktree architecture and TMux workflow patterns.

## Documentation Structure

### Core Design Documents
- **[orc-task-management-system.md](orc-task-management-system.md)** - Complete system specification, architecture, and implementation strategy
- **[rails-minimal-mcp-setup.md](rails-minimal-mcp-setup.md)** - Recommended Rails + FastMCP implementation with code examples
- **[ruby-mcp-server-architecture.md](ruby-mcp-server-architecture.md)** - Alternative Sinatra-based approach (reference)

## Key Features

### 🎯 Purpose-Built for ORC Workflow
- **Worktree Native**: Tasks automatically linked to investigations
- **TMux Integration**: Agent sessions mapped to TMux windows
- **Context Aware**: Orchestrator vs implementation agent detection
- **Git Integrated**: Branch/commit awareness and PR status

### 🔧 MCP-First Architecture
- **Claude Code Native**: Primary interface via slash commands
- **Cross-Session Communication**: Real-time coordination via centralized database
- **Natural Language**: No complex UI, just conversation with tasks
- **Multiple Transports**: STDIO and HTTP/SSE support

### 🚀 Rails + FastMCP Stack
- **Generator Setup**: `rails generate fast_mcp:install`
- **Clean Tools**: Class-based tools with built-in validation
- **ActiveRecord**: Full Rails ecosystem (migrations, validations, testing)
- **API-Only**: Minimal Rails setup focused on MCP functionality

## Implementation Timeline

### Phase 1: Core MCP Server (2 weeks)
- [ ] Rails API setup with FastMCP
- [ ] Basic task CRUD operations
- [ ] Context detection (orchestrator vs worktree)
- [ ] Multi-session HTTP/SSE support

### Phase 2: Command Integration (1 week)
- [ ] Update `/hpmboot` → `/taskboot` 
- [ ] Create `/task` universal command
- [ ] Integrate with existing ORC commands

### Phase 3: Workflow Integration (1 week)
- [ ] Auto-task creation from tech plans
- [ ] Git integration and TMux window mapping
- [ ] Progress sync back to tech plans

### Phase 4: Advanced Features (2 weeks)
- [ ] Task dependencies and cross-worktree relationships
- [ ] Time tracking and automated status reporting

## Quick Start Commands

```bash
# Create Rails API with FastMCP
rails new orc-tasks --api --database=sqlite3 --minimal
cd orc-tasks
bundle add fast-mcp
bin/rails generate fast_mcp:install
bin/rails db:migrate
bin/rails server -p 6970

# Configure Claude Code MCP
# Add to ~/.claude.json:
{
  "mcpServers": {
    "orc-tasks": {
      "type": "sse", 
      "url": "http://localhost:6970/mcp/sse"
    }
  }
}
```

## Core MCP Tools

- **`create_task`** - Orchestrator creates tasks for investigations
- **`get_my_tasks`** - Implementation agent sees context-aware tasks
- **`update_task_status`** - Progress updates with notes
- **`list_all_tasks`** - Cross-worktree visibility for orchestrator

## Success Metrics

- [ ] Replace HPM with native ORC solution
- [ ] Zero-config agent registration per worktree  
- [ ] Real-time task coordination between sessions
- [ ] Natural language task management via `/task` commands
- [ ] Full tech plan ↔ task synchronization

## Decision: Rails + FastMCP

After researching multiple Ruby MCP implementations, we selected **Rails + FastMCP** for:

✅ **Generator Magic**: One-command setup with `rails generate fast_mcp:install`  
✅ **Rails Ecosystem**: Familiar ActiveRecord, migrations, testing, console  
✅ **Clean Architecture**: Class-based tools with built-in validation  
✅ **Multiple Transports**: STDIO and HTTP/SSE for multi-session support  
✅ **Minimal Setup**: API-only Rails focused purely on MCP functionality  

This approach provides the Rails conventions El Presidente prefers with excellent MCP integration and no compromises on functionality.

---

**Status**: Ready for implementation  
**Next Step**: Create Rails prototype with FastMCP integration
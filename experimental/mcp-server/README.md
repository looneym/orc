# ORC MCP Task Management Server

**Experimental Rails-based MCP Server for Claude Code Integration**

This directory contains an experimental task management system built with Rails and FastMCP, designed to provide native Claude Code CLI integration with ORC's worktree architecture.

## Status: Experimental

This MCP server has foundational components built but its purpose within the ORC ecosystem is still evolving. The core ORC workflow (worktree management, global commands, tech plans) takes priority.

## Architecture

### Rails + FastMCP Stack
- **FastMCP Integration**: MCP protocol support via fast-mcp gem
- **Database Models**: Tasks, Worktrees, Repositories, Task History
- **HTTP/SSE Transport**: Multi-session Claude Code support
- **Tool-based API**: MCP tools for task management operations

### Core Components

#### Models
- **Task**: Individual work items with status tracking
- **Worktree**: Integration with ORC worktree architecture
- **Repository**: Git repository management
- **TaskHistory**: Audit trail for task changes

#### MCP Tools
- **CreateTaskTool**: Orchestrator creates tasks for investigations
- **GetMyTasksTool**: Context-aware task retrieval for agents
- **UpdateTaskTool**: Progress updates with notes
- **GlobalStatusTool**: Cross-worktree visibility

## Documentation

Comprehensive design documentation in `docs/`:
- **[System Overview](docs/orc-task-management-system.md)** - Complete specification
- **[Rails + FastMCP Setup](docs/rails-minimal-mcp-setup.md)** - Implementation guide
- **[Domain Model](docs/domain-model.md)** - Data architecture
- **[MCP Integration](docs/mcp-crash-course.md)** - Protocol overview

## Development Setup

```bash
cd experimental/mcp-server

# Install dependencies
bundle install

# Setup database
bin/rails db:migrate
bin/rails db:seed

# Start MCP server
bin/rails server -p 6970
```

### Claude Code Integration

Add to `~/.claude.json`:
```json
{
  "mcpServers": {
    "orc-tasks": {
      "type": "sse",
      "url": "http://localhost:6970/mcp/sse"
    }
  }
}
```

## Original Vision

The system was designed to provide:
- **Worktree Native**: Tasks automatically linked to investigations
- **TMux Integration**: Agent sessions mapped to TMux windows  
- **Context Aware**: Orchestrator vs implementation agent detection
- **Real-time Coordination**: Cross-session task communication

## Current Status

**Foundations Complete**:
- ✅ Rails API with FastMCP integration
- ✅ Core models and database schema
- ✅ Basic MCP tools implementation
- ✅ HTTP/SSE transport support

**Purpose Undefined**:
- 🔄 Integration with existing ORC workflow unclear
- 🔄 Value proposition vs tech-plans system undefined
- 🔄 Relationship to global commands not established

## Future Considerations

This experimental system may evolve into:
1. **Enhanced Tech Plans**: MCP-powered planning with real-time sync
2. **Task Coordination**: Multi-agent task assignment and tracking
3. **Workflow Automation**: Automated tech plan → task → PR workflows
4. **Cross-Session Communication**: Real-time coordination between Claude instances

Or it may be archived if the core ORC workflow proves sufficient.

---

**Status**: Experimental foundations preserved, purpose evolving with ORC ecosystem needs.
# MCP Server Crash Course

**Understanding Model Context Protocol for ORC Task Management**

## What is MCP vs REST API?

### MCP is NOT REST
MCP (Model Context Protocol) is fundamentally different from REST:

- **REST**: Resource-based (GET /tasks, POST /tasks/1, etc.)
- **MCP**: Tool-based (function calls AI can make)
- **Purpose**: Designed specifically for AI model interaction, not human API consumption

### MCP Architecture
```
┌─────────────┐    JSON-RPC 2.0    ┌─────────────┐
│             │◄─────────────────►│             │
│ Claude Code │                   │ MCP Server  │
│   (Client)  │    Tool Calls     │   (Your     │
│             │                   │    App)     │  
└─────────────┘                   └─────────────┘
```

## Core MCP Concepts

### 1. Tools (Functions AI Can Call)
Tools are functions that AI models can discover and execute:

```ruby
class CreateTaskTool < FastMcp::Tool
  description "Create a new task for investigation"
  
  arguments do
    required(:title).filled(:string)
    required(:worktree).filled(:string) 
    optional(:priority).filled(:string, included_in?: %w[low medium high])
  end

  def call(title:, worktree:, priority: 'medium')
    # Your business logic here
    Task.create!(title: title, worktree: worktree, priority: priority)
  end
end
```

**What happens:**
1. Claude Code discovers this tool exists
2. When user says "Create a task to fix the DLQ bug in ml-dlqbot-ems"
3. Claude automatically calls `CreateTaskTool` with extracted parameters
4. Your Ruby code executes and returns results

### 2. Resources (Data AI Can Access)
Resources are read-only data sources:

```ruby
class WorktreeStatusResource < FastMcp::Resource
  name "worktree_status"
  description "Current status of all active worktrees"
  
  def call
    Worktree.active.map do |wt|
      {
        name: wt.name,
        repository: wt.repository,
        active_tasks: wt.tasks.in_progress.count,
        last_commit: wt.last_commit_message
      }
    end
  end
end
```

### 3. Prompts (AI Conversation Templates)
Predefined conversation starters (less commonly used).

## MCP Protocol Flow

### 1. Discovery Phase
```javascript
// Claude Code asks: "What tools do you have?"
{
  "jsonrpc": "2.0", 
  "method": "tools/list",
  "id": 1
}

// Your server responds:
{
  "jsonrpc": "2.0",
  "result": {
    "tools": [
      {
        "name": "create_task",
        "description": "Create a new task",
        "inputSchema": {
          "type": "object",
          "properties": {
            "title": {"type": "string"},
            "worktree": {"type": "string"}
          },
          "required": ["title", "worktree"]
        }
      }
    ]
  },
  "id": 1
}
```

### 2. Execution Phase
```javascript
// Claude Code calls your tool:
{
  "jsonrpc": "2.0",
  "method": "tools/call", 
  "params": {
    "name": "create_task",
    "arguments": {
      "title": "Fix DLQ label length issue",
      "worktree": "ml-dlqbot-ems"
    }
  },
  "id": 2
}

// Your server executes and responds:
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "✅ Created task 'Fix DLQ label length issue' for ml-dlqbot-ems"
      }
    ]
  },
  "id": 2
}
```

## Data Modeling in MCP Context

### Yes, You Still Model Core Entities!

MCP doesn't change your data modeling - it's just the interface layer:

```ruby
# Traditional Rails models
class Task < ApplicationRecord
  belongs_to :worktree
  has_many :task_histories
  
  validates :title, presence: true
  validates :status, inclusion: { in: %w[investigating in_progress blocked completed] }
  
  scope :for_worktree, ->(wt) { where(worktree: wt) }
  scope :active, -> { where(status: ['investigating', 'in_progress']) }
end

class Worktree < ApplicationRecord
  has_many :tasks
  
  def current_branch
    # Git operations
  end
  
  def tmux_window_active?
    # TMux integration
  end
end

class TaskHistory < ApplicationRecord
  belongs_to :task
end
```

### MCP Tools Use Your Models

```ruby
class GetMyTasksTool < FastMcp::Tool
  def call(status: nil)
    # Use your Rails models normally
    worktree = detect_current_worktree()
    tasks = Task.for_worktree(worktree)
    tasks = tasks.where(status: status) if status
    
    # Return formatted data for AI
    {
      success: true,
      tasks: tasks.map(&:as_json),
      summary: format_for_claude(tasks)
    }
  end
end
```

## Key Differences from REST

| Aspect | REST API | MCP Server |
|--------|----------|------------|
| **Consumer** | Humans via HTTP clients | AI models via JSON-RPC |
| **Interface** | Resource endpoints | Function calls |
| **Discovery** | Documentation | Protocol introspection |
| **Parameters** | Query params, JSON body | Structured arguments with validation |
| **Responses** | HTTP status + JSON | Success/error with content array |
| **Authentication** | Various (JWT, OAuth, etc.) | Usually API key or none |

## MCP Transport Options

### 1. STDIO (Standard Input/Output)
- AI launches your server as subprocess
- Communication via stdin/stdout
- Good for single-session tools

### 2. HTTP/SSE (Server-Sent Events)  
- Your server runs as web service
- AI connects via HTTP
- **Better for multi-session** (our use case!)

```ruby
# FastMCP automatically provides both:
# GET  /mcp/sse     (Connection endpoint)
# POST /mcp/tools   (Tool execution)
```

## Context Awareness in MCP

### Environment Variables Available
```ruby
class ContextDetector
  def current_worktree
    # MCP server can access environment where Claude is running
    pwd = ENV['PWD'] || Dir.pwd
    return nil unless pwd.include?('/worktrees/')
    File.basename(pwd)
  end
end
```

### Tool Behavior Changes by Context
```ruby
class CreateTaskTool < FastMcp::Tool  
  def call(title:, worktree:, **args)
    context = ContextDetector.new
    
    if context.in_orchestrator_mode?
      # Create task for specified worktree
      create_orchestrator_task(title, worktree, args)
    else
      # Create task for current worktree
      current_wt = context.current_worktree
      create_implementation_task(title, current_wt, args)
    end
  end
end
```

## Natural Language Interface

### The "NLP Glue" You Mentioned
This happens automatically - no extra code needed:

```
User says: "I need to fix the DLQ bot label bug, it's high priority"

Claude Code:
1. Parses natural language
2. Maps to create_task tool
3. Extracts: title="Fix DLQ bot label bug", priority="high"
4. Calls your MCP tool with those params
5. Shows response to user naturally

You just write: CreateTaskTool.call(title: "Fix DLQ bot label bug", priority: "high")
```

## Practical Example: Task Management Flow

### User Conversation → MCP Calls

```
User: "What tasks do I have for the DLQ investigation?"

Claude Code automatically:
1. Detects current worktree (ml-dlq-investigation-ems)  
2. Calls get_my_tasks tool
3. Your server returns tasks
4. Claude formats response naturally

User: "Mark the first one as in progress, I'm starting work on it"

Claude Code automatically:
1. Identifies "first task" from previous response
2. Calls update_task_status(task_id: 123, status: "in_progress")  
3. Your server updates database
4. Claude confirms the change
```

## Why MCP vs GraphQL/REST for AI?

### MCP Advantages
- **AI-Native**: Designed for model consumption
- **Function-Based**: AI thinks in terms of actions, not resources
- **Self-Describing**: Tools describe their own parameters and validation
- **Context Aware**: Can access Claude's environment (pwd, env vars, etc.)
- **Standardized**: One protocol works with all AI models

### Traditional API Challenges for AI
- AI struggles with REST semantics (when to POST vs PUT?)  
- Complex authentication flows
- No built-in parameter validation discovery
- HTTP status codes don't map well to AI reasoning

## Summary for ORC Task Management

**You're building a traditional Rails app** with models, validations, relationships, etc.

**MCP is just the interface layer** that lets Claude Code call your Ruby methods naturally.

**The magic:** User says "Create a high priority task to fix the DLQ bug" and your Rails code automatically runs `Task.create!(title: "Fix DLQ bug", priority: "high", worktree: current_worktree)`

**No REST endpoints needed** - just Ruby classes that Claude can call as functions!

Ready to build this, El Presidente?
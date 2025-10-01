# Ruby MCP Server Architecture

**Status**: investigating

## Problem & Solution

**Current Issue:** Need to prototype ORC Task Management system with Ruby-based MCP server
**Solution:** Build HTTP MCP server using Sinatra/Roda with SQLite database for task coordination

## Ruby MCP Stack

### Core Dependencies
```ruby
# Gemfile
gem 'sinatra', '~> 3.0'          # Lightweight web framework
gem 'sqlite3', '~> 1.6'          # Database
gem 'sequel', '~> 5.0'           # ORM with migrations
gem 'json', '~> 2.6'             # JSON-RPC handling
gem 'puma', '~> 6.0'             # Web server
gem 'dry-validation', '~> 1.10'  # Input validation
gem 'zeitwerk', '~> 2.6'         # Autoloading
gem 'logger', '~> 1.5'           # Structured logging
```

### Project Structure
```
orc/task-management/
├── Gemfile
├── config.ru                   # Rack config
├── server.rb                   # Main Sinatra app
├── lib/
│   ├── orc_tasks/
│   │   ├── models/
│   │   │   ├── task.rb
│   │   │   ├── worktree.rb
│   │   │   └── task_history.rb
│   │   ├── services/
│   │   │   ├── context_detector.rb
│   │   │   ├── mcp_handler.rb
│   │   │   └── task_coordinator.rb
│   │   ├── database.rb         # Sequel setup
│   │   └── mcp_tools.rb        # Tool definitions
├── db/
│   ├── migrations/
│   └── tasks.db               # SQLite database
├── config/
│   └── database.yml
└── spec/                      # RSpec tests
```

## Core Server Implementation

### Main Sinatra Application
```ruby
# server.rb
require 'sinatra'
require 'json'
require 'logger'
require_relative 'lib/orc_tasks'

class OrcTaskServer < Sinatra::Base
  configure do
    set :port, ENV.fetch('PORT', 6970)
    set :bind, '127.0.0.1'
    set :logging, true
    
    # Setup database
    OrcTasks::Database.setup!
  end

  before do
    content_type :json
    
    # API key authentication
    api_key = request.env['HTTP_X_API_KEY']
    halt 401, {error: 'API key required'}.to_json unless valid_api_key?(api_key)
  end

  # MCP protocol endpoint
  post '/' do
    handle_mcp_request(JSON.parse(request.body.read))
  end

  # Health check
  get '/health' do
    {status: 'ok', version: '1.0.0'}.to_json
  end

  private

  def handle_mcp_request(body)
    handler = OrcTasks::McpHandler.new
    
    case body['method']
    when 'initialize'
      handler.initialize_session(body)
    when 'tools/list'
      handler.list_tools(body)
    when 'tools/call'
      handler.call_tool(body)
    else
      error_response(body['id'], -32601, "Method not found: #{body['method']}")
    end
  rescue => e
    logger.error "MCP request failed: #{e.message}"
    error_response(body['id'], -32603, e.message)
  end

  def valid_api_key?(key)
    # Simple API key validation
    key == ENV.fetch('ORC_TASK_API_KEY', 'orc-tasks-secret')
  end

  def error_response(id, code, message)
    {
      jsonrpc: '2.0',
      error: {code: code, message: message},
      id: id
    }.to_json
  end
end
```

### MCP Protocol Handler
```ruby
# lib/orc_tasks/mcp_handler.rb
module OrcTasks
  class McpHandler
    def initialize
      @context = ContextDetector.new
      @coordinator = TaskCoordinator.new
    end

    def initialize_session(body)
      {
        jsonrpc: '2.0',
        result: {
          protocolVersion: '2024-11-05',
          capabilities: {
            tools: {},
            resources: {}
          },
          serverInfo: {
            name: 'orc-tasks',
            version: '1.0.0'
          }
        },
        id: body['id']
      }.to_json
    end

    def list_tools(body)
      tools = [
        # Orchestrator tools
        {
          name: 'create_task',
          description: 'Create new task for investigation',
          inputSchema: {
            type: 'object',
            properties: {
              title: {type: 'string'},
              description: {type: 'string'},
              worktree: {type: 'string'},
              priority: {type: 'string', enum: ['low', 'medium', 'high'], default: 'medium'},
              repository: {type: 'string'}
            },
            required: ['title', 'worktree']
          }
        },

        # Implementation tools  
        {
          name: 'get_my_tasks',
          description: 'Get tasks for current worktree context'
        },

        {
          name: 'update_task_status',
          description: 'Update task status and add notes',
          inputSchema: {
            type: 'object',
            properties: {
              task_id: {type: 'string'},
              status: {type: 'string', enum: ['investigating', 'in_progress', 'blocked', 'completed']},
              notes: {type: 'string'}
            },
            required: ['task_id', 'status']
          }
        },

        # Cross-worktree visibility
        {
          name: 'list_all_tasks',
          description: 'List all tasks (orchestrator context)',
          inputSchema: {
            type: 'object',
            properties: {
              status: {type: 'string', enum: ['investigating', 'in_progress', 'blocked', 'completed']},
              worktree: {type: 'string'}
            }
          }
        }
      ]

      {
        jsonrpc: '2.0',
        result: {tools: tools},
        id: body['id']
      }.to_json
    end

    def call_tool(body)
      tool_name = body.dig('params', 'name')
      arguments = body.dig('params', 'arguments') || {}
      
      result = case tool_name
               when 'create_task'
                 @coordinator.create_task(arguments)
               when 'get_my_tasks'  
                 @coordinator.get_context_tasks(@context.current_worktree)
               when 'update_task_status'
                 @coordinator.update_task(arguments['task_id'], arguments)
               when 'list_all_tasks'
                 @coordinator.list_tasks(arguments)
               else
                 {error: "Unknown tool: #{tool_name}"}
               end

      {
        jsonrpc: '2.0',
        result: {
          content: [
            {
              type: 'text', 
              text: format_tool_result(result)
            }
          ]
        },
        id: body['id']
      }.to_json
    end

    private

    def format_tool_result(result)
      if result.is_a?(Hash) && result[:error]
        "❌ #{result[:error]}"
      elsif result.is_a?(Array)
        format_task_list(result)
      else
        JSON.pretty_generate(result)
      end
    end

    def format_task_list(tasks)
      return "No tasks found." if tasks.empty?
      
      tasks.map do |task|
        status_emoji = {
          'investigating' => '🔍',
          'in_progress' => '⚡',
          'blocked' => '🚫', 
          'completed' => '✅'
        }[task[:status]] || '📋'
        
        "#{status_emoji} **#{task[:title]}** (#{task[:worktree]})\n" +
        "   Status: #{task[:status]} | Priority: #{task[:priority]}\n" +
        "   #{task[:description]}\n"
      end.join("\n")
    end
  end
end
```

### Database Models
```ruby
# lib/orc_tasks/models/task.rb
module OrcTasks
  class Task < Sequel::Model
    plugin :validation_helpers
    plugin :timestamps, update_on_create: true

    def validate
      super
      validates_presence [:title, :worktree]
      validates_includes %w[investigating in_progress blocked completed], :status
      validates_includes %w[low medium high], :priority
    end

    def to_hash
      {
        id: id,
        title: title,
        description: description,
        status: status,
        priority: priority,
        worktree: worktree,
        repository: repository,
        branch: branch,
        assigned_agent: assigned_agent,
        created_by: created_by,
        created_at: created_at,
        updated_at: updated_at,
        context: context ? JSON.parse(context) : {}
      }
    end
  end
end
```

### Context Detection Service  
```ruby
# lib/orc_tasks/services/context_detector.rb
module OrcTasks
  class ContextDetector
    def current_worktree
      pwd = ENV['PWD'] || Dir.pwd
      
      # Check if we're in a worktree
      if pwd.include?('/worktrees/')
        File.basename(pwd)
      else
        nil # Orchestrator context
      end
    end

    def current_repository
      return nil unless in_worktree?
      
      # Extract repo from worktree name (ml-investigation-repo)
      current_worktree&.split('-')&.last
    end

    def in_worktree?
      !current_worktree.nil?
    end

    def in_orchestrator_context?
      ENV['PWD']&.include?('/orc') || Dir.pwd.include?('/orc')
    end

    def current_branch
      return nil unless in_worktree?
      `git branch --show-current`.strip
    rescue
      nil
    end
  end
end
```

### Task Coordinator Service
```ruby
# lib/orc_tasks/services/task_coordinator.rb
module OrcTasks
  class TaskCoordinator
    def create_task(params)
      task = Task.create(
        title: params['title'],
        description: params['description'],
        worktree: params['worktree'],
        priority: params['priority'] || 'medium',
        repository: params['repository'],
        status: 'investigating',
        created_by: 'orchestrator',
        context: (params['context'] || {}).to_json
      )
      
      task.to_hash
    end

    def get_context_tasks(worktree)
      return {error: "No worktree context detected"} unless worktree
      
      Task.where(worktree: worktree).all.map(&:to_hash)
    end

    def update_task(task_id, params)
      task = Task[task_id]
      return {error: "Task not found"} unless task

      task.update(
        status: params['status'],
        updated_at: Time.now
      )

      # Add history entry if notes provided
      if params['notes']
        TaskHistory.create(
          task_id: task_id,
          action: 'status_update',
          notes: params['notes']
        )
      end

      task.to_hash
    end

    def list_tasks(filters = {})
      dataset = Task.dataset
      dataset = dataset.where(status: filters['status']) if filters['status']
      dataset = dataset.where(worktree: filters['worktree']) if filters['worktree']
      
      dataset.all.map(&:to_hash)
    end
  end
end
```

## Database Migrations

```ruby
# db/migrations/001_create_tasks.rb
Sequel.migration do
  up do
    create_table(:tasks) do
      String :id, primary_key: true, default: Sequel.function(:hex, Sequel.function(:randomblob, 16))
      String :title, null: false
      Text :description
      String :status, null: false, default: 'investigating'
      String :priority, null: false, default: 'medium'
      String :worktree
      String :repository  
      String :branch
      String :assigned_agent
      String :created_by
      DateTime :created_at, default: Sequel::CURRENT_TIMESTAMP
      DateTime :updated_at, default: Sequel::CURRENT_TIMESTAMP
      Text :context # JSON string
      
      index :status
      index :worktree
      index :created_at
    end
  end

  down do
    drop_table(:tasks)
  end
end
```

## Startup and Configuration

### Rack Config
```ruby
# config.ru
require_relative 'server'

# Set environment variables
ENV['ORC_TASK_API_KEY'] ||= 'orc-dev-key'
ENV['DATABASE_URL'] ||= 'sqlite://db/tasks.db'

run OrcTaskServer
```

### Development Startup
```bash
# Setup
bundle install
bundle exec sequel -m db/migrations sqlite://db/tasks.db

# Run server
bundle exec rackup -p 6970
```

## MCP Client Configuration

```json
// ~/.claude.json
{
  "mcpServers": {
    "orc-tasks": {
      "type": "http",
      "url": "http://localhost:6970",
      "headers": {
        "X-API-Key": "orc-dev-key"
      }
    }
  }
}
```

## Testing Strategy

### RSpec Setup
```ruby
# spec/spec_helper.rb
require_relative '../server'
require 'rack/test'

RSpec.configure do |config|
  config.include Rack::Test::Methods
  
  def app
    OrcTaskServer
  end
  
  config.before(:suite) do
    ENV['DATABASE_URL'] = 'sqlite://db/test.db'
    OrcTasks::Database.setup!
  end
  
  config.before(:each) do
    # Clean database
    OrcTasks::Task.truncate
  end
end
```

## Advantages of Ruby Stack

1. **Familiar**: You know Ruby well
2. **Simple**: Sinatra is lightweight and direct  
3. **Database**: Sequel ORM with excellent migration support
4. **Testing**: RSpec for comprehensive test coverage
5. **Deployment**: Easy to containerize or run as service

Ready to start building the prototype, El Presidente?
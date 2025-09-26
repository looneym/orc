require 'fast_mcp'
require 'mcp/transports/rack_transport'

# FastMCP Configuration for ORC Task Management
Rails.application.config.after_initialize do
  # Create MCP server instance
  $mcp_server = MCP::Server.new(name: 'orc-tasks', version: '1.0.0')
  
  # Load all tools and resources
  Dir[Rails.root.join('app/tools/**/*.rb')].each { |f| require f }
  Dir[Rails.root.join('app/resources/**/*.rb')].each { |f| require f }
  
  # Register tools with server
  if defined?(ApplicationTool)
    ApplicationTool.descendants.each do |tool_class|
      $mcp_server.register_tool(tool_class)
    end
  end
  
  # Register resources with server
  if defined?(ApplicationResource)
    ApplicationResource.descendants.each do |resource_class|
      $mcp_server.register_resource(resource_class.new)
    end
  end
  
  Rails.logger.info "🚀 FastMCP server initialized with #{$mcp_server.tools.count} tools"
end
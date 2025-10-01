class TestTool < ApplicationTool
  description "Test tool to verify ORC Task Management MCP server is working"
  
  arguments do
    optional(:message).filled(:string).description("Test message to echo back")
  end
  
  def call(message: "ORC Task Management MCP server is working!")
    {
      success: true,
      message: message,
      timestamp: Time.current.iso8601,
      server: "ORC Task Management v1.0.0"
    }
  end
end
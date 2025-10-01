class ApplicationTool < MCP::Tool
  # Base class for all ORC Task Management MCP tools
  
  protected
  
  def current_context
    @current_context ||= ContextDetector.new
  end
  
  def task_coordinator
    @task_coordinator ||= TaskCoordinator.new
  end
end
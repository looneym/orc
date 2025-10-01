# FastMCP Middleware for Rails Integration
class FastMcpMiddleware
  def initialize(app)
    @app = app
    @mcp_server = nil
  end

  def call(env)
    # Initialize MCP server if not already done
    if @mcp_server.nil? && defined?($mcp_server)
      @mcp_server = $mcp_server
      Rails.logger.info "✅ FastMCP middleware initialized at /mcp"
    end

    request = Rack::Request.new(env)
    
    # Handle MCP endpoints directly
    if request.path.start_with?('/mcp') && @mcp_server
      handle_mcp_request(request, env)
    else
      @app.call(env)
    end
  end

  private

  def handle_mcp_request(request, env)
    subpath = request.path[4..] # Remove '/mcp' prefix
    
    case subpath
    when '/messages', '' # Handle both /mcp/messages and /mcp
      handle_json_rpc_request(request)
    when '/sse'
      handle_sse_request(request, env)
    else
      [404, { 'Content-Type' => 'application/json' }, 
       [JSON.generate({ jsonrpc: '2.0', error: { code: -32601, message: 'Endpoint not found' }, id: nil })]]
    end
  end

  def handle_json_rpc_request(request)
    return [405, {}, ['Method not allowed']] unless request.post?
    
    # Check for Authorization header
    auth_header = request.get_header('HTTP_AUTHORIZATION')
    if auth_header.nil? || !auth_header.start_with?('Bearer ')
      # Return 401 with WWW-Authenticate header pointing to authorization server
      return [401, {
        'Content-Type' => 'application/json',
        'WWW-Authenticate' => 'Bearer realm="MCP", resource_metadata="http://localhost:6970/.well-known/oauth-protected-resource"'
      }, [JSON.generate({
        jsonrpc: '2.0',
        error: { code: -32001, message: 'Authorization required' },
        id: nil
      })]]
    end
    
    # Extract and validate token (simplified for development)
    token = auth_header.sub('Bearer ', '')
    if token.empty?
      return [401, {
        'Content-Type' => 'application/json',
        'WWW-Authenticate' => 'Bearer realm="MCP", error="invalid_token"'
      }, [JSON.generate({
        jsonrpc: '2.0',
        error: { code: -32001, message: 'Invalid token' },
        id: nil
      })]]
    end
    
    begin
      body = request.body.read
      Rails.logger.info "Processing MCP JSON-RPC request: #{body}"
      
      # Capture the response by temporarily overriding the transport
      response_data = nil
      
      # Create a simple response capture
      original_transport = @mcp_server.instance_variable_get(:@transport)
      
      response_capture = Object.new
      response_capture.define_singleton_method(:send_message) do |message|
        response_data = message
        Rails.logger.info "Captured response: #{message.inspect}"
      end
      
      # Temporarily replace transport to capture response
      @mcp_server.instance_variable_set(:@transport, response_capture)
      
      # Process the message
      @mcp_server.handle_json_request(body)
      
      # Restore original transport
      @mcp_server.instance_variable_set(:@transport, original_transport)
      
      if response_data
        json_response = response_data.is_a?(String) ? response_data : JSON.generate(response_data)
        [200, { 'Content-Type' => 'application/json' }, [json_response]]
      else
        [500, { 'Content-Type' => 'application/json' }, 
         [JSON.generate({ jsonrpc: '2.0', error: { code: -32603, message: 'No response generated' }, id: nil })]]
      end
      
    rescue JSON::ParserError => e
      Rails.logger.error "JSON parse error: #{e.message}"
      [400, { 'Content-Type' => 'application/json' }, 
       [JSON.generate({ jsonrpc: '2.0', error: { code: -32700, message: 'Parse error' }, id: nil })]]
    rescue StandardError => e
      Rails.logger.error "Error processing MCP request: #{e.message}"
      [500, { 'Content-Type' => 'application/json' }, 
       [JSON.generate({ jsonrpc: '2.0', error: { code: -32603, message: "Internal error: #{e.message}" }, id: nil })]]
    end
  end

  def handle_sse_request(request, env)
    # Simple SSE response for now
    [200, {
      'Content-Type' => 'text/event-stream',
      'Cache-Control' => 'no-cache',
      'Connection' => 'keep-alive'
    }, [": SSE endpoint available\n\n"]]
  end
end
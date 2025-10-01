class McpOauthController < ApplicationController
  # OAuth Authorization Server Metadata
  def authorization_server
    render json: {
      issuer: "http://localhost:6970",
      authorization_endpoint: "http://localhost:6970/oauth/authorize",
      token_endpoint: "http://localhost:6970/oauth/token",
      registration_endpoint: "http://localhost:6970/register",
      grant_types_supported: ["client_credentials", "authorization_code", "refresh_token"],
      token_endpoint_auth_methods_supported: ["none", "client_secret_basic"],
      response_types_supported: ["code", "token"],
      scopes_supported: ["mcp"],
      code_challenge_methods_supported: ["S256"]
    }
  end

  # OAuth Protected Resource Metadata  
  def protected_resource
    render json: {
      resource: "http://localhost:6970/mcp",
      authorization_servers: ["http://localhost:6970"],
      scopes_supported: ["mcp"],
      bearer_methods_supported: ["header"]
    }
  end

  # MCP-specific Authorization Server Metadata
  def authorization_server_mcp
    authorization_server
  end

  # Dynamic Client Registration
  def register_client
    client_id = SecureRandom.uuid
    
    # Extract client registration parameters
    redirect_uris = params[:redirect_uris] || []
    grant_types = params[:grant_types] || ["client_credentials"]
    response_types = params[:response_types] || ["token"]
    
    render json: {
      client_id: client_id,
      client_secret: SecureRandom.hex(32),
      redirect_uris: redirect_uris,
      grant_types: grant_types,
      response_types: response_types,
      token_endpoint_auth_method: "none",
      scope: "mcp"
    }
  end

  # OAuth Authorization Endpoint
  def authorize
    # For development, auto-approve and redirect with code
    code = SecureRandom.hex(16)
    redirect_uri = params[:redirect_uri] || "http://localhost:3000/callback"
    
    redirect_to "#{redirect_uri}?code=#{code}&state=#{params[:state]}"
  end

  # OAuth Token Endpoint - Grant client_credentials tokens
  def token
    grant_type = params[:grant_type]
    
    case grant_type
    when "client_credentials"
      # Client credentials flow
      render json: {
        access_token: SecureRandom.hex(32),
        token_type: "Bearer",
        expires_in: 3600,
        scope: "mcp"
      }
    when "authorization_code"
      # Authorization code flow
      render json: {
        access_token: SecureRandom.hex(32),
        token_type: "Bearer", 
        expires_in: 3600,
        scope: "mcp",
        refresh_token: SecureRandom.hex(32)
      }
    else
      render json: {
        error: "unsupported_grant_type",
        error_description: "Grant type #{grant_type} is not supported"
      }, status: 400
    end
  end
end
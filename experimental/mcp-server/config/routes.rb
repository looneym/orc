Rails.application.routes.draw do
  # Define your application routes per the DSL in https://guides.rubyonrails.org/routing.html

  # Reveal health status on /up that returns 200 if the app boots with no exceptions, otherwise 500.
  # Can be used by load balancers and uptime monitors to verify that the app is live.
  get "up" => "rails/health#show", as: :rails_health_check

  # OAuth discovery endpoints for MCP authentication
  get ".well-known/oauth-authorization-server", to: "mcp_oauth#authorization_server"
  get ".well-known/oauth-protected-resource", to: "mcp_oauth#protected_resource"  
  get ".well-known/oauth-authorization-server/mcp", to: "mcp_oauth#authorization_server_mcp"
  
  # OAuth client registration endpoint
  post "register", to: "mcp_oauth#register_client"
  
  # OAuth authorization and token endpoints
  get "oauth/authorize", to: "mcp_oauth#authorize"
  post "oauth/token", to: "mcp_oauth#token"

  # Defines the root path route ("/")
  # root "posts#index"
end

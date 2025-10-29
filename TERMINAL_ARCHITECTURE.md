# Terminal Architecture

## Current Status
The terminal functionality should use the **existing MCP infrastructure**, not a separate WebSocket endpoint.

## Architecture

```
Frontend Terminal Panel
  ↓ HTTP/WebSocket
Go Backend API
  ↓ MCP Client
Terminal MCP Server (Python)
  ↓ subprocess
Linux Shell
```

## Existing Components

1. **Terminal MCP Server** (`backend/mcp_servers/terminal/server.py`)
   - Natural language to command conversion
   - Command execution with safety checks
   - Command explanation
   - Already running as part of start-all.sh

2. **MCP Client** (Go backend)
   - Already integrated in `internal/mcp/client.go`
   - Can call MCP tools

3. **Terminal Manager** (Go backend)
   - Exists in `internal/terminal/manager.go`
   - Should use MCP client, not direct execution

## Correct Implementation

The terminal panel should:
1. Send commands to `/api/terminal/execute` (HTTP endpoint)
2. Backend calls MCP terminal server via MCP client
3. MCP server executes command safely
4. Result returned to frontend

## Why Not Direct WebSocket?

- MCP provides safety checks
- Natural language support
- Consistent with system architecture
- Already implemented and tested
- Integrates with agent reasoning

## Next Steps

1. Remove the direct WebSocket terminal handler
2. Create HTTP API endpoint that uses MCP
3. Update frontend to use HTTP API instead of WebSocket
4. Or: Use existing agent/chat interface which already has MCP access

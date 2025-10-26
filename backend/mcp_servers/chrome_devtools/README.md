# Chrome DevTools MCP Integration

Chrome DevTools MCP server is provided by the official MCP project.
This directory contains configuration for integrating it with the agent.

## Installation

```bash
npm install -g @modelcontextprotocol/server-chrome-devtools
```

## Configuration

The Chrome DevTools MCP server provides tools for:
- Browser automation
- DOM inspection
- Network monitoring
- Console interaction
- Performance profiling
- Debugging

## MCP Client Config

```json
{
  "mcpServers": {
    "chrome-devtools": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-chrome-devtools"]
    }
  }
}
```

## Available Tools

### Browser Control
- `navigate` - Navigate to URL
- `click` - Click element
- `type` - Type text
- `screenshot` - Take screenshot

### DOM Inspection
- `query_selector` - Find elements
- `get_html` - Get page HTML
- `get_text` - Get element text

### Console
- `execute_script` - Run JavaScript
- `get_console_logs` - Get console output

### Network
- `get_network_logs` - Get network requests
- `intercept_request` - Intercept and modify requests

### Performance
- `get_performance_metrics` - Get performance data
- `start_profiling` - Start CPU/memory profiling

## Usage from Agent

```go
// Navigate to page
result, _ := mcpClient.CallTool("chrome-devtools", "navigate", map[string]interface{}{
    "url": "https://example.com",
})

// Execute JavaScript
result, _ := mcpClient.CallTool("chrome-devtools", "execute_script", map[string]interface{}{
    "script": "document.title",
})

// Take screenshot
result, _ := mcpClient.CallTool("chrome-devtools", "screenshot", map[string]interface{}{
    "path": "/tmp/screenshot.png",
})
```

## Integration with Dynamic Thinking

The agent can use Chrome DevTools for:
- **Perception**: Analyze web applications visually
- **Reasoning**: Understand page structure and behavior
- **Action**: Automate browser interactions
- **Reflection**: Analyze performance and debug issues

## Security

Chrome DevTools MCP server runs with full browser access.
Ensure proper validation before executing user-provided scripts.


# Terminal & Browser MCP Server

Unified MCP server combining terminal operations and browser automation.

## Features

### Terminal Tools (3 tools)
- **natural_to_command** - Convert natural language to Linux commands using `comanderanch/Linux-Buster:latest`
- **execute_command** - Execute commands with safety validation
- **explain_command** - Explain what commands do

### Browser Tools (8 tools via Chrome DevTools Protocol)
- **browser_navigate** - Navigate to URLs
- **browser_screenshot** - Take screenshots (optionally analyze with Gemma 3 vision)
- **browser_get_dom** - Get DOM tree structure
- **browser_click** - Click elements
- **browser_type** - Type into inputs
- **browser_execute_script** - Run JavaScript
- **browser_get_content** - Get page HTML
- **browser_get_console** - Get console messages
- **browser_get_network** - Get network activity

## Why Chrome DevTools Protocol?

**Chrome DevTools Protocol (CDP)** provides direct access to Chrome's internals:

✅ **Full DOM access** - Query, modify, inspect any element  
✅ **Network monitoring** - See all requests, responses, timing  
✅ **Console access** - Read console.log, errors, warnings  
✅ **JavaScript execution** - Run code in page context  
✅ **Screenshot capture** - Full page or viewport  
✅ **Performance metrics** - Load times, resource usage  
✅ **No overhead** - Direct WebSocket connection  

## Setup

### 1. Start Chrome with Remote Debugging

```bash
# Linux
google-chrome --remote-debugging-port=9222 --no-first-run --no-default-browser-check

# macOS
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222

# With specific user data dir
google-chrome --remote-debugging-port=9222 --user-data-dir=/tmp/chrome-debug
```

### 2. Install Dependencies

```bash
cd backend/mcp_servers/terminal_browser
pip install -r requirements.txt
```

### 3. Configure Environment

```bash
export OLLAMA_BASE_URL="http://localhost:11434"
```

### 4. Ensure Models are Pulled

```bash
# Terminal model (250+ Linux commands)
ollama pull comanderanch/Linux-Buster:latest

# Vision model (screenshot analysis)
ollama pull gemma3:27b
```

### 5. Run Server

```bash
python server.py
```

## Usage Examples

### Terminal Operations

```python
# Convert natural language to command
{
  "tool": "natural_to_command",
  "arguments": {
    "instruction": "list all python files in current directory",
    "context": "/home/user/project"
  }
}
# Returns: {"command": "find . -name '*.py'", "safe": true, ...}

# Execute command
{
  "tool": "execute_command",
  "arguments": {
    "command": "ls -la",
    "timeout": 10
  }
}
# Returns: {"success": true, "stdout": "...", "stderr": "", ...}

# Explain command
{
  "tool": "explain_command",
  "arguments": {
    "command": "grep -r 'TODO' ."
  }
}
# Returns: "Recursively searches for 'TODO' in all files..."
```

### Browser Automation

```python
# Navigate
{
  "tool": "browser_navigate",
  "arguments": {
    "url": "https://github.com"
  }
}

# Take screenshot and analyze with Gemma 3 vision
{
  "tool": "browser_screenshot",
  "arguments": {
    "analyze": true,
    "question": "What is the main heading on this page?",
    "full_page": false
  }
}
# Returns: {"screenshot_path": "/tmp/browser_screenshot.png", "analysis": "The main heading is 'Let's build from here'..."}

# Get DOM structure
{
  "tool": "browser_get_dom",
  "arguments": {}
}
# Returns: Full DOM tree via Chrome DevTools

# Click element
{
  "tool": "browser_click",
  "arguments": {
    "selector": "button.sign-in"
  }
}

# Type into input
{
  "tool": "browser_type",
  "arguments": {
    "selector": "input[name='username']",
    "text": "myusername"
  }
}

# Execute JavaScript
{
  "tool": "browser_execute_script",
  "arguments": {
    "script": "document.title"
  }
}
# Returns: {"success": true, "result": "GitHub"}

# Get console messages
{
  "tool": "browser_get_console",
  "arguments": {}
}
# Returns: [{"type": "log", "args": ["Hello"]}, ...]

# Get network activity
{
  "tool": "browser_get_network",
  "arguments": {}
}
# Returns: All network requests/responses
```

## Safety Features

### Terminal Safety
- **Dangerous command blocking** - Prevents `rm -rf /`, fork bombs, etc.
- **Pattern matching** - Regex-based dangerous command detection
- **Dry run mode** - Validate without executing
- **Timeout protection** - Commands timeout after 30s (configurable)

### Browser Safety
- **No arbitrary code execution** - All actions through CDP
- **Sandboxed JavaScript** - Runs in page context only
- **No file system access** - Browser can't access local files

## Architecture

```
Terminal & Browser MCP Server
├── Terminal Tools
│   ├── Natural language → Command (Linux-Buster model)
│   ├── Command execution (with safety checks)
│   └── Command explanation
└── Browser Tools (Chrome DevTools Protocol)
    ├── Navigation
    ├── Screenshot (+ Gemma 3 vision analysis)
    ├── DOM inspection
    ├── Element interaction (click, type)
    ├── JavaScript execution
    ├── Console monitoring
    └── Network monitoring
```

## Integration with Agent

### Gemma 3 Can Call These Tools

```python
# Agent decides to search GitHub
agent.call_tool("browser_navigate", {"url": "https://github.com/search"})

# Agent takes screenshot to see what's on page
result = agent.call_tool("browser_screenshot", {
    "analyze": true,
    "question": "Where is the search box?"
})
# Gemma 3 vision: "The search box is at the top center with placeholder 'Search GitHub'"

# Agent types search query
agent.call_tool("browser_type", {
    "selector": "input[name='q']",
    "text": "self-improving agents"
})

# Agent clicks search button
agent.call_tool("browser_click", {"selector": "button[type='submit']"})

# Agent gets results
content = agent.call_tool("browser_get_content", {})
```

### Combined Terminal + Browser Workflows

```python
# 1. Agent uses terminal to clone repo
agent.call_tool("natural_to_command", {
    "instruction": "clone the repository at https://github.com/user/repo"
})
# Returns: git clone https://github.com/user/repo

agent.call_tool("execute_command", {
    "command": "git clone https://github.com/user/repo"
})

# 2. Agent uses browser to read documentation
agent.call_tool("browser_navigate", {
    "url": "https://github.com/user/repo"
})

agent.call_tool("browser_screenshot", {
    "analyze": true,
    "question": "What does the README say about installation?"
})
# Gemma 3 analyzes screenshot and extracts installation instructions

# 3. Agent uses terminal to install
agent.call_tool("execute_command", {
    "command": "pip install -r requirements.txt"
})
```

## Chrome DevTools Protocol Benefits

### vs Playwright/Selenium

**CDP Advantages:**
- ✅ **Direct connection** - WebSocket to Chrome, no driver
- ✅ **Full access** - Everything Chrome can do
- ✅ **Real-time events** - Console, network, performance
- ✅ **Lightweight** - No browser automation framework
- ✅ **Native** - Chrome's own protocol

**Playwright/Selenium:**
- ❌ Extra layer (driver)
- ❌ Limited to automation APIs
- ❌ Polling for events
- ❌ Heavy framework
- ❌ Abstraction overhead

## Configuration

### MCP Config

```json
{
  "mcpServers": {
    "terminal-browser": {
      "command": "python",
      "args": ["/path/to/backend/mcp_servers/terminal_browser/server.py"],
      "env": {
        "OLLAMA_BASE_URL": "http://localhost:11434"
      }
    }
  }
}
```

### Chrome Launch Script

```bash
#!/bin/bash
# launch_chrome_debug.sh

google-chrome \
  --remote-debugging-port=9222 \
  --user-data-dir=/tmp/chrome-debug \
  --no-first-run \
  --no-default-browser-check \
  --disable-background-networking \
  --disable-sync
```

## Troubleshooting

### Chrome not connecting

```bash
# Check if Chrome is running with debugging
curl http://localhost:9222/json

# Should return list of targets
```

### Models not found

```bash
# Pull terminal model
ollama pull comanderanch/Linux-Buster:latest

# Pull vision model
ollama pull gemma3:27b

# Verify
ollama list
```

### WebSocket errors

```bash
# Ensure no firewall blocking port 9222
sudo ufw allow 9222

# Check Chrome is listening
netstat -an | grep 9222
```

## Summary

**11 tools total:**
- 3 terminal tools (natural language, execute, explain)
- 8 browser tools (navigate, screenshot, DOM, click, type, script, content, console, network)

**Key Features:**
- ✅ Chrome DevTools Protocol (direct browser access)
- ✅ Gemma 3 vision (screenshot analysis)
- ✅ Linux-Buster model (250+ commands)
- ✅ Safety validation (dangerous command blocking)
- ✅ Unified server (terminal + browser together)

**Perfect for:**
- Web scraping with understanding
- Automated testing with vision
- DevOps tasks (terminal + browser)
- Research workflows (search + analyze)
- Self-improving agents (modify own code + test in browser)


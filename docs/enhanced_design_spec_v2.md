# Enhanced UI/UX Design Specification - Midnight Glassmorphism with Interactive Takeover & Agent Control

## Design Philosophy Enhancement

This enhanced specification builds upon the original midnight glassmorphism theme by adding **interactive browser and terminal takeover capabilities** similar to Manus computer interface. The design maintains visual consistency while enabling users to seamlessly transition between AI-driven automation and manual control. Additionally, it includes **WebSocket-based communication with a Go Fiber v3 backend** and **full API control for Gemma 3-powered agent**.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Frontend (React/Vue)                    │
│                  Midnight Glassmorphism UI                  │
│  ┌──────────────┬──────────────┬──────────────────────────┐ │
│  │ Left Panel   │ Center Chat  │ Right Panel (MCP/Agent)  │ │
│  │ File Tree    │ Messages     │ MCP Integration          │ │
│  │              │              │ Agent Status             │ │
│  └──────────────┴──────────────┴──────────────────────────┘ │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │ Bottom Dock: Terminal | Browser | MCP Tools | Logs      │ │
│  └──────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                            ↕ WebSocket (wss://)
┌─────────────────────────────────────────────────────────────┐
│              Go Fiber v3 Backend Server                     │
│  ┌──────────────┬──────────────┬──────────────────────────┐ │
│  │ WebSocket    │ REST API     │ Agent Controller         │ │
│  │ Handler      │ Endpoints    │ (Gemma 3)                │ │
│  └──────────────┴──────────────┴──────────────────────────┘ │
│  ┌──────────────┬──────────────┬──────────────────────────┐ │
│  │ Browser      │ Terminal     │ MCP Client               │ │
│  │ Automation   │ PTY Manager  │ Integration              │ │
│  │ (Playwright) │              │                          │ │
│  └──────────────┴──────────────┴──────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                   External Services                         │
│  • MCP Servers (configured tools)                          │
│  • Gemma 3 Model (agent intelligence)                      │
│  • File System / Knowledge Base                            │
└─────────────────────────────────────────────────────────────┘
```

---

## Key Enhancements for Consistency

### 1. **Interactive Browser Panel with Takeover Mode**

The browser component should mirror Manus's approach of providing both automated and manual control within the same glassmorphic interface.

#### Browser Panel Structure

```
┌─────────────────────────────────────────────────────────┐
│  Browser Controls [Glass Header Bar]                    │
│  ┌─────┬─────┬─────┬──────────────────┬──────────────┐ │
│  │ ←   │ →   │ ⟳   │ https://...      │ [Takeover]   │ │
│  └─────┴─────┴─────┴──────────────────┴──────────────┘ │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  [Browser Viewport - Glass Container]                  │
│                                                         │
│  • Numbered overlay boxes (Rango-style) in cyan glow   │
│  • Screenshot-based visual feedback                    │
│  • AI command parser status indicator                  │
│  • Takeover mode toggle with smooth transition         │
│  • WebSocket connection status indicator               │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

#### Browser Takeover Features

**Visual States:**

1. **AI-Driven Mode** (Default)
   - Numbered cyan boxes overlay interactive elements
   - AI command parser active (displays: "CLICK 5", "TYPE 3 'text'")
   - Subtle cyan glow border indicating automation
   - Screenshot capture indicator (pulsing cyan dot)
   - WebSocket status: "Connected - AI Control"

2. **Takeover Mode** (User Control)
   - Numbered overlays fade out
   - Direct mouse/keyboard input enabled
   - Border changes to white glow indicating manual control
   - "User Control Active" badge in top-right corner
   - WebSocket status: "Connected - Manual Override"

**Takeover Button Styling:**

```css
.browser-takeover-button {
  background: rgba(255, 255, 255, 0.25);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(74, 210, 255, 0.5);
  border-radius: 8px;
  padding: 8px 16px;
  color: #FFFFFF;
  font-family: var(--font-primary);
  font-weight: var(--font-weight-medium);
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.browser-takeover-button:hover {
  background: rgba(255, 255, 255, 0.35);
  border-color: #15A7FF;
  box-shadow: 0 0 20px rgba(21, 167, 255, 0.4);
}

.browser-takeover-button.active {
  background: rgba(21, 167, 255, 0.3);
  border-color: #FFFFFF;
  box-shadow: 0 0 25px rgba(255, 255, 255, 0.5);
}

.browser-takeover-button::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.3),
    transparent
  );
  transition: left 0.5s ease;
}

.browser-takeover-button:hover::before {
  left: 100%;
}
```

**Numbered Overlay System (Rango-style):**

```css
.browser-element-number {
  position: absolute;
  background: rgba(21, 167, 255, 0.9);
  backdrop-filter: blur(10px);
  border: 2px solid #FFFFFF;
  border-radius: 6px;
  padding: 4px 8px;
  color: #FFFFFF;
  font-family: var(--font-primary);
  font-weight: var(--font-weight-bold);
  font-size: 0.75rem;
  box-shadow: 
    0 4px 12px rgba(21, 167, 255, 0.6),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  z-index: 9999;
  pointer-events: none;
  animation: numberPulse 2s ease-in-out infinite;
}

@keyframes numberPulse {
  0%, 100% { 
    transform: scale(1);
    box-shadow: 0 4px 12px rgba(21, 167, 255, 0.6);
  }
  50% { 
    transform: scale(1.05);
    box-shadow: 0 6px 18px rgba(21, 167, 255, 0.8);
  }
}

.browser-element-number.highlighted {
  background: rgba(255, 42, 109, 0.9);
  border-color: #FF6B9D;
  animation: highlightPulse 0.5s ease-in-out;
}

@keyframes highlightPulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.2); }
}
```

---

### 2. **Enhanced Terminal Panel with Takeover Mode**

The terminal should provide seamless switching between AI-controlled command execution and manual user input.

#### Terminal Panel Structure

```
┌─────────────────────────────────────────────────────────┐
│  Terminal [Glass Header Bar]                            │
│  ┌────────────┬────────────┬────────────┬─────────────┐ │
│  │ Terminal   │ Browser    │ MCP Tools  │ [Takeover]  │ │
│  └────────────┴────────────┴────────────┴─────────────┘ │
├─────────────────────────────────────────────────────────┤
│  ubuntu@agent:~$ █                                      │
│                                                         │
│  [Command History with Glass Bubbles]                  │
│  • AI commands in cyan-bordered bubbles                │
│  • User commands in white-bordered bubbles             │
│  • Output in translucent glass containers              │
│  • WebSocket connection status indicator               │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

#### Terminal Takeover Features

**Visual States:**

1. **AI-Driven Mode** (Default)
   - Commands appear with AI attribution badge
   - Cyan prompt color: `#15A7FF`
   - Auto-scroll enabled
   - "AI Executing" indicator in header
   - WebSocket status: "Connected - AI Control"

2. **Takeover Mode** (User Control)
   - Prompt changes to white: `#FFFFFF`
   - Direct keyboard input enabled
   - User attribution badge appears
   - "Manual Control" indicator in header
   - WebSocket status: "Connected - Manual Override"

**Terminal Command Bubble Styling:**

```css
.terminal-command-bubble {
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(15px);
  border-left: 3px solid #15A7FF;
  border-radius: 8px;
  padding: 12px 16px;
  margin: 8px 0;
  font-family: 'Monaco', 'Consolas', 'Courier New', monospace;
  font-size: 0.9rem;
  line-height: 1.4;
  position: relative;
  transition: all 0.3s ease;
}

.terminal-command-bubble.ai-command {
  border-left-color: #15A7FF;
  box-shadow: -4px 0 12px rgba(21, 167, 255, 0.3);
}

.terminal-command-bubble.user-command {
  border-left-color: #FFFFFF;
  box-shadow: -4px 0 12px rgba(255, 255, 255, 0.3);
}

.terminal-command-bubble::before {
  content: attr(data-attribution);
  position: absolute;
  top: -8px;
  left: 12px;
  background: rgba(21, 167, 255, 0.9);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: var(--font-weight-medium);
  color: #FFFFFF;
}

.terminal-command-bubble.user-command::before {
  background: rgba(255, 255, 255, 0.9);
  color: #050910;
}

.terminal-output {
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  padding: 10px 14px;
  margin: 4px 0 8px 20px;
  color: #8D9AA8;
  font-family: 'Monaco', 'Consolas', 'Courier New', monospace;
  font-size: 0.85rem;
  white-space: pre-wrap;
  word-break: break-all;
}
```

---

### 3. **MCP Integration Panel** (Replaces Neo4j)

The right panel now features MCP (Model Context Protocol) integration for accessing external tools and services.

#### MCP Panel Structure

```
┌─────────────────────────────────────────────────────────┐
│  MCP Integration [Glass Header Bar]                     │
│  ┌──────────────────────────────────────────────────────┐ │
│  │ 🔌 Connected Servers (3)                            │ │
│  │ ┌────────────────────────────────────────────────┐  │ │
│  │ │ ✓ playwright - Browser automation              │  │ │
│  │ │ ✓ filesystem - File operations                 │  │ │
│  │ │ ✓ github - Repository management               │  │ │
│  │ └────────────────────────────────────────────────┘  │ │
│  └──────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────┤
│  Available Tools [Glass Cards]                          │
│  ┌────────────┬────────────┬────────────┐              │
│  │ tool_list  │ tool_call  │ resource   │              │
│  │ [Cyan]     │ [Cyan]     │ [Cyan]     │              │
│  └────────────┴────────────┴────────────┘              │
│                                                         │
│  Recent Activity [Glass Timeline]                      │
│  • playwright.navigate → https://example.com           │
│  • filesystem.read_file → /path/to/file.txt            │
│  • github.create_issue → Success                       │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

#### MCP Panel Styling

```css
.mcp-panel {
  background: rgba(255, 255, 255, 0.25);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  padding: 16px;
  height: 100%;
  overflow-y: auto;
}

.mcp-server-card {
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(15px);
  border: 1px solid rgba(21, 167, 255, 0.3);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
  transition: all 0.3s ease;
}

.mcp-server-card:hover {
  background: rgba(255, 255, 255, 0.2);
  border-color: #15A7FF;
  box-shadow: 0 4px 16px rgba(21, 167, 255, 0.4);
}

.mcp-server-card.connected::before {
  content: '✓';
  position: absolute;
  top: 8px;
  right: 8px;
  width: 20px;
  height: 20px;
  background: rgba(21, 167, 255, 0.9);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #FFFFFF;
  font-size: 0.7rem;
  font-weight: var(--font-weight-bold);
  box-shadow: 0 0 10px rgba(21, 167, 255, 0.6);
}

.mcp-tool-button {
  background: rgba(21, 167, 255, 0.2);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(21, 167, 255, 0.5);
  border-radius: 6px;
  padding: 8px 12px;
  color: #FFFFFF;
  font-family: var(--font-primary);
  font-weight: var(--font-weight-medium);
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.3s ease;
  display: inline-block;
  margin: 4px;
}

.mcp-tool-button:hover {
  background: rgba(21, 167, 255, 0.3);
  border-color: #15A7FF;
  box-shadow: 0 0 15px rgba(21, 167, 255, 0.5);
}

.mcp-activity-timeline {
  margin-top: 16px;
}

.mcp-activity-item {
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(10px);
  border-left: 3px solid #15A7FF;
  border-radius: 6px;
  padding: 10px 12px;
  margin-bottom: 8px;
  color: #8D9AA8;
  font-family: 'Monaco', 'Consolas', 'Courier New', monospace;
  font-size: 0.85rem;
  position: relative;
}

.mcp-activity-item::before {
  content: '';
  position: absolute;
  left: -6px;
  top: 50%;
  transform: translateY(-50%);
  width: 8px;
  height: 8px;
  background: #15A7FF;
  border-radius: 50%;
  box-shadow: 0 0 8px rgba(21, 167, 255, 0.6);
}
```

---

### 4. **WebSocket Communication Architecture**

Implement real-time bidirectional communication between frontend and Go Fiber v3 backend.

#### Frontend WebSocket Client

```javascript
class AgentWebSocket {
  constructor(url = 'wss://localhost:8080/ws') {
    this.url = url;
    this.ws = null;
    this.reconnectInterval = 3000;
    this.messageHandlers = new Map();
    this.connectionStatus = 'disconnected';
    this.listeners = [];
  }
  
  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.connectionStatus = 'connected';
      this.notifyStatusChange('connected');
      this.updateConnectionIndicator('connected');
    };
    
    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };
    
    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.connectionStatus = 'error';
      this.notifyStatusChange('error');
      this.updateConnectionIndicator('error');
    };
    
    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.connectionStatus = 'disconnected';
      this.notifyStatusChange('disconnected');
      this.updateConnectionIndicator('disconnected');
      
      // Auto-reconnect
      setTimeout(() => this.connect(), this.reconnectInterval);
    };
  }
  
  send(type, payload) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message = {
        type,
        payload,
        timestamp: Date.now()
      };
      this.ws.send(JSON.stringify(message));
    } else {
      console.error('WebSocket not connected');
    }
  }
  
  handleMessage(message) {
    const { type, payload } = message;
    
    // Route message to appropriate handler
    if (this.messageHandlers.has(type)) {
      this.messageHandlers.get(type)(payload);
    }
    
    // Handle specific message types
    switch (type) {
      case 'browser_action':
        this.handleBrowserAction(payload);
        break;
      case 'terminal_command':
        this.handleTerminalCommand(payload);
        break;
      case 'terminal_output':
        this.handleTerminalOutput(payload);
        break;
      case 'mcp_tool_call':
        this.handleMCPToolCall(payload);
        break;
      case 'mcp_tool_result':
        this.handleMCPToolResult(payload);
        break;
      case 'agent_status':
        this.handleAgentStatus(payload);
        break;
      case 'screenshot':
        this.handleScreenshot(payload);
        break;
      case 'numbered_elements':
        this.handleNumberedElements(payload);
        break;
      default:
        console.log('Unhandled message type:', type);
    }
  }
  
  // Register custom message handler
  on(type, handler) {
    this.messageHandlers.set(type, handler);
  }
  
  // Browser control methods
  browserNavigate(url) {
    this.send('browser_navigate', { url });
  }
  
  browserClick(elementNumber) {
    this.send('browser_click', { element: elementNumber });
  }
  
  browserType(elementNumber, text) {
    this.send('browser_type', { element: elementNumber, text });
  }
  
  browserScreenshot() {
    this.send('browser_screenshot', {});
  }
  
  // Terminal control methods
  terminalExecute(command) {
    this.send('terminal_execute', { command });
  }
  
  terminalInput(input) {
    this.send('terminal_input', { input });
  }
  
  // MCP control methods
  mcpListTools(server) {
    this.send('mcp_list_tools', { server });
  }
  
  mcpCallTool(server, tool, args) {
    this.send('mcp_call_tool', { server, tool, args });
  }
  
  // Takeover control methods
  toggleBrowserTakeover() {
    this.send('toggle_browser_takeover', {});
  }
  
  toggleTerminalTakeover() {
    this.send('toggle_terminal_takeover', {});
  }
  
  // Message handlers
  handleBrowserAction(payload) {
    const { action, target, value } = payload;
    console.log(`Browser action: ${action} on ${target}`);
    
    // Update UI to show AI action
    this.showAICommandParser(`${action.toUpperCase()} ${target}`, value);
  }
  
  handleTerminalCommand(payload) {
    const { command, attribution } = payload;
    
    // Add command bubble to terminal
    const bubble = document.createElement('div');
    bubble.className = `terminal-command-bubble ${attribution === 'ai' ? 'ai-command' : 'user-command'}`;
    bubble.setAttribute('data-attribution', attribution === 'ai' ? 'AI' : 'USER');
    bubble.textContent = command;
    
    document.querySelector('.terminal-content').appendChild(bubble);
  }
  
  handleTerminalOutput(payload) {
    const { output } = payload;
    
    // Add output to terminal
    const outputDiv = document.createElement('div');
    outputDiv.className = 'terminal-output';
    outputDiv.textContent = output;
    
    document.querySelector('.terminal-content').appendChild(outputDiv);
  }
  
  handleMCPToolCall(payload) {
    const { server, tool, args } = payload;
    
    // Add to MCP activity timeline
    const item = document.createElement('div');
    item.className = 'mcp-activity-item';
    item.textContent = `${server}.${tool} → ${JSON.stringify(args)}`;
    
    document.querySelector('.mcp-activity-timeline').prepend(item);
  }
  
  handleMCPToolResult(payload) {
    const { result, success } = payload;
    
    // Update last activity item with result
    const lastItem = document.querySelector('.mcp-activity-item');
    if (lastItem) {
      lastItem.textContent += ` → ${success ? 'Success' : 'Failed'}`;
      lastItem.style.borderLeftColor = success ? '#15A7FF' : '#FF2A6D';
    }
  }
  
  handleAgentStatus(payload) {
    const { status, message } = payload;
    
    // Update agent status indicator
    const indicator = document.querySelector('.agent-status-indicator');
    if (indicator) {
      indicator.textContent = message;
      indicator.className = `agent-status-indicator ${status}`;
    }
  }
  
  handleScreenshot(payload) {
    const { image, timestamp } = payload;
    
    // Show screenshot capture indicator
    const indicator = document.querySelector('.screenshot-indicator');
    if (indicator) {
      indicator.classList.add('active');
      setTimeout(() => indicator.classList.remove('active'), 2000);
    }
    
    // Optionally add to screenshot gallery
    this.addScreenshotToGallery(image, timestamp);
  }
  
  handleNumberedElements(payload) {
    const { elements } = payload;
    
    // Clear existing numbered overlays
    document.querySelectorAll('.browser-element-number').forEach(el => el.remove());
    
    // Add new numbered overlays
    elements.forEach((el, index) => {
      const overlay = document.createElement('div');
      overlay.className = 'browser-element-number';
      overlay.textContent = index + 1;
      overlay.style.left = `${el.x}px`;
      overlay.style.top = `${el.y}px`;
      
      document.querySelector('.browser-viewport').appendChild(overlay);
    });
  }
  
  showAICommandParser(command, description) {
    const parser = document.querySelector('.ai-command-parser');
    if (parser) {
      parser.querySelector('.command-text').textContent = command;
      parser.querySelector('.command-description').textContent = description || '';
      parser.classList.add('active');
      
      setTimeout(() => parser.classList.remove('active'), 3000);
    }
  }
  
  addScreenshotToGallery(imageData, timestamp) {
    const gallery = document.querySelector('.screenshot-gallery');
    if (gallery) {
      const img = document.createElement('img');
      img.className = 'screenshot-thumbnail';
      img.src = `data:image/png;base64,${imageData}`;
      img.alt = `Screenshot ${timestamp}`;
      img.onclick = () => this.showScreenshotModal(imageData);
      
      gallery.prepend(img);
      
      // Keep only last 10 screenshots
      const thumbnails = gallery.querySelectorAll('.screenshot-thumbnail');
      if (thumbnails.length > 10) {
        thumbnails[thumbnails.length - 1].remove();
      }
    }
  }
  
  updateConnectionIndicator(status) {
    const indicator = document.querySelector('.ws-connection-indicator');
    if (indicator) {
      indicator.className = `ws-connection-indicator ${status}`;
      indicator.querySelector('.status-text').textContent = 
        status === 'connected' ? 'Connected' : 
        status === 'error' ? 'Connection Error' : 
        'Disconnected';
    }
  }
  
  notifyStatusChange(status) {
    this.listeners.forEach(listener => listener(status));
  }
  
  addStatusListener(callback) {
    this.listeners.push(callback);
  }
}

// Initialize WebSocket connection
const agentWS = new AgentWebSocket('wss://localhost:8080/ws');
agentWS.connect();
```

#### WebSocket Connection Indicator

```css
.ws-connection-indicator {
  position: fixed;
  top: 16px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(15px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 8px;
  padding: 8px 16px;
  z-index: 1000;
  transition: all 0.3s ease;
}

.ws-connection-indicator .status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #8D9AA8;
  box-shadow: 0 0 10px rgba(141, 154, 168, 0.6);
  animation: statusPulse 2s ease-in-out infinite;
}

.ws-connection-indicator.connected .status-dot {
  background: #15A7FF;
  box-shadow: 0 0 10px rgba(21, 167, 255, 0.8);
}

.ws-connection-indicator.error .status-dot {
  background: #FF2A6D;
  box-shadow: 0 0 10px rgba(255, 42, 109, 0.8);
  animation: errorPulse 1s ease-in-out infinite;
}

@keyframes errorPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.ws-connection-indicator .status-text {
  color: #FFFFFF;
  font-family: var(--font-primary);
  font-weight: var(--font-weight-medium);
  font-size: 0.85rem;
}
```

---

### 5. **Go Fiber v3 Backend Implementation**

Backend server handling WebSocket connections, browser automation, terminal management, and MCP integration.

#### Main Server Structure (Go)

```go
package main

import (
    "log"
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/websocket/v3"
)

type Message struct {
    Type      string                 `json:"type"`
    Payload   map[string]interface{} `json:"payload"`
    Timestamp int64                  `json:"timestamp"`
}

type AgentServer struct {
    app              *fiber.App
    browserManager   *BrowserManager
    terminalManager  *TerminalManager
    mcpClient        *MCPClient
    agentController  *AgentController
    clients          map[*websocket.Conn]bool
}

func NewAgentServer() *AgentServer {
    return &AgentServer{
        app:             fiber.New(),
        browserManager:  NewBrowserManager(),
        terminalManager: NewTerminalManager(),
        mcpClient:       NewMCPClient(),
        agentController: NewAgentController(),
        clients:         make(map[*websocket.Conn]bool),
    }
}

func (s *AgentServer) Setup() {
    // WebSocket endpoint
    s.app.Get("/ws", websocket.New(s.handleWebSocket))
    
    // REST API endpoints for agent control
    s.app.Post("/api/agent/execute", s.handleAgentExecute)
    s.app.Post("/api/browser/navigate", s.handleBrowserNavigate)
    s.app.Post("/api/terminal/execute", s.handleTerminalExecute)
    s.app.Post("/api/mcp/call", s.handleMCPCall)
    s.app.Get("/api/status", s.handleStatus)
    
    // Static files (optional)
    s.app.Static("/", "./public")
}

func (s *AgentServer) handleWebSocket(c *websocket.Conn) {
    // Register client
    s.clients[c] = true
    defer func() {
        delete(s.clients, c)
        c.Close()
    }()
    
    log.Println("WebSocket client connected")
    
    // Send initial status
    s.sendToClient(c, Message{
        Type: "agent_status",
        Payload: map[string]interface{}{
            "status":  "ready",
            "message": "Agent ready for commands",
        },
    })
    
    // Message loop
    for {
        var msg Message
        if err := c.ReadJSON(&msg); err != nil {
            log.Println("WebSocket read error:", err)
            break
        }
        
        s.handleMessage(c, msg)
    }
}

func (s *AgentServer) handleMessage(c *websocket.Conn, msg Message) {
    switch msg.Type {
    case "browser_navigate":
        url := msg.Payload["url"].(string)
        s.browserManager.Navigate(url)
        s.captureAndSendScreenshot(c)
        
    case "browser_click":
        element := int(msg.Payload["element"].(float64))
        s.browserManager.Click(element)
        s.captureAndSendScreenshot(c)
        
    case "browser_type":
        element := int(msg.Payload["element"].(float64))
        text := msg.Payload["text"].(string)
        s.browserManager.Type(element, text)
        s.captureAndSendScreenshot(c)
        
    case "browser_screenshot":
        s.captureAndSendScreenshot(c)
        
    case "terminal_execute":
        command := msg.Payload["command"].(string)
        output := s.terminalManager.Execute(command)
        s.sendToClient(c, Message{
            Type: "terminal_output",
            Payload: map[string]interface{}{
                "output": output,
            },
        })
        
    case "terminal_input":
        input := msg.Payload["input"].(string)
        s.terminalManager.SendInput(input)
        
    case "mcp_list_tools":
        server := msg.Payload["server"].(string)
        tools := s.mcpClient.ListTools(server)
        s.sendToClient(c, Message{
            Type: "mcp_tools_list",
            Payload: map[string]interface{}{
                "server": server,
                "tools":  tools,
            },
        })
        
    case "mcp_call_tool":
        server := msg.Payload["server"].(string)
        tool := msg.Payload["tool"].(string)
        args := msg.Payload["args"]
        
        result := s.mcpClient.CallTool(server, tool, args)
        s.sendToClient(c, Message{
            Type: "mcp_tool_result",
            Payload: map[string]interface{}{
                "result":  result,
                "success": result != nil,
            },
        })
        
    case "toggle_browser_takeover":
        s.browserManager.ToggleTakeover()
        
    case "toggle_terminal_takeover":
        s.terminalManager.ToggleTakeover()
        
    default:
        log.Printf("Unknown message type: %s", msg.Type)
    }
}

func (s *AgentServer) captureAndSendScreenshot(c *websocket.Conn) {
    screenshot := s.browserManager.CaptureScreenshot()
    elements := s.browserManager.DetectElements()
    
    s.sendToClient(c, Message{
        Type: "screenshot",
        Payload: map[string]interface{}{
            "image":     screenshot,
            "timestamp": time.Now().Unix(),
        },
    })
    
    s.sendToClient(c, Message{
        Type: "numbered_elements",
        Payload: map[string]interface{}{
            "elements": elements,
        },
    })
}

func (s *AgentServer) sendToClient(c *websocket.Conn, msg Message) {
    msg.Timestamp = time.Now().UnixMilli()
    if err := c.WriteJSON(msg); err != nil {
        log.Println("WebSocket write error:", err)
    }
}

func (s *AgentServer) broadcast(msg Message) {
    msg.Timestamp = time.Now().UnixMilli()
    for client := range s.clients {
        if err := client.WriteJSON(msg); err != nil {
            log.Println("Broadcast error:", err)
            client.Close()
            delete(s.clients, client)
        }
    }
}

// REST API handlers for agent control
func (s *AgentServer) handleAgentExecute(c *fiber.Ctx) error {
    var req struct {
        Task string `json:"task"`
    }
    
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    // Execute task with Gemma 3 agent
    result := s.agentController.ExecuteTask(req.Task)
    
    return c.JSON(fiber.Map{
        "success": true,
        "result":  result,
    })
}

func (s *AgentServer) handleBrowserNavigate(c *fiber.Ctx) error {
    var req struct {
        URL string `json:"url"`
    }
    
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    s.browserManager.Navigate(req.URL)
    
    // Broadcast to all connected clients
    s.broadcast(Message{
        Type: "browser_action",
        Payload: map[string]interface{}{
            "action": "navigate",
            "target": req.URL,
        },
    })
    
    return c.JSON(fiber.Map{"success": true})
}

func (s *AgentServer) handleTerminalExecute(c *fiber.Ctx) error {
    var req struct {
        Command string `json:"command"`
    }
    
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    output := s.terminalManager.Execute(req.Command)
    
    // Broadcast to all connected clients
    s.broadcast(Message{
        Type: "terminal_command",
        Payload: map[string]interface{}{
            "command":     req.Command,
            "attribution": "api",
        },
    })
    
    s.broadcast(Message{
        Type: "terminal_output",
        Payload: map[string]interface{}{
            "output": output,
        },
    })
    
    return c.JSON(fiber.Map{
        "success": true,
        "output":  output,
    })
}

func (s *AgentServer) handleMCPCall(c *fiber.Ctx) error {
    var req struct {
        Server string                 `json:"server"`
        Tool   string                 `json:"tool"`
        Args   map[string]interface{} `json:"args"`
    }
    
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    result := s.mcpClient.CallTool(req.Server, req.Tool, req.Args)
    
    // Broadcast to all connected clients
    s.broadcast(Message{
        Type: "mcp_tool_call",
        Payload: map[string]interface{}{
            "server": req.Server,
            "tool":   req.Tool,
            "args":   req.Args,
        },
    })
    
    s.broadcast(Message{
        Type: "mcp_tool_result",
        Payload: map[string]interface{}{
            "result":  result,
            "success": result != nil,
        },
    })
    
    return c.JSON(fiber.Map{
        "success": true,
        "result":  result,
    })
}

func (s *AgentServer) handleStatus(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "browser_ready":  s.browserManager.IsReady(),
        "terminal_ready": s.terminalManager.IsReady(),
        "mcp_connected":  s.mcpClient.IsConnected(),
        "agent_status":   s.agentController.GetStatus(),
    })
}

func (s *AgentServer) Start(port string) {
    log.Printf("Starting Agent Server on port %s", port)
    log.Fatal(s.app.Listen(":" + port))
}

func main() {
    server := NewAgentServer()
    server.Setup()
    server.Start("8080")
}
```

---

### 6. **Gemma 3 Agent Controller**

Integration with Gemma 3 model for intelligent agent control.

#### Agent Controller (Go)

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "log"
)

type AgentController struct {
    modelEndpoint string
    apiKey        string
    conversationHistory []Message
}

func NewAgentController() *AgentController {
    return &AgentController{
        modelEndpoint: "http://localhost:11434/api/generate", // Ollama endpoint
        conversationHistory: make([]Message, 0),
    }
}

type AgentRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
    Stream bool   `json:"stream"`
}

type AgentResponse struct {
    Response string `json:"response"`
    Done     bool   `json:"done"`
}

func (ac *AgentController) ExecuteTask(task string) string {
    // Build context from conversation history
    context := ac.buildContext()
    
    // Create prompt for Gemma 3
    prompt := ac.buildPrompt(task, context)
    
    // Send request to Gemma 3
    reqBody := AgentRequest{
        Model:  "gemma2:3b",
        Prompt: prompt,
        Stream: false,
    }
    
    jsonData, _ := json.Marshal(reqBody)
    resp, err := http.Post(ac.modelEndpoint, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Println("Error calling Gemma 3:", err)
        return "Error executing task"
    }
    defer resp.Body.Close()
    
    var agentResp AgentResponse
    json.NewDecoder(resp.Body).Decode(&agentResp)
    
    // Parse agent response and extract actions
    actions := ac.parseActions(agentResp.Response)
    
    // Store in conversation history
    ac.conversationHistory = append(ac.conversationHistory, Message{
        Type: "agent_task",
        Payload: map[string]interface{}{
            "task":     task,
            "response": agentResp.Response,
            "actions":  actions,
        },
    })
    
    return agentResp.Response
}

func (ac *AgentController) buildContext() string {
    // Build context from recent conversation history
    context := "Previous interactions:\n"
    for _, msg := range ac.conversationHistory {
        context += fmt.Sprintf("- %s\n", msg.Type)
    }
    return context
}

func (ac *AgentController) buildPrompt(task string, context string) string {
    return fmt.Sprintf(`You are an AI agent controlling a web browser and terminal through a GUI interface.

Available actions:
- BROWSER_NAVIGATE <url>
- BROWSER_CLICK <element_number>
- BROWSER_TYPE <element_number> <text>
- TERMINAL_EXECUTE <command>
- MCP_CALL <server> <tool> <args>

%s

User task: %s

Respond with a plan and the specific actions to take in the format above.`, context, task)
}

func (ac *AgentController) parseActions(response string) []map[string]interface{} {
    // Parse agent response to extract structured actions
    // This is a simplified version - implement proper parsing logic
    actions := make([]map[string]interface{}, 0)
    
    // Example parsing logic
    lines := strings.Split(response, "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "BROWSER_NAVIGATE") {
            parts := strings.SplitN(line, " ", 2)
            if len(parts) == 2 {
                actions = append(actions, map[string]interface{}{
                    "type": "browser_navigate",
                    "url":  parts[1],
                })
            }
        } else if strings.HasPrefix(line, "BROWSER_CLICK") {
            parts := strings.SplitN(line, " ", 2)
            if len(parts) == 2 {
                element, _ := strconv.Atoi(parts[1])
                actions = append(actions, map[string]interface{}{
                    "type":    "browser_click",
                    "element": element,
                })
            }
        }
        // Add more action parsers...
    }
    
    return actions
}

func (ac *AgentController) GetStatus() string {
    return "ready"
}
```

---

### 7. **Updated Layout Integration**

Update the original three-panel layout to incorporate WebSocket status, MCP integration, and agent control.

#### Updated Panel Structure

```
┌─────────────────┬─────────────────────┬─────────────────┐
│                 │                     │                 │
│   LEFT PANEL    │    CENTER CHAT      │   RIGHT PANEL   │
│    (25%)        │      (50%)          │     (25%)       │
│  [Glass Card]   │   [Deep Perspective]│   [Glass Card]  │
│                 │                     │                 │
│  • File Tree    │   • Message Bubbles │   • MCP Servers │
│  • Knowledge    │   • Status Header   │   • Tools       │
│                 │   • Input Bar       │   • Activity    │
│                 │   • WS Status       │   • Agent Status│
│                 │                     │                 │
├─────────────────┴─────────────────────┴─────────────────┤
│                                                         │
│         BOTTOM DOCK (30%) - Tabbed Interface           │
│  ┌────────────┬────────────┬────────────┬────────────┐ │
│  │ Terminal   │ Browser    │ MCP Tools  │ Logs       │ │
│  └────────────┴────────────┴────────────┴────────────┘ │
│                                                         │
│  [Active Tab Content with Takeover Controls]           │
│  • Glass panel with 25% opacity                        │
│  • Takeover button in top-right                        │
│  • Status indicator showing AI/Manual mode             │
│  • WebSocket connection status                         │
│  • Smooth transitions between modes                    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## API Endpoints for Agent Control

### REST API Endpoints

```
POST /api/agent/execute
Body: { "task": "Navigate to example.com and click login" }
Response: { "success": true, "result": "Task completed" }

POST /api/browser/navigate
Body: { "url": "https://example.com" }
Response: { "success": true }

POST /api/browser/click
Body: { "element": 5 }
Response: { "success": true }

POST /api/browser/type
Body: { "element": 3, "text": "username" }
Response: { "success": true }

POST /api/terminal/execute
Body: { "command": "ls -la" }
Response: { "success": true, "output": "..." }

POST /api/mcp/call
Body: { "server": "playwright", "tool": "navigate", "args": {...} }
Response: { "success": true, "result": {...} }

GET /api/status
Response: {
  "browser_ready": true,
  "terminal_ready": true,
  "mcp_connected": true,
  "agent_status": "ready"
}
```

### WebSocket Message Types

**Client → Server:**
- `browser_navigate` - Navigate to URL
- `browser_click` - Click element
- `browser_type` - Type text
- `browser_screenshot` - Capture screenshot
- `terminal_execute` - Execute command
- `terminal_input` - Send input
- `mcp_list_tools` - List MCP tools
- `mcp_call_tool` - Call MCP tool
- `toggle_browser_takeover` - Toggle browser control
- `toggle_terminal_takeover` - Toggle terminal control

**Server → Client:**
- `browser_action` - Browser action performed
- `terminal_command` - Command executed
- `terminal_output` - Command output
- `mcp_tool_call` - MCP tool called
- `mcp_tool_result` - MCP tool result
- `agent_status` - Agent status update
- `screenshot` - Screenshot data
- `numbered_elements` - Element positions

---

## Implementation Checklist

### Frontend
- [ ] WebSocket client implementation
- [ ] Connection status indicator
- [ ] MCP integration panel
- [ ] Browser takeover with numbered overlays
- [ ] Terminal takeover with command attribution
- [ ] Real-time message handling
- [ ] Screenshot gallery
- [ ] AI command parser display
- [ ] Keyboard shortcuts

### Backend (Go Fiber v3)
- [ ] WebSocket server setup
- [ ] Browser automation (Playwright)
- [ ] Terminal PTY management
- [ ] MCP client integration
- [ ] REST API endpoints
- [ ] Message routing
- [ ] Screenshot capture
- [ ] Element detection

### Agent Integration (Gemma 3)
- [ ] Agent controller setup
- [ ] Prompt engineering
- [ ] Action parsing
- [ ] Conversation history
- [ ] Task execution
- [ ] Error handling

### MCP Integration
- [ ] MCP server configuration
- [ ] Tool discovery
- [ ] Tool invocation
- [ ] Result handling
- [ ] Activity timeline

---

## Conclusion

This comprehensive design specification provides a complete architecture for a midnight glassmorphism UI with browser/terminal takeover capabilities, WebSocket-based real-time communication with a Go Fiber v3 backend, MCP integration, and full API control for Gemma 3-powered agent. The design maintains visual consistency while enabling seamless transitions between AI-driven automation and manual control.


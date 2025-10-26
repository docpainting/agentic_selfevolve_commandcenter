# Complete UI/UX Design Specification - Midnight Glassmorphism Agent Workspace

## Vision Overview

A unified agent workspace combining VS Code-style file management, modern chat interface, OpenEvolve integration, and full "Manus Computer" capabilities (browser/terminal with takeover) - all wrapped in a sophisticated midnight glassmorphism aesthetic with dual communication protocols (WebSocket + JSON-RPC 2.0 A2A).

---

## Complete Layout Structure

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Top Bar (Optional)                          │
│  Logo | Agent Status | Connection Indicators | Settings            │
└─────────────────────────────────────────────────────────────────────┘

┌──────────────┬────────────────────────────────┬─────────────────────┐
│              │                                │                     │
│  LEFT PANEL  │        CENTER CHAT AREA        │    RIGHT PANEL      │
│    (20%)     │           (60%)                │       (20%)         │
│              │                                │                     │
│ ┌──────────┐ │  ┌──────────────────────────┐ │  ┌────────────────┐ │
│ │📁 Open   │ │  │                          │ │  │ 📁 Open Folder │ │
│ │  Folder  │ │  │   [Initially Empty]      │ │  │  (OpenEvolve)  │ │
│ └──────────┘ │  │                          │ │  └────────────────┘ │
│              │  │                          │ │                     │
│ [File Tree]  │  │                          │ │  [OpenEvolve Tree] │
│ VS Code      │  │      Modern Input        │ │  Project Structure │
│ Style        │  │    ┌──────────────┐      │ │  Components        │
│              │  │    │ Type here... │      │ │  Progress          │
│ • folder/    │  │    └──────────────┘      │ │  Watchdog          │
│   • file.js  │  │         [Send]           │ │                     │
│   • file.css │  │                          │ │ • Component A      │
│ • src/       │  │  [Expands upward as      │ │   ✓ Approved       │
│   • main.go  │  │   conversation grows]    │ │ • Component B      │
│              │  │                          │ │   ⏳ Pending        │
│              │  │                          │ │ • Watchdog Alert   │
│              │  │                          │ │   ⚠️ Review         │
│              │  │                          │ │                     │
└──────────────┴────────────────────────────────┴─────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│                    BOTTOM PANEL (30vh)                              │
│                   "Manus Computer" Experience                       │
│                                                                     │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────────────┐  │
│  │ Terminal │ Browser  │ MCP Tools│ Logs     │ [Takeover] 🔄    │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────────────┘  │
│                                                                     │
│  [Active Tab Content]                                              │
│  • Terminal: AI/User command bubbles with glass styling            │
│  • Browser: Numbered overlays (Rango-style) with screenshot        │
│  • MCP Tools: Connected servers and activity timeline              │
│  • Logs: Real-time agent activity and debug info                   │
│                                                                     │
│  WebSocket Status: ● Connected | A2A Status: ● Ready               │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Detailed Panel Specifications

### 1. Left Panel - VS Code Style File Tree

#### Header with Folder Opener

```html
<div class="left-panel glass-panel">
  <div class="panel-header">
    <button class="open-folder-button">
      <svg class="folder-icon">📁</svg>
      <span>Open Folder</span>
    </button>
  </div>
  
  <div class="file-tree">
    <!-- Tree structure populated after folder selection -->
  </div>
</div>
```

#### Styling

```css
.left-panel {
  width: 20%;
  height: calc(100vh - 30vh); /* Subtract bottom panel */
  background: rgba(255, 255, 255, 0.25);
  backdrop-filter: blur(20px);
  border-right: 1px solid rgba(255, 255, 255, 0.2);
  overflow-y: auto;
  position: relative;
}

.panel-header {
  padding: 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  position: sticky;
  top: 0;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(20px);
  z-index: 10;
}

.open-folder-button {
  width: 100%;
  background: rgba(21, 167, 255, 0.2);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(21, 167, 255, 0.5);
  border-radius: 8px;
  padding: 12px 16px;
  color: #FFFFFF;
  font-family: var(--font-primary);
  font-weight: var(--font-weight-medium);
  font-size: 0.9rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: all 0.3s ease;
}

.open-folder-button:hover {
  background: rgba(21, 167, 255, 0.3);
  border-color: #15A7FF;
  box-shadow: 0 0 20px rgba(21, 167, 255, 0.4);
  transform: translateY(-2px);
}

.open-folder-button::before {
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

.open-folder-button:hover::before {
  left: 100%;
}

.folder-icon {
  font-size: 1.2rem;
  filter: drop-shadow(0 0 5px rgba(21, 167, 255, 0.6));
}

/* File Tree Styling (VS Code inspired) */
.file-tree {
  padding: 12px 8px;
  font-family: var(--font-primary);
  font-size: 0.9rem;
  color: #FFFFFF;
}

.tree-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.tree-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.tree-item.selected {
  background: rgba(21, 167, 255, 0.3);
  border-left: 3px solid #15A7FF;
}

.tree-item.folder {
  font-weight: var(--font-weight-medium);
}

.tree-item.file {
  padding-left: 28px;
}

.tree-item .icon {
  width: 16px;
  height: 16px;
  color: #15A7FF;
  flex-shrink: 0;
}

.tree-item .expand-icon {
  width: 12px;
  height: 12px;
  transition: transform 0.2s ease;
}

.tree-item.expanded .expand-icon {
  transform: rotate(90deg);
}

.tree-item .file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tree-item .file-extension {
  color: #8D9AA8;
  font-size: 0.8rem;
  margin-left: auto;
}

/* Nested items */
.tree-children {
  margin-left: 16px;
  border-left: 1px solid rgba(255, 255, 255, 0.1);
  padding-left: 4px;
}
```

---

### 2. Center Panel - Modern Chat Interface

#### Initial State (Centered Input)

```html
<div class="center-panel">
  <div class="chat-container" data-state="initial">
    <!-- Initially centered, expands upward -->
    <div class="chat-input-wrapper centered">
      <div class="welcome-message">
        <h2>How can I help you today?</h2>
        <p>Ask me anything or give me a task to complete</p>
      </div>
      
      <div class="input-area glass-input">
        <textarea 
          placeholder="Type your message here..."
          rows="3"
        ></textarea>
        <div class="input-actions">
          <button class="attach-button">📎</button>
          <button class="send-button">
            <svg class="send-icon">➤</svg>
            Send
          </button>
        </div>
      </div>
    </div>
  </div>
</div>
```

#### Expanded State (Conversation Mode)

```html
<div class="center-panel">
  <div class="chat-container" data-state="conversation">
    <!-- Messages scroll area -->
    <div class="messages-area">
      <!-- User message -->
      <div class="message-bubble user-message">
        <div class="message-content">
          <p>Navigate to example.com and extract the title</p>
        </div>
        <div class="message-meta">
          <span class="timestamp">2:30 PM</span>
        </div>
      </div>
      
      <!-- Agent message -->
      <div class="message-bubble agent-message">
        <div class="agent-badge">
          <span class="agent-icon">🤖</span>
          <span class="agent-name">Gemma 3</span>
        </div>
        <div class="message-content">
          <p>I'll navigate to example.com and extract the title for you.</p>
          
          <!-- Task status indicator -->
          <div class="task-status working">
            <span class="status-icon">⏳</span>
            <span class="status-text">Working on it...</span>
          </div>
        </div>
        <div class="message-meta">
          <span class="timestamp">2:30 PM</span>
          <span class="task-id">task-12345</span>
        </div>
      </div>
    </div>
    
    <!-- Fixed input at bottom -->
    <div class="input-area-fixed glass-input">
      <textarea 
        placeholder="Type your message here..."
        rows="2"
      ></textarea>
      <div class="input-actions">
        <button class="attach-button">📎</button>
        <button class="send-button">
          <svg class="send-icon">➤</svg>
          Send
        </button>
      </div>
    </div>
  </div>
</div>
```

#### Styling

```css
.center-panel {
  width: 60%;
  height: calc(100vh - 30vh);
  position: relative;
  display: flex;
  flex-direction: column;
}

/* Initial centered state */
.chat-container[data-state="initial"] {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 40px;
}

.chat-input-wrapper.centered {
  width: 100%;
  max-width: 700px;
  display: flex;
  flex-direction: column;
  gap: 32px;
  animation: fadeIn 0.5s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.welcome-message {
  text-align: center;
  color: #FFFFFF;
}

.welcome-message h2 {
  font-family: var(--font-primary);
  font-weight: var(--font-weight-semibold);
  font-size: 2rem;
  margin-bottom: 12px;
  background: linear-gradient(135deg, #FFFFFF 0%, #15A7FF 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.welcome-message p {
  font-family: var(--font-primary);
  font-size: 1.1rem;
  color: #8D9AA8;
}

/* Glass input styling */
.glass-input {
  background: rgba(255, 255, 255, 0.25);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 16px;
  padding: 16px;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.glass-input::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.4),
    transparent
  );
  z-index: 1;
}

.glass-input:focus-within {
  border-color: #15A7FF;
  box-shadow: 
    0 0 30px rgba(21, 167, 255, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.glass-input textarea {
  width: 100%;
  background: transparent;
  border: none;
  outline: none;
  color: #FFFFFF;
  font-family: var(--font-primary);
  font-size: 1rem;
  line-height: 1.5;
  resize: none;
  min-height: 60px;
}

.glass-input textarea::placeholder {
  color: #8D9AA8;
}

.input-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.attach-button {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  padding: 8px 12px;
  color: #FFFFFF;
  font-size: 1.2rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.attach-button:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: translateY(-2px);
}

.send-button {
  background: rgba(21, 167, 255, 0.8);
  backdrop-filter: blur(10px);
  border: 1px solid #15A7FF;
  border-radius: 8px;
  padding: 10px 20px;
  color: #FFFFFF;
  font-family: var(--font-primary);
  font-weight: var(--font-weight-medium);
  font-size: 0.95rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.send-button:hover {
  background: rgba(21, 167, 255, 1);
  box-shadow: 0 0 25px rgba(21, 167, 255, 0.5);
  transform: translateY(-2px);
}

.send-button::before {
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

.send-button:hover::before {
  left: 100%;
}

.send-icon {
  font-size: 1rem;
}

/* Conversation state */
.chat-container[data-state="conversation"] {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.messages-area {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Message bubbles */
.message-bubble {
  max-width: 75%;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(15px);
  border-radius: 12px;
  padding: 14px 18px;
  position: relative;
  animation: messageSlideIn 0.3s ease;
}

@keyframes messageSlideIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.message-bubble.user-message {
  align-self: flex-end;
  background: rgba(21, 167, 255, 0.2);
  border: 1px solid rgba(21, 167, 255, 0.4);
  border-bottom-right-radius: 4px;
}

.message-bubble.agent-message {
  align-self: flex-start;
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-bottom-left-radius: 4px;
}

.agent-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.agent-icon {
  font-size: 1.2rem;
  filter: drop-shadow(0 0 5px rgba(21, 167, 255, 0.6));
}

.agent-name {
  font-family: var(--font-primary);
  font-weight: var(--font-weight-medium);
  font-size: 0.85rem;
  color: #15A7FF;
}

.message-content {
  color: #FFFFFF;
  font-family: var(--font-primary);
  font-size: 0.95rem;
  line-height: 1.5;
}

.message-content p {
  margin: 0 0 8px 0;
}

.task-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(255, 255, 255, 0.1);
  padding: 6px 12px;
  border-radius: 6px;
  margin-top: 8px;
  font-size: 0.85rem;
}

.task-status.working {
  border-left: 3px solid #15A7FF;
}

.task-status.completed {
  border-left: 3px solid #00FF88;
}

.task-status.failed {
  border-left: 3px solid #FF2A6D;
}

.status-icon {
  animation: spin 2s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.message-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 0.75rem;
  color: #8D9AA8;
}

.task-id {
  font-family: 'Monaco', 'Consolas', 'Courier New', monospace;
  background: rgba(255, 255, 255, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
}

/* Fixed input at bottom */
.input-area-fixed {
  margin: 16px 24px;
  position: sticky;
  bottom: 0;
}
```

---

### 3. Right Panel - OpenEvolve Integration

#### Structure

```html
<div class="right-panel glass-panel">
  <div class="panel-header">
    <button class="open-folder-button">
      <svg class="folder-icon">📁</svg>
      <span>Open Folder</span>
    </button>
  </div>
  
  <div class="openevolve-content">
    <div class="section">
      <h3 class="section-title">Components</h3>
      <div class="component-list">
        <div class="component-item approved">
          <div class="component-header">
            <span class="component-icon">✓</span>
            <span class="component-name">Component A</span>
          </div>
          <div class="component-status">Approved</div>
        </div>
        
        <div class="component-item pending">
          <div class="component-header">
            <span class="component-icon">⏳</span>
            <span class="component-name">Component B</span>
          </div>
          <div class="component-status">Pending Review</div>
        </div>
      </div>
    </div>
    
    <div class="section">
      <h3 class="section-title">Progress</h3>
      <div class="progress-visualization">
        <div class="progress-bar">
          <div class="progress-fill" style="width: 65%"></div>
        </div>
        <span class="progress-text">65% Complete</span>
      </div>
    </div>
    
    <div class="section">
      <h3 class="section-title">Watchdog Alerts</h3>
      <div class="alert-list">
        <div class="alert-item warning">
          <span class="alert-icon">⚠️</span>
          <span class="alert-text">Review Required: Security Check</span>
        </div>
      </div>
    </div>
  </div>
</div>
```

#### Styling

```css
.right-panel {
  width: 20%;
  height: calc(100vh - 30vh);
  background: rgba(255, 255, 255, 0.25);
  backdrop-filter: blur(20px);
  border-left: 1px solid rgba(255, 255, 255, 0.2);
  overflow-y: auto;
}

.openevolve-content {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.section {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 10px;
  padding: 14px;
}

.section-title {
  font-family: var(--font-primary);
  font-weight: var(--font-weight-semibold);
  font-size: 0.95rem;
  color: #15A7FF;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(21, 167, 255, 0.3);
}

.component-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.component-item {
  background: rgba(255, 255, 255, 0.1);
  border-left: 3px solid #15A7FF;
  border-radius: 6px;
  padding: 10px 12px;
  transition: all 0.3s ease;
  cursor: pointer;
}

.component-item:hover {
  background: rgba(255, 255, 255, 0.15);
  transform: translateX(4px);
}

.component-item.approved {
  border-left-color: #00FF88;
}

.component-item.pending {
  border-left-color: #FFB800;
}

.component-item.warning {
  border-left-color: #FF2A6D;
}

.component-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.component-icon {
  font-size: 1rem;
}

.component-name {
  font-family: var(--font-primary);
  font-weight: var(--font-weight-medium);
  font-size: 0.9rem;
  color: #FFFFFF;
}

.component-status {
  font-size: 0.8rem;
  color: #8D9AA8;
  padding-left: 24px;
}

.progress-visualization {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.progress-bar {
  width: 100%;
  height: 8px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  overflow: hidden;
  position: relative;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #15A7FF 0%, #1AD0FF 100%);
  border-radius: 4px;
  transition: width 0.5s ease;
  position: relative;
  overflow: hidden;
}

.progress-fill::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.3),
    transparent
  );
  animation: shimmer 2s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.progress-text {
  font-family: var(--font-primary);
  font-size: 0.85rem;
  color: #FFFFFF;
  text-align: center;
}

.alert-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.alert-item {
  display: flex;
  align-items: center;
  gap: 10px;
  background: rgba(255, 42, 109, 0.15);
  border: 1px solid rgba(255, 42, 109, 0.3);
  border-radius: 6px;
  padding: 10px 12px;
  animation: alertPulse 2s ease-in-out infinite;
}

@keyframes alertPulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(255, 42, 109, 0.4);
  }
  50% {
    box-shadow: 0 0 0 4px rgba(255, 42, 109, 0);
  }
}

.alert-icon {
  font-size: 1.1rem;
}

.alert-text {
  font-family: var(--font-primary);
  font-size: 0.85rem;
  color: #FFFFFF;
  flex: 1;
}
```

---

### 4. Bottom Panel - "Manus Computer" Experience

#### Complete Implementation

```html
<div class="bottom-panel glass-panel">
  <div class="dock-tabs">
    <button class="dock-tab active" data-tab="terminal">
      <span class="tab-icon">⌨️</span>
      <span class="tab-label">Terminal</span>
    </button>
    <button class="dock-tab" data-tab="browser">
      <span class="tab-icon">🌐</span>
      <span class="tab-label">Browser</span>
    </button>
    <button class="dock-tab" data-tab="mcp">
      <span class="tab-icon">🔌</span>
      <span class="tab-label">MCP Tools</span>
    </button>
    <button class="dock-tab" data-tab="logs">
      <span class="tab-icon">📋</span>
      <span class="tab-label">Logs</span>
    </button>
    
    <div class="dock-controls">
      <button class="takeover-toggle" data-component="terminal">
        <span class="toggle-icon">🔄</span>
        <span class="toggle-label">Takeover</span>
      </button>
    </div>
  </div>
  
  <div class="dock-content">
    <!-- Terminal Tab -->
    <div class="tab-panel active" data-panel="terminal">
      <div class="terminal-container">
        <div class="terminal-header">
          <div class="terminal-status">
            <span class="status-dot ai-mode"></span>
            <span class="status-text">AI Executing</span>
          </div>
        </div>
        
        <div class="terminal-content">
          <!-- Command bubbles appear here -->
          <div class="terminal-command-bubble ai-command" data-attribution="AI">
            <span class="command-text">$ ls -la</span>
          </div>
          <div class="terminal-output">
            total 48
            drwxr-xr-x  12 user  staff   384 Oct 26 15:30 .
            drwxr-xr-x   8 user  staff   256 Oct 25 10:15 ..
          </div>
        </div>
        
        <div class="terminal-input">
          <span class="terminal-prompt">$</span>
          <input type="text" placeholder="Command..." />
        </div>
      </div>
    </div>
    
    <!-- Browser Tab -->
    <div class="tab-panel" data-panel="browser">
      <div class="browser-container">
        <div class="browser-controls">
          <button class="nav-button">←</button>
          <button class="nav-button">→</button>
          <button class="nav-button">⟳</button>
          <input type="text" class="url-bar" placeholder="https://..." />
          <button class="takeover-button">Take Over</button>
        </div>
        
        <div class="browser-viewport">
          <!-- Screenshot with numbered overlays -->
          <img src="screenshot.png" class="browser-screenshot" />
          
          <!-- Numbered element overlays (Rango-style) -->
          <div class="browser-element-number" style="left: 100px; top: 50px;">1</div>
          <div class="browser-element-number" style="left: 200px; top: 100px;">2</div>
        </div>
      </div>
    </div>
    
    <!-- MCP Tools Tab -->
    <div class="tab-panel" data-panel="mcp">
      <div class="mcp-container">
        <div class="mcp-servers">
          <h4>Connected Servers</h4>
          <div class="mcp-server-card connected">
            <span class="server-icon">✓</span>
            <span class="server-name">playwright</span>
          </div>
        </div>
        
        <div class="mcp-activity-timeline">
          <h4>Recent Activity</h4>
          <div class="mcp-activity-item">
            playwright.navigate → https://example.com
          </div>
        </div>
      </div>
    </div>
    
    <!-- Logs Tab -->
    <div class="tab-panel" data-panel="logs">
      <div class="logs-container">
        <div class="log-entry info">
          <span class="log-time">15:30:45</span>
          <span class="log-level">INFO</span>
          <span class="log-message">WebSocket connected</span>
        </div>
      </div>
    </div>
  </div>
  
  <div class="dock-footer">
    <div class="connection-status">
      <span class="ws-indicator connected">● WebSocket</span>
      <span class="a2a-indicator ready">● A2A Ready</span>
    </div>
  </div>
</div>
```

#### Bottom Panel Styling

```css
.bottom-panel {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 30vh;
  background: rgba(5, 9, 16, 0.95);
  backdrop-filter: blur(25px);
  border-top: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 -8px 32px rgba(0, 0, 0, 0.4);
  z-index: 100;
  display: flex;
  flex-direction: column;
}

.dock-tabs {
  display: flex;
  gap: 4px;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.05);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  position: relative;
}

.dock-tab {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(15px);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-bottom: none;
  border-radius: 8px 8px 0 0;
  padding: 10px 18px;
  color: #8D9AA8;
  font-family: var(--font-primary);
  font-weight: var(--font-weight-medium);
  font-size: 0.9rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.3s ease;
  position: relative;
}

.dock-tab:hover {
  background: rgba(255, 255, 255, 0.15);
  color: #FFFFFF;
}

.dock-tab.active {
  background: rgba(21, 167, 255, 0.2);
  border-color: #15A7FF;
  color: #FFFFFF;
  box-shadow: 
    0 -4px 12px rgba(21, 167, 255, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.dock-tab.active::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(
    90deg,
    transparent,
    #15A7FF 50%,
    transparent
  );
  animation: tabFlare 3s ease-in-out infinite;
}

@keyframes tabFlare {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 1; }
}

.tab-icon {
  font-size: 1.1rem;
}

.dock-controls {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.takeover-toggle {
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(15px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  padding: 8px 16px;
  color: #FFFFFF;
  font-family: var(--font-primary);
  font-weight: var(--font-weight-medium);
  font-size: 0.85rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.3s ease;
}

.takeover-toggle:hover {
  background: rgba(21, 167, 255, 0.2);
  border-color: #15A7FF;
}

.takeover-toggle.active {
  background: rgba(255, 255, 255, 0.3);
  border-color: #FFFFFF;
  box-shadow: 0 0 20px rgba(255, 255, 255, 0.4);
}

.dock-content {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.tab-panel {
  display: none;
  height: 100%;
  padding: 16px;
  overflow-y: auto;
}

.tab-panel.active {
  display: block;
}

.dock-footer {
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.05);
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.connection-status {
  display: flex;
  gap: 24px;
  font-family: var(--font-primary);
  font-size: 0.85rem;
}

.ws-indicator,
.a2a-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #8D9AA8;
}

.ws-indicator.connected,
.a2a-indicator.ready {
  color: #15A7FF;
}

.ws-indicator::before,
.a2a-indicator::before {
  content: '●';
  animation: statusPulse 2s ease-in-out infinite;
}

@keyframes statusPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
```

---

## Dual Communication System

### WebSocket for Real-Time Chat

```javascript
class ChatWebSocket {
  constructor(url = 'wss://localhost:8080/chat') {
    this.url = url;
    this.ws = null;
  }
  
  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.handleChatMessage(data);
    };
  }
  
  sendMessage(text) {
    this.ws.send(JSON.stringify({
      type: 'chat_message',
      content: text,
      timestamp: Date.now()
    }));
  }
  
  handleChatMessage(data) {
    // Add message to chat UI
    this.addMessageToChat(data);
  }
  
  addMessageToChat(data) {
    const messagesArea = document.querySelector('.messages-area');
    const bubble = document.createElement('div');
    bubble.className = `message-bubble ${data.role}-message`;
    bubble.innerHTML = `
      <div class="message-content">
        <p>${data.content}</p>
      </div>
      <div class="message-meta">
        <span class="timestamp">${new Date(data.timestamp).toLocaleTimeString()}</span>
      </div>
    `;
    messagesArea.appendChild(bubble);
    messagesArea.scrollTop = messagesArea.scrollHeight;
  }
}
```

### JSON-RPC 2.0 for A2A Communication

```javascript
class A2AClient {
  constructor(url = 'wss://localhost:8080/a2a') {
    this.url = url;
    this.ws = null;
    this.requestId = 0;
    this.pendingRequests = new Map();
  }
  
  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleJSONRPC(message);
    };
  }
  
  request(method, params) {
    return new Promise((resolve, reject) => {
      const id = `req-${++this.requestId}`;
      const request = {
        jsonrpc: '2.0',
        id,
        method,
        params
      };
      
      this.pendingRequests.set(id, { resolve, reject });
      this.ws.send(JSON.stringify(request));
    });
  }
  
  handleJSONRPC(message) {
    if (message.id && this.pendingRequests.has(message.id)) {
      const pending = this.pendingRequests.get(message.id);
      this.pendingRequests.delete(message.id);
      
      if (message.error) {
        pending.reject(message.error);
      } else {
        pending.resolve(message.result);
      }
    }
  }
}

// Initialize both connections
const chatWS = new ChatWebSocket();
const a2aClient = new A2AClient();

chatWS.connect();
a2aClient.connect();
```

---

## Complete Interaction Flow

1. **User opens folder** → Left panel populates with file tree
2. **User types message** → Center input expands to conversation mode
3. **Agent responds** → Messages appear with task status indicators
4. **Agent executes commands** → Bottom panel shows real-time terminal/browser activity
5. **User can take over** → Click takeover button to manually control browser/terminal
6. **OpenEvolve tracks progress** → Right panel shows component status and alerts
7. **WebSocket handles chat** → Real-time message streaming
8. **JSON-RPC handles A2A** → Structured agent-to-agent communication

---

## Conclusion

This complete specification provides a unified, powerful agent workspace that combines:
- **VS Code-style file management**
- **Modern chat interface** (centered → expanding)
- **OpenEvolve integration** with progress tracking
- **Full "Manus Computer" capabilities** (browser/terminal takeover)
- **Dual communication protocols** (WebSocket + JSON-RPC 2.0 A2A)
- **Midnight glassmorphism aesthetic** throughout

All wrapped in a sophisticated, production-ready design that enables seamless human-agent collaboration.


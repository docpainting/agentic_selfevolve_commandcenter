# Agent Workspace

A complete AI agent workspace with dynamic sequential thinking, browser automation, and knowledge graph memory.

## Features

- 🧠 **Dynamic Sequential Thinking** - Perceive-Reason-Act-Reflect loop with branching
- 🌐 **Browser Automation** - ChromeDP-based automation with takeover mode
- 💾 **Dual Memory System** - LightRAG + Neo4j (long-term) + ChromeDP (short-term)
- 🎨 **Midnight Glassmorphism UI** - Modern, beautiful interface
- 🔌 **MCP Integration** - Model Context Protocol for extensibility
- 🤖 **Gemma 3 Agent** - Powered by Ollama
- 📊 **OpenEvolve** - Watchdog monitoring and code evolution
- 🔄 **Real-time Communication** - WebSocket + JSON-RPC 2.0 A2A protocol

## Architecture

```
Frontend (React + Tailwind)
    ↕ WebSocket (Chat) + JSON-RPC 2.0 (A2A)
Backend (Go Fiber v3)
    ↕
├── Agent Controller (Gemma 3)
├── Memory Systems
│   ├── Long-term (LightRAG + Neo4j + ChromeM + BoltDB)
│   └── Short-term (ChromeDP context)
├── Browser Manager (ChromeDP)
├── Terminal Manager (PTY)
├── MCP Client
└── Watchdog

MCP Server: Dynamic Thinking
├── Perceive (Vision + OCR)
├── Reason (Branched thinking)
├── Act (Tool execution)
└── Reflect (Self-improvement)
```

## Prerequisites

- **Go** 1.21+ ([install](https://go.dev/doc/install))
- **Node.js** 18+ ([install](https://nodejs.org/))
- **Neo4j** 5.26 Community Edition ([install](https://neo4j.com/download/))
- **Ollama** with `gemma3:27b` and `nomic-embed-text:v1.5` ([install](https://ollama.ai/))
- **Chrome/Chromium** (for browser automation)

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/yourusername/agent-workspace.git
cd agent-workspace
```

### 2. Start Neo4j

```bash
# Start Neo4j (adjust path to your installation)
neo4j start

# Verify it's running
neo4j status
```

### 3. Pull Ollama Models

```bash
ollama pull gemma3:27b
ollama pull nomic-embed-text:v1.5
```

### 4. Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your settings
nano .env
```

**Required environment variables:**

```env
# Neo4j Configuration
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=your_password

# Ollama Configuration
OLLAMA_HOST=http://localhost:11434
OLLAMA_MODEL=gemma3:27b
OLLAMA_EMBEDDING_MODEL=nomic-embed-text:v1.5

# Server Configuration
SERVER_PORT=8080
JWT_SECRET=your_random_secret_key_here

# Frontend URL (for CORS)
FRONTEND_URL=http://localhost:3000
```

### 5. Start Backend

```bash
cd backend

# Install dependencies
go mod download

# Run backend
go run cmd/server/main.go
```

Backend will start on `http://localhost:8080`

### 6. Start MCP Dynamic Thinking Server

```bash
cd mcp-dynamic-thinking

# Install dependencies
go mod download

# Build MCP server
go build -o mcp-dynamic-thinking cmd/server/main.go

# The backend will automatically connect to this MCP server via stdio
```

### 7. Start Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

Frontend will start on `http://localhost:3000`

### 8. Access the Application

Open your browser to `http://localhost:3000`

The terminal will automatically initialize:
```
✓ Neo4j started
✓ LightRAG initialized
✓ Watchdog monitoring active
🚀 Agent workspace ready!
```

## Project Structure

```
agent-workspace/
├── backend/                    # Go Fiber v3 backend
│   ├── cmd/
│   │   └── server/
│   │       └── main.go        # Entry point
│   ├── internal/
│   │   ├── agent/             # Agent controller
│   │   ├── browser/           # ChromeDP manager
│   │   ├── memory/            # Memory systems
│   │   ├── mcp/               # MCP client
│   │   ├── terminal/          # Terminal manager
│   │   ├── watchdog/          # Watchdog monitoring
│   │   └── websocket/         # WebSocket handlers
│   ├── pkg/
│   │   ├── jsonrpc/           # JSON-RPC 2.0 implementation
│   │   └── models/            # Shared models
│   ├── go.mod
│   └── go.sum
│
├── frontend/                   # React + Tailwind frontend
│   ├── public/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Chat/          # Chat interface
│   │   │   ├── FileTree/      # VS Code-style file tree
│   │   │   ├── OpenEvolve/    # Evolution tracking
│   │   │   ├── BottomPanel/   # Terminal, Browser, MCP, Logs
│   │   │   └── Layout/        # Main layout
│   │   ├── hooks/             # Custom React hooks
│   │   ├── services/          # API services
│   │   ├── styles/            # Global styles
│   │   ├── App.jsx
│   │   └── main.jsx
│   ├── package.json
│   ├── tailwind.config.js
│   └── vite.config.js
│
├── mcp-dynamic-thinking/       # MCP server for PRAR loop
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── perceive/          # Perception phase
│   │   ├── reason/            # Reasoning phase
│   │   ├── act/               # Action phase
│   │   └── reflect/           # Reflection phase
│   ├── go.mod
│   └── go.sum
│
├── docs/                       # Documentation
│   ├── API.md                 # API documentation
│   ├── ARCHITECTURE.md        # Architecture overview
│   ├── MEMORY.md              # Memory systems guide
│   └── MCP.md                 # MCP integration guide
│
├── scripts/                    # Utility scripts
│   ├── setup.sh               # Initial setup
│   ├── start-all.sh           # Start all services
│   └── stop-all.sh            # Stop all services
│
├── .env.example               # Example environment variables
├── .gitignore
├── docker-compose.yml         # Optional Docker setup
├── LICENSE
└── README.md
```

## Usage

### Basic Chat

Type in the center input field:

```
Create a login page with JWT authentication
```

The agent will:
1. **Perceive** - Analyze current project state
2. **Reason** - Generate multiple approaches (branched thinking)
3. **Act** - Create files and execute commands
4. **Reflect** - Critique performance and evolve strategies

### Browser Automation

```
Navigate to GitHub and find the go-light-rag repository
```

The agent will:
- Open browser in bottom panel
- Navigate to GitHub
- Search for repository
- Display numbered overlays for interaction
- Allow takeover for manual control

### File Operations

Click "Open Folder" in left panel to browse project files. The agent can:
- Create/edit/delete files
- Apply diffs
- Search across files
- Track changes in OpenEvolve panel

### Terminal Commands

```
Run tests for the authentication module
```

Terminal appears in bottom panel with:
- AI-executed commands (cyan)
- User commands (white)
- Takeover mode for manual control

### Knowledge Graph Queries

```
What authentication patterns have we used before?
```

Agent queries Neo4j knowledge graph and returns:
- Related concepts
- Past conversations
- Code references
- Success/failure patterns

## Development

### Backend Development

```bash
cd backend

# Run with hot reload (install air first)
go install github.com/cosmtrek/air@latest
air

# Run tests
go test ./...

# Build for production
go build -o agent-workspace cmd/server/main.go
```

### Frontend Development

```bash
cd frontend

# Development mode with hot reload
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### MCP Server Development

```bash
cd mcp-dynamic-thinking

# Test MCP server
go run cmd/server/main.go

# Build
go build -o mcp-dynamic-thinking cmd/server/main.go
```

## Configuration

### Neo4j Setup

1. Install Neo4j Community Edition 5.26
2. Start Neo4j: `neo4j start`
3. Access Neo4j Browser: `http://localhost:7474`
4. Set password for `neo4j` user
5. Update `.env` with credentials

### Ollama Setup

1. Install Ollama: `curl https://ollama.ai/install.sh | sh`
2. Pull models:
   ```bash
   ollama pull gemma3:27b
   ollama pull nomic-embed-text:v1.5
   ```
3. Verify: `ollama list`

### MCP Configuration

Edit `backend/config/mcp.json`:

```json
{
  "servers": {
    "dynamic-thinking": {
      "command": "../mcp-dynamic-thinking/mcp-dynamic-thinking",
      "args": [],
      "env": {
        "NEO4J_URI": "bolt://localhost:7687",
        "NEO4J_USER": "neo4j",
        "NEO4J_PASSWORD": "password"
      }
    }
  }
}
```

## API Documentation

See [docs/API.md](docs/API.md) for complete API reference.

### WebSocket Chat

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/chat?token=YOUR_JWT');

ws.send(JSON.stringify({
  id: "uuid",
  type: "user_command",
  payload: {
    command: "Create a login page"
  }
}));
```

### JSON-RPC 2.0 A2A

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/a2a');

ws.send(JSON.stringify({
  jsonrpc: "2.0",
  id: "req-001",
  method: "message/send",
  params: {
    message: {
      role: "user",
      parts: [{ type: "text", text: "Hello" }]
    }
  }
}));
```

### REST API

```bash
# Get agent status
curl http://localhost:8080/api/agent/status

# Query memory
curl -X POST http://localhost:8080/api/memory/query \
  -H "Content-Type: application/json" \
  -d '{"query": "authentication patterns"}'
```

## Troubleshooting

### Neo4j Connection Failed

```bash
# Check Neo4j status
neo4j status

# Check logs
tail -f /path/to/neo4j/logs/neo4j.log

# Restart Neo4j
neo4j restart
```

### Ollama Not Responding

```bash
# Check Ollama status
ollama list

# Restart Ollama service
sudo systemctl restart ollama

# Test model
ollama run gemma3:27b "Hello"
```

### Frontend Not Loading

```bash
# Clear node_modules and reinstall
cd frontend
rm -rf node_modules package-lock.json
npm install

# Check for port conflicts
lsof -i :3000
```

### Backend Crashes

```bash
# Check logs
tail -f backend/logs/app.log

# Verify Go version
go version  # Should be 1.21+

# Rebuild dependencies
cd backend
go mod tidy
go mod download
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Go Fiber](https://gofiber.io/) - Web framework
- [LightRAG](https://github.com/MegaGrindStone/go-light-rag) - Knowledge graph + vector DB
- [ChromeDP](https://github.com/chromedp/chromedp) - Browser automation
- [Ollama](https://ollama.ai/) - Local LLM runtime
- [Neo4j](https://neo4j.com/) - Graph database
- [React](https://react.dev/) - Frontend framework
- [Tailwind CSS](https://tailwindcss.com/) - Styling

## Support

- 📧 Email: support@agent-workspace.dev
- 💬 Discord: [Join our community](https://discord.gg/agent-workspace)
- 🐛 Issues: [GitHub Issues](https://github.com/yourusername/agent-workspace/issues)
- 📖 Docs: [Full Documentation](https://docs.agent-workspace.dev)

---

**Built with ❤️ for the AI agent community**


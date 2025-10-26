# Quick Setup Guide - Agent Workspace

## Prerequisites Installation

### 1. Install Go (1.21+)

**Linux/macOS:**
```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

**Windows:**
Download installer from https://go.dev/dl/ and run it.

### 2. Install Node.js (18+)

**Linux:**
```bash
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
node --version
npm --version
```

**macOS:**
```bash
brew install node@18
```

**Windows:**
Download installer from https://nodejs.org/

### 3. Install Neo4j Community Edition 5.26

**Linux (Debian/Ubuntu):**
```bash
# Add Neo4j repository
wget -O - https://debian.neo4j.com/neotechnology.gpg.key | sudo apt-key add -
echo 'deb https://debian.neo4j.com stable latest' | sudo tee /etc/apt/sources.list.d/neo4j.list
sudo apt-get update

# Install Neo4j
sudo apt-get install neo4j=1:5.26.0

# Start Neo4j
sudo systemctl enable neo4j
sudo systemctl start neo4j

# Check status
sudo systemctl status neo4j
```

**macOS:**
```bash
brew install neo4j@5.26
brew services start neo4j
```

**Windows:**
Download from https://neo4j.com/download-center/#community

**Initial Setup:**
```bash
# Access Neo4j Browser
open http://localhost:7474

# Default credentials:
# Username: neo4j
# Password: neo4j
# (You'll be prompted to change password on first login)
```

### 4. Install Ollama

**Linux:**
```bash
curl -fsSL https://ollama.ai/install.sh | sh

# Start Ollama service
sudo systemctl start ollama
sudo systemctl enable ollama
```

**macOS:**
```bash
brew install ollama
brew services start ollama
```

**Windows:**
Download from https://ollama.ai/download

**Pull Required Models:**
```bash
# Pull Gemma 3 27B (main LLM)
ollama pull gemma3:27b

# Pull nomic-embed-text (embeddings)
ollama pull nomic-embed-text:v1.5

# Verify models
ollama list
```

### 5. Install Chrome/Chromium (for browser automation)

**Linux:**
```bash
# Chrome
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
sudo dpkg -i google-chrome-stable_current_amd64.deb
sudo apt-get install -f

# Or Chromium
sudo apt-get install chromium-browser
```

**macOS:**
```bash
brew install --cask google-chrome
```

**Windows:**
Download from https://www.google.com/chrome/

---

## Project Setup

### 1. Extract Project

```bash
tar -xzf agent-workspace.tar.gz
cd agent-workspace
```

### 2. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration
nano .env
```

**Required Configuration (.env):**
```env
# Neo4j Configuration
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=YOUR_NEO4J_PASSWORD_HERE

# Ollama Configuration
OLLAMA_HOST=http://localhost:11434
OLLAMA_MODEL=gemma3:27b
OLLAMA_EMBEDDING_MODEL=nomic-embed-text:v1.5

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# JWT Configuration (generate a random 32+ character string)
JWT_SECRET=your_random_secret_key_minimum_32_characters_long
JWT_EXPIRATION=24h

# Frontend URL (for CORS)
FRONTEND_URL=http://localhost:3000

# Database Paths
CHROMEM_DB_PATH=./data/vec.db
BOLT_DB_PATH=./data/kv.db

# ChromeDP Configuration
CHROMEDP_HEADLESS=true
CHROMEDP_NO_SANDBOX=true

# Logging
LOG_LEVEL=info
LOG_FILE=./logs/app.log
```

### 3. Install Frontend Dependencies

```bash
cd frontend
npm install
cd ..
```

### 4. Install Backend Dependencies

```bash
cd backend
go mod download
cd ..
```

---

## Running the Application

### Option 1: Automated Start (Recommended)

```bash
# Make scripts executable
chmod +x scripts/*.sh

# Start all services
./scripts/start-all.sh
```

This will:
- ✅ Check all prerequisites
- ✅ Start Neo4j (if not running)
- ✅ Verify Ollama models
- ✅ Start backend server
- ✅ Start frontend dev server

### Option 2: Manual Start

**Terminal 1 - Backend:**
```bash
cd backend
go run cmd/server/main.go
```

**Terminal 2 - Frontend:**
```bash
cd frontend
npm run dev
```

### Stopping Services

```bash
./scripts/stop-all.sh
```

Or press `Ctrl+C` in each terminal.

---

## Accessing the Application

- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:8080
- **Neo4j Browser:** http://localhost:7474
- **WebSocket Chat:** ws://localhost:8080/ws/chat
- **A2A Protocol:** ws://localhost:8080/ws/a2a
- **Agent Card:** http://localhost:8080/.well-known/agent.json

---

## Verification Checklist

### ✅ Prerequisites Check

```bash
# Check Go
go version  # Should show 1.21+

# Check Node.js
node --version  # Should show v18+
npm --version

# Check Neo4j
neo4j status  # Should show "Neo4j is running"

# Check Ollama
ollama list  # Should show gemma3:27b and nomic-embed-text:v1.5

# Check Chrome
google-chrome --version  # Or chromium-browser --version
```

### ✅ Services Check

```bash
# Check Neo4j is accessible
curl http://localhost:7474

# Check Ollama is running
curl http://localhost:11434/api/tags

# Check backend is running (after starting)
curl http://localhost:8080/health

# Check frontend is running (after starting)
curl http://localhost:3000
```

---

## Troubleshooting

### Neo4j Won't Start

```bash
# Check logs
sudo journalctl -u neo4j -f

# Or
tail -f /var/log/neo4j/neo4j.log

# Restart
sudo systemctl restart neo4j
```

### Ollama Models Not Found

```bash
# Re-pull models
ollama pull gemma3:27b
ollama pull nomic-embed-text:v1.5

# Check Ollama service
sudo systemctl status ollama
sudo systemctl restart ollama
```

### Port Already in Use

```bash
# Check what's using port 8080
lsof -i :8080

# Check what's using port 3000
lsof -i :3000

# Kill process if needed
kill -9 <PID>
```

### Frontend Dependencies Error

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Backend Dependencies Error

```bash
cd backend
go clean -modcache
go mod tidy
go mod download
```

---

## Development Workflow

### Frontend Development

```bash
cd frontend

# Start dev server with hot reload
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint code
npm run lint
```

### Backend Development

```bash
cd backend

# Run with hot reload (install air first)
go install github.com/cosmtrek/air@latest
air

# Run tests
go test ./...

# Build binary
go build -o agent-workspace cmd/server/main.go

# Run binary
./agent-workspace
```

---

## System Requirements

### Minimum
- **CPU:** 4 cores
- **RAM:** 8 GB
- **Disk:** 10 GB free space
- **OS:** Linux, macOS, or Windows

### Recommended
- **CPU:** 8+ cores
- **RAM:** 16+ GB (for Gemma 3 27B)
- **Disk:** 20+ GB free space
- **OS:** Linux (Ubuntu 22.04+)

---

## Next Steps

1. ✅ Install all prerequisites
2. ✅ Configure environment (.env)
3. ✅ Start services
4. ✅ Access http://localhost:3000
5. 🚧 Implement remaining backend components (see IMPLEMENTATION_GUIDE.md)

---

## Support Resources

- **README.md** - Complete documentation
- **IMPLEMENTATION_GUIDE.md** - Backend implementation details
- **PROJECT_SUMMARY.md** - Project overview
- **docs/** - Design specifications

## Common Issues

### "Neo4j connection refused"
- Ensure Neo4j is running: `neo4j status`
- Check credentials in .env match Neo4j password
- Verify URI is correct: `bolt://localhost:7687`

### "Ollama model not found"
- Pull models: `ollama pull gemma3:27b`
- Check Ollama is running: `ollama list`

### "Frontend shows disconnected"
- Backend must be running first
- Check backend is on port 8080: `curl http://localhost:8080/health`
- Check CORS settings in .env

### "ChromeDP fails to start"
- Install Chrome/Chromium
- Set `CHROMEDP_NO_SANDBOX=true` in .env for Docker/VM environments

---

**Ready to go! 🚀**


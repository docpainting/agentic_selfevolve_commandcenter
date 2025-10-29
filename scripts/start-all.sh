#!/bin/bash

# Agentic Self-Evolving Command Center - Startup Script
# Kills existing processes and starts all services in correct order

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Service ports
NEO4J_BOLT_PORT=7687
NEO4J_HTTP_PORT=7474
OLLAMA_PORT=11434
BACKEND_PORT=8080
FRONTEND_PORT=3000
MCP_PORT=8081

echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  Agentic Self-Evolving Command Center - Startup Script    ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Function to kill process on port
kill_port() {
    local port=$1
    local service=$2
    echo -e "${YELLOW}→ Checking port $port ($service)...${NC}"
    
    # Find and kill process using the port
    local pid=$(lsof -ti:$port 2>/dev/null || true)
    if [ -n "$pid" ]; then
        echo -e "${RED}  ✗ Port $port in use by PID $pid, killing...${NC}"
        kill -9 $pid 2>/dev/null || true
        sleep 1
        echo -e "${GREEN}  ✓ Port $port freed${NC}"
    else
        echo -e "${GREEN}  ✓ Port $port available${NC}"
    fi
}

# Function to check if service is running
check_service() {
    local port=$1
    local service=$2
    local max_attempts=30
    local attempt=0
    
    echo -e "${YELLOW}→ Waiting for $service on port $port...${NC}"
    
    while [ $attempt -lt $max_attempts ]; do
        if lsof -ti:$port > /dev/null 2>&1; then
            echo -e "${GREEN}  ✓ $service is running on port $port${NC}"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done
    
    echo -e "${RED}  ✗ $service failed to start on port $port${NC}"
    return 1
}

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}STEP 1: Killing all previous processes (clean slate)${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Kill processes by PID files first
echo -e "${YELLOW}→ Killing processes from previous run...${NC}"
if [ -d "pids" ]; then
    for pidfile in pids/*.pid; do
        if [ -f "$pidfile" ]; then
            pid=$(cat "$pidfile" 2>/dev/null || echo "")
            if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
                service_name=$(basename "$pidfile" .pid)
                echo -e "${YELLOW}  → Killing $service_name (PID: $pid)${NC}"
                kill -9 "$pid" 2>/dev/null || true
            fi
        fi
    done
    rm -f pids/*.pid
    echo -e "${GREEN}  ✓ Previous PIDs killed${NC}"
else
    echo -e "${GREEN}  ✓ No previous PIDs found${NC}"
fi

# Kill by port (backup method)
kill_port $BACKEND_PORT "Backend Server"
kill_port $FRONTEND_PORT "Frontend Dev Server"

# Kill any lingering Python processes
echo -e "${YELLOW}→ Killing Python agent processes...${NC}"
pkill -f evoagentx_service.py 2>/dev/null || true
pkill -f chrome_devtools_mcp.py 2>/dev/null || true
sleep 1
echo -e "${GREEN}  ✓ Python processes killed${NC}"

# Kill any lingering Chrome processes
echo -e "${YELLOW}→ Killing Chrome browser processes...${NC}"
pkill -f "chrome.*--remote-debugging-port" 2>/dev/null || true
sleep 1
echo -e "${GREEN}  ✓ Chrome processes killed${NC}"

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}STEP 2: Verifying prerequisites${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Check Neo4j
echo -e "${YELLOW}→ Checking Neo4j...${NC}"
if systemctl is-active --quiet neo4j; then
    echo -e "${GREEN}  ✓ Neo4j is running${NC}"
else
    echo -e "${YELLOW}  ⚠ Neo4j not running, attempting to start...${NC}"
    sudo systemctl start neo4j
    sleep 5
    if systemctl is-active --quiet neo4j; then
        echo -e "${GREEN}  ✓ Neo4j started successfully${NC}"
    else
        echo -e "${RED}  ✗ Failed to start Neo4j${NC}"
        exit 1
    fi
fi

# Check Ollama (on host machine)
echo -e "${YELLOW}→ Checking Ollama on host (10.0.2.2:11434)...${NC}"
if curl -s http://10.0.2.2:11434/api/tags > /dev/null 2>&1; then
    echo -e "${GREEN}  ✓ Ollama is accessible${NC}"
else
    echo -e "${RED}  ✗ Ollama not accessible on host machine${NC}"
    echo -e "${YELLOW}  Make sure Ollama is running on the host machine${NC}"
    exit 1
fi

# Check Go
echo -e "${YELLOW}→ Checking Go...${NC}"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}  ✓ Go found: $GO_VERSION${NC}"
else
    echo -e "${RED}  ✗ Go not found${NC}"
    exit 1
fi

# Check Node.js
echo -e "${YELLOW}→ Checking Node.js...${NC}"
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    echo -e "${GREEN}  ✓ Node.js found: $NODE_VERSION${NC}"
else
    echo -e "${RED}  ✗ Node.js not found${NC}"
    exit 1
fi

# Check .env file
echo -e "${YELLOW}→ Checking .env configuration...${NC}"
if [ -f ".env" ]; then
    if grep -q "NEO4J_PASSWORD=your_neo4j_password_here" .env; then
        echo -e "${RED}  ✗ Neo4j password not set in .env${NC}"
        exit 1
    fi
    echo -e "${GREEN}  ✓ .env file configured${NC}"
else
    echo -e "${RED}  ✗ .env file not found${NC}"
    exit 1
fi

# Create data directories
echo -e "${YELLOW}→ Creating data directories...${NC}"
mkdir -p backend/data backend/logs pids
echo -e "${GREEN}  ✓ Data directories ready${NC}"

# Check Python dependencies for MCP server
echo -e "${YELLOW}→ Checking Python dependencies for browser automation...${NC}"
if python3 -c "from selenium import webdriver" 2>/dev/null; then
    echo -e "${GREEN}  ✓ Selenium installed${NC}"
else
    echo -e "${RED}  ✗ Selenium not found, installing dependencies...${NC}"
    sudo pip3 install --target=/home/thedoc/venv/lib/python3.10/site-packages selenium websocket-client typing_extensions certifi urllib3 trio trio-websocket wsproto h11 outcome sniffio sortedcontainers attrs > /dev/null 2>&1
    echo -e "${GREEN}  ✓ Python dependencies installed${NC}"
fi

# Check ChromeDriver
echo -e "${YELLOW}→ Checking ChromeDriver...${NC}"
if command -v chromedriver &> /dev/null; then
    echo -e "${GREEN}  ✓ ChromeDriver found: $(which chromedriver)${NC}"
else
    echo -e "${RED}  ✗ ChromeDriver not found${NC}"
    echo -e "${YELLOW}  Install with: sudo apt install chromium-chromedriver${NC}"
    exit 1
fi

# Note: Chrome will be started automatically by the Python MCP server
# when the backend connects to it. The MCP server controls Chrome via Selenium.
echo -e "${YELLOW}→ Chrome browser automation...${NC}"
echo -e "${GREEN}  ✓ Chrome will be started by MCP server (headless mode)${NC}"

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}STEP 3: Starting services${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Start Backend
echo -e "${CYAN}→ Starting Backend Server (port $BACKEND_PORT)...${NC}"
cd backend
# Build first to ensure latest code
go build -o server ./cmd/server
./server > ../logs/backend.log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > ../pids/backend.pid
cd ..
echo -e "${GREEN}  ✓ Backend started (PID: $BACKEND_PID)${NC}"

# Wait for backend to be ready
if ! check_service $BACKEND_PORT "Backend"; then
    echo -e "${RED}Backend failed to start. Check logs/backend.log${NC}"
    exit 1
fi

# Start Frontend
echo -e "${CYAN}→ Starting Frontend Dev Server (port $FRONTEND_PORT)...${NC}"
cd frontend
npm run dev > ../logs/frontend.log 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > ../pids/frontend.pid
cd ..
echo -e "${GREEN}  ✓ Frontend started (PID: $FRONTEND_PID)${NC}"

# Wait for frontend to be ready
if ! check_service $FRONTEND_PORT "Frontend"; then
    echo -e "${RED}Frontend failed to start. Check logs/frontend.log${NC}"
    exit 1
fi

# Start Python MCP Servers
echo ""
echo -e "${CYAN}→ Starting Python MCP Servers...${NC}"

# 1. Dynamic Thinking MCP
if [ -f "backend/mcp_servers/dynamic_thinking/server.py" ]; then
    python3 backend/mcp_servers/dynamic_thinking/server.py > logs/mcp_dynamic_thinking.log 2>&1 &
    MCP_DYNAMIC_PID=$!
    echo $MCP_DYNAMIC_PID > pids/mcp_dynamic_thinking.pid
    echo -e "${GREEN}  ✓ Dynamic Thinking MCP started (PID: $MCP_DYNAMIC_PID)${NC}"
fi

# 2. OpenEvolve MCP
if [ -f "backend/mcp_servers/openevolve/server.py" ]; then
    python3 backend/mcp_servers/openevolve/server.py > logs/mcp_openevolve.log 2>&1 &
    MCP_OPENEVOLVE_PID=$!
    echo $MCP_OPENEVOLVE_PID > pids/mcp_openevolve.pid
    echo -e "${GREEN}  ✓ OpenEvolve MCP started (PID: $MCP_OPENEVOLVE_PID)${NC}"
fi

# 3. Terminal MCP
if [ -f "backend/mcp_servers/terminal/server.py" ]; then
    python3 backend/mcp_servers/terminal/server.py > logs/mcp_terminal.log 2>&1 &
    MCP_TERMINAL_PID=$!
    echo $MCP_TERMINAL_PID > pids/mcp_terminal.pid
    echo -e "${GREEN}  ✓ Terminal MCP started (PID: $MCP_TERMINAL_PID)${NC}"
fi

# 4. Dynamic Sequential Thinking MCP
if [ -f "backend/mcp_servers/dynamic_sequential_thinking/server.py" ]; then
    python3 backend/mcp_servers/dynamic_sequential_thinking/server.py > logs/mcp_sequential.log 2>&1 &
    MCP_SEQUENTIAL_PID=$!
    echo $MCP_SEQUENTIAL_PID > pids/mcp_sequential.pid
    echo -e "${GREEN}  ✓ Sequential Thinking MCP started (PID: $MCP_SEQUENTIAL_PID)${NC}"
fi

# 5. Terminal Browser MCP
if [ -f "backend/mcp_servers/terminal_browser/server.py" ]; then
    python3 backend/mcp_servers/terminal_browser/server.py > logs/mcp_terminal_browser.log 2>&1 &
    MCP_BROWSER_PID=$!
    echo $MCP_BROWSER_PID > pids/mcp_terminal_browser.pid
    echo -e "${GREEN}  ✓ Terminal Browser MCP started (PID: $MCP_BROWSER_PID)${NC}"
fi

sleep 2
echo -e "${GREEN}  ✓ All Python MCP servers started${NC}"

# Start EvoAgentX Service (optional - for Python agent workflows)
echo ""
echo -e "${CYAN}→ Starting EvoAgentX Service (optional)...${NC}"
if command -v python3.11 &> /dev/null && python3.11 -c "import evoagentx" 2>/dev/null; then
    # EvoAgentX is installed, start the service
    python3.11 evoagentx_service.py > logs/evoagentx.log 2>&1 &
    EVOAGENTX_PID=$!
    echo $EVOAGENTX_PID > pids/evoagentx.pid
    echo -e "${GREEN}  ✓ EvoAgentX Service started (PID: $EVOAGENTX_PID)${NC}"
else
    echo -e "${YELLOW}  ⚠ EvoAgentX not installed (optional - Python agent workflows)${NC}"
    echo -e "${YELLOW}    Install with: ./scripts/install-evoagentx.sh${NC}"
fi

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              All Services Started Successfully!            ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}Service Status:${NC}"
echo -e "  ${GREEN}✓${NC} Neo4j:          bolt://localhost:$NEO4J_BOLT_PORT"
echo -e "  ${GREEN}✓${NC} Neo4j Browser:  http://localhost:$NEO4J_HTTP_PORT"
echo -e "  ${GREEN}✓${NC} Ollama:         http://10.0.2.2:$OLLAMA_PORT"
echo -e "  ${GREEN}✓${NC} Backend API:    http://localhost:$BACKEND_PORT"
echo -e "  ${GREEN}✓${NC} Frontend:       http://localhost:$FRONTEND_PORT"
echo ""
echo -e "${CYAN}Quick Links:${NC}"
echo -e "  • Application:  ${BLUE}http://localhost:$FRONTEND_PORT${NC}"
echo -e "  • Health Check: ${BLUE}http://localhost:$BACKEND_PORT/health${NC}"
echo -e "  • Neo4j Browser:${BLUE}http://localhost:$NEO4J_HTTP_PORT${NC}"
echo ""
echo -e "${CYAN}Logs:${NC}"
echo -e "  • Backend:  tail -f logs/backend.log"
echo -e "  • Frontend: tail -f logs/frontend.log"
echo ""
echo -e "${CYAN}To stop all services:${NC}"
echo -e "  ./stop.sh"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all services...${NC}"
echo ""

# Create PID directory
mkdir -p pids

# Trap Ctrl+C to cleanup
cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down services...${NC}"
    
    # Stop EvoAgentX
    if [ -n "$EVOAGENTX_PID" ]; then
        echo -e "${YELLOW}→ Stopping EvoAgentX Service (PID: $EVOAGENTX_PID)...${NC}"
        kill $EVOAGENTX_PID 2>/dev/null || true
    fi
    
    # Stop all MCP servers
    for mcp_pid in "$MCP_DYNAMIC_PID" "$MCP_OPENEVOLVE_PID" "$MCP_TERMINAL_PID" "$MCP_SEQUENTIAL_PID" "$MCP_BROWSER_PID"; do
        if [ -n "$mcp_pid" ]; then
            kill $mcp_pid 2>/dev/null || true
        fi
    done
    echo -e "${YELLOW}→ Stopped all MCP servers${NC}"
    
    # Stop Frontend
    if [ -n "$FRONTEND_PID" ]; then
        echo -e "${YELLOW}→ Stopping Frontend (PID: $FRONTEND_PID)...${NC}"
        kill $FRONTEND_PID 2>/dev/null || true
    fi
    
    # Stop Backend
    if [ -n "$BACKEND_PID" ]; then
        echo -e "${YELLOW}→ Stopping Backend (PID: $BACKEND_PID)...${NC}"
        kill $BACKEND_PID 2>/dev/null || true
    fi
    
    # Clean up PID files
    rm -f pids/*.pid
    
    echo -e "${GREEN}✓ All services stopped${NC}"
    exit 0
}

trap cleanup INT TERM

# Wait for processes
wait

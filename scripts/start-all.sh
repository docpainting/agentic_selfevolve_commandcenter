#!/bin/bash

# Agent Workspace - Start All Services
# This script starts all required services for the agent workspace

set -e

echo "🚀 Starting Agent Workspace..."
echo ""

# Check prerequisites
echo "Checking prerequisites..."

# Check Neo4j
if ! command -v neo4j &> /dev/null; then
    echo "❌ Neo4j not found. Please install Neo4j 5.26 Community Edition."
    exit 1
fi

# Check Ollama
if ! command -v ollama &> /dev/null; then
    echo "❌ Ollama not found. Please install Ollama."
    exit 1
fi

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Go not found. Please install Go 1.21+."
    exit 1
fi

# Check Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js not found. Please install Node.js 18+."
    exit 1
fi

echo "✅ All prerequisites found"
echo ""

# Start Neo4j
echo "Starting Neo4j..."
neo4j status > /dev/null 2>&1 || neo4j start
sleep 5
echo "✅ Neo4j started"

# Check Ollama models
echo "Checking Ollama models..."
if ! ollama list | grep -q "gemma3:27b"; then
    echo "Pulling gemma3:27b..."
    ollama pull gemma3:27b
fi
if ! ollama list | grep -q "nomic-embed-text:v1.5"; then
    echo "Pulling nomic-embed-text:v1.5..."
    ollama pull nomic-embed-text:v1.5
fi
echo "✅ Ollama models ready"
echo ""

# Start backend
echo "Starting backend..."
cd backend
go mod download > /dev/null 2>&1
go run cmd/server/main.go &
BACKEND_PID=$!
echo "✅ Backend started (PID: $BACKEND_PID)"
echo ""

# Wait for backend to be ready
echo "Waiting for backend to be ready..."
sleep 5

# Start frontend
echo "Starting frontend..."
cd ../frontend
if [ ! -d "node_modules" ]; then
    echo "Installing frontend dependencies..."
    npm install
fi
npm run dev &
FRONTEND_PID=$!
echo "✅ Frontend started (PID: $FRONTEND_PID)"
echo ""

echo "✅ All services started!"
echo ""
echo "Access the application at: http://localhost:3000"
echo ""
echo "To stop all services, run: ./scripts/stop-all.sh"
echo ""

# Save PIDs
echo $BACKEND_PID > /tmp/agent-workspace-backend.pid
echo $FRONTEND_PID > /tmp/agent-workspace-frontend.pid

# Wait for user interrupt
trap "echo ''; echo 'Stopping services...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit 0" INT

wait


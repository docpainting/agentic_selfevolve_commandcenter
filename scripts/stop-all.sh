#!/bin/bash

# Agent Workspace - Stop All Services

echo "Stopping Agent Workspace services..."

# Stop backend
if [ -f /tmp/agent-workspace-backend.pid ]; then
    BACKEND_PID=$(cat /tmp/agent-workspace-backend.pid)
    kill $BACKEND_PID 2>/dev/null && echo "✅ Backend stopped" || echo "⚠️  Backend not running"
    rm /tmp/agent-workspace-backend.pid
fi

# Stop frontend
if [ -f /tmp/agent-workspace-frontend.pid ]; then
    FRONTEND_PID=$(cat /tmp/agent-workspace-frontend.pid)
    kill $FRONTEND_PID 2>/dev/null && echo "✅ Frontend stopped" || echo "⚠️  Frontend not running"
    rm /tmp/agent-workspace-frontend.pid
fi

# Optionally stop Neo4j (commented out by default)
# echo "Stopping Neo4j..."
# neo4j stop

echo ""
echo "All services stopped."


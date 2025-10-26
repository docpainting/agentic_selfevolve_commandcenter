# Dynamic Thinking MCP Server - Usage Guide

## Quick Start

### 1. Install Dependencies
```bash
cd backend/mcp_servers/dynamic_thinking
pip install -r requirements.txt
```

### 2. Configure Environment
```bash
export NEO4J_URI="bolt://localhost:7687"
export NEO4J_USER="neo4j"
export NEO4J_PASSWORD="your_password"
export OLLAMA_BASE_URL="http://localhost:11434"
```

### 3. Run Server
```bash
python server.py
```

## Tools Available to Gemma 3

The model can trigger these 6 tools via function calling:

### perceive
Deep understanding of tasks using systems thinking and 3 reasoning modes.

### reason  
Generate strategies with automatic online retrieval when confidence is low.

### act
Execute plans with watchdog monitoring.

### reflect
Learn from experiences and create patterns.

### query_memory
Search past experiences semantically.

### evolve_prompt
Improve prompts based on learnings.

## Example: Complete PRAR Loop

```python
# Gemma 3 triggers these tools automatically

# 1. PERCEIVE
perception = perceive({
    "task_id": "opt-1",
    "task": "Optimize HandleTask",
    "goal": "Improve performance",
    "entity": {"name": "HandleTask", "type": "function"}
})

# 2. REASON
reasoning = reason({
    "task_id": "opt-1",
    "perception_id": perception["perception_id"],
    "perception": perception
})

# 3. ACT
execution = act({
    "task_id": "opt-1",
    "reasoning_id": reasoning["reasoning_id"],
    "selected_branch": reasoning["selected_branch"],
    "perception": perception
})

# 4. REFLECT
reflection = reflect({
    "task_id": "opt-1",
    "execution_id": execution["execution_id"],
    "perception": perception,
    "reasoning": reasoning,
    "execution": execution
})
```

## MCP Client Configuration

Add to your MCP client config:
```json
{
  "mcpServers": {
    "dynamic-thinking": {
      "command": "python",
      "args": ["/path/to/backend/mcp_servers/dynamic_thinking/server.py"]
    }
  }
}
```

Now Gemma 3 can use these tools! 🧠

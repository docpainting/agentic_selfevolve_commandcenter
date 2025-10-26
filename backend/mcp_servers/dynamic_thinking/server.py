#!/usr/bin/env python3
"""
Dynamic Thinking MCP Server

Core MCP server for agent self-awareness implementing:
- Enhanced multi-modal perception (systems thinking, contextual, meta-perception)
- Multi-branch reasoning with confidence-based online retrieval
- Dynamic execution with watchdog monitoring
- Self-critique and pattern creation
- Semantic memory search
- Prompt evolution

This server enables the agent to think deeply about its own code and make
intelligent decisions based on multiple reasoning modes.
"""

import asyncio
import logging
from typing import Any, Dict, List, Optional
from datetime import datetime
import uuid

from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import Tool, TextContent

# Import tool implementations
from tools.perceive import PerceiveTool
from tools.reason import ReasonTool
from tools.act import ActTool
from tools.reflect import ReflectTool
from tools.query_memory import QueryMemoryTool
from tools.evolve_prompt import EvolvePromptTool

# Import storage
from storage.lightrag_client import LightRAGClient
from storage.session_manager import SessionManager

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class DynamicThinkingServer:
    """MCP Server for dynamic thinking and self-awareness"""
    
    def __init__(self):
        self.server = Server("dynamic-thinking")
        self.lightrag_client = None
        self.session_manager = SessionManager()
        
        # Tool instances
        self.perceive_tool = None
        self.reason_tool = None
        self.act_tool = None
        self.reflect_tool = None
        self.query_memory_tool = None
        self.evolve_prompt_tool = None
        
        # Register handlers
        self.server.list_tools()(self.list_tools)
        self.server.call_tool()(self.call_tool)
    
    async def initialize(self):
        """Initialize LightRAG client and tools"""
        logger.info("Initializing Dynamic Thinking MCP Server...")
        
        # Initialize LightRAG client
        self.lightrag_client = LightRAGClient(
            llm_model="gemma3:27b",
            embedding_model="nomic-embed-text:v1.5",
            ollama_base_url="http://localhost:11434",
            neo4j_uri="bolt://localhost:7687",
            neo4j_user="neo4j",
            neo4j_password="password",  # TODO: Use env var
            vector_db_path="/var/lib/agent/vectors.db",
            kv_db_path="/var/lib/agent/kv.db"
        )
        
        await self.lightrag_client.initialize()
        
        # Initialize tools
        self.perceive_tool = PerceiveTool(self.lightrag_client)
        self.reason_tool = ReasonTool(self.lightrag_client)
        self.act_tool = ActTool(self.lightrag_client)
        self.reflect_tool = ReflectTool(self.lightrag_client)
        self.query_memory_tool = QueryMemoryTool(self.lightrag_client)
        self.evolve_prompt_tool = EvolvePromptTool(self.lightrag_client)
        
        logger.info("Dynamic Thinking MCP Server initialized successfully")
    
    async def list_tools(self) -> List[Tool]:
        """List all available tools"""
        return [
            Tool(
                name="perceive",
                description=(
                    "Deep analytical understanding using systems thinking, contextual reasoning, "
                    "meta-perception, and three reasoning modes (deductive, inductive, abductive). "
                    "Use this when you need to deeply understand a task, analyze code, or make "
                    "complex decisions. Returns comprehensive perception including system view, "
                    "context, alternative framings, and reasoning analysis."
                ),
                inputSchema={
                    "type": "object",
                    "properties": {
                        "task_id": {
                            "type": "string",
                            "description": "Unique identifier for this task"
                        },
                        "task": {
                            "type": "string",
                            "description": "The task to perceive and analyze"
                        },
                        "goal": {
                            "type": "string",
                            "description": "The goal to achieve"
                        },
                        "entity": {
                            "type": "object",
                            "properties": {
                                "name": {"type": "string"},
                                "type": {"type": "string"},
                                "code": {"type": "string"}
                            },
                            "required": ["name", "type"],
                            "description": "The primary entity being analyzed"
                        },
                        "context": {
                            "type": "object",
                            "description": "Additional context about the task"
                        }
                    },
                    "required": ["task_id", "task", "goal", "entity"]
                }
            ),
            
            Tool(
                name="reason",
                description=(
                    "Multi-branch reasoning with confidence checking and online retrieval. "
                    "Generates multiple reasoning strategies, evaluates confidence, and "
                    "automatically searches online (web, GitHub, Stack Overflow, YouTube) "
                    "when confidence is low. Use this after perceive to generate strategies. "
                    "Returns reasoning branches with selected best strategy."
                ),
                inputSchema={
                    "type": "object",
                    "properties": {
                        "task_id": {
                            "type": "string",
                            "description": "Task identifier"
                        },
                        "perception_id": {
                            "type": "string",
                            "description": "UUID from perceive tool output"
                        },
                        "perception": {
                            "type": "object",
                            "description": "Full perception object from perceive tool"
                        },
                        "max_branches": {
                            "type": "integer",
                            "description": "Maximum reasoning branches to generate",
                            "default": 3
                        }
                    },
                    "required": ["task_id", "perception_id", "perception"]
                }
            ),
            
            Tool(
                name="act",
                description=(
                    "Execute action plan with dynamic adjustment and watchdog monitoring. "
                    "Supports parallel execution for cross-entity tasks. Use this after "
                    "reason to execute the selected strategy. Returns execution results "
                    "with performance metrics and watchdog events."
                ),
                inputSchema={
                    "type": "object",
                    "properties": {
                        "task_id": {
                            "type": "string",
                            "description": "Task identifier"
                        },
                        "reasoning_id": {
                            "type": "string",
                            "description": "UUID from reason tool output"
                        },
                        "selected_branch": {
                            "type": "object",
                            "description": "Selected reasoning branch from reason tool"
                        },
                        "perception": {
                            "type": "object",
                            "description": "Perception object for context"
                        },
                        "dry_run": {
                            "type": "boolean",
                            "description": "If true, simulate execution without making changes",
                            "default": False
                        }
                    },
                    "required": ["task_id", "reasoning_id", "selected_branch", "perception"]
                }
            ),
            
            Tool(
                name="reflect",
                description=(
                    "Reflect on complete process, extract learnings, create patterns, and "
                    "generate recommendations. Use this after act to learn from the experience. "
                    "Returns learnings, discovered patterns, recommendations for similar entities, "
                    "and training data for future improvement."
                ),
                inputSchema={
                    "type": "object",
                    "properties": {
                        "task_id": {
                            "type": "string",
                            "description": "Task identifier"
                        },
                        "execution_id": {
                            "type": "string",
                            "description": "UUID from act tool output"
                        },
                        "perception": {
                            "type": "object",
                            "description": "Perception object"
                        },
                        "reasoning": {
                            "type": "object",
                            "description": "Reasoning object"
                        },
                        "execution": {
                            "type": "object",
                            "description": "Execution object"
                        }
                    },
                    "required": ["task_id", "execution_id", "perception", "reasoning", "execution"]
                }
            ),
            
            Tool(
                name="query_memory",
                description=(
                    "Search agent's memory using semantic search. Query past perceptions, "
                    "reasoning, executions, reflections, or patterns. Use this to find "
                    "relevant past experiences before making decisions. Returns relevant "
                    "results with metadata and related entities."
                ),
                inputSchema={
                    "type": "object",
                    "properties": {
                        "query": {
                            "type": "string",
                            "description": "Search query"
                        },
                        "context": {
                            "type": "string",
                            "description": "Additional context for the query"
                        },
                        "filters": {
                            "type": "object",
                            "properties": {
                                "type": {
                                    "type": "string",
                                    "enum": ["perception", "reasoning", "execution", "reflection", "pattern"]
                                },
                                "success": {"type": "boolean"},
                                "min_confidence": {"type": "number"}
                            },
                            "description": "Filters for search results"
                        },
                        "max_results": {
                            "type": "integer",
                            "description": "Maximum number of results",
                            "default": 10
                        }
                    },
                    "required": ["query"]
                }
            ),
            
            Tool(
                name="evolve_prompt",
                description=(
                    "Evolve prompts based on past performance and learnings. Use this to "
                    "improve prompts over time. Returns evolved prompt with explanation of changes."
                ),
                inputSchema={
                    "type": "object",
                    "properties": {
                        "current_prompt": {
                            "type": "string",
                            "description": "Current prompt to evolve"
                        },
                        "context": {
                            "type": "string",
                            "description": "Context for prompt usage"
                        },
                        "past_performance": {
                            "type": "object",
                            "properties": {
                                "success_rate": {"type": "number"},
                                "avg_confidence": {"type": "number"},
                                "common_failures": {
                                    "type": "array",
                                    "items": {"type": "string"}
                                }
                            },
                            "description": "Past performance metrics"
                        },
                        "learnings": {
                            "type": "array",
                            "items": {"type": "string"},
                            "description": "Learnings to incorporate"
                        }
                    },
                    "required": ["current_prompt"]
                }
            )
        ]
    
    async def call_tool(self, name: str, arguments: Dict[str, Any]) -> List[TextContent]:
        """Execute a tool"""
        logger.info(f"Calling tool: {name}")
        logger.debug(f"Arguments: {arguments}")
        
        try:
            # Route to appropriate tool
            if name == "perceive":
                result = await self.perceive_tool.execute(arguments)
            elif name == "reason":
                result = await self.reason_tool.execute(arguments)
            elif name == "act":
                result = await self.act_tool.execute(arguments)
            elif name == "reflect":
                result = await self.reflect_tool.execute(arguments)
            elif name == "query_memory":
                result = await self.query_memory_tool.execute(arguments)
            elif name == "evolve_prompt":
                result = await self.evolve_prompt_tool.execute(arguments)
            else:
                raise ValueError(f"Unknown tool: {name}")
            
            # Store session data
            task_id = arguments.get("task_id")
            if task_id:
                self.session_manager.update_session(task_id, name, result)
            
            # Return result as TextContent
            import json
            return [TextContent(
                type="text",
                text=json.dumps(result, indent=2)
            )]
            
        except Exception as e:
            logger.error(f"Error executing tool {name}: {e}", exc_info=True)
            return [TextContent(
                type="text",
                text=json.dumps({
                    "error": str(e),
                    "tool": name,
                    "timestamp": datetime.utcnow().isoformat()
                }, indent=2)
            )]


async def main():
    """Main entry point"""
    server_instance = DynamicThinkingServer()
    await server_instance.initialize()
    
    # Run MCP server
    async with stdio_server() as (read_stream, write_stream):
        logger.info("Dynamic Thinking MCP Server running on stdio")
        await server_instance.server.run(
            read_stream,
            write_stream,
            server_instance.server.create_initialization_options()
        )


if __name__ == "__main__":
    asyncio.run(main())


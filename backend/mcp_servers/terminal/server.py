#!/usr/bin/env python3
"""
Terminal Agent MCP Server

Natural language to Linux commands using comanderanch/Linux-Buster model.
Trained on 250+ commands for safe terminal operations.
"""

import asyncio
import logging
import subprocess
from typing import Any, Dict, List

from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import Tool, TextContent
import ollama

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class TerminalAgentServer:
    """MCP Server for terminal operations"""
    
    def __init__(self):
        self.server = Server("terminal-agent")
        self.ollama_client = ollama.Client()
        self.model = "comanderanch/Linux-Buster:latest"
        
        self.server.list_tools()(self.list_tools)
        self.server.call_tool()(self.call_tool)
    
    async def list_tools(self) -> List[Tool]:
        """List available tools"""
        return [
            Tool(
                name="natural_to_command",
                description="Convert natural language to Linux command using specialized model.",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "instruction": {"type": "string", "description": "Natural language instruction"},
                        "context": {"type": "string", "description": "Additional context"}
                    },
                    "required": ["instruction"]
                }
            ),
            
            Tool(
                name="execute_command",
                description="Execute Linux command safely with validation.",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "command": {"type": "string"},
                        "dry_run": {"type": "boolean", "default": False}
                    },
                    "required": ["command"]
                }
            ),
            
            Tool(
                name="explain_command",
                description="Explain what a Linux command does.",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "command": {"type": "string"}
                    },
                    "required": ["command"]
                }
            )
        ]
    
    async def call_tool(self, name: str, arguments: Dict[str, Any]) -> List[TextContent]:
        """Execute tool"""
        import json
        
        try:
            if name == "natural_to_command":
                result = await self.natural_to_command(arguments)
            elif name == "execute_command":
                result = await self.execute_command(arguments)
            elif name == "explain_command":
                result = await self.explain_command(arguments)
            else:
                raise ValueError(f"Unknown tool: {name}")
            
            return [TextContent(type="text", text=json.dumps(result, indent=2))]
        except Exception as e:
            logger.error(f"Error in {name}: {e}", exc_info=True)
            return [TextContent(type="text", text=json.dumps({"error": str(e)}, indent=2))]
    
    async def natural_to_command(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """Convert natural language to command"""
        instruction = params["instruction"]
        context = params.get("context", "")
        
        prompt = f"""Convert this natural language instruction to a Linux command:

Instruction: {instruction}
{f'Context: {context}' if context else ''}

Provide only the command, no explanation."""
        
        try:
            response = self.ollama_client.generate(
                model=self.model,
                prompt=prompt
            )
            
            command = response['response'].strip()
            
            return {
                "command": command,
                "instruction": instruction,
                "safe": self._is_safe_command(command),
                "explanation": ""
            }
        except Exception as e:
            logger.error(f"Error generating command: {e}")
            return {"error": str(e)}
    
    async def execute_command(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """Execute command"""
        command = params["command"]
        dry_run = params.get("dry_run", False)
        
        if not self._is_safe_command(command):
            return {
                "error": "Command rejected as unsafe",
                "command": command,
                "safe": False
            }
        
        if dry_run:
            return {
                "command": command,
                "dry_run": True,
                "safe": True,
                "output": "[DRY RUN - Command not executed]"
            }
        
        try:
            result = subprocess.run(
                command,
                shell=True,
                capture_output=True,
                text=True,
                timeout=30
            )
            
            return {
                "command": command,
                "exit_code": result.returncode,
                "stdout": result.stdout,
                "stderr": result.stderr,
                "success": result.returncode == 0
            }
        except subprocess.TimeoutExpired:
            return {"error": "Command timed out"}
        except Exception as e:
            return {"error": str(e)}
    
    async def explain_command(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """Explain command"""
        command = params["command"]
        
        prompt = f"""Explain what this Linux command does:

Command: {command}

Provide a clear, concise explanation."""
        
        try:
            response = self.ollama_client.generate(
                model=self.model,
                prompt=prompt
            )
            
            return {
                "command": command,
                "explanation": response['response'].strip()
            }
        except Exception as e:
            return {"error": str(e)}
    
    def _is_safe_command(self, command: str) -> bool:
        """Check if command is safe to execute"""
        dangerous_patterns = [
            "rm -rf /",
            ":(){ :|:& };:",  # Fork bomb
            "mkfs",
            "dd if=/dev/zero",
            "> /dev/sda",
            "chmod -R 777 /",
            "wget | sh",
            "curl | sh"
        ]
        
        command_lower = command.lower()
        return not any(pattern in command_lower for pattern in dangerous_patterns)


async def main():
    server = TerminalAgentServer()
    async with stdio_server() as (read_stream, write_stream):
        logger.info("Terminal Agent MCP Server running")
        await server.server.run(read_stream, write_stream, server.server.create_initialization_options())


if __name__ == "__main__":
    asyncio.run(main())


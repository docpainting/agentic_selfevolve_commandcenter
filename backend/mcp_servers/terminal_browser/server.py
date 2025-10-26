"""
Terminal & Browser MCP Server
Unified server for terminal operations and browser automation using Chrome DevTools Protocol.
Uses Gemma 3 vision for screenshot analysis when needed.
"""

import asyncio
import os
import subprocess
import base64
import json
from pathlib import Path
from typing import Optional, List, Dict, Any
from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp import types
import ollama
import websockets
import aiohttp
from PIL import Image, ImageDraw, ImageFont
import io

# Configuration
OLLAMA_BASE_URL = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
TERMINAL_MODEL = "comanderanch/Linux-Buster:latest"
VISION_MODEL = "gemma3:27b"  # Multimodal model for screenshot analysis
CHROME_DEBUGGING_PORT = 9222

# Dangerous command patterns to block
DANGEROUS_PATTERNS = [
    r"rm\s+-rf\s+/",
    r"mkfs",
    r"dd\s+if=/dev/zero",
    r":(){ :|:& };:",  # Fork bomb
    r">\s*/dev/sd",
]

class ChromeDevTools:
    """Chrome DevTools Protocol client"""
    
    def __init__(self):
        self.ws = None
        self.session_id = None
        self.message_id = 0
        self.responses = {}
        self.events = []
    
    async def connect(self):
        """Connect to Chrome via DevTools Protocol"""
        # Get list of targets
        async with aiohttp.ClientSession() as session:
            async with session.get(f"http://localhost:{CHROME_DEBUGGING_PORT}/json") as resp:
                targets = await resp.json()
        
        if not targets:
            raise Exception("No Chrome targets found. Is Chrome running with --remote-debugging-port=9222?")
        
        # Connect to first page
        ws_url = targets[0]['webSocketDebuggerUrl']
        self.ws = await websockets.connect(ws_url)
        
        # Start listening for messages
        asyncio.create_task(self._listen())
    
    async def _listen(self):
        """Listen for CDP messages"""
        async for message in self.ws:
            data = json.loads(message)
            
            if 'id' in data:
                # Response to our command
                self.responses[data['id']] = data
            else:
                # Event
                self.events.append(data)
    
    async def send_command(self, method: str, params: dict = None) -> dict:
        """Send CDP command and wait for response"""
        self.message_id += 1
        msg_id = self.message_id
        
        message = {
            "id": msg_id,
            "method": method,
            "params": params or {}
        }
        
        await self.ws.send(json.dumps(message))
        
        # Wait for response
        for _ in range(100):  # 10 second timeout
            if msg_id in self.responses:
                return self.responses.pop(msg_id)
            await asyncio.sleep(0.1)
        
        raise TimeoutError(f"No response for {method}")
    
    async def navigate(self, url: str) -> dict:
        """Navigate to URL"""
        result = await self.send_command("Page.navigate", {"url": url})
        await self.send_command("Page.loadEventFired")
        return result
    
    async def get_interactive_elements(self) -> list:
        """Get all interactive elements with their positions"""
        script = """
        (() => {
            const elements = [];
            const selectors = [
                'a', 'button', 'input', 'textarea', 'select',
                '[onclick]', '[role="button"]', '[role="link"]',
                '[tabindex]', '[contenteditable]'
            ];
            
            selectors.forEach(selector => {
                document.querySelectorAll(selector).forEach((el, idx) => {
                    const rect = el.getBoundingClientRect();
                    if (rect.width > 0 && rect.height > 0) {
                        elements.push({
                            tag: el.tagName.toLowerCase(),
                            type: el.type || null,
                            text: el.innerText?.substring(0, 50) || el.value || el.placeholder || '',
                            x: rect.left + rect.width / 2,
                            y: rect.top + rect.height / 2,
                            width: rect.width,
                            height: rect.height,
                            selector: `${el.tagName.toLowerCase()}${el.id ? '#' + el.id : ''}${el.className ? '.' + el.className.split(' ')[0] : ''}`
                        });
                    }
                });
            });
            
            return elements;
        })()
        """;
        
        result = await self.send_command("Runtime.evaluate", {
            "expression": script,
            "returnByValue": True
        })
        
        return result['result'].get('value', [])
    
    async def screenshot(self, full_page: bool = False, with_overlays: bool = False) -> tuple:
        """Take screenshot"""
        if full_page:
            # Get content size
            layout = await self.send_command("Page.getLayoutMetrics")
            width = layout['result']['contentSize']['width']
            height = layout['result']['contentSize']['height']
            
            # Set viewport
            await self.send_command("Emulation.setDeviceMetricsOverride", {
                "width": width,
                "height": height,
                "deviceScaleFactor": 1,
                "mobile": False
            })
        
        result = await self.send_command("Page.captureScreenshot")
        screenshot_b64 = result['result']['data']
        screenshot_bytes = base64.b64decode(screenshot_b64)
        
        elements = []
        if with_overlays:
            # Get interactive elements
            elements = await self.get_interactive_elements()
            
            # Add numbered overlays
            screenshot_bytes = self._add_overlays(screenshot_bytes, elements)
        
        return screenshot_bytes, elements
    
    async def get_dom(self) -> dict:
        """Get DOM tree"""
        result = await self.send_command("DOM.getDocument")
        return result['result']
    
    async def query_selector(self, selector: str) -> Optional[int]:
        """Find element by CSS selector"""
        doc = await self.get_dom()
        root_id = doc['root']['nodeId']
        
        result = await self.send_command("DOM.querySelector", {
            "nodeId": root_id,
            "selector": selector
        })
        
        return result['result'].get('nodeId')
    
    async def click(self, selector: str) -> dict:
        """Click element"""
        node_id = await self.query_selector(selector)
        if not node_id:
            return {"success": False, "error": f"Element not found: {selector}"}
        
        # Get box model
        result = await self.send_command("DOM.getBoxModel", {"nodeId": node_id})
        box = result['result']['model']['content']
        
        # Calculate center
        x = (box[0] + box[2]) / 2
        y = (box[1] + box[5]) / 2
        
        # Click
        await self.send_command("Input.dispatchMouseEvent", {
            "type": "mousePressed",
            "x": x,
            "y": y,
            "button": "left",
            "clickCount": 1
        })
        await self.send_command("Input.dispatchMouseEvent", {
            "type": "mouseReleased",
            "x": x,
            "y": y,
            "button": "left",
            "clickCount": 1
        })
        
        return {"success": True, "selector": selector}
    
    async def type_text(self, selector: str, text: str) -> dict:
        """Type text into input"""
        # Focus element
        node_id = await self.query_selector(selector)
        if not node_id:
            return {"success": False, "error": f"Element not found: {selector}"}
        
        await self.send_command("DOM.focus", {"nodeId": node_id})
        
        # Type each character
        for char in text:
            await self.send_command("Input.dispatchKeyEvent", {
                "type": "char",
                "text": char
            })
        
        return {"success": True, "selector": selector, "text": text}
    
    async def execute_script(self, script: str) -> dict:
        """Execute JavaScript"""
        result = await self.send_command("Runtime.evaluate", {
            "expression": script,
            "returnByValue": True
        })
        
        if 'exceptionDetails' in result['result']:
            return {
                "success": False,
                "error": result['result']['exceptionDetails']['text']
            }
        
        return {
            "success": True,
            "result": result['result'].get('value')
        }
    
    async def get_content(self) -> str:
        """Get page HTML"""
        doc = await self.get_dom()
        root_id = doc['root']['nodeId']
        
        result = await self.send_command("DOM.getOuterHTML", {"nodeId": root_id})
        return result['result']['outerHTML']
    
    async def get_console_messages(self) -> list:
        """Get console messages"""
        # Filter console events
        console_events = [
            e for e in self.events 
            if e.get('method') == 'Runtime.consoleAPICalled'
        ]
        
        messages = []
        for event in console_events:
            params = event['params']
            messages.append({
                "type": params['type'],
                "args": [arg.get('value') for arg in params['args']]
            })
        
        return messages
    
    async def get_network_activity(self) -> list:
        """Get network requests"""
        network_events = [
            e for e in self.events
            if e.get('method', '').startswith('Network.')
        ]
        
        return network_events
    
    def _add_overlays(self, screenshot_bytes: bytes, elements: list) -> bytes:
        """Add numbered overlays to screenshot"""
        # Load image
        image = Image.open(io.BytesIO(screenshot_bytes))
        draw = ImageDraw.Draw(image, 'RGBA')
        
        # Try to load a font, fallback to default
        try:
            font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 16)
        except:
            font = ImageFont.load_default()
        
        # Draw numbered circles on each element
        for idx, el in enumerate(elements, 1):
            x, y = int(el['x']), int(el['y'])
            
            # Draw blue circle
            radius = 15
            draw.ellipse(
                [(x - radius, y - radius), (x + radius, y + radius)],
                fill=(0, 123, 255, 200),  # Blue with transparency
                outline=(255, 255, 255, 255),  # White border
                width=2
            )
            
            # Draw number
            text = str(idx)
            # Get text bounding box for centering
            bbox = draw.textbbox((0, 0), text, font=font)
            text_width = bbox[2] - bbox[0]
            text_height = bbox[3] - bbox[1]
            
            draw.text(
                (x - text_width // 2, y - text_height // 2),
                text,
                fill=(255, 255, 255, 255),  # White text
                font=font
            )
        
        # Convert back to bytes
        output = io.BytesIO()
        image.save(output, format='PNG')
        return output.getvalue()
    
    async def close(self):
        """Close connection"""
        if self.ws:
            await self.ws.close()


class TerminalBrowserServer:
    """
    Unified MCP server for terminal and browser operations.
    Uses Chrome DevTools Protocol for browser automation.
    Uses Gemma 3 vision for screenshot analysis.
    """
    
    def __init__(self):
        self.server = Server("terminal-browser")
        self.cdp: Optional[ChromeDevTools] = None
        self.last_elements = []  # Store elements from last screenshot
        
        # Register tools
        self._register_tools()
    
    def _register_tools(self):
        """Register all MCP tools"""
        
        @self.server.list_tools()
        async def list_tools() -> list[types.Tool]:
            return [
                # Terminal Tools
                types.Tool(
                    name="natural_to_command",
                    description="Convert natural language to Linux command using specialized model",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "instruction": {
                                "type": "string",
                                "description": "Natural language instruction"
                            },
                            "context": {
                                "type": "string",
                                "description": "Optional context (current directory, etc.)"
                            }
                        },
                        "required": ["instruction"]
                    }
                ),
                types.Tool(
                    name="execute_command",
                    description="Execute Linux command with safety validation",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "command": {
                                "type": "string",
                                "description": "Command to execute"
                            },
                            "dry_run": {
                                "type": "boolean",
                                "description": "If true, validate but don't execute"
                            },
                            "timeout": {
                                "type": "integer",
                                "description": "Timeout in seconds (default: 30)"
                            }
                        },
                        "required": ["command"]
                    }
                ),
                types.Tool(
                    name="explain_command",
                    description="Explain what a Linux command does",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "command": {
                                "type": "string",
                                "description": "Command to explain"
                            }
                        },
                        "required": ["command"]
                    }
                ),
                
                # Browser Tools (Chrome DevTools)
                types.Tool(
                    name="browser_navigate",
                    description="Navigate browser to URL using Chrome DevTools",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "url": {
                                "type": "string",
                                "description": "URL to navigate to"
                            }
                        },
                        "required": ["url"]
                    }
                ),
                types.Tool(
                    name="browser_screenshot",
                    description="Take screenshot with numbered overlays on interactive elements. Gemma 3 can reference elements by number.",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "analyze": {
                                "type": "boolean",
                                "description": "Use Gemma 3 vision to analyze screenshot"
                            },
                            "question": {
                                "type": "string",
                                "description": "Question to ask about screenshot (if analyzing)"
                            },
                            "full_page": {
                                "type": "boolean",
                                "description": "Capture full page (default: false)"
                            },
                            "with_overlays": {
                                "type": "boolean",
                                "description": "Add numbered overlays to interactive elements (default: true)"
                            }
                        }
                    }
                ),
                types.Tool(
                    name="browser_get_dom",
                    description="Get DOM tree structure via Chrome DevTools",
                    inputSchema={
                        "type": "object",
                        "properties": {}
                    }
                ),
                types.Tool(
                    name="browser_click",
                    description="Click element on page. Use element number from screenshot or CSS selector.",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "selector": {
                                "type": "string",
                                "description": "CSS selector or element number (e.g., 'element:7')"
                            }
                        },
                        "required": ["selector"]
                    }
                ),
                types.Tool(
                    name="browser_type",
                    description="Type text into input field. Use element number from screenshot or CSS selector.",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "selector": {
                                "type": "string",
                                "description": "CSS selector or element number (e.g., 'element:7')"
                            },
                            "text": {
                                "type": "string",
                                "description": "Text to type"
                            }
                        },
                        "required": ["selector", "text"]
                    }
                ),
                types.Tool(
                    name="browser_execute_script",
                    description="Execute JavaScript via Chrome DevTools",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "script": {
                                "type": "string",
                                "description": "JavaScript code to execute"
                            }
                        },
                        "required": ["script"]
                    }
                ),
                types.Tool(
                    name="browser_get_content",
                    description="Get page HTML content via Chrome DevTools",
                    inputSchema={
                        "type": "object",
                        "properties": {}
                    }
                ),
                types.Tool(
                    name="browser_get_console",
                    description="Get console messages via Chrome DevTools",
                    inputSchema={
                        "type": "object",
                        "properties": {}
                    }
                ),
                types.Tool(
                    name="browser_get_network",
                    description="Get network activity via Chrome DevTools",
                    inputSchema={
                        "type": "object",
                        "properties": {}
                    }
                ),
            ]
        
        @self.server.call_tool()
        async def call_tool(name: str, arguments: dict) -> list[types.TextContent]:
            """Route tool calls to appropriate handlers"""
            
            # Terminal tools
            if name == "natural_to_command":
                result = await self.natural_to_command(
                    arguments["instruction"],
                    arguments.get("context")
                )
            elif name == "execute_command":
                result = await self.execute_command(
                    arguments["command"],
                    arguments.get("dry_run", False),
                    arguments.get("timeout", 30)
                )
            elif name == "explain_command":
                result = await self.explain_command(arguments["command"])
            
            # Browser tools
            elif name == "browser_navigate":
                result = await self.browser_navigate(arguments["url"])
            elif name == "browser_screenshot":
                result = await self.browser_screenshot(
                    arguments.get("analyze", False),
                    arguments.get("question"),
                    arguments.get("full_page", False),
                    arguments.get("with_overlays", True)
                )
            elif name == "browser_get_dom":
                result = await self.browser_get_dom()
            elif name == "browser_click":
                result = await self.browser_click(arguments["selector"])
            elif name == "browser_type":
                result = await self.browser_type(
                    arguments["selector"],
                    arguments["text"]
                )
            elif name == "browser_execute_script":
                result = await self.browser_execute_script(arguments["script"])
            elif name == "browser_get_content":
                result = await self.browser_get_content()
            elif name == "browser_get_console":
                result = await self.browser_get_console()
            elif name == "browser_get_network":
                result = await self.browser_get_network()
            else:
                raise ValueError(f"Unknown tool: {name}")
            
            return [types.TextContent(type="text", text=json.dumps(result, indent=2))]
    
    # ========== Terminal Tools ==========
    
    async def natural_to_command(self, instruction: str, context: Optional[str] = None) -> dict:
        """Convert natural language to Linux command"""
        
        prompt = f"Convert this instruction to a Linux command:\n{instruction}"
        if context:
            prompt += f"\n\nContext: {context}"
        
        prompt += "\n\nRespond with ONLY the command, no explanation."
        
        response = ollama.chat(
            model=TERMINAL_MODEL,
            messages=[{"role": "user", "content": prompt}]
        )
        
        command = response['message']['content'].strip()
        
        # Safety check
        safe, reason = self._is_command_safe(command)
        
        return {
            "command": command,
            "safe": safe,
            "reason": reason if not safe else None,
            "explanation": await self.explain_command(command) if safe else None
        }
    
    async def execute_command(self, command: str, dry_run: bool = False, timeout: int = 30) -> dict:
        """Execute Linux command with safety validation"""
        
        # Safety check
        safe, reason = self._is_command_safe(command)
        if not safe:
            return {
                "success": False,
                "error": f"Unsafe command blocked: {reason}",
                "command": command
            }
        
        if dry_run:
            return {
                "success": True,
                "dry_run": True,
                "command": command,
                "message": "Command validated (not executed)"
            }
        
        try:
            result = subprocess.run(
                command,
                shell=True,
                capture_output=True,
                text=True,
                timeout=timeout
            )
            
            return {
                "success": result.returncode == 0,
                "exit_code": result.returncode,
                "stdout": result.stdout,
                "stderr": result.stderr,
                "command": command
            }
        except subprocess.TimeoutExpired:
            return {
                "success": False,
                "error": f"Command timed out after {timeout}s",
                "command": command
            }
        except Exception as e:
            return {
                "success": False,
                "error": str(e),
                "command": command
            }
    
    async def explain_command(self, command: str) -> str:
        """Explain what a Linux command does"""
        
        prompt = f"Explain what this Linux command does:\n{command}\n\nBe concise and clear."
        
        response = ollama.chat(
            model=TERMINAL_MODEL,
            messages=[{"role": "user", "content": prompt}]
        )
        
        return response['message']['content']
    
    def _is_command_safe(self, command: str) -> tuple[bool, Optional[str]]:
        """Check if command is safe to execute"""
        import re
        
        for pattern in DANGEROUS_PATTERNS:
            if re.search(pattern, command):
                return False, f"Matches dangerous pattern: {pattern}"
        
        return True, None
    
    # ========== Browser Tools (Chrome DevTools) ==========
    
    async def _ensure_cdp(self):
        """Ensure Chrome DevTools connection"""
        if not self.cdp:
            self.cdp = ChromeDevTools()
            await self.cdp.connect()
    
    async def browser_navigate(self, url: str) -> dict:
        """Navigate to URL"""
        await self._ensure_cdp()
        
        try:
            await self.cdp.navigate(url)
            return {
                "success": True,
                "url": url
            }
        except Exception as e:
            return {
                "success": False,
                "error": str(e)
            }
    
    async def browser_screenshot(self, analyze: bool = False, question: Optional[str] = None, full_page: bool = False, with_overlays: bool = True) -> dict:
        """Take screenshot and optionally analyze with Gemma 3 vision"""
        await self._ensure_cdp()
        
        try:
            # Take screenshot with optional overlays
            screenshot_bytes, elements = await self.cdp.screenshot(full_page=full_page, with_overlays=with_overlays)
            
            # Store elements for later reference
            self.last_elements = elements
            
            # Save screenshot
            screenshot_path = "/tmp/browser_screenshot.png"
            with open(screenshot_path, "wb") as f:
                f.write(screenshot_bytes)
            
            result = {
                "success": True,
                "screenshot_path": screenshot_path,
                "elements": [
                    {
                        "number": idx,
                        "tag": el['tag'],
                        "type": el.get('type'),
                        "text": el['text'],
                        "selector": el['selector']
                    }
                    for idx, el in enumerate(elements, 1)
                ] if with_overlays else []
            }
            
            # Analyze with Gemma 3 vision if requested
            if analyze:
                screenshot_b64 = base64.b64encode(screenshot_bytes).decode()
                
                if question:
                    prompt = question
                else:
                    prompt = "Describe what you see on this webpage. Include key elements, layout, and any important information."
                
                response = ollama.chat(
                    model=VISION_MODEL,
                    messages=[{
                        "role": "user",
                        "content": prompt,
                        "images": [screenshot_b64]
                    }]
                )
                
                result["analysis"] = response['message']['content']
            
            return result
        except Exception as e:
            return {
                "success": False,
                "error": str(e)
            }
    
    async def browser_get_dom(self) -> dict:
        """Get DOM tree"""
        await self._ensure_cdp()
        
        try:
            dom = await self.cdp.get_dom()
            return {
                "success": True,
                "dom": dom
            }
        except Exception as e:
            return {
                "success": False,
                "error": str(e)
            }
    
    async def browser_click(self, selector: str) -> dict:
        """Click element by selector or element number"""
        await self._ensure_cdp()
        
        # Check if selector is element number (e.g., "element:7")
        if selector.startswith("element:"):
            try:
                element_num = int(selector.split(":")[1])
                if hasattr(self, 'last_elements') and 0 < element_num <= len(self.last_elements):
                    element = self.last_elements[element_num - 1]
                    selector = element['selector']
                else:
                    return {
                        "success": False,
                        "error": f"Element number {element_num} not found. Take a screenshot first."
                    }
            except (ValueError, IndexError):
                return {
                    "success": False,
                    "error": f"Invalid element number format: {selector}"
                }
        
        return await self.cdp.click(selector)
    
    async def browser_type(self, selector: str, text: str) -> dict:
        """Type text into input by selector or element number"""
        await self._ensure_cdp()
        
        # Check if selector is element number (e.g., "element:7")
        if selector.startswith("element:"):
            try:
                element_num = int(selector.split(":")[1])
                if hasattr(self, 'last_elements') and 0 < element_num <= len(self.last_elements):
                    element = self.last_elements[element_num - 1]
                    selector = element['selector']
                else:
                    return {
                        "success": False,
                        "error": f"Element number {element_num} not found. Take a screenshot first."
                    }
            except (ValueError, IndexError):
                return {
                    "success": False,
                    "error": f"Invalid element number format: {selector}"
                }
        
        return await self.cdp.type_text(selector, text)
    
    async def browser_execute_script(self, script: str) -> dict:
        """Execute JavaScript"""
        await self._ensure_cdp()
        return await self.cdp.execute_script(script)
    
    async def browser_get_content(self) -> dict:
        """Get page HTML"""
        await self._ensure_cdp()
        
        try:
            content = await self.cdp.get_content()
            return {
                "success": True,
                "content": content
            }
        except Exception as e:
            return {
                "success": False,
                "error": str(e)
            }
    
    async def browser_get_console(self) -> dict:
        """Get console messages"""
        await self._ensure_cdp()
        
        try:
            messages = await self.cdp.get_console_messages()
            return {
                "success": True,
                "messages": messages
            }
        except Exception as e:
            return {
                "success": False,
                "error": str(e)
            }
    
    async def browser_get_network(self) -> dict:
        """Get network activity"""
        await self._ensure_cdp()
        
        try:
            network = await self.cdp.get_network_activity()
            return {
                "success": True,
                "network": network
            }
        except Exception as e:
            return {
                "success": False,
                "error": str(e)
            }
    
    async def cleanup(self):
        """Cleanup resources"""
        if self.cdp:
            await self.cdp.close()

async def main():
    """Run the MCP server"""
    server_instance = TerminalBrowserServer()
    
    try:
        async with stdio_server() as (read_stream, write_stream):
            await server_instance.server.run(
                read_stream,
                write_stream,
                server_instance.server.create_initialization_options()
            )
    finally:
        await server_instance.cleanup()

if __name__ == "__main__":
    asyncio.run(main())


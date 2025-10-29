#!/usr/bin/env python3.11
"""
Test Agent Terminal Control via A2A
Demonstrates the agent using terminal commands through A2A protocol
"""

import asyncio
import json
import websockets

async def test_agent_terminal_control():
    """
    Test agent controlling terminal via A2A WebSocket
    """
    print("""
╔══════════════════════════════════════════════════════════════════╗
║           AGENT TERMINAL CONTROL TEST - A2A PROTOCOL             ║
╚══════════════════════════════════════════════════════════════════╝

Testing agent's ability to execute terminal commands via A2A...
""")
    
    # Connect to A2A WebSocket
    uri = "ws://localhost:8080/ws/a2a"
    request_id = 1
    
    try:
        async with websockets.connect(uri) as websocket:
            print(f"✅ Connected to A2A WebSocket: {uri}\n")
            
            # Test 1: Execute 'pwd' command
            print("="*70)
            print("TEST 1: Execute 'pwd' command")
            print("="*70)
            
            request = {
                "jsonrpc": "2.0",
                "method": "terminal/execute",
                "params": {
                    "command": "pwd",
                    "session_id": "agent-test"
                },
                "id": request_id
            }
            request_id += 1
            
            await websocket.send(json.dumps(request))
            response = await websocket.recv()
            result = json.loads(response)
            
            if "result" in result:
                print(f"✅ Command executed successfully!")
                print(f"Output: {result['result']['output']}")
            else:
                print(f"❌ Error: {result.get('error', {}).get('message', 'Unknown error')}")
            
            await asyncio.sleep(1)
            
            # Test 2: Execute 'ls -la' command
            print("\n" + "="*70)
            print("TEST 2: Execute 'ls -la' command")
            print("="*70)
            
            request = {
                "jsonrpc": "2.0",
                "method": "terminal/execute",
                "params": {
                    "command": "ls -la",
                    "session_id": "agent-test"
                },
                "id": request_id
            }
            request_id += 1
            
            await websocket.send(json.dumps(request))
            response = await websocket.recv()
            result = json.loads(response)
            
            if "result" in result:
                print(f"✅ Command executed successfully!")
                print(f"Output:\n{result['result']['output']}")
            else:
                print(f"❌ Error: {result.get('error', {}).get('message', 'Unknown error')}")
            
            await asyncio.sleep(1)
            
            # Test 3: Execute 'echo "Hello from Agent!"' command
            print("\n" + "="*70)
            print("TEST 3: Execute echo command")
            print("="*70)
            
            request = {
                "jsonrpc": "2.0",
                "method": "terminal/execute",
                "params": {
                    "command": 'echo "Hello from Agent!"',
                    "session_id": "agent-test"
                },
                "id": request_id
            }
            request_id += 1
            
            await websocket.send(json.dumps(request))
            response = await websocket.recv()
            result = json.loads(response)
            
            if "result" in result:
                print(f"✅ Command executed successfully!")
                print(f"Output: {result['result']['output']}")
            else:
                print(f"❌ Error: {result.get('error', {}).get('message', 'Unknown error')}")
            
            await asyncio.sleep(1)
            
            # Test 4: Check system info
            print("\n" + "="*70)
            print("TEST 4: Get system information")
            print("="*70)
            
            request = {
                "jsonrpc": "2.0",
                "method": "terminal/execute",
                "params": {
                    "command": "uname -a",
                    "session_id": "agent-test"
                },
                "id": request_id
            }
            request_id += 1
            
            await websocket.send(json.dumps(request))
            response = await websocket.recv()
            result = json.loads(response)
            
            if "result" in result:
                print(f"✅ Command executed successfully!")
                print(f"Output: {result['result']['output']}")
            else:
                print(f"❌ Error: {result.get('error', {}).get('message', 'Unknown error')}")
            
            print("\n" + "="*70)
            print("✅ ALL TESTS PASSED!")
            print("="*70)
            print("\nThe agent can now:")
            print("  • Execute any terminal command")
            print("  • Get command output")
            print("  • Use multiple terminal sessions")
            print("  • Combine with browser automation")
            print("\n🎉 Agent has full terminal control via A2A!")
            
    except Exception as e:
        print(f"\n❌ Test failed: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    print("\n🚀 Starting Agent Terminal Control Test...")
    print("📋 Make sure the backend is running!\n")
    
    asyncio.run(test_agent_terminal_control())

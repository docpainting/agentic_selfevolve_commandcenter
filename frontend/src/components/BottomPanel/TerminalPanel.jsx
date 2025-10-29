import { useRef, useEffect, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { SearchAddon } from '@xterm/addon-search';
import '@xterm/xterm/css/xterm.css';

export default function TerminalPanel({ takeoverMode }) {
  const terminalRef = useRef(null);
  const xtermRef = useRef(null);
  const wsRef = useRef(null);
  const fitAddonRef = useRef(null);
  const requestIdRef = useRef(1);
  const currentLineRef = useRef('');
  const cwdRef = useRef('~');
  const [isReady, setIsReady] = useState(false);

  // Initialize xterm.js
  useEffect(() => {
    if (!terminalRef.current || xtermRef.current) return;

    // Small delay to ensure DOM is ready
    const timer = setTimeout(() => {
      try {
        console.log('[TerminalPanel] Creating Terminal instance...');
        // Create terminal instance
        const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#0a0e1a',
        foreground: '#e0e0e0',
        cursor: '#00d9ff',
        black: '#000000',
        red: '#ff5555',
        green: '#50fa7b',
        yellow: '#f1fa8c',
        blue: '#bd93f9',
        magenta: '#ff79c6',
        cyan: '#8be9fd',
        white: '#bfbfbf',
        brightBlack: '#4d4d4d',
        brightRed: '#ff6e67',
        brightGreen: '#5af78e',
        brightYellow: '#f4f99d',
        brightBlue: '#caa9fa',
        brightMagenta: '#ff92d0',
        brightCyan: '#9aedfe',
        brightWhite: '#e6e6e6'
      },
      scrollback: 10000,
      allowProposedApi: true
    });

    // Add addons
    const fitAddon = new FitAddon();
    const webLinksAddon = new WebLinksAddon();
    const searchAddon = new SearchAddon();
    
    term.loadAddon(fitAddon);
    term.loadAddon(webLinksAddon);
    term.loadAddon(searchAddon);

    // Open terminal
    term.open(terminalRef.current);
      
    // Fit after a small delay to ensure rendering is complete
    setTimeout(() => {
      fitAddon.fit();
      setIsReady(true);
    }, 100);

    // Store refs
    xtermRef.current = term;
    fitAddonRef.current = fitAddon;

    // Welcome message
    term.writeln('\x1b[1;36m╔══════════════════════════════════════════════════════════════╗\x1b[0m');
    term.writeln('\x1b[1;36m║         Agent Workspace Terminal - A2A Protocol              ║\x1b[0m');
    term.writeln('\x1b[1;36m╚══════════════════════════════════════════════════════════════╝\x1b[0m');
    term.writeln('');
    term.writeln('\x1b[33m🔌 Connecting to backend...\x1b[0m');
    term.writeln('');

    // Handle resize
    const handleResize = () => {
      if (fitAddonRef.current) {
        fitAddonRef.current.fit();
      }
    };
    window.addEventListener('resize', handleResize);

    // Connect to A2A WebSocket
    const websocket = new WebSocket('ws://localhost:8080/ws/a2a');
    
    websocket.onopen = () => {
      term.writeln('\x1b[32m✅ Connected to A2A WebSocket\x1b[0m');
      term.writeln('');
      showPrompt(term);
    };
    
    websocket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      if (data.jsonrpc === '2.0' && data.result) {
        const output = data.result.output || '';
        const success = data.result.success;
        
        // Write output to terminal
        if (output.trim()) {
          term.write('\r\n' + output);
        }
        
        // Show new prompt
        term.write('\r\n');
        showPrompt(term);
      } else if (data.error) {
        term.writeln('\r\n\x1b[31m❌ Error: ' + data.error.message + '\x1b[0m');
        term.write('\r\n');
        showPrompt(term);
      }
    };
    
    websocket.onerror = (error) => {
      term.writeln('\r\n\x1b[31m❌ WebSocket error\x1b[0m');
    };
    
    websocket.onclose = (event) => {
      term.writeln('\r\n\x1b[33m🔌 Disconnected (code: ' + event.code + ')\x1b[0m');
    };
    
    wsRef.current = websocket;

    // Handle terminal input
    term.onData(data => {
      const ws = wsRef.current;
      if (!ws || ws.readyState !== WebSocket.OPEN) return;

      const code = data.charCodeAt(0);

      // Handle Enter key
      if (code === 13) {
        const command = currentLineRef.current.trim();
        if (command) {
          term.write('\r\n');
          
          // Handle special commands
          if (command === 'clear' || command === 'cls') {
            term.clear();
            currentLineRef.current = '';
            showPrompt(term);
            return;
          }

          // Send command via A2A
          ws.send(JSON.stringify({
            jsonrpc: '2.0',
            method: 'terminal/execute',
            params: {
              command: command,
              session_id: 'default'
            },
            id: requestIdRef.current++
          }));
          
          currentLineRef.current = '';
        } else {
          term.write('\r\n');
          showPrompt(term);
        }
      }
      // Handle Backspace
      else if (code === 127) {
        if (currentLineRef.current.length > 0) {
          currentLineRef.current = currentLineRef.current.slice(0, -1);
          term.write('\b \b');
        }
      }
      // Handle Ctrl+C
      else if (code === 3) {
        term.write('^C\r\n');
        currentLineRef.current = '';
        showPrompt(term);
      }
      // Handle Ctrl+L (clear)
      else if (code === 12) {
        term.clear();
        currentLineRef.current = '';
        showPrompt(term);
      }
      // Regular character
      else if (code >= 32) {
        currentLineRef.current += data;
        term.write(data);
      }
    });

        return () => {
          window.removeEventListener('resize', handleResize);
          websocket.close();
          term.dispose();
        };
      } catch (error) {
        console.error('[TerminalPanel] Error initializing terminal:', error);
        console.error('[TerminalPanel] Error stack:', error.stack);
      }
    }, 50);

    return () => clearTimeout(timer);
  }, []);

  const showPrompt = (term) => {
    const cwd = cwdRef.current;
    term.write('\x1b[36m' + cwd + '\x1b[0m \x1b[90m$\x1b[0m ');
  };

  return (
    <div className="flex flex-col h-full bg-midnight-950/30 backdrop-blur-sm">
      <div ref={terminalRef} className="flex-1 w-full" style={{ minHeight: '400px' }} />
      
      {/* Status Bar */}
      <div className="border-t border-white/10 px-4 py-2 flex items-center justify-between text-xs">
        <span className="text-white/40">
          {takeoverMode ? '👤 Manual Control' : '🤖 AI Control'}
        </span>
        <span className="text-cyan-500">
          xterm.js • A2A Protocol
        </span>
      </div>
    </div>
  );
}


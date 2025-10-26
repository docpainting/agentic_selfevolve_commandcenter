import { useState, useRef, useEffect } from 'react';

const mockOutput = [
  { type: 'ai', command: 'neo4j start', output: 'Starting Neo4j...\nNeo4j started successfully.' },
  { type: 'system', output: '✓ Neo4j started' },
  { type: 'ai', command: 'lightrag init', output: 'Initializing LightRAG...' },
  { type: 'system', output: '✓ LightRAG initialized' },
  { type: 'system', output: '✓ Watchdog monitoring active' },
  { type: 'system', output: '🚀 Agent workspace ready!' },
];

export default function TerminalPanel({ takeoverMode }) {
  const [output, setOutput] = useState(mockOutput);
  const [input, setInput] = useState('');
  const terminalRef = useRef(null);

  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [output]);

  const handleCommand = (e) => {
    if (e.key === 'Enter') {
      const newOutput = [
        ...output,
        { type: 'user', command: input, output: '' },
      ];
      setOutput(newOutput);
      setInput('');

      // Simulate command execution
      setTimeout(() => {
        setOutput([
          ...newOutput,
          { type: 'system', output: `Executed: ${input}` },
        ]);
      }, 500);
    }
  };

  return (
    <div className="flex flex-col h-full bg-midnight-950/50 font-mono text-sm">
      {/* Output Area */}
      <div ref={terminalRef} className="flex-1 overflow-y-auto p-4 space-y-2">
        {output.map((line, index) => (
          <div key={index}>
            {line.command && (
              <div className="flex items-center gap-2 mb-1">
                <span className="text-white/40">$</span>
                <span className={line.type === 'ai' ? 'terminal-command-ai' : 'terminal-command-user'}>
                  {line.command}
                </span>
              </div>
            )}
            {line.output && (
              <div className="terminal-output pl-4">
                {line.output}
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Input Area (only in takeover mode) */}
      {takeoverMode && (
        <div className="border-t border-white/10 p-3 flex items-center gap-2">
          <span className="text-cyan-500">$</span>
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={handleCommand}
            placeholder="Enter command..."
            className="flex-1 bg-transparent outline-none text-white"
            autoFocus
          />
        </div>
      )}

      {/* Mode Indicator */}
      <div className="border-t border-white/10 px-4 py-2 flex items-center justify-between text-xs">
        <span className="text-white/40">
          {takeoverMode ? '👤 Manual Control' : '🤖 AI Control'}
        </span>
        <span className="text-white/40">
          Press Ctrl+Shift+T to toggle
        </span>
      </div>
    </div>
  );
}


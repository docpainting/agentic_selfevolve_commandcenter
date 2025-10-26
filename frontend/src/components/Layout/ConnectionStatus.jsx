export default function ConnectionStatus({ wsConnected, a2aConnected, agentState }) {
  const getStateColor = (state) => {
    switch (state) {
      case 'idle': return 'text-white/60';
      case 'thinking': return 'text-cyan-500';
      case 'executing': return 'text-cyan-500 animate-pulse';
      case 'awaiting_approval': return 'text-yellow-500';
      case 'error': return 'text-red-500';
      default: return 'text-white/60';
    }
  };

  return (
    <div className="h-8 glass-panel mx-2 mt-2 rounded-lg flex items-center justify-between px-4 text-xs">
      {/* Left: Connection Status */}
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <div className={wsConnected ? 'status-connected' : 'status-disconnected'} />
          <span className="text-white/60">WebSocket</span>
        </div>
        <div className="flex items-center gap-2">
          <div className={a2aConnected ? 'status-connected' : 'status-disconnected'} />
          <span className="text-white/60">A2A</span>
        </div>
      </div>

      {/* Center: Agent State */}
      <div className={`font-medium ${getStateColor(agentState)}`}>
        {agentState === 'idle' && '● Idle'}
        {agentState === 'thinking' && '● Thinking...'}
        {agentState === 'executing' && '● Executing...'}
        {agentState === 'awaiting_approval' && '⚠ Awaiting Approval'}
        {agentState === 'error' && '✗ Error'}
      </div>

      {/* Right: System Info */}
      <div className="flex items-center gap-4 text-white/60">
        <span>Neo4j: Connected</span>
        <span>Ollama: gemma3:27b</span>
      </div>
    </div>
  );
}


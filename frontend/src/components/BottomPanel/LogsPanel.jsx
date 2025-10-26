import { useEffect, useRef } from 'react';
import { Info, AlertTriangle, XCircle, CheckCircle } from 'lucide-react';

const mockLogs = [
  { level: 'info', message: 'WebSocket connected to ws://localhost:8080/ws/chat', timestamp: '14:30:00' },
  { level: 'info', message: 'A2A protocol initialized', timestamp: '14:30:01' },
  { level: 'success', message: 'Neo4j connection established', timestamp: '14:30:05' },
  { level: 'info', message: 'LightRAG initialized with ChromeM vector store', timestamp: '14:30:06' },
  { level: 'success', message: 'Watchdog monitoring started', timestamp: '14:30:07' },
  { level: 'info', message: 'MCP server "dynamic-thinking" connected', timestamp: '14:30:10' },
  { level: 'warning', message: 'ChromeDP context creation took longer than expected (5.2s)', timestamp: '14:30:15' },
  { level: 'info', message: 'Agent workspace ready', timestamp: '14:30:16' },
];

export default function LogsPanel() {
  const logsRef = useRef(null);

  useEffect(() => {
    if (logsRef.current) {
      logsRef.current.scrollTop = logsRef.current.scrollHeight;
    }
  }, []);

  const getIcon = (level) => {
    switch (level) {
      case 'info':
        return <Info className="w-4 h-4 text-cyan-500" />;
      case 'success':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'warning':
        return <AlertTriangle className="w-4 h-4 text-yellow-500" />;
      case 'error':
        return <XCircle className="w-4 h-4 text-red-500" />;
      default:
        return <Info className="w-4 h-4 text-white/40" />;
    }
  };

  return (
    <div className="flex flex-col h-full bg-midnight-950/50">
      {/* Toolbar */}
      <div className="border-b border-white/10 px-4 py-2 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-xs text-white/60">System Logs</span>
          <span className="text-xs text-white/40">({mockLogs.length} entries)</span>
        </div>
        <div className="flex items-center gap-2">
          <button className="btn-glass text-xs px-3 py-1">Clear</button>
          <button className="btn-glass text-xs px-3 py-1">Export</button>
        </div>
      </div>

      {/* Logs */}
      <div ref={logsRef} className="flex-1 overflow-y-auto p-4 space-y-1 font-mono text-xs">
        {mockLogs.map((log, index) => (
          <div key={index} className="flex items-start gap-3 py-1 hover:bg-white/5 rounded px-2">
            <span className="text-white/40 flex-shrink-0">{log.timestamp}</span>
            <div className="flex-shrink-0 mt-0.5">{getIcon(log.level)}</div>
            <span className="text-white/80 flex-1">{log.message}</span>
          </div>
        ))}
      </div>
    </div>
  );
}


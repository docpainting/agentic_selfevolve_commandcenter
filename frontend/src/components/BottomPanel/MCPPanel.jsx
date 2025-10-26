import { Plug, Play, Clock } from 'lucide-react';

const mockServers = [
  {
    name: 'dynamic-thinking',
    status: 'connected',
    tools: ['perceive', 'reason', 'act', 'reflect'],
  },
  {
    name: 'playwright',
    status: 'connected',
    tools: ['navigate', 'click', 'type', 'screenshot'],
  },
];

const mockActivity = [
  { time: '14:32:15', server: 'dynamic-thinking', tool: 'perceive', status: 'completed' },
  { time: '14:32:18', server: 'dynamic-thinking', tool: 'reason', status: 'completed' },
  { time: '14:32:22', server: 'dynamic-thinking', tool: 'act', status: 'working' },
];

export default function MCPPanel() {
  return (
    <div className="flex h-full bg-midnight-950/50">
      {/* Left: Connected Servers */}
      <div className="w-1/2 border-r border-white/10 p-4 overflow-y-auto">
        <h3 className="text-sm font-semibold text-white/80 mb-3">Connected Servers</h3>
        <div className="space-y-3">
          {mockServers.map((server) => (
            <div key={server.name} className="card-glass">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Plug className="w-4 h-4 text-cyan-500" />
                  <span className="text-sm font-medium">{server.name}</span>
                </div>
                <div className="status-connected" />
              </div>
              <div className="flex flex-wrap gap-1">
                {server.tools.map((tool) => (
                  <button
                    key={tool}
                    className="btn-glass text-xs px-2 py-1"
                  >
                    <Play className="w-3 h-3 mr-1 inline" />
                    {tool}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Right: Activity Timeline */}
      <div className="w-1/2 p-4 overflow-y-auto">
        <h3 className="text-sm font-semibold text-white/80 mb-3">Activity Timeline</h3>
        <div className="space-y-2">
          {mockActivity.map((activity, index) => (
            <div key={index} className="card-glass">
              <div className="flex items-start gap-3">
                <Clock className="w-4 h-4 text-white/40 flex-shrink-0 mt-0.5" />
                <div className="flex-1">
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-xs font-mono text-white/60">{activity.time}</span>
                    <span className={`text-xs px-2 py-0.5 rounded ${
                      activity.status === 'completed'
                        ? 'bg-green-500/20 text-green-500'
                        : 'bg-cyan-500/20 text-cyan-500'
                    }`}>
                      {activity.status}
                    </span>
                  </div>
                  <p className="text-sm">
                    <span className="text-cyan-500">{activity.server}</span>
                    <span className="text-white/40 mx-1">→</span>
                    <span className="text-white">{activity.tool}</span>
                  </p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}


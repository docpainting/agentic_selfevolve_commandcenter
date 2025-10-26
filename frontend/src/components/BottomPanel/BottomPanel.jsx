import { useState } from 'react';
import { Terminal, Globe, Boxes, FileText, Hand } from 'lucide-react';
import TerminalPanel from './TerminalPanel';
import BrowserPanel from './BrowserPanel';
import MCPPanel from './MCPPanel';
import LogsPanel from './LogsPanel';

export default function BottomPanel() {
  const [activeTab, setActiveTab] = useState('terminal');
  const [takeoverMode, setTakeoverMode] = useState(false);

  const tabs = [
    { id: 'terminal', label: 'Terminal', icon: Terminal },
    { id: 'browser', label: 'Browser', icon: Globe },
    { id: 'mcp', label: 'MCP Tools', icon: Boxes },
    { id: 'logs', label: 'Logs', icon: FileText },
  ];

  const renderContent = () => {
    switch (activeTab) {
      case 'terminal':
        return <TerminalPanel takeoverMode={takeoverMode} />;
      case 'browser':
        return <BrowserPanel takeoverMode={takeoverMode} />;
      case 'mcp':
        return <MCPPanel />;
      case 'logs':
        return <LogsPanel />;
      default:
        return null;
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Tab Bar */}
      <div className="flex items-center justify-between border-b border-white/10 px-4">
        <div className="flex">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 px-4 py-3 border-b-2 transition-all ${
                  activeTab === tab.id
                    ? 'border-cyan-500 text-cyan-500'
                    : 'border-transparent text-white/60 hover:text-white/80'
                }`}
              >
                <Icon className="w-4 h-4" />
                <span className="text-sm font-medium">{tab.label}</span>
              </button>
            );
          })}
        </div>

        {/* Takeover Button (only for terminal and browser) */}
        {(activeTab === 'terminal' || activeTab === 'browser') && (
          <button
            onClick={() => setTakeoverMode(!takeoverMode)}
            className={takeoverMode ? 'takeover-button-active' : 'takeover-button'}
          >
            <Hand className="w-3 h-3 mr-1 inline" />
            {takeoverMode ? 'AI Control' : 'Manual Control'}
          </button>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        {renderContent()}
      </div>
    </div>
  );
}


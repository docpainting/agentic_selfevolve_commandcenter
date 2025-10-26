import { useState } from 'react';
import { Eye, CheckCircle, Clock, AlertTriangle, TrendingUp } from 'lucide-react';

const mockComponents = [
  { name: 'Authentication Module', status: 'approved', progress: 100 },
  { name: 'Database Layer', status: 'pending', progress: 65 },
  { name: 'API Endpoints', status: 'review', progress: 80 },
];

const mockAlerts = [
  {
    id: 1,
    type: 'info',
    title: 'Pattern Detected',
    message: 'JWT authentication pattern being implemented',
    timestamp: new Date().toISOString(),
  },
  {
    id: 2,
    type: 'warning',
    title: 'Concept Drift',
    message: 'Database connection approach differs from previous pattern',
    timestamp: new Date().toISOString(),
  },
];

export default function OpenEvolve() {
  const [selectedTab, setSelectedTab] = useState('components');

  const getStatusIcon = (status) => {
    switch (status) {
      case 'approved':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'pending':
        return <Clock className="w-4 h-4 text-cyan-500 animate-pulse" />;
      case 'review':
        return <AlertTriangle className="w-4 h-4 text-yellow-500" />;
      default:
        return null;
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-white/10">
        <div className="flex items-center gap-2 mb-3">
          <Eye className="w-5 h-5 text-cyan-500" />
          <h2 className="text-sm font-semibold text-white/80">OpenEvolve</h2>
        </div>
        
        {/* Tabs */}
        <div className="flex gap-1 text-xs">
          <button
            onClick={() => setSelectedTab('components')}
            className={selectedTab === 'components' ? 'tab-active' : 'tab-inactive'}
          >
            Components
          </button>
          <button
            onClick={() => setSelectedTab('watchdog')}
            className={selectedTab === 'watchdog' ? 'tab-active' : 'tab-inactive'}
          >
            Watchdog
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {selectedTab === 'components' && (
          <>
            {/* Overall Progress */}
            <div className="card-glass">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-white/60">Overall Progress</span>
                <span className="text-sm font-semibold text-cyan-500">82%</span>
              </div>
              <div className="progress-bar">
                <div className="progress-fill" style={{ width: '82%' }} />
              </div>
            </div>

            {/* Components List */}
            {mockComponents.map((component, index) => (
              <div key={index} className="card-glass space-y-2">
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-2">
                    {getStatusIcon(component.status)}
                    <div>
                      <p className="text-sm font-medium">{component.name}</p>
                      <p className="text-xs text-white/40 capitalize">{component.status}</p>
                    </div>
                  </div>
                </div>
                <div className="progress-bar">
                  <div className="progress-fill" style={{ width: `${component.progress}%` }} />
                </div>
              </div>
            ))}
          </>
        )}

        {selectedTab === 'watchdog' && (
          <>
            {/* Watchdog Status */}
            <div className="card-glass">
              <div className="flex items-center gap-2 mb-2">
                <div className="status-working" />
                <span className="text-sm font-medium">Monitoring Active</span>
              </div>
              <div className="flex items-center gap-2 text-xs text-white/60">
                <TrendingUp className="w-4 h-4" />
                <span>3 patterns detected today</span>
              </div>
            </div>

            {/* Alerts */}
            {mockAlerts.map((alert) => (
              <div
                key={alert.id}
                className={`watchdog-alert-${alert.type}`}
              >
                <div className="flex items-start gap-2">
                  {alert.type === 'info' && <Eye className="w-4 h-4 text-cyan-500 flex-shrink-0" />}
                  {alert.type === 'warning' && <AlertTriangle className="w-4 h-4 text-yellow-500 flex-shrink-0" />}
                  <div className="flex-1">
                    <p className="text-sm font-medium mb-1">{alert.title}</p>
                    <p className="text-xs text-white/60">{alert.message}</p>
                    <p className="text-xs text-white/40 mt-2">
                      {new Date(alert.timestamp).toLocaleTimeString()}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </>
        )}
      </div>
    </div>
  );
}


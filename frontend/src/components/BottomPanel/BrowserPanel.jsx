import { useState } from 'react';
import { Camera, MousePointer } from 'lucide-react';

const mockElements = [
  { id: 1, x: 10, y: 10, width: 100, height: 40, label: 'Search' },
  { id: 2, x: 120, y: 10, width: 80, height: 40, label: 'Sign In' },
  { id: 3, x: 10, y: 60, width: 200, height: 100, label: 'Repository' },
];

export default function BrowserPanel({ takeoverMode }) {
  const [screenshotUrl, setScreenshotUrl] = useState('/api/placeholder/800/600');
  const [showOverlays, setShowOverlays] = useState(true);
  const [lastCommand, setLastCommand] = useState('NAVIGATE https://github.com');

  return (
    <div className="flex flex-col h-full bg-midnight-950/50">
      {/* Toolbar */}
      <div className="border-b border-white/10 px-4 py-2 flex items-center justify-between">
        <div className="flex items-center gap-3 text-xs">
          <div className="flex items-center gap-2">
            <div className="status-working" />
            <span className="text-white/60">Capturing...</span>
          </div>
          <span className="text-white/40">|</span>
          <span className="text-cyan-500 font-mono">{lastCommand}</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowOverlays(!showOverlays)}
            className="btn-glass text-xs px-3 py-1"
          >
            <MousePointer className="w-3 h-3 mr-1 inline" />
            {showOverlays ? 'Hide' : 'Show'} Overlays
          </button>
          <button className="btn-glass text-xs px-3 py-1">
            <Camera className="w-3 h-3 mr-1 inline" />
            Capture
          </button>
        </div>
      </div>

      {/* Browser View */}
      <div className="flex-1 overflow-auto p-4 bg-midnight-900/30">
        <div className="relative inline-block">
          {/* Screenshot */}
          <img
            src={screenshotUrl}
            alt="Browser screenshot"
            className="rounded-lg border border-white/10"
          />

          {/* Numbered Overlays */}
          {showOverlays && mockElements.map((element) => (
            <div
              key={element.id}
              className="browser-element-overlay"
              style={{
                left: `${element.x}px`,
                top: `${element.y}px`,
                width: `${element.width}px`,
                height: `${element.height}px`,
              }}
            >
              {element.id}
            </div>
          ))}
        </div>
      </div>

      {/* Command Parser (AI mode only) */}
      {!takeoverMode && (
        <div className="border-t border-white/10 px-4 py-2 bg-cyan-500/5">
          <div className="text-xs font-mono">
            <span className="text-white/40">AI Command:</span>
            <span className="text-cyan-500 ml-2">CLICK 3</span>
            <span className="text-white/60 ml-2">→ Clicking "Repository" element</span>
          </div>
        </div>
      )}

      {/* Mode Indicator */}
      <div className="border-t border-white/10 px-4 py-2 flex items-center justify-between text-xs">
        <span className="text-white/40">
          {takeoverMode ? '👤 Manual Control' : '🤖 AI Control'}
        </span>
        <span className="text-white/40">
          Press Ctrl+Shift+B to toggle
        </span>
      </div>
    </div>
  );
}


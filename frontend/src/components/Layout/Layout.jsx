import { useState } from 'react';

export default function Layout({ leftPanel, centerPanel, rightPanel, bottomPanel }) {
  const [bottomPanelHeight, setBottomPanelHeight] = useState(30); // vh

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Top Section: 3-column layout */}
      <div 
        className="flex gap-2 p-2" 
        style={{ height: `${100 - bottomPanelHeight}vh` }}
      >
        {/* Left Panel - File Tree (20%) */}
        <div className="w-[20%] glass-panel rounded-xl overflow-hidden flex flex-col">
          {leftPanel}
        </div>

        {/* Center Panel - Chat (60%) */}
        <div className="w-[60%] glass-panel rounded-xl overflow-hidden flex flex-col">
          {centerPanel}
        </div>

        {/* Right Panel - OpenEvolve (20%) */}
        <div className="w-[20%] glass-panel rounded-xl overflow-hidden flex flex-col">
          {rightPanel}
        </div>
      </div>

      {/* Bottom Panel - Terminal/Browser/MCP/Logs */}
      <div 
        className="mx-2 mb-2 glass-panel rounded-xl overflow-hidden flex flex-col"
        style={{ height: `${bottomPanelHeight}vh` }}
      >
        {bottomPanel}
      </div>
    </div>
  );
}


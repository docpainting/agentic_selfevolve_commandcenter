import { useState, useEffect } from 'react';
import Layout from './components/Layout/Layout';
import FileTree from './components/FileTree/FileTree';
import Chat from './components/Chat/Chat';
import OpenEvolve from './components/OpenEvolve/OpenEvolve';
import BottomPanel from './components/BottomPanel/BottomPanel';
import ConnectionStatus from './components/Layout/ConnectionStatus';

function App() {
  const [wsConnected, setWsConnected] = useState(false);
  const [a2aConnected, setA2aConnected] = useState(false);
  const [agentState, setAgentState] = useState('idle');

  useEffect(() => {
    // Initialize WebSocket connections
    const wsChat = new WebSocket('ws://localhost:8080/ws/chat');
    const wsA2A = new WebSocket('ws://localhost:8080/ws/a2a');

    wsChat.onopen = () => setWsConnected(true);
    wsChat.onclose = () => setWsConnected(false);

    wsA2A.onopen = () => setA2aConnected(true);
    wsA2A.onclose = () => setA2aConnected(false);

    return () => {
      wsChat.close();
      wsA2A.close();
    };
  }, []);

  return (
    <div className="h-screen w-screen overflow-hidden flex flex-col bg-midnight-950">
      {/* Connection Status Bar */}
      <ConnectionStatus 
        wsConnected={wsConnected}
        a2aConnected={a2aConnected}
        agentState={agentState}
      />

      {/* Main Layout */}
      <Layout
        leftPanel={<FileTree />}
        centerPanel={<Chat onStateChange={setAgentState} />}
        rightPanel={<OpenEvolve />}
        bottomPanel={<BottomPanel />}
      />
    </div>
  );
}

export default App;


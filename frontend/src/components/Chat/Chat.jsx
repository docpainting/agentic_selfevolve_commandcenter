import { useState, useRef, useEffect } from 'react';
import { Send, Sparkles } from 'lucide-react';
import Message from './Message';

export default function Chat({ onStateChange }) {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [isExpanded, setIsExpanded] = useState(false);
  const messagesEndRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = () => {
    if (!input.trim()) return;

    const userMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: input,
      timestamp: new Date().toISOString(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setIsExpanded(true);

    // Simulate agent response
    setTimeout(() => {
      const agentMessage = {
        id: (Date.now() + 1).toString(),
        role: 'agent',
        content: 'I understand. Let me work on that...',
        timestamp: new Date().toISOString(),
        status: 'thinking',
      };
      setMessages(prev => [...prev, agentMessage]);
      onStateChange('thinking');
    }, 500);
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="flex-1 flex flex-col h-full">
      {/* Messages Area */}
      {isExpanded ? (
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.map((message) => (
            <Message key={message.id} message={message} />
          ))}
          <div ref={messagesEndRef} />
        </div>
      ) : (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center space-y-6 max-w-2xl px-8">
            <div className="inline-flex items-center justify-center w-20 h-20 glass-panel rounded-full lens-flare">
              <Sparkles className="w-10 h-10 text-cyan-500" />
            </div>
            <div>
              <h1 className="text-4xl font-bold text-gradient-cyan mb-2">
                Agent Workspace
              </h1>
              <p className="text-white/60 text-lg">
                Dynamic sequential thinking with browser automation and knowledge graph memory
              </p>
            </div>
            <div className="flex flex-wrap gap-2 justify-center">
              <button className="btn-glass text-sm">
                Create a login page
              </button>
              <button className="btn-glass text-sm">
                Navigate to GitHub
              </button>
              <button className="btn-glass text-sm">
                Query past patterns
              </button>
              <button className="btn-glass text-sm">
                Run tests
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Input Area */}
      <div className="p-4 border-t border-white/10">
        <div className="flex gap-2">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Ask me anything..."
            className="flex-1 input-glass resize-none h-12 py-3"
            rows="1"
          />
          <button
            onClick={handleSend}
            disabled={!input.trim()}
            className="btn-cyan px-6 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Send className="w-5 h-5" />
          </button>
        </div>
        <div className="mt-2 text-xs text-white/40 text-center">
          Press Enter to send • Shift+Enter for new line
        </div>
      </div>
    </div>
  );
}


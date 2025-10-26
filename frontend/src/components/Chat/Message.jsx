import { User, Bot, CheckCircle, Clock, XCircle } from 'lucide-react';

export default function Message({ message }) {
  const isUser = message.role === 'user';

  const getStatusIcon = (status) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'thinking':
      case 'working':
        return <Clock className="w-4 h-4 text-cyan-500 animate-spin" />;
      case 'failed':
        return <XCircle className="w-4 h-4 text-red-500" />;
      default:
        return null;
    }
  };

  return (
    <div className={`flex gap-3 animate-slide-up ${isUser ? 'flex-row-reverse' : 'flex-row'}`}>
      {/* Avatar */}
      <div className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center ${
        isUser ? 'bg-cyan-500/20 text-cyan-500' : 'bg-white/10 text-white'
      }`}>
        {isUser ? <User className="w-5 h-5" /> : <Bot className="w-5 h-5" />}
      </div>

      {/* Message Content */}
      <div className={`flex flex-col gap-1 max-w-[70%] ${isUser ? 'items-end' : 'items-start'}`}>
        <div className={isUser ? 'message-user' : 'message-agent'}>
          <p className="text-sm leading-relaxed whitespace-pre-wrap">{message.content}</p>
        </div>
        
        {/* Metadata */}
        <div className="flex items-center gap-2 text-xs text-white/40 px-2">
          {message.status && getStatusIcon(message.status)}
          <span>{new Date(message.timestamp).toLocaleTimeString()}</span>
        </div>
      </div>
    </div>
  );
}


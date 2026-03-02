import { useState, useRef, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { useChatStore } from '../stores/chatStore';
import { useChat } from '../hooks/useChat';

export default function ChatPanel() {
  const [input, setInput] = useState('');
  const { messages, isStreaming, streamingText } = useChatStore();
  const { sendMessage } = useChat();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, streamingText]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || isStreaming) return;
    sendMessage(input.trim());
    setInput('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-6 py-4 space-y-5">
        {messages.length === 0 && !isStreaming && (
          <div className="flex items-center justify-center h-full" style={{ color: 'var(--text-secondary)' }}>
            <div className="text-center">
              <p className="text-2xl font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>StockMind</p>
              <p className="text-sm">输入问题开始投资分析对话</p>
            </div>
          </div>
        )}
        {messages.map((msg) => (
          <div key={msg.id} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            {msg.role === 'assistant' ? (
              <div className="max-w-[85%] rounded-2xl px-5 py-3 shadow-sm"
                style={{ background: 'var(--bg-assistant-msg)', color: 'var(--text-primary)' }}>
                <div className="assistant-prose">
                  <ReactMarkdown>{msg.content}</ReactMarkdown>
                </div>
              </div>
            ) : (
              <div className="max-w-[75%] rounded-2xl px-4 py-2.5 text-sm text-white shadow-sm"
                style={{ background: 'var(--bg-user-msg)' }}>
                <p className="whitespace-pre-wrap">{msg.content}</p>
              </div>
            )}
          </div>
        ))}
        {isStreaming && streamingText && (
          <div className="flex justify-start">
            <div className="max-w-[85%] rounded-2xl px-5 py-3 shadow-sm"
              style={{ background: 'var(--bg-assistant-msg)', color: 'var(--text-primary)' }}>
              <div className="assistant-prose">
                <ReactMarkdown>{streamingText}</ReactMarkdown>
              </div>
            </div>
          </div>
        )}
        {isStreaming && !streamingText && (
          <div className="flex justify-start">
            <div className="rounded-2xl px-5 py-3" style={{ background: 'var(--bg-assistant-msg)', color: 'var(--text-secondary)' }}>
              <span className="inline-flex items-center gap-1 text-sm">
                <span className="animate-pulse">thinking...</span>
              </span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="px-6 pb-4 pt-2">
        <form onSubmit={handleSubmit}
          className="flex items-end gap-2 rounded-2xl px-4 py-3 shadow-sm"
          style={{ background: 'var(--bg-input)', border: '1px solid var(--border-color)' }}>
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="输入问题，如：帮我看看茅台最近走势怎么样"
            className="flex-1 resize-none text-sm leading-relaxed focus:outline-none"
            style={{ background: 'transparent', color: 'var(--text-primary)' }}
            rows={1}
            disabled={isStreaming}
            onInput={(e) => {
              const el = e.currentTarget;
              el.style.height = 'auto';
              el.style.height = Math.min(el.scrollHeight, 120) + 'px';
            }}
          />
          <button
            type="submit"
            disabled={isStreaming || !input.trim()}
            className="shrink-0 rounded-xl px-4 py-1.5 text-sm font-medium text-white transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            style={{ background: 'var(--accent)' }}
            onMouseEnter={e => e.currentTarget.style.background = 'var(--accent-hover)'}
            onMouseLeave={e => e.currentTarget.style.background = 'var(--accent)'}
          >
            发送
          </button>
        </form>
      </div>
    </div>
  );
}

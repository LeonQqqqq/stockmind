import { useEffect } from 'react';
import { useChatStore } from '../stores/chatStore';
import { useChat } from '../hooks/useChat';

export default function Sidebar() {
  const { sessions, currentSessionId, setCurrentSession } = useChatStore();
  const { loadSessions, loadMessages, newSession, deleteSession } = useChat();

  useEffect(() => {
    loadSessions();
  }, [loadSessions]);

  const handleSelectSession = (id: string) => {
    setCurrentSession(id);
    loadMessages(id);
  };

  return (
    <div className="w-56 flex flex-col" style={{ background: 'var(--bg-sidebar)' }}>
      <div className="p-3" style={{ borderBottom: '1px solid var(--border-sidebar)' }}>
        <button
          onClick={newSession}
          className="w-full rounded-lg px-3 py-2 text-sm transition-colors"
          style={{
            border: '1px solid var(--border-sidebar)',
            color: 'var(--text-sidebar)',
          }}
          onMouseEnter={e => e.currentTarget.style.background = 'rgba(255,255,255,0.05)'}
          onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
        >
          + 新对话
        </button>
      </div>
      <div className="flex-1 overflow-y-auto">
        {sessions.map((s) => (
          <div
            key={s.id}
            className="flex items-center justify-between px-3 py-2.5 cursor-pointer text-sm transition-colors"
            style={{
              background: currentSessionId === s.id ? 'rgba(255,255,255,0.08)' : 'transparent',
              color: currentSessionId === s.id ? 'var(--text-sidebar-active)' : 'var(--text-sidebar)',
            }}
            onClick={() => handleSelectSession(s.id)}
            onMouseEnter={e => { if (currentSessionId !== s.id) e.currentTarget.style.background = 'rgba(255,255,255,0.04)'; }}
            onMouseLeave={e => { if (currentSessionId !== s.id) e.currentTarget.style.background = 'transparent'; }}
          >
            <span className="truncate flex-1">{s.title}</span>
            <button
              onClick={(e) => { e.stopPropagation(); deleteSession(s.id); }}
              className="opacity-0 group-hover:opacity-100 hover:text-red-400 ml-1 text-xs"
              style={{ color: 'var(--text-sidebar)', opacity: 0.4 }}
              onMouseEnter={e => { e.currentTarget.style.opacity = '1'; e.currentTarget.style.color = '#ef4444'; }}
              onMouseLeave={e => { e.currentTarget.style.opacity = '0.4'; e.currentTarget.style.color = 'var(--text-sidebar)'; }}
            >
              x
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}

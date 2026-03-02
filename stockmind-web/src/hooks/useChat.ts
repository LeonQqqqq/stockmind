import { useCallback } from 'react';
import { useChatStore } from '../stores/chatStore';

let _msgId = 0;
function genId() {
  return `msg-${Date.now()}-${++_msgId}`;
}

export function useChat() {
  const store = useChatStore;

  const sendMessage = useCallback(async (content: string) => {
    const { currentSessionId, addMessage, setStreaming, setStreamingText } = store.getState();

    addMessage({
      id: genId(),
      role: 'user',
      content,
      timestamp: Date.now(),
    });
    setStreaming(true);
    setStreamingText('');

    try {
      const resp = await fetch('/api/v1/chat/stream', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: content,
          session_id: currentSessionId || undefined,
        }),
      });

      if (!resp.ok || !resp.body) {
        throw new Error(`HTTP ${resp.status}`);
      }

      const reader = resp.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let currentEvent = '';
      let dataLines: string[] = [];

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line.startsWith('event:')) {
            currentEvent = line.slice(6).trim();
            dataLines = [];
          } else if (line.startsWith('data:')) {
            dataLines.push(line.slice(5).trimStart());
          } else if (line.trim() === '' && currentEvent) {
            const data = dataLines.join('\n');
            dataLines = [];
            if (data === '[DONE]') { currentEvent = ''; continue; }

            if (currentEvent === 'session') {
              store.getState().setCurrentSession(data);
            } else if (currentEvent === 'error') {
              store.getState().appendStreamingText('\n\n**Error:** ' + data);
            } else if (currentEvent === 'message') {
              store.getState().appendStreamingText(data);
            }
            currentEvent = '';
          }
        }
      }

      const finalText = store.getState().streamingText;
      if (finalText) {
        store.getState().addMessage({
          id: genId(),
          role: 'assistant',
          content: finalText,
          timestamp: Date.now(),
        });
      }
    } catch (err) {
      console.error('Chat error:', err);
      store.getState().addMessage({
        id: genId(),
        role: 'assistant',
        content: `**Error:** ${err instanceof Error ? err.message : 'Unknown error'}`,
        timestamp: Date.now(),
      });
    } finally {
      store.getState().setStreaming(false);
      store.getState().setStreamingText('');
    }
  }, []);

  const loadSessions = useCallback(async () => {
    const resp = await fetch('/api/v1/sessions');
    const json = await resp.json();
    useChatStore.getState().setSessions(json.data || []);
  }, []);

  const loadMessages = useCallback(async (sessionId: string) => {
    const resp = await fetch(`/api/v1/sessions/${sessionId}/messages`);
    const json = await resp.json();
    const msgs = (json.data || []).map((m: any) => ({
      id: String(m.id),
      role: m.role,
      content: m.content,
      timestamp: new Date(m.created_at).getTime(),
    }));
    useChatStore.getState().setMessages(msgs);
  }, []);

  const newSession = useCallback(() => {
    useChatStore.getState().setCurrentSession(null);
    useChatStore.getState().setMessages([]);
  }, []);

  const deleteSession = useCallback(async (id: string) => {
    await fetch(`/api/v1/sessions/${id}`, { method: 'DELETE' });
    await loadSessions();
    if (useChatStore.getState().currentSessionId === id) {
      newSession();
    }
  }, [loadSessions, newSession]);

  return { sendMessage, loadSessions, loadMessages, newSession, deleteSession };
}

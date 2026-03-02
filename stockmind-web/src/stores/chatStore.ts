import { create } from 'zustand';
import type { ChatMessage, Session } from '../types';

interface ChatState {
  sessions: Session[];
  currentSessionId: string | null;
  messages: ChatMessage[];
  isStreaming: boolean;
  streamingText: string;

  setSessions: (sessions: Session[]) => void;
  setCurrentSession: (id: string | null) => void;
  setMessages: (messages: ChatMessage[]) => void;
  addMessage: (message: ChatMessage) => void;
  setStreaming: (streaming: boolean) => void;
  setStreamingText: (text: string) => void;
  appendStreamingText: (text: string) => void;
}

export const useChatStore = create<ChatState>((set) => ({
  sessions: [],
  currentSessionId: null,
  messages: [],
  isStreaming: false,
  streamingText: '',

  setSessions: (sessions) => set({ sessions }),
  setCurrentSession: (id) => set({ currentSessionId: id }),
  setMessages: (messages) => set({ messages }),
  addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),
  setStreaming: (streaming) => set({ isStreaming: streaming }),
  setStreamingText: (text) => set({ streamingText: text }),
  appendStreamingText: (text) => set((state) => ({ streamingText: state.streamingText + text })),
}));

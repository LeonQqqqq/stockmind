import { useState } from 'react';
import Sidebar from './components/Sidebar';
import ChatPanel from './components/ChatPanel';
import MemoryModal from './components/MemoryModal';

function App() {
  const [showMemory, setShowMemory] = useState(false);

  return (
    <div className="flex h-screen" style={{ background: 'var(--bg-primary)', color: 'var(--text-primary)' }}>
      <Sidebar />
      <div className="flex-1 min-w-0 relative">
        <ChatPanel />
        <button
          onClick={() => setShowMemory(true)}
          className="absolute top-3 right-4 rounded-lg px-3 py-1.5 text-sm font-medium text-white shadow-md transition-colors z-10"
          style={{ background: 'var(--accent)' }}
          onMouseEnter={e => e.currentTarget.style.background = 'var(--accent-hover)'}
          onMouseLeave={e => e.currentTarget.style.background = 'var(--accent)'}
        >
          经验库
        </button>
      </div>
      <MemoryModal open={showMemory} onClose={() => setShowMemory(false)} />
    </div>
  );
}

export default App;

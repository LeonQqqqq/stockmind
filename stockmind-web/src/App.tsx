import Sidebar from './components/Sidebar';
import ChatPanel from './components/ChatPanel';
import MemoryPanel from './components/MemoryPanel';

function App() {
  return (
    <div className="flex h-screen" style={{ background: 'var(--bg-primary)', color: 'var(--text-primary)' }}>
      <Sidebar />
      <div className="flex flex-1 min-w-0">
        <div className="flex-[3] min-w-0" style={{ borderRight: '1px solid var(--border-color)' }}>
          <ChatPanel />
        </div>
        <div className="flex-[2] min-w-0">
          <MemoryPanel />
        </div>
      </div>
    </div>
  );
}

export default App;

import { useState, useEffect, useCallback } from 'react';
import { useMemoryStore } from '../stores/memoryStore';
import type { Experience } from '../types';

export default function MemoryPanel() {
  const { experiences, searchKeyword, setExperiences, setSearchKeyword, addExperience, removeExperience } = useMemoryStore();
  const [showForm, setShowForm] = useState(false);
  const [formTitle, setFormTitle] = useState('');
  const [formContent, setFormContent] = useState('');
  const [formTags, setFormTags] = useState('');

  const loadExperiences = useCallback(async () => {
    const url = searchKeyword
      ? `/api/v1/experiences/search?keyword=${encodeURIComponent(searchKeyword)}`
      : '/api/v1/experiences';
    const resp = await fetch(url);
    const json = await resp.json();
    setExperiences(json.data || []);
  }, [searchKeyword, setExperiences]);

  useEffect(() => {
    loadExperiences();
  }, [loadExperiences]);

  const handleCreate = async () => {
    if (!formTitle.trim()) return;
    const resp = await fetch('/api/v1/experiences', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title: formTitle, content: formContent, tags: formTags }),
    });
    const json = await resp.json();
    if (json.data?.id) {
      addExperience({
        id: json.data.id,
        title: formTitle,
        content: formContent,
        tags: formTags,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      });
    }
    setFormTitle('');
    setFormContent('');
    setFormTags('');
    setShowForm(false);
  };

  const handleDelete = async (id: number) => {
    await fetch(`/api/v1/experiences/${id}`, { method: 'DELETE' });
    removeExperience(id);
  };

  const inputStyle = {
    background: 'var(--bg-input)',
    color: 'var(--text-primary)',
    border: '1px solid var(--border-color)',
  };

  return (
    <div className="flex flex-col h-full" style={{ background: 'var(--bg-secondary)' }}>
      {/* Header */}
      <div className="p-4" style={{ borderBottom: '1px solid var(--border-color)' }}>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-base font-semibold" style={{ color: 'var(--text-primary)' }}>投资经验库</h2>
          <button
            onClick={() => setShowForm(!showForm)}
            className="rounded-lg px-3 py-1 text-xs font-medium text-white transition-colors"
            style={{ background: 'var(--accent)' }}
            onMouseEnter={e => e.currentTarget.style.background = 'var(--accent-hover)'}
            onMouseLeave={e => e.currentTarget.style.background = 'var(--accent)'}
          >
            {showForm ? '取消' : '+ 添加'}
          </button>
        </div>
        <input
          type="text"
          value={searchKeyword}
          onChange={(e) => setSearchKeyword(e.target.value)}
          placeholder="搜索经验..."
          className="w-full rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-1"
          style={{ ...inputStyle, '--tw-ring-color': 'var(--accent)' } as any}
        />
      </div>

      {/* Add form */}
      {showForm && (
        <div className="p-4 space-y-2" style={{ borderBottom: '1px solid var(--border-color)' }}>
          <input type="text" value={formTitle} onChange={(e) => setFormTitle(e.target.value)}
            placeholder="标题" className="w-full rounded-lg px-3 py-1.5 text-sm focus:outline-none" style={inputStyle} />
          <textarea value={formContent} onChange={(e) => setFormContent(e.target.value)}
            placeholder="内容" rows={3} className="w-full resize-none rounded-lg px-3 py-1.5 text-sm focus:outline-none" style={inputStyle} />
          <input type="text" value={formTags} onChange={(e) => setFormTags(e.target.value)}
            placeholder="标签（逗号分隔）" className="w-full rounded-lg px-3 py-1.5 text-sm focus:outline-none" style={inputStyle} />
          <button onClick={handleCreate}
            className="rounded-lg px-3 py-1.5 text-sm font-medium text-white"
            style={{ background: 'var(--accent)' }}>
            保存
          </button>
        </div>
      )}

      {/* Experience list */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {experiences.length === 0 && (
          <p className="text-center text-sm mt-8" style={{ color: 'var(--text-secondary)' }}>暂无经验记录</p>
        )}
        {experiences.map((exp: Experience) => (
          <div key={exp.id} className="rounded-xl p-3 shadow-sm"
            style={{ background: 'var(--bg-input)', border: '1px solid var(--border-color)' }}>
            <div className="flex items-start justify-between">
              <h3 className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{exp.title}</h3>
              <button onClick={() => handleDelete(exp.id)}
                className="text-xs ml-2 transition-colors"
                style={{ color: 'var(--text-secondary)' }}
                onMouseEnter={e => e.currentTarget.style.color = '#ef4444'}
                onMouseLeave={e => e.currentTarget.style.color = 'var(--text-secondary)'}>
                删除
              </button>
            </div>
            <p className="text-xs mt-1 line-clamp-3" style={{ color: 'var(--text-secondary)' }}>{exp.content}</p>
            {exp.tags && (
              <div className="flex flex-wrap gap-1 mt-2">
                {exp.tags.split(',').map((tag, i) => (
                  <span key={i} className="rounded-md px-1.5 py-0.5 text-xs"
                    style={{ background: 'var(--bg-secondary)', color: 'var(--text-secondary)' }}>
                    {tag.trim()}
                  </span>
                ))}
              </div>
            )}
            <p className="text-xs mt-2" style={{ color: 'var(--border-color)' }}>
              {new Date(exp.updated_at).toLocaleDateString()}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}

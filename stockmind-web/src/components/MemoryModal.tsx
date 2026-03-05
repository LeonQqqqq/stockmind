import { useState, useEffect, useCallback } from 'react';
import { useMemoryStore } from '../stores/memoryStore';
import type { Experience, Opinion } from '../types';

interface Props {
  open: boolean;
  onClose: () => void;
}

export default function MemoryModal({ open, onClose }: Props) {
  const {
    experiences, opinions, searchKeyword,
    setExperiences, setOpinions, setSearchKeyword,
    addExperience, removeExperience, removeOpinion,
  } = useMemoryStore();

  const [tab, setTab] = useState<'experiences' | 'opinions'>('experiences');
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [formTitle, setFormTitle] = useState('');
  const [formContent, setFormContent] = useState('');
  const [formTags, setFormTags] = useState('');
  const [authorFilter, setAuthorFilter] = useState('');
  const [authors, setAuthors] = useState<string[]>([]);

  const loadExperiences = useCallback(async () => {
    const url = searchKeyword
      ? `/api/v1/experiences/search?keyword=${encodeURIComponent(searchKeyword)}`
      : '/api/v1/experiences';
    const resp = await fetch(url);
    const json = await resp.json();
    setExperiences(json.data || []);
  }, [searchKeyword, setExperiences]);

  const loadOpinions = useCallback(async () => {
    let url = '/api/v1/opinions';
    if (searchKeyword) {
      url = `/api/v1/opinions/search?keyword=${encodeURIComponent(searchKeyword)}`;
    } else if (authorFilter) {
      url = `/api/v1/opinions?author=${encodeURIComponent(authorFilter)}`;
    }
    const resp = await fetch(url);
    const json = await resp.json();
    setOpinions(json.data || []);
  }, [searchKeyword, authorFilter, setOpinions]);

  const loadAuthors = useCallback(async () => {
    const resp = await fetch('/api/v1/opinions/authors');
    const json = await resp.json();
    setAuthors(json.data || []);
  }, []);

  useEffect(() => {
    if (!open) return;
    setSearchKeyword('');
    setAuthorFilter('');
    setExpandedId(null);
    setShowForm(false);
    loadExperiences();
    loadOpinions();
    loadAuthors();
  }, [open]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!open) return;
    if (tab === 'experiences') loadExperiences();
    else loadOpinions();
  }, [searchKeyword, authorFilter]); // eslint-disable-line react-hooks/exhaustive-deps

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

  const handleDeleteExperience = async (id: number) => {
    await fetch(`/api/v1/experiences/${id}`, { method: 'DELETE' });
    removeExperience(id);
  };

  const handleDeleteOpinion = async (id: number) => {
    await fetch(`/api/v1/opinions/${id}`, { method: 'DELETE' });
    removeOpinion(id);
  };

  const toggleExpand = (key: string) => {
    setExpandedId(expandedId === key ? null : key);
  };

  if (!open) return null;

  const inputStyle = {
    background: 'var(--bg-input)',
    color: 'var(--text-primary)',
    border: '1px solid var(--border-color)',
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      style={{ background: 'rgba(0,0,0,0.5)' }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div
        className="flex flex-col rounded-2xl shadow-2xl"
        style={{
          background: 'var(--bg-primary)',
          width: '700px',
          maxWidth: '90vw',
          height: '80vh',
          maxHeight: '80vh',
        }}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4" style={{ borderBottom: '1px solid var(--border-color)' }}>
          <div className="flex gap-1 rounded-lg p-1" style={{ background: 'var(--bg-secondary)' }}>
            <button
              onClick={() => { setTab('experiences'); setSearchKeyword(''); setExpandedId(null); }}
              className="rounded-md px-4 py-1.5 text-sm font-medium transition-colors"
              style={{
                background: tab === 'experiences' ? 'var(--accent)' : 'transparent',
                color: tab === 'experiences' ? '#fff' : 'var(--text-secondary)',
              }}
            >
              经验库
            </button>
            <button
              onClick={() => { setTab('opinions'); setSearchKeyword(''); setExpandedId(null); }}
              className="rounded-md px-4 py-1.5 text-sm font-medium transition-colors"
              style={{
                background: tab === 'opinions' ? 'var(--accent)' : 'transparent',
                color: tab === 'opinions' ? '#fff' : 'var(--text-secondary)',
              }}
            >
              智囊团观点
            </button>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-lg leading-none transition-colors"
            style={{ color: 'var(--text-secondary)' }}
            onMouseEnter={e => e.currentTarget.style.color = 'var(--text-primary)'}
            onMouseLeave={e => e.currentTarget.style.color = 'var(--text-secondary)'}
          >
            ✕
          </button>
        </div>

        {/* Toolbar */}
        <div className="flex items-center gap-2 px-6 py-3" style={{ borderBottom: '1px solid var(--border-color)' }}>
          {tab === 'opinions' && authors.length > 0 && (
            <select
              value={authorFilter}
              onChange={(e) => setAuthorFilter(e.target.value)}
              className="rounded-lg px-3 py-1.5 text-sm focus:outline-none"
              style={inputStyle}
            >
              <option value="">全部作者</option>
              {authors.map(a => <option key={a} value={a}>{a}</option>)}
            </select>
          )}
          <input
            type="text"
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            placeholder={tab === 'experiences' ? '搜索经验...' : '搜索观点...'}
            className="flex-1 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-1"
            style={{ ...inputStyle, '--tw-ring-color': 'var(--accent)' } as any}
          />
          {tab === 'experiences' && (
            <button
              onClick={() => setShowForm(!showForm)}
              className="rounded-lg px-3 py-1.5 text-sm font-medium text-white whitespace-nowrap transition-colors"
              style={{ background: 'var(--accent)' }}
              onMouseEnter={e => e.currentTarget.style.background = 'var(--accent-hover)'}
              onMouseLeave={e => e.currentTarget.style.background = 'var(--accent)'}
            >
              {showForm ? '取消' : '+ 添加'}
            </button>
          )}
        </div>

        {/* Add form */}
        {showForm && tab === 'experiences' && (
          <div className="px-6 py-3 space-y-2" style={{ borderBottom: '1px solid var(--border-color)' }}>
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

        {/* Content */}
        <div className="flex-1 overflow-y-auto px-6 py-4 space-y-3">
          {tab === 'experiences' && (
            <>
              {experiences.length === 0 && (
                <p className="text-center text-sm mt-8" style={{ color: 'var(--text-secondary)' }}>暂无经验记录</p>
              )}
              {experiences.map((exp: Experience) => {
                const key = `exp-${exp.id}`;
                const expanded = expandedId === key;
                return (
                  <div key={exp.id} className="rounded-xl p-3 shadow-sm cursor-pointer transition-colors"
                    style={{ background: 'var(--bg-input)', border: '1px solid var(--border-color)' }}
                    onClick={() => toggleExpand(key)}
                  >
                    <div className="flex items-start justify-between">
                      <h3 className="text-sm font-medium" style={{ color: 'var(--text-primary)' }}>{exp.title}</h3>
                      <div className="flex items-center gap-2 ml-2">
                        <span className="text-xs" style={{ color: 'var(--text-secondary)' }}>{expanded ? '▲' : '▼'}</span>
                        <button
                          onClick={(e) => { e.stopPropagation(); handleDeleteExperience(exp.id); }}
                          className="text-xs transition-colors"
                          style={{ color: 'var(--text-secondary)' }}
                          onMouseEnter={e => e.currentTarget.style.color = '#ef4444'}
                          onMouseLeave={e => e.currentTarget.style.color = 'var(--text-secondary)'}
                        >
                          删除
                        </button>
                      </div>
                    </div>
                    <p className={`text-xs mt-1 whitespace-pre-wrap ${expanded ? '' : 'line-clamp-2'}`}
                      style={{ color: 'var(--text-secondary)' }}>{exp.content}</p>
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
                );
              })}
            </>
          )}

          {tab === 'opinions' && (
            <>
              {opinions.length === 0 && (
                <p className="text-center text-sm mt-8" style={{ color: 'var(--text-secondary)' }}>暂无观点记录</p>
              )}
              {opinions.map((op: Opinion) => {
                const key = `op-${op.id}`;
                const expanded = expandedId === key;
                return (
                  <div key={op.id} className="rounded-xl p-3 shadow-sm cursor-pointer transition-colors"
                    style={{ background: 'var(--bg-input)', border: '1px solid var(--border-color)' }}
                    onClick={() => toggleExpand(key)}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-2">
                        <span className="rounded-md px-2 py-0.5 text-xs font-medium"
                          style={{ background: 'var(--accent)', color: '#fff' }}>{op.author}</span>
                        <span className="text-xs" style={{ color: 'var(--text-secondary)' }}>{expanded ? '▲' : '▼'}</span>
                      </div>
                      <button
                        onClick={(e) => { e.stopPropagation(); handleDeleteOpinion(op.id); }}
                        className="text-xs ml-2 transition-colors"
                        style={{ color: 'var(--text-secondary)' }}
                        onMouseEnter={e => e.currentTarget.style.color = '#ef4444'}
                        onMouseLeave={e => e.currentTarget.style.color = 'var(--text-secondary)'}
                      >
                        删除
                      </button>
                    </div>
                    <p className={`text-xs mt-2 whitespace-pre-wrap ${expanded ? '' : 'line-clamp-2'}`}
                      style={{ color: 'var(--text-secondary)' }}>{op.content}</p>
                    {op.tags && (
                      <div className="flex flex-wrap gap-1 mt-2">
                        {op.tags.split(',').map((tag, i) => (
                          <span key={i} className="rounded-md px-1.5 py-0.5 text-xs"
                            style={{ background: 'var(--bg-secondary)', color: 'var(--text-secondary)' }}>
                            {tag.trim()}
                          </span>
                        ))}
                      </div>
                    )}
                    <p className="text-xs mt-2" style={{ color: 'var(--border-color)' }}>
                      {new Date(op.created_at).toLocaleDateString()}
                    </p>
                  </div>
                );
              })}
            </>
          )}
        </div>
      </div>
    </div>
  );
}

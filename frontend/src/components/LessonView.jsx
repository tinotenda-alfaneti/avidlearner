import React from 'react';

function linkifyText(text) {
  if (!text) return text;
  const urlRegex = /(https?:\/\/[^\s]+)/g;
  const parts = text.split(urlRegex);
  return parts.map((part, i) => {
    if (part.match(urlRegex)) {
      return <a key={i} href={part} target="_blank" rel="noopener noreferrer" style={{color: '#5b9fff', textDecoration: 'underline'}}>{part}</a>;
    }
    return part;
  });
}

function categoryLabel(value) {
  if (value === 'any') return 'Any';
  if (value === 'random') return 'Random';
  return value.replace(/[-_]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

export default function LessonView({
  lesson,
  categoryOptions = [],
  selectedCategory = 'any',
  onSelectCategory,
  selectedSource = 'all',
  onSelectSource,
  aiEnabled = false,
  onNext,
  onStartQuiz,
  onExit
}) {
  if (!lesson) return null;

  const sourceOptions = [
    { value: 'all', label: 'All Sources' },
    { value: 'local', label: 'Local (70 lessons)' },
    { value: 'github', label: 'GitHub' },
    { value: 'secret-knowledge', label: 'Secret Knowledge' },
    { value: 'devto', label: 'Dev.to' },
    { value: 'ai', label: aiEnabled ? 'AI Generated' : 'AI (Coming Soon)', disabled: !aiEnabled }
  ];

  return (
    <div className="card">
      <div className="nav">
        <button className="home" onClick={()=>onExit && onExit()}>Go Home</button>
      </div>
      
      <div style={{display:'flex', gap: 8, flexWrap: 'wrap', marginBottom:12}}>
        {/* Source Filter */}
        <div className="badge">
          Source:
          <select
            aria-label="Lesson Source"
            value={selectedSource}
            onChange={e => onSelectSource && onSelectSource(e.target.value)}
            className="badge-select"
          >
            {sourceOptions.map(opt => (
              <option key={opt.value} value={opt.value} disabled={opt.disabled}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        {/* Category Filter */}
        <div className="badge">
          Category:
          <select
            aria-label="Category"
            value={selectedCategory}
            onChange={e => onSelectCategory && onSelectCategory(e.target.value)}
            className="badge-select"
          >
            {categoryOptions.map(opt => (
              <option key={opt} value={opt}>{categoryLabel(opt)}</option>
            ))}
          </select>
        </div>
      </div>

      <div style={{color:'#7d89b0', fontWeight:600}}>
        {lesson.category} · {lesson.title}
        {lesson.source && <span style={{marginLeft: 8, fontSize: '0.9em', opacity: 0.7}}>
          ({lesson.source === 'local' ? 'Local' : 
            lesson.source === 'github' ? 'GitHub' :
            lesson.source === 'secret-knowledge' ? 'Secret Knowledge' :
            lesson.source === 'devto' ? 'Dev.to' : 'AI'})
        </span>}
      </div>
      <div className="prompt" style={{marginTop:8}}>{linkifyText(lesson.text)}</div>

      <div style={{marginTop:12}}>
        <h4>About this topic</h4>
        <p>{linkifyText(lesson.explain)}</p>

        <h4>Where to use it</h4>
        <ul>{(lesson.useCases||[]).map((x,i)=>(<li key={i}>{linkifyText(x)}</li>))}</ul>

        <h4>Tips</h4>
        <ul>{(lesson.tips||[]).map((x,i)=>(<li key={i}>{linkifyText(x)}</li>))}</ul>
      </div>

      <div className="row" style={{ marginTop: 12, display: 'flex', justifyContent: 'space-between' }}>
        <button className="primary" onClick={onNext}>Next concept</button>
        <button className="primary" onClick={onStartQuiz}>Start quiz</button>
      </div>

      <div className="footer">You'll only be quizzed on concepts you read. Don't chew what you can't swallow.</div>
    </div>
  );
}

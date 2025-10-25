import React from 'react';

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
  onNext,
  onStartQuiz,
  onExit
}) {
  if (!lesson) return null;
  return (
    <div className="card">
      <div className="nav">
        <button className="home" onClick={()=>onExit && onExit()}>Go Home</button>
      </div>
      <div style={{display:'flex', justifyContent:'flex-start', marginBottom:12}}>
        <div className="badge">
          <span role="img" aria-hidden="true">📚</span>Category:
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
      <div style={{color:'#7d89b0', fontWeight:600}}>{lesson.category} · {lesson.title}</div>
      <div className="prompt" style={{marginTop:8}}>{lesson.text}</div>

      <div style={{marginTop:12}}>
        <h4>About this topic</h4>
        <p>{lesson.explain}</p>

        <h4>Where to use it</h4>
        <ul>{(lesson.useCases||[]).map((x,i)=>(<li key={i}>{x}</li>))}</ul>

        <h4>Tips</h4>
        <ul>{(lesson.tips||[]).map((x,i)=>(<li key={i}>{x}</li>))}</ul>
      </div>

      <div className="row" style={{ marginTop: 12, display: 'flex', justifyContent: 'space-between' }}>
        <button className="primary" onClick={onNext}>Next concept</button>
        <button className="primary" onClick={onStartQuiz}>Start quiz</button>
      </div>

      <div className="footer">You'll only be quizzed on concepts you read. Don't chew what you can't swallow.</div>
    </div>
  );
}

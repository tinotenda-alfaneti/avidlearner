import React, { useEffect, useRef, useState } from 'react';
import { randomLesson, updateTypingScore } from '../api';

function categoryLabel(value) {
  if (value === 'any') return 'Any';
  if (value === 'random') return 'Random';
  return value.replace(/[-_]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

export default function TypingView({
  onExit,
  categoryOptions = [],
  availableCategories = [],
  selectedCategory = 'any',
  onSelectCategory,
  onTypingStats,
  typingBest = 0,
  onSubmitToLeaderboard,
}) {
  const [lesson, setLesson] = useState(null);
  const [text, setText] = useState('');
  const [typed, setTyped] = useState('');
  const [running, setRunning] = useState(false);
  const [duration, setDuration] = useState(60);
  const [startTime, setStartTime] = useState(0);
  const [deadline, setDeadline] = useState(0);
  const [remain, setRemain] = useState(60);
  const [stats, setStats] = useState(() => ({ wpm: 0, acc: 100, streak: 0, best: typingBest }));
  const inputRef = useRef(null);
  const statsRef = useRef({ wpm: 0, acc: 100, streak: 0, best: typingBest });

  function computeStats(t, src, startMs) {
    const correct = t.split('').filter((ch, i) => ch === src[i]).length;
    const minutes = Math.max(0.001, (performance.now() - startMs) / 60000);
    const wpm = Math.round((correct / 5) / minutes);
    const acc = t.length ? Math.round(100 * correct / t.length) : 100;
    return { wpm, acc, streak: stats.streak, best: stats.best };
  }

  function pickCategory() {
    if (selectedCategory === 'random') {
      if (!availableCategories.length) return 'any';
      const choice = availableCategories[Math.floor(Math.random() * availableCategories.length)];
      return choice || 'any';
    }
    return selectedCategory || 'any';
  }

  async function start() {
    const resetStats = { wpm: 0, acc: 100, streak: 0, best: typingBest };
    setStats(resetStats);
    statsRef.current = resetStats;
    onTypingStats && onTypingStats({ streak: 0, best: typingBest });
    const l = await randomLesson(pickCategory());
    setLesson(l);
    setText(l.text);
    setTyped('');
    setRunning(true);
    const now = performance.now();
    setStartTime(now);
    setDeadline(now + duration * 1000);
    setRemain(duration);
    inputRef.current && inputRef.current.focus();
  }

  // realtime countdown
  useEffect(() => {
    if (!running) return;
    const id = setInterval(() => {
      const r = Math.max(0, Math.ceil((deadline - performance.now()) / 1000));
      setRemain(r);
      if (r <= 0) setRunning(false);
    }, 200);
    return () => clearInterval(id);
  }, [running, deadline]);

  // typing handler
  function onChange(e) {
    if (!running) return;
    const v = e.target.value;
    const i = v.length - 1;
    let streak = stats.streak, best = stats.best;
    if (i >= 0) {
      if (v[i] === text[i]) { streak++; best = Math.max(best, streak); }
      else { streak = 0; }
    }
    const nextStats = { ...computeStats(v, text, startTime), streak, best };
    setStats(nextStats);
    onTypingStats && onTypingStats({ streak: nextStats.streak, best: nextStats.best });
    setTyped(v);
    if (v.length >= text.length) setRunning(false);
  }
  useEffect(() => { statsRef.current = stats; }, [stats]);

  useEffect(() => {
    setStats(prev => {
      const updated = { ...prev, best: Math.max(prev.best, typingBest) };
      statsRef.current = updated;
      return updated;
    });
  }, [typingBest]);

  useEffect(() => {
    if (!running && lesson && typed) {
      onTypingStats && onTypingStats({ streak: statsRef.current.streak, best: statsRef.current.best });
      // Update server-side typing score
      updateTypingScore(statsRef.current.wpm).catch(err => {
        console.error('Failed to update typing score:', err);
      });
    }
  }, [running, lesson, typed, onTypingStats]);

  function renderText() {
    return text.split('').map((ch, i) => {
      let cls = '';
      if (i < typed.length) cls = typed[i] === ch ? 'good' : 'bad';
      return <span key={i} className={cls}>{ch}</span>;
    });
  }

  return (
    <div className="card">
      <div className="nav">
        <button className="home" onClick={()=>onExit && onExit()}>Go Home</button>
      </div>

      <div className="row">
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
        <div className="badge">
          <span role="img" aria-hidden="true">⏱</span>Duration:
          <input
            className="badge-input"
            type="number"
            value={duration}
            min="15"
            max="300"
            step="15"
            onChange={e=>setDuration(parseInt(e.target.value||'60',10))}
          />
          <span>sec</span>
        </div>
        <button className="badge badge-button" onClick={start}>▶ Start</button>
      </div>

      <div className="metrics" style={{marginTop:12}}>
        <div className="metric"><h3>Time</h3><p>{running ? `${remain}s` : '—'}</p></div>
        <div className="metric"><h3>WPM</h3><p>{stats.wpm}</p></div>
        <div className="metric"><h3>Accuracy</h3><p>{stats.acc}%</p></div>
        <div className="metric"><h3>Streak</h3><p>{stats.streak}</p></div>
        <div className="metric"><h3>Best</h3><p>{stats.best}</p></div>
      </div>

      {lesson && <div style={{marginTop:12,color:'#7d89b0',fontWeight:600}}>
        {lesson.category} · {lesson.title}
      </div>}

      <div className="typing-box" onClick={()=>inputRef.current && inputRef.current.focus()}>
        <div className="typing-text">{renderText()}</div>
        <input
          ref={inputRef}
          className="typing-input"
          value={typed}
          onChange={onChange}
          autoComplete="off"
          autoCapitalize="none"
          autoCorrect="off"
        />
      </div>

      {/* summary once finished */}
      {!running && lesson && typed && (
        <div className="typing-summary">
          <p>WPM <b>{stats.wpm}</b> · Accuracy <b>{stats.acc}%</b></p>
        
          <div className="row" style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            <button className="primary" onClick={start}>Next</button>
            {onSubmitToLeaderboard && stats.wpm > 0 && (
              <button className="primary secondary" onClick={() => onSubmitToLeaderboard(stats.wpm)}>
                Submit to Leaderboard
              </button>
            )}
            <button className="primary" style={{ backgroundColor: '#e74c3c' }} onClick={()=>onExit && onExit()}>End</button>
          </div>
        </div>
      )}
    </div>
  );
}

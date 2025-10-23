import React, { useEffect, useRef, useState } from 'react';
import { randomLesson } from '../api';

export default function TypingView({ onExit }) {
  const [lesson, setLesson] = useState(null);
  const [text, setText] = useState('');
  const [typed, setTyped] = useState('');
  const [running, setRunning] = useState(false);
  const [duration, setDuration] = useState(60);
  const [startTime, setStartTime] = useState(0);
  const [deadline, setDeadline] = useState(0);
  const [remain, setRemain] = useState(60);
  const [stats, setStats] = useState({ wpm: 0, acc: 100, streak: 0, best: 0 });
  const inputRef = useRef(null);

  function computeStats(t, src, startMs) {
    const correct = t.split('').filter((ch, i) => ch === src[i]).length;
    const minutes = Math.max(0.001, (performance.now() - startMs) / 60000);
    const wpm = Math.round((correct / 5) / minutes);
    const acc = t.length ? Math.round(100 * correct / t.length) : 100;
    return { wpm, acc, streak: stats.streak, best: stats.best };
  }

  async function start() {
    const l = await randomLesson('any');
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
    setStats({ ...computeStats(v, text, startTime), streak, best });
    setTyped(v);
    if (v.length >= text.length) setRunning(false);
  }

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
        <label>Duration&nbsp;
          <input type="number" value={duration} min="15" max="300" step="15"
                 onChange={e=>setDuration(parseInt(e.target.value||'60',10))}/>
        </label>
        <button className="primary" onClick={start}>Start</button>
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
        
          <div className="row" style={{ display: 'flex', justifyContent: 'space-between' }}>
            <button className="primary" onClick={start}>Next</button>
            <button className="primary" style={{ backgroundColor: 'red' }} onClick={()=>onExit && onExit()}>End </button>
          </div>
        </div>
      )}
    </div>
  );
}

import React, { useEffect, useMemo, useState } from 'react';
import Editor from '@monaco-editor/react';
import { getProChallenge, submitProChallenge, requestProHint } from '../api';

const TOPIC_OPTIONS = [
  { value: '', label: 'Any Topic' },
  { value: 'channels', label: 'Channels' },
  { value: 'context', label: 'Context' },
  { value: 'io', label: 'I/O' },
  { value: 'http', label: 'HTTP' },
  { value: 'generics', label: 'Generics' },
  { value: 'perf', label: 'Performance' },
  { value: 'clean-code', label: 'Clean Code' },
  { value: 'configuration', label: 'Configuration' },
  { value: 'errors', label: 'Errors' },
  { value: 'strings', label: 'Strings' },
];

const DIFFICULTY_OPTIONS = [
  { value: 'advanced', label: 'Advanced' },
  { value: 'medium', label: 'Medium' },
  { value: 'any', label: 'Any Level' },
];

function readableTopics(topics = []) {
  if (!topics.length) return 'Advanced Go';
  return topics.map((t) => t.replace(/[-_]/g, ' ')).join(', ');
}

export default function ProModeView({
  coins,
  xp,
  onCoinsChange = () => {},
  onXpChange = () => {},
  onExit = () => {},
}) {
  const [topic, setTopic] = useState('');
  const [difficulty, setDifficulty] = useState('advanced');
  const [challenge, setChallenge] = useState(null);
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [running, setRunning] = useState(false);
  const [hintBusy, setHintBusy] = useState(false);
  const [banner, setBanner] = useState(null);
  const [output, setOutput] = useState('');
  const [failures, setFailures] = useState([]);
  const [hints, setHints] = useState([]);
  const [error, setError] = useState('');

  const topicLabel = useMemo(() => {
    const current = TOPIC_OPTIONS.find((t) => t.value === topic);
    return current ? current.label : 'Any Topic';
  }, [topic]);

  const difficultyLabel = useMemo(() => {
    const current = DIFFICULTY_OPTIONS.find((d) => d.value === difficulty);
    return current ? current.label : 'Advanced';
  }, [difficulty]);

  useEffect(() => {
    loadChallenge('', 'advanced');
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function loadChallenge(nextTopic = topic, nextDifficulty = difficulty) {
    const resolvedTopic = nextTopic ?? '';
    const resolvedDifficulty = nextDifficulty ?? 'advanced';
    setLoading(true);
    setRunning(false);
    setHintBusy(false);
    setBanner(null);
    setOutput('');
    setFailures([]);
    setHints([]);
    setError('');
    try {
      const data = await getProChallenge({
        topic: resolvedTopic || undefined,
        difficulty: resolvedDifficulty === 'any' ? undefined : resolvedDifficulty,
      });
      setChallenge(data);
      setCode(data?.starter?.code || '');
      setTopic(resolvedTopic);
      setDifficulty(resolvedDifficulty);
    } catch (err) {
      setError(err?.message || 'Failed to load challenge');
      setChallenge(null);
    } finally {
      setLoading(false);
    }
  }

  async function handleRunTests() {
    if (!challenge || running) return;
    setRunning(true);
    setBanner(null);
    try {
      const res = await submitProChallenge({ id: challenge.id, code });
      const combined = [res.stdout, res.stderr].filter(Boolean).join('\n\n').trim();
      setOutput(combined);
      setFailures(res.failures || []);
      if (res.passed) {
        setBanner({
          type: 'ok',
          text: res.message || 'All tests passed!',
          reward: challenge.reward,
        });
        if (typeof res.coinsTotal === 'number' && onCoinsChange) {
          onCoinsChange(res.coinsTotal);
        }
        if (typeof res.xpTotal === 'number' && onXpChange) {
          onXpChange(res.xpTotal);
        }
      } else {
        setBanner({
          type: 'bad',
          text: 'Some tests failed. Inspect the output below and iterate.',
        });
      }
    } catch (err) {
      setBanner({
        type: 'bad',
        text: err?.message || 'Submission failed.',
      });
    } finally {
      setRunning(false);
    }
  }

  async function handleHint() {
    if (!challenge) return;
    setHintBusy(true);
    try {
      const res = await requestProHint(challenge.id);
      if (typeof res.coinsTotal === 'number' && onCoinsChange) {
        onCoinsChange(res.coinsTotal);
      }
      if (typeof res.xpTotal === 'number' && onXpChange) {
        onXpChange(res.xpTotal);
      }
      if (res.hint) {
        setHints((prev) => [...prev, res.hint]);
      }
    } catch (err) {
      setBanner({
        type: 'bad',
        text: err?.message || 'Unable to fetch hint.',
      });
    } finally {
      setHintBusy(false);
    }
  }

  const canRun = Boolean(challenge) && !loading && !running;
  const canNext = Boolean(challenge) && banner?.type === 'ok' && !loading && !running;

  return (
    <div className="card pro-mode-card">
      <div className="nav">
        <button className="home" onClick={()=>onExit && onExit()}>Go Home</button>
      </div>

      <div className="pro-mode-controls">
        <label className="badge">
          <select className="badge-select" value={topic} onChange={(e) => setTopic(e.target.value)}>
            {TOPIC_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
        </label>
        <label className="badge">
          <select className="badge-select" value={difficulty} onChange={(e) => setDifficulty(e.target.value)}>
            {DIFFICULTY_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
        </label>
        <button
          className="badge badge-button"
          onClick={() => loadChallenge(topic, difficulty)}
          disabled={loading || running}
        >
          {loading ? 'Loading...' : '▶ New Challenge'}
        </button>
      </div>

      {error && <div className="banner-bad">{error}</div>}
      {loading && !challenge && !error && (
        <div className="muted">Loading challenge...</div>
      )}

      {challenge && !error && (
        <>
          <div className="pro-mode-details">
            <h3>{challenge.title}</h3>
            <p>{challenge.description}</p>
            <div className="badge-row">
              <span className="badge">Difficulty: {challenge.difficulty}</span>
              <span className="badge">Topics: {readableTopics(challenge.topics)}</span>
              <span className="badge">Rewards: +{challenge.reward.coins} coins +{challenge.reward.xp} XP</span>
            </div>
          </div>

          <div className="editor-container">
            <Editor
              height="100%"
              defaultLanguage="go"
              language="go"
              theme="vs-dark"
              value={code}
              onChange={(value) => setCode(value ?? '')}
              options={{
                fontSize: 14,
                minimap: { enabled: false },
                automaticLayout: true,
              }}
            />
          </div>

          <div className="run-actions">
            <button className="badge badge-button" onClick={handleRunTests} disabled={!canRun}>
              {running ? 'Running...' : 'Run Tests'}
            </button>
            <button
              className="badge badge-button"
              onClick={handleHint}
              disabled={!challenge || hintBusy || loading}
            >
              {hintBusy ? 'Fetching...' : 'Hint (-2 coins)'}
            </button>
            <button
              className="badge badge-button"
              onClick={() => loadChallenge(topic, difficulty)}
              disabled={!canNext}
            >
              ▶ Next
            </button>
          </div>

          {banner && (
            <div className={banner.type === 'ok' ? 'banner-ok' : 'banner-bad'}>
              {banner.text}
              {banner.type === 'ok' && banner.reward && (
                <div className="muted">+{banner.reward.coins} coins +{banner.reward.xp} XP awarded.</div>
              )}
            </div>
          )}

          <div className="console">
            {failures.length > 0 ? (
              failures.map((f, idx) => (
                <div key={`${f.name || 'failure'}-${idx}`} className="console-section">
                  <strong>FAIL {f.name || 'Test'}</strong>
                  <pre>{f.output}</pre>
                </div>
              ))
            ) : output ? (
              <pre>{output}</pre>
            ) : (
              <span className="muted">Run tests to view output. stdout/stderr will appear here.</span>
            )}
          </div>

          <div className="hints-block">
            <h4>Hints</h4>
            {hints.length === 0 ? (
              <p className="muted">Hints unlock sequentially. Each one costs 2 coins.</p>
            ) : (
              <ul>
                {hints.map((hint, idx) => (
                  <li key={`${idx}-${hint.slice(0, 8)}`}>{hint}</li>
                ))}
              </ul>
            )}
          </div>
        </>
      )}
    </div>
  );
}

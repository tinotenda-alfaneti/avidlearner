import React, { useState } from 'react';
import { login, signup } from '../api';

export default function AccountModal({ initialMode = 'login', onClose, onAuth }) {
  const [mode, setMode] = useState(initialMode);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [leaderboardOptIn, setLeaderboardOptIn] = useState(true);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  const isSignup = mode === 'signup';
  const canSubmit = username.trim().length >= 3 && password.length >= 8 && (!isSignup || leaderboardOptIn);

  async function handleSubmit(e) {
    e.preventDefault();
    if (!canSubmit) return;
    setBusy(true);
    setError('');
    try {
      const user = isSignup
        ? await signup({ username: username.trim(), password, leaderboardOptIn })
        : await login({ username: username.trim(), password });
      onAuth?.(user);
      onClose?.();
    } catch (err) {
      setError(err?.message || 'Unable to continue');
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-card auth-modal" onClick={(e) => e.stopPropagation()}>
        <button className="modal-close-btn" onClick={onClose}>×</button>
        <h2 className="leaderboard-title">{isSignup ? 'Create Account' : 'Sign In'}</h2>
        <p className="modal-subtitle">
          {isSignup ? 'Save your progress and appear on the global leaderboard.' : 'Welcome back. Pick up where you left off.'}
        </p>

        <div className="auth-tabs">
          <button className={`tab-btn ${!isSignup ? 'active' : ''}`} onClick={() => setMode('login')}>Sign In</button>
          <button className={`tab-btn ${isSignup ? 'active' : ''}`} onClick={() => setMode('signup')}>Create Account</button>
        </div>

        <form onSubmit={handleSubmit} className="auth-form">
          <div className="auth-field">
            <label className="auth-label" htmlFor="auth-username">Username</label>
            <input
              id="auth-username"
              className="auth-input"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="your_name"
              autoComplete="username"
              minLength={3}
              maxLength={30}
              disabled={busy}
              required
            />
            <span className="form-hint">3–30 chars, letters/numbers/._- only.</span>
          </div>

          <div className="auth-field">
            <label className="auth-label" htmlFor="auth-password">Password</label>
            <input
              id="auth-password"
              className="auth-input"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="At least 8 characters"
              autoComplete={isSignup ? 'new-password' : 'current-password'}
              minLength={8}
              disabled={busy}
              required
            />
          </div>

          {isSignup && (
            <label className="checkbox-row">
              <input
                type="checkbox"
                checked={leaderboardOptIn}
                onChange={(e) => setLeaderboardOptIn(e.target.checked)}
                disabled={busy}
                required
              />
              <span>I agree to appear on the global leaderboard.</span>
            </label>
          )}

          {error && <div className="banner-bad">{error}</div>}

          <div className="modal-actions auth-actions">
            <button type="submit" className="primary auth-submit" disabled={!canSubmit || busy}>
              {busy ? 'Working...' : isSignup ? 'Create Account' : 'Sign In'}
            </button>
            <button type="button" className="ghost auth-cancel" onClick={onClose} disabled={busy}>
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

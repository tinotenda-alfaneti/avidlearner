import React, { useState } from 'react';
import { submitToLeaderboard } from '../api';

export default function NameInputModal({ score, mode, category, onSuccess, onSkip }) {
  const [name, setName] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e) {
    e.preventDefault();
    if (!name.trim()) return;

    setSubmitting(true);
    try {
      await submitToLeaderboard(name.trim(), score, mode, category);
      onSuccess?.();
    } catch (error) {
      console.error('Failed to submit score:', error);
      alert('Failed to submit score. Please try again.');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="modal-overlay">
      <div className="modal-card name-input-modal" onClick={(e) => e.stopPropagation()}>
        <h2>Submit to Leaderboard</h2>
        <p className="modal-subtitle">Enter your name to save your score</p>
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <input
              type="text"
              placeholder="Your name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={30}
              autoFocus
              disabled={submitting}
              className="name-input-field"
            />
          </div>
          
          <div className="modal-actions">
            <button 
              type="submit" 
              className="primary" 
              disabled={submitting || !name.trim()}
            >
              {submitting ? 'Submitting...' : 'Submit Score'}
            </button>
            <button 
              type="button" 
              className="ghost" 
              onClick={onSkip}
              disabled={submitting}
            >
              Skip
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

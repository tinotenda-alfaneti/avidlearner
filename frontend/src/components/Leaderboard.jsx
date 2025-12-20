import React, { useEffect, useState } from 'react';
import { getLeaderboard } from '../api';

export default function Leaderboard({ onClose }) {
  const [selectedMode, setSelectedMode] = useState('quiz');
  const [entries, setEntries] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadLeaderboard();
  }, [selectedMode]);

  async function loadLeaderboard() {
    setLoading(true);
    try {
      const data = await getLeaderboard(selectedMode);
      setEntries(data || []);
    } catch (error) {
      console.error('Failed to load leaderboard:', error);
      setEntries([]);
    } finally {
      setLoading(false);
    }
  }

  const getRankDisplay = (rank) => {
    if (rank === 0) return '1st';
    if (rank === 1) return '2nd';
    if (rank === 2) return '3rd';
    return `${rank + 1}th`;
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-card leaderboard-modal" onClick={(e) => e.stopPropagation()}>
        <button className="modal-close-btn" onClick={onClose}>×</button>
        
        <h2 className="leaderboard-title">Leaderboard</h2>
        
        <div className="leaderboard-tabs">
          <button
            className={`tab-btn ${selectedMode === 'quiz' ? 'active' : ''}`}
            onClick={() => setSelectedMode('quiz')}
          >
            Quiz Mode
          </button>
          <button
            className={`tab-btn ${selectedMode === 'coding' ? 'active' : ''}`}
            onClick={() => setSelectedMode('coding')}
          >
            Coding Mode
          </button>
          <button
            className={`tab-btn ${selectedMode === 'typing' ? 'active' : ''}`}
            onClick={() => setSelectedMode('typing')}
          >
            Typing Mode
          </button>
        </div>

        <div className="leaderboard-content">
          {loading ? (
            <div className="loading-state">Loading...</div>
          ) : entries.length === 0 ? (
            <div className="empty-state">
              <p>No entries yet. Be the first!</p>
            </div>
          ) : (
            <div className="leaderboard-list">
              {entries.map((entry, index) => (
                <div 
                  key={index} 
                  className={`leaderboard-entry ${index < 3 ? 'top-3' : ''}`}
                >
                  <div className={`rank ${index < 3 ? 'top-rank' : ''}`}>
                    {getRankDisplay(index)}
                  </div>
                  <div className="player-name">{entry.name}</div>
                  <div className="score">{entry.score.toLocaleString()}</div>
                  <div className="date">
                    {new Date(entry.date).toLocaleDateString()}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="modal-actions">
          <button className="ghost" onClick={onClose}>Close</button>
        </div>
      </div>
    </div>
  );
}

import React from 'react';

function statLabel(value) {
  if (value === undefined || value === null) return 0;
  return value;
}

export default function ProfileModal({ user, onClose, onLogout, onRemoveLesson }) {
  if (!user) return null;
  const profile = user.profile || {};
  const stats = profile.stats || {};
  const savedLessons = profile.savedLessons || [];

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-card profile-modal" onClick={(e) => e.stopPropagation()}>
        <button className="modal-close-btn" onClick={onClose}>×</button>
        <h2 className="leaderboard-title">Your Profile</h2>
        <p className="modal-subtitle">
          Signed in as <strong>{user.username}</strong>
        </p>

        <div className="profile-badges">
          <div className="profile-badge">
            <span>Coins</span>
            <strong>{statLabel(profile.coins)}</strong>
          </div>
          <div className="profile-badge">
            <span>XP</span>
            <strong>{statLabel(profile.xp)}</strong>
          </div>
          <div className="profile-badge">
            <span>Quiz Streak</span>
            <strong>{statLabel(profile.quizStreak)}</strong>
          </div>
          <div className="profile-badge">
            <span>Typing Best</span>
            <strong>{statLabel(profile.typingBest)}</strong>
          </div>
          <div className="profile-badge">
            <span>Typing Streak</span>
            <strong>{statLabel(profile.typingStreak)}</strong>
          </div>
          <div className="profile-badge">
            <span>Coding Score</span>
            <strong>{statLabel(profile.codingScore)}</strong>
          </div>
        </div>

        <div className="profile-stats">
          <div className="profile-stat">Lessons read: {statLabel(stats.lessonsRead)}</div>
          <div className="profile-stat">Quiz questions: {statLabel(stats.quizzesTaken)}</div>
          <div className="profile-stat">Quiz correct: {statLabel(stats.quizCorrect)}</div>
          <div className="profile-stat">Typing sessions: {statLabel(stats.typingSessions)}</div>
          <div className="profile-stat">Coding submissions: {statLabel(stats.codingSubmissions)}</div>
          <div className="profile-stat">Coding passed: {statLabel(stats.codingPassed)}</div>
          <div className="profile-stat">
            Leaderboard opt-in: {user.leaderboardOptIn ? 'Yes' : 'No'}
          </div>
        </div>

        <div className="profile-saved">
          <h3>Saved Lessons</h3>
          {savedLessons.length === 0 ? (
            <div className="muted">No saved lessons yet.</div>
          ) : (
            <div className="saved-list">
              {savedLessons.map((lesson) => (
                <div key={`${lesson.title}-${lesson.category}`} className="saved-item">
                  <div>
                    <div className="saved-title">{lesson.title}</div>
                    <div className="saved-meta">
                      {lesson.category}
                      {lesson.source ? ` · ${lesson.source}` : ''}
                    </div>
                  </div>
                  <button className="ghost" onClick={() => onRemoveLesson?.(lesson)}>
                    Remove
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="modal-actions">
          <button className="ghost" onClick={onClose}>Close</button>
          <button className="primary" onClick={onLogout}>Sign Out</button>
        </div>
      </div>
    </div>
  );
}

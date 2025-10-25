
import React from 'react'

function categoryLabel(value) {
  if (value === 'any') return 'Any';
  if (value === 'random') return 'Random';
  return value.replace(/[-_]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

export default function Dashboard({
  coins,
  quizStreak,
  typingStreak,
  typingBest,
  categoryOptions = [],
  selectedCategory = 'any',
  onSelectCategory,
  onStartLearn,
  onStartTyping
}) {
  return (
    <div className="dashboard-stack">
      <div className="card dash-card">
        <header className="dash-card__header">
          <h2>Learn Mode</h2>
          <p>Collect coins and keep your quiz streak alive.</p>
        </header>
        <div className="badge-row">
          <span className="badge">💰 Coins: {coins}</span>
          <span className="badge">🔥 Quiz Streak: {quizStreak}</span>
        </div>
        <div className="dash-card__actions dash-card__actions--center">
          <button className="primary" onClick={onStartLearn}>Launch Learn Mode</button>
        </div>
        <div className="card-note">Tip: Pick a category or leave "Any" to rotate through everything.</div>
      </div>

      <div className="card dash-card">
        <header className="dash-card__header">
          <h2>Typing Mode</h2>
          <p>Beat your personal-best streak and accuracy.</p>
        </header>
        <div className="badge-row">
          <span className="badge">⌨ Typing Streak: {typingStreak}</span>
          <span className="badge">🏆 Typing Best: {typingBest}</span>
        </div>
        <div className="dash-card__actions dash-card__actions--center">
          <button className="primary" onClick={onStartTyping}>Launch Typing Mode</button>
        </div>
        <div className="card-note">Your streak resets when you exit a session early. Finish strong to keep improving.</div>
      </div>
    </div>
  )
}

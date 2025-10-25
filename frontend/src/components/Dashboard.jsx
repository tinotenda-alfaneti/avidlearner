
import React from 'react'

function categoryLabel(value) {
  if (value === 'any') return 'Any';
  if (value === 'random') return 'Random';
  return value.replace(/[-_]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

export default function Dashboard({
  coins,
  streak,
  categoryOptions = [],
  selectedCategory = 'any',
  onSelectCategory,
  onStartLearn,
  onStartTyping
}) {
  return (
    <div className="card">
      <div className="row" style={{justifyContent:'space-between', alignItems:'center'}}>
        <div style={{display:'flex', gap:'8px', alignItems:'center'}}>
          <span className="badge">💰 Coins: {coins}</span>
          <span className="badge">🔥 Streak: {streak}</span>
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
        <div className="row">
          <button className="primary" onClick={onStartLearn}>Learn Mode</button>
          <button className="primary" style={{ backgroundColor: '#7d89b0' }} onClick={onStartTyping}>Typing Mode</button>
        </div>
      </div>
      <div className="footer">Tip: Pick a category above or leave "Any" to rotate through everything.</div>
    </div>
  )
}

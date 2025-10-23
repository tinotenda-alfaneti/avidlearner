
import React from 'react'

export default function Dashboard({ coins, streak, onStartLearn, onStartTyping }) {
  return (
    <div className="card">
      <div className="row" style={{justifyContent:'space-between', alignItems:'center'}}>
        <div style={{display:'flex', gap:'8px', alignItems:'center'}}>
          <span className="badge">💰 Coins: {coins}</span>
          <span className="badge">🔥 Streak: {streak}</span>
        </div>
        <div className="row">
          <button className="primary" onClick={onStartLearn}>Learn Mode</button>
          <button className="primary" style={{ backgroundColor: '#7d89b0' }} onClick={onStartTyping}>Typing Mode</button>
        </div>
      </div>
      <div className="footer">Tip: Add to your home screen on iOS/Android for a near-app experience.</div>
    </div>
  )
}

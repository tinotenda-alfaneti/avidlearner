
import React from 'react'

export default function ResultView({ correct, earned, total, message, onContinue, onExit }) {
  return (
    <div className="card">
      <div className="nav">
      <button className="home" onClick={()=>onExit && onExit()}>Go Home</button>
    </div>
      <div style={{color:'#7d89b0', fontWeight:600}}>🏅 Results</div>
      <div className="prompt" style={{marginTop:8}}><strong>{message}</strong></div>
      <div className="row" style={{marginTop:10}}>
        <span className="badge">{correct ? '✅ Correct' : '❌ Incorrect'}</span>
        <span className="badge">+{earned} coins</span>
        <span className="badge">Total: {total}</span>
      </div>
      <div style={{marginTop:12}}>
        <button className="primary" onClick={onContinue}>Continue</button>
      </div>
    </div>
  )
}

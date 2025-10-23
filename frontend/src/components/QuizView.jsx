import React from 'react';

export default function QuizView({ question, options, index, total, onAnswer, onExit }) {
  return (
    <div className="card">
      <div className="nav">
        <button className="home" onClick={()=>onExit && onExit()}>Go Home</button>
      </div>

      <div style={{color:'#7d89b0', fontWeight:600}}>🧩 Quiz {index} / {total}</div>
      <div className="prompt" style={{marginTop:8}}>{question}</div>
      <div style={{marginTop:10}}>
        {options?.map((o,i)=>(
          <button key={i} className="option" onClick={()=>onAnswer(i)}>
            {String.fromCharCode(65+i)}. {o}
          </button>
        ))}
      </div>
    </div>
  );
}

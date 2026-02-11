import React, { useEffect, useState } from 'react';
import { getTLDR } from '../api';

export default function TLDRFeed({ onClose }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    let mounted = true;
    setLoading(true);
    getTLDR('tech')
      .then(d => { if (mounted) { setItems(d || []); setError(null); } })
      .catch(err => { if (mounted) setError(err.message || String(err)); })
      .finally(() => { if (mounted) setLoading(false); });
    return () => { mounted = false; };
  }, []);

  return (
    <div className="tldr-section">
      <div className="section-header">
        <h2 className="section-title">TL;DR — Quick Summaries</h2>
        <div style={{marginLeft: 'auto'}}>
          <button className="mode-btn secondary" onClick={onClose}>Close</button>
        </div>
      </div>

      {loading && <div style={{padding:20}}>Loading TLDRs…</div>}

      {error && !loading && (
        <div className="news-error" style={{padding:20}}>{error}</div>
      )}

      {!loading && items && items.length > 0 && (
        <div className="news-grid">
          {items.map(it => (
            <div key={it.id || it.title} className="news-card">
              <a href={it.url || '#'} target="_blank" rel="noopener noreferrer" className="news-title">{it.title}</a>
              <div style={{marginTop:8, color:'#2b2b2b'}}>{it.summary || ''}</div>
              {it.tags && (
                <div className="news-tags" style={{marginTop:10}}>
                  {Array.isArray(it.tags) ? it.tags.slice(0,4).map(tag => (
                    <span key={tag} className="news-tag">#{tag}</span>
                  )) : null}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

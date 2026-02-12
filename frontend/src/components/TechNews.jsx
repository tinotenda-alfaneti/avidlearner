import React, { useState, useEffect } from 'react';

const NEWS_SOURCES = {
  HACKERNEWS: 'HackerNews',
  DEVTO: 'Dev.to',
  ARSTECHNICA: 'Ars Technica',
  TLDR: 'TLDR'
};

// Cache keys for each news source
const CACHE_KEYS = {
  [NEWS_SOURCES.HACKERNEWS]: 'avidlearner_news_hackernews',
  [NEWS_SOURCES.DEVTO]: 'avidlearner_news_devto',
  [NEWS_SOURCES.ARSTECHNICA]: 'avidlearner_news_arstechnica',
  [NEWS_SOURCES.TLDR]: 'avidlearner_news_tldr'

};

// Helper functions for localStorage
const saveToCache = (key, data) => {
  try {
    localStorage.setItem(key, JSON.stringify({
      data,
      timestamp: Date.now()
    }));
  } catch (e) {
    console.warn('Failed to cache news:', e);
  }
};

const getFromCache = (key) => {
  try {
    const cached = localStorage.getItem(key);
    if (!cached) return null;
    return JSON.parse(cached);
  } catch (e) {
    console.warn('Failed to read cached news:', e);
    return null;
  }
};

const isOnline = () => navigator.onLine;

export default function TechNews() {
  const [news, setNews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [activeSource, setActiveSource] = useState(NEWS_SOURCES.HACKERNEWS);
  const [isOffline, setIsOffline] = useState(false);
  const [tldrGroups, setTldrGroups] = useState(null);
  const [tldrCategory, setTldrCategory] = useState('all');

  useEffect(() => {
    fetchNews();
  }, [activeSource, tldrCategory]);

  const fetchNews = async () => {
    setLoading(true);
    setError(null);
    setIsOffline(false);
    
    const cacheKey = activeSource === NEWS_SOURCES.TLDR ? `${CACHE_KEYS[activeSource]}_${tldrCategory}` : CACHE_KEYS[activeSource];
    
    // Check if offline
    if (!isOnline()) {
      const cached = getFromCache(cacheKey);
      if (cached) {
        setNews(cached.data);
        setIsOffline(true);
      } else {
        setError('You are offline and no cached news is available.');
      }
      setLoading(false);
      return;
    }
    
    try {
      let articles = [];
      
      switch (activeSource) {
        case NEWS_SOURCES.HACKERNEWS:
          articles = await fetchHackerNews();
          setTldrGroups(null);
          break;
        case NEWS_SOURCES.DEVTO:
          articles = await fetchDevTo();
          setTldrGroups(null);
          break;
        case NEWS_SOURCES.TLDR:
          // TLDR: fetch selected category (or aggregated 'all')
          const groups = await fetchTLDR(tldrCategory);
          // if single category, normalize to grouped object for rendering
          if (tldrCategory && tldrCategory !== 'all' && Array.isArray(groups)) {
            const map = {};
            map[tldrCategory] = groups;
            setTldrGroups(map);
            saveToCache(cacheKey, map);
          } else {
            setTldrGroups(groups);
            saveToCache(cacheKey, groups);
          }
          setNews([]);
          return;
        case NEWS_SOURCES.ARSTECHNICA:
          articles = await fetchArsTechnica();
          setTldrGroups(null);
          break;
        default:
          articles = await fetchHackerNews();
          setTldrGroups(null);
      }

      // Save to cache
      saveToCache(cacheKey, articles);
      setNews(articles);
    } catch (err) {
      console.error('Error fetching news:', err);
      
      // Try to load from cache on error
      const cached = getFromCache(cacheKey);
      if (cached) {
        setNews(cached.data);
        setError('Using cached news. Unable to fetch latest updates.');
      } else {
        setError('Failed to load news. Please try again later.');
      }
    } finally {
      setLoading(false);
    }
  };

  const fetchArsTechnica = async () => {
    const res = await fetch(`/api/news?source=arstechnica`);
    if (!res.ok) {
      throw new Error('Failed to fetch Ars Technica feed');
    }
    const items = await res.json();
    return (items || []).slice(0, 10).map(it => ({
      id: it.id,
      title: it.title,
      url: it.url,
      points: it.points || 0,
      author: it.author || '',
      time: it.time || 0,
      comments: it.comments || 0,
      tags: it.tags || []
    }));
  };

  const TLDR_CATEGORIES = ['all', 'tech', 'ai', 'devops', 'dev'];

  const fetchTLDR = async (category = 'all') => {
    const res = await fetch(`/api/news?source=tldr&category=${encodeURIComponent(category)}`);
    if (!res.ok) {
      throw new Error('Failed to fetch TLDR feeds');
    }
    const data = await res.json();
    return data;
  };

  const fetchHackerNews = async () => {
    // Fetch top 10 stories from HackerNews
    const topStoriesRes = await fetch('https://hacker-news.firebaseio.com/v0/topstories.json');
    const topStoryIds = await topStoriesRes.json();
    
    // Get first 10 stories
    const storyPromises = topStoryIds.slice(0, 10).map(id =>
      fetch(`https://hacker-news.firebaseio.com/v0/item/${id}.json`).then(r => r.json())
    );
    
    const stories = await Promise.all(storyPromises);
    
    return stories.map(story => ({
      id: story.id,
      title: story.title,
      url: story.url || `https://news.ycombinator.com/item?id=${story.id}`,
      points: story.score,
      author: story.by,
      time: story.time,
      comments: story.descendants || 0
    }));
  };

  const fetchDevTo = async () => {
    const res = await fetch('https://dev.to/api/articles?top=7');
    const articles = await res.json();
    
    return articles.slice(0, 10).map(article => ({
      id: article.id,
      title: article.title,
      url: article.url,
      points: article.public_reactions_count,
      author: article.user.name,
      time: new Date(article.published_at).getTime() / 1000,
      comments: article.comments_count,
      tags: article.tag_list
    }));
  };


  const formatTimeAgo = (timestamp) => {
    const now = Date.now() / 1000;
    const diff = now - timestamp;
    
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    return `${Math.floor(diff / 86400)}d ago`;
  };

  return (
    <div className="tech-news-section">
      <div className="tech-news-header">
        <h2 className="section-title">Tech News</h2>
        <div className="news-source-tabs">
          {Object.values(NEWS_SOURCES).map(source => (
            <button
              key={source}
              className={`news-tab ${activeSource === source ? 'active' : ''}`}
              onClick={() => setActiveSource(source)}
            >
              {source}
            </button>
          ))}
        </div>
        {activeSource === NEWS_SOURCES.TLDR && (
          <div style={{marginTop:8}}>
            <label className="badge">
              <select className="badge-select" value={tldrCategory} onChange={(e) => setTldrCategory(e.target.value)}>
                {['all','tech','ai','devops','dev'].map(c => (
                  <option key={c} value={c}>{c.toUpperCase()}</option>
                ))}
              </select>
            </label>
          </div>
        )}
      </div>

      {isOffline && !loading && !error && (
        <div className="news-offline-banner">
          📡 Offline - Showing cached news
        </div>
      )}

      {loading && (
        <div className="news-loading">
          <div className="spinner"></div>
          <p>Loading tech news...</p>
        </div>
      )}
      {error && news.length === 0 && (
        <div className="news-error">
          <p>{error}</p>
          {isOnline() && <button onClick={fetchNews} className="retry-btn">Retry</button>}
        </div>
      )}

      {/* TLDR grouped view */}
      {!loading && activeSource === NEWS_SOURCES.TLDR && tldrGroups && (
        <div>
          {Object.keys(tldrGroups).map(cat => (
            <div key={cat} style={{marginBottom:20}}>
              <h3 style={{marginTop:12, marginBottom:8}}>{cat.toUpperCase()}</h3>
              <div className="news-grid">
                {(tldrGroups[cat] || []).slice(0,10).map(article => (
                  <div key={article.id || article.title} className="news-card">
                    <a href={article.url || '#'} target="_blank" rel="noopener noreferrer" className="news-title">{article.title}</a>
                    <div className="news-meta">
                      <span className="news-points">▲ {article.points || 0}</span>
                      <span className="news-author">by {article.author || ''}</span>
                      <span className="news-time">{formatTimeAgo(article.time || 0)}</span>
                      <span className="news-comments">💬 {article.comments || 0}</span>
                    </div>
                    {article.summary && (
                      <div style={{marginTop:8,color:'#2b2b2b'}}>{article.summary}</div>
                    )}
                    {article.tags && (
                      <div className="news-tags">
                        {article.tags.slice(0,3).map(tag => (
                          <span key={tag} className="news-tag">#{tag}</span>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Default flat news view */}
      {!loading && news.length > 0 && (
        <>
          {error && (
            <div className="news-error" style={{ marginBottom: '20px' }}>
              <p>{error}</p>
              {isOnline() && <button onClick={fetchNews} className="retry-btn">Retry</button>}
            </div>
          )}
          <div className="news-grid">
            {news.map((article) => (
              <div key={article.id} className="news-card">
                <a 
                  href={article.url} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="news-title"
                >
                  {article.title}
                </a>
                <div className="news-meta">
                  <span className="news-points">▲ {article.points}</span>
                  <span className="news-author">by {article.author}</span>
                  <span className="news-time">{formatTimeAgo(article.time)}</span>
                  <span className="news-comments">💬 {article.comments}</span>
                </div>
                {article.summary && (
                  <div style={{marginTop:8,color:'#2b2b2b'}}>{article.summary}</div>
                )}

                {article.tags && (
                  <div className="news-tags">
                    {article.tags.slice(0, 3).map(tag => (
                      <span key={tag} className="news-tag">#{tag}</span>
                    ))}
                  </div>
                )}
                
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}

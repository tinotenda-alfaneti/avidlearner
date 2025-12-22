import React, { useState, useEffect } from 'react';

const NEWS_SOURCES = {
  HACKERNEWS: 'HackerNews',
  DEVTO: 'Dev.to',
  REDDIT: 'Reddit'
};

export default function TechNews() {
  const [news, setNews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [activeSource, setActiveSource] = useState(NEWS_SOURCES.HACKERNEWS);

  useEffect(() => {
    fetchNews();
  }, [activeSource]);

  const fetchNews = async () => {
    setLoading(true);
    setError(null);
    
    try {
      let articles = [];
      
      switch (activeSource) {
        case NEWS_SOURCES.HACKERNEWS:
          articles = await fetchHackerNews();
          break;
        case NEWS_SOURCES.DEVTO:
          articles = await fetchDevTo();
          break;
        case NEWS_SOURCES.REDDIT:
          articles = await fetchReddit();
          break;
        default:
          articles = await fetchHackerNews();
      }
      
      setNews(articles);
    } catch (err) {
      console.error('Error fetching news:', err);
      setError('Failed to load news. Please try again later.');
    } finally {
      setLoading(false);
    }
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

  const fetchReddit = async () => {
    // Fetch from r/programming (no auth required for public posts)
    const res = await fetch('https://www.reddit.com/r/programming/hot.json?limit=10');
    const data = await res.json();
    
    return data.data.children.map(post => ({
      id: post.data.id,
      title: post.data.title,
      url: post.data.url,
      points: post.data.ups,
      author: post.data.author,
      time: post.data.created_utc,
      comments: post.data.num_comments,
      subreddit: post.data.subreddit
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
      </div>

      {loading && (
        <div className="news-loading">
          <div className="spinner"></div>
          <p>Loading tech news...</p>
        </div>
      )}

      {error && (
        <div className="news-error">
          <p>{error}</p>
          <button onClick={fetchNews} className="retry-btn">Retry</button>
        </div>
      )}

      {!loading && !error && (
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
              {article.tags && (
                <div className="news-tags">
                  {article.tags.slice(0, 3).map(tag => (
                    <span key={tag} className="news-tag">#{tag}</span>
                  ))}
                </div>
              )}
              {article.subreddit && (
                <span className="news-subreddit">r/{article.subreddit}</span>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

import React, { useState } from 'react';
import { generateAILesson } from '../api';

export default function AILessonGenerator({ categories, onLessonGenerated, onCancel }) {
  const [topic, setTopic] = useState('');
  const [category, setCategory] = useState('general');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleGenerate = async (e) => {
    e.preventDefault();
    
    if (!topic.trim()) {
      setError('Please enter a topic');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const result = await generateAILesson(category, topic);
      onLessonGenerated(result);
    } catch (err) {
      setError(err.message || 'Failed to generate lesson');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="ai-generator-container">
      <div className="ai-generator-card">
        <h2>🤖 Generate AI Lesson</h2>
        <p className="ai-description">
          Create a custom lesson on any software engineering topic using AI.
        </p>

        <form onSubmit={handleGenerate}>
          <div className="form-group">
            <label htmlFor="topic">Topic</label>
            <input
              id="topic"
              type="text"
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              placeholder="e.g., Database Indexing Strategies"
              disabled={loading}
              autoFocus
            />
          </div>

          <div className="form-group">
            <label htmlFor="category">Category</label>
            <select
              id="category"
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              disabled={loading}
            >
              <option value="general">General</option>
              <option value="architecture">Architecture</option>
              <option value="clean-code">Clean Code</option>
              <option value="databases">Databases</option>
              <option value="performance">Performance</option>
              <option value="system-design">System Design</option>
              <option value="reliability">Reliability</option>
              <option value="effective-go">Effective Go</option>
              <option value="golang">Golang</option>
              <option value="testing">Testing</option>
              {categories.map(cat => (
                <option key={cat} value={cat}>{cat}</option>
              ))}
            </select>
          </div>

          {error && (
            <div className="error-message">
              ⚠️ {error}
            </div>
          )}

          <div className="button-group">
            <button
              type="submit"
              className="btn-primary"
              disabled={loading}
            >
              {loading ? '⏳ Generating...' : '✨ Generate Lesson'}
            </button>
            <button
              type="button"
              className="btn-secondary"
              onClick={onCancel}
              disabled={loading}
            >
              Cancel
            </button>
          </div>
        </form>

        <div className="ai-tips">
          <p><strong>💡 Tips:</strong></p>
          <ul>
            <li>Be specific with your topic for better results</li>
            <li>Generation takes 10-20 seconds</li>
            <li>AI lessons are experimental - verify content</li>
          </ul>
        </div>
      </div>
    </div>
  );
}

import React from 'react';
import ModeCard from './ModeCard';
import TechNews from './TechNews';

function categoryLabel(value) {
  if (value === 'any') return 'Any';
  if (value === 'random') return 'Random';
  return value.replace(/[-_]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

export default function Dashboard({
  coins,
  xp = 0,
  quizStreak,
  typingStreak,
  typingBest,
  categoryOptions = [],
  selectedCategory = 'any',
  onSelectCategory = () => {},
  onStartLearn,
  onStartAI,
  onStartTyping,
  onStartProMode,
  onOpenLeaderboard
}) {
  const learnBadges = [
    { label: 'Coins', value: coins },
    { label: 'Quiz Streak', value: quizStreak },
  ];

  const codingBadges = [
    { label: 'Coins', value: coins },
    { label: 'XP', value: xp },
  ];

  const typingBadges = [
    { label: 'Typing Streak', value: typingStreak },
    { label: 'Typing Best', value: typingBest },
  ];

  const learnActions = [
    { label: 'Launch Learn Mode', onClick: onStartLearn },
  ];

  // Add AI option if enabled
  if (onStartAI) {
    learnActions.push({ label: 'Generate AI Lesson', onClick: onStartAI });
  }

  return (
    <>
      <div className="hero-section">
        <h1 className="hero-title">AvidLearner</h1>
        <p className="hero-subtitle">Master Software Engineering Through Practice</p>
        <div className="hero-stats">
          <div className="stat-item">
            <div className="stat-value">{coins.toLocaleString()}</div>
            <div className="stat-label">Coins</div>
          </div>
          <div className="stat-item">
            <div className="stat-value">{xp.toLocaleString()}</div>
            <div className="stat-label">Total XP</div>
          </div>
          <div className="stat-item">
            <div className="stat-value">{quizStreak}</div>
            <div className="stat-label">Quiz Streak</div>
          </div>
        </div>
      </div>

      <div className="mode-selection">
        <h2 className="section-title">Choose Your Mode</h2>
        <div className="mode-grid">
          <div className="mode-card-large" onClick={onStartLearn}>
            <div className="mode-card-header">
              <h3>Learn Mode</h3>
              <div className="mode-difficulty easy">Beginner Friendly</div>
            </div>
            <p className="mode-description">
              Study software engineering concepts and test your knowledge with quizzes. 
              Earn coins for every correct answer.
            </p>
            <div className="mode-stats-row">
              <div className="mode-stat">
                <span className="label">Current Streak</span>
                <span className="value">{quizStreak}</span>
              </div>
              <div className="mode-stat">
                <span className="label">Coins Earned</span>
                <span className="value">{coins}</span>
              </div>
            </div>
            <button className="mode-btn primary">Start Learning</button>
            {onStartAI && (
              <button 
                className="mode-btn secondary" 
                onClick={(e) => { e.stopPropagation(); onStartAI(); }}
              >
                AI Custom Lesson
              </button>
            )}
          </div>

          <div className="mode-card-large" onClick={onStartProMode}>
            <div className="mode-card-header">
              <h3>Coding Mode</h3>
              <div className="mode-difficulty hard">Advanced</div>
            </div>
            <p className="mode-description">
              Tackle realistic Go challenges and clean-code exercises. 
              Pass hidden tests to earn XP and improve your skills.
            </p>
            <div className="mode-stats-row">
              <div className="mode-stat">
                <span className="label">Total XP</span>
                <span className="value">{xp}</span>
              </div>
              <div className="mode-stat">
                <span className="label">Coins</span>
                <span className="value">{coins}</span>
              </div>
            </div>
            <button className="mode-btn primary">Start Coding</button>
          </div>

          <div className="mode-card-large" onClick={onStartTyping}>
            <div className="mode-card-header">
              <h3>Typing Mode</h3>
              <div className="mode-difficulty medium">Intermediate</div>
            </div>
            <p className="mode-description">
              Improve your typing speed and accuracy with programming concepts. 
              Beat your personal best streak.
            </p>
            <div className="mode-stats-row">
              <div className="mode-stat">
                <span className="label">Current Streak</span>
                <span className="value">{typingStreak}</span>
              </div>
              <div className="mode-stat">
                <span className="label">Best Streak</span>
                <span className="value">{typingBest}</span>
              </div>
            </div>
            <button className="mode-btn primary">Start Typing</button>
          </div>
        </div>
      </div>

      <div className="leaderboard-section">
        <div className="leaderboard-banner" onClick={onOpenLeaderboard}>
          <div className="leaderboard-banner-content">
            <h3>Global Leaderboard</h3>
            <p>Compete with learners worldwide. See where you rank!</p>
          </div>
          <button className="leaderboard-view-btn">View Leaderboard</button>
        </div>
      </div>

      <TechNews />
    </>
  );
}

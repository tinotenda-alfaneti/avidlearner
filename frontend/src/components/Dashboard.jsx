import React from 'react';
import ModeCard from './ModeCard';

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
  onStartTyping,
  onStartProMode
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

  return (
    <div className="dashboard-grid">
      <ModeCard
        title="Learn Mode"
        description="Collect coins by studying lessons and answering quizzes."
        badges={learnBadges}
        actions={[
          { label: 'Launch Learn Mode', onClick: onStartLearn },
        ]}
        footer='Tip: Pick a category or leave "Any" to rotate through everything.'
      />

      <ModeCard
        title="Coding Mode"
        description="Tackle realistic Go challenges, earn XP, and sharpen clean-code habits."
        badges={codingBadges}
        actions={[
          { label: 'Launch Coding Mode', onClick: onStartProMode },
        ]}
        footer="Passing hidden tests grants XP and coins. Hints cost coins."
      />

      <ModeCard
        title="Typing Mode"
        description="Beat your personal-best typing streak and accuracy."
        badges={typingBadges}
        actions={[
          { label: 'Launch Typing Mode', onClick: onStartTyping },
        ]}
        footer="Your streak resets when you exit a session early. Finish strong to keep improving."
      />
    </div>
  );
}

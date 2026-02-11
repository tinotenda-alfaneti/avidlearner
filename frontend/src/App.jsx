import React, { useEffect, useRef, useState } from 'react';
import './styles.css';
import Dashboard from './components/Dashboard';
import LessonView from './components/LessonView';
import QuizView from './components/QuizView';
import ResultView from './components/ResultView';
import TypingView from './components/TypingView';
import ProModeView from './components/ProModeView';
import AILessonGenerator from './components/AILessonGenerator';
import Leaderboard from './components/Leaderboard';
import NameInputModal from './components/NameInputModal';
import OfflineIndicator from './components/OfflineIndicator';
import { getLessons, getReadingLesson, addLessonToQuiz, startQuiz, answerQuiz, getAIConfig } from './api';

export default function App() {
  const [mode, setMode] = useState('dashboard'); // dashboard | reading | quiz | result | typing | promode | ai-generate
  const [coins, setCoins] = useState(parseInt(localStorage.getItem('coins')||'0',10));
  const [xp, setXp] = useState(parseInt(localStorage.getItem('xp')||'0',10));
  const [quizStreak, setQuizStreak] = useState(parseInt(localStorage.getItem('quizStreak')||'0',10));
  const [typingStreak, setTypingStreak] = useState(parseInt(localStorage.getItem('typingStreak')||'0',10));
  const [typingBest, setTypingBest] = useState(parseInt(localStorage.getItem('typingBest')||'0',10));
  const [aiEnabled, setAiEnabled] = useState(false);
  const [showLeaderboard, setShowLeaderboard] = useState(false);
  const [showNameInput, setShowNameInput] = useState(false);
  const [pendingScore, setPendingScore] = useState(null);

  const [categories, setCategories] = useState([]);
  const [selectedCategory, setSelectedCategory] = useState('any');
  const categoryRef = useRef('any');
  const [selectedSource, setSelectedSource] = useState('all');
  const sourceRef = useRef('all');
  const [currentLesson, setCurrentLesson] = useState(null);
  const [quizQuestion, setQuizQuestion] = useState(null); // {question,options,index,total}
  const [result, setResult] = useState(null);             // {correct,earned,total,message}

  useEffect(() => {
    getLessons()
      .then(d => {
        const list = (d.categories || []).slice().sort((a, b) => a.localeCompare(b));
        setCategories(list);
      })
      .catch((err) => {
        console.error('Failed to load lessons:', err);
        // Check if error is offline-related
        if (err.offline) {
          console.log('Offline mode: using cached data if available');
        }
      });
    
    // Check AI feature flag
    getAIConfig()
      .then(config => setAiEnabled(config.aiEnabled || false))
      .catch(() => setAiEnabled(false));
  }, []);
  useEffect(() => { localStorage.setItem('coins', String(coins)); }, [coins]);
  useEffect(() => { localStorage.setItem('xp', String(xp)); }, [xp]);
  useEffect(() => { localStorage.setItem('quizStreak', String(quizStreak)); }, [quizStreak]);
  useEffect(() => { localStorage.setItem('typingStreak', String(typingStreak)); }, [typingStreak]);
  useEffect(() => { localStorage.setItem('typingBest', String(typingBest)); }, [typingBest]);

  const categoryOptions = ['any', 'random', ...categories].filter((value, index, arr) => {
    if (value === 'any' || value === 'random') {
      return index === arr.indexOf(value);
    }
    return Boolean(value) && arr.indexOf(value) === index;
  });

  function resolveCategory(value) {
    const target = value ?? categoryRef.current ?? selectedCategory;
    if (target === 'random') {
      if (!categories.length) return 'any';
      const choice = categories[Math.floor(Math.random() * categories.length)];
      return choice || 'any';
    }
    return target || 'any';
  }

  async function handleSelectCategory(value) {
    categoryRef.current = value;
    setSelectedCategory(value);
    if (mode === 'reading') {
      try {
        const s = await getReadingLesson(resolveCategory(value), sourceRef.current);
        setCurrentLesson(s.lesson);
        if (typeof s.xpTotal === 'number') setXp(s.xpTotal);
      } catch (err) {
        // swallow; UI will keep previous lesson
      }
    }
  }

  async function handleSelectSource(value) {
    sourceRef.current = value;
    setSelectedSource(value);
    if (mode === 'reading') {
      try {
        const s = await getReadingLesson(resolveCategory(), value);
        setCurrentLesson(s.lesson);
        if (typeof s.xpTotal === 'number') setXp(s.xpTotal);
      } catch (err) {
        // swallow; UI will keep previous lesson
      }
    }
  }

  // ---------- Reading flow ----------
  async function startReading() {
    try {
      const s = await getReadingLesson(resolveCategory(), sourceRef.current);
      setCurrentLesson(s.lesson);
      if (typeof s.xpTotal === 'number') setXp(s.xpTotal);
      setMode('reading');
    } catch (err) {
      console.error('Failed to start reading:', err);
      if (err.offline) {
        alert('You are offline. Reading mode requires an internet connection for session tracking.');
      } else {
        alert('Failed to load lesson. Please try again.');
      }
    }
  }
  
  async function startAIGenerate() {
    setMode('ai-generate');
  }
  
  function handleAILessonGenerated(result) {
    if (result.lesson) {
      setCurrentLesson(result.lesson);
      setMode('reading');
    }
  }
  
  async function nextConcept() {
    try {
      if (currentLesson?.title) {
        await addLessonToQuiz(currentLesson.title);
      }
      const s = await getReadingLesson(resolveCategory(), sourceRef.current);
      setCurrentLesson(s.lesson);
      if (typeof s.xpTotal === 'number') setXp(s.xpTotal);
    } catch (err) {
      console.error('Failed to load next concept:', err);
      if (err.offline) {
        alert('You are offline. Cannot load next lesson.');
      }
    }
  }
  async function beginQuiz() {
    try {
      if (currentLesson?.title) {
        await addLessonToQuiz(currentLesson.title);
      }
      const q = await startQuiz();
      setQuizQuestion({ question: q.question, options: q.options, index: q.index, total: q.total });
      if (typeof q.xpTotal === 'number') setXp(q.xpTotal);
      setMode('quiz');
    } catch (err) {
      console.error('Failed to start quiz:', err);
      if (err.offline) {
        alert('You are offline. Quiz mode requires an internet connection.');
      } else {
        alert('Failed to start quiz. Please try again.');
      }
    }
  }

  // ---------- Quiz flow ----------
  async function answer(i) {
    const s = await answerQuiz(i);
    if (s.stage === 'quiz') {
      // server already advanced us to the next question
      setQuizQuestion({ question: s.question, options: s.options, index: s.index, total: s.total });
      // show a tiny toast? for correctness; coins are cumulative
      if (s.correct) setQuizStreak(x => x + 1); else setQuizStreak(0);
      if (typeof s.coinsTotal === 'number') setCoins(s.coinsTotal);
      if (typeof s.xpTotal === 'number') setXp(s.xpTotal);
      return;
    }
    // end of quiz
    setResult({ correct: s.correct, earned: s.coinsEarned, total: s.coinsTotal, message: s.message });
    if (s.correct) setQuizStreak(x => x + 1); else setQuizStreak(0);
    if (typeof s.coinsTotal === 'number') setCoins(s.coinsTotal);
    if (typeof s.xpTotal === 'number') setXp(s.xpTotal);
    setMode('result');
  }
  function doneResult() {
    // go back to reading to collect more lessons (Option B behavior)
    setMode('reading');
    // fetch a fresh concept to keep learning
    startReading();
  }

  function headerRight() {
    return <div style={{fontSize:12,color:'#7d89b0'}}>Built with Go + React</div>;
  }

  function handleTypingStats({ streak, best }) {
    if (typeof streak === 'number') {
      setTypingStreak(streak);
    }
    if (typeof best === 'number') {
      setTypingBest(prev => Math.max(prev, best));
    }
  }

  function promptLeaderboardSubmit(score, mode, category = '') {
    setPendingScore({ score, mode, category });
    setShowNameInput(true);
  }

  function handleLeaderboardSuccess() {
    setShowNameInput(false);
    setPendingScore(null);
    setShowLeaderboard(true);
  }

  return (
    <>
      <header>
        <div className="brand">AvidLearner — Software Engineering</div>
        {headerRight()}
      </header>

      <div className="container">
        {mode === 'dashboard' && (
          <Dashboard
            coins={coins}
            xp={xp}
            quizStreak={quizStreak}
            typingStreak={typingStreak}
            typingBest={typingBest}
            categoryOptions={categoryOptions}
            selectedCategory={selectedCategory}
            onSelectCategory={handleSelectCategory}
            onStartLearn={startReading}
            onStartAI={aiEnabled ? startAIGenerate : null}
            onStartTyping={()=>setMode('typing')}
            onStartProMode={()=>setMode('promode')}
            onOpenLeaderboard={() => setShowLeaderboard(true)}
          />
        )}

        {mode === 'ai-generate' && (
          <AILessonGenerator
            categories={categories}
            onLessonGenerated={handleAILessonGenerated}
            onCancel={() => setMode('dashboard')}
          />
        )}

        {mode === 'result' && result && (
          <ResultView
            correct={result.correct}
            earned={result.earned}
            total={result.total}
            message={result.message}
            onContinue={doneResult}
            onExit={()=>setMode('dashboard')}
            onSubmitToLeaderboard={() => promptLeaderboardSubmit(quizStreak, 'quiz')}
          />
        )}


        {mode === 'reading' && currentLesson && (
          <LessonView
            lesson={currentLesson}
            categoryOptions={categoryOptions}
            selectedCategory={selectedCategory}
            onSelectCategory={handleSelectCategory}
            selectedSource={selectedSource}
            onSelectSource={handleSelectSource}
            aiEnabled={aiEnabled}
            onNext={nextConcept}
            onStartQuiz={beginQuiz}
            onExit={()=>setMode('dashboard')}
          />
        )}

        {mode === 'quiz' && quizQuestion && (
          <QuizView
            question={quizQuestion.question}
            index={quizQuestion.index}
            total={quizQuestion.total}
            options={quizQuestion.options}
            onAnswer={answer}
            onExit={()=>setMode('dashboard')}
          />
        )}

        {mode === 'promode' && (
          <ProModeView
            coins={coins}
            xp={xp}
            onCoinsChange={setCoins}
            onXpChange={setXp}
            onSubmitToLeaderboard={(score) => promptLeaderboardSubmit(score, 'coding')}
            onExit={()=>setMode('dashboard')}
          />
        )}

        {mode === 'typing' && (
          <TypingView
            categoryOptions={categoryOptions}
            availableCategories={categories}
            selectedCategory={selectedCategory}
            onSelectCategory={handleSelectCategory}
            typingBest={typingBest}
            onTypingStats={handleTypingStats}
            onSubmitToLeaderboard={(wpm) => promptLeaderboardSubmit(wpm, 'typing')}
            onExit={()=>setMode('dashboard')}
          />
        )}

      </div>

      <OfflineIndicator />

      

      {showLeaderboard && <Leaderboard onClose={() => setShowLeaderboard(false)} />}
      
      {showNameInput && pendingScore && (
        <NameInputModal
          score={pendingScore.score}
          mode={pendingScore.mode}
          category={pendingScore.category}
          onSuccess={handleLeaderboardSuccess}
          onSkip={() => {
            setShowNameInput(false);
            setPendingScore(null);
          }}
        />
      )}
    </>
  );
}

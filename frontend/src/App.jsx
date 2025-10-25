import React, { useEffect, useRef, useState } from 'react';
import './styles.css';
import Dashboard from './components/Dashboard';
import LessonView from './components/LessonView';
import QuizView from './components/QuizView';
import ResultView from './components/ResultView';
import TypingView from './components/TypingView';
import { getLessons, getReadingLesson, addLessonToQuiz, startQuiz, answerQuiz } from './api';

export default function App() {
  const [mode, setMode] = useState('dashboard'); // dashboard | reading | quiz | result | typing
  const [coins, setCoins] = useState(parseInt(localStorage.getItem('coins')||'0',10));
  const [quizStreak, setQuizStreak] = useState(parseInt(localStorage.getItem('quizStreak')||'0',10));
  const [typingStreak, setTypingStreak] = useState(parseInt(localStorage.getItem('typingStreak')||'0',10));
  const [typingBest, setTypingBest] = useState(parseInt(localStorage.getItem('typingBest')||'0',10));

  const [categories, setCategories] = useState([]);
  const [selectedCategory, setSelectedCategory] = useState('any');
  const categoryRef = useRef('any');
  const [currentLesson, setCurrentLesson] = useState(null);
  const [quizQuestion, setQuizQuestion] = useState(null); // {question,options,index,total}
  const [result, setResult] = useState(null);             // {correct,earned,total,message}

  useEffect(() => {
    getLessons()
      .then(d => {
        const list = (d.categories || []).slice().sort((a, b) => a.localeCompare(b));
        setCategories(list);
      })
      .catch(()=>{});
  }, []);
  useEffect(() => { localStorage.setItem('coins', String(coins)); }, [coins]);
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
        const s = await getReadingLesson(resolveCategory(value));
        setCurrentLesson(s.lesson);
      } catch (err) {
        // swallow; UI will keep previous lesson
      }
    }
  }

  // ---------- Reading flow ----------
  async function startReading() {
    const s = await getReadingLesson(resolveCategory());
    setCurrentLesson(s.lesson);
    setMode('reading');
  }
  async function nextConcept() {
    if (currentLesson?.title) {
      await addLessonToQuiz(currentLesson.title);
    }
    const s = await getReadingLesson(resolveCategory());
    setCurrentLesson(s.lesson);
  }
  async function beginQuiz() {
    if (currentLesson?.title) {
      await addLessonToQuiz(currentLesson.title);
    }
    const q = await startQuiz();
    setQuizQuestion({ question: q.question, options: q.options, index: q.index, total: q.total });
    setMode('quiz');
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
      return;
    }
    // end of quiz
    setResult({ correct: s.correct, earned: s.coinsEarned, total: s.coinsTotal, message: s.message });
    if (s.correct) setQuizStreak(x => x + 1); else setQuizStreak(0);
    if (typeof s.coinsTotal === 'number') setCoins(s.coinsTotal);
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

  return (
    <>
      <header>
        <div className="brand">⚡ AvidLearner — Software Engineering</div>
        {headerRight()}
      </header>

      <div className="container">
        {mode === 'dashboard' && (
          <Dashboard
            coins={coins}
            quizStreak={quizStreak}
            typingStreak={typingStreak}
            typingBest={typingBest}
            categoryOptions={categoryOptions}
            selectedCategory={selectedCategory}
            onSelectCategory={handleSelectCategory}
            onStartLearn={startReading}
            onStartTyping={()=>setMode('typing')}
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
          />
        )}


        {mode === 'reading' && currentLesson && (
          <LessonView
            lesson={currentLesson}
            categoryOptions={categoryOptions}
            selectedCategory={selectedCategory}
            onSelectCategory={handleSelectCategory}
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

        {mode === 'typing' && (
          <TypingView
            categoryOptions={categoryOptions}
            availableCategories={categories}
            selectedCategory={selectedCategory}
            onSelectCategory={handleSelectCategory}
            typingBest={typingBest}
            onTypingStats={handleTypingStats}
            onExit={()=>setMode('dashboard')}
          />
        )}

      </div>
    </>
  );
}

import React, { useEffect, useState } from 'react';
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
  const [streak, setStreak] = useState(parseInt(localStorage.getItem('streak')||'0',10));

  const [categories, setCategories] = useState([]);
  const [currentLesson, setCurrentLesson] = useState(null);
  const [quizQuestion, setQuizQuestion] = useState(null); // {question,options,index,total}
  const [result, setResult] = useState(null);             // {correct,earned,total,message}

  useEffect(() => {
    getLessons().then(d => setCategories(d.categories||[])).catch(()=>{});
  }, []);
  useEffect(() => { localStorage.setItem('coins', String(coins)); }, [coins]);
  useEffect(() => { localStorage.setItem('streak', String(streak)); }, [streak]);

  // ---------- Reading flow ----------
  async function startReading() {
    const s = await getReadingLesson('any');
    setCurrentLesson(s.lesson);
    setMode('reading');
  }
  async function nextConcept() {
    const s = await getReadingLesson('any');
    await addLessonToQuiz(s.lesson.title);
    setCurrentLesson(s.lesson);
  }
  async function beginQuiz() {
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
      if (s.correct) setStreak(x => x + 1); else setStreak(0);
      if (typeof s.coinsTotal === 'number') setCoins(s.coinsTotal);
      return;
    }
    // end of quiz
    setResult({ correct: s.correct, earned: s.coinsEarned, total: s.coinsTotal, message: s.message });
    if (s.correct) setStreak(x => x + 1); else setStreak(0);
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
            streak={streak}
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
          <TypingView onExit={()=>setMode('dashboard')} />
        )}

      </div>
    </>
  );
}

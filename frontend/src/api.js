// Offline detection and error handling utilities
const OFFLINE_ERROR_MSG = 'You are offline. Please check your connection.';
const STORAGE_KEY_LESSONS = 'avidlearner_lessons_cache';

function isOnline() {
  return navigator.onLine;
}

function createOfflineError() {
  const error = new Error(OFFLINE_ERROR_MSG);
  error.offline = true;
  return error;
}

function saveToLocalStorage(key, data) {
  try {
    localStorage.setItem(key, JSON.stringify(data));
  } catch (e) {
    console.warn('Failed to save to localStorage:', e);
  }
}

function getFromLocalStorage(key) {
  try {
    const data = localStorage.getItem(key);
    return data ? JSON.parse(data) : null;
  } catch (e) {
    console.warn('Failed to read from localStorage:', e);
    return null;
  }
}

export async function getLessons() {
  if (!isOnline()) {
    const cached = getFromLocalStorage(STORAGE_KEY_LESSONS);
    if (cached) return cached;
    throw createOfflineError();
  }
  
  try {
    const res = await fetch('/api/lessons');
    if (!res.ok) throw new Error('Failed to load lessons');
    const data = await res.json();
    saveToLocalStorage(STORAGE_KEY_LESSONS, data);
    return data;
  } catch (error) {
    // Try fallback to localStorage
    const cached = getFromLocalStorage(STORAGE_KEY_LESSONS);
    if (cached) return cached;
    throw error;
  }
}

// AI Configuration
export async function getAIConfig() {
  if (!isOnline()) {
    return { aiEnabled: false };
  }
  
  try {
    const res = await fetch('/api/ai/config');
    if (!res.ok) return { aiEnabled: false };
    return res.json();
  } catch (error) {
    return { aiEnabled: false };
  }
}

// Generate AI lesson
export async function generateAILesson(category, topic) {
  if (!isOnline()) {
    throw new Error('AI generation requires an internet connection');
  }
  
  const res = await fetch('/api/ai/generate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ category, topic })
  });
  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: 'Failed to generate lesson' }));
    throw new Error(error.error || 'Failed to generate lesson');
  }
  return res.json();
}

export async function randomLesson(category = 'any') {
  if (!isOnline()) {
    const cached = getFromLocalStorage(STORAGE_KEY_LESSONS);
    if (cached && cached.lessons && cached.lessons.length > 0) {
      const filtered = category === 'any' ? cached.lessons : cached.lessons.filter(l => l.category === category);
      if (filtered.length > 0) {
        return filtered[Math.floor(Math.random() * filtered.length)];
      }
    }
    throw createOfflineError();
  }
  
  const res = await fetch(`/api/random?category=${encodeURIComponent(category)}`);
  if (!res.ok) throw new Error('No lesson available');
  return res.json();
}

// Reading: fetch a random lesson to read
export async function getReadingLesson(category = 'any') {
  if (!isOnline()) {
    throw createOfflineError();
  }
  
  const res = await fetch(`/api/session?stage=lesson&category=${encodeURIComponent(category)}`);
  if (!res.ok) throw new Error('Failed to get lesson');
  return res.json();
}

// Add current lesson to study list (server-side)
export async function addLessonToQuiz(title) {
  const res = await fetch('/api/session?stage=add', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title })
  });
  if (!res.ok) throw new Error('Failed to add lesson');
  return res.json();
}

// Build quiz from lessons seen (or all if none)
export async function startQuiz() {
  const res = await fetch('/api/session?stage=startQuiz', { method: 'POST' });
  if (!res.ok) throw new Error('Failed to start quiz');
  return res.json();
}

// Fetch current question (index + total)
export async function getCurrentQuiz() {
  const res = await fetch('/api/session?stage=quiz');
  if (!res.ok) throw new Error('No active quiz');
  return res.json();
}

// Submit answer; server returns either the next question or end-of-quiz result
export async function answerQuiz(answerIndex) {
  const res = await fetch('/api/session?stage=answer', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ answerIndex })
  });
  if (!res.ok) throw new Error('Answer failed');
  return res.json();
}

export async function getProChallenge({ topic, difficulty } = {}) {
  const params = new URLSearchParams();
  if (topic) params.set('topic', topic);
  if (difficulty) params.set('difficulty', difficulty);
  const qs = params.toString();
  const res = await fetch(`/api/prochallenge${qs ? `?${qs}` : ''}`);
  if (!res.ok) throw new Error('Unable to fetch challenge');
  return res.json();
}

export async function submitProChallenge({ id, code }) {
  const res = await fetch('/api/prochallenge/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id, code })
  });
  if (!res.ok) throw new Error('Submission failed');
  return res.json();
}

export async function requestProHint(id) {
  const res = await fetch('/api/prochallenge/hint', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id })
  });
  if (!res.ok) throw new Error('Hint unavailable');
  return res.json();
}

// ---------- Leaderboard ----------

export async function getLeaderboard(mode = '') {
  const url = mode ? `/api/leaderboard?mode=${encodeURIComponent(mode)}` : '/api/leaderboard';
  const res = await fetch(url);
  if (!res.ok) throw new Error('Failed to get leaderboard');
  return res.json();
}

export async function submitToLeaderboard(name, score, mode, category = '') {
  const res = await fetch('/api/leaderboard/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, score, mode, category })
  });
  if (!res.ok) throw new Error('Failed to submit score');
  return res.json();
}

// Update typing score server-side
export async function updateTypingScore(score) {
  const res = await fetch('/api/typing/score', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ score })
  });
  if (!res.ok) throw new Error('Failed to update typing score');
  return res.json();
}

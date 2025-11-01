export async function getLessons() {
  const res = await fetch('/api/lessons');
  if (!res.ok) throw new Error('Failed to load lessons');
  return res.json();
}

export async function randomLesson(category = 'any') {
  const res = await fetch(`/api/random?category=${encodeURIComponent(category)}`);
  if (!res.ok) throw new Error('No lesson available');
  return res.json();
}

// Reading: fetch a random lesson to read
export async function getReadingLesson(category = 'any') {
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

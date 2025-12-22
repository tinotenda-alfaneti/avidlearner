import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  getLessons,
  getAIConfig,
  generateAILesson,
  randomLesson,
  getReadingLesson,
  addLessonToQuiz,
  startQuiz,
  getCurrentQuiz,
  answerQuiz,
  getProChallenge,
  submitProChallenge,
  requestProHint
} from './api.js'

describe('API Module', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    global.fetch = vi.fn()
  })

  describe('getLessons', () => {
    it('fetches lessons successfully', async () => {
      const mockResponse = {
        categories: ['Go', 'Python'],
        lessons: {
          Go: [{ title: 'Go Basics', category: 'Go', text: 'Learn Go' }]
        }
      }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const result = await getLessons()

      expect(global.fetch).toHaveBeenCalledWith('/api/lessons')
      expect(result).toEqual(mockResponse)
    })

    it('throws error when fetch fails', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 500
      })

      // Mock localStorage to return null so fallback doesn't work
      const getItemSpy = vi.spyOn(Storage.prototype, 'getItem').mockReturnValue(null)

      await expect(getLessons()).rejects.toThrow('Failed to load lessons')
      
      getItemSpy.mockRestore()
    })
  })

  describe('getAIConfig', () => {
    it('returns AI configuration when successful', async () => {
      const mockConfig = { aiEnabled: true, provider: 'openai' }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockConfig
      })

      const result = await getAIConfig()

      expect(global.fetch).toHaveBeenCalledWith('/api/ai/config')
      expect(result).toEqual(mockConfig)
    })

    it('returns disabled config when fetch fails', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 404
      })

      const result = await getAIConfig()

      expect(result).toEqual({ aiEnabled: false })
    })
  })

  describe('generateAILesson', () => {
    it('generates AI lesson successfully', async () => {
      const mockLesson = {
        title: 'Test Lesson',
        category: 'Go',
        text: 'Test content',
        explain: 'Explanation',
        useCases: ['case1'],
        tips: ['tip1']
      }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ lesson: mockLesson })
      })

      const result = await generateAILesson('Go', 'concurrency')

      expect(global.fetch).toHaveBeenCalledWith('/api/ai/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ category: 'Go', topic: 'concurrency' })
      })
      expect(result).toEqual({ lesson: mockLesson })
    })

    it('throws error with error message from response', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => ({ error: 'Invalid topic' })
      })

      await expect(generateAILesson('Go', '')).rejects.toThrow('Invalid topic')
    })

    it('throws default error when response parsing fails', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => { throw new Error('Parse error') }
      })

      await expect(generateAILesson('Go', 'test')).rejects.toThrow('Failed to generate lesson')
    })
  })

  describe('randomLesson', () => {
    it('fetches random lesson with category', async () => {
      const mockLesson = { title: 'Random Lesson', category: 'Go', text: 'Content' }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockLesson
      })

      const result = await randomLesson('Go')

      expect(global.fetch).toHaveBeenCalledWith('/api/random?category=Go')
      expect(result).toEqual(mockLesson)
    })

    it('defaults to "any" category', async () => {
      const mockLesson = { title: 'Random Lesson', category: 'Python', text: 'Content' }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockLesson
      })

      const result = await randomLesson()

      expect(global.fetch).toHaveBeenCalledWith('/api/random?category=any')
      expect(result).toEqual(mockLesson)
    })

    it('throws error when no lesson available', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 404
      })

      await expect(randomLesson()).rejects.toThrow('No lesson available')
    })
  })

  describe('getReadingLesson', () => {
    it('fetches reading lesson successfully', async () => {
      const mockResponse = {
        stage: 'lesson',
        lesson: { title: 'Test', category: 'Go', text: 'Content' }
      }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const result = await getReadingLesson('Go')

      expect(global.fetch).toHaveBeenCalledWith('/api/session?stage=lesson&category=Go&source=all')
      expect(result).toEqual(mockResponse)
    })
  })

  describe('addLessonToQuiz', () => {
    it('adds lesson to quiz successfully', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true })
      })

      await addLessonToQuiz('Test Lesson')

      expect(global.fetch).toHaveBeenCalledWith('/api/session?stage=add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: 'Test Lesson' })
      })
    })

    it('throws error when add fails', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 500
      })

      await expect(addLessonToQuiz('Test')).rejects.toThrow('Failed to add lesson')
    })
  })

  describe('startQuiz', () => {
    it('starts quiz successfully', async () => {
      const mockResponse = {
        stage: 'quiz',
        question: 'What is Go?',
        options: ['A', 'B', 'C', 'D']
      }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const result = await startQuiz()

      expect(global.fetch).toHaveBeenCalledWith('/api/session?stage=startQuiz', {
        method: 'POST'
      })
      expect(result).toEqual(mockResponse)
    })
  })

  describe('getCurrentQuiz', () => {
    it('fetches current quiz question', async () => {
      const mockQuestion = {
        question: 'Test?',
        options: ['A', 'B'],
        index: 1,
        total: 5
      }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockQuestion
      })

      const result = await getCurrentQuiz()

      expect(global.fetch).toHaveBeenCalledWith('/api/session?stage=quiz')
      expect(result).toEqual(mockQuestion)
    })
  })

  describe('answerQuiz', () => {
    it('submits quiz answer successfully', async () => {
      const mockResponse = {
        correct: true,
        coinsEarned: 10,
        message: 'Correct!'
      }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      })

      const result = await answerQuiz(2)

      expect(global.fetch).toHaveBeenCalledWith('/api/session?stage=answer', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ answerIndex: 2 })
      })
      expect(result).toEqual(mockResponse)
    })
  })

  describe('getProChallenge', () => {
    it('fetches challenge without filters', async () => {
      const mockChallenge = { id: 'test', title: 'Test Challenge' }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockChallenge
      })

      const result = await getProChallenge()

      expect(global.fetch).toHaveBeenCalledWith('/api/prochallenge')
      expect(result).toEqual(mockChallenge)
    })

    it('fetches challenge with topic and difficulty filters', async () => {
      const mockChallenge = { id: 'test', title: 'Test Challenge' }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockChallenge
      })

      const result = await getProChallenge({ topic: 'algorithms', difficulty: 'hard' })

      expect(global.fetch).toHaveBeenCalledWith('/api/prochallenge?topic=algorithms&difficulty=hard')
      expect(result).toEqual(mockChallenge)
    })

    it('throws error when unable to fetch', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 404
      })

      await expect(getProChallenge()).rejects.toThrow('Unable to fetch challenge')
    })
  })

  describe('submitProChallenge', () => {
    it('submits challenge code successfully', async () => {
      const mockResult = {
        passed: true,
        xpEarned: 100,
        coinsEarned: 50
      }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResult
      })

      const result = await submitProChallenge({
        id: 'challenge1',
        code: 'package main\nfunc main() {}'
      })

      expect(global.fetch).toHaveBeenCalledWith('/api/prochallenge/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: 'challenge1', code: 'package main\nfunc main() {}' })
      })
      expect(result).toEqual(mockResult)
    })

    it('throws error when submission fails', async () => {
      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 400
      })

      await expect(submitProChallenge({ id: 'test', code: 'code' }))
        .rejects.toThrow('Submission failed')
    })
  })

  describe('requestProHint', () => {
    it('requests hint successfully', async () => {
      const mockHint = { hint: 'Use goroutines' }

      global.fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockHint
      })

      const result = await requestProHint('challenge1')

      expect(global.fetch).toHaveBeenCalledWith('/api/prochallenge/hint', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: expect.stringContaining('challenge1')
      })
      expect(result).toEqual(mockHint)
    })
  })
})

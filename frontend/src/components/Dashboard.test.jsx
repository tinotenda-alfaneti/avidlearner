import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import Dashboard from '../components/Dashboard'

// Mock fetch for TechNews component
global.fetch = vi.fn();

describe('Dashboard Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock fetch to prevent TechNews from making real API calls
    fetch.mockResolvedValue({
      json: async () => []
    });
  });

  const defaultProps = {
    onStartLearn: vi.fn(),
    onStartTyping: vi.fn(),
    onStartProMode: vi.fn(),
    onStartAI: null,
    coins: 100,
    xp: 500,
    quizStreak: 5,
    typingStreak: 3,
    typingBest: 85,
    categoryOptions: ['Go', 'Python', 'JavaScript'],
    selectedCategory: 'any',
    onSelectCategory: vi.fn()
  }

  it('renders dashboard with stats', () => {
    render(<Dashboard {...defaultProps} />)

    // Check hero stats section - these appear multiple times so use getAllByText
    expect(screen.getAllByText(/Coins/).length).toBeGreaterThan(0)
    expect(screen.getAllByText(/Total XP/).length).toBeGreaterThan(0)
    expect(screen.getAllByText(/Quiz Streak/).length).toBeGreaterThan(0)
    // Verify numbers appear
    const allValues = screen.getAllByText(/100|500|5/)
    expect(allValues.length).toBeGreaterThan(0)
  })

  it('renders learn mode section', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByRole('heading', { name: /Learn Mode/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Start Learning/i })).toBeInTheDocument()
  })

  it('calls onStartLearn when Start Learning clicked', () => {
    render(<Dashboard {...defaultProps} />)

    const learnButton = screen.getByText(/Start Learning/i)
    fireEvent.click(learnButton)

    expect(defaultProps.onStartLearn).toHaveBeenCalledTimes(1)
  })

  it('calls onStartProMode when Start Coding clicked', () => {
    render(<Dashboard {...defaultProps} />)

    const codingButton = screen.getByText(/Start Coding/i)
    fireEvent.click(codingButton)

    expect(defaultProps.onStartProMode).toHaveBeenCalledTimes(1)
  })

  it('calls onStartTyping when typing button clicked', () => {
    render(<Dashboard {...defaultProps} />)

    const typingButton = screen.getByText(/Start Typing/i)
    fireEvent.click(typingButton)
    expect(defaultProps.onStartTyping).toHaveBeenCalledTimes(1)
  })

  it('renders coding mode section', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByRole('heading', { name: /Coding Mode/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Start Coding/i })).toBeInTheDocument()
  })

  it('shows AI Generate button when AI is enabled', () => {
    const propsWithAI = { ...defaultProps, onStartAI: vi.fn() }
    render(<Dashboard {...propsWithAI} />)

    const aiButton = screen.getByText(/AI Custom Lesson/i)
    expect(aiButton).toBeInTheDocument()
  })

  it('hides AI Generate button when AI is disabled', () => {
    render(<Dashboard {...defaultProps} />)

    const aiButton = screen.queryByText(/AI Custom Lesson/i)
    expect(aiButton).not.toBeInTheDocument()
  })

  it('displays quiz streak correctly', () => {
    render(<Dashboard {...defaultProps} />)

    // Quiz streak appears in hero stats and learn mode card
    expect(screen.getByText(/Quiz Streak/)).toBeInTheDocument()
    // Verify the value appears somewhere (don't need to match exact count since it appears multiple times)
    const allValues = screen.getAllByText('5')
    expect(allValues.length).toBeGreaterThan(0)
  })

  it('displays typing stats correctly', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByText(/3/)).toBeInTheDocument() // streak
    expect(screen.getByText(/85/)).toBeInTheDocument() // best WPM
  })

  it('renders typing mode section', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByRole('heading', { name: /Typing Mode/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Start Typing/i })).toBeInTheDocument()
  })

  it('calls onStartAI when AI button clicked', () => {
    const onStartAI = vi.fn()
    const propsWithAI = { ...defaultProps, onStartAI }
    render(<Dashboard {...propsWithAI} />)

    const aiButton = screen.getByText(/AI Custom Lesson/i)
    fireEvent.click(aiButton)
    expect(onStartAI).toHaveBeenCalledTimes(1)
  })

  it('displays all mode cards', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByRole('heading', { name: /Learn Mode/i })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /Coding Mode/i })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /Typing Mode/i })).toBeInTheDocument()
  })

  it('renders leaderboard section', () => {
    const onOpenLeaderboard = vi.fn()
    const propsWithLeaderboard = { ...defaultProps, onOpenLeaderboard }
    render(<Dashboard {...propsWithLeaderboard} />)

    expect(screen.getByRole('heading', { name: /Global Leaderboard/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /View Leaderboard/i })).toBeInTheDocument()
  })

  it('calls onOpenLeaderboard when View Leaderboard clicked', () => {
    const onOpenLeaderboard = vi.fn()
    const propsWithLeaderboard = { ...defaultProps, onOpenLeaderboard }
    render(<Dashboard {...propsWithLeaderboard} />)

    const leaderboardButton = screen.getByText(/View Leaderboard/i)
    fireEvent.click(leaderboardButton)
    expect(onOpenLeaderboard).toHaveBeenCalledTimes(1)
  })
})

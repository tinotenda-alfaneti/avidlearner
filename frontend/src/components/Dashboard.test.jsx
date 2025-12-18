import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import Dashboard from '../components/Dashboard'

describe('Dashboard Component', () => {
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

    const coinsBadges = screen.getAllByText(/Coins: 100/)
    expect(coinsBadges.length).toBeGreaterThan(0)
    expect(screen.getByText(/XP: 500/)).toBeInTheDocument()
  })

  it('renders learn mode section', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByRole('heading', { name: /Learn Mode/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Launch Learn Mode/i })).toBeInTheDocument()
  })

  it('calls onStartLearn when Launch Learn Mode clicked', () => {
    render(<Dashboard {...defaultProps} />)

    const learnButton = screen.getByText(/Launch Learn Mode/i)
    fireEvent.click(learnButton)

    expect(defaultProps.onStartLearn).toHaveBeenCalledTimes(1)
  })

  it('calls onStartProMode when Launch Coding Mode clicked', () => {
    render(<Dashboard {...defaultProps} />)

    const codingButton = screen.getByText(/Launch Coding Mode/i)
    fireEvent.click(codingButton)

    expect(defaultProps.onStartProMode).toHaveBeenCalledTimes(1)
  })

  it('calls onStartTyping when typing button clicked', () => {
    render(<Dashboard {...defaultProps} />)

    const typingButton = screen.getByText(/Launch Typing Mode/i)
    fireEvent.click(typingButton)
    expect(defaultProps.onStartTyping).toHaveBeenCalledTimes(1)
  })

  it('renders coding mode section', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByRole('heading', { name: /Coding Mode/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Launch Coding Mode/i })).toBeInTheDocument()
  })

  it('shows AI Generate button when AI is enabled', () => {
    const propsWithAI = { ...defaultProps, onStartAI: vi.fn() }
    render(<Dashboard {...propsWithAI} />)

    const aiButton = screen.getByText(/Generate AI Lesson/i)
    expect(aiButton).toBeInTheDocument()
  })

  it('hides AI Generate button when AI is disabled', () => {
    render(<Dashboard {...defaultProps} />)

    const aiButton = screen.queryByText(/Generate AI Lesson/i)
    expect(aiButton).not.toBeInTheDocument()
  })

  it('displays quiz streak correctly', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByText(/Quiz Streak: 5/)).toBeInTheDocument()
  })

  it('displays typing stats correctly', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByText(/3/)).toBeInTheDocument() // streak
    expect(screen.getByText(/85/)).toBeInTheDocument() // best WPM
  })

  it('renders typing mode section', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByRole('heading', { name: /Typing Mode/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Launch Typing Mode/i })).toBeInTheDocument()
  })

  it('calls onStartAI when AI button clicked', () => {
    const onStartAI = vi.fn()
    const propsWithAI = { ...defaultProps, onStartAI }
    render(<Dashboard {...propsWithAI} />)

    const aiButton = screen.getByText(/Generate AI Lesson/i)
    fireEvent.click(aiButton)
    expect(onStartAI).toHaveBeenCalledTimes(1)
  })

  it('displays all mode cards', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByRole('heading', { name: /Learn Mode/i })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /Coding Mode/i })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /Typing Mode/i })).toBeInTheDocument()
  })
})

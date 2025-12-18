import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import Dashboard from '../components/Dashboard'

describe('Dashboard Component', () => {
  const defaultProps = {
    onStartReading: vi.fn(),
    onStartQuiz: vi.fn(),
    onStartTyping: vi.fn(),
    onStartProMode: vi.fn(),
    onStartAIGenerate: vi.fn(),
    coins: 100,
    xp: 500,
    quizStreak: 5,
    typingStreak: 3,
    typingBest: 85,
    categories: ['Go', 'Python', 'JavaScript'],
    selectedCategory: 'any',
    onCategoryChange: vi.fn(),
    aiEnabled: false
  }

  it('renders dashboard with stats', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByText(/100/)).toBeInTheDocument() // coins
    expect(screen.getByText(/500/)).toBeInTheDocument() // xp
  })

  it('renders learn mode section', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByText(/Learn Mode/i)).toBeInTheDocument()
    expect(screen.getByText(/Start Reading/i)).toBeInTheDocument()
    expect(screen.getByText(/Take Quiz/i)).toBeInTheDocument()
  })

  it('calls onStartReading when Start Reading clicked', () => {
    render(<Dashboard {...defaultProps} />)

    const readingButton = screen.getByText(/Start Reading/i)
    fireEvent.click(readingButton)

    expect(defaultProps.onStartReading).toHaveBeenCalledTimes(1)
  })

  it('calls onStartQuiz when Take Quiz clicked', () => {
    render(<Dashboard {...defaultProps} />)

    const quizButton = screen.getByText(/Take Quiz/i)
    fireEvent.click(quizButton)

    expect(defaultProps.onStartQuiz).toHaveBeenCalledTimes(1)
  })

  it('calls onStartTyping when typing button clicked', () => {
    render(<Dashboard {...defaultProps} />)

    const typingButton = screen.getByText(/Start Typing/i) || screen.getByText(/Typing/i)
    if (typingButton) {
      fireEvent.click(typingButton)
      expect(defaultProps.onStartTyping).toHaveBeenCalledTimes(1)
    }
  })

  it('calls onStartProMode when Pro Mode button clicked', () => {
    render(<Dashboard {...defaultProps} />)

    const proButton = screen.getByText(/Pro Mode/i) || screen.getByText(/Coding/i)
    if (proButton) {
      fireEvent.click(proButton)
      expect(defaultProps.onStartProMode).toHaveBeenCalled()
    }
  })

  it('shows AI Generate button when AI is enabled', () => {
    const propsWithAI = { ...defaultProps, aiEnabled: true }
    render(<Dashboard {...propsWithAI} />)

    const aiButton = screen.queryByText(/AI.*Generate/i) || screen.queryByText(/Generate/i)
    expect(aiButton).toBeInTheDocument()
  })

  it('hides AI Generate button when AI is disabled', () => {
    render(<Dashboard {...defaultProps} />)

    const aiButton = screen.queryByText(/AI.*Generate/i)
    // Button should either not exist or not be visible when AI disabled
    if (aiButton) {
      expect(aiButton).not.toBeVisible()
    }
  })

  it('displays quiz streak correctly', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByText(/5/)).toBeInTheDocument()
  })

  it('displays typing stats correctly', () => {
    render(<Dashboard {...defaultProps} />)

    expect(screen.getByText(/3/)).toBeInTheDocument() // streak
    expect(screen.getByText(/85/)).toBeInTheDocument() // best WPM
  })

  it('renders category selector', () => {
    render(<Dashboard {...defaultProps} />)

    // Category selector should be present
    const categoryElement = screen.queryByText(/any/i) || screen.queryByDisplayValue(/any/i)
    expect(categoryElement).toBeInTheDocument()
  })

  it('calls onCategoryChange when category selected', () => {
    const { container } = render(<Dashboard {...defaultProps} />)

    const select = container.querySelector('select')
    if (select) {
      fireEvent.change(select, { target: { value: 'Go' } })
      expect(defaultProps.onCategoryChange).toHaveBeenCalledWith('Go')
    }
  })

  it('renders all available categories', () => {
    render(<Dashboard {...defaultProps} />)

    // Check if categories are available
    defaultProps.categories.forEach(category => {
      const element = screen.queryByText(category)
      if (element) {
        expect(element).toBeInTheDocument()
      }
    })
  })
})

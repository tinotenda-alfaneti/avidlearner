import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import AILessonGenerator from '../components/AILessonGenerator'
import * as api from '../api'

vi.mock('../api')

describe('AILessonGenerator Component', () => {
  const defaultProps = {
    categories: ['Go', 'Python', 'JavaScript'],
    onLessonGenerated: vi.fn(),
    onBack: vi.fn()
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders AI lesson generator form', () => {
    render(<AILessonGenerator {...defaultProps} />)
    
    expect(screen.getByText(/Generate AI Lesson/i)).toBeInTheDocument()
    expect(screen.getByText(/Category/i)).toBeInTheDocument()
    expect(screen.getByText(/Topic/i)).toBeInTheDocument()
  })

  it('renders all categories in select', () => {
    render(<AILessonGenerator {...defaultProps} />)
    
    const categorySelect = screen.getByRole('combobox', { name: /category/i }) || 
                          screen.getAllByRole('combobox')[0]
    
    if (categorySelect) {
      defaultProps.categories.forEach(category => {
        expect(screen.queryByText(category)).toBeTruthy()
      })
    }
  })

  it('renders topic input field', () => {
    render(<AILessonGenerator {...defaultProps} />)
    
    const topicInput = screen.getByRole('textbox', { name: /topic/i }) ||
                       screen.getByPlaceholderText(/topic/i)
    
    expect(topicInput).toBeInTheDocument()
  })

  it('updates category when changed', () => {
    const { container } = render(<AILessonGenerator {...defaultProps} />)
    
    const select = container.querySelector('select')
    if (select) {
      fireEvent.change(select, { target: { value: 'Python' } })
      expect(select.value).toBe('Python')
    }
  })

  it('updates topic when typed', () => {
    render(<AILessonGenerator {...defaultProps} />)
    
    const topicInput = screen.getByRole('textbox') || screen.getByPlaceholderText(/topic/i)
    
    fireEvent.change(topicInput, { target: { value: 'concurrency' } })
    expect(topicInput.value).toBe('concurrency')
  })

  it('calls generateAILesson when form submitted', async () => {
    const mockLesson = {
      title: 'Test Lesson',
      category: 'Go',
      text: 'Test content'
    }
    
    api.generateAILesson.mockResolvedValueOnce({ lesson: mockLesson })
    
    render(<AILessonGenerator {...defaultProps} />)
    
    const topicInput = screen.getByRole('textbox') || screen.getByPlaceholderText(/topic/i)
    fireEvent.change(topicInput, { target: { value: 'goroutines' } })
    
    const generateButton = screen.getByText(/Generate/i)
    fireEvent.click(generateButton)
    
    await waitFor(() => {
      expect(api.generateAILesson).toHaveBeenCalled()
    })
  })

  it('calls onLessonGenerated on successful generation', async () => {
    const mockLesson = {
      title: 'Test Lesson',
      category: 'Go',
      text: 'Test content'
    }
    
    api.generateAILesson.mockResolvedValueOnce({ lesson: mockLesson })
    
    render(<AILessonGenerator {...defaultProps} />)
    
    const topicInput = screen.getByRole('textbox') || screen.getByPlaceholderText(/topic/i)
    fireEvent.change(topicInput, { target: { value: 'goroutines' } })
    
    const generateButton = screen.getByText(/Generate/i)
    fireEvent.click(generateButton)
    
    await waitFor(() => {
      expect(defaultProps.onLessonGenerated).toHaveBeenCalledWith({ lesson: mockLesson })
    })
  })

  it('shows loading state during generation', async () => {
    api.generateAILesson.mockImplementationOnce(() => 
      new Promise(resolve => setTimeout(() => resolve({ lesson: {} }), 100))
    )
    
    render(<AILessonGenerator {...defaultProps} />)
    
    const topicInput = screen.getByRole('textbox') || screen.getByPlaceholderText(/topic/i)
    fireEvent.change(topicInput, { target: { value: 'test' } })
    
    const generateButton = screen.getByText(/Generate/i)
    fireEvent.click(generateButton)
    
    // Should show loading state
    expect(screen.queryByText(/Generating/i) || screen.queryByText(/Loading/i)).toBeTruthy()
  })

  it('shows error message on generation failure', async () => {
    api.generateAILesson.mockRejectedValueOnce(new Error('Generation failed'))
    
    render(<AILessonGenerator {...defaultProps} />)
    
    const topicInput = screen.getByRole('textbox') || screen.getByPlaceholderText(/topic/i)
    fireEvent.change(topicInput, { target: { value: 'test' } })
    
    const generateButton = screen.getByText(/Generate/i)
    fireEvent.click(generateButton)
    
    await waitFor(() => {
      expect(screen.queryByText(/failed/i) || screen.queryByText(/error/i)).toBeTruthy()
    })
  })

  it('calls onBack when back button clicked', () => {
    render(<AILessonGenerator {...defaultProps} />)
    
    const backButton = screen.getByText(/Back/i) || screen.getByText(/Cancel/i)
    if (backButton) {
      fireEvent.click(backButton)
      expect(defaultProps.onBack).toHaveBeenCalled()
    }
  })

  it('disables generate button when topic is empty', () => {
    render(<AILessonGenerator {...defaultProps} />)
    
    const generateButton = screen.getByText(/Generate/i)
    
    // Button should be disabled initially
    expect(generateButton).toBeDisabled()
  })

  it('enables generate button when topic is provided', () => {
    render(<AILessonGenerator {...defaultProps} />)
    
    const topicInput = screen.getByRole('textbox') || screen.getByPlaceholderText(/topic/i)
    fireEvent.change(topicInput, { target: { value: 'goroutines' } })
    
    const generateButton = screen.getByText(/Generate/i)
    expect(generateButton).not.toBeDisabled()
  })
})

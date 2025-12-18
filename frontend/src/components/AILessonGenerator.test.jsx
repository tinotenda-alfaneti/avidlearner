import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import AILessonGenerator from '../components/AILessonGenerator'
import * as api from '../api'

vi.mock('../api')

describe('AILessonGenerator Component', () => {
  const defaultProps = {
    categories: ['Go', 'Python', 'JavaScript'],
    onLessonGenerated: vi.fn(),
    onCancel: vi.fn()
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders AI lesson generator form', () => {
    render(<AILessonGenerator {...defaultProps} />)
    
    expect(screen.getByText(/Generate AI Lesson/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Category/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Topic/i)).toBeInTheDocument()
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
    
    const generateButton = screen.getByRole('button', { name: /Generate Lesson/i })
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
    
    const generateButton = screen.getByRole('button', { name: /Generate Lesson/i })
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
    
    const generateButton = screen.getByRole('button', { name: /Generate Lesson/i })
    fireEvent.click(generateButton)
    
    // Should show loading state
    expect(screen.queryByText(/Generating/i) || screen.queryByText(/Loading/i)).toBeTruthy()
  })

  it('shows error message on generation failure', async () => {
    api.generateAILesson.mockRejectedValueOnce(new Error('Generation failed'))
    
    render(<AILessonGenerator {...defaultProps} />)
    
    const topicInput = screen.getByRole('textbox') || screen.getByPlaceholderText(/topic/i)
    fireEvent.change(topicInput, { target: { value: 'test' } })
    
    const generateButton = screen.getByRole('button', { name: /Generate Lesson/i })
    fireEvent.click(generateButton)
    
    await waitFor(() => {
      expect(screen.queryByText(/failed/i) || screen.queryByText(/error/i)).toBeTruthy()
    })
  })

  it('calls onCancel when cancel button clicked', () => {
    render(<AILessonGenerator {...defaultProps} />)
    
    const cancelButton = screen.getByRole('button', { name: /Cancel/i })
    fireEvent.click(cancelButton)
    expect(defaultProps.onCancel).toHaveBeenCalled()
  })

  it('shows error when generating with empty topic', async () => {
    render(<AILessonGenerator {...defaultProps} />)
    
    const generateButton = screen.getByRole('button', { name: /Generate Lesson/i })
    fireEvent.click(generateButton)
    
    // Should show error message
    await waitFor(() => {
      expect(screen.getByText(/enter a topic/i)).toBeInTheDocument()
    })
  })

  it('enables generate button when topic is provided', () => {
    render(<AILessonGenerator {...defaultProps} />)
    
    const topicInput = screen.getByRole('textbox') || screen.getByPlaceholderText(/topic/i)
    fireEvent.change(topicInput, { target: { value: 'goroutines' } })
    
    const generateButton = screen.getByRole('button', { name: /Generate Lesson/i })
    expect(generateButton).not.toBeDisabled()
  })
})

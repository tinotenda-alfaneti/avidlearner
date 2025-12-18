import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import ModeCard from '../components/ModeCard'

describe('ModeCard Component', () => {
  const defaultProps = {
    title: 'Test Mode',
    description: 'This is a test mode description',
    badges: [
      { label: 'Coins', value: 100 },
      { label: 'Streak', value: 5 }
    ],
    actions: [
      { label: 'Start', onClick: vi.fn() }
    ],
    footer: 'This is a footer message'
  }

  it('renders title correctly', () => {
    render(<ModeCard {...defaultProps} />)
    expect(screen.getByText('Test Mode')).toBeInTheDocument()
  })

  it('renders description correctly', () => {
    render(<ModeCard {...defaultProps} />)
    expect(screen.getByText('This is a test mode description')).toBeInTheDocument()
  })

  it('renders all badges', () => {
    render(<ModeCard {...defaultProps} />)
    
    expect(screen.getByText('Coins: 100')).toBeInTheDocument()
    expect(screen.getByText('Streak: 5')).toBeInTheDocument()
  })

  it('renders action buttons', () => {
    render(<ModeCard {...defaultProps} />)
    
    const button = screen.getByText('Start')
    expect(button).toBeInTheDocument()
  })

  it('calls onClick when action button clicked', () => {
    const onClick = vi.fn()
    const props = {
      ...defaultProps,
      actions: [{ label: 'Click Me', onClick }]
    }
    
    render(<ModeCard {...props} />)
    
    const button = screen.getByText('Click Me')
    fireEvent.click(button)
    
    expect(onClick).toHaveBeenCalledTimes(1)
  })

  it('renders footer message', () => {
    render(<ModeCard {...defaultProps} />)
    expect(screen.getByText('This is a footer message')).toBeInTheDocument()
  })

  it('renders multiple action buttons', () => {
    const props = {
      ...defaultProps,
      actions: [
        { label: 'Action 1', onClick: vi.fn() },
        { label: 'Action 2', onClick: vi.fn() },
        { label: 'Action 3', onClick: vi.fn() }
      ]
    }
    
    render(<ModeCard {...props} />)
    
    expect(screen.getByText('Action 1')).toBeInTheDocument()
    expect(screen.getByText('Action 2')).toBeInTheDocument()
    expect(screen.getByText('Action 3')).toBeInTheDocument()
  })

  it('handles empty badges array', () => {
    const props = {
      ...defaultProps,
      badges: []
    }
    
    render(<ModeCard {...props} />)
    expect(screen.getByText('Test Mode')).toBeInTheDocument()
  })

  it('handles missing footer', () => {
    const props = {
      ...defaultProps,
      footer: undefined
    }
    
    const { container } = render(<ModeCard {...props} />)
    expect(container).toBeInTheDocument()
  })
})

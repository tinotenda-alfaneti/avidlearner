import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import TechNews from './TechNews';

global.fetch = vi.fn();

describe('TechNews', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    fetch.mockImplementation(() => new Promise(() => {})); // Never resolves
    render(<TechNews />);
    expect(screen.getByText(/loading tech news/i)).toBeInTheDocument();
  });

  it('fetches and displays HackerNews articles by default', async () => {
    const mockTopStories = [1];
    const mockStory = {
      id: 1,
      title: 'Test HN Story',
      url: 'https://example.com',
      score: 100,
      by: 'testuser',
      time: Date.now() / 1000,
      descendants: 50
    };

    fetch
      .mockResolvedValueOnce({
        json: async () => mockTopStories
      })
      .mockResolvedValue({
        json: async () => mockStory
      });

    render(<TechNews />);

    await waitFor(() => {
      expect(screen.getAllByText('Test HN Story').length).toBe(1);
    });
    
    expect(screen.getByText(/100/)).toBeInTheDocument();
    expect(screen.getByText(/testuser/)).toBeInTheDocument();
  });

  it('switches to Dev.to when tab is clicked', async () => {
    const user = userEvent.setup();
    
    // Mock HackerNews initial load
    fetch.mockResolvedValueOnce({
      json: async () => [1]
    }).mockResolvedValueOnce({
      json: async () => ({
        id: 1,
        title: 'HN Story',
        url: 'https://example.com',
        score: 100,
        by: 'user',
        time: Date.now() / 1000,
        descendants: 10
      })
    });

    render(<TechNews />);
    
    await waitFor(() => {
      expect(screen.getByText('HN Story')).toBeInTheDocument();
    });

    // Mock Dev.to response
    const mockDevToArticles = [{
      id: 1,
      title: 'Test Dev.to Article',
      url: 'https://dev.to/test',
      public_reactions_count: 50,
      user: { name: 'Dev User' },
      published_at: new Date().toISOString(),
      comments_count: 25,
      tag_list: ['javascript', 'react']
    }];

    fetch.mockResolvedValueOnce({
      json: async () => mockDevToArticles
    });

    await user.click(screen.getByText('Dev.to'));

    await waitFor(() => {
      expect(screen.getByText('Test Dev.to Article')).toBeInTheDocument();
    });
  });

  it('displays error message when fetch fails', async () => {
    fetch.mockRejectedValue(new Error('Network error'));

    render(<TechNews />);

    await waitFor(() => {
      expect(screen.getByText(/failed to load news/i)).toBeInTheDocument();
    });
    
    expect(screen.getByText('Retry')).toBeInTheDocument();
  });

  it('formats time ago correctly', async () => {
    const now = Date.now() / 1000;
    const mockStory = {
      id: 1,
      title: 'Recent Story',
      url: 'https://example.com',
      score: 100,
      by: 'user',
      time: now - 3600, // 1 hour ago
      descendants: 10
    };

    fetch
      .mockResolvedValueOnce({ json: async () => [1] })
      .mockResolvedValueOnce({ json: async () => mockStory });

    render(<TechNews />);

    await waitFor(() => {
      expect(screen.getByText(/1h ago/)).toBeInTheDocument();
    });
  });
});

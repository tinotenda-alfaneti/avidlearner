import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import TechNews from './TechNews';

global.fetch = vi.fn();

// Mock navigator.onLine
const mockOnline = (isOnline) => {
  Object.defineProperty(navigator, 'onLine', {
    writable: true,
    value: isOnline
  });
};

// Mock localStorage
const localStorageMock = (() => {
  let store = {};
  return {
    getItem: (key) => store[key] || null,
    setItem: (key, value) => { store[key] = value; },
    clear: () => { store = {}; }
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
});

describe('TechNews', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorageMock.clear();
    mockOnline(true);
  });

  afterEach(() => {
    mockOnline(true);
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
    }, { timeout: 3000 });

    // Clear previous mocks and set up new mock for Dev.to
    vi.clearAllMocks();
    
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
    }, { timeout: 10000 });
  }, 15000); // Increase test timeout to 15 seconds

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

  it('shows cached news when offline', async () => {
    // First, fetch news while online
    const mockStory = {
      id: 1,
      title: 'Cached Story',
      url: 'https://example.com',
      score: 100,
      by: 'user',
      time: Date.now() / 1000,
      descendants: 10
    };

    fetch
      .mockResolvedValueOnce({ json: async () => [1] })
      .mockResolvedValueOnce({ json: async () => mockStory });

    const { unmount } = render(<TechNews />);

    await waitFor(() => {
      expect(screen.getByText('Cached Story')).toBeInTheDocument();
    });

    unmount();

    // Now go offline and render again
    mockOnline(false);
    
    render(<TechNews />);

    await waitFor(() => {
      expect(screen.getByText('Cached Story')).toBeInTheDocument();
      expect(screen.getByText(/offline - showing cached news/i)).toBeInTheDocument();
    });
  });

  it('shows error when offline and no cache available', async () => {
    mockOnline(false);
    
    render(<TechNews />);

    await waitFor(() => {
      expect(screen.getByText(/you are offline and no cached news is available/i)).toBeInTheDocument();
    });
    
    // Retry button should not be shown when offline
    expect(screen.queryByText('Retry')).not.toBeInTheDocument();
  });

  it('hides retry button when offline', async () => {
    mockOnline(false);
    
    render(<TechNews />);

    await waitFor(() => {
      expect(screen.getByText(/you are offline/i)).toBeInTheDocument();
    });
    
    expect(screen.queryByText('Retry')).not.toBeInTheDocument();
  });

  it('shows retry button when online but fetch fails', async () => {
    fetch.mockRejectedValue(new Error('Network error'));

    render(<TechNews />);

    await waitFor(() => {
      expect(screen.getByText('Retry')).toBeInTheDocument();
    });
  });

  it('uses cached news when fetch fails', async () => {
    // First, fetch news successfully
    const mockStory = {
      id: 1,
      title: 'Cached on Error Story',
      url: 'https://example.com',
      score: 100,
      by: 'user',
      time: Date.now() / 1000,
      descendants: 10
    };

    fetch
      .mockResolvedValueOnce({ json: async () => [1] })
      .mockResolvedValueOnce({ json: async () => mockStory });

    const { unmount } = render(<TechNews />);

    await waitFor(() => {
      expect(screen.getByText('Cached on Error Story')).toBeInTheDocument();
    });

    unmount();

    // Now make fetch fail
    fetch.mockRejectedValue(new Error('Network error'));
    
    render(<TechNews />);

    await waitFor(() => {
      expect(screen.getByText('Cached on Error Story')).toBeInTheDocument();
      expect(screen.getByText(/using cached news/i)).toBeInTheDocument();
    });
  });
});

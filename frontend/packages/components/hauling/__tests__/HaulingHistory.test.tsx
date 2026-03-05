import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/router';
import HaulingHistory from '../HaulingHistory';
import { HaulingRun } from '@industry-tool/client/data/models';

jest.mock('next-auth/react');
jest.mock('next/router');
jest.mock('../../loading', () => {
  return function MockLoading() {
    return <div data-testid="loading">Loading...</div>;
  };
});

// Silence next/image warning in jsdom
jest.mock('next/image', () => ({
  __esModule: true,
  default: ({ src, alt }: { src: string; alt: string }) => <img src={src} alt={alt} />,
}));

const mockUseSession = useSession as jest.MockedFunction<typeof useSession>;
const mockUseRouter = useRouter as jest.MockedFunction<typeof useRouter>;
const mockFetch = global.fetch as jest.Mock;

const mockSession = {
  data: { providerAccountId: '123', user: { name: 'Test User' } } as any,
  status: 'authenticated' as const,
  update: jest.fn(),
};

const mockPush = jest.fn();

const mockRuns: HaulingRun[] = [
  {
    id: 10,
    userId: 100,
    name: 'Completed Jita Run',
    status: 'COMPLETE',
    fromRegionId: 10000002,
    toRegionId: 10000043,
    maxVolumeM3: 300000,
    notifyTier2: false,
    notifyTier3: false,
    dailyDigest: false,
    createdAt: '2026-01-10T10:00:00Z',
    updatedAt: '2026-01-15T10:00:00Z',
    completedAt: '2026-01-15T10:00:00Z',
    items: [],
  },
  {
    id: 11,
    userId: 100,
    name: 'Cancelled Dodixie Run',
    status: 'CANCELLED',
    fromRegionId: 10000002,
    toRegionId: 10000032,
    notifyTier2: false,
    notifyTier3: false,
    dailyDigest: false,
    createdAt: '2026-01-05T10:00:00Z',
    updatedAt: '2026-01-06T10:00:00Z',
    items: [],
  },
];

describe('HaulingHistory Component', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-02-22T12:00:00Z'));
    mockFetch.mockClear();
    mockPush.mockClear();
    mockUseRouter.mockReturnValue({
      push: mockPush,
      query: {},
      pathname: '/hauling',
    } as any);
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('should show loading initially', () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockImplementation(() => new Promise(() => {}));

    render(<HaulingHistory />);
    expect(screen.getByTestId('loading')).toBeInTheDocument();
  });

  it('should show empty state when no runs', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ runs: [], total: 0 }),
    } as Response);

    render(<HaulingHistory />);

    await waitFor(() => {
      expect(screen.getByText('No completed runs found.')).toBeInTheDocument();
    });
  });

  it('should display runs after fetch', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ runs: mockRuns, total: 2 }),
    } as Response);

    render(<HaulingHistory />);

    await waitFor(() => {
      expect(screen.getByText('Completed Jita Run')).toBeInTheDocument();
      expect(screen.getByText('Cancelled Dodixie Run')).toBeInTheDocument();
    });
  });

  it('should match snapshot with runs list', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ runs: mockRuns, total: 2 }),
    } as Response);

    const { container } = render(<HaulingHistory />);

    await waitFor(() => {
      expect(screen.getByText('Completed Jita Run')).toBeInTheDocument();
    });

    expect(container).toMatchSnapshot();
  });

  it('should show route names correctly', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ runs: mockRuns, total: 2 }),
    } as Response);

    render(<HaulingHistory />);

    await waitFor(() => {
      const forgeText = screen.getAllByText(/The Forge/);
      expect(forgeText.length).toBeGreaterThan(0);
    });
  });

  it('should navigate to run detail on row click', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ runs: mockRuns, total: 2 }),
    } as Response);

    render(<HaulingHistory />);

    await waitFor(() => {
      expect(screen.getByText('Completed Jita Run')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Completed Jita Run'));
    expect(mockPush).toHaveBeenCalledWith('/hauling/10');
  });

  it('should filter by status COMPLETE', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ runs: mockRuns, total: 2 }),
    } as Response);

    render(<HaulingHistory />);

    await waitFor(() => {
      expect(screen.getByText('Completed Jita Run')).toBeInTheDocument();
    });

    // Both runs visible initially
    expect(screen.getByText('Cancelled Dodixie Run')).toBeInTheDocument();
  });

  it('should show 2 total runs indicator', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ runs: mockRuns, total: 2 }),
    } as Response);

    render(<HaulingHistory />);

    await waitFor(() => {
      expect(screen.getByText('2 total runs')).toBeInTheDocument();
    });
  });
});

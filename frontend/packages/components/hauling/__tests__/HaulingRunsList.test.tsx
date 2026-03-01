import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/router';
import HaulingRunsList from '../HaulingRunsList';
import { HaulingRun } from '@industry-tool/client/data/models';

jest.mock('next-auth/react');
jest.mock('next/router');
jest.mock('next/link', () => {
  return function MockLink({ children, href }: { children: React.ReactNode; href: string }) {
    return <a href={href}>{children}</a>;
  };
});
jest.mock('../../Navbar', () => {
  return function MockNavbar() {
    return <div data-testid="navbar">Navbar</div>;
  };
});
jest.mock('../../loading', () => {
  return function MockLoading() {
    return <div data-testid="loading">Loading...</div>;
  };
});

const mockUseSession = useSession as jest.MockedFunction<typeof useSession>;
const mockUseRouter = useRouter as jest.MockedFunction<typeof useRouter>;
const mockFetch = global.fetch as jest.Mock;

const mockSession = {
  data: { providerAccountId: '123', user: { name: 'Test User' } } as any,
  status: 'authenticated' as const,
  update: jest.fn(),
};

const mockRuns: HaulingRun[] = [
  {
    id: 1,
    userId: 100,
    name: 'Jita to Amarr Run',
    status: 'PLANNING',
    fromRegionId: 10000002,
    toRegionId: 10000043,
    maxVolumeM3: 300000,
    notifyTier2: false,
    notifyTier3: false,
    dailyDigest: false,
    createdAt: '2026-02-22T12:00:00Z',
    updatedAt: '2026-02-22T12:00:00Z',
    items: [
      {
        id: 10,
        runId: 1,
        typeId: 34,
        typeName: 'Tritanium',
        quantityPlanned: 10000,
        quantityAcquired: 5000,
        buyPriceIsk: 5.5,
        sellPriceIsk: 7.0,
        volumeM3: 0.01,
        fillPercent: 50,
        createdAt: '2026-02-22T12:00:00Z',
        updatedAt: '2026-02-22T12:00:00Z',
      },
    ],
  },
  {
    id: 2,
    userId: 100,
    name: 'Dodixie Run',
    status: 'IN_TRANSIT',
    fromRegionId: 10000002,
    toRegionId: 10000032,
    notifyTier2: true,
    notifyTier3: false,
    dailyDigest: true,
    createdAt: '2026-02-20T10:00:00Z',
    updatedAt: '2026-02-21T10:00:00Z',
    items: [],
  },
];

describe('HaulingRunsList Component', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-02-22T12:00:00Z'));
    mockFetch.mockClear();
    mockUseRouter.mockReturnValue({
      push: jest.fn(),
      query: {},
      pathname: '/hauling',
    } as any);
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('should show loading state initially', () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockImplementation(() => new Promise(() => {}));

    render(<HaulingRunsList />);
    expect(screen.getByTestId('loading')).toBeInTheDocument();
  });

  it('should show empty state when no runs', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    } as Response);

    render(<HaulingRunsList />);

    await waitFor(() => {
      expect(screen.getByText(/No hauling runs yet/)).toBeInTheDocument();
    });
  });

  it('should match snapshot with runs list', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockRuns,
    } as Response);

    const { container } = render(<HaulingRunsList />);

    await waitFor(() => {
      expect(screen.getByText('Jita to Amarr Run')).toBeInTheDocument();
    });

    expect(container).toMatchSnapshot();
  });

  it('should display run names and statuses', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockRuns,
    } as Response);

    render(<HaulingRunsList />);

    await waitFor(() => {
      expect(screen.getByText('Jita to Amarr Run')).toBeInTheDocument();
      expect(screen.getByText('Dodixie Run')).toBeInTheDocument();
    });
  });

  it('should open new run dialog when New Run clicked', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    } as Response);

    render(<HaulingRunsList />);

    await waitFor(() => {
      expect(screen.getByText(/No hauling runs yet/)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('New Run'));
    expect(screen.getByText('New Hauling Run')).toBeInTheDocument();
  });

  it('should show Discord notification checkboxes in new run dialog', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    } as Response);

    render(<HaulingRunsList />);

    await waitFor(() => {
      expect(screen.getByText(/No hauling runs yet/)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('New Run'));

    expect(screen.getByText('Discord Notifications')).toBeInTheDocument();
    expect(screen.getByText('Notify when fill crosses 80% (requires Discord)')).toBeInTheDocument();
    expect(screen.getByText('Notify when items are slow to fill (requires Discord)')).toBeInTheDocument();
    expect(screen.getByText('Daily digest in Discord')).toBeInTheDocument();
    expect(screen.getByText('Requires Discord linked in Settings.')).toBeInTheDocument();
  });

  it('should match snapshot with new run dialog open', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    } as Response);

    render(<HaulingRunsList />);

    await waitFor(() => {
      expect(screen.getByText(/No hauling runs yet/)).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('New Run'));

    const { container } = render(<HaulingRunsList />);
    expect(container).toMatchSnapshot();
  });

  it('should return null when no session', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated',
      update: jest.fn(),
    });

    const { container } = render(<HaulingRunsList />);
    expect(container.firstChild).toBeNull();
  });
});

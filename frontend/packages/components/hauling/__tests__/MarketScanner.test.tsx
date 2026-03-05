import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import MarketScanner from '../MarketScanner';
import { HaulingArbitrageRow, HaulingRun, TradingStation, UserTradingStructure } from '@industry-tool/client/data/models';

jest.mock('next-auth/react');
jest.mock('next/image', () => ({
  __esModule: true,
  default: ({ src, alt }: { src: string; alt: string }) => <img src={src} alt={alt} />,
}));
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

// Mock the UserStructuresDialog to keep snapshot stable
jest.mock('../UserStructuresDialog', () => {
  return function MockUserStructuresDialog({ open }: { open: boolean }) {
    return open ? <div data-testid="user-structures-dialog">Structures Dialog</div> : null;
  };
});

// Mock shadcn dialog
jest.mock('@/components/ui/dialog', () => ({
  Dialog: ({ children, open }: { children: React.ReactNode; open: boolean }) =>
    open ? <div data-testid="dialog">{children}</div> : null,
  DialogContent: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="dialog-content">{children}</div>
  ),
  DialogHeader: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="dialog-header">{children}</div>
  ),
  DialogTitle: ({ children }: { children: React.ReactNode }) => (
    <h2 data-testid="dialog-title">{children}</h2>
  ),
  DialogFooter: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="dialog-footer">{children}</div>
  ),
}));

const mockUseSession = useSession as jest.MockedFunction<typeof useSession>;
const mockFetch = global.fetch as jest.Mock;

const mockSession = {
  data: { providerAccountId: '123', user: { name: 'Test User' } } as any,
  status: 'authenticated' as const,
  update: jest.fn(),
};

const mockResults: HaulingArbitrageRow[] = [
  {
    typeId: 34,
    typeName: 'Tritanium',
    volumeM3: 0.01,
    buyPrice: 5.5,
    sellPrice: 7.0,
    netProfitIsk: 150000,
    spread: 27.3,
    volumeAvailable: 10000000,
    daysToSell: 0.5,
    indicator: 'gap',
    updatedAt: '2026-02-22T10:00:00Z',
  },
  {
    typeId: 35,
    typeName: 'Pyerite',
    volumeM3: 0.01,
    buyPrice: 9.0,
    sellPrice: 10.5,
    netProfitIsk: 75000,
    spread: 16.7,
    volumeAvailable: 5000000,
    daysToSell: 1.2,
    indicator: 'markup',
    updatedAt: '2026-02-22T10:00:00Z',
  },
];

const mockRuns: HaulingRun[] = [
  {
    id: 1,
    userId: 100,
    name: 'Jita to Amarr',
    status: 'PLANNING',
    fromRegionId: 10000002,
    toRegionId: 10000043,
    notifyTier2: false,
    notifyTier3: false,
    dailyDigest: false,
    createdAt: '2026-02-22T10:00:00Z',
    updatedAt: '2026-02-22T10:00:00Z',
  },
];

const mockStations: TradingStation[] = [
  {
    id: 1,
    stationId: 60003760,
    name: 'Jita IV - Moon 4 - Caldari Navy Assembly Plant',
    systemId: 30000142,
    regionId: 10000002,
    isPreset: true,
  },
];

const mockStructures: UserTradingStructure[] = [
  {
    id: 1,
    userId: 100,
    structureId: 1035466617946,
    name: 'Perimeter TTT',
    systemId: 30000144,
    regionId: 10000002,
    characterId: 456,
    accessOk: true,
    lastScannedAt: '2026-02-22T10:00:00Z',
    createdAt: '2026-02-20T08:00:00Z',
  },
];

beforeEach(() => {
  jest.useFakeTimers();
  jest.setSystemTime(new Date('2026-02-22T12:00:00Z'));
  mockFetch.mockClear();
  jest.clearAllMocks();
  mockUseSession.mockReturnValue(mockSession);
});

afterEach(() => {
  jest.useRealTimers();
});

describe('MarketScanner', () => {
  function setupDefaultFetches() {
    mockFetch
      .mockResolvedValueOnce({ ok: true, json: async () => mockRuns })      // /api/hauling/runs
      .mockResolvedValueOnce({ ok: true, json: async () => mockStations })  // /api/hauling/stations
      .mockResolvedValueOnce({ ok: true, json: async () => mockStructures }); // /api/hauling/structures
  }

  it('renders loading skeleton while fetching results', async () => {
    setupDefaultFetches();

    const { container } = render(<MarketScanner />);

    // Initial render before data loads
    expect(container).toMatchSnapshot();
  });

  it('renders empty state when no results', async () => {
    setupDefaultFetches();

    const { container } = render(<MarketScanner />);

    await waitFor(() => {
      // Wait for stations/structures/runs to load
    });

    expect(container).toMatchSnapshot();
  });

  it('renders results table with arbitrage rows', async () => {
    setupDefaultFetches();

    // Mock results fetch
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockResults,
    });

    const { container } = render(<MarketScanner />);

    // Trigger load
    const loadButton = await waitFor(() => screen.getByText('Load'));
    fireEvent.click(loadButton);

    await waitFor(() => {
      expect(screen.getByText('Tritanium')).toBeTruthy();
    });

    expect(container).toMatchSnapshot();
  });

  it('renders with stations in location picker after load', async () => {
    setupDefaultFetches();

    const { container } = render(<MarketScanner />);

    await waitFor(() => {
      // After fetches complete, stations should be available
    });

    expect(container).toMatchSnapshot();
  });

  it('opens structures dialog when settings icon is clicked', async () => {
    setupDefaultFetches();

    render(<MarketScanner />);

    await waitFor(() => {
      const settingsBtn = screen.getByTitle('Manage trading structures');
      fireEvent.click(settingsBtn);
    });

    expect(screen.getByTestId('user-structures-dialog')).toBeTruthy();
  });

  it('renders nothing when no session', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated' as const,
      update: jest.fn(),
    });

    const { container } = render(<MarketScanner />);
    expect(container.firstChild).toBeNull();
  });
});

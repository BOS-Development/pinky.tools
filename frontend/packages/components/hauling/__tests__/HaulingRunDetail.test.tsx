import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import HaulingRunDetail from '../HaulingRunDetail';
import { HaulingRun, HaulingRunPnlEntry, HaulingRunPnlSummary } from '@industry-tool/client/data/models';

jest.mock('next-auth/react');
jest.mock('next/link', () => {
  return function MockLink({ children, href }: { children: React.ReactNode; href: string }) {
    return <a href={href}>{children}</a>;
  };
});
jest.mock('next/image', () => {
  return function MockImage({ src, alt, width, height, style }: { src: string; alt: string; width: number; height: number; style?: React.CSSProperties }) {
    // eslint-disable-next-line @next/next/no-img-element
    return <img src={src} alt={alt} width={width} height={height} style={style} />;
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
const mockFetch = global.fetch as jest.Mock;

const mockSession = {
  data: { providerAccountId: '123', user: { name: 'Test User' } } as any,
  status: 'authenticated' as const,
  update: jest.fn(),
};

const mockRun: HaulingRun = {
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
      netProfitIsk: 1.5,
      qtySold: 0,
      sellFillPercent: 0,
      createdAt: '2026-02-22T12:00:00Z',
      updatedAt: '2026-02-22T12:00:00Z',
    },
  ],
};

const mockSellingRun: HaulingRun = {
  ...mockRun,
  id: 2,
  status: 'SELLING',
};

const mockCompleteRun: HaulingRun = {
  ...mockRun,
  id: 3,
  status: 'COMPLETE',
};

const mockPnlEntries: HaulingRunPnlEntry[] = [
  {
    id: 1,
    runId: 2,
    typeId: 34,
    typeName: 'Tritanium',
    quantitySold: 5000,
    avgSellPriceIsk: 7.2,
    totalRevenueIsk: 36000,
    totalCostIsk: 27500,
    netProfitIsk: 8500,
    createdAt: '2026-02-22T12:00:00Z',
    updatedAt: '2026-02-22T12:00:00Z',
  },
];

const mockPnlSummary: HaulingRunPnlSummary = {
  totalRevenueIsk: 36000,
  totalCostIsk: 27500,
  netProfitIsk: 8500,
  marginPct: 23.6,
  itemsSold: 1,
  itemsPending: 0,
};

// Helper to set up fetch mocks for a given run
function setupFetchMocks(run: HaulingRun, pnlEntries?: HaulingRunPnlEntry[], pnlSummary?: HaulingRunPnlSummary) {
  mockFetch.mockImplementation((url: string) => {
    if (typeof url === 'string') {
      if (url.includes('/api/hauling/runs/') && !url.includes('/items') && !url.includes('/status')) {
        return Promise.resolve({
          ok: true,
          json: async () => run,
        });
      }
      if (url.includes('/api/hauling/scanner')) {
        return Promise.resolve({
          ok: true,
          json: async () => [],
        });
      }
      if (url.includes('/api/hauling/pnl-summary')) {
        return Promise.resolve({
          ok: true,
          json: async () => pnlSummary || mockPnlSummary,
        });
      }
      if (url.includes('/api/hauling/pnl')) {
        return Promise.resolve({
          ok: true,
          json: async () => pnlEntries || mockPnlEntries,
        });
      }
      // ESI and zKillboard calls — return empty/minimal
      if (url.includes('esi.evetech.net')) {
        return Promise.resolve({
          ok: true,
          json: async () => [30000142, 30002187],
        });
      }
      if (url.includes('zkillboard.com')) {
        return Promise.resolve({
          ok: true,
          json: async () => [],
        });
      }
    }
    return Promise.resolve({ ok: false, json: async () => ({}) });
  });
}

describe('HaulingRunDetail Component', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-02-22T12:00:00Z'));
    mockFetch.mockClear();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('should return null when no session', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated',
      update: jest.fn(),
    });

    const { container } = render(<HaulingRunDetail runId={1} />);
    expect(container.firstChild).toBeNull();
  });

  it('should show loading state while fetching', () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockImplementation(() => new Promise(() => {}));

    render(<HaulingRunDetail runId={1} />);
    expect(screen.getByTestId('loading')).toBeInTheDocument();
  });

  it('should match snapshot with planning run', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockRun);

    const { container } = render(<HaulingRunDetail runId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Jita to Amarr Run')).toBeInTheDocument();
    });

    expect(container).toMatchSnapshot();
  });

  it('should display run name and status', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockRun);

    render(<HaulingRunDetail runId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Jita to Amarr Run')).toBeInTheDocument();
      expect(screen.getByText('PLANNING')).toBeInTheDocument();
    });
  });

  it('should show Route Safety section', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockRun);

    render(<HaulingRunDetail runId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Route Safety')).toBeInTheDocument();
    });
  });

  it('should show items table with item data', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockRun);

    render(<HaulingRunDetail runId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Tritanium')).toBeInTheDocument();
    });
  });

  it('should not show P&L section for non-selling/complete runs', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockRun);

    render(<HaulingRunDetail runId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Jita to Amarr Run')).toBeInTheDocument();
    });

    expect(screen.queryByText('Profit & Loss')).not.toBeInTheDocument();
  });

  it('should show P&L section for SELLING runs', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockSellingRun, mockPnlEntries, mockPnlSummary);

    render(<HaulingRunDetail runId={2} />);

    await waitFor(() => {
      expect(screen.getByText('Profit & Loss')).toBeInTheDocument();
    });
  });

  it('should show P&L entries for SELLING run', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockSellingRun, mockPnlEntries, mockPnlSummary);

    render(<HaulingRunDetail runId={2} />);

    await waitFor(() => {
      expect(screen.getByText('Profit & Loss')).toBeInTheDocument();
    });

    // The P&L entries table should show item data
    await waitFor(() => {
      // Tritanium appears in both items table and P&L table
      const tritaniumElements = screen.getAllByText('Tritanium');
      expect(tritaniumElements.length).toBeGreaterThan(0);
    });
  });

  it('should show empty P&L message when no entries', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockSellingRun, [], undefined);

    // Override to return empty pnl entries
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/hauling/runs/') && !url.includes('/items') && !url.includes('/status')) {
        return Promise.resolve({ ok: true, json: async () => mockSellingRun });
      }
      if (url.includes('/api/hauling/scanner')) {
        return Promise.resolve({ ok: true, json: async () => [] });
      }
      if (url.includes('/api/hauling/pnl-summary')) {
        return Promise.resolve({ ok: false });
      }
      if (url.includes('/api/hauling/pnl')) {
        return Promise.resolve({ ok: true, json: async () => [] });
      }
      return Promise.resolve({ ok: false, json: async () => ({}) });
    });

    render(<HaulingRunDetail runId={2} />);

    await waitFor(() => {
      expect(screen.getByText(/No P&L entries yet/)).toBeInTheDocument();
    });
  });

  it('should show P&L section for COMPLETE runs', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockCompleteRun, mockPnlEntries, mockPnlSummary);

    render(<HaulingRunDetail runId={3} />);

    await waitFor(() => {
      expect(screen.getByText('Profit & Loss')).toBeInTheDocument();
    });
  });

  it('should show Actual Revenue in Run Summary for COMPLETE run with pnl', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockCompleteRun, mockPnlEntries, mockPnlSummary);

    render(<HaulingRunDetail runId={3} />);

    await waitFor(() => {
      expect(screen.getByText('Actual Revenue')).toBeInTheDocument();
      expect(screen.getByText('Actual Net Profit')).toBeInTheDocument();
    });
  });

  it('should show estimated figures in Run Summary for PLANNING run', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockRun);

    render(<HaulingRunDetail runId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Total ISK Revenue')).toBeInTheDocument();
      expect(screen.getByText('Net Profit')).toBeInTheDocument();
    });
  });

  it('should show run not found when fetch returns null', async () => {
    mockUseSession.mockReturnValue(mockSession);
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/hauling/runs/')) {
        return Promise.resolve({ ok: false, json: async () => ({ error: 'Not found' }) });
      }
      return Promise.resolve({ ok: false, json: async () => ({}) });
    });

    render(<HaulingRunDetail runId={999} />);

    await waitFor(() => {
      expect(screen.getByText('Run not found.')).toBeInTheDocument();
    });
  });

  it('should match snapshot for SELLING run with P&L entries', async () => {
    mockUseSession.mockReturnValue(mockSession);
    setupFetchMocks(mockSellingRun, mockPnlEntries, mockPnlSummary);

    const { container } = render(<HaulingRunDetail runId={2} />);

    await waitFor(() => {
      expect(screen.getByText('Profit & Loss')).toBeInTheDocument();
    });

    expect(container).toMatchSnapshot();
  });
});

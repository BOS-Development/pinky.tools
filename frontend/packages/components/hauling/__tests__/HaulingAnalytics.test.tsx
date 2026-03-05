import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import HaulingAnalytics from '../HaulingAnalytics';
import {
  HaulingRouteAnalytics,
  HaulingItemAnalytics,
  HaulingProfitDataPoint,
  HaulingRunDurationSummary,
} from '@industry-tool/client/data/models';

jest.mock('next-auth/react');
jest.mock('../../loading', () => {
  return function MockLoading() {
    return <div data-testid="loading">Loading...</div>;
  };
});

// Mock recharts to avoid complex SVG rendering in jsdom
jest.mock('recharts', () => {
  return {
    LineChart: ({ children }: { children: React.ReactNode }) => <div data-testid="line-chart">{children}</div>,
    Line: () => null,
    XAxis: () => null,
    YAxis: () => null,
    CartesianGrid: () => null,
    Tooltip: () => null,
    Legend: () => null,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
      <div data-testid="responsive-container">{children}</div>
    ),
  };
});

// Silence next/image warning in jsdom
jest.mock('next/image', () => ({
  __esModule: true,
  default: ({ src, alt }: { src: string; alt: string }) => <img src={src} alt={alt} />,
}));

const mockUseSession = useSession as jest.MockedFunction<typeof useSession>;
const mockFetch = global.fetch as jest.Mock;

const mockSession = {
  data: { providerAccountId: '123', user: { name: 'Test User' } } as any,
  status: 'authenticated' as const,
  update: jest.fn(),
};

const mockRoutes: HaulingRouteAnalytics[] = [
  {
    fromRegionId: 10000002,
    toRegionId: 10000043,
    totalRuns: 5,
    totalProfitIsk: 500000000,
    avgProfitIsk: 100000000,
    avgMarginPct: 12.5,
    avgIskPerM3: 300,
    bestRunProfitIsk: 200000000,
    worstRunProfitIsk: 50000000,
  },
];

const mockItems: HaulingItemAnalytics[] = [
  {
    typeId: 34,
    typeName: 'Tritanium',
    totalRuns: 3,
    totalQtySold: 150000,
    totalProfitIsk: 250000000,
    avgMarginPct: 10.0,
  },
];

const mockTimeseries: HaulingProfitDataPoint[] = [
  { date: '2026-01-10', fromRegionId: 10000002, toRegionId: 10000043, profitIsk: 100000000, runCount: 1 },
  { date: '2026-01-15', fromRegionId: 10000002, toRegionId: 10000043, profitIsk: 150000000, runCount: 1 },
];

const mockSummary: HaulingRunDurationSummary = {
  totalCompletedRuns: 5,
  avgDurationDays: 4.2,
  minDurationDays: 1.5,
  maxDurationDays: 8.0,
  totalProfitIsk: 500000000,
};

function setupMocks(
  routes: HaulingRouteAnalytics[] = mockRoutes,
  items: HaulingItemAnalytics[] = mockItems,
  timeseries: HaulingProfitDataPoint[] = mockTimeseries,
  summary: HaulingRunDurationSummary = mockSummary,
) {
  mockFetch
    .mockResolvedValueOnce({ ok: true, json: async () => routes } as Response)
    .mockResolvedValueOnce({ ok: true, json: async () => items } as Response)
    .mockResolvedValueOnce({ ok: true, json: async () => timeseries } as Response)
    .mockResolvedValueOnce({ ok: true, json: async () => summary } as Response);
}

describe('HaulingAnalytics Component', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-02-22T12:00:00Z'));
    mockFetch.mockClear();
    mockUseSession.mockReturnValue(mockSession);
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('should show loading initially', () => {
    mockFetch.mockImplementation(() => new Promise(() => {}));

    render(<HaulingAnalytics />);
    expect(screen.getByTestId('loading')).toBeInTheDocument();
  });

  it('should show overview cards after loading', async () => {
    setupMocks();
    render(<HaulingAnalytics />);

    await waitFor(() => {
      expect(screen.getByText('Overview')).toBeInTheDocument();
    });

    expect(screen.getByText('Completed Runs')).toBeInTheDocument();
    // "Total Profit" appears in card label and table headers — just check it's present
    expect(screen.getAllByText('Total Profit').length).toBeGreaterThan(0);
    expect(screen.getByText('Avg Run Duration')).toBeInTheDocument();
    expect(screen.getByText('Avg Margin')).toBeInTheDocument();
  });

  it('should display total completed runs count', async () => {
    setupMocks();
    render(<HaulingAnalytics />);

    await waitFor(() => {
      // The stat card shows "5" as the total completed runs
      expect(screen.getByText('Completed Runs')).toBeInTheDocument();
    });

    // The "5" value is in the stat card — there may be multiple elements with "5"
    // so use getAllByText and check at least one exists
    expect(screen.getAllByText('5').length).toBeGreaterThan(0);
  });

  it('should show route performance section', async () => {
    setupMocks();
    render(<HaulingAnalytics />);

    await waitFor(() => {
      expect(screen.getByText('Route Performance')).toBeInTheDocument();
    });

    expect(screen.getByText(/The Forge/)).toBeInTheDocument();
    expect(screen.getByText(/Domain/)).toBeInTheDocument();
  });

  it('should show item performance section', async () => {
    setupMocks();
    render(<HaulingAnalytics />);

    await waitFor(() => {
      expect(screen.getByText('Item Performance')).toBeInTheDocument();
    });

    expect(screen.getByText('Tritanium')).toBeInTheDocument();
  });

  it('should show run duration section when runs exist', async () => {
    setupMocks();
    render(<HaulingAnalytics />);

    await waitFor(() => {
      expect(screen.getByText('Run Duration')).toBeInTheDocument();
    });

    expect(screen.getByText('Average Duration')).toBeInTheDocument();
    expect(screen.getByText('Fastest Run')).toBeInTheDocument();
    expect(screen.getByText('Slowest Run')).toBeInTheDocument();
  });

  it('should not show duration section when no completed runs', async () => {
    const emptySummary: HaulingRunDurationSummary = {
      ...mockSummary,
      totalCompletedRuns: 0,
    };
    setupMocks(mockRoutes, mockItems, mockTimeseries, emptySummary);
    render(<HaulingAnalytics />);

    await waitFor(() => {
      expect(screen.getByText('Overview')).toBeInTheDocument();
    });

    expect(screen.queryByText('Run Duration')).not.toBeInTheDocument();
  });

  it('should show empty state in route table when no routes', async () => {
    setupMocks([], mockItems, mockTimeseries, mockSummary);
    render(<HaulingAnalytics />);

    await waitFor(() => {
      expect(screen.getByText(/No route data yet/)).toBeInTheDocument();
    });
  });

  it('should show empty state in item table when no items', async () => {
    setupMocks(mockRoutes, [], mockTimeseries, mockSummary);
    render(<HaulingAnalytics />);

    await waitFor(() => {
      expect(screen.getByText(/No item data yet/)).toBeInTheDocument();
    });
  });

  it('should match snapshot with data loaded', async () => {
    setupMocks();
    const { container } = render(<HaulingAnalytics />);

    await waitFor(() => {
      expect(screen.getByText('Tritanium')).toBeInTheDocument();
    });

    expect(container).toMatchSnapshot();
  });
});

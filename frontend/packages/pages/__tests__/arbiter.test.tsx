import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import ArbiterPage from '../arbiter';

// Mock next-auth
jest.mock('next-auth/react');
const mockUseSession = useSession as jest.MockedFunction<typeof useSession>;

// Mock next/head
jest.mock('next/head', () => ({
  __esModule: true,
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

// Mock Navbar
jest.mock('@industry-tool/components/Navbar', () => ({
  __esModule: true,
  default: () => <nav data-testid="navbar" />,
}));

// Mock Loading and Unauthorized
jest.mock('@industry-tool/components/loading', () => ({
  __esModule: true,
  default: () => <div data-testid="loading" />,
}));
jest.mock('@industry-tool/components/unauthorized', () => ({
  __esModule: true,
  default: () => <div data-testid="unauthorized" />,
}));

// Mock shadcn Select (Radix doesn't work in jsdom)
jest.mock('@/components/ui/select', () => ({
  Select: ({ children, value, onValueChange }: any) => (
    <div data-testid="select" data-value={value} onClick={() => onValueChange?.('raitaru')}>
      {children}
    </div>
  ),
  SelectTrigger: ({ children }: any) => <div role="combobox">{children}</div>,
  SelectValue: ({ placeholder }: any) => <span>{placeholder}</span>,
  SelectContent: ({ children }: any) => <div>{children}</div>,
  SelectItem: ({ children, value }: any) => (
    <div role="option" data-value={value}>
      {children}
    </div>
  ),
  SelectGroup: ({ children }: any) => <div>{children}</div>,
  SelectLabel: ({ children }: any) => <div>{children}</div>,
}));

// Mock shadcn Collapsible (Radix)
jest.mock('@/components/ui/collapsible', () => ({
  Collapsible: ({ children, open, onOpenChange }: any) => (
    <div data-testid="collapsible" data-open={open}>
      {children}
    </div>
  ),
  CollapsibleTrigger: ({ children, asChild }: any) => (
    <div data-testid="collapsible-trigger">{children}</div>
  ),
  CollapsibleContent: ({ children }: any) => (
    <div data-testid="collapsible-content">{children}</div>
  ),
}));

const mockFetch = jest.fn();
global.fetch = mockFetch;

const mockOpportunitiesResponse = {
  opportunities: [
    {
      product_type_id: 12345,
      product_name: 'Vagabond',
      category: 'ship',
      jita_sell_price: 450000000,
      jita_buy_price: 420000000,
      best_decryptor: {
        type_id: 34201,
        name: 'Accelerant Decryptor',
        probability_multiplier: 1.2,
        me_modifier: 2,
        te_modifier: 10,
        run_modifier: 1,
        resulting_me: 4,
        resulting_runs: 2,
        invention_cost: 15000000,
        material_cost: 290000000,
        job_cost: 15000000,
        total_cost: 320000000,
        profit: 130000000,
        roi: 40.6,
        isk_per_day: 130000000,
        build_time_sec: 86400,
      },
      all_decryptors: [
        {
          type_id: 0,
          name: 'No Decryptor',
          probability_multiplier: 1.0,
          me_modifier: 0,
          te_modifier: 0,
          run_modifier: 0,
          resulting_me: 2,
          resulting_runs: 1,
          invention_cost: 12000000,
          material_cost: 290000000,
          job_cost: 12000000,
          total_cost: 302000000,
          profit: 118000000,
          roi: 39.1,
          isk_per_day: 118000000,
          build_time_sec: 86400,
        },
        {
          type_id: 34201,
          name: 'Accelerant Decryptor',
          probability_multiplier: 1.2,
          me_modifier: 2,
          te_modifier: 10,
          run_modifier: 1,
          resulting_me: 4,
          resulting_runs: 2,
          invention_cost: 15000000,
          material_cost: 290000000,
          job_cost: 15000000,
          total_cost: 320000000,
          profit: 130000000,
          roi: 40.6,
          isk_per_day: 130000000,
          build_time_sec: 86400,
        },
      ],
    },
    {
      product_type_id: 67890,
      product_name: 'Caldari Navy Kinetic Plating',
      category: 'module',
      jita_sell_price: 12000000,
      jita_buy_price: 11000000,
      best_decryptor: {
        type_id: 34202,
        name: 'Symmetry Decryptor',
        probability_multiplier: 1.5,
        me_modifier: 2,
        te_modifier: -2,
        run_modifier: 3,
        resulting_me: 4,
        resulting_runs: 4,
        invention_cost: 2000000,
        material_cost: 8000000,
        job_cost: 1000000,
        total_cost: 11000000,
        profit: 1000000,
        roi: 9.1,
        isk_per_day: 4000000,
        build_time_sec: 21600,
      },
      all_decryptors: [],
    },
  ],
  generated_at: '2026-03-19T12:00:00Z',
  total_scanned: 150,
  best_character_name: 'Inventor Alt',
};

describe('ArbiterPage', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-03-19T12:00:00Z'));
    mockFetch.mockClear();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('renders loading state', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'loading',
      update: jest.fn(),
    });
    const { container } = render(<ArbiterPage />);
    expect(screen.getByTestId('loading')).toBeInTheDocument();
    expect(container).toMatchSnapshot();
  });

  it('renders unauthorized state', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated',
      update: jest.fn(),
    });
    const { container } = render(<ArbiterPage />);
    expect(screen.getByTestId('unauthorized')).toBeInTheDocument();
    expect(container).toMatchSnapshot();
  });

  it('renders initial authenticated state with empty results', () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });
    const { container } = render(<ArbiterPage />);
    expect(screen.getByTestId('navbar')).toBeInTheDocument();
    expect(screen.getByText('Arbiter')).toBeInTheDocument();
    expect(screen.getByText('Industry Advisor')).toBeInTheDocument();
    expect(
      screen.getByText(/configure your structures above/i)
    ).toBeInTheDocument();
    expect(container).toMatchSnapshot();
  });

  it('renders settings panel with structure sections', () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });
    render(<ArbiterPage />);
    expect(screen.getByText('Production Structure Settings')).toBeInTheDocument();
    expect(screen.getByText('Reaction')).toBeInTheDocument();
    expect(screen.getByText('Invention')).toBeInTheDocument();
    expect(screen.getByText('Component Build')).toBeInTheDocument();
    expect(screen.getByText('Final Build')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /save settings/i })).toBeInTheDocument();
  });

  it('renders filter and sort controls', () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });
    render(<ArbiterPage />);
    expect(screen.getByRole('button', { name: /all/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ships/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /modules/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /net profit/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /roi/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /isk\/day/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /scan opportunities/i })).toBeInTheDocument();
  });

  it('shows scanning state when scan is triggered', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    // Make fetch hang so we can observe the scanning state
    mockFetch.mockImplementation(() => new Promise(() => {}));

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText(/scanning t2 blueprints/i)).toBeInTheDocument();
    });
  });

  it('renders opportunities table after successful scan', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockOpportunitiesResponse,
    } as Response);

    const { container } = render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Vagabond')).toBeInTheDocument();
    });

    expect(screen.getByText('Caldari Navy Kinetic Plating')).toBeInTheDocument();
    expect(screen.getByText('Scanned 150 blueprints')).toBeInTheDocument();
    expect(screen.getByText('Inventor Alt')).toBeInTheDocument();
    expect(container).toMatchSnapshot();
  });

  it('renders error state after failed scan', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockResolvedValueOnce({
      ok: false,
      text: async () => 'Internal server error',
    } as Response);

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Internal server error')).toBeInTheDocument();
    });
  });

  it('expands row to show decryptor sub-table', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockOpportunitiesResponse,
    } as Response);

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Vagabond')).toBeInTheDocument();
    });

    // Click the Vagabond row to expand it
    fireEvent.click(screen.getByText('Vagabond'));

    await waitFor(() => {
      expect(screen.getByText('No Decryptor')).toBeInTheDocument();
      // "Accelerant Decryptor" appears in both the main row (Best Decryptor column) and the sub-table
      expect(screen.getAllByText('Accelerant Decryptor').length).toBeGreaterThanOrEqual(2);
    });
  });

  it('filters by ship category', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockOpportunitiesResponse,
    } as Response);

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Vagabond')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /ships/i }));
    expect(screen.getByText('Vagabond')).toBeInTheDocument();
    expect(screen.queryByText('Caldari Navy Kinetic Plating')).not.toBeInTheDocument();
  });

  it('filters by module category', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockOpportunitiesResponse,
    } as Response);

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Caldari Navy Kinetic Plating')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /modules/i }));
    expect(screen.queryByText('Vagabond')).not.toBeInTheDocument();
    expect(screen.getByText('Caldari Navy Kinetic Plating')).toBeInTheDocument();
  });

  it('filters by search text', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockOpportunitiesResponse,
    } as Response);

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Vagabond')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search items...');
    fireEvent.change(searchInput, { target: { value: 'vagabond' } });

    expect(screen.getByText('Vagabond')).toBeInTheDocument();
    expect(screen.queryByText('Caldari Navy Kinetic Plating')).not.toBeInTheDocument();
  });

  it('clears search text with clear button', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockOpportunitiesResponse,
    } as Response);

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Vagabond')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search items...');
    fireEvent.change(searchInput, { target: { value: 'vagabond' } });

    const clearButton = screen.getByRole('button', { name: /clear search/i });
    fireEvent.click(clearButton);

    expect(screen.getByText('Vagabond')).toBeInTheDocument();
    expect(screen.getByText('Caldari Navy Kinetic Plating')).toBeInTheDocument();
  });

  it('renders search input in opportunities section', () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });
    render(<ArbiterPage />);
    expect(screen.getByPlaceholderText('Search items...')).toBeInTheDocument();
  });

  it('shows category sort button in table header after scan', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockOpportunitiesResponse,
    } as Response);

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Vagabond')).toBeInTheDocument();
    });

    // Category header is a button with text "Category"
    const catSortBtn = screen.getByRole('button', { name: /^category$/i });
    expect(catSortBtn).toBeInTheDocument();
  });

  it('cycles category sort on header click', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockOpportunitiesResponse,
    } as Response);

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Vagabond')).toBeInTheDocument();
    });

    // Category sort cycles: none -> ships_first -> modules_first -> none
    // The button title changes with each state
    const catSortBtn = screen.getByRole('button', { name: /^category$/i });
    expect(catSortBtn).toHaveAttribute('title', 'Sort: ships first');

    fireEvent.click(catSortBtn);
    expect(catSortBtn).toHaveAttribute('title', 'Sort: modules first');

    fireEvent.click(catSortBtn);
    expect(catSortBtn).toHaveAttribute('title', 'Remove category sort');

    fireEvent.click(catSortBtn);
    expect(catSortBtn).toHaveAttribute('title', 'Sort: ships first');
  });

  it('saves settings on button click', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
    } as Response);

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /save settings/i }));

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/arbiter/settings',
        expect.objectContaining({ method: 'PUT' })
      );
    });
  });
});

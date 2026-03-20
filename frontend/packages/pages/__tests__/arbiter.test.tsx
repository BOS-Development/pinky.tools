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

// Mock next/link
jest.mock('next/link', () => ({
  __esModule: true,
  default: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
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
    <div data-testid="select" data-value={value}>
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
  Collapsible: ({ children, open }: any) => (
    <div data-testid="collapsible" data-open={open}>
      {children}
    </div>
  ),
  CollapsibleTrigger: ({ children }: any) => (
    <div data-testid="collapsible-trigger">{children}</div>
  ),
  CollapsibleContent: ({ children }: any) => (
    <div data-testid="collapsible-content">{children}</div>
  ),
}));

// Mock shadcn Tabs
jest.mock('@/components/ui/tabs', () => ({
  Tabs: ({ children, value, onValueChange }: any) => (
    <div data-testid="tabs" data-value={value}>
      {children}
    </div>
  ),
  TabsList: ({ children }: any) => <div role="tablist">{children}</div>,
  TabsTrigger: ({ children, value }: any) => (
    <button role="tab" data-value={value}>
      {children}
    </button>
  ),
  TabsContent: ({ children, value }: any) => (
    <div role="tabpanel" data-value={value}>
      {children}
    </div>
  ),
}));

// Mock shadcn Switch
jest.mock('@/components/ui/switch', () => ({
  Switch: ({ checked, onCheckedChange }: any) => (
    <input
      type="checkbox"
      checked={checked}
      onChange={(e) => onCheckedChange?.(e.target.checked)}
      data-testid="switch"
    />
  ),
}));

// Mock Tooltip (Radix)
jest.mock('@/components/ui/tooltip', () => ({
  TooltipProvider: ({ children }: any) => <>{children}</>,
  Tooltip: ({ children }: any) => <>{children}</>,
  TooltipTrigger: ({ children }: any) => <>{children}</>,
  TooltipContent: ({ children }: any) => <div>{children}</div>,
}));

const mockFetch = jest.fn();
global.fetch = mockFetch;

const mockOpportunitiesResponse = {
  opportunities: [
    {
      product_type_id: 12345,
      product_name: 'Vagabond',
      category: 'ship',
      group: 'Heavy Assault Cruisers',
      tech_level: 'Tech II',
      jita_sell_price: 450000000,
      jita_buy_price: 420000000,
      demand_per_day: 5.2,
      days_of_supply: 14.3,
      duration_sec: 86400,
      runs: 1,
      me: 4,
      te: 4,
      material_cost: 290000000,
      job_cost: 15000000,
      invention_cost: 15000000,
      total_cost: 320000000,
      revenue: 420000000,
      sales_tax: 16800000,
      broker_fee: 12600000,
      profit: 70600000,
      roi: 22.1,
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
        profit: 70600000,
        roi: 22.1,
        isk_per_day: 70600000,
        build_time_sec: 86400,
      },
      all_decryptors: [],
      is_blacklisted: false,
      is_whitelisted: false,
    },
    {
      product_type_id: 67890,
      product_name: 'Caldari Navy Kinetic Plating',
      category: 'module',
      group: 'Armor Platings',
      tech_level: 'Tech II',
      jita_sell_price: 12000000,
      jita_buy_price: 11000000,
      demand_per_day: 2.1,
      days_of_supply: 30.0,
      duration_sec: 21600,
      runs: 3,
      me: 2,
      te: 2,
      material_cost: 8000000,
      job_cost: 1000000,
      invention_cost: 2000000,
      total_cost: 11000000,
      revenue: 11000000,
      sales_tax: 440000,
      broker_fee: 330000,
      profit: 230000,
      roi: 2.1,
      best_decryptor: null,
      all_decryptors: [],
      is_blacklisted: false,
      is_whitelisted: false,
    },
  ],
  generated_at: '2026-03-19T12:00:00Z',
  total_scanned: 150,
  best_character_name: 'Inventor Alt',
};

// Default mock: fetch returns 404 for all initial data loads
function mockDefaultFetches() {
  mockFetch.mockImplementation((url: string) => {
    if (
      url.includes('/api/arbiter/settings') ||
      url.includes('/api/arbiter/tax-profile') ||
      url.includes('/api/arbiter/scopes') ||
      url.includes('/api/arbiter/decryptors') ||
      url.includes('/api/arbiter/blacklist') ||
      url.includes('/api/arbiter/whitelist')
    ) {
      return Promise.resolve({ ok: false, json: async () => null } as Response);
    }
    return Promise.resolve({ ok: false, json: async () => null } as Response);
  });
}

describe('ArbiterPage', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-03-19T12:00:00Z'));
    mockFetch.mockClear();
    mockDefaultFetches();
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

  it('renders initial authenticated state', () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });
    const { container } = render(<ArbiterPage />);
    expect(screen.getByTestId('navbar')).toBeInTheDocument();
    expect(screen.getByText('Arbiter')).toBeInTheDocument();
    expect(screen.getByText('Industry Advisor')).toBeInTheDocument();
    expect(container).toMatchSnapshot();
  });

  it('renders settings panel', () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });
    render(<ArbiterPage />);
    expect(screen.getByText('Settings')).toBeInTheDocument();
    // Settings tabs are rendered
    expect(screen.getByText('Structures')).toBeInTheDocument();
    expect(screen.getByText('Tax & Pricing')).toBeInTheDocument();
    expect(screen.getByText('Lists')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /save settings/i })).toBeInTheDocument();
  });

  it('renders structure sections', () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });
    render(<ArbiterPage />);
    expect(screen.getByText('Reaction')).toBeInTheDocument();
    expect(screen.getByText('Invention')).toBeInTheDocument();
    expect(screen.getByText('Component Build')).toBeInTheDocument();
    expect(screen.getByText('Final Build')).toBeInTheDocument();
  });

  it('renders main controls', () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });
    render(<ArbiterPage />);
    expect(screen.getByRole('button', { name: /scan opportunities/i })).toBeInTheDocument();
    // Price toggles
    const sellButtons = screen.getAllByText('Sell');
    expect(sellButtons.length).toBeGreaterThanOrEqual(2);
    const buyButtons = screen.getAllByText('Buy');
    expect(buyButtons.length).toBeGreaterThanOrEqual(2);
  });

  it('renders filter bar with category and tech filters', () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });
    render(<ArbiterPage />);
    expect(screen.getByText('ship')).toBeInTheDocument();
    expect(screen.getByText('module')).toBeInTheDocument();
    expect(screen.getByText('Tech II')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Search...')).toBeInTheDocument();
  });

  it('renders empty state before scan', () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });
    render(<ArbiterPage />);
    expect(
      screen.getByText(/configure your settings above/i),
    ).toBeInTheDocument();
  });

  it('shows scanning state when scan triggered', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    // Override to make the scan call hang
    mockFetch.mockImplementation(() => new Promise(() => {}));

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText(/scanning blueprints/i)).toBeInTheDocument();
    });
  });

  it('renders opportunities after successful scan', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    // The scan call is the only fetch that matters here (initial data loads fail with 404)
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/arbiter/opportunities')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockOpportunitiesResponse,
        } as Response);
      }
      return Promise.resolve({ ok: false } as Response);
    });

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

    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/arbiter/opportunities')) {
        return Promise.resolve({
          ok: false,
          text: async () => 'Internal server error',
        } as Response);
      }
      return Promise.resolve({ ok: false } as Response);
    });

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Internal server error')).toBeInTheDocument();
    });
  });

  it('filters by category checkbox', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/arbiter/opportunities')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockOpportunitiesResponse,
        } as Response);
      }
      return Promise.resolve({ ok: false } as Response);
    });

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Vagabond')).toBeInTheDocument();
    });

    // Both items show initially (both ship and module filters active)
    expect(screen.getByText('Caldari Navy Kinetic Plating')).toBeInTheDocument();

    // Toggle off 'module' filter - find the filter bar button (not the badge in results)
    // The filter buttons are rounded-full, while badges in results are different
    const moduleFilterBtns = screen.getAllByText('module');
    // The filter bar button is a <button> element (not a div badge), click it
    const moduleFilterBtn = moduleFilterBtns.find(
      (el) => el.tagName.toLowerCase() === 'button',
    );
    expect(moduleFilterBtn).toBeTruthy();
    fireEvent.click(moduleFilterBtn!);

    // Caldari module should disappear
    expect(screen.queryByText('Caldari Navy Kinetic Plating')).not.toBeInTheDocument();
    expect(screen.getByText('Vagabond')).toBeInTheDocument();
  });

  it('filters by search text', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/arbiter/opportunities')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockOpportunitiesResponse,
        } as Response);
      }
      return Promise.resolve({ ok: false } as Response);
    });

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Vagabond')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search...');
    fireEvent.change(searchInput, { target: { value: 'vagabond' } });

    expect(screen.getByText('Vagabond')).toBeInTheDocument();
    expect(screen.queryByText('Caldari Navy Kinetic Plating')).not.toBeInTheDocument();
  });

  it('clears search text', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/arbiter/opportunities')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockOpportunitiesResponse,
        } as Response);
      }
      return Promise.resolve({ ok: false } as Response);
    });

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Vagabond')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search...');
    fireEvent.change(searchInput, { target: { value: 'vagabond' } });
    expect(screen.queryByText('Caldari Navy Kinetic Plating')).not.toBeInTheDocument();

    // Clear by setting empty value
    fireEvent.change(searchInput, { target: { value: '' } });

    await waitFor(() => {
      expect(screen.getByText('Caldari Navy Kinetic Plating')).toBeInTheDocument();
    });
  });

  it('saves settings when button clicked', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/arbiter/settings') || url.includes('/api/arbiter/tax-profile')) {
        return Promise.resolve({
          ok: true,
          json: async () => ({}),
        } as Response);
      }
      return Promise.resolve({ ok: false } as Response);
    });

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /save settings/i }));

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/arbiter/settings',
        expect.objectContaining({ method: 'PUT' }),
      );
    });
  });

  it('renders warehouse panel', () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });
    render(<ArbiterPage />);
    expect(screen.getByText('Ingredients')).toBeInTheDocument();
    expect(screen.getByText('Shopping List')).toBeInTheDocument();
    expect(screen.getByText('Sell Orders')).toBeInTheDocument();
    expect(screen.getByText('Buy Orders')).toBeInTheDocument();
  });

  it('renders column headers after scan', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/api/arbiter/opportunities')) {
        return Promise.resolve({
          ok: true,
          json: async () => mockOpportunitiesResponse,
        } as Response);
      }
      return Promise.resolve({ ok: false } as Response);
    });

    render(<ArbiterPage />);
    fireEvent.click(screen.getByRole('button', { name: /scan opportunities/i }));

    await waitFor(() => {
      expect(screen.getByText('Vagabond')).toBeInTheDocument();
    });

    // Check key column headers
    expect(screen.getByText('Name')).toBeInTheDocument();
    expect(screen.getByText('ROI')).toBeInTheDocument();
    expect(screen.getByText('Profit')).toBeInTheDocument();
    expect(screen.getByText('Category')).toBeInTheDocument();
  });
});

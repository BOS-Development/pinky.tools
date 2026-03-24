import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import ArbiterListsPage from '../arbiter-lists';

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

const mockFetch = jest.fn();
global.fetch = mockFetch;

const mockWhitelist = [
  { type_id: 22436, name: 'Widow' },
  { type_id: 12005, name: 'Vagabond' },
];

const mockBlacklist = [
  { type_id: 55555, name: 'Ibis' },
];

describe('ArbiterListsPage', () => {
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
    const { container } = render(<ArbiterListsPage />);
    expect(screen.getByTestId('loading')).toBeInTheDocument();
    expect(container).toMatchSnapshot();
  });

  it('renders unauthorized state', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated',
      update: jest.fn(),
    });
    const { container } = render(<ArbiterListsPage />);
    expect(screen.getByTestId('unauthorized')).toBeInTheDocument();
    expect(container).toMatchSnapshot();
  });

  it('renders page with whitelist and blacklist after load', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockWhitelist,
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockBlacklist,
      } as Response);

    const { container } = render(<ArbiterListsPage />);

    await waitFor(() => {
      expect(screen.getByText('Widow')).toBeInTheDocument();
    });

    expect(screen.getByText('Vagabond')).toBeInTheDocument();
    expect(screen.getByText('Ibis')).toBeInTheDocument();
    expect(screen.getByText(/whitelist \(2\)/i)).toBeInTheDocument();
    expect(screen.getByText(/blacklist \(1\)/i)).toBeInTheDocument();
    expect(container).toMatchSnapshot();
  });

  it('renders empty lists when data is empty', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      } as Response);

    render(<ArbiterListsPage />);

    await waitFor(() => {
      expect(screen.getByText('No items in whitelist')).toBeInTheDocument();
    });

    expect(screen.getByText('No items in blacklist')).toBeInTheDocument();
    expect(screen.getByText(/whitelist \(0\)/i)).toBeInTheDocument();
    expect(screen.getByText(/blacklist \(0\)/i)).toBeInTheDocument();
  });

  it('renders back link to arbiter page', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch
      .mockResolvedValueOnce({ ok: true, json: async () => [] } as Response)
      .mockResolvedValueOnce({ ok: true, json: async () => [] } as Response);

    render(<ArbiterListsPage />);

    await waitFor(() => {
      const backLink = screen.getByRole('link', { name: '' });
      expect(backLink).toHaveAttribute('href', '/arbiter');
    });
  });

  it('renders search boxes for adding items', async () => {
    mockUseSession.mockReturnValue({
      data: { user: { name: 'Test User' }, providerAccountId: '12345' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    mockFetch
      .mockResolvedValueOnce({ ok: true, json: async () => [] } as Response)
      .mockResolvedValueOnce({ ok: true, json: async () => [] } as Response);

    render(<ArbiterListsPage />);

    await waitFor(() => {
      const searchBoxes = screen.getAllByPlaceholderText('Search items to add...');
      expect(searchBoxes.length).toBe(2);
    });
  });
});

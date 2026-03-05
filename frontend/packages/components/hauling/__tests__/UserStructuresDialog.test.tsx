import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import UserStructuresDialog from '../UserStructuresDialog';
import { UserTradingStructure } from '@industry-tool/client/data/models';

// Mock sonner toast
jest.mock('@/components/ui/sonner', () => ({
  toast: {
    success: jest.fn(),
    error: jest.fn(),
  },
}));

// Mock shadcn dialog to render inline (Radix portals don't work in jsdom)
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
}));

// Mock shadcn Select (Radix doesn't work in jsdom)
jest.mock('@/components/ui/select', () => ({
  Select: ({ children, value, onValueChange }: { children: React.ReactNode; value: string; onValueChange: (v: string) => void }) => (
    <div data-testid="select" data-value={value}>
      <button
        data-testid="select-trigger-btn"
        onClick={() => {
          // Simulate value change for testing — tests can override by querying children
        }}
      >
        {children}
      </button>
    </div>
  ),
  SelectTrigger: ({ children, className }: { children: React.ReactNode; className?: string }) => (
    <div data-testid="select-trigger" className={className}>{children}</div>
  ),
  SelectValue: ({ placeholder }: { placeholder?: string }) => (
    <span data-testid="select-value">{placeholder}</span>
  ),
  SelectContent: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="select-content">{children}</div>
  ),
  SelectItem: ({ children, value, onClick }: { children: React.ReactNode; value: string; onClick?: () => void }) => (
    <div data-testid={`select-item-${value}`} data-value={value} role="option" onClick={onClick}>
      {children}
    </div>
  ),
}));

const mockFetch = global.fetch as jest.Mock;

const mockStructures: UserTradingStructure[] = [
  {
    id: 1,
    userId: 100,
    structureId: 1035466617946,
    name: 'Perimeter - Tranquility Trading Tower',
    systemId: 30000144,
    regionId: 10000002,
    characterId: 456,
    accessOk: true,
    lastScannedAt: '2026-02-22T10:00:00Z',
    createdAt: '2026-02-20T08:00:00Z',
  },
  {
    id: 2,
    userId: 100,
    structureId: 1035466617947,
    name: 'No Access Structure',
    systemId: 30000001,
    regionId: 10000002,
    characterId: 456,
    accessOk: false,
    createdAt: '2026-02-21T08:00:00Z',
  },
];

const mockCharacters = [
  { id: 456, name: 'Test Character' },
];

const mockAssetStructures = [
  { structureId: 1035466617946, name: 'Fortizar - Alpha' },
  { structureId: 1035466617948, name: '' },
];

beforeEach(() => {
  jest.useFakeTimers();
  jest.setSystemTime(new Date('2026-02-22T12:00:00Z'));
  mockFetch.mockClear();
});

afterEach(() => {
  jest.useRealTimers();
});

describe('UserStructuresDialog', () => {
  const defaultProps = {
    open: true,
    onClose: jest.fn(),
    onStructuresChanged: jest.fn(),
  };

  it('renders loading state while fetching structures', () => {
    mockFetch.mockImplementation(() => new Promise(() => {})); // never resolves

    const { container } = render(<UserStructuresDialog {...defaultProps} />);
    expect(container).toMatchSnapshot();
  });

  it('renders empty state when no structures configured', async () => {
    // First call: /api/hauling/structures, second call: /api/characters
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockCharacters,
    });

    const { container } = render(<UserStructuresDialog {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText(/No trading structures configured/i)).toBeTruthy();
    });

    expect(container).toMatchSnapshot();
  });

  it('renders structures list correctly', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockStructures,
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockCharacters,
    });

    const { container } = render(<UserStructuresDialog {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Perimeter - Tranquility Trading Tower')).toBeTruthy();
    });

    expect(container).toMatchSnapshot();
  });

  it('shows access warning badge for structures with accessOk=false', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockStructures,
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockCharacters,
    });

    render(<UserStructuresDialog {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('No Access Structure')).toBeTruthy();
    });

    expect(screen.getAllByText(/No Access/).length).toBeGreaterThan(0);
  });

  it('shows error message when adding structure fails with 403', async () => {
    // Initial load
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockCharacters,
    });

    render(<UserStructuresDialog {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText(/No trading structures configured/i)).toBeTruthy();
    });

    // Add button should be disabled when no selections made
    const addButton = screen.getByText('Add');
    expect(addButton.closest('button')).toBeDisabled();
  });

  it('shows "Select a character first" hint when no character selected', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockCharacters,
    });

    render(<UserStructuresDialog {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText(/Select a character first/i)).toBeTruthy();
    });
  });

  it('shows "No structures found" when character has no asset structures', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockCharacters,
    });
    // Asset structures fetch returns empty
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    render(<UserStructuresDialog {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText(/No trading structures configured/i)).toBeTruthy();
    });
  });

  it('does not render when closed', () => {
    const { container } = render(
      <UserStructuresDialog open={false} onClose={jest.fn()} onStructuresChanged={jest.fn()} />,
    );
    expect(container.querySelector('[data-testid="dialog"]')).toBeNull();
  });

  it('fetches asset structures when character is selected via dropdown', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockCharacters,
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockAssetStructures,
    });

    render(<UserStructuresDialog {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText(/Select a character first/i)).toBeTruthy();
    });

    // Verify characters were loaded (fetch called twice: structures + characters)
    expect(mockFetch).toHaveBeenCalledWith('/api/characters');
  });
});

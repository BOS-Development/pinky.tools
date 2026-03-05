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
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
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

    render(<UserStructuresDialog {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('No Access Structure')).toBeTruthy();
    });

    expect(screen.getAllByText(/No Access/).length).toBeGreaterThan(0);
  });

  it('shows error message when adding structure with no access', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    render(<UserStructuresDialog {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText(/No trading structures configured/i)).toBeTruthy();
    });

    // Fill in add form
    fireEvent.change(screen.getByLabelText(/Structure ID/i), {
      target: { value: '1035466617946' },
    });
    fireEvent.change(screen.getByLabelText(/Character ID/i), {
      target: { value: '456' },
    });

    // Mock 403 access denied response
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 403,
      json: async () => ({ accessOk: false, error: 'Character does not have docking rights.' }),
    });

    fireEvent.click(screen.getByText('Add'));

    await waitFor(() => {
      expect(screen.getByText(/Character does not have docking rights/i)).toBeTruthy();
    });
  });

  it('does not render when closed', () => {
    const { container } = render(
      <UserStructuresDialog open={false} onClose={jest.fn()} onStructuresChanged={jest.fn()} />,
    );
    expect(container.querySelector('[data-testid="dialog"]')).toBeNull();
  });
});

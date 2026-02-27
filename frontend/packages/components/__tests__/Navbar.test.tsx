import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import Navbar from '../Navbar';

jest.mock('next-auth/react');

const mockUseSession = useSession as jest.MockedFunction<typeof useSession>;

describe('Navbar Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (global.fetch as jest.Mock).mockClear();
  });

  it('should match snapshot when not authenticated', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated',
      update: jest.fn(),
    });

    const { container } = render(<Navbar />);
    expect(container).toMatchSnapshot();
  });

  it('should display app title', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated',
      update: jest.fn(),
    });

    render(<Navbar />);
    expect(screen.getByText('EVE Industry Tool')).toBeInTheDocument();
  });

  it('should have all navigation links', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated',
      update: jest.fn(),
    });

    render(<Navbar />);

    // All 5 dropdown trigger buttons should always be visible
    expect(screen.getByRole('button', { name: /account/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /assets/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /trading/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /industry/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /logistics/i })).toBeInTheDocument();
    // Settings is a standalone button/link
    expect(screen.getByRole('link', { name: /settings/i })).toBeInTheDocument();

    // --- Account dropdown ---
    fireEvent.click(screen.getByRole('button', { name: /account/i }));
    expect(screen.getByRole('menuitem', { name: /characters/i })).toHaveAttribute('href', '/characters');
    expect(screen.getByRole('menuitem', { name: /corporations/i })).toHaveAttribute('href', '/corporations');
    // Close by clicking the button again
    fireEvent.click(screen.getByRole('button', { name: /account/i }));

    // --- Assets dropdown ---
    fireEvent.click(screen.getByRole('button', { name: /assets/i }));
    expect(screen.getByRole('menuitem', { name: /inventory/i })).toHaveAttribute('href', '/inventory');
    expect(screen.getByRole('menuitem', { name: /stockpiles/i })).toHaveAttribute('href', '/stockpiles');
    fireEvent.click(screen.getByRole('button', { name: /assets/i }));

    // --- Trading dropdown ---
    fireEvent.click(screen.getByRole('button', { name: /trading/i }));
    expect(screen.getByRole('menuitem', { name: /contacts/i })).toHaveAttribute('href', '/contacts');
    expect(screen.getByRole('menuitem', { name: /marketplace/i })).toHaveAttribute('href', '/marketplace');
    fireEvent.click(screen.getByRole('button', { name: /trading/i }));

    // --- Industry dropdown ---
    fireEvent.click(screen.getByRole('button', { name: /industry/i }));
    expect(screen.getByRole('menuitem', { name: /reactions/i })).toHaveAttribute('href', '/reactions');
    expect(screen.getByRole('menuitem', { name: /^industry$/i })).toHaveAttribute('href', '/industry');
    expect(screen.getByRole('menuitem', { name: /plans/i })).toHaveAttribute('href', '/production-plans');
    expect(screen.getByRole('menuitem', { name: /runs/i })).toHaveAttribute('href', '/plan-runs');
    expect(screen.getByRole('menuitem', { name: /planets/i })).toHaveAttribute('href', '/pi');
    fireEvent.click(screen.getByRole('button', { name: /industry/i }));

    // --- Logistics dropdown ---
    fireEvent.click(screen.getByRole('button', { name: /logistics/i }));
    expect(screen.getByRole('menuitem', { name: /transport/i })).toHaveAttribute('href', '/transport');
    expect(screen.getByRole('menuitem', { name: /stations/i })).toHaveAttribute('href', '/stations');
    fireEvent.click(screen.getByRole('button', { name: /logistics/i }));
  });

  it('should have rocket icon', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated',
      update: jest.fn(),
    });

    render(<Navbar />);
    const menuButton = screen.getByLabelText('menu');
    expect(menuButton).toBeInTheDocument();
  });

  it('should not fetch contacts when user is not authenticated', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated',
      update: jest.fn(),
    });

    render(<Navbar />);
    expect(global.fetch).not.toHaveBeenCalled();
  });

  it('should fetch contacts when user is authenticated', async () => {
    mockUseSession.mockReturnValue({
      data: { providerAccountId: '123' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => [],
    });

    render(<Navbar />);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith('/api/contacts');
    });
  });

  it('should display badge with pending contact count', async () => {
    mockUseSession.mockReturnValue({
      data: { providerAccountId: '123' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    const mockContacts = [
      {
        id: 1,
        requesterUserId: 456,
        recipientUserId: 123,
        requesterName: 'Test User 1',
        recipientName: 'Current User',
        status: 'pending',
        requestedAt: '2024-01-01',
      },
      {
        id: 2,
        requesterUserId: 789,
        recipientUserId: 123,
        requesterName: 'Test User 2',
        recipientName: 'Current User',
        status: 'pending',
        requestedAt: '2024-01-02',
      },
      {
        id: 3,
        requesterUserId: 123,
        recipientUserId: 999,
        requesterName: 'Current User',
        recipientName: 'Test User 3',
        status: 'pending',
        requestedAt: '2024-01-03',
      },
    ];

    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockContacts,
    });

    render(<Navbar />);

    // Badge count "2" appears on the Trading trigger button (visible without opening)
    await waitFor(() => {
      const badge = screen.getByText('2');
      expect(badge).toBeInTheDocument();
    });

    // Also open the Trading dropdown and verify the badge on the Contacts item
    fireEvent.click(screen.getByRole('button', { name: /trading/i }));
    // The Contacts menuitem should be present and its badge should also show "2"
    const contactsItem = screen.getByRole('menuitem', { name: /contacts/i });
    expect(contactsItem).toBeInTheDocument();
    // There should be two elements showing "2" — one on the trigger, one inside the dropdown
    const badgeElements = screen.getAllByText('2');
    expect(badgeElements.length).toBeGreaterThanOrEqual(2);
  });

  it('should not display badge when there are no pending requests', async () => {
    mockUseSession.mockReturnValue({
      data: { providerAccountId: '123' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    const mockContacts = [
      {
        id: 1,
        requesterUserId: 123,
        recipientUserId: 456,
        requesterName: 'Current User',
        recipientName: 'Test User',
        status: 'accepted',
        requestedAt: '2024-01-01',
      },
    ];

    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockContacts,
    });

    render(<Navbar />);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalled();
    });

    // Badge should not be visible when count is 0
    expect(screen.queryByText('0')).not.toBeInTheDocument();
  });

  it('should handle fetch errors gracefully', async () => {
    mockUseSession.mockReturnValue({
      data: { providerAccountId: '123' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    (global.fetch as jest.Mock).mockResolvedValue({
      ok: false,
    });

    const consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation();

    render(<Navbar />);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalled();
    });

    // Should not crash on error
    expect(screen.getByText('EVE Industry Tool')).toBeInTheDocument();

    consoleErrorSpy.mockRestore();
  });

  it('should set up polling interval for contact updates', async () => {
    jest.useFakeTimers();

    mockUseSession.mockReturnValue({
      data: { providerAccountId: '123' } as any,
      status: 'authenticated',
      update: jest.fn(),
    });

    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => [],
    });

    render(<Navbar />);

    // On mount: contacts fetch + scope-status fetch (both fire once)
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(2);
    });

    // Fast forward 30 seconds — contacts polls again (scope-status does not poll)
    jest.advanceTimersByTime(30000);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(3);
    });

    // Fast forward another 30 seconds — contacts polls again
    jest.advanceTimersByTime(30000);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(4);
    });

    jest.useRealTimers();
  });
});

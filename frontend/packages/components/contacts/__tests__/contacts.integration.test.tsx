import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import ContactsList from '../ContactsList';

const mockToastError = jest.fn();
const mockToastSuccess = jest.fn();
jest.mock('@/components/ui/sonner', () => ({
  toast: {
    error: (...args: unknown[]) => mockToastError(...args),
    success: (...args: unknown[]) => mockToastSuccess(...args),
  },
}));

jest.mock('next-auth/react');
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
jest.mock('../PermissionsDialog', () => {
  return function MockPermissionsDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
    return open ? (
      <div data-testid="permissions-dialog">
        <button onClick={onClose}>Close</button>
      </div>
    ) : null;
  };
});

const mockUseSession = useSession as jest.MockedFunction<typeof useSession>;

// Helper to click Radix UI tabs (requires mouseDown event)
const clickTab = (tabText: string) => {
  const tab = screen.getByText(tabText);
  fireEvent.mouseDown(tab);
};

describe('Contacts Integration Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (global.fetch as jest.Mock).mockClear();
    mockToastError.mockClear();
    mockToastSuccess.mockClear();
  });

  const mockSession = {
    data: { providerAccountId: '123' } as any,
    status: 'authenticated' as const,
    update: jest.fn(),
  };

  it('should complete full contact request workflow', async () => {
    mockUseSession.mockReturnValue(mockSession);

    // Initial state: no contacts (contacts + contact-rules)
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],  // /api/contacts
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],  // /api/contact-rules
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(
        screen.getByText('No contacts yet. Add a contact to get started!')
      ).toBeInTheDocument();
    });

    // Open add contact dialog
    const addButton = screen.getByRole('button', { name: /add contact/i });
    fireEvent.click(addButton);

    // Enter character name
    const input = screen.getByLabelText(/character name/i);
    fireEvent.change(input, { target: { value: 'Test Character' } });

    // Mock successful contact creation + refetch
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 1,
          requesterUserId: 123,
          recipientUserId: 456,
          requesterName: 'Current User',
          recipientName: 'Test Character',
          status: 'pending',
          requestedAt: '2024-01-01T00:00:00Z',
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [
          {
            id: 1,
            requesterUserId: 123,
            recipientUserId: 456,
            requesterName: 'Current User',
            recipientName: 'Test Character',
            status: 'pending',
            requestedAt: '2024-01-01T00:00:00Z',
          },
        ],
      });

    // Send request
    const sendButton = screen.getByRole('button', { name: /send request/i });
    fireEvent.click(sendButton);

    // Should show success via toast
    await waitFor(() => {
      expect(mockToastSuccess).toHaveBeenCalledWith('Contact request sent!');
    });

    // Should now show in Sent tab
    await waitFor(() => {
      expect(screen.getByText('Sent (1)')).toBeInTheDocument();
    });

    clickTab('Sent (1)');

    await waitFor(() => {
      expect(screen.getByText('Test Character')).toBeInTheDocument();
    });
  });

  it('should complete accept contact workflow with permissions', async () => {
    mockUseSession.mockReturnValue(mockSession);

    const pendingContact = {
      id: 2,
      requesterUserId: 789,
      recipientUserId: 123,
      requesterName: 'Requester User',
      recipientName: 'Current User',
      status: 'pending' as const,
      requestedAt: '2024-01-01T00:00:00Z',
    };

    // Initial fetch with pending contact (contacts + contact-rules)
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [pendingContact],  // /api/contacts
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],  // /api/contact-rules
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Pending (1)')).toBeInTheDocument();
    });

    // Go to pending requests
    clickTab('Pending (1)');

    await waitFor(() => {
      expect(screen.getByText('Requester User')).toBeInTheDocument();
    });

    // Accept the contact
    const acceptedContact = {
      ...pendingContact,
      status: 'accepted' as const,
      respondedAt: '2024-01-02T00:00:00Z',
    };

    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => acceptedContact,
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [acceptedContact],
      });

    const acceptButton = screen.getByTitle('Accept');
    fireEvent.click(acceptButton);

    // Should show success via toast
    await waitFor(() => {
      expect(mockToastSuccess).toHaveBeenCalledWith('Contact accepted!');
    });

    // Should now appear in My Contacts
    await waitFor(() => {
      expect(screen.getByText('My Contacts (1)')).toBeInTheDocument();
    });

    clickTab('My Contacts (1)');

    await waitFor(() => {
      expect(screen.getByText('Requester User')).toBeInTheDocument();
      expect(screen.getByText('Connected')).toBeInTheDocument();
    });
  });

  it('should handle reject workflow', async () => {
    mockUseSession.mockReturnValue(mockSession);

    const pendingContact = {
      id: 3,
      requesterUserId: 999,
      recipientUserId: 123,
      requesterName: 'Unwanted User',
      recipientName: 'Current User',
      status: 'pending' as const,
      requestedAt: '2024-01-01T00:00:00Z',
    };

    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [pendingContact],  // /api/contacts
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],  // /api/contact-rules
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Pending (1)')).toBeInTheDocument();
    });

    clickTab('Pending (1)');

    await waitFor(() => {
      expect(screen.getByText('Unwanted User')).toBeInTheDocument();
    });

    // Reject the contact
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ ...pendingContact, status: 'rejected' }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      });

    const rejectButton = screen.getByTitle('Reject');
    fireEvent.click(rejectButton);

    await waitFor(() => {
      expect(mockToastSuccess).toHaveBeenCalledWith('Contact rejected');
    });

    // Should no longer show pending requests
    await waitFor(() => {
      expect(screen.getByText('Pending (0)')).toBeInTheDocument();
    });
  });

  it('should handle delete contact workflow', async () => {
    mockUseSession.mockReturnValue(mockSession);

    const acceptedContact = {
      id: 4,
      requesterUserId: 123,
      recipientUserId: 888,
      requesterName: 'Current User',
      recipientName: 'Old Friend',
      status: 'accepted' as const,
      requestedAt: '2024-01-01T00:00:00Z',
      respondedAt: '2024-01-02T00:00:00Z',
    };

    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [acceptedContact],  // /api/contacts
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],  // /api/contact-rules
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Old Friend')).toBeInTheDocument();
    });

    // Delete the contact
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      });

    const deleteButton = screen.getByTitle('Remove Contact');
    fireEvent.click(deleteButton);

    await waitFor(() => {
      expect(mockToastSuccess).toHaveBeenCalledWith('Contact removed');
    });

    // Should show empty state
    await waitFor(() => {
      expect(
        screen.getByText('No contacts yet. Add a contact to get started!')
      ).toBeInTheDocument();
    });
  });

  it('should handle error when adding duplicate contact', async () => {
    mockUseSession.mockReturnValue(mockSession);

    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],  // /api/contacts
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],  // /api/contact-rules
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Add Contact')).toBeInTheDocument();
    });

    // Open add contact dialog
    const addButton = screen.getByRole('button', { name: /add contact/i });
    fireEvent.click(addButton);

    const input = screen.getByLabelText(/character name/i);
    fireEvent.change(input, { target: { value: 'Existing Contact' } });

    // Mock duplicate error
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ error: 'Contact already exists' }),
    });

    const sendButton = screen.getByRole('button', { name: /send request/i });
    fireEvent.click(sendButton);

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('Contact already exists');
    });
  });

  it('should handle self-contact error', async () => {
    mockUseSession.mockReturnValue(mockSession);

    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],  // /api/contacts
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],  // /api/contact-rules
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Add Contact')).toBeInTheDocument();
    });

    // Open add contact dialog
    const addButton = screen.getByRole('button', { name: /add contact/i });
    fireEvent.click(addButton);

    const input = screen.getByLabelText(/character name/i);
    fireEvent.change(input, { target: { value: 'Current User' } });

    // Mock self-contact error
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ error: 'cannot add yourself as a contact' }),
    });

    const sendButton = screen.getByRole('button', { name: /send request/i });
    fireEvent.click(sendButton);

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('cannot add yourself as a contact');
    });
  });

  it('should cancel sent request workflow', async () => {
    mockUseSession.mockReturnValue(mockSession);

    const sentRequest = {
      id: 5,
      requesterUserId: 123,
      recipientUserId: 777,
      requesterName: 'Current User',
      recipientName: 'Pending User',
      status: 'pending' as const,
      requestedAt: '2024-01-01T00:00:00Z',
    };

    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [sentRequest],  // /api/contacts
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],  // /api/contact-rules
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Sent (1)')).toBeInTheDocument();
    });

    clickTab('Sent (1)');

    await waitFor(() => {
      expect(screen.getByText('Pending User')).toBeInTheDocument();
    });

    // Cancel the request (delete)
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      });

    const cancelButton = screen.getByTitle('Cancel Request');
    fireEvent.click(cancelButton);

    await waitFor(() => {
      expect(mockToastSuccess).toHaveBeenCalledWith('Contact removed');
    });

    await waitFor(() => {
      expect(screen.getByText('Sent (0)')).toBeInTheDocument();
    });
  });

  it('should show multiple contacts across all tabs', async () => {
    mockUseSession.mockReturnValue(mockSession);

    const contacts = [
      {
        id: 1,
        requesterUserId: 123,
        recipientUserId: 456,
        requesterName: 'Current User',
        recipientName: 'Friend 1',
        status: 'accepted' as const,
        requestedAt: '2024-01-01T00:00:00Z',
        respondedAt: '2024-01-02T00:00:00Z',
      },
      {
        id: 2,
        requesterUserId: 789,
        recipientUserId: 123,
        requesterName: 'Friend 2',
        recipientName: 'Current User',
        status: 'accepted' as const,
        requestedAt: '2024-01-03T00:00:00Z',
        respondedAt: '2024-01-04T00:00:00Z',
      },
      {
        id: 3,
        requesterUserId: 999,
        recipientUserId: 123,
        requesterName: 'Requester 1',
        recipientName: 'Current User',
        status: 'pending' as const,
        requestedAt: '2024-01-05T00:00:00Z',
      },
      {
        id: 4,
        requesterUserId: 123,
        recipientUserId: 888,
        requesterName: 'Current User',
        recipientName: 'Sent To 1',
        status: 'pending' as const,
        requestedAt: '2024-01-06T00:00:00Z',
      },
    ];

    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => contacts,  // /api/contacts
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],  // /api/contact-rules
      });

    render(<ContactsList />);

    // Check counts in tabs (correct tab labels)
    await waitFor(() => {
      expect(screen.getByText('My Contacts (2)')).toBeInTheDocument();
      expect(screen.getByText('Pending (1)')).toBeInTheDocument();
      expect(screen.getByText('Sent (1)')).toBeInTheDocument();
    });

    // Check My Contacts tab
    await waitFor(() => {
      expect(screen.getByText('Friend 1')).toBeInTheDocument();
      expect(screen.getByText('Friend 2')).toBeInTheDocument();
    });

    // Check Pending tab
    clickTab('Pending (1)');
    await waitFor(() => {
      expect(screen.getByText('Requester 1')).toBeInTheDocument();
    });

    // Check Sent tab
    clickTab('Sent (1)');
    await waitFor(() => {
      expect(screen.getByText('Sent To 1')).toBeInTheDocument();
    });
  });
});

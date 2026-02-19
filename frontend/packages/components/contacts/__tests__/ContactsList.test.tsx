import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import ContactsList from '../ContactsList';

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
  return function MockPermissionsDialog({ open, onClose }: any) {
    return open ? (
      <div data-testid="permissions-dialog">
        <button onClick={onClose}>Close</button>
      </div>
    ) : null;
  };
});

const mockUseSession = useSession as jest.MockedFunction<typeof useSession>;

describe('ContactsList Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (global.fetch as jest.Mock).mockClear();
  });

  const mockSession = {
    data: { providerAccountId: '123' } as any,
    status: 'authenticated' as const,
    update: jest.fn(),
  };

  const mockContacts = [
    {
      id: 1,
      requesterUserId: 123,
      recipientUserId: 456,
      requesterName: 'Current User',
      recipientName: 'Contact 1',
      status: 'accepted' as const,
      requestedAt: '2024-01-01T00:00:00Z',
      respondedAt: '2024-01-02T00:00:00Z',
    },
    {
      id: 2,
      requesterUserId: 789,
      recipientUserId: 123,
      requesterName: 'Contact 2',
      recipientName: 'Current User',
      status: 'pending' as const,
      requestedAt: '2024-01-03T00:00:00Z',
    },
    {
      id: 3,
      requesterUserId: 123,
      recipientUserId: 999,
      requesterName: 'Current User',
      recipientName: 'Contact 3',
      status: 'pending' as const,
      requestedAt: '2024-01-04T00:00:00Z',
    },
  ];

  it('should show loading state initially', () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    );

    render(<ContactsList />);
    expect(screen.getByTestId('loading')).toBeInTheDocument();
  });

  it('should not render when user is not authenticated', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated',
      update: jest.fn(),
    });

    const { container } = render(<ContactsList />);
    expect(container).toBeEmptyDOMElement();
  });

  it('should fetch and display contacts', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockContacts,
    });

    render(<ContactsList />);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith('/api/contacts');
    });

    await waitFor(() => {
      expect(screen.getByText('Contacts')).toBeInTheDocument();
    });
  });

  it('should display tabs with correct counts', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockContacts,
    });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('My Contacts (1)')).toBeInTheDocument();
      expect(screen.getByText('Pending Requests (1)')).toBeInTheDocument();
      expect(screen.getByText('Sent Requests (1)')).toBeInTheDocument();
    });
  });

  it('should display accepted contacts in My Contacts tab', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockContacts,
    });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Contact 1')).toBeInTheDocument();
      expect(screen.getByText('Connected')).toBeInTheDocument();
    });
  });

  it('should switch between tabs', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockContacts,
    });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Contact 1')).toBeInTheDocument();
    });

    // Switch to Pending Requests tab
    fireEvent.click(screen.getByText('Pending Requests (1)'));

    await waitFor(() => {
      expect(screen.getByText('Contact 2')).toBeInTheDocument();
    });

    // Switch to Sent Requests tab
    fireEvent.click(screen.getByText('Sent Requests (1)'));

    await waitFor(() => {
      expect(screen.getByText('Contact 3')).toBeInTheDocument();
    });
  });

  it('should open add contact dialog', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => [],
    });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Add Contact')).toBeInTheDocument();
    });

    const addButton = screen.getByRole('button', { name: /add contact/i });
    fireEvent.click(addButton);

    await waitFor(() => {
      expect(screen.getByLabelText(/character name/i)).toBeInTheDocument();
    });
  });

  it('should send contact request with character name', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ id: 4, status: 'pending' }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Add Contact')).toBeInTheDocument();
    });

    // Open add contact dialog
    const addButton = screen.getByRole('button', { name: /add contact/i });
    fireEvent.click(addButton);

    // Enter character name
    const input = screen.getByLabelText(/character name/i);
    fireEvent.change(input, { target: { value: 'Test Character' } });

    // Submit
    const sendButton = screen.getByRole('button', { name: /send request/i });
    fireEvent.click(sendButton);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/contacts',
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ characterName: 'Test Character' }),
        })
      );
    });
  });

  it('should accept contact request', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockContacts,
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ ...mockContacts[1], status: 'accepted' }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockContacts,
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Pending Requests (1)')).toBeInTheDocument();
    });

    // Switch to Pending Requests tab
    fireEvent.click(screen.getByText('Pending Requests (1)'));

    await waitFor(() => {
      expect(screen.getByText('Contact 2')).toBeInTheDocument();
    });

    // Accept the contact
    const acceptButton = screen.getByTitle('Accept');
    fireEvent.click(acceptButton);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/contacts/2/accept',
        expect.objectContaining({
          method: 'POST',
        })
      );
    });
  });

  it('should reject contact request', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockContacts,
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ ...mockContacts[1], status: 'rejected' }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockContacts,
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Pending Requests (1)')).toBeInTheDocument();
    });

    // Switch to Pending Requests tab
    fireEvent.click(screen.getByText('Pending Requests (1)'));

    await waitFor(() => {
      expect(screen.getByText('Contact 2')).toBeInTheDocument();
    });

    // Reject the contact
    const rejectButton = screen.getByTitle('Reject');
    fireEvent.click(rejectButton);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/contacts/2/reject',
        expect.objectContaining({
          method: 'POST',
        })
      );
    });
  });

  it('should delete contact', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockContacts,
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Contact 1')).toBeInTheDocument();
    });

    // Delete the contact
    const deleteButton = screen.getByTitle('Remove Contact');
    fireEvent.click(deleteButton);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/contacts/1',
        expect.objectContaining({
          method: 'DELETE',
        })
      );
    });
  });

  it('should open permissions dialog', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockContacts,
    });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Contact 1')).toBeInTheDocument();
    });

    // Open permissions dialog
    const settingsButton = screen.getByTitle('Manage Permissions');
    fireEvent.click(settingsButton);

    await waitFor(() => {
      expect(screen.getByTestId('permissions-dialog')).toBeInTheDocument();
    });
  });

  it('should display error message on failed request', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      })
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({ error: 'Character not found' }),
      });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Add Contact')).toBeInTheDocument();
    });

    // Open add contact dialog
    const addButton = screen.getByRole('button', { name: /add contact/i });
    fireEvent.click(addButton);

    // Enter character name
    const input = screen.getByLabelText(/character name/i);
    fireEvent.change(input, { target: { value: 'Invalid Character' } });

    // Submit
    const sendButton = screen.getByRole('button', { name: /send request/i });
    fireEvent.click(sendButton);

    await waitFor(() => {
      expect(screen.getByText('Character not found')).toBeInTheDocument();
    });
  });

  it('should display empty state when no contacts', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => [],
    });

    render(<ContactsList />);

    await waitFor(() => {
      expect(
        screen.getByText('No contacts yet. Add a contact to get started!')
      ).toBeInTheDocument();
    });
  });

  it('should display empty state for pending requests', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => [mockContacts[0]], // Only accepted contact
    });

    render(<ContactsList />);

    await waitFor(() => {
      expect(screen.getByText('Pending Requests (0)')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Pending Requests (0)'));

    await waitFor(() => {
      expect(screen.getByText('No pending requests')).toBeInTheDocument();
    });
  });

  it('should only fetch contacts once with ref guard', async () => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockContacts,
    });

    const { rerender } = render(<ContactsList />);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(2);
    });

    // Re-render with same session
    rerender(<ContactsList />);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(2);
    });
  });
});

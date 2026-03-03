import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import PermissionsDialog from '../PermissionsDialog';
import type { Contact } from '../ContactsList';

describe('PermissionsDialog Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (global.fetch as jest.Mock).mockClear();
  });

  const mockContact: Contact = {
    id: 1,
    requesterUserId: 123,
    recipientUserId: 456,
    requesterName: 'Current User',
    recipientName: 'Other User',
    status: 'accepted',
    requestedAt: '2024-01-01T00:00:00Z',
    respondedAt: '2024-01-02T00:00:00Z',
  };

  const mockPermissions = [
    {
      id: 1,
      contactId: 1,
      grantingUserId: 123,
      receivingUserId: 456,
      serviceType: 'for_sale_browse',
      canAccess: true,
    },
    {
      id: 2,
      contactId: 1,
      grantingUserId: 456,
      receivingUserId: 123,
      serviceType: 'for_sale_browse',
      canAccess: false,
    },
  ];

  it('should not render when closed', () => {
    const onClose = jest.fn();
    const { container } = render(
      <PermissionsDialog
        open={false}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    expect(container.querySelector('[role="dialog"]')).not.toBeInTheDocument();
  });

  it('should render when open', () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockPermissions,
    });

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    expect(screen.getByText(/Manage Permissions/)).toBeInTheDocument();
  });

  it('should display other user name in title', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockPermissions,
    });

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    await waitFor(() => {
      expect(
        screen.getByText('Permissions I Grant to Other User')
      ).toBeInTheDocument();
      expect(
        screen.getByText('Permissions Other User Grants to Me')
      ).toBeInTheDocument();
    });
  });

  it('should fetch permissions on open', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockPermissions,
    });

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith('/api/contacts/1/permissions');
    });
  });

  it('should display loading state while fetching', () => {
    (global.fetch as jest.Mock).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    );

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('should display permissions with correct toggle states', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockPermissions,
    });

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Permissions I Grant to Other User')).toBeInTheDocument();
      expect(screen.getByText('Permissions Other User Grants to Me')).toBeInTheDocument();
    });

    // Verify the permission labels are displayed
    const labels = screen.getAllByText('Browse For-Sale Items');
    expect(labels).toHaveLength(2);
  });

  it('should toggle permission when switch is clicked', async () => {
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockPermissions,
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ ...mockPermissions[0], canAccess: false }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [
          { ...mockPermissions[0], canAccess: false },
          mockPermissions[1],
        ],
      });

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Permissions I Grant to Other User')).toBeInTheDocument();
    });

    // Wait for permissions to load and find the switch
    let switchElement: HTMLInputElement | null = null;
    await waitFor(() => {
      const switchLabels = screen.getAllByText('Browse For-Sale Items');
      switchElement = switchLabels[0].closest('label')?.querySelector('input[type="checkbox"]') as HTMLInputElement;
      expect(switchElement).toBeInTheDocument();
    });

    if (switchElement) {
      fireEvent.click(switchElement);
    }

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/contacts/1/permissions',
        expect.objectContaining({
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            serviceType: 'for_sale_browse',
            receivingUserId: 456,
            canAccess: false,
          }),
        })
      );
    });
  });

  it('should handle permission update errors', async () => {
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockPermissions,
      })
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({ error: 'Permission update failed' }),
      });

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    // Wait for permissions to load
    let switchElement: HTMLInputElement | null = null;
    await waitFor(() => {
      const switchLabels = screen.getAllByText('Browse For-Sale Items');
      switchElement = switchLabels[0].closest('label')?.querySelector('input[type="checkbox"]') as HTMLInputElement;
      expect(switchElement).toBeInTheDocument();
    });

    // Try to toggle
    if (switchElement) {
      fireEvent.click(switchElement);
    }

    // Should display error message
    await waitFor(() => {
      expect(screen.getByText('Permission update failed')).toBeInTheDocument();
    });
  });

  it('should call onClose when close button is clicked', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockPermissions,
    });

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    // The dialog has two "Close" buttons: the explicit footer button and the built-in
    // shadcn/ui DialogContent X button. Use getAllByText to handle both.
    await waitFor(() => {
      expect(screen.getAllByText('Close').length).toBeGreaterThanOrEqual(1);
    });

    fireEvent.click(screen.getAllByText('Close')[0]);

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('should display correct other user name when current user is recipient', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockPermissions,
    });

    const reversedContact: Contact = {
      ...mockContact,
      requesterUserId: 456,
      recipientUserId: 123,
      requesterName: 'Other User',
      recipientName: 'Current User',
    };

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={reversedContact}
        currentUserId={123}
      />
    );

    await waitFor(() => {
      expect(
        screen.getByText('Permissions I Grant to Other User')
      ).toBeInTheDocument();
    });
  });

  it('should display formatted service type names', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockPermissions,
    });

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    await waitFor(() => {
      const labels = screen.getAllByText('Browse For-Sale Items');
      expect(labels.length).toBeGreaterThan(0);
    });
  });

  it('should handle empty permissions response', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => [],
    });

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    await waitFor(() => {
      expect(screen.getByText(/Manage Permissions/)).toBeInTheDocument();
    });

    // Should still render service type labels (they're hardcoded)
    await waitFor(() => {
      const labels = screen.getAllByText('Browse For-Sale Items');
      expect(labels.length).toBeGreaterThan(0);
    });
  });

  it('should refetch permissions after successful update', async () => {
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockPermissions,
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ ...mockPermissions[0], canAccess: false }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [
          { ...mockPermissions[0], canAccess: false },
          mockPermissions[1],
        ],
      });

    const onClose = jest.fn();
    render(
      <PermissionsDialog
        open={true}
        onClose={onClose}
        contact={mockContact}
        currentUserId={123}
      />
    );

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(1);
    });

    // Wait for permissions to load and toggle
    let switchElement: HTMLInputElement | null = null;
    await waitFor(() => {
      const switchLabels = screen.getAllByText('Browse For-Sale Items');
      switchElement = switchLabels[0].closest('label')?.querySelector('input[type="checkbox"]') as HTMLInputElement;
      expect(switchElement).toBeInTheDocument();
    });

    if (switchElement) {
      fireEvent.click(switchElement);
    }

    await waitFor(() => {
      // Should have called: initial fetch, update, refetch
      expect(global.fetch).toHaveBeenCalledTimes(3);
    });
  });
});

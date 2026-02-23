import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import PlanRunsList from '../PlanRunsList';
import { PlanRun } from '@industry-tool/client/data/models';

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

const mockUseSession = useSession as jest.MockedFunction<typeof useSession>;

const mockSession = {
  data: {
    providerAccountId: '1001',
    expires: '2099-01-01',
  },
  status: 'authenticated' as const,
  update: jest.fn(),
};

describe('PlanRunsList Component', () => {
  beforeEach(() => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockClear();
  });

  it('should match snapshot while loading', async () => {
    (global.fetch as jest.Mock).mockImplementation(
      () => new Promise(() => {}),
    );

    const { container } = render(<PlanRunsList />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with empty runs', async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    let container: HTMLElement;
    await act(async () => {
      const result = render(<PlanRunsList />);
      container = result.container;
    });

    expect(container!).toMatchSnapshot();
  });

  it('should display empty message when no runs', async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    await act(async () => {
      render(<PlanRunsList />);
    });

    expect(screen.getByText(/No plan runs yet/)).toBeInTheDocument();
  });

  it('should match snapshot with runs', async () => {
    const runs: PlanRun[] = [
      {
        id: 1,
        planId: 5,
        userId: 100,
        quantity: 10,
        createdAt: '2026-02-22T23:00:00Z',
        planName: 'Rifter Plan',
        productName: 'Rifter',
        status: 'in_progress',
        jobSummary: {
          total: 5,
          planned: 2,
          active: 2,
          completed: 1,
          cancelled: 0,
        },
      },
      {
        id: 2,
        planId: 8,
        userId: 100,
        quantity: 3,
        createdAt: '2026-02-22T20:00:00Z',
        planName: 'Slasher Plan',
        productName: 'Slasher',
        status: 'completed',
        jobSummary: {
          total: 3,
          planned: 0,
          active: 0,
          completed: 3,
          cancelled: 0,
        },
      },
    ];

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => runs,
    });

    let container: HTMLElement;
    await act(async () => {
      const result = render(<PlanRunsList />);
      container = result.container;
    });

    expect(container!).toMatchSnapshot();
  });

  it('should display run details in table', async () => {
    const runs: PlanRun[] = [
      {
        id: 1,
        planId: 5,
        userId: 100,
        quantity: 10,
        createdAt: '2026-02-22T23:00:00Z',
        planName: 'Rifter Plan',
        productName: 'Rifter',
        status: 'in_progress',
        jobSummary: {
          total: 5,
          planned: 2,
          active: 2,
          completed: 1,
          cancelled: 0,
        },
      },
    ];

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => runs,
    });

    await act(async () => {
      render(<PlanRunsList />);
    });

    expect(screen.getByText('Rifter')).toBeInTheDocument();
    expect(screen.getByText('Rifter Plan')).toBeInTheDocument();
    expect(screen.getByText('10')).toBeInTheDocument();
    expect(screen.getByText('In Progress')).toBeInTheDocument();
    expect(screen.getByText('1/5 done, 2 active')).toBeInTheDocument();
  });

  it('should show cancel button for runs with planned jobs', async () => {
    const runs: PlanRun[] = [
      {
        id: 1,
        planId: 5,
        userId: 100,
        quantity: 10,
        createdAt: '2026-02-22T23:00:00Z',
        planName: 'Rifter Plan',
        productName: 'Rifter',
        status: 'in_progress',
        jobSummary: {
          total: 5,
          planned: 2,
          active: 2,
          completed: 1,
          cancelled: 0,
        },
      },
    ];

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => runs,
    });

    await act(async () => {
      render(<PlanRunsList />);
    });

    const cancelButton = screen.getByTitle('Cancel planned jobs');
    expect(cancelButton).toBeInTheDocument();
  });

  it('should not show cancel button for completed runs', async () => {
    const runs: PlanRun[] = [
      {
        id: 2,
        planId: 8,
        userId: 100,
        quantity: 3,
        createdAt: '2026-02-22T20:00:00Z',
        planName: 'Slasher Plan',
        productName: 'Slasher',
        status: 'completed',
        jobSummary: {
          total: 3,
          planned: 0,
          active: 0,
          completed: 3,
          cancelled: 0,
        },
      },
    ];

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => runs,
    });

    await act(async () => {
      render(<PlanRunsList />);
    });

    expect(screen.queryByTitle('Cancel planned jobs')).not.toBeInTheDocument();
  });

  it('should call cancel API when cancel is confirmed', async () => {
    const runs: PlanRun[] = [
      {
        id: 1,
        planId: 5,
        userId: 100,
        quantity: 10,
        createdAt: '2026-02-22T23:00:00Z',
        planName: 'Rifter Plan',
        productName: 'Rifter',
        status: 'pending',
        jobSummary: {
          total: 3,
          planned: 3,
          active: 0,
          completed: 0,
          cancelled: 0,
        },
      },
    ];

    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({ ok: true, json: async () => runs })
      .mockResolvedValueOnce({ ok: true, json: async () => ({ status: 'cancelled', jobsCancelled: 3 }) })
      .mockResolvedValueOnce({ ok: true, json: async () => [] });

    window.confirm = jest.fn(() => true);

    await act(async () => {
      render(<PlanRunsList />);
    });

    const cancelButton = screen.getByTitle('Cancel planned jobs');

    await act(async () => {
      fireEvent.click(cancelButton);
    });

    expect(window.confirm).toHaveBeenCalledWith(
      'Are you sure you want to cancel all planned jobs for this run?',
    );
    expect(global.fetch).toHaveBeenCalledWith(
      '/api/industry/plans/runs/1/cancel',
      { method: 'POST' },
    );
  });
});

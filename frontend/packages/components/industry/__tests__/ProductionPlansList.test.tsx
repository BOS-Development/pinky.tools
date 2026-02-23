import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import ProductionPlansList from '../ProductionPlansList';
import { ProductionPlan } from '@industry-tool/client/data/models';

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
jest.mock('../ProductionPlanEditor', () => {
  return function MockEditor({ planId }: { planId: number }) {
    return <div data-testid="plan-editor">Editor for plan {planId}</div>;
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

describe('ProductionPlansList Component', () => {
  beforeEach(() => {
    mockUseSession.mockReturnValue(mockSession);
    (global.fetch as jest.Mock).mockClear();
  });

  it('should match snapshot while loading', async () => {
    (global.fetch as jest.Mock).mockImplementation(
      () => new Promise(() => {}), // never resolves
    );

    const { container } = render(<ProductionPlansList />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with empty plans', async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    let container: HTMLElement;
    await act(async () => {
      const result = render(<ProductionPlansList />);
      container = result.container;
    });

    expect(container!).toMatchSnapshot();
  });

  it('should display empty message when no plans', async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    await act(async () => {
      render(<ProductionPlansList />);
    });

    expect(
      screen.getByText(/No production plans yet/),
    ).toBeInTheDocument();
  });

  it('should match snapshot with plans', async () => {
    const plans: ProductionPlan[] = [
      {
        id: 1,
        userId: 100,
        productTypeId: 587,
        name: 'Rifter Production',
        createdAt: '2026-02-22T12:00:00Z',
        updatedAt: '2026-02-22T12:00:00Z',
        productName: 'Rifter',
        steps: [
          {
            id: 10,
            planId: 1,
            productTypeId: 587,
            blueprintTypeId: 787,
            activity: 'manufacturing',
            meLevel: 10,
            teLevel: 20,
            industrySkill: 5,
            advIndustrySkill: 5,
            structure: 'raitaru',
            rig: 't2',
            security: 'high',
            facilityTax: 1.0,
            productName: 'Rifter',
            blueprintName: 'Rifter Blueprint',
          },
        ],
      },
      {
        id: 2,
        userId: 100,
        productTypeId: 11379,
        name: 'Muninn Production',
        createdAt: '2026-02-22T12:00:00Z',
        updatedAt: '2026-02-22T12:00:00Z',
        productName: 'Muninn',
        steps: [
          {
            id: 20,
            planId: 2,
            productTypeId: 11379,
            blueprintTypeId: 11380,
            activity: 'manufacturing',
            meLevel: 10,
            teLevel: 20,
            industrySkill: 5,
            advIndustrySkill: 5,
            structure: 'raitaru',
            rig: 't2',
            security: 'low',
            facilityTax: 1.0,
            productName: 'Muninn',
            blueprintName: 'Muninn Blueprint',
          },
          {
            id: 21,
            planId: 2,
            parentStepId: 20,
            productTypeId: 11370,
            blueprintTypeId: 11371,
            activity: 'manufacturing',
            meLevel: 10,
            teLevel: 20,
            industrySkill: 5,
            advIndustrySkill: 5,
            structure: 'raitaru',
            rig: 't2',
            security: 'low',
            facilityTax: 1.0,
            productName: 'Rupture',
            blueprintName: 'Rupture Blueprint',
          },
        ],
      },
    ];

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => plans,
    });

    let container: HTMLElement;
    await act(async () => {
      const result = render(<ProductionPlansList />);
      container = result.container;
    });

    expect(container!).toMatchSnapshot();
  });

  it('should display plan details in table', async () => {
    const plans: ProductionPlan[] = [
      {
        id: 1,
        userId: 100,
        productTypeId: 587,
        name: 'Rifter Production',
        createdAt: '2026-02-22T12:00:00Z',
        updatedAt: '2026-02-22T12:00:00Z',
        productName: 'Rifter',
        steps: [
          {
            id: 10,
            planId: 1,
            productTypeId: 587,
            blueprintTypeId: 787,
            activity: 'manufacturing',
            meLevel: 10,
            teLevel: 20,
            industrySkill: 5,
            advIndustrySkill: 5,
            structure: 'raitaru',
            rig: 't2',
            security: 'high',
            facilityTax: 1.0,
          },
        ],
      },
    ];

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => plans,
    });

    await act(async () => {
      render(<ProductionPlansList />);
    });

    expect(screen.getByText('Rifter')).toBeInTheDocument();
    expect(screen.getByText('Rifter Production')).toBeInTheDocument();
  });

  it('should navigate to editor when plan row is clicked', async () => {
    const plans: ProductionPlan[] = [
      {
        id: 42,
        userId: 100,
        productTypeId: 587,
        name: 'Rifter Production',
        createdAt: '2026-02-22T12:00:00Z',
        updatedAt: '2026-02-22T12:00:00Z',
        productName: 'Rifter',
        steps: [],
      },
    ];

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => plans,
    });

    await act(async () => {
      render(<ProductionPlansList />);
    });

    // Click the plan row
    fireEvent.click(screen.getByText('Rifter'));

    expect(screen.getByTestId('plan-editor')).toBeInTheDocument();
    expect(screen.getByText('Editor for plan 42')).toBeInTheDocument();
  });

  it('should show back button when viewing editor', async () => {
    const plans: ProductionPlan[] = [
      {
        id: 42,
        userId: 100,
        productTypeId: 587,
        name: 'Rifter Production',
        createdAt: '2026-02-22T12:00:00Z',
        updatedAt: '2026-02-22T12:00:00Z',
        productName: 'Rifter',
        steps: [],
      },
    ];

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => plans,
    });

    await act(async () => {
      render(<ProductionPlansList />);
    });

    fireEvent.click(screen.getByText('Rifter'));
    expect(screen.getByText('Back to Plans')).toBeInTheDocument();
  });

  it('should call delete API when delete button is clicked', async () => {
    const plans: ProductionPlan[] = [
      {
        id: 5,
        userId: 100,
        productTypeId: 587,
        name: 'Rifter Production',
        createdAt: '2026-02-22T12:00:00Z',
        updatedAt: '2026-02-22T12:00:00Z',
        productName: 'Rifter',
        steps: [],
      },
    ];

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => plans,
    });

    await act(async () => {
      render(<ProductionPlansList />);
    });

    // Mock for delete + refetch
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce({ ok: true })
      .mockResolvedValueOnce({ ok: true, json: async () => [] });

    // Find the delete button (second icon button in actions cell)
    const deleteButtons = screen.getAllByTestId('DeleteIcon');
    await act(async () => {
      fireEvent.click(deleteButtons[0].closest('button')!);
    });

    expect(global.fetch).toHaveBeenCalledWith('/api/industry/plans/5', {
      method: 'DELETE',
    });
  });

  it('should not fetch plans when not authenticated', () => {
    mockUseSession.mockReturnValue({
      data: null,
      status: 'unauthenticated',
      update: jest.fn(),
    });

    render(<ProductionPlansList />);
    expect(global.fetch).not.toHaveBeenCalled();
  });

  it('should open create dialog when New Plan button is clicked', async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    await act(async () => {
      render(<ProductionPlansList />);
    });

    // Mock the stations fetch that fires when dialog opens
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    await act(async () => {
      fireEvent.click(screen.getByText('New Plan'));
    });

    expect(screen.getByText('Create Production Plan')).toBeInTheDocument();
    expect(screen.getByLabelText('Default Manufacturing Station')).toBeInTheDocument();
    expect(screen.getByLabelText('Default Reaction Station')).toBeInTheDocument();
  });
});

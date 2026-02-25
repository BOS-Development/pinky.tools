import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import ProductionPlanEditor from '../ProductionPlanEditor';
import { ProductionPlan, ProductionPlanStep, PlanMaterial } from '@industry-tool/client/data/models';

// Mock formatISK to avoid import issues
jest.mock('@industry-tool/utils/formatting', () => ({
  formatISK: (val: number) => `${val.toLocaleString()} ISK`,
  formatNumber: (val: number) => val.toLocaleString(),
  formatCompact: (val: number) => val.toLocaleString(),
}));

const mockRootStep: ProductionPlanStep = {
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
};

const mockPlan: ProductionPlan = {
  id: 1,
  userId: 100,
  productTypeId: 587,
  name: 'Rifter Production',
  createdAt: '2026-02-22T00:00:00Z',
  updatedAt: '2026-02-22T00:00:00Z',
  productName: 'Rifter',
  steps: [mockRootStep],
};

const mockMaterials: PlanMaterial[] = [
  {
    typeId: 34,
    typeName: 'Tritanium',
    quantity: 22222,
    volume: 0.01,
    hasBlueprint: false,
    isProduced: false,
  },
  {
    typeId: 35,
    typeName: 'Pyerite',
    quantity: 5555,
    volume: 0.01,
    hasBlueprint: false,
    isProduced: false,
  },
  {
    typeId: 11399,
    typeName: 'Morphite',
    quantity: 10,
    volume: 0.01,
    hasBlueprint: true,
    blueprintTypeId: 99999,
    activity: 'reaction',
    isProduced: false,
  },
];

// URL-based fetch mock to handle concurrent requests from multiple useEffects
function mockFetchForPlan(plan: ProductionPlan | null, materials?: PlanMaterial[]) {
  (global.fetch as jest.Mock).mockImplementation((url: string, opts?: any) => {
    if (url === `/api/industry/plans/${plan?.id ?? 1}`) {
      return Promise.resolve({ ok: true, json: async () => plan });
    }
    if (url === '/api/transport/profiles') {
      return Promise.resolve({ ok: true, json: async () => [] });
    }
    if (url.includes('/materials') && materials) {
      return Promise.resolve({ ok: true, json: async () => materials });
    }
    return Promise.resolve({ ok: true, json: async () => [] });
  });
}

describe('ProductionPlanEditor Component', () => {
  beforeEach(() => {
    (global.fetch as jest.Mock).mockClear();
  });

  it('should match snapshot while loading', () => {
    (global.fetch as jest.Mock).mockImplementation(
      () => new Promise(() => {}), // never resolves
    );

    const { container } = render(<ProductionPlanEditor planId={1} />);
    expect(container).toMatchSnapshot();
  });

  it('should display loading text initially', () => {
    (global.fetch as jest.Mock).mockImplementation(
      () => new Promise(() => {}),
    );

    render(<ProductionPlanEditor planId={1} />);
    expect(screen.getByText('Loading plan...')).toBeInTheDocument();
  });

  it('should match snapshot with loaded plan', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    let container: HTMLElement;
    await act(async () => {
      const result = render(<ProductionPlanEditor planId={1} />);
      container = result.container;
    });

    expect(container!).toMatchSnapshot();
  });

  it('should display plan name and product info', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    expect(screen.getByText('Rifter Production')).toBeInTheDocument();
    expect(screen.getByText('1 production step(s)')).toBeInTheDocument();
  });

  it('should display root step with details', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    expect(screen.getByText('Rifter')).toBeInTheDocument();
    expect(screen.getByText('manufacturing')).toBeInTheDocument();
    expect(screen.getByText('ME 10 / TE 20')).toBeInTheDocument();
    expect(screen.getByText('raitaru / t2 / high')).toBeInTheDocument();
  });

  it('should display materials when root step is auto-expanded', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    expect(screen.getByText('Tritanium')).toBeInTheDocument();
    expect(screen.getByText('Pyerite')).toBeInTheDocument();
    expect(screen.getByText('Morphite')).toBeInTheDocument();
  });

  it('should show produce toggle for materials with blueprints', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    // Morphite has hasBlueprint=true, so it should have a toggle button
    expect(screen.getByTitle('Switch to Produce')).toBeInTheDocument();
  });

  it('should show Buy chip for non-produced materials', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    const buyChips = screen.getAllByText('Buy');
    expect(buyChips.length).toBe(3);
  });

  it('should have Generate Jobs button', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    expect(screen.getByText('Generate Jobs')).toBeInTheDocument();
  });

  it('should open generate dialog when Generate Jobs is clicked', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    fireEvent.click(screen.getByText('Generate Jobs'));
    expect(screen.getByText('Generate Production Jobs')).toBeInTheDocument();
    expect(screen.getByText(/How many Rifter do you want to produce/)).toBeInTheDocument();
  });

  it('should open edit dialog when edit button is clicked', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    const editButtons = screen.getAllByTestId('EditIcon');
    // First EditIcon is plan name edit, second is the step edit button
    fireEvent.click(editButtons[1].closest('button')!);

    expect(screen.getByText('Edit Step: Rifter')).toBeInTheDocument();
  });

  it('should display "No steps" message for plan without steps', async () => {
    const emptyPlan: ProductionPlan = {
      ...mockPlan,
      steps: [],
    };

    mockFetchForPlan(emptyPlan);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    expect(screen.getByText('No steps in this plan')).toBeInTheDocument();
  });

  it('should match snapshot with multi-step plan', async () => {
    const childStep: ProductionPlanStep = {
      id: 11,
      planId: 1,
      parentStepId: 10,
      productTypeId: 11399,
      blueprintTypeId: 99999,
      activity: 'reaction',
      meLevel: 0,
      teLevel: 0,
      industrySkill: 5,
      advIndustrySkill: 5,
      structure: 'raitaru',
      rig: 't2',
      security: 'low',
      facilityTax: 0.25,
      productName: 'Morphite',
      blueprintName: 'Morphite Reaction Formula',
    };

    const multiStepPlan: ProductionPlan = {
      ...mockPlan,
      steps: [mockRootStep, childStep],
    };

    const materialsWithProduced: PlanMaterial[] = [
      ...mockMaterials.slice(0, 2),
      { ...mockMaterials[2], isProduced: true },
    ];

    mockFetchForPlan(multiStepPlan, materialsWithProduced);

    let container: HTMLElement;
    await act(async () => {
      const result = render(<ProductionPlanEditor planId={1} />);
      container = result.container;
    });

    expect(container!).toMatchSnapshot();
  });

  it('should fetch plan on mount', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    expect(global.fetch).toHaveBeenCalledWith('/api/industry/plans/1');
  });

  it('should fetch transport profiles on mount', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    expect(global.fetch).toHaveBeenCalledWith('/api/transport/profiles');
  });

  it('should fetch materials for root step on load', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    expect(global.fetch).toHaveBeenCalledWith(
      '/api/industry/plans/1/steps/10/materials',
    );
  });

  it('should show Transport tab', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    expect(screen.getByText('Transport')).toBeInTheDocument();
  });
});

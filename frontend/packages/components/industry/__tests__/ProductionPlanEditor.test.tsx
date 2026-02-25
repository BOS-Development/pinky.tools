import React from 'react';
import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import ProductionPlanEditor from '../ProductionPlanEditor';
import { ProductionPlan, ProductionPlanStep, PlanMaterial, PlanPreviewResult } from '@industry-tool/client/data/models';

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

const mockPreviewResult: PlanPreviewResult = {
  eligibleCharacters: 2,
  totalJobs: 8,
  options: [
    {
      parallelism: 1,
      estimatedDurationSec: 172800,
      estimatedDurationLabel: '2d 0h',
      characters: [
        {
          characterId: 101,
          name: 'Main',
          jobCount: 8,
          durationSec: 172800,
          mfgSlotsUsed: 5,
          mfgSlotsMax: 11,
          reactSlotsUsed: 0,
          reactSlotsMax: 0,
        },
      ],
    },
    {
      parallelism: 2,
      estimatedDurationSec: 93600,
      estimatedDurationLabel: '1d 2h',
      characters: [
        {
          characterId: 101,
          name: 'Main',
          jobCount: 4,
          durationSec: 93600,
          mfgSlotsUsed: 3,
          mfgSlotsMax: 11,
          reactSlotsUsed: 0,
          reactSlotsMax: 0,
        },
        {
          characterId: 102,
          name: 'Alt 1',
          jobCount: 4,
          durationSec: 93600,
          mfgSlotsUsed: 2,
          mfgSlotsMax: 11,
          reactSlotsUsed: 0,
          reactSlotsMax: 0,
        },
      ],
    },
  ],
};

const mockBlueprintLevel = {
  materialEfficiency: 8,
  timeEfficiency: 16,
  isCopy: false,
  ownerName: 'Test Character',
  runs: -1,
};

// URL-based fetch mock to handle concurrent requests from multiple useEffects
function mockFetchForPlan(
  plan: ProductionPlan | null,
  materials?: PlanMaterial[],
  preview?: PlanPreviewResult,
  blueprintLevels?: Record<string, typeof mockBlueprintLevel | null>,
) {
  (global.fetch as jest.Mock).mockImplementation((url: string, opts?: any) => {
    if (url === `/api/industry/plans/${plan?.id ?? 1}`) {
      return Promise.resolve({ ok: true, json: async () => plan });
    }
    if (url === '/api/transport/profiles') {
      return Promise.resolve({ ok: true, json: async () => [] });
    }
    if (url === '/api/industry/blueprint-levels') {
      return Promise.resolve({ ok: true, json: async () => blueprintLevels ?? {} });
    }
    if (url.includes('/materials') && materials) {
      return Promise.resolve({ ok: true, json: async () => materials });
    }
    if (url.includes('/preview') && preview) {
      return Promise.resolve({ ok: true, json: async () => preview });
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

  it('should show Preview button in generate dialog', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    fireEvent.click(screen.getByText('Generate Jobs'));

    expect(screen.getByText('Preview')).toBeInTheDocument();
  });

  it('should call preview API and show parallelism options', async () => {
    mockFetchForPlan(mockPlan, mockMaterials, mockPreviewResult);

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    fireEvent.click(screen.getByText('Generate Jobs'));

    await act(async () => {
      fireEvent.click(screen.getByText('Preview'));
    });

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/industry/plans/1/preview',
        expect.objectContaining({ method: 'POST' }),
      );
    });

    expect(screen.getByText('No assignment')).toBeInTheDocument();
    expect(screen.getByText('2d 0h')).toBeInTheDocument();
    expect(screen.getByText('1d 2h')).toBeInTheDocument();
  });

  it('should show preview error if preview fails', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);
    (global.fetch as jest.Mock).mockImplementation((url: string, opts?: any) => {
      if (url === `/api/industry/plans/1`) {
        return Promise.resolve({ ok: true, json: async () => mockPlan });
      }
      if (url === '/api/transport/profiles') {
        return Promise.resolve({ ok: true, json: async () => [] });
      }
      if (url.includes('/materials')) {
        return Promise.resolve({ ok: true, json: async () => mockMaterials });
      }
      if (url.includes('/preview')) {
        return Promise.resolve({ ok: false, status: 400, json: async () => ({ error: 'No eligible characters' }) });
      }
      return Promise.resolve({ ok: true, json: async () => [] });
    });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    fireEvent.click(screen.getByText('Generate Jobs'));

    await act(async () => {
      fireEvent.click(screen.getByText('Preview'));
    });

    await waitFor(() => {
      expect(screen.getByText(/No eligible characters/)).toBeInTheDocument();
    });
  });

  it('should include parallelism in generate request when option is selected', async () => {
    mockFetchForPlan(mockPlan, mockMaterials, mockPreviewResult);
    // Also mock generate endpoint
    (global.fetch as jest.Mock).mockImplementation((url: string, opts?: any) => {
      if (url === `/api/industry/plans/1`) {
        return Promise.resolve({ ok: true, json: async () => mockPlan });
      }
      if (url === '/api/transport/profiles') {
        return Promise.resolve({ ok: true, json: async () => [] });
      }
      if (url.includes('/materials')) {
        return Promise.resolve({ ok: true, json: async () => mockMaterials });
      }
      if (url.includes('/preview')) {
        return Promise.resolve({ ok: true, json: async () => mockPreviewResult });
      }
      if (url.includes('/generate')) {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            run: { id: 1, planId: 1, userId: 100, quantity: 1, createdAt: '', status: 'planned' },
            created: [],
            skipped: [],
            transportJobs: [],
          }),
        });
      }
      return Promise.resolve({ ok: true, json: async () => [] });
    });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    // Open dialog via toolbar button
    const generateButtons = screen.getAllByText('Generate Jobs');
    fireEvent.click(generateButtons[0]);

    await act(async () => {
      fireEvent.click(screen.getByText('Preview'));
    });

    await waitFor(() => {
      expect(screen.getByText('No assignment')).toBeInTheDocument();
    });

    // Select parallelism = 2
    const radioButtons = screen.getAllByRole('radio');
    await act(async () => {
      // radioButtons[0] = no assignment (value=0), radioButtons[1] = parallelism 1, radioButtons[2] = parallelism 2
      fireEvent.click(radioButtons[2]);
    });

    // Click dialog's Generate Jobs button (last one)
    const generateJobsButtons = screen.getAllByText('Generate Jobs');
    await act(async () => {
      fireEvent.click(generateJobsButtons[generateJobsButtons.length - 1]);
    });

    await waitFor(() => {
      const generateCall = (global.fetch as jest.Mock).mock.calls.find(
        ([url]: [string]) => url.includes('/generate'),
      );
      expect(generateCall).toBeDefined();
      const body = JSON.parse(generateCall[1].body);
      expect(body.parallelism).toBe(2);
    });
  });

  it('should fetch blueprint levels on plan load', async () => {
    mockFetchForPlan(mockPlan, mockMaterials, undefined, { '787': mockBlueprintLevel });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      const blueprintLevelCall = (global.fetch as jest.Mock).mock.calls.find(
        ([url]: [string]) => url === '/api/industry/blueprint-levels',
      );
      expect(blueprintLevelCall).toBeDefined();
    });
  });

  it('should call blueprint-levels with all step blueprint type IDs', async () => {
    mockFetchForPlan(mockPlan, mockMaterials, undefined, { '787': mockBlueprintLevel });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      const blueprintLevelCall = (global.fetch as jest.Mock).mock.calls.find(
        ([url]: [string]) => url === '/api/industry/blueprint-levels',
      );
      expect(blueprintLevelCall).toBeDefined();
      const body = JSON.parse(blueprintLevelCall[1].body);
      expect(body.type_ids).toContain(787);
    });
  });

  it('should show info icon in step tree when detected ME/TE differs from step', async () => {
    // Step has ME 10/TE 20, but detected level is ME 8/TE 16
    mockFetchForPlan(mockPlan, mockMaterials, undefined, { '787': mockBlueprintLevel });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      // The info icon should be present since ME 8 != 10 or TE 16 != 20
      const infoIcons = screen.queryAllByTestId('InfoIcon');
      expect(infoIcons.length).toBeGreaterThanOrEqual(1);
    });
  });

  it('should show green check icon and no info icon when detected ME/TE matches step values', async () => {
    // Step has ME 10/TE 20, detected level also ME 10/TE 20 - values match
    const matchingLevel = { materialEfficiency: 10, timeEfficiency: 20, isCopy: false, ownerName: 'Test', runs: -1 };
    mockFetchForPlan(mockPlan, mockMaterials, undefined, { '787': matchingLevel });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      const blueprintLevelCall = (global.fetch as jest.Mock).mock.calls.find(
        ([url]: [string]) => url === '/api/industry/blueprint-levels',
      );
      expect(blueprintLevelCall).toBeDefined();
    });

    // No info icon since values match
    const infoIcons = screen.queryAllByTestId('InfoIcon');
    expect(infoIcons.length).toBe(0);

    // Green check icon should appear since blueprint was detected and ME/TE matches
    await waitFor(() => {
      const checkIcons = screen.queryAllByTestId('CheckCircleOutlineIcon');
      expect(checkIcons.length).toBeGreaterThanOrEqual(1);
    });
  });

  it('should show green check icon for reaction step even when detected ME/TE differs from step values', async () => {
    // Reaction step has ME 0/TE 0 (always fixed), but detected level has ME 5/TE 10
    // For reactions, mismatch should be ignored — green check should show, no info icon
    const reactionStep: ProductionPlanStep = {
      id: 20,
      planId: 1,
      productTypeId: 16671,
      blueprintTypeId: 46166,
      activity: 'reaction',
      meLevel: 0,
      teLevel: 0,
      industrySkill: 5,
      advIndustrySkill: 5,
      structure: 'tatara',
      rig: 't1',
      security: 'low',
      facilityTax: 0.25,
      productName: 'Fullerite-C32',
      blueprintName: 'Fullerite-C32 Reaction Formula',
    };

    const reactionPlan: ProductionPlan = {
      ...mockPlan,
      steps: [reactionStep],
    };

    const detectedReactionLevel = {
      materialEfficiency: 5,
      timeEfficiency: 10,
      isCopy: false,
      ownerName: 'Reaction Character',
      runs: -1,
    };

    mockFetchForPlan(reactionPlan, [], undefined, { '46166': detectedReactionLevel });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      const blueprintLevelCall = (global.fetch as jest.Mock).mock.calls.find(
        ([url]: [string]) => url === '/api/industry/blueprint-levels',
      );
      expect(blueprintLevelCall).toBeDefined();
    });

    // Info icon should NOT appear for reaction steps (even though ME/TE differs)
    const infoIcons = screen.queryAllByTestId('InfoIcon');
    expect(infoIcons.length).toBe(0);

    // Green check icon should appear since blueprint was detected (reaction always shows green)
    await waitFor(() => {
      const checkIcons = screen.queryAllByTestId('CheckCircleOutlineIcon');
      expect(checkIcons.length).toBeGreaterThanOrEqual(1);
    });
  });

  it('should show no Apply button in edit dialog for reaction steps when blueprint detected', async () => {
    // Reaction steps don't have ME/TE research, so Apply button should not appear
    const reactionStep: ProductionPlanStep = {
      id: 20,
      planId: 1,
      productTypeId: 16671,
      blueprintTypeId: 46166,
      activity: 'reaction',
      meLevel: 0,
      teLevel: 0,
      industrySkill: 5,
      advIndustrySkill: 5,
      structure: 'tatara',
      rig: 't1',
      security: 'low',
      facilityTax: 0.25,
      productName: 'Fullerite-C32',
      blueprintName: 'Fullerite-C32 Reaction Formula',
    };

    const reactionPlan: ProductionPlan = {
      ...mockPlan,
      steps: [reactionStep],
    };

    const detectedReactionLevel = {
      materialEfficiency: 0,
      timeEfficiency: 0,
      isCopy: false,
      ownerName: 'Reaction Character',
      runs: -1,
    };

    mockFetchForPlan(reactionPlan, [], undefined, { '46166': detectedReactionLevel });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      const blueprintLevelCall = (global.fetch as jest.Mock).mock.calls.find(
        ([url]: [string]) => url === '/api/industry/blueprint-levels',
      );
      expect(blueprintLevelCall).toBeDefined();
    });

    // Open edit dialog for the reaction step
    const editButtons = screen.getAllByTestId('EditIcon');
    await act(async () => {
      fireEvent.click(editButtons[1].closest('button')!);
    });

    await waitFor(() => {
      expect(screen.getByText('Edit Step: Fullerite-C32')).toBeInTheDocument();
    });

    // Should show "Blueprint detected from ..." without ME/TE values
    await waitFor(() => {
      expect(screen.getByText(/Blueprint detected from Reaction Character/)).toBeInTheDocument();
    });

    // Apply button should NOT be present for reactions
    expect(screen.queryByText('Apply')).not.toBeInTheDocument();
  });

  it('should show Detected chip in edit step dialog when blueprint level exists', async () => {
    mockFetchForPlan(mockPlan, mockMaterials, undefined, { '787': mockBlueprintLevel });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      const blueprintLevelCall = (global.fetch as jest.Mock).mock.calls.find(
        ([url]: [string]) => url === '/api/industry/blueprint-levels',
      );
      expect(blueprintLevelCall).toBeDefined();
    });

    // Open edit dialog for the root step
    const editButtons = screen.getAllByTestId('EditIcon');
    await act(async () => {
      fireEvent.click(editButtons[1].closest('button')!);
    });

    await waitFor(() => {
      expect(screen.getByText('Edit Step: Rifter')).toBeInTheDocument();
      expect(screen.getByText(/Blueprint detected:/)).toBeInTheDocument();
      expect(screen.getByText(/ME 8 \/ TE 16/)).toBeInTheDocument();
    });
  });

  it('should apply detected ME/TE values when Apply button clicked in edit dialog', async () => {
    mockFetchForPlan(mockPlan, mockMaterials, undefined, { '787': mockBlueprintLevel });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      const blueprintLevelCall = (global.fetch as jest.Mock).mock.calls.find(
        ([url]: [string]) => url === '/api/industry/blueprint-levels',
      );
      expect(blueprintLevelCall).toBeDefined();
    });

    // Open edit dialog
    const editButtons = screen.getAllByTestId('EditIcon');
    await act(async () => {
      fireEvent.click(editButtons[1].closest('button')!);
    });

    await waitFor(() => {
      expect(screen.getByText('Apply')).toBeInTheDocument();
    });

    // The ME field should currently show 10 (step's value)
    const meInput = screen.getByLabelText('ME Level') as HTMLInputElement;
    expect(meInput.value).toBe('10');

    // Click Apply
    await act(async () => {
      fireEvent.click(screen.getByText('Apply'));
    });

    // ME should now be 8 (detected value)
    expect(meInput.value).toBe('8');
  });

  it('should show warning icon in step tree when blueprint-levels loaded but step blueprint not found', async () => {
    // blueprint-levels returns a result for a different ID (fetch has completed, keys exist)
    // but the step's blueprintTypeId (787) is not in the response — warning icon should show
    mockFetchForPlan(mockPlan, mockMaterials, undefined, { '99999': mockBlueprintLevel });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      const blueprintLevelCall = (global.fetch as jest.Mock).mock.calls.find(
        ([url]: [string]) => url === '/api/industry/blueprint-levels',
      );
      expect(blueprintLevelCall).toBeDefined();
    });

    // WarningAmberIcon should appear because detectedLevels has keys but 787 is not in it
    await waitFor(() => {
      const warningIcons = screen.queryAllByTestId('WarningAmberIcon');
      expect(warningIcons.length).toBeGreaterThanOrEqual(1);
    });
  });

  it('should show no-blueprint warning chip in EditStepDialog when no detected level', async () => {
    // blueprint-levels returns empty — no blueprint found for step (ID 787)
    mockFetchForPlan(mockPlan, mockMaterials, undefined, {});

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      const blueprintLevelCall = (global.fetch as jest.Mock).mock.calls.find(
        ([url]: [string]) => url === '/api/industry/blueprint-levels',
      );
      expect(blueprintLevelCall).toBeDefined();
    });

    // Open edit dialog for the root step
    const editButtons = screen.getAllByTestId('EditIcon');
    await act(async () => {
      fireEvent.click(editButtons[1].closest('button')!);
    });

    await waitFor(() => {
      expect(screen.getByText('Edit Step: Rifter')).toBeInTheDocument();
      expect(screen.getByText('No blueprint detected — using manual values')).toBeInTheDocument();
    });
  });

  it('should show warning icon on material row when blueprint not detected for that material', async () => {
    // blueprint-levels returns the step blueprint (787) but NOT the material blueprint (99999)
    // detectedLevels will have keys (787), so the "fetch completed" check passes
    // Morphite has hasBlueprint=true, blueprintTypeId=99999 — not in detectedLevels → warning
    mockFetchForPlan(mockPlan, mockMaterials, undefined, { '787': mockBlueprintLevel });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      const blueprintLevelCalls = (global.fetch as jest.Mock).mock.calls.filter(
        ([url]: [string]) => url === '/api/industry/blueprint-levels',
      );
      expect(blueprintLevelCalls.length).toBeGreaterThanOrEqual(1);
    });

    // WarningAmberIcon should appear on the material row for Morphite
    await waitFor(() => {
      const warningIcons = screen.queryAllByTestId('WarningAmberIcon');
      // At minimum one warning icon (could be on step row too, but at least the material one)
      expect(warningIcons.length).toBeGreaterThanOrEqual(1);
    });
  });

  it('should show detected ME/TE chip on material row when blueprint is detected', async () => {
    // blueprint-levels returns the material blueprint (99999)
    // Morphite has hasBlueprint=true, blueprintTypeId=99999 → chip should show
    mockFetchForPlan(mockPlan, mockMaterials, undefined, { '99999': mockBlueprintLevel });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    await waitFor(() => {
      const blueprintLevelCalls = (global.fetch as jest.Mock).mock.calls.filter(
        ([url]: [string]) => url === '/api/industry/blueprint-levels',
      );
      expect(blueprintLevelCalls.length).toBeGreaterThanOrEqual(1);
    });

    // ME/TE chip should appear on the Morphite material row
    await waitFor(() => {
      const meTeChips = screen.queryAllByText('ME 8 / TE 16');
      expect(meTeChips.length).toBeGreaterThanOrEqual(1);
    });
  });

  it('should show character assignments in result when present', async () => {
    mockFetchForPlan(mockPlan, mockMaterials);
    (global.fetch as jest.Mock).mockImplementation((url: string, opts?: any) => {
      if (url === `/api/industry/plans/1`) {
        return Promise.resolve({ ok: true, json: async () => mockPlan });
      }
      if (url === '/api/transport/profiles') {
        return Promise.resolve({ ok: true, json: async () => [] });
      }
      if (url.includes('/materials')) {
        return Promise.resolve({ ok: true, json: async () => mockMaterials });
      }
      if (url.includes('/generate')) {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            run: { id: 1, planId: 1, userId: 100, quantity: 1, createdAt: '', status: 'planned' },
            created: [
              { id: 5, blueprintTypeId: 787, blueprintName: 'Rifter Blueprint', runs: 1, status: 'planned', activity: 'manufacturing', meLevel: 10, teLevel: 20, facilityTax: 1, sortOrder: 0, createdAt: '', updatedAt: '' },
            ],
            skipped: [],
            transportJobs: [],
            characterAssignments: { 5: 'Main' },
            unassignedCount: 0,
          }),
        });
      }
      return Promise.resolve({ ok: true, json: async () => [] });
    });

    await act(async () => {
      render(<ProductionPlanEditor planId={1} />);
    });

    // Open dialog
    const toolbarButton = screen.getAllByText('Generate Jobs')[0];
    fireEvent.click(toolbarButton);

    // Click dialog's Generate Jobs button
    const dialogButton = screen.getAllByText('Generate Jobs');
    await act(async () => {
      fireEvent.click(dialogButton[dialogButton.length - 1]);
    });

    await waitFor(() => {
      expect(screen.getByText('Created 1 job(s)')).toBeInTheDocument();
      expect(screen.getByText('Main')).toBeInTheDocument();
    });
  });
});

import React from 'react';
import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import BatchConfigureTab from '../BatchConfigureTab';
import { ProductionPlan, ProductionPlanStep } from '@industry-tool/client/data/models';

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

const mockFuelStep1: ProductionPlanStep = {
  id: 20,
  planId: 1,
  parentStepId: 10,
  productTypeId: 4247,
  blueprintTypeId: 4248,
  activity: 'manufacturing',
  meLevel: 10,
  teLevel: 20,
  industrySkill: 5,
  advIndustrySkill: 5,
  structure: 'raitaru',
  rig: 't2',
  security: 'high',
  facilityTax: 1.0,
  productName: 'Nitrogen Fuel Block',
  blueprintName: 'Nitrogen Fuel Block Blueprint',
};

const mockFuelStep2: ProductionPlanStep = {
  id: 21,
  planId: 1,
  parentStepId: 10,
  productTypeId: 4247,
  blueprintTypeId: 4248,
  activity: 'manufacturing',
  meLevel: 10,
  teLevel: 20,
  industrySkill: 5,
  advIndustrySkill: 5,
  structure: 'raitaru',
  rig: 't2',
  security: 'high',
  facilityTax: 1.0,
  productName: 'Nitrogen Fuel Block',
  blueprintName: 'Nitrogen Fuel Block Blueprint',
};

const mockReactionStep: ProductionPlanStep = {
  id: 30,
  planId: 1,
  parentStepId: 10,
  productTypeId: 16634,
  blueprintTypeId: 16635,
  activity: 'reaction',
  meLevel: 0,
  teLevel: 0,
  industrySkill: 5,
  advIndustrySkill: 5,
  structure: 'tatara',
  rig: 't1',
  security: 'low',
  facilityTax: 0.25,
  productName: 'Chromium',
  blueprintName: 'Chromium Reaction Formula',
};

const mockPlan: ProductionPlan = {
  id: 1,
  userId: 100,
  productTypeId: 587,
  name: 'Rifter Production',
  createdAt: '2026-02-22T00:00:00Z',
  updatedAt: '2026-02-22T00:00:00Z',
  productName: 'Rifter',
  steps: [mockRootStep, mockFuelStep1, mockFuelStep2, mockReactionStep],
};

const emptyPlan: ProductionPlan = {
  id: 2,
  userId: 100,
  productTypeId: 587,
  name: 'Empty Plan',
  createdAt: '2026-02-22T00:00:00Z',
  updatedAt: '2026-02-22T00:00:00Z',
  productName: 'Rifter',
  steps: [],
};

const mockMaterials = [
  {
    typeId: 34,
    typeName: 'Tritanium',
    quantity: 2000,
    volume: 0.01,
    hasBlueprint: false,
    isProduced: false,
  },
  {
    typeId: 35,
    typeName: 'Pyerite',
    quantity: 1000,
    volume: 0.01,
    hasBlueprint: true,
    blueprintTypeId: 135,
    activity: 'manufacturing',
    isProduced: false,
  },
];

describe('BatchConfigureTab Component', () => {
  const mockOnUpdate = jest.fn();

  beforeEach(() => {
    (global.fetch as jest.Mock).mockClear();
    mockOnUpdate.mockClear();
  });

  it('should match snapshot with grouped steps', () => {
    const { container } = render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should show empty state when no steps', () => {
    render(
      <BatchConfigureTab plan={emptyPlan} planId={2} onUpdate={mockOnUpdate} />,
    );
    expect(screen.getByText('No steps in this plan')).toBeInTheDocument();
  });

  it('should match snapshot with empty plan', () => {
    const { container } = render(
      <BatchConfigureTab plan={emptyPlan} planId={2} onUpdate={mockOnUpdate} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should group steps by product type and activity', () => {
    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    // Should show 3 groups: Chromium (1), Nitrogen Fuel Block (2), Rifter (1)
    expect(screen.getByText('Chromium')).toBeInTheDocument();
    expect(screen.getByText('Nitrogen Fuel Block')).toBeInTheDocument();
    expect(screen.getByText('Rifter')).toBeInTheDocument();
  });

  it('should show correct count for grouped steps', () => {
    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    // Fuel block group has 2 steps
    const chips = screen.getAllByText('2');
    expect(chips.length).toBeGreaterThanOrEqual(1);
  });

  it('should show activity chips', () => {
    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    const mfgChips = screen.getAllByText('manufacturing');
    expect(mfgChips.length).toBe(2); // Rifter and Fuel blocks
    expect(screen.getByText('reaction')).toBeInTheDocument();
  });

  it('should open edit dialog when edit button is clicked', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => [],
    });

    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    const editButtons = screen.getAllByTestId('EditIcon');
    await act(async () => {
      fireEvent.click(editButtons[0].closest('button')!);
    });

    expect(screen.getByText(/Batch Edit:/)).toBeInTheDocument();
  });

  it('should show warning in edit dialog about number of steps', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => [],
    });

    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    // Click edit on the fuel block group (which has 2 steps)
    const editButtons = screen.getAllByTestId('EditIcon');
    // Groups are sorted alphabetically: Chromium, Nitrogen Fuel Block, Rifter
    await act(async () => {
      fireEvent.click(editButtons[1].closest('button')!); // Nitrogen Fuel Block
    });

    expect(screen.getByText(/Changes will apply to all 2/)).toBeInTheDocument();
  });

  it('should display Mixed chip when steps have different ME levels', () => {
    const mixedPlan: ProductionPlan = {
      ...mockPlan,
      steps: [
        mockRootStep,
        mockFuelStep1,
        { ...mockFuelStep2, meLevel: 5 },
        mockReactionStep,
      ],
    };

    render(
      <BatchConfigureTab plan={mixedPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    const mixedChips = screen.getAllByText('Mixed');
    expect(mixedChips.length).toBeGreaterThanOrEqual(1);
  });

  it('should display description text', () => {
    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    expect(
      screen.getByText(/Steps producing the same item are grouped together/),
    ).toBeInTheDocument();
  });

  it('should show expand/collapse icons on each group row', () => {
    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    // All groups should show ChevronRight (collapsed) by default
    const chevrons = screen.getAllByTestId('ChevronRightIcon');
    expect(chevrons.length).toBe(3); // 3 groups
  });

  it('should expand group and fetch materials on click', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockMaterials,
    });

    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    // Click expand on first group (Chromium)
    const chevrons = screen.getAllByTestId('ChevronRightIcon');
    await act(async () => {
      fireEvent.click(chevrons[0].closest('button')!);
    });

    // Should have fetched materials
    expect(global.fetch).toHaveBeenCalledWith(
      '/api/industry/plans/1/steps/30/materials',
    );

    // Should show material names
    await waitFor(() => {
      expect(screen.getByText('Tritanium')).toBeInTheDocument();
      expect(screen.getByText('Pyerite')).toBeInTheDocument();
    });
  });

  it('should show Buy chip for materials not produced', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockMaterials,
    });

    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    // Expand Chromium group
    const chevrons = screen.getAllByTestId('ChevronRightIcon');
    await act(async () => {
      fireEvent.click(chevrons[0].closest('button')!);
    });

    await waitFor(() => {
      const buyChips = screen.getAllByText('Buy');
      expect(buyChips.length).toBeGreaterThanOrEqual(1);
    });
  });

  it('should show toggle button only for materials with blueprints', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockMaterials,
    });

    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    // Expand Chromium group
    const chevrons = screen.getAllByTestId('ChevronRightIcon');
    await act(async () => {
      fireEvent.click(chevrons[0].closest('button')!);
    });

    await waitFor(() => {
      // Pyerite has hasBlueprint=true, so its row should have a BuildIcon toggle button
      // BuildIcon also appears in the build count chip, so check for the toggle button specifically
      const buildIcons = screen.getAllByTestId('BuildIcon');
      const toggleButton = buildIcons.find(
        (icon) => icon.closest('button[title="Switch all to Produce"]'),
      );
      expect(toggleButton).toBeTruthy();
    });
  });

  it('should collapse expanded group on second click', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockMaterials,
    });

    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    // Expand first group
    const chevrons = screen.getAllByTestId('ChevronRightIcon');
    await act(async () => {
      fireEvent.click(chevrons[0].closest('button')!);
    });

    await waitFor(() => {
      expect(screen.getByText('Tritanium')).toBeInTheDocument();
    });

    // Collapse it - the ExpandMore icon should now be visible
    const expandIcon = screen.getByTestId('ExpandMoreIcon');
    await act(async () => {
      fireEvent.click(expandIcon.closest('button')!);
    });

    // Materials should be hidden
    expect(screen.queryByText('Tritanium')).not.toBeInTheDocument();
  });

  it('should show Produce chip for materials with child steps', async () => {
    // Create a plan where fuel step 1 has a child for material typeId 35
    const planWithChild: ProductionPlan = {
      ...mockPlan,
      steps: [
        mockRootStep,
        mockFuelStep1,
        mockFuelStep2,
        mockReactionStep,
        // Child step of fuelStep1 producing Pyerite
        {
          id: 40,
          planId: 1,
          parentStepId: 20,
          productTypeId: 35,
          blueprintTypeId: 135,
          activity: 'manufacturing',
          meLevel: 10,
          teLevel: 20,
          industrySkill: 5,
          advIndustrySkill: 5,
          structure: 'raitaru',
          rig: 't2',
          security: 'high',
          facilityTax: 1.0,
          productName: 'Pyerite',
          blueprintName: 'Pyerite Blueprint',
        },
        // Child step of fuelStep2 producing Pyerite
        {
          id: 41,
          planId: 1,
          parentStepId: 21,
          productTypeId: 35,
          blueprintTypeId: 135,
          activity: 'manufacturing',
          meLevel: 10,
          teLevel: 20,
          industrySkill: 5,
          advIndustrySkill: 5,
          structure: 'raitaru',
          rig: 't2',
          security: 'high',
          facilityTax: 1.0,
          productName: 'Pyerite',
          blueprintName: 'Pyerite Blueprint',
        },
      ],
    };

    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockMaterials,
    });

    render(
      <BatchConfigureTab plan={planWithChild} planId={1} onUpdate={mockOnUpdate} />,
    );

    // Expand fuel block group (index 1 alphabetically: Nitrogen Fuel Block)
    const chevrons = screen.getAllByTestId('ChevronRightIcon');
    await act(async () => {
      // Groups: Chromium, Nitrogen Fuel Block, Pyerite, Rifter
      fireEvent.click(chevrons[1].closest('button')!);
    });

    await waitFor(() => {
      // Pyerite should show "Produce" chip since both fuel steps have child steps for it
      expect(screen.getByText('Produce')).toBeInTheDocument();
    });
  });

  it('should show input/output location when configured', () => {
    const planWithLocations: ProductionPlan = {
      ...mockPlan,
      steps: [
        mockRootStep,
        {
          ...mockFuelStep1,
          sourceOwnerName: 'My Corp',
          sourceDivisionName: 'Hangar 1',
          sourceContainerName: 'Materials',
          outputOwnerName: 'My Corp',
          outputDivisionName: 'Hangar 2',
        },
        {
          ...mockFuelStep2,
          sourceOwnerName: 'My Corp',
          sourceDivisionName: 'Hangar 1',
          sourceContainerName: 'Materials',
          outputOwnerName: 'My Corp',
          outputDivisionName: 'Hangar 2',
        },
        mockReactionStep,
      ],
    };

    render(
      <BatchConfigureTab plan={planWithLocations} planId={1} onUpdate={mockOnUpdate} />,
    );

    expect(screen.getByText('In: My Corp / Hangar 1 / Materials')).toBeInTheDocument();
    expect(screen.getByText('Out: My Corp / Hangar 2')).toBeInTheDocument();
  });

  it('should show Mixed chip for input location when steps differ', () => {
    const planWithMixedLocations: ProductionPlan = {
      ...mockPlan,
      steps: [
        mockRootStep,
        {
          ...mockFuelStep1,
          sourceOwnerName: 'Corp A',
        },
        {
          ...mockFuelStep2,
          sourceOwnerName: 'Corp B',
        },
        mockReactionStep,
      ],
    };

    render(
      <BatchConfigureTab plan={planWithMixedLocations} planId={1} onUpdate={mockOnUpdate} />,
    );

    // The fuel block group should show "In: Mixed"
    const inTexts = screen.getAllByText(/^In:/);
    expect(inTexts.length).toBeGreaterThanOrEqual(1);
  });

  it('should show set-all-build and set-all-buy buttons on each group row', () => {
    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    // Each group row has a build button for "Set all to Build"
    // 3 groups = 3 build buttons + 1 chip icon (Rifter group has children) = 4 total BuildIcons
    const allBuildIcons = screen.getAllByTestId('BuildIcon');
    const buildButtons = allBuildIcons.filter(
      (icon) => icon.closest('button[class*="IconButton"]'),
    );
    expect(buildButtons.length).toBeGreaterThanOrEqual(3);

    // Each group row has a buy chip + a "Set all to Buy" button = 2 per group = 6 total
    const allCartIcons = screen.getAllByTestId('ShoppingCartIcon');
    expect(allCartIcons.length).toBe(6);
  });

  it('should call onUpdate after set all to buy removes child steps', async () => {
    // Plan with child steps under fuel steps
    const planWithChildren: ProductionPlan = {
      ...mockPlan,
      steps: [
        mockRootStep,
        mockFuelStep1,
        mockFuelStep2,
        mockReactionStep,
        {
          id: 40,
          planId: 1,
          parentStepId: 20,
          productTypeId: 35,
          blueprintTypeId: 135,
          activity: 'manufacturing',
          meLevel: 10,
          teLevel: 20,
          industrySkill: 5,
          advIndustrySkill: 5,
          structure: 'raitaru',
          rig: 't2',
          security: 'high',
          facilityTax: 1.0,
          productName: 'Pyerite',
          blueprintName: 'Pyerite Blueprint',
        },
      ],
    };

    (global.fetch as jest.Mock).mockResolvedValue({ ok: true, json: async () => ({}) });

    render(
      <BatchConfigureTab plan={planWithChildren} planId={1} onUpdate={mockOnUpdate} />,
    );

    // Find "Set all to Buy" buttons — these are IconButtons containing ShoppingCartIcon
    // (distinct from the Chip icons which are not inside IconButtons)
    const cartButtons = screen.getAllByTestId('ShoppingCartIcon')
      .filter((icon) => icon.closest('button.MuiIconButton-root'))
      .map((icon) => icon.closest('button')!);
    // Groups alphabetically: Chromium (0), Nitrogen Fuel Block (1), Pyerite (2), Rifter (3)
    await act(async () => {
      fireEvent.click(cartButtons[1]);
    });

    // Should have called DELETE for the child step
    expect(global.fetch).toHaveBeenCalledWith(
      '/api/industry/plans/1/steps/40',
      { method: 'DELETE' },
    );
    expect(mockOnUpdate).toHaveBeenCalled();
  });

  it('should mention expand feature in description text', () => {
    render(
      <BatchConfigureTab plan={mockPlan} planId={1} onUpdate={mockOnUpdate} />,
    );

    expect(
      screen.getByText(/Expand a group to toggle materials between buy and produce/),
    ).toBeInTheDocument();
  });

  describe('Blueprint auto-detection', () => {
    const mockDetectedLevels = {
      787: {
        materialEfficiency: 8,
        timeEfficiency: 16,
        isCopy: false,
        ownerName: 'Test Character',
        runs: -1,
      },
    };

    it('should show Apply Blueprint ME/TE button for each group row', () => {
      render(
        <BatchConfigureTab
          plan={mockPlan}
          planId={1}
          onUpdate={mockOnUpdate}
          detectedLevels={mockDetectedLevels}
        />,
      );

      // AutoFixHighIcon should be present in the actions column for groups
      const autoFixIcons = screen.getAllByTestId('AutoFixHighIcon');
      expect(autoFixIcons.length).toBe(3); // 3 groups
    });

    it('should enable Apply button for group with detected levels', () => {
      render(
        <BatchConfigureTab
          plan={mockPlan}
          planId={1}
          onUpdate={mockOnUpdate}
          detectedLevels={mockDetectedLevels}
        />,
      );

      // The Rifter group has blueprint 787 with a detected level
      // The button for that group should be enabled (not disabled)
      const autoFixButtons = screen.getAllByTestId('AutoFixHighIcon')
        .map((icon) => icon.closest('button')!);

      // Groups are sorted alphabetically: Chromium (bp 16635), Nitrogen Fuel Block (bp 4248), Rifter (bp 787)
      // Only Rifter (index 2) has a detected level
      expect(autoFixButtons[2]).not.toBeDisabled();
    });

    it('should disable Apply button for group with no detected levels', () => {
      render(
        <BatchConfigureTab
          plan={mockPlan}
          planId={1}
          onUpdate={mockOnUpdate}
          detectedLevels={mockDetectedLevels}
        />,
      );

      const autoFixButtons = screen.getAllByTestId('AutoFixHighIcon')
        .map((icon) => icon.closest('button')!);

      // Chromium (index 0) has blueprint 16635 - no detected level
      expect(autoFixButtons[0]).toBeDisabled();
    });

    it('should call batch update API when Apply Blueprint ME/TE is clicked', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({}),
      });

      render(
        <BatchConfigureTab
          plan={mockPlan}
          planId={1}
          onUpdate={mockOnUpdate}
          detectedLevels={mockDetectedLevels}
        />,
      );

      // Click Apply for Rifter group (index 2, blueprint 787 has detected level)
      const autoFixButtons = screen.getAllByTestId('AutoFixHighIcon')
        .map((icon) => icon.closest('button')!);

      await act(async () => {
        fireEvent.click(autoFixButtons[2]);
      });

      await waitFor(() => {
        const batchCall = (global.fetch as jest.Mock).mock.calls.find(
          ([url]: [string]) => url === '/api/industry/plans/1/steps/batch',
        );
        expect(batchCall).toBeDefined();
        const body = JSON.parse(batchCall[1].body);
        expect(body.me_level).toBe(8);
        expect(body.te_level).toBe(16);
      });
    });

    it('should show success snackbar after applying detected ME/TE', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({}),
      });

      render(
        <BatchConfigureTab
          plan={mockPlan}
          planId={1}
          onUpdate={mockOnUpdate}
          detectedLevels={mockDetectedLevels}
        />,
      );

      const autoFixButtons = screen.getAllByTestId('AutoFixHighIcon')
        .map((icon) => icon.closest('button')!);

      await act(async () => {
        fireEvent.click(autoFixButtons[2]);
      });

      await waitFor(() => {
        expect(screen.getByText(/Applied detected ME\/TE to/)).toBeInTheDocument();
      });
    });

    it('should call onUpdate after successfully applying detected ME/TE', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({}),
      });

      render(
        <BatchConfigureTab
          plan={mockPlan}
          planId={1}
          onUpdate={mockOnUpdate}
          detectedLevels={mockDetectedLevels}
        />,
      );

      const autoFixButtons = screen.getAllByTestId('AutoFixHighIcon')
        .map((icon) => icon.closest('button')!);

      await act(async () => {
        fireEvent.click(autoFixButtons[2]);
      });

      await waitFor(() => {
        expect(mockOnUpdate).toHaveBeenCalled();
      });
    });

    it('should show Detected chip in batch edit dialog when detectedLevel is provided', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => [],
      });

      render(
        <BatchConfigureTab
          plan={mockPlan}
          planId={1}
          onUpdate={mockOnUpdate}
          detectedLevels={mockDetectedLevels}
        />,
      );

      // Open edit dialog for Rifter group (index 2, has detected level for bp 787)
      const editButtons = screen.getAllByTestId('EditIcon');
      // Groups alphabetically: Chromium, Nitrogen Fuel Block, Rifter — editButtons index 2 = Rifter
      await act(async () => {
        fireEvent.click(editButtons[2].closest('button')!);
      });

      await waitFor(() => {
        expect(screen.getByText(/Blueprint detected:/)).toBeInTheDocument();
        expect(screen.getByText(/ME 8 \/ TE 16/)).toBeInTheDocument();
      });
    });

    it('should apply detected values when Apply clicked in batch edit dialog', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => [],
      });

      render(
        <BatchConfigureTab
          plan={mockPlan}
          planId={1}
          onUpdate={mockOnUpdate}
          detectedLevels={mockDetectedLevels}
        />,
      );

      // Open edit for Rifter group (blueprint 787, detected ME 8/TE 16)
      const editButtons = screen.getAllByTestId('EditIcon');
      await act(async () => {
        fireEvent.click(editButtons[2].closest('button')!);
      });

      await waitFor(() => {
        expect(screen.getByText('Apply')).toBeInTheDocument();
      });

      // ME field should be 10 (group default)
      const meInput = screen.getByLabelText('ME Level') as HTMLInputElement;
      expect(meInput.value).toBe('10');

      // Click Apply
      await act(async () => {
        fireEvent.click(screen.getByText('Apply'));
      });

      // ME should be updated to detected value 8
      expect(meInput.value).toBe('8');
    });

    it('should not show Apply button in batch edit dialog when no detected level', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => [],
      });

      render(
        <BatchConfigureTab
          plan={mockPlan}
          planId={1}
          onUpdate={mockOnUpdate}
          detectedLevels={mockDetectedLevels}
        />,
      );

      // Open edit for Chromium (index 0, no detected level for bp 16635)
      const editButtons = screen.getAllByTestId('EditIcon');
      await act(async () => {
        fireEvent.click(editButtons[0].closest('button')!);
      });

      await waitFor(() => {
        expect(screen.getByText(/Batch Edit:/)).toBeInTheDocument();
      });

      // No detected chip for Chromium
      expect(screen.queryByText(/Blueprint detected:/)).not.toBeInTheDocument();
    });
  });
});

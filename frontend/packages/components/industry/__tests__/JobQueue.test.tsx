import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import JobQueue from '../JobQueue';
import { IndustryJobQueueEntry } from '@industry-tool/client/data/models';

describe('JobQueue Component', () => {
  const mockOnCancel = jest.fn();

  beforeEach(() => {
    mockOnCancel.mockClear();
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-02-22T12:00:00Z'));
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('should match snapshot when loading', () => {
    const { container } = render(
      <JobQueue entries={[]} loading={true} onCancel={mockOnCancel} />
    );
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with empty queue', () => {
    const { container } = render(
      <JobQueue entries={[]} loading={false} onCancel={mockOnCancel} />
    );
    expect(container).toMatchSnapshot();
  });

  it('should display empty message when no entries', () => {
    render(<JobQueue entries={[]} loading={false} onCancel={mockOnCancel} />);
    expect(screen.getByText('No jobs in queue')).toBeInTheDocument();
  });

  it('should match snapshot with entries', () => {
    const entries: IndustryJobQueueEntry[] = [
      {
        id: 1,
        userId: 100,
        blueprintTypeId: 787,
        activity: 'manufacturing',
        runs: 10,
        meLevel: 10,
        teLevel: 20,
        facilityTax: 1.0,
        status: 'planned',
        sortOrder: 2,
        estimatedCost: 5000000,
        estimatedDuration: 3600,
        createdAt: '2026-02-22T00:00:00Z',
        updatedAt: '2026-02-22T00:00:00Z',
        blueprintName: 'Rifter Blueprint',
        productName: 'Rifter',
        notes: 'Test note',
        stationName: 'Jita IV - Moon 4 - Caldari Navy Assembly Plant',
        inputLocation: 'CharOwner > Big Container',
        outputLocation: 'CorpOwner > Hangar 1',
      },
      {
        id: 2,
        userId: 100,
        blueprintTypeId: 46166,
        activity: 'reaction',
        runs: 100,
        meLevel: 0,
        teLevel: 0,
        facilityTax: 0.25,
        status: 'active',
        sortOrder: 4,
        esiJobId: 12345,
        createdAt: '2026-02-22T00:00:00Z',
        updatedAt: '2026-02-22T00:00:00Z',
        blueprintName: 'Reaction Formula',
        esiJobEndDate: '2026-02-26T04:33:00Z',
      },
    ];

    const { container } = render(
      <JobQueue entries={entries} loading={false} onCancel={mockOnCancel} />
    );
    expect(container).toMatchSnapshot();
  });

  it('should display entry details with station and locations', () => {
    const entries: IndustryJobQueueEntry[] = [
      {
        id: 1,
        userId: 100,
        blueprintTypeId: 787,
        activity: 'manufacturing',
        runs: 10,
        meLevel: 10,
        teLevel: 20,
        facilityTax: 1.0,
        status: 'planned',
        sortOrder: 0,
        createdAt: '2026-02-22T00:00:00Z',
        updatedAt: '2026-02-22T00:00:00Z',
        blueprintName: 'Rifter Blueprint',
        stationName: 'Jita IV - Moon 4 - Caldari Navy Assembly Plant',
        inputLocation: 'MyCharacter > Ammo Box',
        outputLocation: 'MyCorp > Division 1',
      },
    ];

    render(<JobQueue entries={entries} loading={false} onCancel={mockOnCancel} />);
    expect(screen.getByText('Rifter Blueprint')).toBeInTheDocument();
    expect(screen.getByText('planned')).toBeInTheDocument();
    expect(screen.getByText('10/20')).toBeInTheDocument();
    expect(screen.getByText('Jita IV - Moon 4 - Caldari Navy Assembly Plant')).toBeInTheDocument();
    expect(screen.getByText('MyCharacter > Ammo Box')).toBeInTheDocument();
    expect(screen.getByText('MyCorp > Division 1')).toBeInTheDocument();
  });

  it('should call onCancel when cancel button is clicked', () => {
    const entries: IndustryJobQueueEntry[] = [
      {
        id: 5,
        userId: 100,
        blueprintTypeId: 787,
        activity: 'manufacturing',
        runs: 10,
        meLevel: 10,
        teLevel: 20,
        facilityTax: 1.0,
        status: 'planned',
        sortOrder: 0,
        createdAt: '2026-02-22T00:00:00Z',
        updatedAt: '2026-02-22T00:00:00Z',
      },
    ];

    render(<JobQueue entries={entries} loading={false} onCancel={mockOnCancel} />);
    const cancelButton = screen.getByTitle('Cancel job');
    fireEvent.click(cancelButton);
    expect(mockOnCancel).toHaveBeenCalledWith(5);
  });

  it('should display finish time for active entries with esiJobEndDate', () => {
    const entries: IndustryJobQueueEntry[] = [
      {
        id: 10,
        userId: 100,
        blueprintTypeId: 787,
        activity: 'manufacturing',
        runs: 1,
        meLevel: 10,
        teLevel: 20,
        facilityTax: 1.0,
        status: 'active',
        sortOrder: 0,
        esiJobId: 99999,
        esiJobEndDate: '2026-02-26T04:33:00Z',
        createdAt: '2026-02-22T00:00:00Z',
        updatedAt: '2026-02-22T00:00:00Z',
        blueprintName: 'Rifter Blueprint',
      },
    ];

    render(<JobQueue entries={entries} loading={false} onCancel={mockOnCancel} />);
    expect(screen.getByText('2026.02.26 04:33')).toBeInTheDocument();
  });

  it('should match snapshot with transport entry', () => {
    const entries: IndustryJobQueueEntry[] = [
      {
        id: 20,
        userId: 100,
        blueprintTypeId: 0,
        activity: 'transport',
        runs: 0,
        meLevel: 0,
        teLevel: 0,
        facilityTax: 0,
        status: 'planned',
        sortOrder: 3,
        estimatedCost: 12500000,
        createdAt: '2026-02-22T00:00:00Z',
        updatedAt: '2026-02-22T00:00:00Z',
        transportJobId: 5,
        transportOriginName: 'Jita IV - Moon 4 - Caldari Navy Assembly Plant',
        transportDestName: 'Amarr VIII - Oris - Emperor Family Academy',
        transportMethod: 'jump_freighter',
        transportFulfillment: 'self_haul',
        transportVolumeM3: 125000,
        transportJumps: 12,
        transportItemsSummary: 'Hydrogen Fuel Block x500, Oxygen Fuel Block x300',
      },
    ];

    const { container } = render(
      <JobQueue entries={entries} loading={false} onCancel={mockOnCancel} />
    );
    expect(container).toMatchSnapshot();
  });

  it('should display transport route, items, and method', () => {
    const entries: IndustryJobQueueEntry[] = [
      {
        id: 21,
        userId: 100,
        blueprintTypeId: 0,
        activity: 'transport',
        runs: 0,
        meLevel: 0,
        teLevel: 0,
        facilityTax: 0,
        status: 'planned',
        sortOrder: 1,
        createdAt: '2026-02-22T00:00:00Z',
        updatedAt: '2026-02-22T00:00:00Z',
        transportOriginName: 'Jita IV - Moon 4 - Caldari Navy Assembly Plant',
        transportDestName: 'Amarr VIII - Oris - Emperor Family Academy',
        transportMethod: 'freighter',
        transportFulfillment: 'courier_contract',
        transportVolumeM3: 50000,
        transportJumps: 15,
        transportItemsSummary: 'Tritanium x10000, Pyerite x5000',
      },
    ];

    render(<JobQueue entries={entries} loading={false} onCancel={mockOnCancel} />);
    expect(screen.getByText(/Jita IV.*→.*Amarr VIII/)).toBeInTheDocument();
    expect(screen.getByText('Freighter')).toBeInTheDocument();
    expect(screen.getByText('15 jumps')).toBeInTheDocument();
    expect(screen.getByText('50,000 m³')).toBeInTheDocument();
    expect(screen.getByText('Courier Contract')).toBeInTheDocument();
    expect(screen.getByText('Tritanium x10000, Pyerite x5000')).toBeInTheDocument();
  });

  it('should not show cancel button for completed entries', () => {
    const entries: IndustryJobQueueEntry[] = [
      {
        id: 3,
        userId: 100,
        blueprintTypeId: 787,
        activity: 'manufacturing',
        runs: 10,
        meLevel: 10,
        teLevel: 20,
        facilityTax: 1.0,
        status: 'completed',
        sortOrder: 0,
        createdAt: '2026-02-22T00:00:00Z',
        updatedAt: '2026-02-22T00:00:00Z',
      },
    ];

    render(<JobQueue entries={entries} loading={false} onCancel={mockOnCancel} />);
    expect(screen.queryByTitle('Cancel job')).not.toBeInTheDocument();
  });
});

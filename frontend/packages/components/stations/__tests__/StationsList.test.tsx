import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import StationsList from '../StationsList';
import { UserStation } from '@industry-tool/client/data/models';

jest.mock('../StationDialog', () => {
  return function MockStationDialog({ open, station, onClose }: any) {
    if (!open) return null;
    return (
      <div data-testid="station-dialog">
        <span>{station ? 'Edit' : 'Add'} Station Dialog</span>
        <button onClick={() => onClose(false)}>Close</button>
      </div>
    );
  };
});

const mockStations: UserStation[] = [
  {
    id: 1,
    userId: 100,
    stationId: 60003760,
    structure: 'sotiyo',
    facilityTax: 1.0,
    createdAt: '2026-02-22T12:00:00Z',
    updatedAt: '2026-02-22T12:00:00Z',
    stationName: 'Test Sotiyo',
    solarSystemName: 'Jita',
    securityStatus: 0.9,
    security: 'high',
    rigs: [
      { id: 1, userStationId: 1, rigName: 'Standup XL-Set Ship Manufacturing Efficiency I', category: 'ship', tier: 't1' },
      { id: 2, userStationId: 1, rigName: 'Standup XL-Set Structure and Component Manufacturing Efficiency I', category: 'component', tier: 't1' },
    ],
    services: [
      { id: 1, userStationId: 1, serviceName: 'Standup Manufacturing Plant I', activity: 'manufacturing' },
    ],
    activities: ['manufacturing'],
  },
  {
    id: 2,
    userId: 100,
    stationId: 60003761,
    structure: 'tatara',
    facilityTax: 2.5,
    createdAt: '2026-02-22T12:00:00Z',
    updatedAt: '2026-02-22T12:00:00Z',
    stationName: 'Test Tatara',
    solarSystemName: 'Perimeter',
    securityStatus: 0.3,
    security: 'low',
    rigs: [
      { id: 3, userStationId: 2, rigName: 'Standup L-Set Biochemical Reactor Efficiency II', category: 'reaction', tier: 't2' },
    ],
    services: [
      { id: 2, userStationId: 2, serviceName: 'Standup Biochemical Reactor I', activity: 'reaction' },
    ],
    activities: ['reaction'],
  },
];

describe('StationsList Component', () => {
  beforeEach(() => {
    (global.fetch as jest.Mock).mockClear();
  });

  it('should match snapshot while loading', () => {
    const { container } = render(
      <StationsList stations={[]} loading={true} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with empty stations', () => {
    const { container } = render(
      <StationsList stations={[]} loading={false} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with stations', () => {
    const { container } = render(
      <StationsList stations={mockStations} loading={false} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should display empty message when no stations', () => {
    render(
      <StationsList stations={[]} loading={false} onRefresh={jest.fn()} />,
    );
    expect(screen.getByText(/No preferred stations configured/)).toBeInTheDocument();
  });

  it('should display station details', () => {
    render(
      <StationsList stations={mockStations} loading={false} onRefresh={jest.fn()} />,
    );

    expect(screen.getByText('Test Sotiyo')).toBeInTheDocument();
    expect(screen.getByText('Jita')).toBeInTheDocument();
    expect(screen.getByText('high')).toBeInTheDocument();
    expect(screen.getByText('sotiyo')).toBeInTheDocument();

    expect(screen.getByText('Test Tatara')).toBeInTheDocument();
    expect(screen.getByText('Perimeter')).toBeInTheDocument();
    expect(screen.getByText('low')).toBeInTheDocument();
  });

  it('should display rig chips', () => {
    render(
      <StationsList stations={mockStations} loading={false} onRefresh={jest.fn()} />,
    );

    expect(screen.getByText('ship T1')).toBeInTheDocument();
    expect(screen.getByText('component T1')).toBeInTheDocument();
    expect(screen.getByText('reaction T2')).toBeInTheDocument();
  });

  it('should display activity chips', () => {
    render(
      <StationsList stations={mockStations} loading={false} onRefresh={jest.fn()} />,
    );

    expect(screen.getAllByText('manufacturing')).toHaveLength(1);
    expect(screen.getAllByText('reaction')).toHaveLength(1);
  });

  it('should open add dialog when Add Station is clicked', () => {
    render(
      <StationsList stations={mockStations} loading={false} onRefresh={jest.fn()} />,
    );

    fireEvent.click(screen.getByText('Add Station'));
    expect(screen.getByTestId('station-dialog')).toBeInTheDocument();
    expect(screen.getByText('Add Station Dialog')).toBeInTheDocument();
  });

  it('should open edit dialog when edit button is clicked', () => {
    render(
      <StationsList stations={mockStations} loading={false} onRefresh={jest.fn()} />,
    );

    const editButtons = screen.getAllByTestId('EditIcon');
    fireEvent.click(editButtons[0].closest('button')!);
    expect(screen.getByTestId('station-dialog')).toBeInTheDocument();
    expect(screen.getByText('Edit Station Dialog')).toBeInTheDocument();
  });

  it('should display facility tax', () => {
    render(
      <StationsList stations={mockStations} loading={false} onRefresh={jest.fn()} />,
    );

    expect(screen.getByText('1%')).toBeInTheDocument();
    expect(screen.getByText('2.5%')).toBeInTheDocument();
  });
});

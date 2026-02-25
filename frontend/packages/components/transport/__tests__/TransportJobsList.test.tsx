import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { TransportJobsList } from '../TransportJobsList';
import { TransportJob, TransportProfile, JFRoute } from '@industry-tool/pages/transport';

jest.mock('../TransportJobDialog', () => ({
  TransportJobDialog: function MockDialog({ open, onClose }: any) {
    if (!open) return null;
    return (
      <div data-testid="job-dialog">
        <span>Create Job Dialog</span>
        <button onClick={() => onClose(false)}>Close</button>
      </div>
    );
  },
}));

const mockJobs: TransportJob[] = [
  {
    id: 1,
    userId: 42,
    originStationId: 60003760,
    originStationName: 'Jita IV - Moon 4 - Caldari Navy',
    originSystemId: 30000142,
    originSystemName: 'Jita',
    destinationStationId: 60008494,
    destinationStationName: 'Amarr VIII',
    destinationSystemId: 30002187,
    destinationSystemName: 'Amarr',
    transportMethod: 'freighter',
    routePreference: 'shortest',
    totalVolumeM3: 50000,
    totalCollateral: 1000000000,
    estimatedCost: 5000000,
    jumps: 9,
    fulfillmentType: 'self_haul',
    transportProfileId: 1,
    transportProfileName: 'Charon',
    status: 'planned',
    items: [],
    createdAt: '2026-02-24T12:00:00Z',
  },
  {
    id: 2,
    userId: 42,
    originStationId: 60003760,
    originStationName: 'Jita IV - Moon 4 - Caldari Navy',
    originSystemId: 30000142,
    originSystemName: 'Jita',
    destinationStationId: 60008494,
    destinationStationName: 'Amarr VIII',
    destinationSystemId: 30002187,
    destinationSystemName: 'Amarr',
    transportMethod: 'jump_freighter',
    routePreference: 'shortest',
    totalVolumeM3: 30000,
    totalCollateral: 500000000,
    estimatedCost: 25000000,
    jumps: 2,
    distanceLy: 8.5,
    fulfillmentType: 'self_haul',
    status: 'in_transit',
    items: [],
    createdAt: '2026-02-24T12:00:00Z',
  },
];

const mockProfiles: TransportProfile[] = [];
const mockJFRoutes: JFRoute[] = [];

describe('TransportJobsList', () => {
  beforeEach(() => {
    (global.fetch as jest.Mock).mockClear();
  });

  it('should match snapshot while loading', () => {
    const { container } = render(
      <TransportJobsList jobs={[]} loading={true} profiles={mockProfiles} jfRoutes={mockJFRoutes} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with empty jobs', () => {
    const { container } = render(
      <TransportJobsList jobs={[]} loading={false} profiles={mockProfiles} jfRoutes={mockJFRoutes} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with jobs', () => {
    const { container } = render(
      <TransportJobsList jobs={mockJobs} loading={false} profiles={mockProfiles} jfRoutes={mockJFRoutes} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should display empty message when no jobs', () => {
    render(
      <TransportJobsList jobs={[]} loading={false} profiles={mockProfiles} jfRoutes={mockJFRoutes} onRefresh={jest.fn()} />,
    );
    expect(screen.getByText(/No transport jobs/)).toBeInTheDocument();
  });

  it('should display job status chips', () => {
    render(
      <TransportJobsList jobs={mockJobs} loading={false} profiles={mockProfiles} jfRoutes={mockJFRoutes} onRefresh={jest.fn()} />,
    );

    expect(screen.getByText('Planned')).toBeInTheDocument();
    expect(screen.getByText('In Transit')).toBeInTheDocument();
  });

  it('should display route information', () => {
    render(
      <TransportJobsList jobs={mockJobs} loading={false} profiles={mockProfiles} jfRoutes={mockJFRoutes} onRefresh={jest.fn()} />,
    );

    expect(screen.getAllByText(/Jita → Amarr/)).toHaveLength(2);
  });

  it('should show Start and Cancel buttons for planned jobs', () => {
    render(
      <TransportJobsList jobs={mockJobs} loading={false} profiles={mockProfiles} jfRoutes={mockJFRoutes} onRefresh={jest.fn()} />,
    );

    expect(screen.getByText('Start')).toBeInTheDocument();
    expect(screen.getByText('Cancel')).toBeInTheDocument();
  });

  it('should show Delivered button for in_transit jobs', () => {
    render(
      <TransportJobsList jobs={mockJobs} loading={false} profiles={mockProfiles} jfRoutes={mockJFRoutes} onRefresh={jest.fn()} />,
    );

    expect(screen.getByText('Delivered')).toBeInTheDocument();
  });

  it('should open create dialog when Create Transport Job is clicked', () => {
    render(
      <TransportJobsList jobs={mockJobs} loading={false} profiles={mockProfiles} jfRoutes={mockJFRoutes} onRefresh={jest.fn()} />,
    );

    fireEvent.click(screen.getByText('Create Transport Job'));
    expect(screen.getByTestId('job-dialog')).toBeInTheDocument();
    expect(screen.getByText('Create Job Dialog')).toBeInTheDocument();
  });
});

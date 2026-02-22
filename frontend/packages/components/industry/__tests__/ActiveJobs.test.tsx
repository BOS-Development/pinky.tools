import React from 'react';
import { render, screen } from '@testing-library/react';
import ActiveJobs from '../ActiveJobs';
import { IndustryJob } from '@industry-tool/client/data/models';

describe('ActiveJobs Component', () => {
  it('should match snapshot when loading', () => {
    const { container } = render(<ActiveJobs jobs={[]} loading={true} />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with empty jobs', () => {
    const { container } = render(<ActiveJobs jobs={[]} loading={false} />);
    expect(container).toMatchSnapshot();
  });

  it('should display empty message when no jobs', () => {
    render(<ActiveJobs jobs={[]} loading={false} />);
    expect(screen.getByText('No active industry jobs')).toBeInTheDocument();
  });

  it('should match snapshot with jobs', () => {
    const jobs: IndustryJob[] = [
      {
        jobId: 10001,
        installerId: 1001,
        userId: 100,
        facilityId: 60003760,
        stationId: 60003760,
        activityId: 1,
        blueprintId: 9876,
        blueprintTypeId: 787,
        blueprintLocationId: 60003760,
        outputLocationId: 60003760,
        runs: 10,
        cost: 1500000,
        status: 'active',
        source: 'character',
        duration: 3600,
        startDate: '2026-02-22T00:00:00Z',
        endDate: '2026-02-22T01:00:00Z',
        updatedAt: '2026-02-22T00:00:00Z',
        blueprintName: 'Rifter Blueprint',
        productName: 'Rifter',
        activityName: 'Manufacturing',
        installerName: 'Test Pilot',
        systemName: 'Jita',
      },
      {
        jobId: 10002,
        installerId: 1002,
        userId: 100,
        facilityId: 60003760,
        stationId: 60003760,
        activityId: 9,
        blueprintId: 1111,
        blueprintTypeId: 46166,
        blueprintLocationId: 60003760,
        outputLocationId: 60003760,
        runs: 100,
        status: 'ready',
        source: 'corporation',
        duration: 7200,
        startDate: '2026-02-21T22:00:00Z',
        endDate: '2026-02-22T00:00:00Z',
        updatedAt: '2026-02-22T00:00:00Z',
        blueprintName: 'Reaction Blueprint',
        activityName: 'Reaction',
        installerName: 'Corp Alt',
      },
    ];

    const { container } = render(<ActiveJobs jobs={jobs} loading={false} />);
    expect(container).toMatchSnapshot();
  });

  it('should display job details', () => {
    const jobs: IndustryJob[] = [
      {
        jobId: 10001,
        installerId: 1001,
        userId: 100,
        facilityId: 60003760,
        stationId: 60003760,
        activityId: 1,
        blueprintId: 9876,
        blueprintTypeId: 787,
        blueprintLocationId: 60003760,
        outputLocationId: 60003760,
        runs: 10,
        cost: 1500000,
        status: 'active',
        source: 'character',
        duration: 3600,
        startDate: '2026-02-22T00:00:00Z',
        endDate: '2026-02-22T01:00:00Z',
        updatedAt: '2026-02-22T00:00:00Z',
        blueprintName: 'Rifter Blueprint',
        productName: 'Rifter',
        activityName: 'Manufacturing',
        installerName: 'Test Pilot',
        systemName: 'Jita',
      },
    ];

    render(<ActiveJobs jobs={jobs} loading={false} />);
    expect(screen.getByText('Rifter Blueprint')).toBeInTheDocument();
    expect(screen.getByText('Rifter')).toBeInTheDocument();
    expect(screen.getByText('Manufacturing')).toBeInTheDocument();
    expect(screen.getByText('active')).toBeInTheDocument();
    expect(screen.getByText('Test Pilot')).toBeInTheDocument();
    expect(screen.getByText('Jita')).toBeInTheDocument();
  });
});

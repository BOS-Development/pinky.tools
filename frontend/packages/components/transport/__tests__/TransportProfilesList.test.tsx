import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { TransportProfilesList } from '../TransportProfilesList';
import { TransportProfile } from '@industry-tool/pages/transport';

jest.mock('../TransportProfileDialog', () => ({
  TransportProfileDialog: function MockDialog({ open, profile, onClose }: any) {
    if (!open) return null;
    return (
      <div data-testid="profile-dialog">
        <span>{profile ? 'Edit' : 'Add'} Profile Dialog</span>
        <button onClick={() => onClose(false)}>Close</button>
      </div>
    );
  },
}));

const mockProfiles: TransportProfile[] = [
  {
    id: 1,
    userId: 42,
    name: 'Charon Freighter',
    transportMethod: 'freighter',
    cargoM3: 435000,
    ratePerM3PerJump: 800,
    collateralRate: 0.01,
    collateralPriceBasis: 'sell',
    fuelConservationLevel: 0,
    routePreference: 'shortest',
    isDefault: true,
    createdAt: '2026-02-24T12:00:00Z',
  },
  {
    id: 2,
    userId: 42,
    name: 'Rhea JF',
    transportMethod: 'jump_freighter',
    cargoM3: 320000,
    ratePerM3PerJump: 0,
    collateralRate: 0.02,
    collateralPriceBasis: 'sell',
    fuelTypeId: 16274,
    fuelTypeName: 'Nitrogen Isotopes',
    fuelPerLy: 500,
    fuelConservationLevel: 5,
    routePreference: 'shortest',
    isDefault: false,
    createdAt: '2026-02-24T12:00:00Z',
  },
];

describe('TransportProfilesList', () => {
  beforeEach(() => {
    (global.fetch as jest.Mock).mockClear();
  });

  it('should match snapshot while loading', () => {
    const { container } = render(
      <TransportProfilesList profiles={[]} loading={true} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with empty profiles', () => {
    const { container } = render(
      <TransportProfilesList profiles={[]} loading={false} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with profiles', () => {
    const { container } = render(
      <TransportProfilesList profiles={mockProfiles} loading={false} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should display empty message when no profiles', () => {
    render(
      <TransportProfilesList profiles={[]} loading={false} onRefresh={jest.fn()} />,
    );
    expect(screen.getByText(/No transport profiles configured/)).toBeInTheDocument();
  });

  it('should display profile details', () => {
    render(
      <TransportProfilesList profiles={mockProfiles} loading={false} onRefresh={jest.fn()} />,
    );

    expect(screen.getByText('Charon Freighter')).toBeInTheDocument();
    expect(screen.getByText('Freighter')).toBeInTheDocument();
    expect(screen.getByText('Rhea JF')).toBeInTheDocument();
    expect(screen.getByText('Jump Freighter')).toBeInTheDocument();
  });

  it('should show Default chip for default profiles', () => {
    render(
      <TransportProfilesList profiles={mockProfiles} loading={false} onRefresh={jest.fn()} />,
    );

    const defaultElements = screen.getAllByText('Default');
    expect(defaultElements.length).toBeGreaterThanOrEqual(2); // column header + chip
  });

  it('should open add dialog when Add Profile is clicked', () => {
    render(
      <TransportProfilesList profiles={mockProfiles} loading={false} onRefresh={jest.fn()} />,
    );

    fireEvent.click(screen.getByText('Add Profile'));
    expect(screen.getByTestId('profile-dialog')).toBeInTheDocument();
    expect(screen.getByText('Add Profile Dialog')).toBeInTheDocument();
  });

  it('should open edit dialog when edit button is clicked', () => {
    render(
      <TransportProfilesList profiles={mockProfiles} loading={false} onRefresh={jest.fn()} />,
    );

    const editButtons = screen.getAllByTestId('EditIcon');
    fireEvent.click(editButtons[0].closest('button')!);
    expect(screen.getByTestId('profile-dialog')).toBeInTheDocument();
    expect(screen.getByText('Edit Profile Dialog')).toBeInTheDocument();
  });
});

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { JFRoutesList } from '../JFRoutesList';
import { JFRoute } from '@industry-tool/pages/transport';

jest.mock('../JFRouteDialog', () => ({
  JFRouteDialog: function MockDialog({ open, route, onClose }: any) {
    if (!open) return null;
    return (
      <div data-testid="route-dialog">
        <span>{route ? 'Edit' : 'Add'} Route Dialog</span>
        <button onClick={() => onClose(false)}>Close</button>
      </div>
    );
  },
}));

const mockRoutes: JFRoute[] = [
  {
    id: 1,
    userId: 42,
    name: 'Jita to Amarr',
    originSystemId: 30000142,
    originSystemName: 'Jita',
    destinationSystemId: 30002187,
    destinationSystemName: 'Amarr',
    totalDistanceLy: 8.52,
    waypoints: [
      { id: 1, routeId: 1, sequence: 0, systemId: 30000142, systemName: 'Jita', distanceLy: 0 },
      { id: 2, routeId: 1, sequence: 1, systemId: 30003000, systemName: 'Ignoitton', distanceLy: 4.2 },
      { id: 3, routeId: 1, sequence: 2, systemId: 30002187, systemName: 'Amarr', distanceLy: 4.32 },
    ],
    createdAt: '2026-02-24T12:00:00Z',
  },
];

describe('JFRoutesList', () => {
  beforeEach(() => {
    (global.fetch as jest.Mock).mockClear();
  });

  it('should match snapshot while loading', () => {
    const { container } = render(
      <JFRoutesList routes={[]} loading={true} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with empty routes', () => {
    const { container } = render(
      <JFRoutesList routes={[]} loading={false} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot with routes', () => {
    const { container } = render(
      <JFRoutesList routes={mockRoutes} loading={false} onRefresh={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should display empty message when no routes', () => {
    render(
      <JFRoutesList routes={[]} loading={false} onRefresh={jest.fn()} />,
    );
    expect(screen.getByText(/No JF routes configured/)).toBeInTheDocument();
  });

  it('should display route details', () => {
    render(
      <JFRoutesList routes={mockRoutes} loading={false} onRefresh={jest.fn()} />,
    );

    expect(screen.getByText('Jita to Amarr')).toBeInTheDocument();
    expect(screen.getByText('8.52 LY')).toBeInTheDocument();
  });

  it('should display waypoint chips', () => {
    render(
      <JFRoutesList routes={mockRoutes} loading={false} onRefresh={jest.fn()} />,
    );

    expect(screen.getByText('Ignoitton')).toBeInTheDocument();
  });

  it('should open add dialog when Add JF Route is clicked', () => {
    render(
      <JFRoutesList routes={mockRoutes} loading={false} onRefresh={jest.fn()} />,
    );

    fireEvent.click(screen.getByText('Add JF Route'));
    expect(screen.getByTestId('route-dialog')).toBeInTheDocument();
    expect(screen.getByText('Add Route Dialog')).toBeInTheDocument();
  });

  it('should open edit dialog when edit button is clicked', () => {
    render(
      <JFRoutesList routes={mockRoutes} loading={false} onRefresh={jest.fn()} />,
    );

    const editButtons = screen.getAllByTestId('EditIcon');
    fireEvent.click(editButtons[0].closest('button')!);
    expect(screen.getByTestId('route-dialog')).toBeInTheDocument();
    expect(screen.getByText('Edit Route Dialog')).toBeInTheDocument();
  });
});

import React from 'react';
import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import StationDialog from '../StationDialog';
import { UserStation } from '@industry-tool/client/data/models';

const mockStation: UserStation = {
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
};

describe('StationDialog Component', () => {
  beforeEach(() => {
    (global.fetch as jest.Mock).mockClear();
  });

  it('should match snapshot when closed', () => {
    const { container } = render(
      <StationDialog open={false} station={null} onClose={jest.fn()} />,
    );
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot when open for adding', () => {
    const { baseElement } = render(
      <StationDialog open={true} station={null} onClose={jest.fn()} />,
    );
    expect(baseElement).toMatchSnapshot();
  });

  it('should match snapshot when open for editing', () => {
    const { baseElement } = render(
      <StationDialog open={true} station={mockStation} onClose={jest.fn()} />,
    );
    expect(baseElement).toMatchSnapshot();
  });

  it('should display Add Station title when creating', () => {
    render(
      <StationDialog open={true} station={null} onClose={jest.fn()} />,
    );
    expect(screen.getByText('Add Station')).toBeInTheDocument();
  });

  it('should display Edit Station title when editing', () => {
    render(
      <StationDialog open={true} station={mockStation} onClose={jest.fn()} />,
    );
    expect(screen.getByText('Edit Station')).toBeInTheDocument();
  });

  it('should populate fields when editing', () => {
    render(
      <StationDialog open={true} station={mockStation} onClose={jest.fn()} />,
    );

    // Structure should be pre-filled
    expect(screen.getByText('Sotiyo')).toBeInTheDocument();

    // Rig names should be shown as title attributes on rig rows
    expect(screen.getByTitle('Standup XL-Set Ship Manufacturing Efficiency I')).toBeInTheDocument();
    expect(screen.getByTitle('Standup XL-Set Structure and Component Manufacturing Efficiency I')).toBeInTheDocument();

    // Services should be displayed
    expect(screen.getByText('Standup Manufacturing Plant I')).toBeInTheDocument();
  });

  it('should disable station search when editing', () => {
    render(
      <StationDialog open={true} station={mockStation} onClose={jest.fn()} />,
    );

    const stationInput = screen.getByLabelText('Station');
    expect(stationInput).toBeDisabled();
  });

  it('should call onClose(false) when Cancel is clicked', () => {
    const onClose = jest.fn();
    render(
      <StationDialog open={true} station={null} onClose={onClose} />,
    );

    fireEvent.click(screen.getByText('Cancel'));
    expect(onClose).toHaveBeenCalledWith(false);
  });

  it('should disable Add button when no station is selected', () => {
    render(
      <StationDialog open={true} station={null} onClose={jest.fn()} />,
    );

    const addButton = screen.getByText('Add');
    expect(addButton.closest('button')).toBeDisabled();
  });

  it('should parse scan when Parse Scan button is clicked', async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        structure: 'sotiyo',
        rigs: [
          { name: 'Standup XL-Set Ship Manufacturing Efficiency I', category: 'ship', tier: 't1' },
        ],
        services: [
          { name: 'Standup Manufacturing Plant I', activity: 'manufacturing' },
        ],
      }),
    });

    render(
      <StationDialog open={true} station={null} onClose={jest.fn()} />,
    );

    const scanInput = screen.getByLabelText('Structure Fitting Scan');
    fireEvent.change(scanInput, {
      target: { value: 'Rig Slots\nStandup XL-Set Ship Manufacturing Efficiency I\nService Slots\nStandup Manufacturing Plant I' },
    });

    await act(async () => {
      fireEvent.click(screen.getByText('Parse Scan'));
    });

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/stations/parse-scan',
        expect.objectContaining({ method: 'POST' }),
      );
    });
  });

  it('should disable Parse Scan button when scan text is empty', () => {
    render(
      <StationDialog open={true} station={null} onClose={jest.fn()} />,
    );

    const parseButton = screen.getByText('Parse Scan');
    expect(parseButton.closest('button')).toBeDisabled();
  });

  it('should add manual rig when Add Rig is clicked', () => {
    render(
      <StationDialog open={true} station={null} onClose={jest.fn()} />,
    );

    expect(screen.getByText('No rigs. Paste a scan or add manually.')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Add Rig'));

    // Should now have a rig row with category and tier selects
    expect(screen.queryByText('No rigs. Paste a scan or add manually.')).not.toBeInTheDocument();
  });
});

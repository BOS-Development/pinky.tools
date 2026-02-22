import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import AddJob from '../AddJob';

// Mock fetch
beforeEach(() => {
  (global.fetch as jest.Mock).mockClear();
});

describe('AddJob Component', () => {
  const mockOnJobAdded = jest.fn();

  beforeEach(() => {
    mockOnJobAdded.mockClear();
    // Mock systems endpoint
    (global.fetch as jest.Mock).mockImplementation((url: string) => {
      if (url === '/api/industry/systems') {
        return Promise.resolve({
          ok: true,
          json: async () => [
            { system_id: 30000142, name: 'Jita', security_status: 0.9, cost_index: 0.05 },
          ],
        });
      }
      return Promise.resolve({ ok: true, json: async () => [] });
    });
  });

  it('should match snapshot', async () => {
    const { container } = render(<AddJob onJobAdded={mockOnJobAdded} />);
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith('/api/industry/systems');
    });
    expect(container).toMatchSnapshot();
  });

  it('should render the blueprint search field', () => {
    render(<AddJob onJobAdded={mockOnJobAdded} />);
    expect(screen.getByLabelText('Search Blueprint')).toBeInTheDocument();
  });

  it('should render activity selector', () => {
    render(<AddJob onJobAdded={mockOnJobAdded} />);
    // MUI Select renders 'Activity' as label and 'Manufacturing' as default value
    expect(screen.getByText('Manufacturing')).toBeInTheDocument();
  });

  it('should render runs field', () => {
    render(<AddJob onJobAdded={mockOnJobAdded} />);
    expect(screen.getByLabelText('Runs')).toBeInTheDocument();
  });

  it('should render ME and TE fields', () => {
    render(<AddJob onJobAdded={mockOnJobAdded} />);
    expect(screen.getByLabelText('ME Level')).toBeInTheDocument();
    expect(screen.getByLabelText('TE Level')).toBeInTheDocument();
  });

  it('should render Add to Queue button', () => {
    render(<AddJob onJobAdded={mockOnJobAdded} />);
    expect(screen.getByText('Add to Queue')).toBeInTheDocument();
  });

  it('should have Add to Queue button disabled by default', () => {
    render(<AddJob onJobAdded={mockOnJobAdded} />);
    const button = screen.getByText('Add to Queue').closest('button');
    expect(button).toBeDisabled();
  });
});

import React from 'react';
import { render } from '@testing-library/react';
import Navbar from '../Navbar';

describe('Navbar Component', () => {
  it('should match snapshot', () => {
    const { container } = render(<Navbar />);
    expect(container).toMatchSnapshot();
  });

  it('should display app title', () => {
    const { getByText } = render(<Navbar />);
    expect(getByText('EVE Industry Tool')).toBeInTheDocument();
  });

  it('should have navigation links', () => {
    const { getByRole } = render(<Navbar />);

    const charactersLink = getByRole('link', { name: /characters/i });
    const corporationsLink = getByRole('link', { name: /corporations/i });
    const inventoryLink = getByRole('link', { name: /inventory/i });
    const stockpilesLink = getByRole('link', { name: /stockpiles/i });

    expect(charactersLink).toHaveAttribute('href', '/characters');
    expect(corporationsLink).toHaveAttribute('href', '/corporations');
    expect(inventoryLink).toHaveAttribute('href', '/inventory');
    expect(stockpilesLink).toHaveAttribute('href', '/stockpiles');
  });

  it('should have rocket icon', () => {
    const { getByLabelText } = render(<Navbar />);
    const menuButton = getByLabelText('menu');
    expect(menuButton).toBeInTheDocument();
  });
});

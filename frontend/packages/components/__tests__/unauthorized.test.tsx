import React from 'react';
import { render } from '@testing-library/react';
import Unauthorized from '../unauthorized';

describe('Unauthorized Component', () => {
  it('should match snapshot', () => {
    const { container } = render(<Unauthorized />);
    expect(container).toMatchSnapshot();
  });

  it('should display authentication required message', () => {
    const { getByText } = render(<Unauthorized />);
    expect(getByText('Authentication Required')).toBeInTheDocument();
    expect(getByText('You must sign in to access this page')).toBeInTheDocument();
  });

  it('should have sign in link with correct href', () => {
    const { getByRole } = render(<Unauthorized />);
    const signInLink = getByRole('link', { name: /sign in/i });
    expect(signInLink).toBeInTheDocument();
    expect(signInLink).toHaveAttribute('href', 'api/auth/signin');
  });
});

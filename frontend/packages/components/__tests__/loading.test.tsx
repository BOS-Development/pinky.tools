import React from 'react';
import { render } from '@testing-library/react';
import Loading from '../loading';

describe('Loading Component', () => {
  it('should match snapshot', () => {
    const { container } = render(<Loading />);
    expect(container).toMatchSnapshot();
  });

  it('should display loading text', () => {
    const { getByText } = render(<Loading />);
    expect(getByText('Loading...')).toBeInTheDocument();
  });
});

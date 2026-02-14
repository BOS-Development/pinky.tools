import React from 'react';
import { render } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import Inventory from '../inventory';

// Mock the AssetsList component to avoid complex dependencies
jest.mock('@industry-tool/components/assets/AssetsList', () => {
  return function MockAssetsList() {
    return <div data-testid="assets-list">Assets List Component</div>;
  };
});

// Mock other components
jest.mock('@industry-tool/components/loading', () => {
  return function MockLoading() {
    return <div data-testid="loading">Loading...</div>;
  };
});

jest.mock('@industry-tool/components/unauthorized', () => {
  return function MockUnauthorized() {
    return <div data-testid="unauthorized">Unauthorized</div>;
  };
});

describe('Inventory Page', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should match snapshot when loading', () => {
    (useSession as jest.Mock).mockReturnValue({
      data: null,
      status: 'loading',
    });

    const { container } = render(<Inventory />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot when authenticated', () => {
    (useSession as jest.Mock).mockReturnValue({
      data: { user: { name: 'Test User' } },
      status: 'authenticated',
    });

    const { container } = render(<Inventory />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot when unauthenticated', () => {
    (useSession as jest.Mock).mockReturnValue({
      data: null,
      status: 'unauthenticated',
    });

    const { container } = render(<Inventory />);
    expect(container).toMatchSnapshot();
  });

  it('should show loading component when session is loading', () => {
    (useSession as jest.Mock).mockReturnValue({
      data: null,
      status: 'loading',
    });

    const { getByTestId } = render(<Inventory />);
    expect(getByTestId('loading')).toBeInTheDocument();
  });

  it('should show unauthorized component when not authenticated', () => {
    (useSession as jest.Mock).mockReturnValue({
      data: null,
      status: 'unauthenticated',
    });

    const { getByTestId } = render(<Inventory />);
    expect(getByTestId('unauthorized')).toBeInTheDocument();
  });

  it('should show assets list when authenticated', () => {
    (useSession as jest.Mock).mockReturnValue({
      data: { user: { name: 'Test User' } },
      status: 'authenticated',
    });

    const { getByTestId } = render(<Inventory />);
    expect(getByTestId('assets-list')).toBeInTheDocument();
  });
});

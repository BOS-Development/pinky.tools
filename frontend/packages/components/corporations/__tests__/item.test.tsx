import React from 'react';
import { render, screen } from '@testing-library/react';
import Item from '../item';

describe('Corporation Item Component', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-02-22T12:00:00Z'));
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  const upToDateCorporation = {
    id: 98765,
    name: 'Test Corporation',
    esiScopes: [
      'esi-wallet.read_corporation_wallets.v1',
      'esi-assets.read_corporation_assets.v1',
      'esi-corporations.read_blueprints.v1',
      'esi-corporations.read_starbases.v1',
      'esi-industry.read_corporation_jobs.v1',
      'esi-markets.read_corporation_orders.v1',
      'esi-contracts.read_corporation_contracts.v1',
      'esi-corporations.read_container_logs.v1',
      'esi-industry.read_corporation_mining.v1',
      'esi-corporations.read_facilities.v1',
      'esi-corporations.read_divisions.v1',
      'esi-universe.read_structures.v1',
    ].join(' '),
  };

  const outdatedCorporation = {
    id: 54321,
    name: 'Outdated Corp',
    esiScopes: 'esi-assets.read_corporation_assets.v1',
  };

  const noScopesCorporation = {
    id: 11111,
    name: 'No Scopes Corp',
    esiScopes: '',
  };

  it('should match snapshot for up-to-date corporation', () => {
    const { container } = render(<Item corporation={upToDateCorporation} />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot for outdated corporation', () => {
    const { container } = render(<Item corporation={outdatedCorporation} />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot for corporation with no scopes', () => {
    const { container } = render(<Item corporation={noScopesCorporation} />);
    expect(container).toMatchSnapshot();
  });

  it('should display corporation name', () => {
    render(<Item corporation={upToDateCorporation} />);
    expect(screen.getByText('Test Corporation')).toBeInTheDocument();
  });

  it('should display corporation logo image', () => {
    render(<Item corporation={upToDateCorporation} />);
    const img = screen.getByRole('img', { name: 'Test Corporation' });
    expect(img).toHaveAttribute(
      'src',
      'https://images.evetech.net/corporations/98765/logo?size=256&tenant=tranquility'
    );
  });

  it('should display Corporation chip', () => {
    render(<Item corporation={upToDateCorporation} />);
    expect(screen.getByText('Corporation')).toBeInTheDocument();
  });

  it('should not show warning for up-to-date corporation', () => {
    render(<Item corporation={upToDateCorporation} />);
    expect(screen.queryByText('Re-authorize')).not.toBeInTheDocument();
  });

  it('should show warning for outdated corporation', () => {
    render(<Item corporation={outdatedCorporation} />);
    expect(screen.getByText('Re-authorize')).toBeInTheDocument();
  });

  it('should show warning for corporation with no scopes', () => {
    render(<Item corporation={noScopesCorporation} />);
    expect(screen.getByText('Re-authorize')).toBeInTheDocument();
  });

  it('should link re-authorize button to /api/corporations/add', () => {
    render(<Item corporation={outdatedCorporation} />);
    const button = screen.getByRole('link', { name: 'Re-authorize' });
    expect(button).toHaveAttribute('href', '/api/corporations/add');
  });
});

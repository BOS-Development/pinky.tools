import React from 'react';
import { render, screen } from '@testing-library/react';
import Item from '../item';

describe('Character Item Component', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-02-22T12:00:00Z'));
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  const upToDateCharacter = {
    id: 12345,
    name: 'Test Pilot',
    esiScopes: [
      'publicData',
      'esi-skills.read_skills.v1',
      'esi-skills.read_skillqueue.v1',
      'esi-wallet.read_character_wallet.v1',
      'esi-search.search_structures.v1',
      'esi-clones.read_clones.v1',
      'esi-universe.read_structures.v1',
      'esi-assets.read_assets.v1',
      'esi-planets.manage_planets.v1',
      'esi-markets.structure_markets.v1',
      'esi-industry.read_character_jobs.v1',
      'esi-markets.read_character_orders.v1',
      'esi-characters.read_blueprints.v1',
      'esi-contracts.read_character_contracts.v1',
      'esi-clones.read_implants.v1',
      'esi-industry.read_character_mining.v1',
    ].join(' '),
  };

  const outdatedCharacter = {
    id: 67890,
    name: 'Outdated Pilot',
    esiScopes: 'publicData esi-skills.read_skills.v1',
  };

  const noScopesCharacter = {
    id: 11111,
    name: 'No Scopes Pilot',
    esiScopes: '',
  };

  it('should match snapshot for up-to-date character', () => {
    const { container } = render(<Item character={upToDateCharacter} />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot for outdated character', () => {
    const { container } = render(<Item character={outdatedCharacter} />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot for character with no scopes', () => {
    const { container } = render(<Item character={noScopesCharacter} />);
    expect(container).toMatchSnapshot();
  });

  it('should display character name', () => {
    render(<Item character={upToDateCharacter} />);
    expect(screen.getByText('Test Pilot')).toBeInTheDocument();
  });

  it('should display character portrait image', () => {
    render(<Item character={upToDateCharacter} />);
    const img = screen.getByRole('img', { name: 'Test Pilot' });
    expect(img).toHaveAttribute(
      'src',
      'https://image.eveonline.com/Character/12345_128.jpg'
    );
  });

  it('should not show warning for up-to-date character', () => {
    render(<Item character={upToDateCharacter} />);
    expect(screen.queryByText('Re-authorize')).not.toBeInTheDocument();
  });

  it('should show warning for outdated character', () => {
    render(<Item character={outdatedCharacter} />);
    expect(screen.getByText('Re-authorize')).toBeInTheDocument();
  });

  it('should show warning for character with no scopes', () => {
    render(<Item character={noScopesCharacter} />);
    expect(screen.getByText('Re-authorize')).toBeInTheDocument();
  });

  it('should link re-authorize button to /api/characters/add', () => {
    render(<Item character={outdatedCharacter} />);
    const button = screen.getByRole('link', { name: 'Re-authorize' });
    expect(button).toHaveAttribute('href', '/api/characters/add');
  });
});

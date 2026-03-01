import React from 'react';
import { render, screen } from '@testing-library/react';
import List from '../list';

jest.mock('@industry-tool/components/Navbar', () => {
  return function MockNavbar() {
    return <div data-testid="navbar">Navbar</div>;
  };
});

// Mock the Item component to keep list snapshots stable
jest.mock('../item', () => {
  return function MockItem({ character }: { character: { id: number; name: string } }) {
    return <div data-testid={`character-item-${character.id}`}>{character.name}</div>;
  };
});

describe('Character List Component', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2026-02-22T12:00:00Z'));
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  const normalCharacter = {
    id: 12345,
    name: 'Test Pilot',
    esiScopes: 'publicData esi-assets.read_assets.v1',
  };

  const reauthCharacter = {
    id: 22222,
    name: 'Revoked Pilot',
    esiScopes: 'publicData',
    needsReauth: true,
  };

  it('should match snapshot for empty character list', () => {
    const { container } = render(<List characters={[]} />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot for list with normal characters', () => {
    const { container } = render(<List characters={[normalCharacter]} />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot for list with reauth character', () => {
    const { container } = render(<List characters={[reauthCharacter]} />);
    expect(container).toMatchSnapshot();
  });

  it('should match snapshot for list with mixed characters', () => {
    const { container } = render(<List characters={[normalCharacter, reauthCharacter]} />);
    expect(container).toMatchSnapshot();
  });

  it('should show empty state when no characters', () => {
    render(<List characters={[]} />);
    expect(screen.getByText('No Characters')).toBeInTheDocument();
    expect(screen.getByText('Get started by adding your first character')).toBeInTheDocument();
  });

  it('should show Characters heading when characters exist', () => {
    render(<List characters={[normalCharacter]} />);
    expect(screen.getByText('Characters')).toBeInTheDocument();
  });

  it('should render character items', () => {
    render(<List characters={[normalCharacter]} />);
    expect(screen.getByTestId('character-item-12345')).toBeInTheDocument();
  });

  it('should show error alert for character needing reauth', () => {
    render(<List characters={[reauthCharacter]} />);
    expect(screen.getByText(/ESI authorization for/)).toBeInTheDocument();
    expect(screen.getAllByText('Revoked Pilot').length).toBeGreaterThan(0);
    expect(screen.getByText(/has been revoked/)).toBeInTheDocument();
  });

  it('should show Re-authorize button in alert for reauth character', () => {
    render(<List characters={[reauthCharacter]} />);
    // The alert contains a Re-authorize button linking to /api/characters/add
    const button = screen.getByRole('link', { name: 'Re-authorize' });
    expect(button).toHaveAttribute('href', '/api/characters/add');
  });

  it('should not show reauth alert for normal character', () => {
    render(<List characters={[normalCharacter]} />);
    expect(screen.queryByText(/ESI authorization for/)).not.toBeInTheDocument();
  });

  it('should show one alert per reauth character', () => {
    const secondReauth = { id: 33333, name: 'Another Revoked', esiScopes: '', needsReauth: true };
    render(<List characters={[reauthCharacter, secondReauth]} />);
    const alerts = screen.getAllByText(/ESI authorization for/);
    expect(alerts).toHaveLength(2);
  });
});

import { render } from '@testing-library/react';
import { useSession } from 'next-auth/react';
import AssetsList from '../AssetsList';
import { AssetsResponse } from '@industry-tool/client/data/models';

// Mock Navbar component
jest.mock('@industry-tool/components/Navbar', () => {
  return function MockNavbar() {
    return <div data-testid="navbar">Navbar</div>;
  };
});

describe('AssetsList Component', () => {
  const mockSession = {
    data: {
      user: { name: 'Test User' },
      providerAccountId: '123456',
    },
    status: 'authenticated',
  };

  beforeEach(() => {
    jest.clearAllMocks();
    (useSession as jest.Mock).mockReturnValue(mockSession);
    Storage.prototype.getItem = jest.fn(() => null);
    Storage.prototype.setItem = jest.fn();
  });

  describe('Loading State', () => {
    it('should match snapshot when loading', () => {
      // Mock unauthenticated to avoid fetch attempt
      (useSession as jest.Mock).mockReturnValue({ data: null, status: 'unauthenticated' });
      const { container } = render(<AssetsList />);
      expect(container).toMatchSnapshot();
    });

    it('should display loading message when no assets provided', () => {
      (useSession as jest.Mock).mockReturnValue({ data: null, status: 'unauthenticated' });
      const { getByText } = render(<AssetsList />);
      expect(getByText('Loading assets...')).toBeInTheDocument();
    });
  });

  describe('Empty State', () => {
    it('should match snapshot when no assets exist', () => {
      const emptyAssets: AssetsResponse = {
        structures: [],
      };

      const { container } = render(<AssetsList assets={emptyAssets} />);
      expect(container).toMatchSnapshot();
    });

    it('should display no assets message', () => {
      const emptyAssets: AssetsResponse = {
        structures: [],
      };

      const { getByText } = render(<AssetsList assets={emptyAssets} />);
      expect(getByText('No Assets Found')).toBeInTheDocument();
      expect(getByText("You don't have any assets yet, or they haven't been synced.")).toBeInTheDocument();
    });

    it('should handle null structures without crashing', () => {
      const nullAssets = {
        structures: null,
      } as any;

      const { getByText } = render(<AssetsList assets={nullAssets} />);
      expect(getByText('No Assets Found')).toBeInTheDocument();
    });

    it('should handle undefined structures without crashing', () => {
      const undefinedAssets = {} as any;

      const { getByText } = render(<AssetsList assets={undefinedAssets} />);
      expect(getByText('No Assets Found')).toBeInTheDocument();
    });
  });

  describe('With Basic Assets', () => {
    const mockAssets: AssetsResponse = {
      structures: [
        {
          id: 60003760,
          name: 'Jita IV - Moon 4 - Caldari Navy Assembly Plant',
          hangarAssets: [
            {
              name: 'Tritanium',
              typeId: 34,
              quantity: 1000,
              volume: 10.0,
              ownerType: 'character',
              ownerName: 'Test Character',
              ownerId: 12345,
            },
            {
              name: 'Pyerite',
              typeId: 35,
              quantity: 500,
              volume: 8.0,
              ownerType: 'character',
              ownerName: 'Test Character',
              ownerId: 12345,
            },
          ],
          solarSystem: 'Test System',
          region: 'Test Region',
          hangarContainers: [],
          deliveries: [],
          assetSafety: [],
          corporationHangers: [],
        },
      ],
    };

    it('should match snapshot with basic assets', () => {
      const { container } = render(<AssetsList assets={mockAssets} />);
      expect(container).toMatchSnapshot();
    });

    it('should render Asset Inventory header', () => {
      const { getByText } = render(<AssetsList assets={mockAssets} />);
      expect(getByText('Asset Inventory')).toBeInTheDocument();
    });

    it('should display structure name in list', () => {
      const { getByText } = render(<AssetsList assets={mockAssets} />);
      expect(getByText('Jita IV - Moon 4 - Caldari Navy Assembly Plant')).toBeInTheDocument();
    });
  });

  describe('With Containers', () => {
    const mockAssetsWithContainers: AssetsResponse = {
      structures: [
        {
          id: 60003760,
          name: 'Jita IV - Moon 4 - Caldari Navy Assembly Plant',
          solarSystem: 'Test System',
          region: 'Test Region',
          hangarAssets: [],
          hangarContainers: [
            {
              id: 1001,
              name: 'Materials Container',
              ownerType: 'character',
              ownerName: 'Test Character',
              ownerId: 12345,
              assets: [
                {
                  name: 'Mexallon',
                  typeId: 36,
                  quantity: 2000,
                  volume: 20.0,
                  ownerType: 'character',
                  ownerName: 'Test Character',
                  ownerId: 12345,
                },
              ],
            },
          ],
          deliveries: [],
          assetSafety: [],
          corporationHangers: [],
        },
      ],
    };

    it('should match snapshot with containers', () => {
      const { container } = render(<AssetsList assets={mockAssetsWithContainers} />);
      expect(container).toMatchSnapshot();
    });
  });

  describe('With Corporation Hangars', () => {
    const mockAssetsWithCorpHangars: AssetsResponse = {
      structures: [
        {
          id: 60003760,
          name: 'Jita IV - Moon 4 - Caldari Navy Assembly Plant',
          solarSystem: 'Test System',
          region: 'Test Region',
          hangarAssets: [],
          hangarContainers: [],
          deliveries: [],
          assetSafety: [],
          corporationHangers: [
            {
              id: 1,
              name: 'Hangar 1',
              corporationId: 98765,
              corporationName: 'Test Corp',
              assets: [
                {
                  name: 'Isogen',
                  typeId: 37,
                  quantity: 3000,
                  volume: 30.0,
                  ownerType: 'corporation',
                  ownerName: 'Test Corp',
                  ownerId: 98765,
                },
              ],
              hangarContainers: [],
            },
          ],
        },
      ],
    };

    it('should match snapshot with corporation hangars', () => {
      const { container } = render(<AssetsList assets={mockAssetsWithCorpHangars} />);
      expect(container).toMatchSnapshot();
    });
  });

  describe('With Stockpile Markers', () => {
    const mockAssetsWithStockpiles: AssetsResponse = {
      structures: [
        {
          id: 60003760,
          name: 'Jita IV - Moon 4 - Caldari Navy Assembly Plant',
          hangarAssets: [
            {
              name: 'Nocxium',
              typeId: 38,
              quantity: 800,
              volume: 16.0,
              ownerType: 'character',
              ownerName: 'Test Character',
              ownerId: 12345,
              desiredQuantity: 1000,
              stockpileDelta: -200,
            },
            {
              name: 'Zydrine',
              typeId: 39,
              quantity: 1500,
              volume: 30.0,
              ownerType: 'character',
              ownerName: 'Test Character',
              ownerId: 12345,
              desiredQuantity: 1000,
              stockpileDelta: 500,
            },
          ],
          solarSystem: 'Test System',
          region: 'Test Region',
          hangarContainers: [],
          deliveries: [],
          assetSafety: [],
          corporationHangers: [],
        },
      ],
    };

    it('should match snapshot with stockpile markers', () => {
      const { container } = render(<AssetsList assets={mockAssetsWithStockpiles} />);
      expect(container).toMatchSnapshot();
    });

    it('should render Below target only switch', () => {
      const { getByText } = render(<AssetsList assets={mockAssetsWithStockpiles} />);
      expect(getByText('Below target only')).toBeInTheDocument();
    });
  });

  describe('Multiple Structures', () => {
    const mockMultipleStructures: AssetsResponse = {
      structures: [
        {
          id: 60003760,
          name: 'Jita IV - Moon 4',
          hangarAssets: [
            {
              name: 'Item A',
              typeId: 1,
              quantity: 100,
              volume: 10.0,
              ownerType: 'character',
              ownerName: 'Test',
              ownerId: 1,
            },
          ],
          solarSystem: 'Test System',
          region: 'Test Region',
          hangarContainers: [],
          deliveries: [],
          assetSafety: [],
          corporationHangers: [],
        },
        {
          id: 60003761,
          name: 'Amarr VIII - Station',
          hangarAssets: [
            {
              name: 'Item B',
              typeId: 2,
              quantity: 200,
              volume: 20.0,
              ownerType: 'character',
              ownerName: 'Test',
              ownerId: 1,
            },
          ],
          solarSystem: 'Test System',
          region: 'Test Region',
          hangarContainers: [],
          deliveries: [],
          assetSafety: [],
          corporationHangers: [],
        },
      ],
    };

    it('should match snapshot with multiple structures', () => {
      const { container } = render(<AssetsList assets={mockMultipleStructures} />);
      expect(container).toMatchSnapshot();
    });

    it('should display both structure names', () => {
      const { container } = render(<AssetsList assets={mockMultipleStructures} />);
      expect(container.textContent).toContain('Jita IV - Moon 4');
      expect(container.textContent).toContain('Amarr VIII - Station');
    });
  });
});

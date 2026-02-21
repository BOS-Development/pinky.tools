import { aggregateAssetsByTypeId, getUniqueOwners } from '../assetAggregation';
import { AssetStructure } from '@industry-tool/client/data/models';

const makeAsset = (typeId: number, quantity: number) => ({
  name: `Item ${typeId}`,
  typeId,
  quantity,
  volume: quantity * 0.01,
  ownerType: 'character',
  ownerName: 'Test Char',
  ownerId: 1,
});

const emptyStructure: AssetStructure = {
  id: 1000,
  name: 'Test Station',
  solarSystem: 'Jita',
  region: 'The Forge',
  hangarAssets: [],
  hangarContainers: [],
  deliveries: [],
  assetSafety: [],
  corporationHangers: [],
};

describe('aggregateAssetsByTypeId', () => {
  it('returns empty map for empty structure', () => {
    const result = aggregateAssetsByTypeId(emptyStructure);
    expect(result.size).toBe(0);
  });

  it('aggregates hangar assets', () => {
    const structure: AssetStructure = {
      ...emptyStructure,
      hangarAssets: [makeAsset(100, 500), makeAsset(200, 300)],
    };
    const result = aggregateAssetsByTypeId(structure);
    expect(result.get(100)).toBe(500);
    expect(result.get(200)).toBe(300);
  });

  it('aggregates container assets', () => {
    const structure: AssetStructure = {
      ...emptyStructure,
      hangarContainers: [
        { id: 1, name: 'Container 1', ownerType: 'character', ownerName: 'Test', ownerId: 1, assets: [makeAsset(100, 200)] },
        { id: 2, name: 'Container 2', ownerType: 'character', ownerName: 'Test', ownerId: 1, assets: [makeAsset(100, 100)] },
      ],
    };
    const result = aggregateAssetsByTypeId(structure);
    expect(result.get(100)).toBe(300);
  });

  it('aggregates corporation hangar assets and containers', () => {
    const structure: AssetStructure = {
      ...emptyStructure,
      corporationHangers: [
        {
          id: 1,
          name: 'Division 1',
          corporationId: 99,
          corporationName: 'Test Corp',
          assets: [makeAsset(100, 1000)],
          hangarContainers: [
            { id: 10, name: 'Corp Container', ownerType: 'corporation', ownerName: 'Test Corp', ownerId: 99, assets: [makeAsset(100, 500)] },
          ],
        },
      ],
    };
    const result = aggregateAssetsByTypeId(structure);
    expect(result.get(100)).toBe(1500);
  });

  it('sums quantities across all sources for same typeId', () => {
    const structure: AssetStructure = {
      ...emptyStructure,
      hangarAssets: [makeAsset(100, 100)],
      hangarContainers: [
        { id: 1, name: 'C1', ownerType: 'character', ownerName: 'Test', ownerId: 1, assets: [makeAsset(100, 200)] },
      ],
      deliveries: [makeAsset(100, 50)],
      assetSafety: [makeAsset(100, 25)],
      corporationHangers: [
        {
          id: 1,
          name: 'Div 1',
          corporationId: 99,
          corporationName: 'Corp',
          assets: [makeAsset(100, 300)],
          hangarContainers: [
            { id: 10, name: 'CC', ownerType: 'corporation', ownerName: 'Corp', ownerId: 99, assets: [makeAsset(100, 125)] },
          ],
        },
      ],
    };
    const result = aggregateAssetsByTypeId(structure);
    expect(result.get(100)).toBe(800); // 100 + 200 + 50 + 25 + 300 + 125
  });
});

describe('getUniqueOwners', () => {
  it('returns empty array for empty structure', () => {
    const result = getUniqueOwners(emptyStructure);
    expect(result).toEqual([]);
  });

  it('returns unique owners from hangar and corp assets', () => {
    const structure: AssetStructure = {
      ...emptyStructure,
      hangarAssets: [makeAsset(100, 500)],
      corporationHangers: [
        {
          id: 1,
          name: 'Div 1',
          corporationId: 99,
          corporationName: 'Test Corp',
          assets: [{ ...makeAsset(200, 300), ownerType: 'corporation', ownerName: 'Test Corp', ownerId: 99 }],
          hangarContainers: [],
        },
      ],
    };
    const result = getUniqueOwners(structure);
    expect(result).toHaveLength(2);
    expect(result).toContainEqual({ ownerType: 'character', ownerId: 1, ownerName: 'Test Char' });
    expect(result).toContainEqual({ ownerType: 'corporation', ownerId: 99, ownerName: 'Test Corp' });
  });

  it('deduplicates same owner across sources', () => {
    const structure: AssetStructure = {
      ...emptyStructure,
      hangarAssets: [makeAsset(100, 500), makeAsset(200, 300)],
      hangarContainers: [
        { id: 1, name: 'C1', ownerType: 'character', ownerName: 'Test Char', ownerId: 1, assets: [makeAsset(300, 100)] },
      ],
    };
    const result = getUniqueOwners(structure);
    expect(result).toHaveLength(1);
    expect(result[0]).toEqual({ ownerType: 'character', ownerId: 1, ownerName: 'Test Char' });
  });
});

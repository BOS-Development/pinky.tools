export type Character = {
  id: number;
  name: string;
};

export type Corporation = {
  id: number;
  name: string;
};

export type Asset = {
  name: string;
  typeId: number;
  quantity: number;
  volume: number;
  ownerType: string;
  ownerName: string;
  ownerId: number;
  desiredQuantity?: number;
  stockpileDelta?: number;
  unitPrice?: number;
  totalValue?: number;
  deficitValue?: number;
};

export type AssetContainer = {
  id: number;
  name: string;
  ownerType: string;
  ownerName: string;
  ownerId: number;
  assets: Asset[];
};

export type CorporationHanger = {
  id: number;
  name: string;
  corporationId: number;
  corporationName: string;
  assets: Asset[];
  hangarContainers: AssetContainer[];
};

export type AssetStructure = {
  id: number;
  name: string;
  solarSystem: string;
  region: string;
  hangarAssets: Asset[];
  hangarContainers: AssetContainer[];
  deliveries: Asset[];
  assetSafety: Asset[];
  corporationHangers: CorporationHanger[];
};

export type AssetsResponse = {
  structures: AssetStructure[];
};

export type StockpileMarker = {
  userId: number;
  typeId: number;
  ownerType: string;
  ownerId: number;
  locationId: number;
  containerId?: number;
  divisionNumber?: number;
  desiredQuantity: number;
  notes?: string;
};

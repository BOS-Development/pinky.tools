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

// Item Type Search Result (Go struct has no json tags -> PascalCase)
export type EveInventoryType = {
  TypeID: number;
  TypeName: string;
  Volume: number;
  IconID: number | null;
};

// Reactions Calculator Types

export type ReactionMaterial = {
  type_id: number;
  name: string;
  base_qty: number;
  adj_qty: number;
  price: number;
  cost: number;
  volume: number;
  is_intermediate: boolean;
};

export type Reaction = {
  reaction_type_id: number;
  product_type_id: number;
  product_name: string;
  group_name: string;
  product_qty_per_run: number;
  runs_per_cycle: number;
  secs_per_run: number;
  complex_instances: number;
  num_intermediates: number;
  input_cost_per_run: number;
  job_cost_per_run: number;
  output_value_per_run: number;
  output_fees_per_run: number;
  shipping_in_per_run: number;
  shipping_out_per_run: number;
  profit_per_run: number;
  profit_per_cycle: number;
  margin: number;
  materials: ReactionMaterial[];
};

export type ReactionsResponse = {
  reactions: Reaction[];
  count: number;
  cost_index: number;
  me_factor: number;
  te_factor: number;
  runs_per_cycle: number;
};

export type ReactionSystem = {
  system_id: number;
  name: string;
  security_status: number;
  cost_index: number;
};

export type PlanSelection = {
  reaction_type_id: number;
  instances: number;
};

export type IntermediatePlan = {
  type_id: number;
  name: string;
  slots: number;
  runs: number;
  produced: number;
};

export type ShoppingItem = {
  type_id: number;
  name: string;
  quantity: number;
  price: number;
  cost: number;
  volume: number;
};

export type PlanSummary = {
  total_slots: number;
  intermediate_slots: number;
  complex_slots: number;
  investment: number;
  revenue: number;
  profit: number;
  margin: number;
};

export type PlanResponse = {
  intermediates: IntermediatePlan[];
  shopping_list: ShoppingItem[];
  summary: PlanSummary;
};

// Planetary Industry Types

export type PiPinContent = {
  typeId: number;
  name: string;
  amount: number;
};

export type PiExtractor = {
  pinId: number;
  typeId: number;
  productTypeId: number;
  productName: string;
  qtyPerCycle: number;
  cycleTimeSec: number;
  ratePerHour: number;
  expiryTime: string | null;
  status: string;
  numHeads: number;
};

export type PiFactory = {
  pinId: number;
  typeId: number;
  schematicId: number;
  schematicName: string;
  outputTypeId: number;
  outputName: string;
  outputQty: number;
  cycleTimeSec: number;
  ratePerHour: number;
  lastCycleStart: string | null;
  status: string;
  pinCategory: string;
};

export type PiLaunchpad = {
  pinId: number;
  typeId: number;
  label?: string;
  contents: PiPinContent[];
};

// Launchpad Detail Types

export type LaunchpadInputRequirement = {
  typeId: number;
  name: string;
  qtyPerCycle: number;
  cyclesPerHour: number;
  consumedPerHour: number;
  currentStock: number;
  depletionHours: number;
};

export type LaunchpadConnectedFactory = {
  pinId: number;
  schematicName: string;
  outputName: string;
  outputTypeId: number;
  cycleTimeSec: number;
  inputs: LaunchpadInputRequirement[];
};

export type LaunchpadDetailResponse = {
  pinId: number;
  characterId: number;
  planetId: number;
  label?: string;
  contents: PiPinContent[];
  factories: LaunchpadConnectedFactory[];
};

export type PiPlanet = {
  planetId: number;
  planetType: string;
  solarSystemId: number;
  solarSystemName: string;
  characterId: number;
  characterName: string;
  upgradeLevel: number;
  numPins: number;
  lastUpdate: string;
  status: string;
  extractors: PiExtractor[];
  factories: PiFactory[];
  launchpads: PiLaunchpad[];
};

export type PiPlanetsResponse = {
  planets: PiPlanet[];
};

export type PiTaxConfig = {
  id?: number;
  userId?: number;
  planetId: number | null;
  taxRate: number;
};

// PI Profit Types

export type PiFactoryInput = {
  typeId: number;
  name: string;
  tier: string;
  quantity: number;
  pricePerUnit: number;
  costPerHour: number;
  importTaxPerHour: number;
  isLocal: boolean;
};

export type PiFactoryProfit = {
  pinId: number;
  schematicId: number;
  schematicName: string;
  outputTypeId: number;
  outputName: string;
  outputTier: string;
  outputQty: number;
  cycleTimeSec: number;
  ratePerHour: number;
  outputValuePerHour: number;
  inputCostPerHour: number;
  exportTaxPerHour: number;
  importTaxPerHour: number;
  profitPerHour: number;
  inputs: PiFactoryInput[];
};

export type PiPlanetProfit = {
  planetId: number;
  planetType: string;
  solarSystemId: number;
  solarSystemName: string;
  characterId: number;
  characterName: string;
  taxRate: number;
  totalOutputValue: number;
  totalInputCost: number;
  totalExportTax: number;
  totalImportTax: number;
  netProfitPerHour: number;
  factories: PiFactoryProfit[];
};

export type PiProfitResponse = {
  planets: PiPlanetProfit[];
  priceSource: string;
  totalOutputValue: number;
  totalInputCost: number;
  totalExportTax: number;
  totalImportTax: number;
  totalProfit: number;
};

// PI Supply Chain Types

export type SupplyChainPlanetEntry = {
  characterId: number;
  characterName: string;
  planetId: number;
  solarSystemName: string;
  planetType: string;
  ratePerHour: number;
};

export type SupplyChainItem = {
  typeId: number;
  name: string;
  tier: number;
  tierName: string;
  producedPerHour: number;
  consumedPerHour: number;
  netPerHour: number;
  stockpileQty: number;
  depletionHours: number;
  source: string;
  producers: SupplyChainPlanetEntry[];
  consumers: SupplyChainPlanetEntry[];
  stockpileMarkers?: StockpileMarker[];
};

export type SupplyChainResponse = {
  items: SupplyChainItem[];
};

// Industry Job Manager Types

export type IndustryJob = {
  jobId: number;
  installerId: number;
  userId: number;
  facilityId: number;
  stationId: number;
  activityId: number;
  blueprintId: number;
  blueprintTypeId: number;
  blueprintLocationId: number;
  outputLocationId: number;
  runs: number;
  cost?: number;
  licensedRuns?: number;
  probability?: number;
  productTypeId?: number;
  status: string;
  duration: number;
  startDate: string;
  endDate: string;
  pauseDate?: string;
  completedDate?: string;
  completedCharacterId?: number;
  successfulRuns?: number;
  solarSystemId?: number;
  source: string;
  updatedAt: string;
  blueprintName?: string;
  productName?: string;
  installerName?: string;
  systemName?: string;
  activityName?: string;
};

export type IndustryJobQueueEntry = {
  id: number;
  userId: number;
  characterId?: number;
  blueprintTypeId: number;
  activity: string;
  runs: number;
  meLevel: number;
  teLevel: number;
  systemId?: number;
  facilityTax: number;
  status: string;
  esiJobId?: number;
  productTypeId?: number;
  estimatedCost?: number;
  estimatedDuration?: number;
  notes?: string;
  createdAt: string;
  updatedAt: string;
  blueprintName?: string;
  productName?: string;
  characterName?: string;
  systemName?: string;
  esiJobEndDate?: string;
  esiJobSource?: string;
};

export type ManufacturingCalcResult = {
  blueprintTypeId: number;
  productTypeId: number;
  productName: string;
  runs: number;
  meFactor: number;
  teFactor: number;
  secsPerRun: number;
  totalDuration: number;
  totalProducts: number;
  inputCost: number;
  jobCost: number;
  totalCost: number;
  outputValue: number;
  profit: number;
  margin: number;
  materials: ManufacturingMaterial[];
};

export type ManufacturingMaterial = {
  typeId: number;
  name: string;
  baseQty: number;
  batchQty: number;
  price: number;
  cost: number;
};

export type BlueprintSearchResult = {
  BlueprintTypeID: number;
  BlueprintName: string;
  ProductTypeID: number;
  ProductName: string;
  Activity: string;
};

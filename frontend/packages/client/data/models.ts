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

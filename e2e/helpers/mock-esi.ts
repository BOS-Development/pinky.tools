/**
 * mock-esi.ts — Helper module for the mock ESI admin API.
 *
 * The mock ESI server exposes admin endpoints under /_admin/ when
 * E2E_TESTING=true. These allow tests to override canned ESI data
 * on a per-test basis and reset to defaults between tests.
 *
 * Base URL is resolved from MOCK_ESI_URL (default: http://localhost:8090).
 * In CI (Docker), set MOCK_ESI_URL=http://mock-esi:8090.
 */

const MOCK_ESI_BASE = process.env.MOCK_ESI_URL ?? 'http://localhost:8090';

// ---------------------------------------------------------------------------
// TypeScript interfaces — mirror the Go structs in cmd/mock-esi/main.go
// ---------------------------------------------------------------------------

export interface Asset {
  item_id: number;
  is_blueprint_copy: boolean;
  is_singleton: boolean;
  location_flag: string;
  location_id: number;
  location_type: string;
  quantity: number;
  type_id: number;
}

export interface SkillEntry {
  skill_id: number;
  trained_skill_level: number;
  active_skill_level: number;
  skillpoints_in_skill: number;
}

export interface SkillsResponse {
  skills: SkillEntry[];
  total_sp: number;
}

export interface IndustryJob {
  job_id: number;
  installer_id: number;
  facility_id: number;
  station_id: number;
  activity_id: number;
  blueprint_id: number;
  blueprint_type_id: number;
  blueprint_location_id: number;
  output_location_id: number;
  runs: number;
  cost: number;
  product_type_id: number;
  status: string;
  duration: number;
  start_date: string;
  end_date: string;
}

export interface BlueprintEntry {
  item_id: number;
  type_id: number;
  location_id: number;
  location_flag: string;
  /** -1 = BPO, -2 = BPC */
  quantity: number;
  material_efficiency: number;
  time_efficiency: number;
  /** -1 for BPOs (unlimited runs) */
  runs: number;
}

export interface MarketOrder {
  order_id: number;
  type_id: number;
  location_id: number;
  volume_total: number;
  volume_remain: number;
  min_volume: number;
  price: number;
  is_buy_order: boolean;
  duration: number;
  issued: string;
  range: string;
}

// ---------------------------------------------------------------------------
// PI interfaces — mirror the Go piPlanet / piColony types in cmd/mock-esi/main.go
// ---------------------------------------------------------------------------

export interface PiPlanet {
  last_update: string;
  num_pins: number;
  owner_id: number;
  planet_id: number;
  planet_type: string;
  solar_system_id: number;
  upgrade_level: number;
}

export interface PiPinContent {
  amount: number;
  type_id: number;
}

export interface PiExtractorDetail {
  cycle_time: number;
  head_radius: number;
  heads: object[];
  product_type_id: number;
  qty_per_cycle: number;
}

export interface PiFactoryDetail {
  schematic_id: number;
}

export interface PiPin {
  pin_id: number;
  type_id: number;
  latitude: number;
  longitude: number;
  install_time?: string;
  expiry_time?: string;
  last_cycle_start?: string;
  schematic_id?: number;
  contents: PiPinContent[];
  extractor_details?: PiExtractorDetail;
  factory_details?: PiFactoryDetail;
}

export interface PiLink {
  source_pin_id: number;
  destination_pin_id: number;
  link_level: number;
}

export interface PiRoute {
  route_id: number;
  source_pin_id: number;
  destination_pin_id: number;
  content_type_id: number;
  quantity: number;
  waypoints: number[];
}

export interface PiColony {
  links: PiLink[];
  pins: PiPin[];
  routes: PiRoute[];
}

// ---------------------------------------------------------------------------
// Internal helper
// ---------------------------------------------------------------------------

async function adminRequest(
  method: 'POST' | 'PUT',
  path: string,
  body?: unknown,
): Promise<void> {
  const url = `${MOCK_ESI_BASE}${path}`;
  const init: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json' },
  };
  if (body !== undefined) {
    init.body = JSON.stringify(body);
  }

  const res = await fetch(url, init);
  if (!res.ok) {
    let detail = '';
    try {
      detail = await res.text();
    } catch {
      // ignore
    }
    throw new Error(
      `Mock ESI admin request failed: ${method} ${path} → ${res.status} ${res.statusText}${detail ? ': ' + detail : ''}`,
    );
  }
}

// ---------------------------------------------------------------------------
// Exported API
// ---------------------------------------------------------------------------

/**
 * Reset all mock ESI state to the default canned fixtures.
 * Call this in afterEach/afterAll when a test mutates mock data.
 */
export async function resetMockESI(): Promise<void> {
  await adminRequest('POST', '/_admin/reset');
}

/**
 * Replace the asset list for a single character.
 * The backend will serve this data on the next /characters/{charID}/assets call.
 */
export async function setCharacterAssets(
  charID: number,
  assets: Asset[],
): Promise<void> {
  await adminRequest('PUT', `/_admin/character-assets/${charID}`, assets);
}

/**
 * Replace the skills response for a single character.
 */
export async function setCharacterSkills(
  charID: number,
  skills: SkillsResponse,
): Promise<void> {
  await adminRequest('PUT', `/_admin/character-skills/${charID}`, skills);
}

/**
 * Replace the industry jobs list for a single character.
 */
export async function setCharacterIndustryJobs(
  charID: number,
  jobs: IndustryJob[],
): Promise<void> {
  await adminRequest(
    'PUT',
    `/_admin/character-industry-jobs/${charID}`,
    jobs,
  );
}

/**
 * Replace the blueprint list for a single character.
 */
export async function setCharacterBlueprints(
  charID: number,
  blueprints: BlueprintEntry[],
): Promise<void> {
  await adminRequest(
    'PUT',
    `/_admin/character-blueprints/${charID}`,
    blueprints,
  );
}

/**
 * Replace the asset list for a corporation.
 */
export async function setCorpAssets(
  corpID: number,
  assets: Asset[],
): Promise<void> {
  await adminRequest('PUT', `/_admin/corp-assets/${corpID}`, assets);
}

/**
 * Replace the full market orders list (all regions share the same mock orders).
 */
export async function setMarketOrders(orders: MarketOrder[]): Promise<void> {
  await adminRequest('PUT', '/_admin/market-orders', orders);
}

/**
 * Set the planets list for a single character.
 * The backend will serve this data on the next /v1/characters/{charID}/planets/ call.
 */
export async function setCharacterPlanets(
  charID: number,
  planets: PiPlanet[],
): Promise<void> {
  await adminRequest('PUT', `/_admin/character-planets/${charID}`, planets);
}

/**
 * Set the colony details for a specific character+planet combination.
 * The backend will serve this on GET /v3/characters/{charID}/planets/{planetID}/.
 */
export async function setPlanetDetails(
  charID: number,
  planetID: number,
  colony: PiColony,
): Promise<void> {
  await adminRequest(
    'PUT',
    `/_admin/planet-details/${charID}/${planetID}`,
    colony,
  );
}

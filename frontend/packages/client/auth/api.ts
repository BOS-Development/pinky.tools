import * as client from "openid-client";

// Lazy discovery — only connects to EVE SSO when an OAuth flow is triggered,
// not at module import time (avoids failures in E2E where SSO is unreachable).
let config: client.Configuration | null = null;

async function getConfig(): Promise<client.Configuration> {
  if (!config) {
    config = await client.discovery(
      new URL("https://login.eveonline.com"),
      process.env.EVE_CLIENT_ID as string,
      process.env.EVE_CLIENT_SECRET as string,
    );
  }
  return config;
}

let redirect_uri = process.env.NEXTAUTH_URL + "api/auth/callback";

type StateEntry = {
  flowType: FlowType;
  codeVerifier: string;
};

const stateMaps: { [key: string]: StateEntry } = {};

export type FlowType = "login" | "char" | "corp";

export let scopesArray = [
  "publicData",
  "esi-skills.read_skills.v1",
  "esi-skills.read_skillqueue.v1",
  "esi-wallet.read_character_wallet.v1",
  "esi-search.search_structures.v1",
  "esi-clones.read_clones.v1",
  "esi-universe.read_structures.v1",
  "esi-assets.read_assets.v1",
  "esi-planets.manage_planets.v1",
  "esi-markets.structure_markets.v1",
  "esi-industry.read_character_jobs.v1",
  "esi-markets.read_character_orders.v1",
  "esi-characters.read_blueprints.v1",
  "esi-contracts.read_character_contracts.v1",
  "esi-clones.read_implants.v1",
  "esi-industry.read_character_mining.v1",
];
export let playerScope = scopesArray.join(" ");

export let corpScopesArray = [
  "esi-wallet.read_corporation_wallets.v1",
  "esi-assets.read_corporation_assets.v1",
  "esi-corporations.read_blueprints.v1",
  "esi-corporations.read_starbases.v1",
  "esi-industry.read_corporation_jobs.v1",
  "esi-markets.read_corporation_orders.v1",
  "esi-corporations.read_container_logs.v1",
  "esi-industry.read_corporation_mining.v1",
  "esi-corporations.read_facilities.v1",
  "esi-corporations.read_divisions.v1",
];
export let corpScope = corpScopesArray.join(" ");

let base = process.env.NEXTAUTH_URL as string;

export type VerifyTokenResponse = {
  tokenResponse: client.TokenEndpointResponse;
  flowType: FlowType;
};

export async function verifyToken(
  url: string,
  state: string,
): Promise<VerifyTokenResponse> {
  let entry = stateMaps[state];
  if (!entry) {
    throw new Error(`unknown state: ${state}`);
  }

  let cfg = await getConfig();
  let tokens: client.TokenEndpointResponse =
    await client.authorizationCodeGrant(cfg, new URL(base + url), {
      pkceCodeVerifier: entry.codeVerifier,
      expectedState: state,
    });

  // Clean up used state
  delete stateMaps[state];

  return {
    tokenResponse: tokens,
    flowType: entry.flowType,
  };
}

export default async function getAuthUrl(flowType: FlowType): Promise<string> {
  let code_verifier = client.randomPKCECodeVerifier();
  let code_challenge: string =
    await client.calculatePKCECodeChallenge(code_verifier);

  let scope: string;
  switch (flowType) {
    case "login":
      scope = "publicData";
      break;
    case "char":
      scope = playerScope;
      break;
    case "corp":
      scope = corpScope;
      break;
  }

  let parameters: Record<string, string> = {
    redirect_uri: redirect_uri,
    scope,
    code_challenge,
    code_challenge_method: "S256",
  };

  let state = client.randomState();
  parameters.state = state;
  stateMaps[state] = { flowType, codeVerifier: code_verifier };

  let cfg = await getConfig();
  let redirectTo: URL = client.buildAuthorizationUrl(cfg, parameters);
  return redirectTo.toString();
}

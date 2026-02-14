import * as client from "openid-client";

let config: client.Configuration = await client.discovery(
  new URL("https://login.eveonline.com"),
  process.env.ALT_EVE_CLIENT_ID as string,
  process.env.ALT_EVE_CLIENT_SECRET as string,
);

const code_verifier: string = client.randomPKCECodeVerifier();

let state!: string;
let redirect_uri = process.env.NEXTAUTH_URL + "api/altAuth/callback";

const stateMaps: { [key: string]: string } = {};

let scopesArray = [
  "publicData",
  "esi-skills.read_skills.v1",
  "esi-skills.read_skillqueue.v1",
  "esi-wallet.read_character_wallet.v1",
  "esi-wallet.read_corporation_wallet.v1",
  "esi-search.search_structures.v1",
  "esi-clones.read_clones.v1",
  "esi-universe.read_structures.v1",
  "esi-assets.read_assets.v1",
  "esi-planets.manage_planets.v1",
  "esi-markets.structure_markets.v1",
  "esi-corporations.read_structures.v1",
  "esi-industry.read_character_jobs.v1",
  "esi-markets.read_character_orders.v1",
  "esi-characters.read_blueprints.v1",
  "esi-contracts.read_character_contracts.v1",
  "esi-clones.read_implants.v1",
  "esi-wallet.read_corporation_wallets.v1",
  "esi-assets.read_corporation_assets.v1",
  "esi-corporations.read_blueprints.v1",
  "esi-contracts.read_corporation_contracts.v1",
  "esi-corporations.read_starbases.v1",
  "esi-industry.read_corporation_jobs.v1",
  "esi-markets.read_corporation_orders.v1",
  "esi-corporations.read_container_logs.v1",
  "esi-industry.read_character_mining.v1",
  "esi-industry.read_corporation_mining.v1",
  "esi-planets.read_customs_offices.v1",
  "esi-corporations.read_facilities.v1",
  "esi-corporations.read_freelance_jobs.v1",
];
let playerScope = scopesArray.join(" ");
let corpScopesArray = [
  "esi-wallet.read_corporation_wallet.v1",
  "esi-search.search_structures.v1",
  "esi-universe.read_structures.v1",
  "esi-wallet.read_corporation_wallets.v1",
  "esi-assets.read_corporation_assets.v1",
  "esi-corporations.read_blueprints.v1",
  "esi-corporations.read_starbases.v1",
  "esi-industry.read_corporation_jobs.v1",
  "esi-markets.read_corporation_orders.v1",
  "esi-corporations.read_container_logs.v1",
  "esi-industry.read_corporation_mining.v1",
  "esi-corporations.read_facilities.v1",
  "esi-corporations.read_freelance_jobs.v1",
  "esi-corporations.read_divisions.v1",
];
let corpScope = corpScopesArray.join(" ");

let base = process.env.NEXTAUTH_URL as string;

export type VerifyTokenResponse = {
  tokenResponse: client.TokenEndpointResponse;
  redirectType: string;
};
export async function verifyToken(
  url: string,
  state: string,
): Promise<VerifyTokenResponse> {
  let redirectType = stateMaps[state];
  let tokens: client.TokenEndpointResponse =
    await client.authorizationCodeGrant(config, new URL(base + url), {
      pkceCodeVerifier: code_verifier,
      expectedState: state,
    });

  return {
    tokenResponse: tokens,
    redirectType,
  };
}

export default async function getAuthUrl(isCorp: boolean): Promise<string> {
  let code_challenge: string =
    await client.calculatePKCECodeChallenge(code_verifier);

  let scope = isCorp ? corpScope : playerScope;
  let parameters: Record<string, string> = {
    redirect_uri: redirect_uri,
    scope,
    code_challenge,
    code_challenge_method: "S256",
  };

  state = client.randomState();
  parameters.state = state;
  stateMaps[state] = isCorp ? "corp" : "char";

  let redirectTo: URL = client.buildAuthorizationUrl(config, parameters);
  return redirectTo.toString();
}

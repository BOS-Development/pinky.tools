import * as client from "openid-client";
import { characterScopes, corporationScopes } from "../scope-definitions";

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

export let scopesArray = characterScopes;
export let playerScope = scopesArray.join(" ");

export let corpScopesArray = corporationScopes;
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

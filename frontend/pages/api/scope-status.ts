import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from "next-auth/next";
import { authOptions } from "./auth/[...nextauth]";
import client from "@industry-tool/client/api";
import { characterScopes, corporationScopes } from "@industry-tool/client/scope-definitions";

let backend = process.env.BACKEND_URL as string;

function hasMissingScopes(stored: string, required: string[]): boolean {
  if (!stored) return true;
  const set = new Set(stored.split(" "));
  return !required.every((s) => set.has(s));
}

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  const api = client(backend, session.providerAccountId);

  const [charsResp, corpsResp] = await Promise.all([
    api.getCharacters(),
    api.getCorporations(),
  ]);

  let outdatedCharacters = 0;
  let outdatedCorporations = 0;

  if (charsResp.kind === "success" && charsResp.data) {
    outdatedCharacters = charsResp.data.filter(
      (c) => hasMissingScopes(c.esiScopes, characterScopes)
    ).length;
  }

  if (corpsResp.kind === "success" && corpsResp.data) {
    outdatedCorporations = corpsResp.data.filter(
      (c) => hasMissingScopes(c.esiScopes, corporationScopes)
    ).length;
  }

  return res.status(200).json({
    outdatedCharacters,
    outdatedCorporations,
    hasOutdated: outdatedCharacters > 0 || outdatedCorporations > 0,
  });
}

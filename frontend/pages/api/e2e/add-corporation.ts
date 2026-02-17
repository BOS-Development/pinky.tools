import type { NextApiRequest, NextApiResponse } from "next";
import client from "@industry-tool/client/api";

let backend = process.env.BACKEND_URL as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  if (process.env.E2E_TESTING !== "true") {
    return res.status(404).json({ error: "Not found" });
  }

  if (req.method !== "POST") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  const { userId, characterId, characterName } = req.body;

  if (!userId || !characterId || !characterName) {
    return res.status(400).json({ error: "Missing required fields: userId, characterId, characterName" });
  }

  const expiresOn = new Date(Date.now() + 3600 * 1000);

  // This calls POST /v1/corporations which triggers the full ESI flow:
  // 1. Backend calls mock ESI POST /characters/affiliation to get corp ID
  // 2. Backend calls mock ESI GET /corporations/{corpID} to get corp name
  // 3. Backend upserts into player_corporations
  const response = await client(backend, String(userId)).addCharacterCorporation({
    id: characterId,
    name: characterName,
    esiToken: `fake-token-${characterName.toLowerCase().replace(/\s+/g, "-")}`,
    esiRefreshToken: `fake-refresh-${characterName.toLowerCase().replace(/\s+/g, "-")}`,
    esiTokenExpiresOn: expiresOn,
    esiScopes: "esi-assets.read_corporation_assets.v1 esi-corporations.read_divisions.v1",
  });

  if (response.kind === "error") {
    return res.status(500).json({ error: `Failed to add corporation: ${response.error}` });
  }

  return res.status(200).json({ success: true });
}

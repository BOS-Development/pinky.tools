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

  const response = await client(backend, String(userId)).addCharacter({
    id: characterId,
    name: characterName,
    esiToken: `fake-token-${characterName.toLowerCase().replace(/\s+/g, "-")}`,
    esiRefreshToken: `fake-refresh-${characterName.toLowerCase().replace(/\s+/g, "-")}`,
    esiTokenExpiresOn: expiresOn,
    esiScopes: [
      "publicData",
      "esi-assets.read_assets.v1",
      "esi-skills.read_skills.v1",
      "esi-industry.read_character_jobs.v1",
      "esi-characters.read_blueprints.v1",
      "esi-planets.manage_planets.v1",
      "esi-contracts.read_character_contracts.v1",
    ].join(" "),
  });

  if (response.kind === "error") {
    return res.status(500).json({ error: `Failed to add character: ${response.error}` });
  }

  return res.status(200).json({ success: true });
}

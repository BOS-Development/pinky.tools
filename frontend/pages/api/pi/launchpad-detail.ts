import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from "next-auth/next";
import { authOptions } from "../auth/[...nextauth]";

const backend = process.env.BACKEND_URL as string;
const backendKey = process.env.BACKEND_KEY as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  if (req.method !== "GET") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  const { characterId, planetId, pinId } = req.query;

  try {
    const response = await fetch(
      `${backend}v1/pi/launchpad-detail?characterId=${characterId}&planetId=${planetId}&pinId=${pinId}`,
      {
        headers: {
          "Content-Type": "application/json",
          "USER-ID": session.providerAccountId,
          "BACKEND-KEY": backendKey,
        },
      },
    );

    if (!response.ok) {
      const errorText = await response.text();
      return res.status(response.status).json({ error: errorText });
    }

    const data = await response.json();
    return res.status(200).json(data);
  } catch (error) {
    console.error("PI launchpad detail API error:", error);
    return res
      .status(500)
      .json({ error: "Failed to fetch launchpad detail" });
  }
}

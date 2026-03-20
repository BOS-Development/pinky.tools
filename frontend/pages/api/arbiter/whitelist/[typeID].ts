import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from "next-auth/next";
import { authOptions } from "../../auth/[...nextauth]";

const backend = process.env.BACKEND_URL as string;
const backendKey = process.env.BACKEND_KEY as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  if (req.method !== "DELETE") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  const { typeID } = req.query;

  try {
    const response = await fetch(`${backend}v1/arbiter/whitelist/${typeID}`, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
        "USER-ID": session.providerAccountId,
        "BACKEND-KEY": backendKey,
      },
    });

    if (!response.ok) {
      const errorText = await response.text();
      return res.status(response.status).json({ error: errorText });
    }

    return res.status(204).end();
  } catch (error) {
    console.error("Arbiter whitelist delete API error:", error);
    return res.status(500).json({ error: "Failed to remove from whitelist" });
  }
}

import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from "next-auth/next";
import { authOptions } from "../../auth/[...nextauth]";

const backend = process.env.BACKEND_URL as string;
const backendKey = process.env.BACKEND_KEY as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  if (req.method !== "GET") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  const { typeID, scope_id, quantity, build_all, me } = req.query;

  try {
    const params = new URLSearchParams();
    if (scope_id) params.set("scope_id", String(scope_id));
    if (quantity) params.set("quantity", String(quantity));
    if (build_all) params.set("build_all", String(build_all));
    if (me) params.set("me", String(me));

    const response = await fetch(
      `${backend}v1/arbiter/opportunities/${typeID}/bom?${params.toString()}`,
      {
        method: "GET",
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
    console.error("Arbiter BOM API error:", error);
    return res.status(500).json({ error: "Failed to fetch BOM" });
  }
}

import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from "next-auth/next";
import { authOptions } from "../auth/[...nextauth]";

const backend = process.env.BACKEND_URL as string;
const backendKey = process.env.BACKEND_KEY as string;

const getHeaders = (userId: string) => ({
  "Content-Type": "application/json",
  "USER-ID": userId,
  "BACKEND-KEY": backendKey,
});

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  const { runId } = req.query;

  if (!runId) {
    return res.status(400).json({ error: "runId query parameter is required" });
  }

  try {
    if (req.method === "GET") {
      const response = await fetch(
        `${backend}v1/hauling/runs/${runId}/pnl/summary`,
        {
          method: "GET",
          headers: getHeaders(session.providerAccountId),
        },
      );

      if (!response.ok) {
        const errorText = await response.text();
        return res.status(response.status).json({ error: errorText });
      }

      const data = await response.json();
      return res.status(200).json(data);
    } else {
      return res.status(405).json({ error: "Method not allowed" });
    }
  } catch (error) {
    console.error("Hauling P&L summary API error:", error);
    return res.status(500).json({ error: "Failed to process hauling P&L summary request" });
  }
}

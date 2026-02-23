import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from "next-auth/next";
import { authOptions } from "../../auth/[...nextauth]";

const backend = process.env.BACKEND_URL as string;
const backendKey = process.env.BACKEND_KEY as string;

const getHeaders = (id: string) => ({
  "Content-Type": "application/json",
  "USER-ID": id,
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

  if (req.method !== "GET") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  try {
    const userStationId = req.query.user_station_id;
    if (!userStationId) {
      return res.status(400).json({ error: "user_station_id is required" });
    }

    const response = await fetch(
      `${backend}v1/industry/plans/hangars?user_station_id=${userStationId}`,
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
  } catch (error) {
    console.error("Hangars API error:", error);
    return res.status(500).json({ error: "Failed to fetch hangars" });
  }
}

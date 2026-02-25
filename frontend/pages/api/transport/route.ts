import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from "next-auth/next";
import { authOptions } from "../auth/[...nextauth]";

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

  try {
    if (req.method === "GET") {
      const { origin, destination, flag } = req.query;
      const params = new URLSearchParams();
      if (origin) params.set("origin", String(origin));
      if (destination) params.set("destination", String(destination));
      if (flag) params.set("flag", String(flag));

      const response = await fetch(`${backend}v1/transport/route?${params.toString()}`, {
        method: "GET",
        headers: getHeaders(session.providerAccountId),
      });

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
    console.error("Transport route API error:", error);
    return res.status(500).json({ error: "Failed to process transport route request" });
  }
}

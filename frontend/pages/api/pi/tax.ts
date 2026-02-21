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

  const allowedMethods = ["GET", "POST", "DELETE"];
  if (!allowedMethods.includes(req.method || "")) {
    return res.status(405).json({ error: "Method not allowed" });
  }

  try {
    const response = await fetch(`${backend}v1/pi/tax`, {
      method: req.method,
      headers: {
        "Content-Type": "application/json",
        "USER-ID": session.providerAccountId,
        "BACKEND-KEY": backendKey,
      },
      body: req.method !== "GET" ? JSON.stringify(req.body) : undefined,
    });

    if (!response.ok) {
      const errorText = await response.text();
      return res.status(response.status).json({ error: errorText });
    }

    const text = await response.text();
    if (text) {
      return res.status(200).json(JSON.parse(text));
    }
    return res.status(200).json(null);
  } catch (error) {
    console.error("PI tax API error:", error);
    return res.status(500).json({ error: "Failed to handle PI tax config" });
  }
}

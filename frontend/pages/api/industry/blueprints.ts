import type { NextApiRequest, NextApiResponse } from "next";

const backend = process.env.BACKEND_URL as string;
const backendKey = process.env.BACKEND_KEY as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  if (req.method !== "GET") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  try {
    const queryString = new URLSearchParams(
      req.query as Record<string, string>,
    ).toString();
    const url = `${backend}v1/industry/blueprints${queryString ? `?${queryString}` : ""}`;

    const response = await fetch(url, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "BACKEND-KEY": backendKey,
      },
    });

    if (!response.ok) {
      const errorText = await response.text();
      return res.status(response.status).json({ error: errorText });
    }

    const data = await response.json();
    return res.status(200).json(data);
  } catch (error) {
    console.error("Industry blueprints API error:", error);
    return res.status(500).json({ error: "Failed to search blueprints" });
  }
}

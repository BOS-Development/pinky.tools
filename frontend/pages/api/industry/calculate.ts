import type { NextApiRequest, NextApiResponse } from "next";

const backend = process.env.BACKEND_URL as string;
const backendKey = process.env.BACKEND_KEY as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  if (req.method !== "POST") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  try {
    const response = await fetch(`${backend}v1/industry/calculate`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "BACKEND-KEY": backendKey,
      },
      body: JSON.stringify(req.body),
    });

    if (!response.ok) {
      const errorText = await response.text();
      return res.status(response.status).json({ error: errorText });
    }

    const data = await response.json();
    return res.status(200).json(data);
  } catch (error) {
    console.error("Industry calculate API error:", error);
    return res.status(500).json({ error: "Failed to calculate manufacturing cost" });
  }
}

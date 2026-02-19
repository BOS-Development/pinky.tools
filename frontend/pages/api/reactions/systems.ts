import type { NextApiRequest, NextApiResponse } from "next";

let backend = process.env.BACKEND_URL as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  if (req.method !== "GET") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  const response = await fetch(backend + "v1/reaction-systems", {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      "BACKEND-KEY": process.env.BACKEND_KEY as string,
    },
  });

  if (response.status !== 200) {
    return res.status(response.status).json({ error: "Failed to get reaction systems" });
  }

  const data = await response.json();
  return res.status(200).json(data);
}

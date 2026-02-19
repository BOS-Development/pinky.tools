import type { NextApiRequest, NextApiResponse } from "next";

let backend = process.env.BACKEND_URL as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  if (req.method !== "POST") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  const response = await fetch(backend + "v1/reactions/plan", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "BACKEND-KEY": process.env.BACKEND_KEY as string,
    },
    body: JSON.stringify(req.body),
  });

  if (response.status !== 200) {
    return res.status(response.status).json({ error: "Failed to compute plan" });
  }

  const data = await response.json();
  return res.status(200).json(data);
}

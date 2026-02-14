import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from "next-auth/next";
import { authOptions } from "../auth/[...nextauth]";

let backend = process.env.BACKEND_URL as string;

const getHeaders = (id: string) => {
  return {
    "Content-Type": "application/json",
    "USER-ID": id,
    "BACKEND-KEY": process.env.BACKEND_KEY as string,
  };
};

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  if (req.method !== "POST") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  const response = await fetch(backend + "v1/stockpiles", {
    method: "POST",
    headers: getHeaders(session.providerAccountId),
    body: JSON.stringify(req.body),
  });

  if (response.status !== 200) {
    return res.status(response.status).json({ error: "Failed to upsert stockpile" });
  }

  res.status(200).json({ success: true });
}

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
  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  const { id } = req.query;

  if (!id || typeof id !== "string") {
    return res.status(400).json({ error: "Invalid auto-sell config ID" });
  }

  if (req.method === "PUT") {
    const response = await fetch(backend + `v1/auto-sell/${id}`, {
      method: "PUT",
      headers: getHeaders(session.providerAccountId),
      body: JSON.stringify(req.body),
    });

    if (response.status !== 200) {
      return res.status(response.status).json({ error: "Failed to update auto-sell config" });
    }

    const data = await response.json();
    return res.status(200).json(data);
  }

  if (req.method === "DELETE") {
    const response = await fetch(backend + `v1/auto-sell/${id}`, {
      method: "DELETE",
      headers: getHeaders(session.providerAccountId),
    });

    if (response.status !== 200) {
      return res.status(response.status).json({ error: "Failed to delete auto-sell config" });
    }

    return res.status(200).json({ success: true });
  }

  return res.status(405).json({ error: "Method not allowed" });
}

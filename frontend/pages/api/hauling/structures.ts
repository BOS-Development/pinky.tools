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

  try {
    if (req.method === "GET") {
      const response = await fetch(`${backend}v1/hauling/structures`, {
        method: "GET",
        headers: getHeaders(session.providerAccountId),
      });

      if (!response.ok) {
        const errorText = await response.text();
        return res.status(response.status).json({ error: errorText });
      }

      const data = await response.json();
      return res.status(200).json(data);
    } else if (req.method === "POST") {
      const response = await fetch(`${backend}v1/hauling/structures`, {
        method: "POST",
        headers: getHeaders(session.providerAccountId),
        body: JSON.stringify(req.body),
      });

      if (!response.ok && response.status !== 403) {
        const errorText = await response.text();
        return res.status(response.status).json({ error: errorText });
      }

      const data = await response.json();
      return res.status(response.status === 403 ? 403 : 200).json(data);
    } else if (req.method === "DELETE") {
      const { id } = req.query;
      if (!id) {
        return res.status(400).json({ error: "Missing structure id" });
      }

      const response = await fetch(`${backend}v1/hauling/structures/${id}`, {
        method: "DELETE",
        headers: getHeaders(session.providerAccountId),
      });

      if (!response.ok) {
        const errorText = await response.text();
        return res.status(response.status).json({ error: errorText });
      }

      return res.status(204).end();
    } else {
      return res.status(405).json({ error: "Method not allowed" });
    }
  } catch (error) {
    console.error("Hauling structures API error:", error);
    return res.status(500).json({ error: "Failed to process hauling structures request" });
  }
}

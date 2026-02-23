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

  const { id } = req.query;

  try {
    if (req.method === "GET") {
      const response = await fetch(`${backend}v1/industry/plans/${id}`, {
        method: "GET",
        headers: getHeaders(session.providerAccountId),
      });

      if (!response.ok) {
        const errorText = await response.text();
        return res.status(response.status).json({ error: errorText });
      }

      const data = await response.json();
      return res.status(200).json(data);
    } else if (req.method === "PUT") {
      const response = await fetch(`${backend}v1/industry/plans/${id}`, {
        method: "PUT",
        headers: getHeaders(session.providerAccountId),
        body: JSON.stringify(req.body),
      });

      if (!response.ok) {
        const errorText = await response.text();
        return res.status(response.status).json({ error: errorText });
      }

      const data = await response.json();
      return res.status(200).json(data);
    } else if (req.method === "DELETE") {
      const response = await fetch(`${backend}v1/industry/plans/${id}`, {
        method: "DELETE",
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
    console.error("Production plan API error:", error);
    return res.status(500).json({ error: "Failed to process plan request" });
  }
}

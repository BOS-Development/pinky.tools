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
  res: NextApiResponse,
) {
  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  if (req.method === "GET") {
    const response = await fetch(backend + "v1/discord/link", {
      method: "GET",
      headers: getHeaders(session.providerAccountId),
    });

    if (response.status !== 200) {
      return res.status(response.status).json({ error: "Failed to get discord link" });
    }

    const data = await response.json();
    return res.status(200).json(data);
  }

  if (req.method === "DELETE") {
    const response = await fetch(backend + "v1/discord/link", {
      method: "DELETE",
      headers: getHeaders(session.providerAccountId),
    });

    if (response.status !== 200) {
      return res.status(response.status).json({ error: "Failed to unlink discord" });
    }

    const data = await response.json();
    return res.status(200).json(data);
  }

  return res.status(405).json({ error: "Method not allowed" });
}

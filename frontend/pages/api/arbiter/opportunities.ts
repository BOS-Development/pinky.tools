import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from "next-auth/next";
import { authOptions } from "../auth/[...nextauth]";

const backend = process.env.BACKEND_URL as string;
const backendKey = process.env.BACKEND_KEY as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  if (req.method !== "GET") {
    return res.status(405).json({ error: "Method not allowed" });
  }

  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  try {
    const params = new URLSearchParams();
    if (req.query.scope_id) params.set("scope_id", String(req.query.scope_id));
    if (req.query.input_price) params.set("input_price", String(req.query.input_price));
    if (req.query.output_price) params.set("output_price", String(req.query.output_price));
    if (req.query.decryptor_type_id) params.set("decryptor_type_id", String(req.query.decryptor_type_id));
    if (req.query.build_all) params.set("build_all", String(req.query.build_all));

    const url = `${backend}v1/arbiter/opportunities?${params.toString()}`;

    const response = await fetch(url, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "USER-ID": session.providerAccountId,
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
    console.error("Arbiter opportunities API error:", error);
    return res.status(500).json({ error: "Failed to fetch arbiter opportunities" });
  }
}

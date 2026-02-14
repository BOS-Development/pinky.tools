import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from "next-auth/next";
import { authOptions } from "../auth/[...nextauth]";
import client from "@industry-tool/client/api";

let backend = process.env.BACKEND_URL as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  const session = await getServerSession(req, res, authOptions);

  let response = await client(
    backend,
    session.providerAccountId,
  ).refreshAssets();

  if (response.kind === "error") {
    throw `error refreshing assets ${response.error}`;
  }

  res.redirect("/characters");
}

import type { NextApiRequest, NextApiResponse } from "next";
import getAuthUrl from "@/packages/client/auth/api";

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  let redirectTo = await getAuthUrl(true);

  res.redirect(redirectTo);
}

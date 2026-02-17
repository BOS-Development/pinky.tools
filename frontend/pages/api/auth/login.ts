import type { NextApiRequest, NextApiResponse } from "next";
import getAuthUrl from "@industry-tool/client/auth/api";

export default async function handler(
  _req: NextApiRequest,
  res: NextApiResponse,
) {
  let redirectTo = await getAuthUrl("login");
  res.redirect(redirectTo);
}

import type { NextApiRequest, NextApiResponse } from "next";
import { verifyToken } from "@industry-tool/client/auth/api";
import { getServerSession } from "next-auth/next";
import { authOptions } from "../auth/[...nextauth]";
import client from "@industry-tool/client/api";
import { jwtDecode } from "jwt-decode";
import moment from "moment";
import { TokenEndpointResponse } from "openid-client";

let backend = process.env.BACKEND_URL as string;

let handleCharAuthFlow = async (
  res: NextApiResponse,
  session: { providerAccountId: string },
  tokenResponse: TokenEndpointResponse,
) => {
  let jwtDecoded = jwtDecode(tokenResponse.access_token);
  let altId = jwtDecoded.sub?.split(":")[2];
  let expiresOn = moment().add(tokenResponse.expires_in, "seconds").toDate();

  let response = await client(backend, session.providerAccountId).addCharacter({
    id: Number(altId),
    name: (jwtDecoded as any).name as string,
    esiToken: tokenResponse.access_token,
    esiRefreshToken: tokenResponse.refresh_token as string,
    esiTokenExpiresOn: expiresOn,
  });

  if (response.kind === "error") {
    throw `error adding alt ${response.error}`;
  }

  res.redirect("/characters");
};

let handleCorpAuthFlow = async (
  res: NextApiResponse,
  session: { providerAccountId: string },
  tokenResponse: TokenEndpointResponse,
) => {
  let jwtDecoded = jwtDecode(tokenResponse.access_token);
  let altId = jwtDecoded.sub?.split(":")[2];
  let expiresOn = moment().add(tokenResponse.expires_in, "seconds").toDate();

  let response = await client(
    backend,
    session.providerAccountId,
  ).addCharacterCorporation({
    id: Number(altId),
    name: (jwtDecoded as any).name as string,
    esiToken: tokenResponse.access_token,
    esiRefreshToken: tokenResponse.refresh_token as string,
    esiTokenExpiresOn: expiresOn,
  });

  if (response.kind === "error") {
    throw `error adding alt corporation ${JSON.stringify(response)}`;
  }

  res.redirect("/corporations");
};

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  const session = await getServerSession(req, res, authOptions);

  let { tokenResponse, redirectType } = await verifyToken(
    req.url as string,
    req.query["state"] as string,
  );

  switch (redirectType) {
    case "char":
      await handleCharAuthFlow(res, session, tokenResponse);
      return;
    case "corp":
      await handleCorpAuthFlow(res, session, tokenResponse);
      return;
    default:
      throw `unknown redirect type ${redirectType}`;
  }
}

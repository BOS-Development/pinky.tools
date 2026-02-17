import type { NextApiRequest, NextApiResponse } from "next";
import { verifyToken, playerScope, corpScope } from "@industry-tool/client/auth/api";
import { getServerSession } from "next-auth/next";
import { authOptions, getUser, addUser } from "./[...nextauth]";
import client from "@industry-tool/client/api";
import { jwtDecode } from "jwt-decode";
import moment from "moment";
import { encode } from "next-auth/jwt";

let backend = process.env.BACKEND_URL as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  let { tokenResponse, flowType } = await verifyToken(
    req.url as string,
    req.query["state"] as string,
  );

  switch (flowType) {
    case "login":
      await handleLoginFlow(req, res, tokenResponse);
      return;
    case "char":
      await handleCharFlow(req, res, tokenResponse);
      return;
    case "corp":
      await handleCorpFlow(req, res, tokenResponse);
      return;
    default:
      throw `unknown flow type ${flowType}`;
  }
}

async function handleLoginFlow(
  _req: NextApiRequest,
  res: NextApiResponse,
  tokenResponse: any,
) {
  let jwtDecoded = jwtDecode(tokenResponse.access_token);
  let characterId = jwtDecoded.sub?.split(":")[2];
  let characterName = (jwtDecoded as any).name as string;

  // Create or lookup user in backend
  let existingUser = await getUser(Number(characterId));
  if (existingUser == null) {
    await addUser({
      id: Number(characterId),
      name: characterName,
    });
  }

  // Create NextAuth-compatible session JWT
  let token = await encode({
    token: {
      providerAccountId: characterId,
      name: characterName,
    },
    secret: process.env.NEXTAUTH_SECRET as string,
  });

  // Determine cookie name based on HTTPS vs HTTP
  let isSecure = (process.env.NEXTAUTH_URL || "").startsWith("https");
  let cookieName = isSecure
    ? "__Secure-next-auth.session-token"
    : "next-auth.session-token";

  let cookieFlags = `Path=/; HttpOnly; SameSite=Lax`;
  if (isSecure) {
    cookieFlags += "; Secure";
  }

  res.setHeader("Set-Cookie", `${cookieName}=${token}; ${cookieFlags}`);
  res.redirect("/");
}

async function handleCharFlow(
  req: NextApiRequest,
  res: NextApiResponse,
  tokenResponse: any,
) {
  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  let jwtDecoded = jwtDecode(tokenResponse.access_token);
  let altId = jwtDecoded.sub?.split(":")[2];
  let expiresOn = moment().add(tokenResponse.expires_in, "seconds").toDate();

  let response = await client(backend, session.providerAccountId).addCharacter({
    id: Number(altId),
    name: (jwtDecoded as any).name as string,
    esiToken: tokenResponse.access_token,
    esiRefreshToken: tokenResponse.refresh_token as string,
    esiTokenExpiresOn: expiresOn,
    esiScopes: playerScope,
  });

  if (response.kind === "error") {
    throw `error adding character ${response.error}`;
  }

  res.redirect("/characters");
}

async function handleCorpFlow(
  req: NextApiRequest,
  res: NextApiResponse,
  tokenResponse: any,
) {
  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

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
    esiScopes: corpScope,
  });

  if (response.kind === "error") {
    throw `error adding corporation ${JSON.stringify(response)}`;
  }

  res.redirect("/corporations");
}

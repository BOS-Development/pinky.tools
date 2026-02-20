import type { NextApiRequest, NextApiResponse } from "next";
import { getServerSession } from "next-auth/next";
import { authOptions } from "../auth/[...nextauth]";

let backend = process.env.BACKEND_URL as string;

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse,
) {
  const session = await getServerSession(req, res, authOptions);
  if (!session) {
    return res.status(401).json({ error: "Unauthorized" });
  }

  const code = req.query.code as string;
  if (!code) {
    return res.redirect("/settings?error=no_code");
  }

  const clientId = process.env.DISCORD_CLIENT_ID as string;
  const clientSecret = process.env.DISCORD_CLIENT_SECRET as string;
  const redirectUri = `${process.env.NEXTAUTH_URL}api/discord/callback`;

  // Exchange code for tokens
  const tokenResponse = await fetch("https://discord.com/api/v10/oauth2/token", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      client_id: clientId,
      client_secret: clientSecret,
      grant_type: "authorization_code",
      code: code,
      redirect_uri: redirectUri,
    }),
  });

  if (!tokenResponse.ok) {
    return res.redirect("/settings?error=token_exchange_failed");
  }

  const tokens = await tokenResponse.json();

  // Get Discord user info
  const userResponse = await fetch("https://discord.com/api/v10/users/@me", {
    headers: { Authorization: `Bearer ${tokens.access_token}` },
  });

  if (!userResponse.ok) {
    return res.redirect("/settings?error=user_fetch_failed");
  }

  const discordUser = await userResponse.json();

  // Save link to backend
  const saveResponse = await fetch(backend + "v1/discord/link", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "USER-ID": session.providerAccountId,
      "BACKEND-KEY": process.env.BACKEND_KEY as string,
    },
    body: JSON.stringify({
      discordUserId: discordUser.id,
      discordUsername: discordUser.username,
      accessToken: tokens.access_token,
      refreshToken: tokens.refresh_token,
      expiresIn: tokens.expires_in,
    }),
  });

  if (!saveResponse.ok) {
    return res.redirect("/settings?error=save_failed");
  }

  res.redirect("/settings?discord=linked");
}

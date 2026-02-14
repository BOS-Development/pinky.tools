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

  let path = backend + "v1/assets/";
  const response = await fetch(path, {
    method: "GET",
    headers: getHeaders(session.providerAccountId),
  });

  if (response.status == 404) {
    return {
      kind: "error",
      statusCode: 404,
      error: "",
    };
  }

  if (response.status != 200) {
    throw `call to ${path} reponse code ${response.status}`;
  }

  const resp = await response.json();

  res.status(200).json(resp);
}

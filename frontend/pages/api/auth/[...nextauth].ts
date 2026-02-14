import NextAuth from "next-auth";
import EVEOnlineProvider from "next-auth/providers/eveonline";

type User = {
  id: number;
  name: string;
};

let commonHeaders = {
  "Content-Type": "application/json",
  "BACKEND-KEY": process.env.BACKEND_KEY as string,
};

let getUser = async (id: number): Promise<User | null> => {
  let path = process.env.BACKEND_URL + "v1/users/" + id;
  const response = await fetch(path, {
    method: "GET",
    headers: commonHeaders,
  });

  if (response.status == 404) {
    return null;
  }

  if (response.status != 200) {
    throw `call to ${path} reponse code ${response.status}`;
  }

  const resp = await response.json();
  return resp;
};

let addUser = async (user: User): Promise<boolean> => {
  let path = process.env.BACKEND_URL + "v1/users/";
  const response = await fetch(path, {
    method: "POST",
    headers: commonHeaders,
    body: JSON.stringify(user),
  });

  if (response.status != 200) {
    throw `call to ${path} reponse code ${response.status}`;
  }

  return true;
};

export const authOptions = {
  providers: [
    EVEOnlineProvider({
      clientId: process.env.EVE_CLIENT_ID as string,
      clientSecret: process.env.EVE_CLIENT_SECRET as string,
    }),
  ],
  callbacks: {
    async jwt({ token, account }) {
      if (account) {
        token.providerAccountId = account.providerAccountId;
        let user = await getUser(account.providerAccountId);

        if (user == null) {
          await addUser({
            id: Number(token.providerAccountId),
            name: token.name,
          });
        }
      }

      return token;
    },
    async session({ session, token, user }) {
      session.providerAccountId = token.providerAccountId;
      return session;
    },
  },
};

export default NextAuth(authOptions);

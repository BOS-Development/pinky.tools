import NextAuth from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";

type User = {
  id: number;
  name: string;
};

let commonHeaders = {
  "Content-Type": "application/json",
  "BACKEND-KEY": process.env.BACKEND_KEY as string,
};

export let getUser = async (id: number): Promise<User | null> => {
  let path = process.env.BACKEND_URL + "v1/users/" + id;

  try {
    console.log('[NextAuth] Fetching user from backend:', { path, userId: id });

    const response = await fetch(path, {
      method: "GET",
      headers: commonHeaders,
    });

    console.log('[NextAuth] Backend response:', {
      status: response.status,
      statusText: response.statusText
    });

    if (response.status == 404) {
      console.log('[NextAuth] User not found (404)');
      return null;
    }

    if (response.status != 200) {
      const errorText = await response.text();
      console.error('[NextAuth] Backend error:', {
        status: response.status,
        body: errorText
      });
      throw `call to ${path} response code ${response.status}: ${errorText}`;
    }

    const resp = await response.json();
    return resp;
  } catch (error) {
    console.error('[NextAuth] Error fetching user:', {
      path,
      error: error instanceof Error ? error.message : error,
      stack: error instanceof Error ? error.stack : undefined,
      backendUrl: process.env.BACKEND_URL,
      backendKey: process.env.BACKEND_KEY ? '***set***' : 'NOT SET',
    });
    throw error;
  }
};

export let addUser = async (user: User): Promise<boolean> => {
  let path = process.env.BACKEND_URL + "v1/users/";

  try {
    console.log('[NextAuth] Creating new user in backend:', { path, userId: user.id, userName: user.name });

    const response = await fetch(path, {
      method: "POST",
      headers: commonHeaders,
      body: JSON.stringify(user),
    });

    console.log('[NextAuth] Backend create user response:', {
      status: response.status,
      statusText: response.statusText
    });

    if (response.status != 200) {
      const errorText = await response.text();
      console.error('[NextAuth] Backend error creating user:', {
        status: response.status,
        body: errorText
      });
      throw `call to ${path} response code ${response.status}: ${errorText}`;
    }

    console.log('[NextAuth] User created successfully');
    return true;
  } catch (error) {
    console.error('[NextAuth] Error creating user:', {
      path,
      user,
      error: error instanceof Error ? error.message : error,
      stack: error instanceof Error ? error.stack : undefined,
    });
    throw error;
  }
};

const providers: any[] = [
  CredentialsProvider({
    name: "E2E Test Credentials",
    credentials: {
      userId: { label: "User ID", type: "text" },
      userName: { label: "User Name", type: "text" },
    },
    async authorize(credentials) {
      if (!credentials?.userId || !credentials?.userName) {
        return null;
      }
      return {
        id: credentials.userId,
        name: credentials.userName,
      };
    },
  }),
];

export const authOptions = {
  providers,
  pages: {
    signIn: "/api/auth/login",
  },
  debug: process.env.NODE_ENV === 'development',
  logger: {
    error(code, metadata) {
      console.error('[NextAuth Error]', code, metadata);
    },
    warn(code) {
      console.warn('[NextAuth Warning]', code);
    },
    debug(code, metadata) {
      if (process.env.NODE_ENV === 'development') {
        console.debug('[NextAuth Debug]', code, metadata);
      }
    },
  },
  events: {
    async signIn(message) {
      console.log('[NextAuth] Sign in event:', {
        user: message.user?.name,
        provider: message.account?.provider,
        accountId: message.account?.providerAccountId,
      });
    },
    async signOut() {
      console.log('[NextAuth] Sign out event');
    },
  },
  callbacks: {
    async jwt({ token, account, user }) {
      try {
        if (account) {
          console.log('[NextAuth JWT] Processing new account:', {
            provider: account.provider,
            accountId: account.providerAccountId,
          });

          token.providerAccountId = user?.id;

          console.log('[NextAuth JWT] Fetching user from backend...');
          let existingUser = await getUser(Number(token.providerAccountId));

          if (existingUser == null) {
            console.log('[NextAuth JWT] User not found, creating new user');
            await addUser({
              id: Number(token.providerAccountId),
              name: token.name,
            });
          } else {
            console.log('[NextAuth JWT] Existing user found:', existingUser.name);
          }
        }

        return token;
      } catch (error) {
        console.error('[NextAuth JWT] Error in JWT callback:', error);
        throw error;
      }
    },
    async session({ session, token, user }) {
      try {
        session.providerAccountId = token.providerAccountId;
        return session;
      } catch (error) {
        console.error('[NextAuth Session] Error in session callback:', error);
        throw error;
      }
    },
  },
};

export default NextAuth(authOptions);

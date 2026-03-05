import Head from "next/head";
import { GetServerSideProps } from "next";
import { getSession, useSession } from "next-auth/react";
import client from "@industry-tool/client/api";
import { Character } from "@industry-tool/client/data/models";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import AccountInfo from "@industry-tool/components/settings/AccountInfo";
import LinkedCharacters from "@industry-tool/components/settings/LinkedCharacters";
import DiscordSettings from "@industry-tool/components/settings/DiscordSettings";

export type SettingsPageProps = {
  characters?: Character[];
};

export default function Settings({ characters }: SettingsPageProps) {
  const { data: session, status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return (
    <>
      <Head><title>Settings — pinky.tools</title></Head>
      <Navbar />
      <div className="max-w-3xl mx-auto px-4 py-8">
        <h1 className="text-2xl font-display font-semibold mb-6">Settings</h1>
        <div className="space-y-4">
          <AccountInfo
            userName={session.user?.name ?? ""}
            characterId={session.providerAccountId}
          />
          <LinkedCharacters characters={characters ?? []} />
          <DiscordSettings />
        </div>
      </div>
    </>
  );
}

export const getServerSideProps: GetServerSideProps<SettingsPageProps> = async (
  context
) => {
  const session = await getSession(context);
  if (!session) {
    return {
      props: {},
    };
  }

  const api = client(
    process.env.BACKEND_URL as string,
    session.providerAccountId
  );

  const charactersResponse = await api.getCharacters();
  if (charactersResponse.kind !== "success") {
    return {
      props: {},
    };
  }

  return {
    props: {
      characters: charactersResponse.data ?? [],
    },
  };
};

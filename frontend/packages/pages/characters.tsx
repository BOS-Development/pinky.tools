import { GetServerSideProps } from "next";
import { getSession, useSession } from "next-auth/react";
import client from "@industry-tool/client/api";
import { Character } from "@industry-tool/client/data/models";
import Loading from "@industry-tool/components/loading";
import Unuathorized from "@industry-tool/components/unauthorized";
import List from "@industry-tool/components/characters/list";

export type CharactersPageProps = {
  characters?: Character[];
};

export default function Home(props: CharactersPageProps) {
  const { data: session, status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unuathorized />;
  }

  console.log(props.characters);
  return <List characters={props.characters ?? []} />;
}

export const getServerSideProps: GetServerSideProps<
  CharactersPageProps
> = async (context) => {
  const session = await getSession(context);
  if (!session) {
    return {
      props: {},
    };
  }

  const api = client(
    process.env.BACKEND_URL as string,
    session.providerAccountId,
  );

  const charactersResponse = await api.getCharacters();
  if (charactersResponse.kind != "success") {
    return {
      props: {},
    };
  }

  return {
    props: {
      characters: charactersResponse.data,
    },
  };
};

import { GetServerSideProps } from "next";
import { getSession, useSession } from "next-auth/react";
import client from "@industry-tool/client/api";
import { Corporation } from "@industry-tool/client/data/models";
import Loading from "@industry-tool/components/loading";
import Unuathorized from "@industry-tool/components/unauthorized";
import List from "@industry-tool/components/corporations/list";

export type CorporationsPageProps = {
  corporations?: Corporation[];
};

export default function Corporations(props: CorporationsPageProps) {
  const { data: session, status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unuathorized />;
  }

  return <List corporations={props.corporations ?? []} />;
}

export const getServerSideProps: GetServerSideProps<
  CorporationsPageProps
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

  const corporationsResponse = await api.getCorporations();
  if (corporationsResponse.kind != "success") {
    return {
      props: {},
    };
  }

  return {
    props: {
      corporations: corporationsResponse.data,
    },
  };
};

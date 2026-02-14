import { GetServerSideProps } from "next";
import { getSession, useSession } from "next-auth/react";
import client from "@industry-tool/client/api";
import { AssetsResponse } from "@industry-tool/client/data/models";
import Loading from "@industry-tool/components/loading";
import Unuathorized from "@industry-tool/components/unauthorized";
import StockpilesList from "@industry-tool/components/stockpiles/StockpilesList";

export type StockpilesProps = {
  assets?: AssetsResponse;
};

export default function Stockpiles(props: StockpilesProps) {
  const { status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unuathorized />;
  }

  return <StockpilesList assets={props.assets ?? { structures: [] }} />;
}

export const getServerSideProps: GetServerSideProps<StockpilesProps> = async (
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

  const assetsResponse = await api.getAssets();
  if (assetsResponse.kind != "success") {
    return {
      props: {},
    };
  }

  return {
    props: {
      assets: assetsResponse.data,
    },
  };
};

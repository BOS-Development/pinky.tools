import { GetServerSideProps } from "next";
import { getSession, useSession } from "next-auth/react";
import client from "@industry-tool/client/api";
import { AssetsResponse } from "@industry-tool/client/data/models";
import Loading from "@industry-tool/components/loading";
import Unuathorized from "@industry-tool/components/unauthorized";
import AssetsList from "@industry-tool/components/assets/AssetsList";

export type AssetsProps = {
  assets?: AssetsResponse;
};

export default function Assets(props: AssetsProps) {
  const { status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unuathorized />;
  }

  return <AssetsList assets={props.assets ?? { structures: [] }} />;
}

export const getServerSideProps: GetServerSideProps<AssetsProps> = async (
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

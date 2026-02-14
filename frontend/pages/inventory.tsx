import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unuathorized from "@industry-tool/components/unauthorized";
import AssetsList from "@industry-tool/components/assets/AssetsList";

export default function Inventory() {
  const { status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unuathorized />;
  }

  // Assets will be loaded client-side by AssetsList component
  return <AssetsList />;
}

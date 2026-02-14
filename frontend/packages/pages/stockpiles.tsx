import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unuathorized from "@industry-tool/components/unauthorized";
import StockpilesList from "@industry-tool/components/stockpiles/StockpilesList";

export default function Stockpiles() {
  const { status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unuathorized />;
  }

  // Assets will be fetched client-side by StockpilesList component
  return <StockpilesList />;
}

import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import HaulingRunsList from "@industry-tool/components/hauling/HaulingRunsList";

export default function HaulingPage() {
  const { status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return <HaulingRunsList />;
}

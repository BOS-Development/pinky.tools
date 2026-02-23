import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import PlanRunsList from "@industry-tool/components/industry/PlanRunsList";

export default function PlanRunsPage() {
  const { status } = useSession();

  if (status === "loading") return <Loading />;
  if (status !== "authenticated") return <Unauthorized />;

  return <PlanRunsList />;
}

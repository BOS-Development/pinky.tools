import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import ProductionPlansList from "@industry-tool/components/industry/ProductionPlansList";

export default function ProductionPlansPage() {
  const { status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return <ProductionPlansList />;
}

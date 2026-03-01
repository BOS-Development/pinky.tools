import { useRouter } from "next/router";
import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import MarketScanner from "@industry-tool/components/hauling/MarketScanner";

export default function HaulingScannerPage() {
  const { status } = useSession();
  const router = useRouter();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  const sourceRegion = router.query.source_region
    ? Number(router.query.source_region)
    : undefined;
  const destRegion = router.query.dest_region
    ? Number(router.query.dest_region)
    : undefined;

  return (
    <MarketScanner
      initialSourceRegion={sourceRegion}
      initialDestRegion={destRegion}
    />
  );
}

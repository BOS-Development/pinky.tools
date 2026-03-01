import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import HaulingRunDetail from "@industry-tool/components/hauling/HaulingRunDetail";

interface HaulingDetailPageProps {
  runId: number;
}

export default function HaulingDetailPage({ runId }: HaulingDetailPageProps) {
  const { status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return <HaulingRunDetail runId={runId} />;
}

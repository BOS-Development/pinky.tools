import { GetServerSideProps } from "next";
import HaulingDetailPage from "@industry-tool/pages/hauling/detail";

interface Props {
  runId: number;
}

export default function HaulingRunPage({ runId }: Props) {
  return <HaulingDetailPage runId={runId} />;
}

export const getServerSideProps: GetServerSideProps = async (context) => {
  const { id } = context.params as { id: string };
  const runId = parseInt(id, 10);

  if (isNaN(runId)) {
    return { notFound: true };
  }

  return {
    props: {
      runId,
    },
  };
};

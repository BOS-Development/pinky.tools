import Head from "next/head";
import { useSession } from "next-auth/react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import Navbar from "@industry-tool/components/Navbar";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import HaulingRunsList from "@industry-tool/components/hauling/HaulingRunsList";
import HaulingHistory from "@industry-tool/components/hauling/HaulingHistory";
import HaulingAnalytics from "@industry-tool/components/hauling/HaulingAnalytics";

export default function HaulingPage() {
  const { status } = useSession();

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return (
    <>
      <Head><title>Hauling — pinky.tools</title></Head>
      <Navbar />
      <div className="w-full px-4 mt-8 mb-8">
        <Tabs defaultValue="active">
          <TabsList className="border-b border-overlay-medium bg-transparent w-full justify-start rounded-none p-0 h-auto mb-6">
            <TabsTrigger
              value="active"
              className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2"
            >
              Active Runs
            </TabsTrigger>
            <TabsTrigger
              value="history"
              className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2"
            >
              History
            </TabsTrigger>
            <TabsTrigger
              value="analytics"
              className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2"
            >
              Analytics
            </TabsTrigger>
          </TabsList>
          <TabsContent value="active">
            <HaulingRunsList />
          </TabsContent>
          <TabsContent value="history">
            <HaulingHistory />
          </TabsContent>
          <TabsContent value="analytics">
            <HaulingAnalytics />
          </TabsContent>
        </Tabs>
      </div>
    </>
  );
}

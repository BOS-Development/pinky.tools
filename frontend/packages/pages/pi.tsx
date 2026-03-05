import Head from "next/head";
import { useState } from "react";
import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unuathorized from "@industry-tool/components/unauthorized";
import PlanetOverview from "@industry-tool/components/pi/PlanetOverview";
import ProfitTable from "@industry-tool/components/pi/ProfitTable";
import SupplyChain from "@industry-tool/components/pi/SupplyChain";
import Navbar from "@industry-tool/components/Navbar";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";

export default function PlanetaryIndustry() {
  const { status } = useSession();
  const [tab, setTab] = useState(() => {
    if (typeof window !== "undefined") {
      return localStorage.getItem("pi-tab") || "overview";
    }
    return "overview";
  });

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unuathorized />;
  }

  return (
    <>
      <Head><title>Planetary Industry — pinky.tools</title></Head>
      <Navbar />
      <div className="max-w-screen-xl mx-auto px-4 mt-2 mb-8">
        <h2 className="text-xl font-semibold text-text-emphasis mb-4">
          Planetary Industry
        </h2>
        <Tabs
          value={tab}
          onValueChange={(v) => {
            setTab(v);
            localStorage.setItem("pi-tab", v);
          }}
        >
          <TabsList className="w-full justify-start mb-4">
            <TabsTrigger
              value="overview"
              className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none bg-transparent px-4 py-2 font-medium data-[state=active]:shadow-none"
            >
              Overview
            </TabsTrigger>
            <TabsTrigger
              value="profit"
              className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none bg-transparent px-4 py-2 font-medium data-[state=active]:shadow-none"
            >
              Profit
            </TabsTrigger>
            <TabsTrigger
              value="supply"
              className="text-text-muted data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none bg-transparent px-4 py-2 font-medium data-[state=active]:shadow-none"
            >
              Supply Chain
            </TabsTrigger>
          </TabsList>
          <TabsContent value="overview">
            <PlanetOverview embedded />
          </TabsContent>
          <TabsContent value="profit">
            <ProfitTable />
          </TabsContent>
          <TabsContent value="supply">
            <SupplyChain />
          </TabsContent>
        </Tabs>
      </div>
    </>
  );
}

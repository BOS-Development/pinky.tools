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
      <Navbar />
      <div className="max-w-screen-xl mx-auto px-4 mt-2 mb-8">
        <h2 className="text-xl font-semibold text-[#e2e8f0] mb-4">
          Planetary Industry
        </h2>
        <Tabs
          value={tab}
          onValueChange={(v) => {
            setTab(v);
            localStorage.setItem("pi-tab", v);
          }}
        >
          <TabsList className="border-b border-[rgba(148,163,184,0.15)] bg-transparent w-full justify-start rounded-none p-0 h-auto mb-4">
            <TabsTrigger
              value="overview"
              className="text-[#64748b] data-[state=active]:text-[#00d4ff] data-[state=active]:border-b-2 data-[state=active]:border-[#00d4ff] rounded-none bg-transparent px-4 py-2 font-medium data-[state=active]:shadow-none"
            >
              Overview
            </TabsTrigger>
            <TabsTrigger
              value="profit"
              className="text-[#64748b] data-[state=active]:text-[#00d4ff] data-[state=active]:border-b-2 data-[state=active]:border-[#00d4ff] rounded-none bg-transparent px-4 py-2 font-medium data-[state=active]:shadow-none"
            >
              Profit
            </TabsTrigger>
            <TabsTrigger
              value="supply"
              className="text-[#64748b] data-[state=active]:text-[#00d4ff] data-[state=active]:border-b-2 data-[state=active]:border-[#00d4ff] rounded-none bg-transparent px-4 py-2 font-medium data-[state=active]:shadow-none"
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

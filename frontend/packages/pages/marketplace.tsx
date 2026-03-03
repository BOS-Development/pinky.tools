import { useState } from 'react';
import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import MyListings from "@industry-tool/components/marketplace/MyListings";
import MarketplaceBrowser from "@industry-tool/components/marketplace/MarketplaceBrowser";
import PurchaseHistory from "@industry-tool/components/marketplace/PurchaseHistory";
import PendingSales from "@industry-tool/components/marketplace/PendingSales";
import BuyOrders from "@industry-tool/components/marketplace/BuyOrders";
import DemandViewer from "@industry-tool/components/marketplace/DemandViewer";
import SalesMetrics from "@industry-tool/components/analytics/SalesMetrics";
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';

const tabMap = ['listings', 'browse', 'pending', 'history', 'buy-orders', 'demand', 'analytics'];

export default function Marketplace() {
  const { status } = useSession();
  const [tabValue, setTabValue] = useState(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('marketplaceTab');
      return saved ? (tabMap[parseInt(saved, 10)] || 'listings') : 'listings';
    }
    return 'listings';
  });

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return (
    <>
      <Navbar />
      <div className="w-full px-4">
        <Tabs
          value={tabValue}
          onValueChange={(v) => {
            setTabValue(v);
            localStorage.setItem('marketplaceTab', String(tabMap.indexOf(v)));
          }}
        >
          <TabsList className="border-b border-[rgba(148,163,184,0.15)] bg-transparent w-full justify-start rounded-none p-0 h-auto mb-6 overflow-x-auto">
            <TabsTrigger
              value="listings"
              className="text-[#94a3b8] data-[state=active]:text-[#00d4ff] data-[state=active]:border-b-2 data-[state=active]:border-[#00d4ff] rounded-none bg-transparent px-4 py-2 whitespace-nowrap"
            >
              My Listings
            </TabsTrigger>
            <TabsTrigger
              value="browse"
              className="text-[#94a3b8] data-[state=active]:text-[#00d4ff] data-[state=active]:border-b-2 data-[state=active]:border-[#00d4ff] rounded-none bg-transparent px-4 py-2 whitespace-nowrap"
            >
              Browse
            </TabsTrigger>
            <TabsTrigger
              value="pending"
              className="text-[#94a3b8] data-[state=active]:text-[#00d4ff] data-[state=active]:border-b-2 data-[state=active]:border-[#00d4ff] rounded-none bg-transparent px-4 py-2 whitespace-nowrap"
            >
              Pending Sales
            </TabsTrigger>
            <TabsTrigger
              value="history"
              className="text-[#94a3b8] data-[state=active]:text-[#00d4ff] data-[state=active]:border-b-2 data-[state=active]:border-[#00d4ff] rounded-none bg-transparent px-4 py-2 whitespace-nowrap"
            >
              History
            </TabsTrigger>
            <TabsTrigger
              value="buy-orders"
              className="text-[#94a3b8] data-[state=active]:text-[#00d4ff] data-[state=active]:border-b-2 data-[state=active]:border-[#00d4ff] rounded-none bg-transparent px-4 py-2 whitespace-nowrap"
            >
              My Buy Orders
            </TabsTrigger>
            <TabsTrigger
              value="demand"
              className="text-[#94a3b8] data-[state=active]:text-[#00d4ff] data-[state=active]:border-b-2 data-[state=active]:border-[#00d4ff] rounded-none bg-transparent px-4 py-2 whitespace-nowrap"
            >
              Demand
            </TabsTrigger>
            <TabsTrigger
              value="analytics"
              className="text-[#94a3b8] data-[state=active]:text-[#00d4ff] data-[state=active]:border-b-2 data-[state=active]:border-[#00d4ff] rounded-none bg-transparent px-4 py-2 whitespace-nowrap"
            >
              Analytics
            </TabsTrigger>
          </TabsList>

          <TabsContent value="listings"><MyListings /></TabsContent>
          <TabsContent value="browse"><MarketplaceBrowser /></TabsContent>
          <TabsContent value="pending"><PendingSales /></TabsContent>
          <TabsContent value="history"><PurchaseHistory /></TabsContent>
          <TabsContent value="buy-orders"><BuyOrders /></TabsContent>
          <TabsContent value="demand"><DemandViewer /></TabsContent>
          <TabsContent value="analytics"><SalesMetrics /></TabsContent>
        </Tabs>
      </div>
    </>
  );
}

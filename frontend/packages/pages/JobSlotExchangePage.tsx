import Head from "next/head";
import { useState, useEffect } from 'react';
import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import SlotInventoryPanel from "@industry-tool/components/job-slots/SlotInventoryPanel";
import MyListings from "@industry-tool/components/job-slots/MyListings";
import ListingsBrowser from "@industry-tool/components/job-slots/ListingsBrowser";
import InterestRequests from "@industry-tool/components/job-slots/InterestRequests";
import Agreements from "@industry-tool/components/job-slots/Agreements";

export default function JobSlotExchangePage() {
  const { status } = useSession();
  const [tab, setTab] = useState(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('jobSlotExchangeTab') || 'inventory';
    }
    return 'inventory';
  });

  useEffect(() => {
    localStorage.setItem('jobSlotExchangeTab', tab);
  }, [tab]);

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return (
    <>
      <Head><title>Job Slots — pinky.tools</title></Head>
      <Navbar />
      <div className="px-4">
        <Tabs value={tab} onValueChange={setTab}>
          <TabsList className="mb-3">
            <TabsTrigger value="inventory">Slot Inventory</TabsTrigger>
            <TabsTrigger value="my-listings">My Listings</TabsTrigger>
            <TabsTrigger value="browse">Browse Listings</TabsTrigger>
            <TabsTrigger value="interest">Interest Requests</TabsTrigger>
            <TabsTrigger value="agreements">Agreements</TabsTrigger>
          </TabsList>
          <TabsContent value="inventory"><SlotInventoryPanel /></TabsContent>
          <TabsContent value="my-listings"><MyListings /></TabsContent>
          <TabsContent value="browse"><ListingsBrowser /></TabsContent>
          <TabsContent value="interest"><InterestRequests /></TabsContent>
          <TabsContent value="agreements"><Agreements /></TabsContent>
        </Tabs>
      </div>
    </>
  );
}

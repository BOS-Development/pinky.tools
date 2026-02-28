import { useState, useEffect } from 'react';
import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import SlotInventoryPanel from "@industry-tool/components/job-slots/SlotInventoryPanel";
import MyListings from "@industry-tool/components/job-slots/MyListings";
import ListingsBrowser from "@industry-tool/components/job-slots/ListingsBrowser";
import InterestRequests from "@industry-tool/components/job-slots/InterestRequests";

export default function JobSlotExchangePage() {
  const { status } = useSession();
  const [tabIndex, setTabIndex] = useState(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('jobSlotExchangeTab');
      return saved ? parseInt(saved, 10) : 0;
    }
    return 0;
  });

  useEffect(() => {
    localStorage.setItem('jobSlotExchangeTab', tabIndex.toString());
  }, [tabIndex]);

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return (
    <>
      <Navbar />
      <Container maxWidth={false}>
        <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
          <Tabs value={tabIndex} onChange={(_, newValue) => setTabIndex(newValue)}>
            <Tab label="Slot Inventory" />
            <Tab label="My Listings" />
            <Tab label="Browse Listings" />
            <Tab label="Interest Requests" />
          </Tabs>
        </Box>

        {tabIndex === 0 && <SlotInventoryPanel />}
        {tabIndex === 1 && <MyListings />}
        {tabIndex === 2 && <ListingsBrowser />}
        {tabIndex === 3 && <InterestRequests />}
      </Container>
    </>
  );
}

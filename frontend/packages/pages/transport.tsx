import React, { useCallback, useEffect, useState } from "react";
import { useSession } from "next-auth/react";
import { Container, Typography, Tabs, Tab, Box } from "@mui/material";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import { TransportProfilesList } from "../components/transport/TransportProfilesList";
import { JFRoutesList } from "../components/transport/JFRoutesList";
import { TransportJobsList } from "../components/transport/TransportJobsList";

interface TabPanelProps {
  children?: React.ReactNode;
  value: number;
  index: number;
}

function TabPanel({ children, value, index }: TabPanelProps) {
  return (
    <div role="tabpanel" hidden={value !== index}>
      {value === index && <Box sx={{ pt: 2 }}>{children}</Box>}
    </div>
  );
}

export interface TransportProfile {
  id: number;
  userId: number;
  name: string;
  transportMethod: string;
  characterId?: number;
  characterName?: string;
  cargoM3: number;
  ratePerM3PerJump: number;
  collateralRate: number;
  collateralPriceBasis: string;
  fuelTypeId?: number;
  fuelTypeName?: string;
  fuelPerLy?: number;
  fuelConservationLevel: number;
  routePreference: string;
  isDefault: boolean;
  createdAt: string;
}

export interface JFRouteWaypoint {
  id: number;
  routeId: number;
  sequence: number;
  systemId: number;
  systemName?: string;
  distanceLy: number;
}

export interface JFRoute {
  id: number;
  userId: number;
  name: string;
  originSystemId: number;
  originSystemName?: string;
  destinationSystemId: number;
  destinationSystemName?: string;
  totalDistanceLy: number;
  waypoints: JFRouteWaypoint[];
  createdAt: string;
}

export interface TransportJobItem {
  id: number;
  transportJobId: number;
  typeId: number;
  typeName?: string;
  quantity: number;
  volumeM3: number;
  estimatedValue: number;
}

export interface TransportJob {
  id: number;
  userId: number;
  originStationId: number;
  originStationName?: string;
  originSystemId: number;
  originSystemName?: string;
  destinationStationId: number;
  destinationStationName?: string;
  destinationSystemId: number;
  destinationSystemName?: string;
  transportMethod: string;
  routePreference: string;
  totalVolumeM3: number;
  totalCollateral: number;
  estimatedCost: number;
  jumps: number;
  distanceLy?: number;
  jfRouteId?: number;
  jfRouteName?: string;
  fulfillmentType: string;
  transportProfileId?: number;
  transportProfileName?: string;
  status: string;
  notes?: string;
  queueEntryId?: number;
  items: TransportJobItem[];
  createdAt: string;
}

export default function TransportPage() {
  const { data: session, status } = useSession();
  const [tabIndex, setTabIndex] = useState(0);
  const [profiles, setProfiles] = useState<TransportProfile[]>([]);
  const [jfRoutes, setJFRoutes] = useState<JFRoute[]>([]);
  const [jobs, setJobs] = useState<TransportJob[]>([]);
  const [loadingProfiles, setLoadingProfiles] = useState(true);
  const [loadingRoutes, setLoadingRoutes] = useState(true);
  const [loadingJobs, setLoadingJobs] = useState(true);

  const fetchProfiles = useCallback(async () => {
    try {
      setLoadingProfiles(true);
      const res = await fetch("/api/transport/profiles");
      if (res.ok) {
        const data = await res.json();
        setProfiles(data || []);
      }
    } catch (error) {
      console.error("Failed to fetch profiles:", error);
    } finally {
      setLoadingProfiles(false);
    }
  }, []);

  const fetchJFRoutes = useCallback(async () => {
    try {
      setLoadingRoutes(true);
      const res = await fetch("/api/transport/jf-routes");
      if (res.ok) {
        const data = await res.json();
        setJFRoutes(data || []);
      }
    } catch (error) {
      console.error("Failed to fetch JF routes:", error);
    } finally {
      setLoadingRoutes(false);
    }
  }, []);

  const fetchJobs = useCallback(async () => {
    try {
      setLoadingJobs(true);
      const res = await fetch("/api/transport/jobs");
      if (res.ok) {
        const data = await res.json();
        setJobs(data || []);
      }
    } catch (error) {
      console.error("Failed to fetch jobs:", error);
    } finally {
      setLoadingJobs(false);
    }
  }, []);

  useEffect(() => {
    if (status === "authenticated") {
      fetchProfiles();
      fetchJFRoutes();
      fetchJobs();
    }
  }, [status, fetchProfiles, fetchJFRoutes, fetchJobs]);

  if (status === "loading") return <Loading />;
  if (status !== "authenticated") return <Unauthorized />;

  return (
    <>
    <Navbar />
    <Container maxWidth="xl" sx={{ mt: 2 }}>
      <Typography variant="h5" sx={{ mb: 2, fontWeight: 600 }}>
        Transport
      </Typography>

      <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
        <Tabs
          value={tabIndex}
          onChange={(_, v) => setTabIndex(v)}
          sx={{
            "& .MuiTab-root": { color: "#94a3b8" },
            "& .Mui-selected": { color: "#3b82f6" },
            "& .MuiTabs-indicator": { backgroundColor: "#3b82f6" },
          }}
        >
          <Tab label="Transport Jobs" />
          <Tab label="Transport Profiles" />
          <Tab label="JF Routes" />
        </Tabs>
      </Box>

      <TabPanel value={tabIndex} index={0}>
        <TransportJobsList
          jobs={jobs}
          loading={loadingJobs}
          profiles={profiles}
          jfRoutes={jfRoutes}
          onRefresh={fetchJobs}
        />
      </TabPanel>

      <TabPanel value={tabIndex} index={1}>
        <TransportProfilesList
          profiles={profiles}
          loading={loadingProfiles}
          onRefresh={fetchProfiles}
        />
      </TabPanel>

      <TabPanel value={tabIndex} index={2}>
        <JFRoutesList
          routes={jfRoutes}
          loading={loadingRoutes}
          onRefresh={fetchJFRoutes}
        />
      </TabPanel>
    </Container>
    </>
  );
}

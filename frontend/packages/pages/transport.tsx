import React, { useCallback, useEffect, useState } from "react";
import { useSession } from "next-auth/react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import { TransportProfilesList } from "../components/transport/TransportProfilesList";
import { JFRoutesList } from "../components/transport/JFRoutesList";
import { TransportJobsList } from "../components/transport/TransportJobsList";

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
      <div className="max-w-screen-xl mx-auto px-4 mt-4">
        <h2 className="text-xl font-semibold text-text-emphasis mb-4">Transport</h2>

        <Tabs defaultValue="jobs">
          <TabsList className="w-full justify-start">
            <TabsTrigger
              value="jobs"
              className="text-text-secondary data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2"
            >
              Transport Jobs
            </TabsTrigger>
            <TabsTrigger
              value="profiles"
              className="text-text-secondary data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2"
            >
              Transport Profiles
            </TabsTrigger>
            <TabsTrigger
              value="routes"
              className="text-text-secondary data-[state=active]:text-primary data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none bg-transparent px-4 py-2"
            >
              JF Routes
            </TabsTrigger>
          </TabsList>

          <TabsContent value="jobs" className="mt-4">
            <TransportJobsList
              jobs={jobs}
              loading={loadingJobs}
              profiles={profiles}
              jfRoutes={jfRoutes}
              onRefresh={fetchJobs}
            />
          </TabsContent>

          <TabsContent value="profiles" className="mt-4">
            <TransportProfilesList
              profiles={profiles}
              loading={loadingProfiles}
              onRefresh={fetchProfiles}
            />
          </TabsContent>

          <TabsContent value="routes" className="mt-4">
            <JFRoutesList
              routes={jfRoutes}
              loading={loadingRoutes}
              onRefresh={fetchJFRoutes}
            />
          </TabsContent>
        </Tabs>
      </div>
    </>
  );
}

import { useState, useEffect, useCallback } from "react";
import { useSession } from "next-auth/react";
import { UserStation } from "@industry-tool/client/data/models";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import StationsList from "@industry-tool/components/stations/StationsList";

export default function Stations() {
  const { status } = useSession();
  const [stations, setStations] = useState<UserStation[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchStations = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/stations/user-stations");
      if (res.ok) {
        const data = await res.json();
        setStations(data || []);
      }
    } catch (err) {
      console.error("Failed to fetch stations:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (status === "authenticated") {
      fetchStations();
    }
  }, [status, fetchStations]);

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  return (
    <>
      <Navbar />
      <div className="max-w-7xl mx-auto px-4 py-6">
        <h2 className="text-xl font-display font-semibold text-[var(--color-text-emphasis)] mb-4">
          Preferred Stations
        </h2>
        <StationsList
          stations={stations}
          loading={loading}
          onRefresh={fetchStations}
        />
      </div>
    </>
  );
}

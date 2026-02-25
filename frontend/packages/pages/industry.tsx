import { useState, useEffect, useCallback } from "react";
import { useSession } from "next-auth/react";
import { IndustryJob, IndustryJobQueueEntry } from "@industry-tool/client/data/models";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import ActiveJobs from "@industry-tool/components/industry/ActiveJobs";
import JobQueue from "@industry-tool/components/industry/JobQueue";
import AddJob from "@industry-tool/components/industry/AddJob";
import Container from "@mui/material/Container";
import Typography from "@mui/material/Typography";
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";
import Box from "@mui/material/Box";

export default function Industry() {
  const { status } = useSession();
  const [tab, setTab] = useState(() => {
    if (typeof window !== "undefined") {
      return parseInt(localStorage.getItem("industry-tab") || "0", 10);
    }
    return 0;
  });

  const [jobs, setJobs] = useState<IndustryJob[]>([]);
  const [queue, setQueue] = useState<IndustryJobQueueEntry[]>([]);
  const [jobsLoading, setJobsLoading] = useState(true);
  const [queueLoading, setQueueLoading] = useState(true);

  const fetchJobs = useCallback(async () => {
    setJobsLoading(true);
    try {
      const res = await fetch("/api/industry/jobs");
      if (res.ok) {
        const data = await res.json();
        setJobs(data || []);
      }
    } catch (err) {
      console.error("Failed to fetch jobs:", err);
    } finally {
      setJobsLoading(false);
    }
  }, []);

  const fetchQueue = useCallback(async () => {
    setQueueLoading(true);
    try {
      const res = await fetch("/api/industry/queue");
      if (res.ok) {
        const data = await res.json();
        setQueue(data || []);
      }
    } catch (err) {
      console.error("Failed to fetch queue:", err);
    } finally {
      setQueueLoading(false);
    }
  }, []);

  useEffect(() => {
    if (status === "authenticated") {
      fetchJobs();
      fetchQueue();
    }
  }, [status, fetchJobs, fetchQueue]);

  const handleCancelEntry = async (id: number) => {
    try {
      const res = await fetch(`/api/industry/queue/${id}`, { method: "DELETE" });
      if (res.ok) {
        fetchQueue();
      }
    } catch (err) {
      console.error("Failed to cancel queue entry:", err);
    }
  };

  const handleJobAdded = () => {
    fetchQueue();
  };

  if (status === "loading") {
    return <Loading />;
  }

  if (status !== "authenticated") {
    return <Unauthorized />;
  }

  const handleTabChange = (_: React.SyntheticEvent, newValue: number) => {
    setTab(newValue);
    localStorage.setItem("industry-tab", String(newValue));
  };

  const plannedCount = queue.filter((e) => e.status === "planned").length;

  return (
    <>
      <Navbar />
      <Container maxWidth={false} sx={{ mt: 2, mb: 4 }}>
        <Typography variant="h5" sx={{ color: "#e2e8f0", mb: 2, fontWeight: 600 }}>
          Industry Jobs
        </Typography>
        <Box sx={{ borderBottom: 1, borderColor: "rgba(148, 163, 184, 0.15)", mb: 2 }}>
          <Tabs
            value={tab}
            onChange={handleTabChange}
            sx={{
              "& .MuiTab-root": { color: "#64748b", textTransform: "none", fontWeight: 500 },
              "& .Mui-selected": { color: "#3b82f6" },
              "& .MuiTabs-indicator": { backgroundColor: "#3b82f6" },
            }}
          >
            <Tab label={`Active Jobs (${jobs.length})`} />
            <Tab label={`Queue${plannedCount > 0 ? ` (${plannedCount})` : ""}`} />
            <Tab label="Add Job" />
          </Tabs>
        </Box>
        {tab === 0 && <ActiveJobs jobs={jobs} loading={jobsLoading} />}
        {tab === 1 && <JobQueue entries={queue} loading={queueLoading} onCancel={handleCancelEntry} />}
        {tab === 2 && <AddJob onJobAdded={handleJobAdded} />}
      </Container>
    </>
  );
}

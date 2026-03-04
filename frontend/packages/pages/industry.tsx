import { useState, useEffect, useCallback } from "react";
import { useSession } from "next-auth/react";
import { IndustryJob, IndustryJobQueueEntry } from "@industry-tool/client/data/models";
import Loading from "@industry-tool/components/loading";
import Unauthorized from "@industry-tool/components/unauthorized";
import Navbar from "@industry-tool/components/Navbar";
import ActiveJobs from "@industry-tool/components/industry/ActiveJobs";
import JobQueue from "@industry-tool/components/industry/JobQueue";
import AddJob from "@industry-tool/components/industry/AddJob";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";

export default function Industry() {
  const { status } = useSession();
  const [tab, setTab] = useState(() => {
    if (typeof window !== "undefined") {
      return localStorage.getItem("industry-tab") || "active";
    }
    return "active";
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

  const handleTabChange = (value: string) => {
    setTab(value);
    localStorage.setItem("industry-tab", value);
  };

  const plannedCount = queue.filter((e) => e.status === "planned").length;

  return (
    <>
      <Navbar />
      <div className="px-4 mt-2 mb-4">
        <h2 className="text-xl font-semibold text-text-emphasis mb-2">
          Industry Jobs
        </h2>
        <Tabs value={tab} onValueChange={handleTabChange}>
          <TabsList>
            <TabsTrigger value="active">{`Active Jobs (${jobs.length})`}</TabsTrigger>
            <TabsTrigger value="queue">{`Queue${plannedCount > 0 ? ` (${plannedCount})` : ""}`}</TabsTrigger>
            <TabsTrigger value="add">Add Job</TabsTrigger>
          </TabsList>
          <TabsContent value="active">
            <ActiveJobs jobs={jobs} loading={jobsLoading} />
          </TabsContent>
          <TabsContent value="queue">
            <JobQueue entries={queue} loading={queueLoading} onCancel={handleCancelEntry} onRefresh={fetchQueue} />
          </TabsContent>
          <TabsContent value="add">
            <AddJob onJobAdded={handleJobAdded} />
          </TabsContent>
        </Tabs>
      </div>
    </>
  );
}

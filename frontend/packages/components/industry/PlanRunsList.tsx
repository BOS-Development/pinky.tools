import { useState, useEffect, useRef, useCallback } from "react";
import { useSession } from "next-auth/react";
import { PlanRun } from "@industry-tool/client/data/models";
import Navbar from "@industry-tool/components/Navbar";
import Loading from "@industry-tool/components/loading";
import { XCircle, Play } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import { toast } from "@/components/ui/sonner";

const STATUS_CLASSES: Record<string, string> = {
  pending: "bg-accent-blue/10 border-accent-blue/30 text-accent-blue",
  in_progress: "bg-amber-manufacturing/10 border-amber-manufacturing/30 text-amber-manufacturing",
  completed: "bg-overlay-subtle border-overlay-strong text-text-secondary",
};

function formatStatus(status: string): string {
  switch (status) {
    case "in_progress":
      return "In Progress";
    case "pending":
      return "Pending";
    case "completed":
      return "Completed";
    default:
      return status;
  }
}

function formatJobProgress(run: PlanRun): string {
  if (!run.jobSummary || run.jobSummary.total === 0) return "\u2014";
  const { completed, active, total } = run.jobSummary;
  return `${completed}/${total} done${active > 0 ? `, ${active} active` : ""}`;
}

export default function PlanRunsList() {
  const { data: session } = useSession();
  const [runs, setRuns] = useState<PlanRun[]>([]);
  const [loading, setLoading] = useState(true);
  const hasFetchedRef = useRef(false);

  const fetchRuns = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/industry/plans/runs");
      if (res.ok) {
        const data = await res.json();
        setRuns(data || []);
      }
    } catch (err) {
      console.error("Failed to fetch plan runs:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchRuns();
    }
  }, [session, fetchRuns]);

  const handleCancel = async (runId: number) => {
    if (
      !confirm(
        "Are you sure you want to cancel all planned jobs for this run?",
      )
    ) {
      return;
    }

    try {
      const res = await fetch(`/api/industry/plans/runs/${runId}/cancel`, {
        method: "POST",
      });
      if (res.ok) {
        const data = await res.json();
        toast.success(`Cancelled ${data.jobsCancelled} planned job${data.jobsCancelled !== 1 ? "s" : ""}`);
        fetchRuns();
      } else {
        toast.error("Failed to cancel plan run");
      }
    } catch (err) {
      console.error("Failed to cancel plan run:", err);
      toast.error("Failed to cancel plan run");
    }
  };

  return (
    <>
      <Navbar />
      <div className="px-4 mt-2 mb-4">
        <div className="mb-2">
          <h2 className="text-xl font-semibold text-text-emphasis">
            Plan Runs
          </h2>
        </div>

        {loading ? (
          <Loading />
        ) : runs.length === 0 ? (
          <div className="empty-state">
            <Play className="empty-state-icon" />
            <p className="empty-state-title">
              No plan runs yet. Generate jobs from a production plan to create a run.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
            <Table>
              <TableHeader>
                <TableRow className="bg-background-void">
                  <TableHead>Product</TableHead>
                  <TableHead>Plan</TableHead>
                  <TableHead className="text-right">Qty</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Jobs</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="text-center">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {runs.map((run, idx) => (
                  <TableRow
                    key={run.id}
                    className={`${idx % 2 === 0 ? "bg-background-panel" : "bg-background-void"} hover:bg-background-elevated`}
                  >
                    <TableCell className="text-text-emphasis text-sm">
                      {run.productName || "\u2014"}
                    </TableCell>
                    <TableCell className="text-text-primary text-sm">
                      {run.planName || `Plan #${run.planId}`}
                    </TableCell>
                    <TableCell className="text-right text-text-primary text-sm">
                      {run.quantity}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant="outline"
                        className={`capitalize cursor-default ${STATUS_CLASSES[run.status] || STATUS_CLASSES.completed}`}
                      >
                        {formatStatus(run.status)}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-text-secondary text-[13px]">
                      {formatJobProgress(run)}
                    </TableCell>
                    <TableCell className="text-text-secondary text-[13px]">
                      {new Date(run.createdAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-center">
                      {run.status !== "completed" &&
                        run.jobSummary &&
                        run.jobSummary.planned > 0 && (
                          <button
                            className="p-1 rounded hover:bg-rose-danger/10 text-rose-danger"
                            onClick={() => handleCancel(run.id)}
                            title="Cancel planned jobs"
                          >
                            <XCircle className="h-4 w-4" />
                          </button>
                        )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </div>
    </>
  );
}

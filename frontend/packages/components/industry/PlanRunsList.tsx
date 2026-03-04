import { useState, useEffect, useRef, useCallback } from "react";
import { useSession } from "next-auth/react";
import { PlanRun } from "@industry-tool/client/data/models";
import Navbar from "@industry-tool/components/Navbar";
import Loading from "@industry-tool/components/loading";
import { XCircle } from "lucide-react";
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
  pending: "bg-[rgba(59,130,246,0.1)] border-[rgba(59,130,246,0.3)] text-[#3b82f6]",
  in_progress: "bg-[rgba(245,158,11,0.1)] border-[rgba(245,158,11,0.3)] text-[#f59e0b]",
  completed: "bg-[rgba(148,163,184,0.1)] border-[rgba(148,163,184,0.3)] text-[#94a3b8]",
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
          <h2 className="text-xl font-semibold text-[#e2e8f0]">
            Plan Runs
          </h2>
        </div>

        {loading ? (
          <Loading />
        ) : runs.length === 0 ? (
          <div className="bg-[#12151f] rounded-sm border border-[rgba(148,163,184,0.1)] p-8 text-center">
            <p className="text-[#64748b]">
              No plan runs yet. Generate jobs from a production plan to create a
              run.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
            <Table>
              <TableHeader>
                <TableRow className="bg-[#0f1219]">
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
                    className={`${idx % 2 === 0 ? "bg-[#12151f]" : "bg-[#0f1219]"} hover:bg-[#1a1d2e]`}
                  >
                    <TableCell className="text-[#e2e8f0] text-sm">
                      {run.productName || "\u2014"}
                    </TableCell>
                    <TableCell className="text-[#cbd5e1] text-sm">
                      {run.planName || `Plan #${run.planId}`}
                    </TableCell>
                    <TableCell className="text-right text-[#cbd5e1] text-sm">
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
                    <TableCell className="text-[#94a3b8] text-[13px]">
                      {formatJobProgress(run)}
                    </TableCell>
                    <TableCell className="text-[#94a3b8] text-[13px]">
                      {new Date(run.createdAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-center">
                      {run.status !== "completed" &&
                        run.jobSummary &&
                        run.jobSummary.planned > 0 && (
                          <button
                            className="p-1 rounded hover:bg-[rgba(239,68,68,0.1)] text-[#ef4444]"
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

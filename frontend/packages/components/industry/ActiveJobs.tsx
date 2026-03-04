import { IndustryJob } from "@industry-tool/client/data/models";
import { formatISK, formatNumber } from "@industry-tool/utils/formatting";
import { Loader2, User, Building2 } from "lucide-react";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "@/components/ui/tooltip";

type Props = {
  jobs: IndustryJob[];
  loading: boolean;
};

function getStatusClasses(status: string): string {
  switch (status) {
    case "active": return "bg-[rgba(16,185,129,0.1)] border-[rgba(16,185,129,0.3)] text-[#10b981]";
    case "ready": return "bg-[rgba(0,212,255,0.1)] border-[rgba(0,212,255,0.3)] text-[#00d4ff]";
    case "paused": return "bg-[rgba(245,158,11,0.1)] border-[rgba(245,158,11,0.3)] text-[#f59e0b]";
    case "delivered": return "bg-[rgba(148,163,184,0.1)] border-[rgba(148,163,184,0.3)] text-[#94a3b8]";
    case "cancelled": return "bg-[rgba(239,68,68,0.1)] border-[rgba(239,68,68,0.3)] text-[#ef4444]";
    default: return "bg-[rgba(148,163,184,0.1)] border-[rgba(148,163,184,0.3)] text-[#94a3b8]";
  }
}

function formatTimeRemaining(endDate: string): string {
  const end = new Date(endDate);
  const now = new Date();
  const diff = end.getTime() - now.getTime();

  if (diff <= 0) return "Ready";

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours > 24) {
    const days = Math.floor(hours / 24);
    const remHours = hours % 24;
    return `${days}d ${remHours}h`;
  }

  return `${hours}h ${minutes}m`;
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (hours > 24) {
    const days = Math.floor(hours / 24);
    const remHours = hours % 24;
    return `${days}d ${remHours}h`;
  }

  return `${hours}h ${minutes}m`;
}

export default function ActiveJobs({ jobs, loading }: Props) {
  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-[#00d4ff]" />
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
        <Table>
          <TableHeader>
            <TableRow className="bg-[#0f1219]">
              <TableHead className="text-center">Source</TableHead>
              <TableHead>Blueprint</TableHead>
              <TableHead>Product</TableHead>
              <TableHead>Activity</TableHead>
              <TableHead className="text-right">Runs</TableHead>
              <TableHead>Character</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Time Left</TableHead>
              <TableHead>Duration</TableHead>
              <TableHead className="text-right">Cost</TableHead>
              <TableHead>System</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {jobs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={11} className="text-center py-4 text-[#64748b]">
                  No active industry jobs
                </TableCell>
              </TableRow>
            ) : (
              jobs.map((job, idx) => (
                <TableRow
                  key={job.jobId}
                  className={idx % 2 === 0 ? "bg-[#0d1117]" : "bg-[#12151f]"}
                >
                  <TableCell className="text-center">
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="inline-flex">
                          {job.source === "corporation" ? (
                            <Building2 className="h-[18px] w-[18px] text-[#f59e0b]" />
                          ) : (
                            <User className="h-[18px] w-[18px] text-[#94a3b8]" />
                          )}
                        </span>
                      </TooltipTrigger>
                      <TooltipContent>
                        {job.source === "corporation" ? "Corporation Job" : "Character Job"}
                      </TooltipContent>
                    </Tooltip>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-[#e2e8f0]">
                      {job.blueprintName || `Type ${job.blueprintTypeId}`}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-[#cbd5e1]">
                      {job.productName || "-"}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-[#cbd5e1]">
                      {job.activityName || `Activity ${job.activityId}`}
                    </span>
                  </TableCell>
                  <TableCell className="text-right">
                    <span className="text-sm text-[#e2e8f0]">
                      {formatNumber(job.runs)}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-[#94a3b8]">
                      {job.installerName || "-"}
                    </span>
                  </TableCell>
                  <TableCell>
                    <Badge className={`border ${getStatusClasses(job.status)} cursor-default`}>
                      {job.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <span className={`text-sm ${job.status === "active" ? "text-[#00d4ff]" : "text-[#94a3b8]"}`}>
                      {formatTimeRemaining(job.endDate)}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-[#94a3b8]">
                      {formatDuration(job.duration)}
                    </span>
                  </TableCell>
                  <TableCell className="text-right">
                    <span className="text-sm text-[#cbd5e1]">
                      {job.cost ? formatISK(job.cost) : "-"}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-[#94a3b8]">
                      {job.systemName || "-"}
                    </span>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </TooltipProvider>
  );
}

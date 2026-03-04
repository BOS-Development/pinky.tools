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
    case "active": return "bg-teal-success/10 border-teal-success/30 text-teal-success";
    case "ready": return "bg-interactive-selected border-border-active text-primary";
    case "paused": return "bg-amber-manufacturing/10 border-amber-manufacturing/30 text-amber-manufacturing";
    case "delivered": return "bg-overlay-subtle border-overlay-strong text-text-secondary";
    case "cancelled": return "bg-rose-danger/10 border-rose-danger/30 text-rose-danger";
    default: return "bg-overlay-subtle border-overlay-strong text-text-secondary";
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
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
        <Table>
          <TableHeader>
            <TableRow className="bg-background-void">
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
                <TableCell colSpan={11} className="text-center py-4 text-text-muted">
                  No active industry jobs
                </TableCell>
              </TableRow>
            ) : (
              jobs.map((job, idx) => (
                <TableRow
                  key={job.jobId}
                  className={idx % 2 === 0 ? "bg-background-void" : "bg-background-panel"}
                >
                  <TableCell className="text-center">
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="inline-flex">
                          {job.source === "corporation" ? (
                            <Building2 className="h-[18px] w-[18px] text-amber-manufacturing" />
                          ) : (
                            <User className="h-[18px] w-[18px] text-text-secondary" />
                          )}
                        </span>
                      </TooltipTrigger>
                      <TooltipContent>
                        {job.source === "corporation" ? "Corporation Job" : "Character Job"}
                      </TooltipContent>
                    </Tooltip>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1.5">
                      <img
                        src={`https://images.evetech.net/types/${job.blueprintTypeId}/icon?size=32`}
                        alt=""
                        width={24}
                        height={24}
                        className="flex-shrink-0"
                        loading="lazy"
                        style={{ filter: "sepia(1) saturate(3) hue-rotate(180deg)" }}
                      />
                      <span className="text-sm text-text-emphasis">
                        {job.blueprintName || `Type ${job.blueprintTypeId}`}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1.5">
                      {job.productTypeId && (
                        <img
                          src={`https://images.evetech.net/types/${job.productTypeId}/icon?size=32`}
                          alt=""
                          width={24}
                          height={24}
                          className="flex-shrink-0"
                          loading="lazy"
                        />
                      )}
                      <span className="text-sm text-text-primary">
                        {job.productName || "-"}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-text-primary">
                      {job.activityName || `Activity ${job.activityId}`}
                    </span>
                  </TableCell>
                  <TableCell className="text-right">
                    <span className="text-sm text-text-emphasis">
                      {formatNumber(job.runs)}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-text-secondary">
                      {job.installerName || "-"}
                    </span>
                  </TableCell>
                  <TableCell>
                    <Badge className={`border ${getStatusClasses(job.status)} cursor-default`}>
                      {job.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div>
                      <span className={`text-sm ${job.status === "active" ? "text-primary" : "text-text-secondary"}`}>
                        {formatTimeRemaining(job.endDate)}
                      </span>
                      {job.status === "active" && (() => {
                        const total = job.duration * 1000;
                        const elapsed = total - (new Date(job.endDate).getTime() - Date.now());
                        const pct = Math.min(100, Math.max(0, (elapsed / total) * 100));
                        return (
                          <div className="mt-0.5 h-1 w-full rounded-full bg-overlay-subtle">
                            <div
                              className="h-1 rounded-full bg-primary"
                              style={{ width: `${pct}%` }}
                            />
                          </div>
                        );
                      })()}
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-text-secondary">
                      {formatDuration(job.duration)}
                    </span>
                  </TableCell>
                  <TableCell className="text-right">
                    <span className="text-sm text-text-primary">
                      {job.cost ? formatISK(job.cost) : "-"}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="text-sm text-text-secondary">
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

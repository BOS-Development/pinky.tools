import React, { useState } from "react";
import { Plus, ChevronDown, ChevronUp, Loader2, Truck } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { TransportJob, TransportProfile, JFRoute } from "../../pages/transport";
import { TransportJobDialog } from "./TransportJobDialog";
import { formatISK, formatNumber } from "../../utils/formatting";

interface Props {
  jobs: TransportJob[];
  loading: boolean;
  profiles: TransportProfile[];
  jfRoutes: JFRoute[];
  onRefresh: () => void;
}

const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    planned: "var(--color-primary-cyan)",
    in_transit: "var(--color-manufacturing-amber)",
    delivered: "var(--color-success-teal)",
    cancelled: "var(--color-danger-rose)",
  };
  return colors[status] || "var(--color-text-secondary)";
};

const getStatusBgColor = (status: string) => {
  const colors: Record<string, string> = {
    planned: "var(--color-info-tint)",
    in_transit: "var(--color-warning-tint)",
    delivered: "var(--color-success-tint)",
    cancelled: "var(--color-error-tint)",
  };
  return colors[status] || "var(--color-neutral-tint)";
};

const getStatusLabel = (status: string) => {
  const labels: Record<string, string> = {
    planned: "Planned",
    in_transit: "In Transit",
    delivered: "Delivered",
    cancelled: "Cancelled",
  };
  return labels[status] || status;
};

const getMethodLabel = (method: string) => {
  const labels: Record<string, string> = {
    freighter: "Freighter",
    jump_freighter: "Jump Freighter",
    dst: "DST",
    blockade_runner: "Blockade Runner",
  };
  return labels[method] || method;
};

const getFulfillmentLabel = (type: string) => {
  const labels: Record<string, string> = {
    self_haul: "Self Haul",
    courier_contract: "Courier",
    contact_haul: "Contact Haul",
  };
  return labels[type] || type;
};

export function TransportJobsList({ jobs, loading, profiles, jfRoutes, onRefresh }: Props) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [expandedJobId, setExpandedJobId] = useState<number | null>(null);

  const handleAdd = () => {
    setDialogOpen(true);
  };

  const handleDialogClose = (saved: boolean) => {
    setDialogOpen(false);
    if (saved) onRefresh();
  };

  const handleStatusChange = async (jobId: number, status: string) => {
    try {
      const res = await fetch(`/api/transport/jobs/${jobId}/status`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status }),
      });
      if (res.ok) onRefresh();
    } catch (error) {
      console.error("Failed to update status:", error);
    }
  };

  const toggleExpand = (jobId: number) => {
    setExpandedJobId(expandedJobId === jobId ? null : jobId);
  };

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  const colCount = 11;

  return (
    <>
      <div className="flex justify-end mb-3">
        <Button size="sm" onClick={handleAdd}>
          <Plus className="h-4 w-4 mr-2" />
          Create Transport Job
        </Button>
      </div>

      <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow className="bg-background-void hover:bg-background-void">
              <TableHead className="w-10" />
              <TableHead>Status</TableHead>
              <TableHead>Route</TableHead>
              <TableHead>Method</TableHead>
              <TableHead>Fulfillment</TableHead>
              <TableHead className="text-right">Volume (m3)</TableHead>
              <TableHead className="text-right">Collateral</TableHead>
              <TableHead className="text-right">Est. Cost</TableHead>
              <TableHead className="text-right">Jumps</TableHead>
              <TableHead>Profile</TableHead>
              <TableHead className="text-center">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {jobs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={colCount} className="p-0 border-0">
                  <div className="empty-state">
                    <Truck className="empty-state-icon" />
                    <p className="empty-state-title">No transport jobs</p>
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              jobs.map((job) => (
                <React.Fragment key={job.id}>
                  <TableRow
                    className={`hover:bg-interactive-hover ${expandedJobId === job.id ? "[&>td]:border-b-0" : ""}`}
                  >
                    <TableCell className="px-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6 text-text-secondary"
                        onClick={() => toggleExpand(job.id)}
                      >
                        {expandedJobId === job.id ? (
                          <ChevronUp className="h-4 w-4" />
                        ) : (
                          <ChevronDown className="h-4 w-4" />
                        )}
                      </Button>
                    </TableCell>
                    <TableCell>
                      <span
                        className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
                        style={{
                          backgroundColor: getStatusBgColor(job.status),
                          color: getStatusColor(job.status),
                        }}
                      >
                        {getStatusLabel(job.status)}
                      </span>
                    </TableCell>
                    <TableCell>
                      <p className="text-sm font-medium text-text-emphasis">
                        {job.originSystemName || "?"} → {job.destinationSystemName || "?"}
                      </p>
                      <span className="text-xs text-text-secondary">
                        {job.originStationName || ""} → {job.destinationStationName || ""}
                      </span>
                    </TableCell>
                    <TableCell className="text-text-secondary">{getMethodLabel(job.transportMethod)}</TableCell>
                    <TableCell className="text-text-secondary">{getFulfillmentLabel(job.fulfillmentType)}</TableCell>
                    <TableCell className="text-right">{formatNumber(job.totalVolumeM3)}</TableCell>
                    <TableCell className="text-right">{formatISK(job.totalCollateral)}</TableCell>
                    <TableCell className="text-right text-rose-danger">
                      {formatISK(job.estimatedCost)}
                    </TableCell>
                    <TableCell className="text-right">{job.jumps}</TableCell>
                    <TableCell className="text-text-secondary">{job.transportProfileName || "—"}</TableCell>
                    <TableCell className="text-center">
                      {job.status === "planned" && (
                        <div className="flex gap-1 justify-center">
                          <Button
                            size="sm"
                            variant="outline"
                            className="text-[0.65rem] py-0 h-6"
                            onClick={() => handleStatusChange(job.id, "in_transit")}
                          >
                            Start
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            className="text-[0.65rem] py-0 h-6 border-rose-danger text-rose-danger hover:bg-rose-danger/10"
                            onClick={() => handleStatusChange(job.id, "cancelled")}
                          >
                            Cancel
                          </Button>
                        </div>
                      )}
                      {job.status === "in_transit" && (
                        <Button
                          size="sm"
                          variant="outline"
                          className="text-[0.65rem] py-0 h-6 border-teal-success text-teal-success hover:bg-teal-success/10"
                          onClick={() => handleStatusChange(job.id, "delivered")}
                        >
                          Delivered
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                  {expandedJobId === job.id && (
                    <TableRow>
                      <TableCell colSpan={colCount} className="py-0 px-0">
                        <div className="px-6 py-3">
                          {job.items && job.items.length > 0 ? (
                            <Table>
                              <TableHeader>
                                <TableRow className="[&>th]:text-text-secondary [&>th]:border-overlay-subtle [&>th]:py-1">
                                  <TableHead>Item</TableHead>
                                  <TableHead className="text-right">Quantity</TableHead>
                                  <TableHead className="text-right">Volume (m³)</TableHead>
                                </TableRow>
                              </TableHeader>
                              <TableBody>
                                {job.items.map((item) => (
                                  <TableRow key={item.id} className="[&>td]:border-overlay-subtle [&>td]:py-1">
                                    <TableCell>
                                      <div className="flex gap-2 items-center">
                                        <img
                                          src={`https://images.evetech.net/types/${item.typeId}/icon?size=32`}
                                          alt=""
                                          style={{ width: 20, height: 20 }}
                                        />
                                        {item.typeName || `Type ${item.typeId}`}
                                      </div>
                                    </TableCell>
                                    <TableCell className="text-right">{formatNumber(item.quantity)}</TableCell>
                                    <TableCell className="text-right">{formatNumber(item.volumeM3)}</TableCell>
                                  </TableRow>
                                ))}
                              </TableBody>
                            </Table>
                          ) : (
                            <p className="text-sm text-text-muted">No items in this transport job.</p>
                          )}
                          {job.notes && (
                            <p className="text-sm text-text-secondary mt-2">Notes: {job.notes}</p>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  )}
                </React.Fragment>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <TransportJobDialog
        open={dialogOpen}
        onClose={handleDialogClose}
        profiles={profiles}
        jfRoutes={jfRoutes}
      />
    </>
  );
}

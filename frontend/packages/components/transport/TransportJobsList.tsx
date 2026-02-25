import React, { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Button,
  Box,
  Chip,
  CircularProgress,
  Typography,
  Collapse,
  IconButton,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
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
    planned: "#3b82f6",
    in_transit: "#f59e0b",
    delivered: "#10b981",
    cancelled: "#ef4444",
  };
  return colors[status] || "#94a3b8";
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
      <Box sx={{ display: "flex", justifyContent: "center", py: 4 }}>
        <CircularProgress size={32} />
      </Box>
    );
  }

  const colCount = 11;

  return (
    <>
      <Box sx={{ display: "flex", justifyContent: "flex-end", mb: 2 }}>
        <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd} size="small">
          Create Transport Job
        </Button>
      </Box>

      <TableContainer>
        <Table size="small">
          <TableHead>
            <TableRow sx={{ backgroundColor: "#0f1219" }}>
              <TableCell sx={{ width: 40 }} />
              <TableCell>Status</TableCell>
              <TableCell>Route</TableCell>
              <TableCell>Method</TableCell>
              <TableCell>Fulfillment</TableCell>
              <TableCell align="right">Volume (m3)</TableCell>
              <TableCell align="right">Collateral</TableCell>
              <TableCell align="right">Est. Cost</TableCell>
              <TableCell align="right">Jumps</TableCell>
              <TableCell>Profile</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {jobs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={colCount} align="center" sx={{ py: 4, color: "#94a3b8" }}>
                  <Typography variant="body2">No transport jobs</Typography>
                </TableCell>
              </TableRow>
            ) : (
              jobs.map((job) => (
                <React.Fragment key={job.id}>
                  <TableRow
                    sx={{
                      "&:hover": { backgroundColor: "rgba(59, 130, 246, 0.05)" },
                      "& > td": expandedJobId === job.id ? { borderBottom: "none" } : {},
                    }}
                  >
                    <TableCell sx={{ px: 0.5 }}>
                      <IconButton
                        size="small"
                        onClick={() => toggleExpand(job.id)}
                        sx={{ color: "#94a3b8" }}
                      >
                        {expandedJobId === job.id ? (
                          <KeyboardArrowUpIcon fontSize="small" />
                        ) : (
                          <KeyboardArrowDownIcon fontSize="small" />
                        )}
                      </IconButton>
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={getStatusLabel(job.status)}
                        size="small"
                        sx={{
                          backgroundColor: `${getStatusColor(job.status)}20`,
                          color: getStatusColor(job.status),
                          fontWeight: 500,
                        }}
                      />
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontWeight: 500 }}>
                        {job.originSystemName || "?"} → {job.destinationSystemName || "?"}
                      </Typography>
                      <Typography variant="caption" sx={{ color: "#94a3b8" }}>
                        {job.originStationName || ""} → {job.destinationStationName || ""}
                      </Typography>
                    </TableCell>
                    <TableCell>{getMethodLabel(job.transportMethod)}</TableCell>
                    <TableCell>{getFulfillmentLabel(job.fulfillmentType)}</TableCell>
                    <TableCell align="right">{formatNumber(job.totalVolumeM3)}</TableCell>
                    <TableCell align="right">{formatISK(job.totalCollateral)}</TableCell>
                    <TableCell align="right" sx={{ color: "#ef4444" }}>
                      {formatISK(job.estimatedCost)}
                    </TableCell>
                    <TableCell align="right">{job.jumps}</TableCell>
                    <TableCell>{job.transportProfileName || "—"}</TableCell>
                    <TableCell align="center">
                      {job.status === "planned" && (
                        <Box sx={{ display: "flex", gap: 0.5, justifyContent: "center" }}>
                          <Button
                            size="small"
                            variant="outlined"
                            onClick={() => handleStatusChange(job.id, "in_transit")}
                            sx={{ fontSize: "0.65rem", py: 0 }}
                          >
                            Start
                          </Button>
                          <Button
                            size="small"
                            variant="outlined"
                            color="error"
                            onClick={() => handleStatusChange(job.id, "cancelled")}
                            sx={{ fontSize: "0.65rem", py: 0 }}
                          >
                            Cancel
                          </Button>
                        </Box>
                      )}
                      {job.status === "in_transit" && (
                        <Button
                          size="small"
                          variant="outlined"
                          color="success"
                          onClick={() => handleStatusChange(job.id, "delivered")}
                          sx={{ fontSize: "0.65rem", py: 0 }}
                        >
                          Delivered
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell sx={{ py: 0, px: 0 }} colSpan={colCount}>
                      <Collapse in={expandedJobId === job.id} timeout="auto" unmountOnExit>
                        <Box sx={{ px: 3, py: 1.5 }}>
                          {job.items && job.items.length > 0 ? (
                            <Table size="small">
                              <TableHead>
                                <TableRow sx={{ "& th": { color: "#94a3b8", borderColor: "#1e2231", py: 0.5 } }}>
                                  <TableCell>Item</TableCell>
                                  <TableCell align="right">Quantity</TableCell>
                                  <TableCell align="right">Volume (m³)</TableCell>
                                </TableRow>
                              </TableHead>
                              <TableBody>
                                {job.items.map((item) => (
                                  <TableRow key={item.id} sx={{ "& td": { borderColor: "#1e2231", py: 0.5 } }}>
                                    <TableCell>
                                      <Box sx={{ display: "flex", gap: 1, alignItems: "center" }}>
                                        <img
                                          src={`https://images.evetech.net/types/${item.typeId}/icon?size=32`}
                                          alt=""
                                          style={{ width: 20, height: 20 }}
                                        />
                                        {item.typeName || `Type ${item.typeId}`}
                                      </Box>
                                    </TableCell>
                                    <TableCell align="right">{formatNumber(item.quantity)}</TableCell>
                                    <TableCell align="right">{formatNumber(item.volumeM3)}</TableCell>
                                  </TableRow>
                                ))}
                              </TableBody>
                            </Table>
                          ) : (
                            <Typography variant="body2" sx={{ color: "#64748b" }}>
                              No items in this transport job.
                            </Typography>
                          )}
                          {job.notes && (
                            <Typography variant="body2" sx={{ color: "#94a3b8", mt: 1 }}>
                              Notes: {job.notes}
                            </Typography>
                          )}
                        </Box>
                      </Collapse>
                    </TableCell>
                  </TableRow>
                </React.Fragment>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <TransportJobDialog
        open={dialogOpen}
        onClose={handleDialogClose}
        profiles={profiles}
        jfRoutes={jfRoutes}
      />
    </>
  );
}

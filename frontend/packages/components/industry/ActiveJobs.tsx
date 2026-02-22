import { IndustryJob } from "@industry-tool/client/data/models";
import { formatISK, formatNumber } from "@industry-tool/utils/formatting";
import Box from "@mui/material/Box";
import Chip from "@mui/material/Chip";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Typography from "@mui/material/Typography";
import CircularProgress from "@mui/material/CircularProgress";
import Tooltip from "@mui/material/Tooltip";
import PersonIcon from "@mui/icons-material/Person";
import CorporateFareIcon from "@mui/icons-material/CorporateFare";

type Props = {
  jobs: IndustryJob[];
  loading: boolean;
};

function getStatusColor(status: string): "success" | "warning" | "info" | "error" | "default" {
  switch (status) {
    case "active": return "success";
    case "ready": return "info";
    case "paused": return "warning";
    case "delivered": return "default";
    case "cancelled": return "error";
    default: return "default";
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
      <Box sx={{ display: "flex", justifyContent: "center", py: 8 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <TableContainer>
      <Table size="small">
        <TableHead>
          <TableRow sx={{ backgroundColor: "#0f1219" }}>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="center">Source</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Blueprint</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Product</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Activity</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="right">Runs</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Character</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Status</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Time Left</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Duration</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="right">Cost</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>System</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {jobs.length === 0 ? (
            <TableRow>
              <TableCell colSpan={11} align="center" sx={{ py: 4, color: "#64748b" }}>
                No active industry jobs
              </TableCell>
            </TableRow>
          ) : (
            jobs.map((job) => (
              <TableRow
                key={job.jobId}
                sx={{
                  "&:nth-of-type(odd)": { backgroundColor: "#0d1117" },
                  "&:nth-of-type(even)": { backgroundColor: "#12151f" },
                }}
              >
                <TableCell align="center">
                  <Tooltip title={job.source === "corporation" ? "Corporation Job" : "Character Job"}>
                    {job.source === "corporation" ? (
                      <CorporateFareIcon sx={{ color: "#f59e0b", fontSize: 18 }} />
                    ) : (
                      <PersonIcon sx={{ color: "#94a3b8", fontSize: 18 }} />
                    )}
                  </Tooltip>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#e2e8f0" }}>
                    {job.blueprintName || `Type ${job.blueprintTypeId}`}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#cbd5e1" }}>
                    {job.productName || "-"}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#cbd5e1" }}>
                    {job.activityName || `Activity ${job.activityId}`}
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2" sx={{ color: "#e2e8f0" }}>
                    {formatNumber(job.runs)}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#94a3b8" }}>
                    {job.installerName || "-"}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Chip
                    label={job.status}
                    color={getStatusColor(job.status)}
                    size="small"
                    variant="outlined"
                  />
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: job.status === "active" ? "#3b82f6" : "#94a3b8" }}>
                    {formatTimeRemaining(job.endDate)}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#94a3b8" }}>
                    {formatDuration(job.duration)}
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2" sx={{ color: "#cbd5e1" }}>
                    {job.cost ? formatISK(job.cost) : "-"}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#94a3b8" }}>
                    {job.systemName || "-"}
                  </Typography>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

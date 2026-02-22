import { IndustryJobQueueEntry } from "@industry-tool/client/data/models";
import { formatISK, formatNumber } from "@industry-tool/utils/formatting";
import Box from "@mui/material/Box";
import Chip from "@mui/material/Chip";
import IconButton from "@mui/material/IconButton";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Typography from "@mui/material/Typography";
import CircularProgress from "@mui/material/CircularProgress";
import CancelIcon from "@mui/icons-material/Cancel";
import Tooltip from "@mui/material/Tooltip";
import PersonIcon from "@mui/icons-material/Person";
import CorporateFareIcon from "@mui/icons-material/CorporateFare";

type Props = {
  entries: IndustryJobQueueEntry[];
  loading: boolean;
  onCancel: (id: number) => void;
};

function getStatusColor(status: string): "success" | "warning" | "info" | "error" | "default" {
  switch (status) {
    case "planned": return "info";
    case "active": return "success";
    case "completed": return "default";
    case "cancelled": return "error";
    default: return "default";
  }
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

function formatTimeRemaining(endDateStr: string): string {
  const end = new Date(endDateStr);
  const now = new Date();
  const diffMs = end.getTime() - now.getTime();

  if (diffMs <= 0) return "Ready";

  const totalSecs = Math.floor(diffMs / 1000);
  const days = Math.floor(totalSecs / 86400);
  const hours = Math.floor((totalSecs % 86400) / 3600);
  const minutes = Math.floor((totalSecs % 3600) / 60);
  const seconds = totalSecs % 60;

  const pad = (n: number) => n.toString().padStart(2, "0");

  if (days > 0) {
    return `${days}D ${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
  }
  return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
}

function formatEndDate(endDateStr: string): string {
  const d = new Date(endDateStr);
  const year = d.getUTCFullYear();
  const month = (d.getUTCMonth() + 1).toString().padStart(2, "0");
  const day = d.getUTCDate().toString().padStart(2, "0");
  const hours = d.getUTCHours().toString().padStart(2, "0");
  const minutes = d.getUTCMinutes().toString().padStart(2, "0");
  return `${year}.${month}.${day} ${hours}:${minutes}`;
}

export default function JobQueue({ entries, loading, onCancel }: Props) {
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
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Blueprint</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Product</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Activity</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="right">Runs</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="right">ME/TE</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Character</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="right">Est. Cost</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Est. Duration</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Finishes</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="center">Source</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Status</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Notes</TableCell>
            <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="center">Actions</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {entries.length === 0 ? (
            <TableRow>
              <TableCell colSpan={13} align="center" sx={{ py: 4, color: "#64748b" }}>
                No jobs in queue
              </TableCell>
            </TableRow>
          ) : (
            entries.map((entry) => (
              <TableRow
                key={entry.id}
                sx={{
                  "&:nth-of-type(odd)": { backgroundColor: "#0d1117" },
                  "&:nth-of-type(even)": { backgroundColor: "#12151f" },
                }}
              >
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#e2e8f0" }}>
                    {entry.blueprintName || `Type ${entry.blueprintTypeId}`}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#cbd5e1" }}>
                    {entry.productName || "-"}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#cbd5e1", textTransform: "capitalize" }}>
                    {entry.activity}
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2" sx={{ color: "#e2e8f0" }}>
                    {formatNumber(entry.runs)}
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2" sx={{ color: "#94a3b8" }}>
                    {entry.meLevel}/{entry.teLevel}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#94a3b8" }}>
                    {entry.characterName || "-"}
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Typography variant="body2" sx={{ color: "#cbd5e1" }}>
                    {entry.estimatedCost ? formatISK(entry.estimatedCost) : "-"}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#94a3b8" }}>
                    {entry.estimatedDuration ? formatDuration(entry.estimatedDuration) : "-"}
                  </Typography>
                </TableCell>
                <TableCell>
                  {entry.esiJobEndDate ? (
                    <Box>
                      <Typography variant="body2" sx={{ color: "#3b82f6", fontFamily: "monospace", fontWeight: 600 }}>
                        {formatTimeRemaining(entry.esiJobEndDate)}
                      </Typography>
                      <Typography variant="caption" sx={{ color: "#64748b" }}>
                        {formatEndDate(entry.esiJobEndDate)}
                      </Typography>
                    </Box>
                  ) : (
                    <Typography variant="body2" sx={{ color: "#64748b" }}>-</Typography>
                  )}
                </TableCell>
                <TableCell align="center">
                  {entry.esiJobSource ? (
                    <Tooltip title={entry.esiJobSource === "corporation" ? "Corporation Job" : "Character Job"}>
                      {entry.esiJobSource === "corporation" ? (
                        <CorporateFareIcon sx={{ color: "#f59e0b", fontSize: 18 }} />
                      ) : (
                        <PersonIcon sx={{ color: "#94a3b8", fontSize: 18 }} />
                      )}
                    </Tooltip>
                  ) : (
                    <Typography variant="body2" sx={{ color: "#64748b" }}>-</Typography>
                  )}
                </TableCell>
                <TableCell>
                  <Chip
                    label={entry.status}
                    color={getStatusColor(entry.status)}
                    size="small"
                    variant="outlined"
                  />
                </TableCell>
                <TableCell>
                  <Typography variant="body2" sx={{ color: "#94a3b8", maxWidth: 150, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                    {entry.notes || "-"}
                  </Typography>
                </TableCell>
                <TableCell align="center">
                  {(entry.status === "planned" || entry.status === "active") && (
                    <IconButton
                      size="small"
                      onClick={() => onCancel(entry.id)}
                      sx={{ color: "#ef4444" }}
                      title="Cancel job"
                    >
                      <CancelIcon fontSize="small" />
                    </IconButton>
                  )}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </TableContainer>
  );
}

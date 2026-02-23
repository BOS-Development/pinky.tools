import { useState, useEffect, useRef, useCallback } from "react";
import { useSession } from "next-auth/react";
import { PlanRun } from "@industry-tool/client/data/models";
import Navbar from "@industry-tool/components/Navbar";
import Loading from "@industry-tool/components/loading";
import Container from "@mui/material/Container";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import Chip from "@mui/material/Chip";
import IconButton from "@mui/material/IconButton";
import CancelIcon from "@mui/icons-material/Cancel";
import Snackbar from "@mui/material/Snackbar";
import Alert from "@mui/material/Alert";

function getStatusColor(
  status: string,
): "success" | "warning" | "info" | "error" | "default" {
  switch (status) {
    case "pending":
      return "info";
    case "in_progress":
      return "warning";
    case "completed":
      return "default";
    default:
      return "default";
  }
}

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
  if (!run.jobSummary || run.jobSummary.total === 0) return "—";
  const { completed, active, total } = run.jobSummary;
  return `${completed}/${total} done${active > 0 ? `, ${active} active` : ""}`;
}

export default function PlanRunsList() {
  const { data: session } = useSession();
  const [runs, setRuns] = useState<PlanRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [snackbar, setSnackbar] = useState<{
    open: boolean;
    message: string;
    severity: "success" | "error";
  }>({ open: false, message: "", severity: "success" });
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
        setSnackbar({
          open: true,
          message: `Cancelled ${data.jobsCancelled} planned job${data.jobsCancelled !== 1 ? "s" : ""}`,
          severity: "success",
        });
        fetchRuns();
      } else {
        setSnackbar({
          open: true,
          message: "Failed to cancel plan run",
          severity: "error",
        });
      }
    } catch (err) {
      console.error("Failed to cancel plan run:", err);
      setSnackbar({
        open: true,
        message: "Failed to cancel plan run",
        severity: "error",
      });
    }
  };

  return (
    <>
      <Navbar />
      <Container maxWidth="xl" sx={{ mt: 2, mb: 4 }}>
        <Box sx={{ mb: 2 }}>
          <Typography
            variant="h5"
            sx={{ color: "#e2e8f0", fontWeight: 600 }}
          >
            Plan Runs
          </Typography>
        </Box>

        {loading ? (
          <Loading />
        ) : runs.length === 0 ? (
          <Paper
            sx={{
              backgroundColor: "#12151f",
              p: 4,
              textAlign: "center",
            }}
          >
            <Typography sx={{ color: "#64748b" }}>
              No plan runs yet. Generate jobs from a production plan to create a
              run.
            </Typography>
          </Paper>
        ) : (
          <TableContainer
            component={Paper}
            sx={{ backgroundColor: "#12151f" }}
          >
            <Table size="small">
              <TableHead>
                <TableRow sx={{ backgroundColor: "#0f1219" }}>
                  <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                    Product
                  </TableCell>
                  <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                    Plan
                  </TableCell>
                  <TableCell
                    sx={{ color: "#94a3b8", fontWeight: 600 }}
                    align="right"
                  >
                    Qty
                  </TableCell>
                  <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                    Status
                  </TableCell>
                  <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                    Jobs
                  </TableCell>
                  <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                    Created
                  </TableCell>
                  <TableCell
                    sx={{ color: "#94a3b8", fontWeight: 600 }}
                    align="center"
                  >
                    Actions
                  </TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {runs.map((run, idx) => (
                  <TableRow
                    key={run.id}
                    sx={{
                      backgroundColor:
                        idx % 2 === 0 ? "#12151f" : "#0f1219",
                      "&:hover": { backgroundColor: "#1a1d2e" },
                    }}
                  >
                    <TableCell>
                      <Box
                        sx={{
                          display: "flex",
                          alignItems: "center",
                          gap: 1,
                        }}
                      >
                        <Typography sx={{ color: "#e2e8f0", fontSize: 14 }}>
                          {run.productName || "—"}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell sx={{ color: "#cbd5e1", fontSize: 14 }}>
                      {run.planName || `Plan #${run.planId}`}
                    </TableCell>
                    <TableCell
                      align="right"
                      sx={{ color: "#cbd5e1", fontSize: 14 }}
                    >
                      {run.quantity}
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={formatStatus(run.status)}
                        color={getStatusColor(run.status)}
                        size="small"
                        variant="outlined"
                      />
                    </TableCell>
                    <TableCell sx={{ color: "#94a3b8", fontSize: 13 }}>
                      {formatJobProgress(run)}
                    </TableCell>
                    <TableCell sx={{ color: "#94a3b8", fontSize: 13 }}>
                      {new Date(run.createdAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell align="center">
                      {run.status !== "completed" &&
                        run.jobSummary &&
                        run.jobSummary.planned > 0 && (
                          <IconButton
                            size="small"
                            onClick={() => handleCancel(run.id)}
                            sx={{ color: "#ef4444" }}
                            title="Cancel planned jobs"
                          >
                            <CancelIcon fontSize="small" />
                          </IconButton>
                        )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </Container>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
      >
        <Alert
          severity={snackbar.severity}
          onClose={() => setSnackbar({ ...snackbar, open: false })}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </>
  );
}

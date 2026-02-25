import React, { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Button,
  IconButton,
  Box,
  Chip,
  CircularProgress,
  Typography,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import EditIcon from "@mui/icons-material/Edit";
import DeleteIcon from "@mui/icons-material/Delete";
import { JFRoute } from "../../pages/transport";
import { JFRouteDialog } from "./JFRouteDialog";

interface Props {
  routes: JFRoute[];
  loading: boolean;
  onRefresh: () => void;
}

export function JFRoutesList({ routes, loading, onRefresh }: Props) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editRoute, setEditRoute] = useState<JFRoute | null>(null);

  const handleAdd = () => {
    setEditRoute(null);
    setDialogOpen(true);
  };

  const handleEdit = (route: JFRoute) => {
    setEditRoute(route);
    setDialogOpen(true);
  };

  const handleDelete = async (route: JFRoute) => {
    if (!confirm(`Delete route "${route.name}"?`)) return;
    try {
      await fetch(`/api/transport/jf-routes/${route.id}`, { method: "DELETE" });
      onRefresh();
    } catch (error) {
      console.error("Failed to delete route:", error);
    }
  };

  const handleDialogClose = (saved: boolean) => {
    setDialogOpen(false);
    if (saved) onRefresh();
  };

  if (loading) {
    return (
      <Box sx={{ display: "flex", justifyContent: "center", py: 4 }}>
        <CircularProgress size={32} />
      </Box>
    );
  }

  return (
    <>
      <Box sx={{ display: "flex", justifyContent: "flex-end", mb: 2 }}>
        <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd} size="small">
          Add JF Route
        </Button>
      </Box>

      <TableContainer>
        <Table size="small">
          <TableHead>
            <TableRow sx={{ backgroundColor: "#0f1219" }}>
              <TableCell>Name</TableCell>
              <TableCell>Origin</TableCell>
              <TableCell>Destination</TableCell>
              <TableCell align="right">Total Distance (LY)</TableCell>
              <TableCell align="right">Waypoints</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {routes.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} align="center" sx={{ py: 4, color: "#94a3b8" }}>
                  <Typography variant="body2">No JF routes configured</Typography>
                </TableCell>
              </TableRow>
            ) : (
              routes.map((r) => (
                <TableRow
                  key={r.id}
                  sx={{ "&:hover": { backgroundColor: "rgba(59, 130, 246, 0.05)" } }}
                >
                  <TableCell sx={{ fontWeight: 500 }}>{r.name}</TableCell>
                  <TableCell>{r.originSystemName || r.originSystemId}</TableCell>
                  <TableCell>{r.destinationSystemName || r.destinationSystemId}</TableCell>
                  <TableCell align="right">{r.totalDistanceLy.toFixed(2)} LY</TableCell>
                  <TableCell align="right">
                    <Box sx={{ display: "flex", gap: 0.5, justifyContent: "flex-end", flexWrap: "wrap" }}>
                      {(r.waypoints || []).map((wp) => (
                        <Chip
                          key={wp.id}
                          label={wp.systemName || wp.systemId}
                          size="small"
                          sx={{
                            backgroundColor: "rgba(139, 92, 246, 0.15)",
                            color: "#8b5cf6",
                            fontSize: "0.7rem",
                          }}
                        />
                      ))}
                    </Box>
                  </TableCell>
                  <TableCell align="center">
                    <IconButton size="small" onClick={() => handleEdit(r)} sx={{ color: "#94a3b8" }}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" onClick={() => handleDelete(r)} sx={{ color: "#ef4444" }}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <JFRouteDialog
        open={dialogOpen}
        onClose={handleDialogClose}
        route={editRoute}
      />
    </>
  );
}

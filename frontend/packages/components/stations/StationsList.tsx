import { useState } from "react";
import { UserStation } from "@industry-tool/client/data/models";
import StationDialog from "./StationDialog";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Button from "@mui/material/Button";
import Chip from "@mui/material/Chip";
import CircularProgress from "@mui/material/CircularProgress";
import Box from "@mui/material/Box";
import IconButton from "@mui/material/IconButton";
import EditIcon from "@mui/icons-material/Edit";
import DeleteIcon from "@mui/icons-material/Delete";
import AddIcon from "@mui/icons-material/Add";

interface Props {
  stations: UserStation[];
  loading: boolean;
  onRefresh: () => void;
}

const getSecurityColor = (security: string) => {
  switch (security) {
    case "high": return "#10b981";
    case "low": return "#f59e0b";
    case "null": return "#ef4444";
    default: return "#94a3b8";
  }
};

const getCategoryColor = (category: string) => {
  switch (category) {
    case "ship": return "#3b82f6";
    case "component": return "#8b5cf6";
    case "equipment": return "#10b981";
    case "ammo": return "#f59e0b";
    case "drone": return "#06b6d4";
    case "reaction": return "#ec4899";
    case "reprocessing": return "#f97316";
    default: return "#94a3b8";
  }
};

export default function StationsList({ stations, loading, onRefresh }: Props) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editStation, setEditStation] = useState<UserStation | null>(null);

  const handleAdd = () => {
    setEditStation(null);
    setDialogOpen(true);
  };

  const handleEdit = (station: UserStation) => {
    setEditStation(station);
    setDialogOpen(true);
  };

  const handleDelete = async (station: UserStation) => {
    if (!confirm(`Delete station "${station.stationName}"?`)) return;

    try {
      const res = await fetch(`/api/stations/user-stations/${station.id}`, {
        method: "DELETE",
      });
      if (res.ok) {
        onRefresh();
      }
    } catch (err) {
      console.error("Failed to delete station:", err);
    }
  };

  const handleDialogClose = (saved: boolean) => {
    setDialogOpen(false);
    setEditStation(null);
    if (saved) {
      onRefresh();
    }
  };

  if (loading) {
    return (
      <Box sx={{ display: "flex", justifyContent: "center", py: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <>
      <Box sx={{ mb: 2 }}>
        <Button variant="contained" startIcon={<AddIcon />} onClick={handleAdd}>
          Add Station
        </Button>
      </Box>

      <TableContainer>
        <Table size="small">
          <TableHead>
            <TableRow sx={{ backgroundColor: "#0f1219" }}>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Station</TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>System</TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Security</TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Structure</TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Activities</TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Rigs</TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="right">Tax</TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {stations.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} align="center" sx={{ color: "#64748b", py: 4 }}>
                  No preferred stations configured. Click &quot;Add Station&quot; to get started.
                </TableCell>
              </TableRow>
            ) : (
              stations.map((station) => (
                <TableRow key={station.id} sx={{ "&:hover": { backgroundColor: "rgba(59, 130, 246, 0.05)" } }}>
                  <TableCell sx={{ color: "#e2e8f0" }}>{station.stationName}</TableCell>
                  <TableCell sx={{ color: "#94a3b8" }}>{station.solarSystemName}</TableCell>
                  <TableCell>
                    <Chip
                      label={station.security}
                      size="small"
                      sx={{
                        backgroundColor: `${getSecurityColor(station.security || "")}20`,
                        color: getSecurityColor(station.security || ""),
                        fontWeight: 600,
                        textTransform: "capitalize",
                      }}
                    />
                  </TableCell>
                  <TableCell sx={{ color: "#94a3b8", textTransform: "capitalize" }}>
                    {station.structure}
                  </TableCell>
                  <TableCell>
                    <Box sx={{ display: "flex", gap: 0.5, flexWrap: "wrap" }}>
                      {station.activities.map((activity) => (
                        <Chip
                          key={activity}
                          label={activity}
                          size="small"
                          sx={{
                            backgroundColor: activity === "manufacturing" ? "#3b82f620" : "#ec489920",
                            color: activity === "manufacturing" ? "#3b82f6" : "#ec4899",
                            textTransform: "capitalize",
                            fontSize: "0.7rem",
                          }}
                        />
                      ))}
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Box sx={{ display: "flex", gap: 0.5, flexWrap: "wrap" }}>
                      {station.rigs.map((rig) => (
                        <Chip
                          key={rig.id}
                          label={`${rig.category} ${rig.tier.toUpperCase()}`}
                          size="small"
                          sx={{
                            backgroundColor: `${getCategoryColor(rig.category)}20`,
                            color: getCategoryColor(rig.category),
                            fontSize: "0.7rem",
                          }}
                        />
                      ))}
                      {station.rigs.length === 0 && (
                        <span style={{ color: "#64748b", fontSize: "0.8rem" }}>None</span>
                      )}
                    </Box>
                  </TableCell>
                  <TableCell align="right" sx={{ color: "#94a3b8" }}>
                    {station.facilityTax}%
                  </TableCell>
                  <TableCell align="right">
                    <IconButton size="small" onClick={() => handleEdit(station)} sx={{ color: "#3b82f6" }}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" onClick={() => handleDelete(station)} sx={{ color: "#ef4444" }}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <StationDialog
        open={dialogOpen}
        station={editStation}
        onClose={handleDialogClose}
      />
    </>
  );
}

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
import { TransportProfile } from "../../pages/transport";
import { TransportProfileDialog } from "./TransportProfileDialog";
import { formatNumber } from "../../utils/formatting";

interface Props {
  profiles: TransportProfile[];
  loading: boolean;
  onRefresh: () => void;
}

const getMethodLabel = (method: string) => {
  const labels: Record<string, string> = {
    freighter: "Freighter",
    jump_freighter: "Jump Freighter",
    dst: "DST",
    blockade_runner: "Blockade Runner",
  };
  return labels[method] || method;
};

const getMethodColor = (method: string) => {
  const colors: Record<string, string> = {
    freighter: "#3b82f6",
    jump_freighter: "#8b5cf6",
    dst: "#06b6d4",
    blockade_runner: "#f59e0b",
  };
  return colors[method] || "#94a3b8";
};

export function TransportProfilesList({ profiles, loading, onRefresh }: Props) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editProfile, setEditProfile] = useState<TransportProfile | null>(null);

  const handleAdd = () => {
    setEditProfile(null);
    setDialogOpen(true);
  };

  const handleEdit = (profile: TransportProfile) => {
    setEditProfile(profile);
    setDialogOpen(true);
  };

  const handleDelete = async (profile: TransportProfile) => {
    if (!confirm(`Delete profile "${profile.name}"?`)) return;
    try {
      await fetch(`/api/transport/profiles/${profile.id}`, { method: "DELETE" });
      onRefresh();
    } catch (error) {
      console.error("Failed to delete profile:", error);
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
          Add Profile
        </Button>
      </Box>

      <TableContainer>
        <Table size="small">
          <TableHead>
            <TableRow sx={{ backgroundColor: "#0f1219" }}>
              <TableCell>Name</TableCell>
              <TableCell>Method</TableCell>
              <TableCell>Character</TableCell>
              <TableCell align="right">Cargo (m3)</TableCell>
              <TableCell align="right">Rate/m3/Jump</TableCell>
              <TableCell align="right">Collateral Rate</TableCell>
              <TableCell>Route Pref</TableCell>
              <TableCell>Default</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {profiles.length === 0 ? (
              <TableRow>
                <TableCell colSpan={9} align="center" sx={{ py: 4, color: "#94a3b8" }}>
                  <Typography variant="body2">No transport profiles configured</Typography>
                </TableCell>
              </TableRow>
            ) : (
              profiles.map((p) => (
                <TableRow
                  key={p.id}
                  sx={{ "&:hover": { backgroundColor: "rgba(59, 130, 246, 0.05)" } }}
                >
                  <TableCell sx={{ fontWeight: 500 }}>{p.name}</TableCell>
                  <TableCell>
                    <Chip
                      label={getMethodLabel(p.transportMethod)}
                      size="small"
                      sx={{
                        backgroundColor: `${getMethodColor(p.transportMethod)}20`,
                        color: getMethodColor(p.transportMethod),
                        fontWeight: 500,
                      }}
                    />
                  </TableCell>
                  <TableCell>{p.characterName || "—"}</TableCell>
                  <TableCell align="right">{formatNumber(p.cargoM3)}</TableCell>
                  <TableCell align="right">{formatNumber(p.ratePerM3PerJump)}</TableCell>
                  <TableCell align="right">{(p.collateralRate * 100).toFixed(1)}%</TableCell>
                  <TableCell>{p.routePreference}</TableCell>
                  <TableCell>
                    {p.isDefault && (
                      <Chip
                        label="Default"
                        size="small"
                        sx={{
                          backgroundColor: "rgba(16, 185, 129, 0.15)",
                          color: "#10b981",
                          fontWeight: 500,
                        }}
                      />
                    )}
                  </TableCell>
                  <TableCell align="center">
                    <IconButton size="small" onClick={() => handleEdit(p)} sx={{ color: "#94a3b8" }}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" onClick={() => handleDelete(p)} sx={{ color: "#ef4444" }}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <TransportProfileDialog
        open={dialogOpen}
        onClose={handleDialogClose}
        profile={editProfile}
      />
    </>
  );
}

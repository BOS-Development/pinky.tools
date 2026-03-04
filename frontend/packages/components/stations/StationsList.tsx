import { useState } from "react";
import { UserStation } from "@industry-tool/client/data/models";
import StationDialog from "./StationDialog";
import { Plus, Pencil, Trash2, Loader2, Building2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Table, TableHeader, TableBody, TableRow, TableHead, TableCell,
} from '@/components/ui/table';

interface Props {
  stations: UserStation[];
  loading: boolean;
  onRefresh: () => void;
}

const getSecurityColor = (security: string) => {
  switch (security) {
    case "high": return "text-teal-success bg-teal-success/10 border-teal-success/30";
    case "low": return "text-amber-manufacturing bg-amber-manufacturing/10 border-amber-manufacturing/30";
    case "null": return "text-rose-danger bg-rose-danger/10 border-rose-danger/30";
    default: return "text-text-secondary bg-[#94a3b8]/10 border-[#94a3b8]/30";
  }
};

const getActivityColor = (activity: string) => {
  return activity === "manufacturing"
    ? "text-[var(--color-primary-cyan)] bg-[var(--color-primary-cyan)]/10 border-[var(--color-primary-cyan)]/30"
    : "text-[#ec4899] bg-[#ec4899]/10 border-[#ec4899]/30";
};

const getCategoryColor = (category: string) => {
  const colors: Record<string, string> = {
    ship: "text-primary bg-primary/10 border-primary/30",
    component: "text-[#8b5cf6] bg-[#8b5cf6]/10 border-[#8b5cf6]/30",
    equipment: "text-teal-success bg-teal-success/10 border-teal-success/30",
    ammo: "text-amber-manufacturing bg-amber-manufacturing/10 border-amber-manufacturing/30",
    drone: "text-[#06b6d4] bg-[#06b6d4]/10 border-[#06b6d4]/30",
    reaction: "text-[#ec4899] bg-[#ec4899]/10 border-[#ec4899]/30",
    reprocessing: "text-[#f97316] bg-[#f97316]/10 border-[#f97316]/30",
  };
  return colors[category] || "text-text-secondary bg-[#94a3b8]/10 border-[#94a3b8]/30";
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
      <div className="flex justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary-cyan)]" />
      </div>
    );
  }

  return (
    <>
      <div className="mb-4">
        <Button onClick={handleAdd}>
          <Plus className="h-4 w-4 mr-2" />
          Add Station
        </Button>
      </div>

      <Table>
        <TableHeader>
          <TableRow className="bg-[var(--color-bg-void)]">
            <TableHead>Station</TableHead>
            <TableHead>System</TableHead>
            <TableHead>Security</TableHead>
            <TableHead>Structure</TableHead>
            <TableHead>Activities</TableHead>
            <TableHead>Rigs</TableHead>
            <TableHead className="text-right">Tax</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {stations.length === 0 ? (
            <TableRow>
              <TableCell colSpan={8} className="p-0 border-0">
                <div className="empty-state">
                  <Building2 className="empty-state-icon" />
                  <p className="empty-state-title">No preferred stations configured. Click &quot;Add Station&quot; to get started.</p>
                </div>
              </TableCell>
            </TableRow>
          ) : (
            stations.map((station) => (
              <TableRow key={station.id}>
                <TableCell className="text-[var(--color-text-emphasis)] font-medium">{station.stationName}</TableCell>
                <TableCell className="text-[var(--color-text-secondary)]">{station.solarSystemName}</TableCell>
                <TableCell>
                  <Badge className={`${getSecurityColor(station.security || "")} capitalize font-semibold border`}>
                    {station.security}
                  </Badge>
                </TableCell>
                <TableCell className="text-[var(--color-text-secondary)] capitalize">{station.structure}</TableCell>
                <TableCell>
                  <div className="flex gap-1 flex-wrap">
                    {station.activities.map((activity) => (
                      <Badge key={activity} className={`${getActivityColor(activity)} capitalize text-[10px] border`}>
                        {activity}
                      </Badge>
                    ))}
                  </div>
                </TableCell>
                <TableCell>
                  <div className="flex gap-1 flex-wrap">
                    {station.rigs.map((rig) => (
                      <Badge key={rig.id} className={`${getCategoryColor(rig.category)} text-[10px] border`}>
                        {rig.category} {rig.tier.toUpperCase()}
                      </Badge>
                    ))}
                    {station.rigs.length === 0 && (
                      <span className="text-[var(--color-text-muted)] text-xs">None</span>
                    )}
                  </div>
                </TableCell>
                <TableCell className="text-right text-[var(--color-text-secondary)]">
                  {station.facilityTax}%
                </TableCell>
                <TableCell className="text-right">
                  <Button variant="ghost" size="icon" onClick={() => handleEdit(station)} className="text-[var(--color-primary-cyan)]" title="Edit Station">
                    <Pencil className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon" onClick={() => handleDelete(station)} className="text-[var(--color-danger-rose)] hover:text-[var(--color-danger-rose)]" title="Delete Station">
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>

      <StationDialog
        open={dialogOpen}
        station={editStation}
        onClose={handleDialogClose}
      />
    </>
  );
}

import React, { useState } from "react";
import { Plus, Pencil, Trash2, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
    freighter: "var(--color-primary-cyan)",
    jump_freighter: "var(--color-category-violet)",
    dst: "var(--color-category-teal)",
    blockade_runner: "var(--color-manufacturing-amber)",
  };
  return colors[method] || "var(--color-text-secondary)";
};

const getMethodBgColor = (method: string) => {
  const colors: Record<string, string> = {
    freighter: "var(--color-info-tint)",
    jump_freighter: "var(--color-neutral-tint)",
    dst: "var(--color-neutral-tint)",
    blockade_runner: "var(--color-warning-tint)",
  };
  return colors[method] || "var(--color-neutral-tint)";
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
      <div className="flex justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <>
      <div className="flex justify-end mb-3">
        <Button size="sm" onClick={handleAdd}>
          <Plus className="h-4 w-4 mr-2" />
          Add Profile
        </Button>
      </div>

      <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow className="bg-background-void hover:bg-background-void">
              <TableHead>Name</TableHead>
              <TableHead>Method</TableHead>
              <TableHead>Character</TableHead>
              <TableHead className="text-right">Cargo (m3)</TableHead>
              <TableHead className="text-right">Rate/m3/Jump</TableHead>
              <TableHead className="text-right">Collateral Rate</TableHead>
              <TableHead>Route Pref</TableHead>
              <TableHead>Default</TableHead>
              <TableHead className="text-center">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {profiles.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={9}
                  className="text-center py-8 text-text-secondary"
                >
                  No transport profiles configured
                </TableCell>
              </TableRow>
            ) : (
              profiles.map((p) => (
                <TableRow key={p.id} className="hover:bg-interactive-hover">
                  <TableCell className="font-medium text-text-emphasis">{p.name}</TableCell>
                  <TableCell>
                    <span
                      className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
                      style={{
                        backgroundColor: getMethodBgColor(p.transportMethod),
                        color: getMethodColor(p.transportMethod),
                      }}
                    >
                      {getMethodLabel(p.transportMethod)}
                    </span>
                  </TableCell>
                  <TableCell className="text-text-secondary">{p.characterName || "—"}</TableCell>
                  <TableCell className="text-right">{formatNumber(p.cargoM3)}</TableCell>
                  <TableCell className="text-right">{formatNumber(p.ratePerM3PerJump)}</TableCell>
                  <TableCell className="text-right">{(p.collateralRate * 100).toFixed(1)}%</TableCell>
                  <TableCell className="text-text-secondary">{p.routePreference}</TableCell>
                  <TableCell>
                    {p.isDefault && (
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-teal-success/15 text-teal-success">
                        Default
                      </span>
                    )}
                  </TableCell>
                  <TableCell className="text-center">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 text-text-secondary hover:text-text-emphasis"
                      onClick={() => handleEdit(p)}
                    >
                      <Pencil className="h-3.5 w-3.5" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 text-rose-danger hover:text-rose-danger hover:bg-rose-danger/10"
                      onClick={() => handleDelete(p)}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <TransportProfileDialog
        open={dialogOpen}
        onClose={handleDialogClose}
        profile={editProfile}
      />
    </>
  );
}

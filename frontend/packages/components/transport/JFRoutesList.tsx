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
      <div className="flex justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-[#00d4ff]" />
      </div>
    );
  }

  return (
    <>
      <div className="flex justify-end mb-3">
        <Button size="sm" onClick={handleAdd}>
          <Plus className="h-4 w-4 mr-2" />
          Add JF Route
        </Button>
      </div>

      <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow className="bg-[#0f1219] hover:bg-[#0f1219]">
              <TableHead>Name</TableHead>
              <TableHead>Origin</TableHead>
              <TableHead>Destination</TableHead>
              <TableHead className="text-right">Total Distance (LY)</TableHead>
              <TableHead className="text-right">Waypoints</TableHead>
              <TableHead className="text-center">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {routes.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className="text-center py-8 text-[#94a3b8]"
                >
                  No JF routes configured
                </TableCell>
              </TableRow>
            ) : (
              routes.map((r) => (
                <TableRow key={r.id} className="hover:bg-[rgba(0,212,255,0.05)]">
                  <TableCell className="font-medium text-[#e2e8f0]">{r.name}</TableCell>
                  <TableCell className="text-[#94a3b8]">{r.originSystemName || r.originSystemId}</TableCell>
                  <TableCell className="text-[#94a3b8]">{r.destinationSystemName || r.destinationSystemId}</TableCell>
                  <TableCell className="text-right">{r.totalDistanceLy.toFixed(2)} LY</TableCell>
                  <TableCell className="text-right">
                    <div className="flex gap-1 justify-end flex-wrap">
                      {(r.waypoints || []).map((wp) => (
                        <span
                          key={wp.id}
                          className="inline-flex items-center px-1.5 py-0.5 rounded text-[0.7rem] font-medium bg-[rgba(139,92,246,0.15)] text-[#8b5cf6]"
                        >
                          {wp.systemName || wp.systemId}
                        </span>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell className="text-center">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 text-[#94a3b8] hover:text-[#e2e8f0]"
                      onClick={() => handleEdit(r)}
                    >
                      <Pencil className="h-3.5 w-3.5" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 text-[#ef4444] hover:text-[#ef4444] hover:bg-[rgba(239,68,68,0.1)]"
                      onClick={() => handleDelete(r)}
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

      <JFRouteDialog
        open={dialogOpen}
        onClose={handleDialogClose}
        route={editRoute}
      />
    </>
  );
}

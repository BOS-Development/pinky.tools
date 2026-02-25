import React, { useCallback, useEffect, useRef, useState } from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Box,
  IconButton,
  Typography,
  Autocomplete,
  CircularProgress,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import { JFRoute } from "../../pages/transport";

interface SolarSystemOption {
  id: number;
  name: string;
  security: number;
}

interface WaypointEntry {
  sequence: number;
  system: SolarSystemOption | null;
}

interface Props {
  open: boolean;
  onClose: (saved: boolean) => void;
  route: JFRoute | null;
}

const getSecurityColor = (sec: number) => {
  if (sec >= 0.5) return "#10b981";
  if (sec > 0.0) return "#f59e0b";
  return "#ef4444";
};

export function JFRouteDialog({ open, onClose, route }: Props) {
  const isEdit = !!route;
  const [saving, setSaving] = useState(false);
  const [name, setName] = useState("");
  const [originSystem, setOriginSystem] = useState<SolarSystemOption | null>(null);
  const [destinationSystem, setDestinationSystem] = useState<SolarSystemOption | null>(null);
  const [waypoints, setWaypoints] = useState<WaypointEntry[]>([]);

  // Search state for origin/destination
  const [originOptions, setOriginOptions] = useState<SolarSystemOption[]>([]);
  const [originLoading, setOriginLoading] = useState(false);
  const [destOptions, setDestOptions] = useState<SolarSystemOption[]>([]);
  const [destLoading, setDestLoading] = useState(false);

  // Per-waypoint search state
  const [waypointOptions, setWaypointOptions] = useState<Record<number, SolarSystemOption[]>>({});
  const [waypointLoading, setWaypointLoading] = useState<Record<number, boolean>>({});

  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const searchSystems = useCallback(async (query: string): Promise<SolarSystemOption[]> => {
    if (!query || query.length < 2) return [];
    try {
      const res = await fetch(`/api/transport/systems/search?q=${encodeURIComponent(query)}`);
      if (res.ok) {
        const data = await res.json();
        return (data || []).map((s: any) => ({
          id: s.id,
          name: s.name,
          security: s.security,
        }));
      }
    } catch (err) {
      console.error("Failed to search systems:", err);
    }
    return [];
  }, []);

  const debouncedSearch = useCallback(
    (query: string, setOptions: (opts: SolarSystemOption[]) => void, setLoading: (l: boolean) => void) => {
      if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current);
      searchTimeoutRef.current = setTimeout(async () => {
        setLoading(true);
        const results = await searchSystems(query);
        setOptions(results);
        setLoading(false);
      }, 300);
    },
    [searchSystems],
  );

  useEffect(() => {
    if (open && route) {
      setName(route.name);
      setOriginSystem({
        id: route.originSystemId,
        name: route.originSystemName || String(route.originSystemId),
        security: 0,
      });
      setDestinationSystem({
        id: route.destinationSystemId,
        name: route.destinationSystemName || String(route.destinationSystemId),
        security: 0,
      });
      setWaypoints(
        (route.waypoints || []).map((wp) => ({
          sequence: wp.sequence,
          system: {
            id: wp.systemId,
            name: wp.systemName || String(wp.systemId),
            security: 0,
          },
        })),
      );
    } else if (open) {
      setName("");
      setOriginSystem(null);
      setDestinationSystem(null);
      setWaypoints([
        { sequence: 0, system: null },
        { sequence: 1, system: null },
      ]);
    }
    setOriginOptions([]);
    setDestOptions([]);
    setWaypointOptions({});
    setWaypointLoading({});
  }, [open, route]);

  const handleAddWaypoint = () => {
    setWaypoints([...waypoints, { sequence: waypoints.length, system: null }]);
  };

  const handleRemoveWaypoint = (index: number) => {
    if (waypoints.length <= 2) return;
    const updated = waypoints.filter((_, i) => i !== index).map((wp, i) => ({ ...wp, sequence: i }));
    setWaypoints(updated);
  };

  const handleWaypointSystemChange = (index: number, system: SolarSystemOption | null) => {
    const updated = [...waypoints];
    updated[index] = { ...updated[index], system };
    setWaypoints(updated);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      const payload = {
        name,
        originSystemId: originSystem?.id,
        destinationSystemId: destinationSystem?.id,
        waypoints: waypoints.map((wp) => ({
          sequence: wp.sequence,
          systemId: wp.system?.id,
        })),
      };

      const url = isEdit ? `/api/transport/jf-routes/${route!.id}` : "/api/transport/jf-routes";
      const method = isEdit ? "PUT" : "POST";

      const res = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        onClose(true);
      }
    } catch (error) {
      console.error("Failed to save JF route:", error);
    } finally {
      setSaving(false);
    }
  };

  const canSave = name && originSystem && destinationSystem && waypoints.length >= 2 && waypoints.every((wp) => wp.system);

  const renderSystemAutocomplete = (
    label: string,
    value: SolarSystemOption | null,
    onChange: (v: SolarSystemOption | null) => void,
    options: SolarSystemOption[],
    loading: boolean,
    onInputChange: (v: string) => void,
  ) => (
    <Autocomplete
      value={value}
      onChange={(_, newValue) => onChange(newValue)}
      onInputChange={(_, inputValue) => onInputChange(inputValue)}
      options={options}
      getOptionLabel={(option) => option.name}
      isOptionEqualToValue={(a, b) => a.id === b.id}
      loading={loading}
      filterOptions={(x) => x}
      size="small"
      renderOption={(props, option) => (
        <Box component="li" {...props}>
          <Box sx={{ display: "flex", gap: 1, alignItems: "center" }}>
            <Typography variant="body2">{option.name}</Typography>
            <Typography
              variant="caption"
              sx={{ color: getSecurityColor(option.security ?? 0) }}
            >
              ({(option.security ?? 0).toFixed(1)})
            </Typography>
          </Box>
        </Box>
      )}
      renderInput={(params) => (
        <TextField
          {...params}
          label={label}
          placeholder="Search for a system..."
          InputProps={{
            ...params.InputProps,
            endAdornment: (
              <>
                {loading ? <CircularProgress color="inherit" size={20} /> : null}
                {params.InputProps.endAdornment}
              </>
            ),
          }}
        />
      )}
    />
  );

  return (
    <Dialog
      open={open}
      onClose={() => onClose(false)}
      maxWidth="sm"
      fullWidth
      PaperProps={{ sx: { backgroundColor: "#12151f", backgroundImage: "none" } }}
    >
      <DialogTitle>{isEdit ? "Edit JF Route" : "Add JF Route"}</DialogTitle>
      <DialogContent>
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2, mt: 1 }}>
          <TextField
            label="Route Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            fullWidth
            size="small"
          />

          {renderSystemAutocomplete(
            "Origin System",
            originSystem,
            setOriginSystem,
            originOptions,
            originLoading,
            (v) => debouncedSearch(v, setOriginOptions, setOriginLoading),
          )}

          {renderSystemAutocomplete(
            "Destination System",
            destinationSystem,
            setDestinationSystem,
            destOptions,
            destLoading,
            (v) => debouncedSearch(v, setDestOptions, setDestLoading),
          )}

          <Box>
            <Box sx={{ display: "flex", alignItems: "center", justifyContent: "space-between", mb: 1 }}>
              <Typography variant="subtitle2" sx={{ color: "#94a3b8" }}>
                Waypoints (cyno systems in order)
              </Typography>
              <IconButton size="small" onClick={handleAddWaypoint} sx={{ color: "#3b82f6" }}>
                <AddIcon fontSize="small" />
              </IconButton>
            </Box>
            {waypoints.map((wp, index) => (
              <Box key={index} sx={{ display: "flex", gap: 1, mb: 1, alignItems: "center" }}>
                <Typography variant="caption" sx={{ color: "#94a3b8", minWidth: 20 }}>
                  {wp.sequence}
                </Typography>
                <Box sx={{ flex: 1 }}>
                  {renderSystemAutocomplete(
                    `Waypoint #${wp.sequence}`,
                    wp.system,
                    (system) => handleWaypointSystemChange(index, system),
                    waypointOptions[index] || [],
                    waypointLoading[index] || false,
                    (v) =>
                      debouncedSearch(
                        v,
                        (opts) => setWaypointOptions((prev) => ({ ...prev, [index]: opts })),
                        (l) => setWaypointLoading((prev) => ({ ...prev, [index]: l })),
                      ),
                  )}
                </Box>
                {waypoints.length > 2 && (
                  <IconButton
                    size="small"
                    onClick={() => handleRemoveWaypoint(index)}
                    sx={{ color: "#ef4444" }}
                  >
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                )}
              </Box>
            ))}
          </Box>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => onClose(false)} sx={{ color: "#94a3b8" }}>
          Cancel
        </Button>
        <Button variant="contained" onClick={handleSave} disabled={saving || !canSave}>
          {saving ? "Saving..." : isEdit ? "Update" : "Create"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

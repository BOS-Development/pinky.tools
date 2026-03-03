import React, { useCallback, useEffect, useRef, useState } from "react";
import { Plus, Trash2, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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

interface SystemSearchDropdownProps {
  label: string;
  value: SolarSystemOption | null;
  onSelect: (option: SolarSystemOption | null) => void;
  options: SolarSystemOption[];
  loading: boolean;
  onSearch: (value: string) => void;
  displayValue: string;
  setDisplayValue: (v: string) => void;
}

function SystemSearchDropdown({
  label,
  value,
  onSelect,
  options,
  loading,
  onSearch,
  displayValue,
  setDisplayValue,
}: SystemSearchDropdownProps) {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Sync displayValue when value is cleared externally
  useEffect(() => {
    if (!value) {
      setDisplayValue("");
    }
  }, [value, setDisplayValue]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const v = e.target.value;
    setDisplayValue(v);
    onSearch(v);
    setOpen(true);
  };

  const handleSelect = (opt: SolarSystemOption) => {
    onSelect(opt);
    setDisplayValue(opt.name);
    setOpen(false);
  };

  return (
    <div className="relative" ref={containerRef}>
      <div className="flex flex-col gap-1">
        {label && <label className="text-xs text-[#94a3b8]">{label}</label>}
        <div className="relative">
          <Input
            value={displayValue}
            onChange={handleInputChange}
            onFocus={() => { if (options.length > 0) setOpen(true); }}
            placeholder="Search for a system..."
          />
          {loading && (
            <Loader2 className="absolute right-3 top-2.5 h-4 w-4 animate-spin text-[#64748b]" />
          )}
        </div>
      </div>
      {open && options.length > 0 && (
        <div className="absolute z-50 w-full mt-1 bg-[#1a1f2e] border border-[rgba(148,163,184,0.15)] rounded-sm shadow-lg max-h-48 overflow-y-auto">
          {options.map((opt) => (
            <button
              key={opt.id}
              type="button"
              className="w-full text-left px-3 py-2 hover:bg-[rgba(0,212,255,0.08)] flex items-center gap-2"
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => handleSelect(opt)}
            >
              <span className="text-sm text-[#e2e8f0]">{opt.name}</span>
              <span
                className="text-xs"
                style={{ color: getSecurityColor(opt.security ?? 0) }}
              >
                ({(opt.security ?? 0).toFixed(1)})
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

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
  const [originDisplay, setOriginDisplay] = useState("");
  const [destOptions, setDestOptions] = useState<SolarSystemOption[]>([]);
  const [destLoading, setDestLoading] = useState(false);
  const [destDisplay, setDestDisplay] = useState("");

  // Per-waypoint search state
  const [waypointOptions, setWaypointOptions] = useState<Record<number, SolarSystemOption[]>>({});
  const [waypointLoading, setWaypointLoading] = useState<Record<number, boolean>>({});
  const [waypointDisplays, setWaypointDisplays] = useState<Record<number, string>>({});

  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const searchSystems = useCallback(async (query: string): Promise<SolarSystemOption[]> => {
    if (!query || query.length < 2) return [];
    try {
      const res = await fetch(`/api/transport/systems/search?q=${encodeURIComponent(query)}`);
      if (res.ok) {
        const data = await res.json();
        return (data || []).map((s: { id: number; name: string; security: number }) => ({
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
      const origin: SolarSystemOption = {
        id: route.originSystemId,
        name: route.originSystemName || String(route.originSystemId),
        security: 0,
      };
      const dest: SolarSystemOption = {
        id: route.destinationSystemId,
        name: route.destinationSystemName || String(route.destinationSystemId),
        security: 0,
      };
      setOriginSystem(origin);
      setDestinationSystem(dest);
      setOriginDisplay(origin.name);
      setDestDisplay(dest.name);

      const wps = (route.waypoints || []).map((wp) => ({
        sequence: wp.sequence,
        system: {
          id: wp.systemId,
          name: wp.systemName || String(wp.systemId),
          security: 0,
        },
      }));
      setWaypoints(wps);

      const displays: Record<number, string> = {};
      wps.forEach((wp, index) => {
        if (wp.system) displays[index] = wp.system.name;
      });
      setWaypointDisplays(displays);
    } else if (open) {
      setName("");
      setOriginSystem(null);
      setDestinationSystem(null);
      setOriginDisplay("");
      setDestDisplay("");
      setWaypoints([
        { sequence: 0, system: null },
        { sequence: 1, system: null },
      ]);
      setWaypointDisplays({});
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
    // Reindex waypoint displays
    const newDisplays: Record<number, string> = {};
    updated.forEach((wp, i) => {
      if (wp.system) newDisplays[i] = waypointDisplays[i < index ? i : i + 1] || wp.system.name;
    });
    setWaypointDisplays(newDisplays);
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

  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose(false); }}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit JF Route" : "Add JF Route"}</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-3 pt-1 max-h-[70vh] overflow-y-auto pr-1">
          <div className="flex flex-col gap-1">
            <label className="text-xs text-[#94a3b8]">Route Name</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Route name..."
            />
          </div>

          <SystemSearchDropdown
            label="Origin System"
            value={originSystem}
            onSelect={setOriginSystem}
            options={originOptions}
            loading={originLoading}
            onSearch={(v) => debouncedSearch(v, setOriginOptions, setOriginLoading)}
            displayValue={originDisplay}
            setDisplayValue={setOriginDisplay}
          />

          <SystemSearchDropdown
            label="Destination System"
            value={destinationSystem}
            onSelect={setDestinationSystem}
            options={destOptions}
            loading={destLoading}
            onSearch={(v) => debouncedSearch(v, setDestOptions, setDestLoading)}
            displayValue={destDisplay}
            setDisplayValue={setDestDisplay}
          />

          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-[#94a3b8]">Waypoints (cyno systems in order)</span>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 text-[#00d4ff] hover:text-[#00d4ff] hover:bg-[rgba(0,212,255,0.1)]"
                onClick={handleAddWaypoint}
              >
                <Plus className="h-4 w-4" />
              </Button>
            </div>

            {waypoints.map((wp, index) => (
              <div key={index} className="flex gap-2 mb-2 items-end">
                <span className="text-xs text-[#94a3b8] min-w-5 pb-2">{wp.sequence}</span>
                <div className="flex-1">
                  <SystemSearchDropdown
                    label={`Waypoint #${wp.sequence}`}
                    value={wp.system}
                    onSelect={(system) => handleWaypointSystemChange(index, system)}
                    options={waypointOptions[index] || []}
                    loading={waypointLoading[index] || false}
                    onSearch={(v) =>
                      debouncedSearch(
                        v,
                        (opts) => setWaypointOptions((prev) => ({ ...prev, [index]: opts })),
                        (l) => setWaypointLoading((prev) => ({ ...prev, [index]: l })),
                      )
                    }
                    displayValue={waypointDisplays[index] || ""}
                    setDisplayValue={(v) => setWaypointDisplays((prev) => ({ ...prev, [index]: v }))}
                  />
                </div>
                {waypoints.length > 2 && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 text-[#ef4444] hover:text-[#ef4444] hover:bg-[rgba(239,68,68,0.1)] mb-0.5"
                    onClick={() => handleRemoveWaypoint(index)}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                )}
              </div>
            ))}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onClose(false)} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving || !canSave}>
            {saving ? "Saving..." : isEdit ? "Update" : "Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

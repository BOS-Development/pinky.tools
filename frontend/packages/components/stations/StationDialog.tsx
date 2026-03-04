import { useState, useEffect, useRef, useCallback } from "react";
import {
  UserStation,
  UserStationRig,
  UserStationService,
  ScanResult,
} from "@industry-tool/client/data/models";
import { Plus, Trash2, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from '@/components/ui/dialog';

type StationOption = {
  stationId: number;
  name: string;
  solarSystemName: string;
};

interface Props {
  open: boolean;
  station: UserStation | null;
  onClose: (saved: boolean) => void;
}

const getRigCategoriesForStructure = (structure: string): string[] => {
  if (["athanor", "tatara"].includes(structure)) {
    return ["reaction", "reprocessing"];
  }
  return ["ship", "component", "equipment", "ammo", "drone", "thukker"];
};
const getRigTiersForCategory = (category: string): string[] => {
  if (category === "component") {
    return ["t1", "t2", "thukker"];
  }
  return ["t1", "t2"];
};

const getCategoryColor = (category: string) => {
  const colors: Record<string, string> = {
    ship: "bg-primary/10 text-primary border-primary/30",
    component: "bg-category-violet/10 text-category-violet border-category-violet/30",
    equipment: "bg-teal-success/10 text-teal-success border-teal-success/30",
    ammo: "bg-amber-manufacturing/10 text-amber-manufacturing border-amber-manufacturing/30",
    drone: "bg-category-teal/10 text-category-teal border-category-teal/30",
    reaction: "bg-category-pink/10 text-category-pink border-category-pink/30",
    reprocessing: "bg-category-orange/10 text-category-orange border-category-orange/30",
    thukker: "bg-category-orange/10 text-category-orange border-category-orange/30",
  };
  return colors[category] || "bg-category-slate/10 text-text-secondary border-category-slate/30";
};

export default function StationDialog({ open, station, onClose }: Props) {
  const isEdit = !!station;

  const [stationOptions, setStationOptions] = useState<StationOption[]>([]);
  const [stationSearchLoading, setStationSearchLoading] = useState(false);
  const [selectedStation, setSelectedStation] = useState<StationOption | null>(null);
  const [stationSearchQuery, setStationSearchQuery] = useState("");

  const [structure, setStructure] = useState("raitaru");
  const [facilityTax, setFacilityTax] = useState(1.0);
  const [rigs, setRigs] = useState<{ rigName: string; category: string; tier: string }[]>([]);
  const [services, setServices] = useState<{ serviceName: string; activity: string }[]>([]);

  const [scanText, setScanText] = useState("");
  const [scanParsing, setScanParsing] = useState(false);

  const [saving, setSaving] = useState(false);

  const stationSearchTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (station) {
      setSelectedStation({
        stationId: station.stationId,
        name: station.stationName || "",
        solarSystemName: station.solarSystemName || "",
      });
      setStationSearchQuery(station.stationName || "");
      setStructure(station.structure);
      setFacilityTax(station.facilityTax);
      setRigs(
        station.rigs.map((r) => ({
          rigName: r.rigName,
          category: r.category,
          tier: r.tier,
        })),
      );
      setServices(
        station.services.map((s) => ({
          serviceName: s.serviceName,
          activity: s.activity,
        })),
      );
    } else {
      setSelectedStation(null);
      setStationSearchQuery("");
      setStructure("raitaru");
      setFacilityTax(1.0);
      setRigs([]);
      setServices([]);
      setScanText("");
    }
    setStationOptions([]);
  }, [station, open]);

  const searchStations = useCallback(async (query: string) => {
    if (!query || query.length < 2) {
      setStationOptions([]);
      return;
    }

    setStationSearchLoading(true);
    try {
      const res = await fetch(`/api/stations/search?q=${encodeURIComponent(query)}`);
      if (res.ok) {
        const data = await res.json();
        setStationOptions(data || []);
      }
    } catch (err) {
      console.error("Failed to search stations:", err);
      setStationOptions([]);
    } finally {
      setStationSearchLoading(false);
    }
  }, []);

  const handleStationSearch = (value: string) => {
    setStationSearchQuery(value);
    setSelectedStation(null);
    if (stationSearchTimeoutRef.current) {
      clearTimeout(stationSearchTimeoutRef.current);
    }
    stationSearchTimeoutRef.current = setTimeout(() => {
      searchStations(value);
    }, 300);
  };

  const handleParseScan = async () => {
    if (!scanText.trim()) return;

    setScanParsing(true);
    try {
      const res = await fetch("/api/stations/parse-scan", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ scanText: scanText.trim() }),
      });

      if (res.ok) {
        const data: ScanResult = await res.json();
        if (data.structure) {
          setStructure(data.structure);
        }
        if (data.rigs?.length > 0) {
          setRigs(
            data.rigs.map((r) => ({
              rigName: r.name,
              category: r.category,
              tier: r.tier,
            })),
          );
        }
        if (data.services?.length > 0) {
          setServices(
            data.services.map((s) => ({
              serviceName: s.name,
              activity: s.activity,
            })),
          );
        }
      }
    } catch (err) {
      console.error("Failed to parse scan:", err);
    } finally {
      setScanParsing(false);
    }
  };

  const handleAddRig = () => {
    const validCategories = getRigCategoriesForStructure(structure);
    setRigs([...rigs, { rigName: "", category: validCategories[0], tier: "t1" }]);
  };

  const handleRemoveRig = (index: number) => {
    setRigs(rigs.filter((_, i) => i !== index));
  };

  const handleRigChange = (index: number, field: string, value: string) => {
    const updated = [...rigs];
    updated[index] = { ...updated[index], [field]: value };
    if (field === "category" && value !== "component" && updated[index].tier === "thukker") {
      updated[index] = { ...updated[index], tier: "t1" };
    }
    setRigs(updated);
  };

  const handleSave = async () => {
    if (!selectedStation) return;

    setSaving(true);
    try {
      const payload = {
        stationId: selectedStation.stationId,
        structure,
        facilityTax,
        rigs: rigs.map((r) => ({
          rigName: r.rigName,
          category: r.category,
          tier: r.tier,
        })),
        services: services.map((s) => ({
          serviceName: s.serviceName,
          activity: s.activity,
        })),
      };

      const url = isEdit
        ? `/api/stations/user-stations/${station!.id}`
        : "/api/stations/user-stations";
      const method = isEdit ? "PUT" : "POST";

      const res = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        onClose(true);
      }
    } catch (err) {
      console.error("Failed to save station:", err);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose(false); }}>
      <DialogContent className="max-w-md max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit Station" : "Add Station"}</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* Station Search */}
          <div>
            <Label htmlFor="station-search">Station</Label>
            <Input
              id="station-search"
              value={stationSearchQuery}
              onChange={(e) => handleStationSearch(e.target.value)}
              placeholder="Search for a station..."
              disabled={isEdit}
            />
            {stationSearchLoading && (
              <div className="flex items-center gap-2 mt-1">
                <Loader2 className="h-3 w-3 animate-spin text-primary" />
                <span className="text-xs text-text-muted">Searching...</span>
              </div>
            )}
            {stationOptions.length > 0 && !selectedStation && (
              <div className="mt-1 border border-dim rounded-sm max-h-40 overflow-y-auto bg-background-panel">
                {stationOptions.map(opt => (
                  <button
                    key={opt.stationId}
                    role="option"
                    onClick={() => {
                      setSelectedStation(opt);
                      setStationSearchQuery(opt.name);
                      setStationOptions([]);
                    }}
                    className="w-full text-left px-3 py-2 hover:bg-background-elevated transition-colors cursor-pointer"
                  >
                    <div className="text-sm text-text-primary">{opt.name}</div>
                    <div className="text-xs text-text-muted">{opt.solarSystemName}</div>
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Scan Input */}
          <div>
            <Label htmlFor="structure-fitting-scan">Structure Fitting Scan</Label>
            <textarea
              id="structure-fitting-scan"
              value={scanText}
              onChange={(e) => setScanText(e.target.value)}
              placeholder="Paste structure fitting scan here..."
              rows={4}
              className="flex w-full rounded-sm border border-dim bg-background-void px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary resize-none"
            />
            <Button
              variant="outline"
              size="sm"
              onClick={handleParseScan}
              disabled={!scanText.trim() || scanParsing}
              className="mt-2"
            >
              {scanParsing ? "Parsing..." : "Parse Scan"}
            </Button>
          </div>

          {/* Structure */}
          <div>
            <Label>Structure</Label>
            <Select
              value={structure}
              onValueChange={(newStructure) => {
                setStructure(newStructure);
                const validCategories = getRigCategoriesForStructure(newStructure);
                setRigs((prev) => prev.filter((r) => validCategories.includes(r.category)));
              }}
            >
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="raitaru">Raitaru</SelectItem>
                <SelectItem value="azbel">Azbel</SelectItem>
                <SelectItem value="sotiyo">Sotiyo</SelectItem>
                <SelectItem value="athanor">Athanor</SelectItem>
                <SelectItem value="tatara">Tatara</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Facility Tax */}
          <div>
            <Label htmlFor="facility-tax">Facility Tax %</Label>
            <Input
              id="facility-tax"
              type="number"
              value={facilityTax}
              onChange={(e) => setFacilityTax(parseFloat(e.target.value) || 0)}
              min={0}
              step={0.1}
            />
          </div>

          {/* Services */}
          {services.length > 0 && (
            <div>
              <Label className="text-text-muted">Services</Label>
              <div className="flex gap-1 flex-wrap mt-1">
                {services.map((svc, i) => (
                  <Badge
                    key={i}
                    className={`text-[11px] border ${svc.activity === "manufacturing" ? "bg-primary/10 text-primary border-primary/30" : "bg-category-pink/10 text-category-pink border-category-pink/30"}`}
                  >
                    {svc.serviceName}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          {/* Rigs */}
          <div>
            <div className="flex justify-between items-center mb-2">
              <Label className="text-text-muted">Rigs</Label>
              <Button variant="ghost" size="sm" onClick={handleAddRig}>
                <Plus className="h-4 w-4 mr-1" />
                Add Rig
              </Button>
            </div>

            {rigs.length === 0 && (
              <p className="text-text-muted text-xs">
                No rigs. Paste a scan or add manually.
              </p>
            )}

            {rigs.map((rig, index) => (
              <div key={index} className="flex gap-2 mb-2 items-center">
                <Select value={rig.category} onValueChange={(val) => handleRigChange(index, "category", val)}>
                  <SelectTrigger className="w-[140px]"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {getRigCategoriesForStructure(structure).map((cat) => (
                      <SelectItem key={cat} value={cat}>
                        <span className="flex items-center gap-1.5">
                          <span className={`inline-block w-2 h-2 rounded-full ${getCategoryColor(cat).split(' ')[1]}`} />
                          <span className="capitalize">{cat}</span>
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select value={rig.tier} onValueChange={(val) => handleRigChange(index, "tier", val)}>
                  <SelectTrigger className="w-[80px]"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {getRigTiersForCategory(rig.category).map((t) => (
                      <SelectItem key={t} value={t}>{t.toUpperCase()}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <span className="text-text-muted text-xs flex-1 truncate" title={rig.rigName}>
                  {rig.rigName || "Manual"}
                </span>

                <Button variant="ghost" size="icon" onClick={() => handleRemoveRig(index)} className="text-rose-danger hover:text-rose-danger h-8 w-8">
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              </div>
            ))}
          </div>
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => onClose(false)}>Cancel</Button>
          <Button onClick={handleSave} disabled={!selectedStation || saving}>
            {saving ? "Saving..." : isEdit ? "Update" : "Add"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

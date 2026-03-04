import { useState, useEffect, useCallback, useRef } from "react";
import {
  BlueprintSearchResult,
  BlueprintLevel,
  ManufacturingCalcResult,
  ReactionSystem,
} from "@industry-tool/client/data/models";
import { formatISK, formatNumber, formatDuration } from "@industry-tool/utils/formatting";
import { Loader2, Plus, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

type Props = {
  onJobAdded: () => void;
};

export default function AddJob({ onJobAdded }: Props) {
  const [blueprintQuery, setBlueprintQuery] = useState("");
  const [blueprintOptions, setBlueprintOptions] = useState<BlueprintSearchResult[]>([]);
  const [selectedBlueprint, setSelectedBlueprint] = useState<BlueprintSearchResult | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);

  const [activity, setActivity] = useState("manufacturing");
  const [runs, setRuns] = useState(1);
  const [meLevel, setMeLevel] = useState(10);
  const [teLevel, setTeLevel] = useState(20);
  const [industrySkill, setIndustrySkill] = useState(5);
  const [advIndustrySkill, setAdvIndustrySkill] = useState(5);
  const [structure, setStructure] = useState("raitaru");
  const [rig, setRig] = useState("t2");
  const [security, setSecurity] = useState("high");
  const [facilityTax, setFacilityTax] = useState(1.0);
  const [systemId, setSystemId] = useState<number>(0);
  const [notes, setNotes] = useState("");

  const [detectedLevel, setDetectedLevel] = useState<BlueprintLevel | null>(null);
  const [detectedForBlueprintId, setDetectedForBlueprintId] = useState<number | null>(null);

  const [systems, setSystems] = useState<ReactionSystem[]>([]);
  const [systemQuery, setSystemQuery] = useState("");
  const [systemSearchOpen, setSystemSearchOpen] = useState(false);
  const [calcResult, setCalcResult] = useState<ManufacturingCalcResult | null>(null);
  const [calcLoading, setCalcLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    fetch("/api/industry/systems")
      .then((res) => res.json())
      .then((data) => setSystems(data))
      .catch((err) => console.error("Failed to fetch systems:", err));
  }, []);

  useEffect(() => {
    if (blueprintQuery.length < 2) {
      setBlueprintOptions([]);
      return;
    }

    if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current);
    searchTimeoutRef.current = setTimeout(async () => {
      setSearchLoading(true);
      try {
        const params = new URLSearchParams({ q: blueprintQuery, activity, limit: "20" });
        const res = await fetch(`/api/industry/blueprints?${params.toString()}`);
        const data = await res.json();
        setBlueprintOptions(data || []);
      } catch (err) {
        console.error("Failed to search blueprints:", err);
      } finally {
        setSearchLoading(false);
      }
    }, 300);

    return () => {
      if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current);
    };
  }, [blueprintQuery, activity]);

  const calculate = useCallback(async () => {
    if (!selectedBlueprint || runs <= 0) {
      setCalcResult(null);
      return;
    }

    if (activity !== "manufacturing") {
      setCalcResult(null);
      return;
    }

    setCalcLoading(true);
    try {
      const res = await fetch("/api/industry/calculate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          blueprint_type_id: selectedBlueprint.BlueprintTypeID,
          runs,
          me_level: meLevel,
          te_level: teLevel,
          industry_skill: industrySkill,
          adv_industry_skill: advIndustrySkill,
          system_id: systemId || undefined,
          facility_tax: facilityTax,
          structure,
          rig,
          security,
        }),
      });
      if (res.ok) {
        const data = await res.json();
        setCalcResult(data);
      }
    } catch (err) {
      console.error("Failed to calculate:", err);
    } finally {
      setCalcLoading(false);
    }
  }, [selectedBlueprint, runs, meLevel, teLevel, industrySkill, advIndustrySkill, systemId, facilityTax, structure, rig, security, activity]);

  useEffect(() => {
    calculate();
  }, [calculate]);

  const handleSelectBlueprint = (bp: BlueprintSearchResult) => {
    setSelectedBlueprint(bp);
    setBlueprintQuery(bp.ProductName || bp.BlueprintName);
    setSearchOpen(false);

    fetch("/api/industry/blueprint-levels", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ type_ids: [bp.BlueprintTypeID] }),
    })
      .then((res) => res.json())
      .then((data: Record<string, BlueprintLevel | null>) => {
        const level = data[String(bp.BlueprintTypeID)] ?? null;
        setDetectedLevel(level);
        setDetectedForBlueprintId(bp.BlueprintTypeID);
        if (level) {
          setMeLevel(level.materialEfficiency);
          setTeLevel(level.timeEfficiency);
        } else {
          setMeLevel(10);
          setTeLevel(20);
        }
      })
      .catch((err) => console.error("Failed to fetch blueprint levels:", err));
  };

  const handleSubmit = async () => {
    if (!selectedBlueprint || runs <= 0) return;

    setSubmitting(true);
    try {
      const res = await fetch("/api/industry/queue", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          blueprint_type_id: selectedBlueprint.BlueprintTypeID,
          activity,
          runs,
          me_level: meLevel,
          te_level: teLevel,
          industry_skill: industrySkill,
          adv_industry_skill: advIndustrySkill,
          system_id: systemId || undefined,
          facility_tax: facilityTax,
          structure,
          rig,
          security,
          product_type_id: selectedBlueprint.ProductTypeID,
          notes: notes || undefined,
        }),
      });

      if (res.ok) {
        setSelectedBlueprint(null);
        setBlueprintQuery("");
        setNotes("");
        setCalcResult(null);
        onJobAdded();
      }
    } catch (err) {
      console.error("Failed to add job:", err);
    } finally {
      setSubmitting(false);
    }
  };

  const filteredSystems = systems.filter((s) =>
    systemQuery.length < 1 ? true : s.name.toLowerCase().includes(systemQuery.toLowerCase())
  );

  const selectedSystem = systems.find((s) => s.system_id === systemId);

  return (
    <div>
      {/* Settings Row */}
      <div className="flex gap-2 flex-wrap mb-3">
        <div className="min-w-[300px] flex-grow relative">
          <Label htmlFor="search-blueprint" className="text-xs text-text-secondary mb-1 block">Search Blueprint</Label>
          <Popover open={searchOpen && blueprintOptions.length > 0} onOpenChange={setSearchOpen}>
            <PopoverTrigger asChild>
              <Input
                id="search-blueprint"
                placeholder="Search Blueprint"
                value={blueprintQuery}
                onChange={(e) => {
                  setBlueprintQuery(e.target.value);
                  setSearchOpen(true);
                  if (selectedBlueprint && e.target.value !== (selectedBlueprint.ProductName || selectedBlueprint.BlueprintName)) {
                    setSelectedBlueprint(null);
                    setDetectedLevel(null);
                    setDetectedForBlueprintId(null);
                  }
                }}
                onFocus={() => blueprintOptions.length > 0 && setSearchOpen(true)}
              />
            </PopoverTrigger>
            <PopoverContent className="p-0 w-[var(--radix-popover-trigger-width)]" align="start">
              <div className="max-h-60 overflow-y-auto">
                {searchLoading && (
                  <div className="flex items-center gap-2 px-3 py-2 text-sm text-text-secondary">
                    <Loader2 className="h-4 w-4 animate-spin" /> Searching...
                  </div>
                )}
                {blueprintOptions.map((opt) => (
                  <button
                    key={opt.BlueprintTypeID}
                    className="flex flex-col w-full px-3 py-1.5 text-left text-sm hover:bg-[var(--color-surface-elevated)] cursor-pointer"
                    onClick={() => handleSelectBlueprint(opt)}
                  >
                    <span className="text-text-emphasis">{opt.ProductName}</span>
                    <span className="text-xs text-text-muted">{opt.BlueprintName} - {opt.Activity}</span>
                  </button>
                ))}
              </div>
            </PopoverContent>
          </Popover>
        </div>

        <div className="min-w-[150px]">
          <Label className="text-xs text-text-secondary mb-1 block">Activity</Label>
          <Select value={activity} onValueChange={setActivity}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="manufacturing">Manufacturing</SelectItem>
              <SelectItem value="reaction">Reaction</SelectItem>
              <SelectItem value="invention">Invention</SelectItem>
              <SelectItem value="copying">Copying</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="w-[100px]">
          <Label htmlFor="runs" className="text-xs text-text-secondary mb-1 block">Runs</Label>
          <Input
            id="runs"
            type="number"
            value={runs}
            onChange={(e) => setRuns(Math.max(1, parseInt(e.target.value) || 1))}
            min={1}
          />
        </div>
      </div>

      <div className="flex gap-2 flex-wrap mb-3">
        <div className="w-[90px]">
          <Label htmlFor="me-level" className="text-xs text-text-secondary mb-1 block">ME Level</Label>
          <Input
            id="me-level"
            type="number"
            value={meLevel}
            onChange={(e) => setMeLevel(Math.max(0, Math.min(10, parseInt(e.target.value) || 0)))}
            min={0}
            max={10}
          />
        </div>
        <div className="w-[90px]">
          <Label htmlFor="te-level" className="text-xs text-text-secondary mb-1 block">TE Level</Label>
          <Input
            id="te-level"
            type="number"
            value={teLevel}
            onChange={(e) => setTeLevel(Math.max(0, Math.min(20, parseInt(e.target.value) || 0)))}
            min={0}
            max={20}
          />
        </div>
        {selectedBlueprint && detectedForBlueprintId === selectedBlueprint.BlueprintTypeID ? (
          detectedLevel ? (
            <div className="flex items-center gap-1 self-end pb-1">
              <Badge className="bg-interactive-selected border border-border-active text-primary hover:bg-interactive-active cursor-default text-[11px]">
                Detected: ME {detectedLevel.materialEfficiency} / TE {detectedLevel.timeEfficiency} from {detectedLevel.ownerName}{detectedLevel.isCopy ? " (BPC)" : ""}
              </Badge>
              {(meLevel !== detectedLevel.materialEfficiency || teLevel !== detectedLevel.timeEfficiency) && (
                <Badge className="bg-amber-manufacturing/10 border border-[rgba(245,158,11,0.3)] text-amber-manufacturing hover:bg-[rgba(245,158,11,0.15)] cursor-default text-[11px]">
                  Overridden
                </Badge>
              )}
            </div>
          ) : (
            <div className="flex items-center gap-1 self-end pb-1">
              <Badge className="bg-amber-manufacturing/10 border border-[rgba(245,158,11,0.3)] text-amber-manufacturing hover:bg-[rgba(245,158,11,0.15)] cursor-default text-[11px]">
                <AlertTriangle className="h-3 w-3 mr-1" />
                No blueprint detected — using manual values
              </Badge>
            </div>
          )
        ) : null}
        <div className="w-[120px]">
          <Label className="text-xs text-text-secondary mb-1 block">Industry Skill</Label>
          <Input
            type="number"
            value={industrySkill}
            onChange={(e) => setIndustrySkill(Math.max(0, Math.min(5, parseInt(e.target.value) || 0)))}
            min={0}
            max={5}
          />
        </div>
        <div className="w-[120px]">
          <Label className="text-xs text-text-secondary mb-1 block">Adv Industry</Label>
          <Input
            type="number"
            value={advIndustrySkill}
            onChange={(e) => setAdvIndustrySkill(Math.max(0, Math.min(5, parseInt(e.target.value) || 0)))}
            min={0}
            max={5}
          />
        </div>

        <div className="min-w-[120px]">
          <Label className="text-xs text-text-secondary mb-1 block">Structure</Label>
          <Select value={structure} onValueChange={setStructure}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="station">NPC Station</SelectItem>
              <SelectItem value="raitaru">Raitaru</SelectItem>
              <SelectItem value="azbel">Azbel</SelectItem>
              <SelectItem value="sotiyo">Sotiyo</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="min-w-[90px]">
          <Label className="text-xs text-text-secondary mb-1 block">Rig</Label>
          <Select value={rig} onValueChange={setRig}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="none">None</SelectItem>
              <SelectItem value="t1">T1</SelectItem>
              <SelectItem value="t2">T2</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="min-w-[100px]">
          <Label className="text-xs text-text-secondary mb-1 block">Security</Label>
          <Select value={security} onValueChange={setSecurity}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="high">Highsec</SelectItem>
              <SelectItem value="low">Lowsec</SelectItem>
              <SelectItem value="null">Nullsec</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="w-[120px]">
          <Label className="text-xs text-text-secondary mb-1 block">Facility Tax %</Label>
          <Input
            type="number"
            value={facilityTax}
            onChange={(e) => setFacilityTax(parseFloat(e.target.value) || 0)}
          />
        </div>
      </div>

      <div className="flex gap-2 flex-wrap mb-3 items-end">
        <div className="min-w-[250px]">
          <Label className="text-xs text-text-secondary mb-1 block">System (optional)</Label>
          <Popover open={systemSearchOpen} onOpenChange={setSystemSearchOpen}>
            <PopoverTrigger asChild>
              <button
                className="flex h-9 w-full items-center justify-between rounded-sm border border-[var(--color-border-dim)] bg-[var(--color-bg-void)] px-3 py-2 text-sm text-left"
                onClick={() => setSystemSearchOpen(true)}
              >
                <span className={selectedSystem ? "text-[var(--color-text-primary)]" : "text-[var(--color-text-muted)]"}>
                  {selectedSystem ? `${selectedSystem.name} (${(selectedSystem.cost_index * 100).toFixed(2)}%)` : "Select system..."}
                </span>
              </button>
            </PopoverTrigger>
            <PopoverContent className="p-0 w-[300px]" align="start">
              <div className="p-2">
                <Input
                  placeholder="Search systems..."
                  value={systemQuery}
                  onChange={(e) => setSystemQuery(e.target.value)}
                  className="h-8"
                  autoFocus
                />
              </div>
              <div className="max-h-60 overflow-y-auto">
                {filteredSystems.slice(0, 50).map((sys) => (
                  <button
                    key={sys.system_id}
                    className="flex w-full px-3 py-1.5 text-sm hover:bg-[var(--color-surface-elevated)] cursor-pointer text-left"
                    onClick={() => {
                      setSystemId(sys.system_id);
                      setSystemSearchOpen(false);
                      setSystemQuery("");
                    }}
                  >
                    {sys.name} ({(sys.cost_index * 100).toFixed(2)}%)
                  </button>
                ))}
              </div>
            </PopoverContent>
          </Popover>
        </div>

        <div className="flex-grow min-w-[200px]">
          <Label className="text-xs text-text-secondary mb-1 block">Notes</Label>
          <Input
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
          />
        </div>

        <Button
          onClick={handleSubmit}
          disabled={!selectedBlueprint || runs <= 0 || submitting}
          className="h-9"
        >
          {submitting ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : <Plus className="h-4 w-4 mr-1" />}
          Add to Queue
        </Button>
      </div>

      {/* Calculation Result */}
      {calcLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="h-6 w-6 animate-spin text-primary" />
        </div>
      )}

      {calcResult && !calcLoading && (
        <div className="bg-background-panel rounded-sm border border-overlay-subtle p-4">
          <h3 className="text-sm font-semibold text-primary mb-2">
            Cost Estimate: {calcResult.productName} x{formatNumber(calcResult.totalProducts)}
          </h3>
          <div className="flex gap-6 flex-wrap mb-3">
            <div>
              <span className="text-xs text-text-muted block">Input Cost</span>
              <span className="text-sm text-text-emphasis">{formatISK(calcResult.inputCost)}</span>
            </div>
            <div>
              <span className="text-xs text-text-muted block">Job Cost</span>
              <span className="text-sm text-text-emphasis">{formatISK(calcResult.jobCost)}</span>
            </div>
            <div>
              <span className="text-xs text-text-muted block">Total Cost</span>
              <span className="text-sm text-text-emphasis font-semibold">{formatISK(calcResult.totalCost)}</span>
            </div>
            <div>
              <span className="text-xs text-text-muted block">Output Value</span>
              <span className="text-sm text-text-emphasis">{formatISK(calcResult.outputValue)}</span>
            </div>
            <div>
              <span className="text-xs text-text-muted block">Profit</span>
              <span className={`text-sm ${calcResult.profit >= 0 ? "text-teal-success" : "text-rose-danger"}`}>
                {formatISK(calcResult.profit)}
              </span>
            </div>
            <div>
              <span className="text-xs text-text-muted block">Margin</span>
              <span className={`text-sm ${calcResult.margin >= 0 ? "text-teal-success" : "text-rose-danger"}`}>
                {calcResult.margin.toFixed(1)}%
              </span>
            </div>
            <div>
              <span className="text-xs text-text-muted block">Per Run</span>
              <span className="text-sm text-text-emphasis">{formatDuration(calcResult.secsPerRun)}</span>
            </div>
            <div>
              <span className="text-xs text-text-muted block">Total Time</span>
              <span className="text-sm text-text-emphasis">{formatDuration(calcResult.totalDuration)}</span>
            </div>
          </div>

          {calcResult.materials.length > 0 && (
            <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
              <Table>
                <TableHeader>
                  <TableRow className="bg-background-void">
                    <TableHead>Material</TableHead>
                    <TableHead className="text-right">Base Qty</TableHead>
                    <TableHead className="text-right">Required</TableHead>
                    <TableHead className="text-right">Price</TableHead>
                    <TableHead className="text-right">Cost</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {calcResult.materials.map((mat) => (
                    <TableRow key={mat.typeId}>
                      <TableCell className="text-text-emphasis">{mat.name}</TableCell>
                      <TableCell className="text-right text-text-secondary">{formatNumber(mat.baseQty)}</TableCell>
                      <TableCell className="text-right text-text-emphasis">{formatNumber(mat.batchQty)}</TableCell>
                      <TableCell className="text-right text-text-secondary">{formatISK(mat.price)}</TableCell>
                      <TableCell className="text-right text-text-primary">{formatISK(mat.cost)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

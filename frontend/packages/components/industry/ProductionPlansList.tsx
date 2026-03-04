import { useState, useEffect, useRef, useCallback } from "react";
import { useSession } from "next-auth/react";
import {
  ProductionPlan,
  BlueprintSearchResult,
  UserStation,
} from "@industry-tool/client/data/models";
import Navbar from "@industry-tool/components/Navbar";
import Loading from "@industry-tool/components/loading";
import ProductionPlanEditor from "./ProductionPlanEditor";
import { ArrowLeft, Plus, Pencil, Trash2, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import {
  Popover,
  PopoverTrigger,
  PopoverContent,
} from "@/components/ui/popover";

export default function ProductionPlansList() {
  const { data: session } = useSession();
  const [plans, setPlans] = useState<ProductionPlan[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedPlanId, setSelectedPlanId] = useState<number | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const hasFetchedRef = useRef(false);

  const fetchPlans = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/industry/plans");
      if (res.ok) {
        const data = await res.json();
        setPlans(data || []);
      }
    } catch (err) {
      console.error("Failed to fetch plans:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchPlans();
    }
  }, [session, fetchPlans]);

  const handleDelete = async (id: number) => {
    try {
      const res = await fetch(`/api/industry/plans/${id}`, {
        method: "DELETE",
      });
      if (res.ok) {
        fetchPlans();
      }
    } catch (err) {
      console.error("Failed to delete plan:", err);
    }
  };

  const handlePlanCreated = (plan: ProductionPlan) => {
    setCreateDialogOpen(false);
    fetchPlans();
    setSelectedPlanId(plan.id);
  };

  const handleBack = () => {
    setSelectedPlanId(null);
    fetchPlans();
  };

  // Show plan editor if a plan is selected
  if (selectedPlanId !== null) {
    return (
      <>
        <Navbar />
        <div className="px-4 mt-2 mb-4">
          <Button
            variant="ghost"
            onClick={handleBack}
            className="text-[#94a3b8] mb-2"
          >
            <ArrowLeft className="h-4 w-4 mr-1" />
            Back to Plans
          </Button>
          <ProductionPlanEditor planId={selectedPlanId} />
        </div>
      </>
    );
  }

  return (
    <>
      <Navbar />
      <div className="px-4 mt-2 mb-4">
        <div className="flex justify-between items-center mb-2">
          <h2 className="text-xl font-semibold text-[#e2e8f0]">
            Production Plans
          </h2>
          <Button onClick={() => setCreateDialogOpen(true)}>
            <Plus className="h-4 w-4 mr-1" />
            New Plan
          </Button>
        </div>

        {loading ? (
          <Loading />
        ) : plans.length === 0 ? (
          <div className="bg-[#12151f] rounded-sm border border-[rgba(148,163,184,0.1)] p-8 text-center">
            <p className="text-[#64748b]">
              No production plans yet. Create one to define how items should be
              produced.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
            <Table>
              <TableHeader>
                <TableRow className="bg-[#0f1219]">
                  <TableHead>Product</TableHead>
                  <TableHead>Plan Name</TableHead>
                  <TableHead className="text-right">Steps</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="text-center">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {plans.map((plan, idx) => (
                  <TableRow
                    key={plan.id}
                    className={`${idx % 2 === 0 ? "bg-[#12151f]" : "bg-[#0f1219]"} hover:bg-[#1a1d2e] cursor-pointer`}
                    onClick={() => setSelectedPlanId(plan.id)}
                  >
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <img
                          src={`https://images.evetech.net/types/${plan.productTypeId}/icon?size=32`}
                          alt=""
                          width={24}
                          height={24}
                          className="rounded-sm"
                        />
                        <span className="text-[#e2e8f0] text-sm">
                          {plan.productName || `Type ${plan.productTypeId}`}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="text-[#cbd5e1] text-sm">
                      {plan.name}
                    </TableCell>
                    <TableCell className="text-right text-[#cbd5e1] text-sm">
                      {plan.steps?.length || 0}
                    </TableCell>
                    <TableCell className="text-[#64748b] text-[13px]">
                      {new Date(plan.createdAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-center">
                      <button
                        className="p-1 rounded hover:bg-[rgba(0,212,255,0.1)] text-[#00d4ff]"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedPlanId(plan.id);
                        }}
                        aria-label="Edit"
                      >
                        <Pencil className="h-4 w-4" />
                      </button>
                      <button
                        className="p-1 rounded hover:bg-[rgba(239,68,68,0.1)] text-[#ef4444]"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDelete(plan.id);
                        }}
                        aria-label="Delete"
                      >
                        <Trash2 className="h-4 w-4" data-testid="DeleteIcon" />
                      </button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}

        <CreatePlanDialog
          open={createDialogOpen}
          onClose={() => setCreateDialogOpen(false)}
          onCreated={handlePlanCreated}
        />
      </div>
    </>
  );
}

// --- Create Plan Dialog ---

function CreatePlanDialog({
  open,
  onClose,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  onCreated: (plan: ProductionPlan) => void;
}) {
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<BlueprintSearchResult[]>(
    [],
  );
  const [searchLoading, setSearchLoading] = useState(false);
  const [selectedBlueprint, setSelectedBlueprint] =
    useState<BlueprintSearchResult | null>(null);
  const [creating, setCreating] = useState(false);
  const searchTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [userStations, setUserStations] = useState<UserStation[]>([]);
  const [defaultManufacturingStation, setDefaultManufacturingStation] =
    useState<UserStation | null>(null);
  const [defaultReactionStation, setDefaultReactionStation] =
    useState<UserStation | null>(null);
  const [bpSearchOpen, setBpSearchOpen] = useState(false);

  useEffect(() => {
    if (!open) {
      setDefaultManufacturingStation(null);
      setDefaultReactionStation(null);
      return;
    }
    const fetchStations = async () => {
      try {
        const res = await fetch("/api/stations/user-stations");
        if (res.ok) {
          const data: UserStation[] = await res.json();
          setUserStations(data || []);
        }
      } catch (err) {
        console.error("Failed to fetch user stations:", err);
      }
    };
    fetchStations();
  }, [open]);

  const manufacturingStations = userStations.filter((s) =>
    s.activities.includes("manufacturing"),
  );
  const reactionStations = userStations.filter((s) =>
    s.activities.includes("reaction"),
  );

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current);
    if (query.length < 2) {
      setSearchResults([]);
      return;
    }
    searchTimeoutRef.current = setTimeout(async () => {
      setSearchLoading(true);
      try {
        const res = await fetch(
          `/api/industry/blueprints?q=${encodeURIComponent(query)}&limit=20`,
        );
        if (res.ok) {
          const data = await res.json();
          setSearchResults(data || []);
        }
      } catch (err) {
        console.error("Blueprint search failed:", err);
      } finally {
        setSearchLoading(false);
      }
    }, 300);
  };

  const handleCreate = async () => {
    if (!selectedBlueprint) return;
    setCreating(true);
    try {
      const res = await fetch("/api/industry/plans", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          product_type_id: selectedBlueprint.ProductTypeID,
          default_manufacturing_station_id:
            defaultManufacturingStation?.id || null,
          default_reaction_station_id: defaultReactionStation?.id || null,
        }),
      });
      if (res.ok) {
        const plan = await res.json();
        setSelectedBlueprint(null);
        setSearchQuery("");
        onCreated(plan);
      }
    } catch (err) {
      console.error("Failed to create plan:", err);
    } finally {
      setCreating(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md bg-[#12151f] border-[rgba(148,163,184,0.15)]">
        <DialogHeader>
          <DialogTitle className="text-[#e2e8f0]">Create Production Plan</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-3 pt-1">
          {/* Blueprint Search */}
          <div>
            <Label className="text-sm text-[#94a3b8] mb-1 block">Product</Label>
            <Popover open={bpSearchOpen} onOpenChange={setBpSearchOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  className="w-full justify-start font-normal"
                >
                  {selectedBlueprint ? (
                    <div className="flex items-center gap-1">
                      <img
                        src={`https://images.evetech.net/types/${selectedBlueprint.ProductTypeID}/icon?size=32`}
                        alt=""
                        width={20}
                        height={20}
                      />
                      <span>{selectedBlueprint.ProductName} ({selectedBlueprint.Activity})</span>
                    </div>
                  ) : (
                    <span className="text-muted-foreground">Search for a product...</span>
                  )}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[400px] p-2" align="start">
                <Input
                  placeholder="e.g. Rifter, Damage Control II..."
                  value={searchQuery}
                  onChange={(e) => handleSearch(e.target.value)}
                  autoFocus
                />
                <div className="max-h-[250px] overflow-y-auto mt-1">
                  {searchLoading && (
                    <div className="flex items-center gap-2 p-2 text-sm text-[#94a3b8]">
                      <Loader2 className="h-4 w-4 animate-spin" />
                      Searching...
                    </div>
                  )}
                  {!searchLoading && searchQuery.length < 2 && (
                    <p className="text-sm text-[#64748b] p-2">Type to search...</p>
                  )}
                  {!searchLoading && searchQuery.length >= 2 && searchResults.length === 0 && (
                    <p className="text-sm text-[#64748b] p-2">No blueprints found</p>
                  )}
                  {searchResults.map((option) => (
                    <button
                      key={`${option.BlueprintTypeID}-${option.Activity}`}
                      className="w-full flex items-center gap-2 p-2 rounded hover:bg-[rgba(0,212,255,0.08)] text-left"
                      onClick={() => {
                        setSelectedBlueprint(option);
                        setBpSearchOpen(false);
                      }}
                    >
                      <img
                        src={`https://images.evetech.net/types/${option.ProductTypeID}/icon?size=32`}
                        alt=""
                        width={24}
                        height={24}
                      />
                      <div>
                        <div className="text-sm text-[#e2e8f0]">
                          {option.ProductName}
                        </div>
                        <div className="text-xs text-[#64748b]">
                          {option.BlueprintName} &middot; {option.Activity}
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              </PopoverContent>
            </Popover>
          </div>

          {/* Manufacturing Station */}
          <div>
            <Label htmlFor="default-mfg-station" className="text-sm text-[#94a3b8] mb-1 block">Default Manufacturing Station</Label>
            <Select
              value={defaultManufacturingStation ? String(defaultManufacturingStation.id) : "none"}
              onValueChange={(val) => {
                if (val === "none") {
                  setDefaultManufacturingStation(null);
                } else {
                  const station = manufacturingStations.find((s) => String(s.id) === val);
                  setDefaultManufacturingStation(station || null);
                }
              }}
            >
              <SelectTrigger id="default-mfg-station">
                <SelectValue placeholder="Optional" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">None</SelectItem>
                {manufacturingStations.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>
                    {s.stationName || "Unknown"} ({s.solarSystemName})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Reaction Station */}
          <div>
            <Label htmlFor="default-rxn-station" className="text-sm text-[#94a3b8] mb-1 block">Default Reaction Station</Label>
            <Select
              value={defaultReactionStation ? String(defaultReactionStation.id) : "none"}
              onValueChange={(val) => {
                if (val === "none") {
                  setDefaultReactionStation(null);
                } else {
                  const station = reactionStations.find((s) => String(s.id) === val);
                  setDefaultReactionStation(station || null);
                }
              }}
            >
              <SelectTrigger id="default-rxn-station">
                <SelectValue placeholder="Optional" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">None</SelectItem>
                {reactionStations.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>
                    {s.stationName || "Unknown"} ({s.solarSystemName})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={handleCreate}
            disabled={!selectedBlueprint || creating}
          >
            {creating ? "Creating..." : "Create Plan"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

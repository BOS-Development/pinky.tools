import { useState, useEffect, useCallback, useRef } from "react";
import {
  ProductionPlan,
  ProductionPlanStep,
  PlanMaterial,
  GenerateJobsResult,
  PlanPreviewResult,
  UserStation,
  HangarsResponse,
  BlueprintLevel,
} from "@industry-tool/client/data/models";
import { formatISK } from "@industry-tool/utils/formatting";
import {
  ChevronDown,
  ChevronRight,
  Wrench,
  Trash2,
  Pencil,
  ShoppingCart,
  Play,
  Check,
  X,
  ArrowLeftRight,
  Info,
  AlertTriangle,
  CheckCircle,
  Truck,
  Loader2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
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
  Tooltip,
  TooltipTrigger,
  TooltipContent,
  TooltipProvider,
} from "@/components/ui/tooltip";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { toast } from "@/components/ui/sonner";
import BatchConfigureTab from "./BatchConfigureTab";
import { formatNumber, formatCompact } from "@industry-tool/utils/formatting";

type Props = {
  planId: number;
};

type StepMaterials = {
  [stepId: number]: PlanMaterial[];
};

export default function ProductionPlanEditor({ planId }: Props) {
  const [plan, setPlan] = useState<ProductionPlan | null>(null);
  const [loading, setLoading] = useState(true);
  const [expandedSteps, setExpandedSteps] = useState<Set<number>>(new Set());
  const [stepMaterials, setStepMaterials] = useState<StepMaterials>({});
  const [loadingMaterials, setLoadingMaterials] = useState<Set<number>>(
    new Set(),
  );
  const [editStepId, setEditStepId] = useState<number | null>(null);
  const [generateDialogOpen, setGenerateDialogOpen] = useState(false);
  const [generateQuantity, setGenerateQuantity] = useState(1);
  const [generating, setGenerating] = useState(false);
  const [generateResult, setGenerateResult] =
    useState<GenerateJobsResult | null>(null);
  const [previewResult, setPreviewResult] = useState<PlanPreviewResult | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [selectedParallelism, setSelectedParallelism] = useState<number>(0);
  const [editingName, setEditingName] = useState(false);
  const [nameValue, setNameValue] = useState("");
  const [tab, setTab] = useState("step-tree");
  const [transportProfiles, setTransportProfiles] = useState<{ id: number; name: string; transportMethod: string }[]>([]);
  const [detectedLevels, setDetectedLevels] = useState<Record<number, BlueprintLevel>>({});

  const initialLoadRef = useRef(true);

  const fetchBlueprintLevels = useCallback(async (steps: ProductionPlanStep[]) => {
    const typeIds = [...new Set(steps.map((s) => s.blueprintTypeId))];
    if (typeIds.length === 0) return;
    try {
      const res = await fetch("/api/industry/blueprint-levels", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ type_ids: typeIds }),
      });
      if (res.ok) {
        const data: Record<string, BlueprintLevel | null> = await res.json();
        const levels: Record<number, BlueprintLevel> = {};
        for (const [key, val] of Object.entries(data)) {
          if (val !== null) {
            levels[parseInt(key)] = val;
          }
        }
        setDetectedLevels((prev) => ({ ...prev, ...levels }));
      }
    } catch (err) {
      console.error("Failed to fetch blueprint levels:", err);
    }
  }, []);

  const fetchPlan = useCallback(async () => {
    if (initialLoadRef.current) {
      setLoading(true);
    }
    try {
      const res = await fetch(`/api/industry/plans/${planId}`);
      if (res.ok) {
        const data = await res.json();
        setPlan(data);
        if (data.steps?.length > 0) {
          fetchBlueprintLevels(data.steps);
        }
        if (initialLoadRef.current && data.steps?.length > 0) {
          initialLoadRef.current = false;
          const rootStep = data.steps.find(
            (s: ProductionPlanStep) => !s.parentStepId,
          );
          if (rootStep) {
            setExpandedSteps(new Set([rootStep.id]));
            fetchMaterials(rootStep.id);
          }
        }
      }
    } catch (err) {
      console.error("Failed to fetch plan:", err);
    } finally {
      setLoading(false);
    }
  }, [planId, fetchBlueprintLevels]);

  useEffect(() => {
    fetchPlan();
  }, [fetchPlan]);

  useEffect(() => {
    const fetchProfiles = async () => {
      try {
        const res = await fetch("/api/transport/profiles");
        if (res.ok) {
          const data = await res.json();
          setTransportProfiles(data || []);
        }
      } catch (err) {
        console.error("Failed to fetch transport profiles:", err);
      }
    };
    fetchProfiles();
  }, []);

  const fetchMaterials = async (stepId: number) => {
    setLoadingMaterials((prev) => new Set([...prev, stepId]));
    try {
      const res = await fetch(
        `/api/industry/plans/${planId}/steps/${stepId}/materials`,
      );
      if (res.ok) {
        const data: PlanMaterial[] = await res.json() || [];
        setStepMaterials((prev) => ({ ...prev, [stepId]: data }));
        const materialBlueprintTypeIds = data
          .filter((m) => m.hasBlueprint && m.blueprintTypeId != null)
          .map((m) => m.blueprintTypeId as number);
        if (materialBlueprintTypeIds.length > 0) {
          fetchBlueprintLevels(
            materialBlueprintTypeIds.map((id) => ({ blueprintTypeId: id } as ProductionPlanStep)),
          );
        }
      }
    } catch (err) {
      console.error("Failed to fetch materials:", err);
    } finally {
      setLoadingMaterials((prev) => {
        const next = new Set(prev);
        next.delete(stepId);
        return next;
      });
    }
  };

  const toggleExpand = (stepId: number) => {
    setExpandedSteps((prev) => {
      const next = new Set(prev);
      if (next.has(stepId)) {
        next.delete(stepId);
      } else {
        next.add(stepId);
        if (!stepMaterials[stepId]) {
          fetchMaterials(stepId);
        }
      }
      return next;
    });
  };

  const handleToggleProduce = async (
    parentStepId: number,
    material: PlanMaterial,
  ) => {
    if (material.isProduced) {
      const childStep = plan?.steps?.find(
        (s) =>
          s.parentStepId === parentStepId &&
          s.productTypeId === material.typeId,
      );
      if (childStep) {
        try {
          await fetch(
            `/api/industry/plans/${planId}/steps/${childStep.id}`,
            { method: "DELETE" },
          );
          await fetchPlan();
          fetchMaterials(parentStepId);
        } catch (err) {
          console.error("Failed to remove step:", err);
        }
      }
    } else {
      try {
        const body: Record<string, unknown> = {
          parent_step_id: parentStepId,
          product_type_id: material.typeId,
        };
        if (material.blueprintTypeId) {
          const detected = detectedLevels[material.blueprintTypeId];
          if (detected) {
            body.me_level = detected.materialEfficiency;
            body.te_level = detected.timeEfficiency;
          }
        }
        const res = await fetch(`/api/industry/plans/${planId}/steps`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        });
        if (res.ok) {
          const newStep = await res.json();
          await fetchPlan();
          fetchMaterials(parentStepId);
          if (newStep?.id) {
            setExpandedSteps((prev) => new Set([...prev, newStep.id]));
            fetchMaterials(newStep.id);
            if (material.blueprintTypeId && !detectedLevels[material.blueprintTypeId]) {
              fetchBlueprintLevels([{ blueprintTypeId: material.blueprintTypeId } as ProductionPlanStep]);
            }
          }
        }
      } catch (err) {
        console.error("Failed to create step:", err);
      }
    }
  };

  const handleUpdateStep = async (
    stepId: number,
    updates: Partial<ProductionPlanStep>,
  ) => {
    try {
      const res = await fetch(
        `/api/industry/plans/${planId}/steps/${stepId}`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(updates),
        },
      );
      if (res.ok) {
        setEditStepId(null);
        fetchPlan();
        toast.success("Step updated");
      }
    } catch (err) {
      console.error("Failed to update step:", err);
    }
  };

  const handlePreview = async (quantity: number) => {
    setPreviewLoading(true);
    setPreviewError(null);
    setPreviewResult(null);
    try {
      const res = await fetch(`/api/industry/plans/${planId}/preview`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ quantity }),
      });
      if (res.ok) {
        const result: PlanPreviewResult = await res.json();
        setPreviewResult(result);
        setSelectedParallelism(0);
      } else {
        const err = await res.json();
        setPreviewError(err.error || "Preview failed");
      }
    } catch (err) {
      console.error("Failed to preview jobs:", err);
      setPreviewError("Failed to load preview");
    } finally {
      setPreviewLoading(false);
    }
  };

  const handleGenerate = async () => {
    setGenerating(true);
    try {
      const body: { quantity: number; parallelism?: number } = { quantity: generateQuantity };
      if (selectedParallelism > 0) {
        body.parallelism = selectedParallelism;
      }
      const res = await fetch(`/api/industry/plans/${planId}/generate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (res.ok) {
        const result: GenerateJobsResult = await res.json();
        setGenerateResult(result);
        toast.success(`Created ${result.created.length} job(s), skipped ${result.skipped.length}`);
      }
    } catch (err) {
      console.error("Failed to generate jobs:", err);
      toast.error("Failed to generate jobs");
    } finally {
      setGenerating(false);
    }
  };

  const handleSaveName = async () => {
    const trimmed = nameValue.trim();
    if (!trimmed || trimmed === plan?.name) {
      setEditingName(false);
      return;
    }
    try {
      const res = await fetch(`/api/industry/plans/${planId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: trimmed }),
      });
      if (res.ok) {
        setEditingName(false);
        fetchPlan();
      }
    } catch (err) {
      console.error("Failed to update plan name:", err);
    }
  };

  const handleSaveTransportSettings = async (settings: {
    transport_fulfillment?: string;
    transport_method?: string;
    transport_profile_id?: number;
    courier_rate_per_m3: number;
    courier_collateral_rate: number;
  }) => {
    try {
      const res = await fetch(`/api/industry/plans/${planId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: plan?.name,
          ...settings,
        }),
      });
      if (res.ok) {
        fetchPlan();
        toast.success("Transport settings saved");
      }
    } catch (err) {
      console.error("Failed to save transport settings:", err);
      toast.error("Failed to save transport settings");
    }
  };

  if (loading || !plan) {
    return (
      <div className="text-center py-4">
        <p className="text-text-muted">Loading plan...</p>
      </div>
    );
  }

  // Build step tree structure
  const stepMap = new Map<number, ProductionPlanStep>();
  const childrenMap = new Map<number, ProductionPlanStep[]>();
  let rootStep: ProductionPlanStep | null = null;

  for (const step of plan.steps || []) {
    stepMap.set(step.id, step);
    if (!step.parentStepId) {
      rootStep = step;
    } else {
      const children = childrenMap.get(step.parentStepId) || [];
      children.push(step);
      childrenMap.set(step.parentStepId, children);
    }
  }

  const depthColors = ["var(--color-primary-cyan)", "var(--color-success-teal)", "var(--color-manufacturing-amber)", "#a78bfa", "#ec4899", "#06b6d4"];

  const renderDepthIndicators = (colorPath: number[]) =>
    colorPath.map((colorIndex, i) => (
      <div
        key={`depth-bar-${i}`}
        className="absolute top-0 bottom-0 w-[3px]"
        style={{
          left: `${i * 32 + 8}px`,
          backgroundColor: depthColors[colorIndex % depthColors.length],
        }}
      />
    ));

  const renderStepRow = (step: ProductionPlanStep, depth: number, colorPath: number[], parentStep?: ProductionPlanStep) => {
    const isExpanded = expandedSteps.has(step.id);
    const materials = stepMaterials[step.id] || [];
    const isLoadingMats = loadingMaterials.has(step.id);
    const children = childrenMap.get(step.id) || [];

    const rows: React.ReactNode[] = [];

    // Step header row
    rows.push(
      <TableRow
        key={`step-${step.id}`}
        className={`${depth === 0 ? "bg-[#1a1d2e]" : "bg-background-panel"} hover:bg-[#1e2235]`}
      >
        <TableCell className="relative pl-0">
          {renderDepthIndicators(colorPath)}
          <div className="flex items-center gap-1" style={{ paddingLeft: `${depth * 32 + 20}px` }}>
            <button
              className="p-0.5 rounded text-text-secondary hover:bg-overlay-subtle"
              onClick={() => toggleExpand(step.id)}
            >
              {isExpanded ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronRight className="h-4 w-4" />
              )}
            </button>
            <img
              src={`https://images.evetech.net/types/${step.productTypeId}/icon?size=32`}
              alt=""
              width={20}
              height={20}
              className="rounded-sm"
            />
            <span className={`text-text-emphasis text-sm ${depth === 0 ? "font-semibold" : ""}`}>
              {step.productName || `Type ${step.productTypeId}`}
            </span>
            <Badge
              className={`ml-1 h-5 text-[11px] cursor-default ${
                step.activity === "manufacturing"
                  ? "bg-[#1e3a5f] text-blue-science hover:bg-[#1e3a5f]"
                  : "bg-[#3a1e5f] text-[#a78bfa] hover:bg-[#3a1e5f]"
              }`}
            >
              {step.activity}
            </Badge>
          </div>
        </TableCell>
        <TableCell className="text-text-secondary text-[13px]">
          <TooltipProvider>
            <div className="flex items-center gap-1">
              ME {step.meLevel} / TE {step.teLevel}
              {detectedLevels[step.blueprintTypeId] ? (
                (step.activity !== "reaction" &&
                 (detectedLevels[step.blueprintTypeId].materialEfficiency !== step.meLevel ||
                  detectedLevels[step.blueprintTypeId].timeEfficiency !== step.teLevel)) ? (
                  <Tooltip>
                    <TooltipTrigger>
                      <Info className="h-3.5 w-3.5 text-primary" data-testid="InfoIcon" />
                    </TooltipTrigger>
                    <TooltipContent>
                      Detected: ME {detectedLevels[step.blueprintTypeId].materialEfficiency} / TE {detectedLevels[step.blueprintTypeId].timeEfficiency} from {detectedLevels[step.blueprintTypeId].ownerName}
                    </TooltipContent>
                  </Tooltip>
                ) : (
                  <Tooltip>
                    <TooltipTrigger>
                      <CheckCircle className="h-3.5 w-3.5 text-teal-success" data-testid="CheckCircleOutlineIcon" />
                    </TooltipTrigger>
                    <TooltipContent>
                      Blueprint detected from {detectedLevels[step.blueprintTypeId].ownerName}
                    </TooltipContent>
                  </Tooltip>
                )
              ) : Object.keys(detectedLevels).length > 0 && (
                <Tooltip>
                  <TooltipTrigger>
                    <AlertTriangle className="h-3.5 w-3.5 text-amber-manufacturing" data-testid="WarningAmberIcon" />
                  </TooltipTrigger>
                  <TooltipContent>
                    No blueprint detected — ME/TE values are manual
                  </TooltipContent>
                </Tooltip>
              )}
            </div>
          </TooltipProvider>
        </TableCell>
        <TableCell className="text-text-secondary text-[13px]">
          {step.structure} / {step.rig} / {step.security}
        </TableCell>
        <TableCell className="text-text-secondary text-[13px]">
          {step.stationName || "\u2014"}
          {step.sourceOwnerName && (
            <span className="block text-text-muted text-[11px]">
              In: {step.sourceOwnerName}
              {step.sourceDivisionName ? ` / ${step.sourceDivisionName}` : ""}
              {step.sourceContainerName ? ` / ${step.sourceContainerName}` : ""}
            </span>
          )}
          {step.outputOwnerName ? (
            <span className="block text-text-muted text-[11px]">
              Out: {step.outputOwnerName}
              {step.outputDivisionName ? ` / ${step.outputDivisionName}` : ""}
              {step.outputContainerName ? ` / ${step.outputContainerName}` : ""}
            </span>
          ) : !step.parentStepId ? (
            <span className="block text-text-muted text-[11px] italic">
              Out: set at build time
            </span>
          ) : null}
          {step.parentStepId && parentStep &&
           step.userStationId && parentStep.userStationId &&
           step.userStationId !== parentStep.userStationId && (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <div className="flex items-center gap-1 mt-0.5">
                    <ArrowLeftRight className="h-3.5 w-3.5 text-amber-manufacturing" />
                    <span className="text-amber-manufacturing text-[11px]">Transfer</span>
                  </div>
                </TooltipTrigger>
                <TooltipContent>Items must be moved between stations</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )}
        </TableCell>
        <TableCell className="text-center">
          <button
            className="p-1 rounded hover:bg-interactive-selected text-primary"
            onClick={() => setEditStepId(step.id)}
            aria-label="Edit"
          >
            <Pencil className="h-4 w-4" data-testid="EditIcon" />
          </button>
          {depth > 0 && (
            <button
              className="p-1 rounded hover:bg-rose-danger/10 text-rose-danger"
              onClick={async () => {
                await fetch(
                  `/api/industry/plans/${planId}/steps/${step.id}`,
                  { method: "DELETE" },
                );
                fetchPlan();
                if (step.parentStepId) fetchMaterials(step.parentStepId);
              }}
              aria-label="Delete"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          )}
        </TableCell>
      </TableRow>,
    );

    // Material rows (when expanded)
    if (isExpanded) {
      if (isLoadingMats) {
        rows.push(
          <TableRow key={`loading-${step.id}`}>
            <TableCell
              colSpan={5}
              className="relative pl-0 text-text-muted text-[13px]"
            >
              {renderDepthIndicators(colorPath)}
              <div style={{ paddingLeft: `${(depth + 1) * 32 + 20}px` }}>Loading materials...</div>
            </TableCell>
          </TableRow>,
        );
      } else {
        const sortedMaterials = [...materials].sort((a, b) => {
          if (a.isProduced !== b.isProduced) return a.isProduced ? -1 : 1;
          if (a.hasBlueprint !== b.hasBlueprint) return a.hasBlueprint ? -1 : 1;
          return 0;
        });
        let childIndex = 0;
        for (const mat of sortedMaterials) {
          const childStep = children.find(
            (c) => c.productTypeId === mat.typeId,
          );

          if (childStep) {
            rows.push(...renderStepRow(childStep, depth + 1, [...colorPath, childIndex], step));
            childIndex++;
          } else {
            rows.push(
              <TableRow
                key={`mat-${step.id}-${mat.typeId}`}
                className="bg-background-void hover:bg-[#151825]"
              >
                <TableCell className="relative pl-0">
                  {renderDepthIndicators(colorPath)}
                  <div
                    className="flex items-center gap-1"
                    style={{ paddingLeft: `${(depth + 1) * 32 + 20}px` }}
                  >
                    <img
                      src={`https://images.evetech.net/types/${mat.typeId}/icon?size=32`}
                      alt=""
                      width={18}
                      height={18}
                      className="rounded-sm"
                    />
                    <span className="text-text-primary text-[13px]">
                      {mat.typeName}
                    </span>
                    <span className="text-text-muted text-xs ml-1">
                      x{mat.quantity}
                    </span>
                    {mat.isProduced ? (
                      <Badge className="ml-1 h-[18px] text-[10px] bg-[#1e3a5f] text-blue-science hover:bg-[#1e3a5f] cursor-default">
                        Produce
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="ml-1 h-[18px] text-[10px] border-[#334155] text-text-muted cursor-default">
                        Buy
                      </Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell className="text-text-muted text-xs">
                  <TooltipProvider>
                    {mat.hasBlueprint && mat.blueprintTypeId && (
                      detectedLevels[mat.blueprintTypeId] ? (
                        <Tooltip>
                          <TooltipTrigger>
                            <Badge className="h-[18px] text-[10px] bg-[#1a3a2a] text-teal-success hover:bg-[#1a3a2a] cursor-default">
                              ME {detectedLevels[mat.blueprintTypeId].materialEfficiency} / TE {detectedLevels[mat.blueprintTypeId].timeEfficiency}
                            </Badge>
                          </TooltipTrigger>
                          <TooltipContent>
                            Blueprint detected: ME {detectedLevels[mat.blueprintTypeId].materialEfficiency} / TE {detectedLevels[mat.blueprintTypeId].timeEfficiency} from {detectedLevels[mat.blueprintTypeId].ownerName}
                          </TooltipContent>
                        </Tooltip>
                      ) : Object.keys(detectedLevels).length > 0 ? (
                        <Tooltip>
                          <TooltipTrigger>
                            <AlertTriangle className="h-3.5 w-3.5 text-amber-manufacturing" data-testid="WarningAmberIcon" />
                          </TooltipTrigger>
                          <TooltipContent>No blueprint detected</TooltipContent>
                        </Tooltip>
                      ) : null
                    )}
                  </TooltipProvider>
                </TableCell>
                <TableCell />
                <TableCell />
                <TableCell className="text-center">
                  {mat.hasBlueprint && (
                    <button
                      className={`p-1 rounded hover:bg-overlay-subtle ${mat.isProduced ? "text-rose-danger" : "text-teal-success"}`}
                      onClick={() => handleToggleProduce(step.id, mat)}
                      title={mat.isProduced ? "Switch to Buy" : "Switch to Produce"}
                    >
                      {mat.isProduced ? (
                        <ShoppingCart className="h-4 w-4" />
                      ) : (
                        <Wrench className="h-4 w-4" />
                      )}
                    </button>
                  )}
                </TableCell>
              </TableRow>,
            );
          }
        }
      }
    }

    return rows;
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-2">
        <div className="flex items-center gap-2">
          <img
            src={`https://images.evetech.net/types/${plan.productTypeId}/icon?size=64`}
            alt=""
            width={40}
            height={40}
            className="rounded"
          />
          <div>
            {editingName ? (
              <div className="flex items-center gap-1">
                <Input
                  value={nameValue}
                  onChange={(e) => setNameValue(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") handleSaveName();
                    if (e.key === "Escape") setEditingName(false);
                  }}
                  autoFocus
                  className="min-w-[250px]"
                />
                <button className="p-1 rounded text-teal-success hover:bg-teal-success/10" onClick={handleSaveName}>
                  <Check className="h-4 w-4" />
                </button>
                <button className="p-1 rounded text-text-secondary hover:bg-overlay-subtle" onClick={() => setEditingName(false)}>
                  <X className="h-4 w-4" />
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-1">
                <h2 className="text-xl font-semibold text-text-emphasis">
                  {plan.name}
                </h2>
                <button
                  className="p-1 rounded text-text-muted hover:text-primary hover:bg-interactive-selected"
                  onClick={() => {
                    setNameValue(plan.name);
                    setEditingName(true);
                  }}
                >
                  <Pencil className="h-4 w-4" data-testid="EditIcon" />
                </button>
              </div>
            )}
            <p className="text-text-muted text-[13px]">
              {plan.steps?.length || 0} production step(s)
            </p>
          </div>
        </div>
        <Button
          className="bg-teal-success hover:bg-[#059669]"
          onClick={() => setGenerateDialogOpen(true)}
        >
          <Play className="h-4 w-4 mr-1" />
          Generate Jobs
        </Button>
      </div>

      <Tabs value={tab} onValueChange={setTab}>
        <TabsList className="mb-2">
          <TabsTrigger value="step-tree">Step Tree</TabsTrigger>
          <TabsTrigger value="batch-configure">Batch Configure</TabsTrigger>
          <TabsTrigger value="transport" className="flex items-center gap-1">
            <Truck className="h-4 w-4" />
            Transport
          </TabsTrigger>
        </TabsList>

        <TabsContent value="step-tree">
          <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
            <Table>
              <TableHeader>
                <TableRow className="bg-background-void">
                  <TableHead>Item / Material</TableHead>
                  <TableHead>ME / TE</TableHead>
                  <TableHead>Structure / Rig / Sec</TableHead>
                  <TableHead>Station</TableHead>
                  <TableHead className="text-center">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {rootStep ? (
                  renderStepRow(rootStep, 0, [0])
                ) : (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center">
                      <p className="text-text-muted">No steps in this plan</p>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </TabsContent>

        <TabsContent value="batch-configure">
          <BatchConfigureTab plan={plan} planId={planId} onUpdate={fetchPlan} detectedLevels={detectedLevels} />
        </TabsContent>

        <TabsContent value="transport">
          <TransportSettingsTab
            plan={plan}
            profiles={transportProfiles}
            onSave={handleSaveTransportSettings}
          />
        </TabsContent>
      </Tabs>

      {/* Edit Step Dialog */}
      {editStepId && (
        <EditStepDialog
          step={
            plan.steps?.find((s) => s.id === editStepId) || null
          }
          open={!!editStepId}
          onClose={() => setEditStepId(null)}
          onSave={(updates) => handleUpdateStep(editStepId, updates)}
          detectedLevel={
            (() => {
              const s = plan.steps?.find((st) => st.id === editStepId);
              return s ? (detectedLevels[s.blueprintTypeId] ?? null) : null;
            })()
          }
        />
      )}

      {/* Generate Jobs Dialog */}
      <Dialog
        open={generateDialogOpen}
        onOpenChange={(v) => {
          if (!v) {
            setGenerateDialogOpen(false);
            setGenerateResult(null);
            setPreviewResult(null);
            setPreviewError(null);
            setSelectedParallelism(0);
          }
        }}
      >
        <DialogContent className="max-w-2xl bg-background-panel border-overlay-medium">
          <DialogHeader>
            <DialogTitle className="text-text-emphasis">Generate Production Jobs</DialogTitle>
          </DialogHeader>
          {generateResult ? (
            <div>
              <p className="text-teal-success mb-1">
                Created {generateResult.created.length} job(s)
              </p>
              {generateResult.created.map((job) => (
                <p
                  key={job.id}
                  className="text-text-primary text-[13px] ml-4"
                >
                  {job.blueprintName || `BP ${job.blueprintTypeId}`} &mdash;{" "}
                  {job.runs} runs
                  {job.estimatedCost
                    ? ` (${formatISK(job.estimatedCost)})`
                    : ""}
                  {generateResult.characterAssignments?.[job.id] && (
                    <Badge className="ml-1 h-[18px] text-[11px] bg-[#1e3a5f] text-[#93c5fd] hover:bg-[#1e3a5f] cursor-default">
                      {generateResult.characterAssignments[job.id]}
                    </Badge>
                  )}
                </p>
              ))}
              {generateResult.unassignedCount != null && generateResult.unassignedCount > 0 && (
                <div className="mt-2 p-3 rounded bg-[#2d2000] border border-[rgba(251,191,36,0.3)] text-amber-manufacturing text-sm flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 shrink-0" />
                  {generateResult.unassignedCount} job(s) could not be assigned to a character
                </div>
              )}
              {generateResult.transportJobs?.length > 0 && (
                <>
                  <p className="text-primary mt-2 mb-1 flex items-center gap-1">
                    <Truck className="h-4 w-4" />
                    Created {generateResult.transportJobs.length} transport job(s)
                  </p>
                  {generateResult.transportJobs.map((tj) => (
                    <p
                      key={tj.id}
                      className="text-text-primary text-[13px] ml-4"
                    >
                      {tj.originStationName} &rarr; {tj.destinationStationName}
                      {" "}&mdash; {tj.items.length} item type(s), {formatNumber(tj.totalVolumeM3)} m&sup3;
                      {tj.estimatedCost ? ` (${formatISK(tj.estimatedCost)})` : ""}
                    </p>
                  ))}
                </>
              )}
              {generateResult.skipped.length > 0 && (
                <>
                  <p className="text-amber-manufacturing mt-2 mb-1">
                    Skipped {generateResult.skipped.length} item(s)
                  </p>
                  {generateResult.skipped.map((skip, i) => (
                    <p
                      key={i}
                      className="text-text-secondary text-[13px] ml-4"
                    >
                      {skip.typeName} &mdash; {skip.reason}
                    </p>
                  ))}
                </>
              )}
            </div>
          ) : (
            <div className="mt-1">
              <p className="text-text-secondary mb-2">
                How many {plan.productName || "units"} do you want to produce?
                Job queue entries will be created for each step in the
                production chain.
              </p>
              <div className="flex items-start gap-2">
                <Input
                  type="number"
                  value={generateQuantity}
                  onChange={(e) =>
                    setGenerateQuantity(Math.max(1, parseInt(e.target.value) || 1))
                  }
                  min={1}
                  className="w-40"
                />
                <Button
                  variant="outline"
                  onClick={() => handlePreview(generateQuantity)}
                  disabled={previewLoading}
                  className="text-primary border-primary hover:bg-interactive-selected"
                >
                  {previewLoading ? (
                    <span className="flex items-center gap-1">
                      <Loader2 className="h-4 w-4 animate-spin" />
                      Previewing...
                    </span>
                  ) : (
                    "Preview"
                  )}
                </Button>
              </div>

              {previewError && (
                <div className="mt-2 p-3 rounded bg-[#2d2000] border border-[rgba(251,191,36,0.3)] text-amber-manufacturing text-sm flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 shrink-0" />
                  {previewError} — you can still generate without parallelism.
                </div>
              )}

              {previewResult && (
                <div className="mt-2">
                  <p className="text-text-secondary text-[13px] mb-1">
                    Select how many characters to spread jobs across ({previewResult.eligibleCharacters} eligible character{previewResult.eligibleCharacters !== 1 ? "s" : ""}, {previewResult.totalJobs} total job{previewResult.totalJobs !== 1 ? "s" : ""})
                  </p>
                  <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
                    <Table>
                      <TableHeader>
                        <TableRow className="bg-background-void">
                          <TableHead className="w-10" />
                          <TableHead>Characters</TableHead>
                          <TableHead>Est. Time</TableHead>
                          <TableHead>Jobs</TableHead>
                          <TableHead>Details</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {/* No-assignment row */}
                        <TableRow
                          className={`cursor-pointer hover:bg-[#161c2c] ${selectedParallelism === 0 ? "bg-[#1e293b]" : ""}`}
                          onClick={() => setSelectedParallelism(0)}
                        >
                          <TableCell>
                            <input
                              type="radio"
                              name="parallelism"
                              value="0"
                              checked={selectedParallelism === 0}
                              onChange={() => setSelectedParallelism(0)}
                              className="accent-primary"
                            />
                          </TableCell>
                          <TableCell className="text-text-secondary">No assignment</TableCell>
                          <TableCell className="text-text-secondary">&mdash;</TableCell>
                          <TableCell className="text-text-secondary text-right">{previewResult.totalJobs}</TableCell>
                          <TableCell className="text-text-muted text-xs">Jobs created without character assignment</TableCell>
                        </TableRow>
                        {/* Parallelism options */}
                        {previewResult.options.map((option) => {
                          const isSelected = selectedParallelism === option.parallelism;
                          const detailsText = option.characters
                            .map((c) => {
                              const slots: string[] = [];
                              if (c.mfgSlotsMax > 0) slots.push(`${c.mfgSlotsUsed}/${c.mfgSlotsMax} mfg`);
                              if (c.reactSlotsMax > 0) slots.push(`${c.reactSlotsUsed}/${c.reactSlotsMax} react`);
                              return `${c.name} (${c.jobCount} job${c.jobCount !== 1 ? "s" : ""}${slots.length ? ", " + slots.join(", ") : ""})`;
                            })
                            .join("; ");
                          return (
                            <TableRow
                              key={option.parallelism}
                              className={`cursor-pointer hover:bg-[#161c2c] ${isSelected ? "bg-[#1e293b]" : ""}`}
                              onClick={() => setSelectedParallelism(option.parallelism)}
                            >
                              <TableCell>
                                <input
                                  type="radio"
                                  name="parallelism"
                                  value={String(option.parallelism)}
                                  checked={isSelected}
                                  onChange={() => setSelectedParallelism(option.parallelism)}
                                  className="accent-primary"
                                />
                              </TableCell>
                              <TableCell className="text-text-emphasis">{option.parallelism}</TableCell>
                              <TableCell className="text-teal-success">{option.estimatedDurationLabel}</TableCell>
                              <TableCell className="text-text-emphasis text-right">{previewResult.totalJobs}</TableCell>
                              <TableCell className="text-text-secondary text-xs">{detailsText}</TableCell>
                            </TableRow>
                          );
                        })}
                      </TableBody>
                    </Table>
                  </div>
                </div>
              )}
            </div>
          )}
          <DialogFooter>
            {generateResult ? (
              <Button
                onClick={() => {
                  setGenerateDialogOpen(false);
                  setGenerateResult(null);
                  setPreviewResult(null);
                  setPreviewError(null);
                  setSelectedParallelism(0);
                }}
              >
                Done
              </Button>
            ) : (
              <>
                <Button
                  variant="outline"
                  onClick={() => {
                    setGenerateDialogOpen(false);
                    setPreviewResult(null);
                    setPreviewError(null);
                    setSelectedParallelism(0);
                  }}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleGenerate}
                  disabled={generating}
                  className="bg-teal-success hover:bg-[#059669]"
                >
                  {generating ? "Generating..." : "Generate Jobs"}
                </Button>
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// --- Edit Step Dialog ---

type OwnerOption = {
  id: number;
  name: string;
  type: "character" | "corporation";
};

type DivisionOption = {
  number: number;
  name: string;
};

type ContainerOption = {
  id: number;
  name: string;
};

function EditStepDialog({
  step,
  open,
  onClose,
  onSave,
  detectedLevel,
}: {
  step: ProductionPlanStep | null;
  open: boolean;
  onClose: () => void;
  onSave: (updates: Partial<ProductionPlanStep>) => void;
  detectedLevel?: BlueprintLevel | null;
}) {
  const [meLevel, setMeLevel] = useState(step?.meLevel || 10);
  const [teLevel, setTeLevel] = useState(step?.teLevel || 20);
  const [industrySkill, setIndustrySkill] = useState(
    step?.industrySkill || 5,
  );
  const [advIndustrySkill, setAdvIndustrySkill] = useState(
    step?.advIndustrySkill || 5,
  );
  const [structure, setStructure] = useState(step?.structure || "raitaru");
  const [rig, setRig] = useState(step?.rig || "t2");
  const [security, setSecurity] = useState(step?.security || "high");
  const [facilityTax, setFacilityTax] = useState(step?.facilityTax || 1.0);
  const [stationName, setStationName] = useState(step?.stationName || "");

  const [userStations, setUserStations] = useState<UserStation[]>([]);
  const [selectedUserStation, setSelectedUserStation] = useState<UserStation | null>(null);
  const [stationsLoaded, setStationsLoaded] = useState(false);

  const [hangarsData, setHangarsData] = useState<HangarsResponse | null>(null);
  const [hangarsLoaded, setHangarsLoaded] = useState(false);
  const [selectedOwner, setSelectedOwner] = useState<OwnerOption | null>(null);
  const [selectedDivision, setSelectedDivision] = useState<DivisionOption | null>(null);
  const [selectedContainer, setSelectedContainer] = useState<ContainerOption | null>(null);

  const [selectedOutputOwner, setSelectedOutputOwner] = useState<OwnerOption | null>(null);
  const [selectedOutputDivision, setSelectedOutputDivision] = useState<DivisionOption | null>(null);
  const [selectedOutputContainer, setSelectedOutputContainer] = useState<ContainerOption | null>(null);

  useEffect(() => {
    if (step) {
      setMeLevel(step.meLevel);
      setTeLevel(step.teLevel);
      setIndustrySkill(step.industrySkill);
      setAdvIndustrySkill(step.advIndustrySkill);
      setStructure(step.structure);
      setRig(step.rig);
      setSecurity(step.security);
      setFacilityTax(step.facilityTax);
      setStationName(step.stationName || "");
    }
  }, [step]);

  useEffect(() => {
    if (!open || stationsLoaded) return;
    const fetchStations = async () => {
      try {
        const res = await fetch("/api/stations/user-stations");
        if (res.ok) {
          const data: UserStation[] = await res.json();
          setUserStations(data || []);
          if (step?.userStationId) {
            const match = (data || []).find((s) => s.id === step.userStationId);
            if (match) setSelectedUserStation(match);
          }
        }
      } catch (err) {
        console.error("Failed to fetch user stations:", err);
      } finally {
        setStationsLoaded(true);
      }
    };
    fetchStations();
  }, [open, stationsLoaded, step]);

  const stationIdForHangars = selectedUserStation?.id;
  useEffect(() => {
    if (!open || !stationIdForHangars) {
      setHangarsData(null);
      setHangarsLoaded(false);
      return;
    }
    const fetchHangars = async () => {
      try {
        const res = await fetch(
          `/api/industry/plans/hangars?user_station_id=${stationIdForHangars}`,
        );
        if (res.ok) {
          const data: HangarsResponse = await res.json();
          setHangarsData(data);

          if (step?.sourceOwnerType && step?.sourceOwnerId) {
            const ownerMatch = step.sourceOwnerType === "character"
              ? data.characters.find((c) => c.id === step.sourceOwnerId)
              : data.corporations.find((c) => c.id === step.sourceOwnerId);
            if (ownerMatch) {
              setSelectedOwner({
                id: ownerMatch.id,
                name: ownerMatch.name,
                type: step.sourceOwnerType as "character" | "corporation",
              });

              if (step.sourceOwnerType === "corporation" && step.sourceDivisionNumber != null) {
                const corp = data.corporations.find((c) => c.id === step.sourceOwnerId);
                if (corp) {
                  const divName = corp.divisions[String(step.sourceDivisionNumber)] || `Division ${step.sourceDivisionNumber}`;
                  setSelectedDivision({ number: step.sourceDivisionNumber, name: divName });
                }
              }

              if (step.sourceContainerId) {
                const container = data.containers.find((c) => c.id === step.sourceContainerId);
                if (container) {
                  setSelectedContainer({ id: container.id, name: container.name });
                }
              }
            }
          }

          if (step?.outputOwnerType && step?.outputOwnerId) {
            const outOwnerMatch = step.outputOwnerType === "character"
              ? data.characters.find((c) => c.id === step.outputOwnerId)
              : data.corporations.find((c) => c.id === step.outputOwnerId);
            if (outOwnerMatch) {
              setSelectedOutputOwner({
                id: outOwnerMatch.id,
                name: outOwnerMatch.name,
                type: step.outputOwnerType as "character" | "corporation",
              });

              if (step.outputOwnerType === "corporation" && step.outputDivisionNumber != null) {
                const corp = data.corporations.find((c) => c.id === step.outputOwnerId);
                if (corp) {
                  const divName = corp.divisions[String(step.outputDivisionNumber)] || `Division ${step.outputDivisionNumber}`;
                  setSelectedOutputDivision({ number: step.outputDivisionNumber, name: divName });
                }
              }

              if (step.outputContainerId) {
                const container = data.containers.find((c) => c.id === step.outputContainerId);
                if (container) {
                  setSelectedOutputContainer({ id: container.id, name: container.name });
                }
              }
            }
          }
        }
      } catch (err) {
        console.error("Failed to fetch hangars:", err);
      } finally {
        setHangarsLoaded(true);
      }
    };
    fetchHangars();
  }, [open, stationIdForHangars, step]);

  useEffect(() => {
    if (!open) {
      setStationsLoaded(false);
      setSelectedUserStation(null);
      setHangarsData(null);
      setHangarsLoaded(false);
      setSelectedOwner(null);
      setSelectedDivision(null);
      setSelectedContainer(null);
      setSelectedOutputOwner(null);
      setSelectedOutputDivision(null);
      setSelectedOutputContainer(null);
    }
  }, [open]);

  const handleStationSelect = (stationId: string) => {
    if (stationId === "none") {
      setSelectedUserStation(null);
      setSelectedOwner(null);
      setSelectedDivision(null);
      setSelectedContainer(null);
      setSelectedOutputOwner(null);
      setSelectedOutputDivision(null);
      setSelectedOutputContainer(null);
      return;
    }
    const station = userStations.find((s) => String(s.id) === stationId) || null;
    setSelectedUserStation(station);
    setSelectedOwner(null);
    setSelectedDivision(null);
    setSelectedContainer(null);
    setSelectedOutputOwner(null);
    setSelectedOutputDivision(null);
    setSelectedOutputContainer(null);
    if (station) {
      setStructure(station.structure);
      setFacilityTax(station.facilityTax);
      setSecurity(station.security || "high");
      setStationName(station.stationName || "");
      if (step?.rigCategory) {
        const matchingRig = station.rigs.find(
          (r) => r.category === step.rigCategory,
        );
        setRig(matchingRig ? matchingRig.tier : "none");
      } else {
        setRig("none");
      }
    }
  };

  const ownerOptions: OwnerOption[] = [];
  if (hangarsData) {
    for (const char of hangarsData.characters) {
      ownerOptions.push({ id: char.id, name: char.name, type: "character" });
    }
    for (const corp of hangarsData.corporations) {
      ownerOptions.push({ id: corp.id, name: corp.name, type: "corporation" });
    }
  }

  const divisionOptions: DivisionOption[] = [];
  if (selectedOwner?.type === "corporation" && hangarsData) {
    const corp = hangarsData.corporations.find((c) => c.id === selectedOwner.id);
    if (corp) {
      for (const [num, name] of Object.entries(corp.divisions)) {
        divisionOptions.push({ number: parseInt(num), name });
      }
      divisionOptions.sort((a, b) => a.number - b.number);
    }
  }

  const containerOptions: ContainerOption[] = [];
  if (selectedOwner && hangarsData) {
    for (const c of hangarsData.containers) {
      if (c.ownerType !== selectedOwner.type || c.ownerId !== selectedOwner.id) continue;
      if (selectedOwner.type === "corporation" && selectedDivision) {
        if (c.divisionNumber !== selectedDivision.number) continue;
      }
      containerOptions.push({ id: c.id, name: c.name });
    }
  }

  const handleOwnerSelect = (val: string) => {
    if (val === "none") {
      setSelectedOwner(null);
      setSelectedDivision(null);
      setSelectedContainer(null);
      return;
    }
    const owner = ownerOptions.find((o) => `${o.type}-${o.id}` === val) || null;
    setSelectedOwner(owner);
    setSelectedDivision(null);
    setSelectedContainer(null);
  };

  const handleDivisionSelect = (val: string) => {
    if (val === "none") {
      setSelectedDivision(null);
      setSelectedContainer(null);
      return;
    }
    const division = divisionOptions.find((d) => String(d.number) === val) || null;
    setSelectedDivision(division);
    setSelectedContainer(null);
  };

  const outputDivisionOptions: DivisionOption[] = [];
  if (selectedOutputOwner?.type === "corporation" && hangarsData) {
    const corp = hangarsData.corporations.find((c) => c.id === selectedOutputOwner.id);
    if (corp) {
      for (const [num, name] of Object.entries(corp.divisions)) {
        outputDivisionOptions.push({ number: parseInt(num), name });
      }
      outputDivisionOptions.sort((a, b) => a.number - b.number);
    }
  }

  const outputContainerOptions: ContainerOption[] = [];
  if (selectedOutputOwner && hangarsData) {
    for (const c of hangarsData.containers) {
      if (c.ownerType !== selectedOutputOwner.type || c.ownerId !== selectedOutputOwner.id) continue;
      if (selectedOutputOwner.type === "corporation" && selectedOutputDivision) {
        if (c.divisionNumber !== selectedOutputDivision.number) continue;
      }
      outputContainerOptions.push({ id: c.id, name: c.name });
    }
  }

  const handleOutputOwnerSelect = (val: string) => {
    if (val === "none") {
      setSelectedOutputOwner(null);
      setSelectedOutputDivision(null);
      setSelectedOutputContainer(null);
      return;
    }
    const owner = ownerOptions.find((o) => `${o.type}-${o.id}` === val) || null;
    setSelectedOutputOwner(owner);
    setSelectedOutputDivision(null);
    setSelectedOutputContainer(null);
  };

  const handleOutputDivisionSelect = (val: string) => {
    if (val === "none") {
      setSelectedOutputDivision(null);
      setSelectedOutputContainer(null);
      return;
    }
    const division = outputDivisionOptions.find((d) => String(d.number) === val) || null;
    setSelectedOutputDivision(division);
    setSelectedOutputContainer(null);
  };

  const filteredStations = userStations.filter((s) =>
    step?.activity ? s.activities.includes(step.activity) : true,
  );

  const hasStation = !!selectedUserStation;

  const resolvedSourceLocationId = selectedUserStation
    ? selectedUserStation.stationId
    : null;

  if (!step) return null;

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md bg-background-panel border-overlay-medium max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-text-emphasis">
            Edit Step: {step.productName || `Type ${step.productTypeId}`}
          </DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-3 pt-1">
          {/* Preferred Station Selection */}
          <div>
            <Label className="text-sm text-text-secondary mb-1 block">Preferred Station</Label>
            <Select
              value={selectedUserStation ? String(selectedUserStation.id) : "none"}
              onValueChange={handleStationSelect}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select a saved station or leave empty" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">None (manual config)</SelectItem>
                {filteredStations.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>
                    {s.stationName || "Unknown"} ({s.solarSystemName})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label htmlFor="edit-me-level" className="text-sm text-text-secondary mb-1 block">ME Level</Label>
              <Input
                id="edit-me-level"
                type="number"
                value={meLevel}
                onChange={(e) => setMeLevel(parseInt(e.target.value) || 0)}
                min={0}
                max={10}
              />
            </div>
            <div>
              <Label htmlFor="edit-te-level" className="text-sm text-text-secondary mb-1 block">TE Level</Label>
              <Input
                id="edit-te-level"
                type="number"
                value={teLevel}
                onChange={(e) => setTeLevel(parseInt(e.target.value) || 0)}
                min={0}
                max={20}
              />
            </div>
            <div>
              <Label className="text-sm text-text-secondary mb-1 block">Industry Skill</Label>
              <Input
                type="number"
                value={industrySkill}
                onChange={(e) => setIndustrySkill(parseInt(e.target.value) || 0)}
                min={0}
                max={5}
              />
            </div>
            <div>
              <Label className="text-sm text-text-secondary mb-1 block">Adv. Industry Skill</Label>
              <Input
                type="number"
                value={advIndustrySkill}
                onChange={(e) => setAdvIndustrySkill(parseInt(e.target.value) || 0)}
                min={0}
                max={5}
              />
            </div>
            <div>
              <Label className="text-sm text-text-secondary mb-1 block">Structure</Label>
              <Select value={structure} onValueChange={setStructure} disabled={hasStation}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="station">Station</SelectItem>
                  <SelectItem value="raitaru">Raitaru</SelectItem>
                  <SelectItem value="azbel">Azbel</SelectItem>
                  <SelectItem value="sotiyo">Sotiyo</SelectItem>
                  <SelectItem value="athanor">Athanor</SelectItem>
                  <SelectItem value="tatara">Tatara</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label className="text-sm text-text-secondary mb-1 block">Rig</Label>
              <Select value={rig} onValueChange={setRig} disabled={hasStation}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">None</SelectItem>
                  <SelectItem value="t1">T1</SelectItem>
                  <SelectItem value="t2">T2</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label className="text-sm text-text-secondary mb-1 block">Security</Label>
              <Select value={security} onValueChange={setSecurity} disabled={hasStation}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="high">Highsec</SelectItem>
                  <SelectItem value="low">Lowsec</SelectItem>
                  <SelectItem value="null">Nullsec / WH</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label className="text-sm text-text-secondary mb-1 block">Facility Tax %</Label>
              <Input
                type="number"
                value={facilityTax}
                onChange={(e) => setFacilityTax(parseFloat(e.target.value) || 0)}
                min={0}
                step={0.1}
                disabled={hasStation}
              />
            </div>
            <div className="col-span-2">
              <Label className="text-sm text-text-secondary mb-1 block">Station Name</Label>
              <Input
                value={stationName}
                onChange={(e) => setStationName(e.target.value)}
                placeholder="e.g. Jita 4-4 or player structure name"
                disabled={hasStation}
              />
            </div>
          </div>

          {/* Detected Blueprint Info */}
          {detectedLevel ? (
            <div className="flex items-center gap-2 flex-wrap">
              <Badge variant="outline" className="text-[11px] border-[#0ea5e9] text-[#38bdf8]">
                {step?.activity === "reaction"
                  ? `Blueprint detected from ${detectedLevel.ownerName}${detectedLevel.isCopy ? " (BPC)" : ""}`
                  : `Blueprint detected: ME ${detectedLevel.materialEfficiency} / TE ${detectedLevel.timeEfficiency} (${detectedLevel.ownerName}${detectedLevel.isCopy ? ", BPC" : ""})`}
              </Badge>
              {step?.activity !== "reaction" && (
                <Button
                  size="sm"
                  variant="outline"
                  className="text-[11px] py-0.5 px-2 h-auto text-primary border-primary"
                  onClick={() => {
                    setMeLevel(detectedLevel.materialEfficiency);
                    setTeLevel(detectedLevel.timeEfficiency);
                  }}
                >
                  Apply
                </Button>
              )}
            </div>
          ) : (
            <div className="flex items-center gap-1">
              <AlertTriangle className="h-3.5 w-3.5 text-amber-manufacturing" />
              <span className="text-[11px] text-amber-manufacturing">No blueprint detected — using manual values</span>
            </div>
          )}

          {/* Input Location Section */}
          {hasStation && (
            <>
              <Separator className="border-[#1e293b] mt-1" />
              <p className="text-text-secondary text-sm font-semibold">Input Location</p>
              <p className="text-text-muted text-xs">
                Where should materials for this step be pulled from?
              </p>

              {!hangarsLoaded ? (
                <p className="text-text-muted text-[13px]">Loading hangars...</p>
              ) : (
                <div className="flex flex-col gap-3">
                  <div>
                    <Label className="text-sm text-text-secondary mb-1 block">Owner</Label>
                    <Select
                      value={selectedOwner ? `${selectedOwner.type}-${selectedOwner.id}` : "none"}
                      onValueChange={handleOwnerSelect}
                    >
                      <SelectTrigger><SelectValue placeholder="Select character or corporation" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="none">None</SelectItem>
                        {ownerOptions.map((o) => (
                          <SelectItem key={`${o.type}-${o.id}`} value={`${o.type}-${o.id}`}>
                            {o.name} ({o.type === "character" ? "Character" : "Corporation"})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {selectedOwner?.type === "corporation" && divisionOptions.length > 0 && (
                    <div>
                      <Label className="text-sm text-text-secondary mb-1 block">Hangar Division</Label>
                      <Select
                        value={selectedDivision ? String(selectedDivision.number) : "none"}
                        onValueChange={handleDivisionSelect}
                      >
                        <SelectTrigger><SelectValue placeholder="Select division" /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="none">None</SelectItem>
                          {divisionOptions.map((d) => (
                            <SelectItem key={d.number} value={String(d.number)}>
                              {d.number}. {d.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}

                  {selectedOwner && (
                    <div>
                      <Label className="text-sm text-text-secondary mb-1 block">Container (optional)</Label>
                      <Select
                        value={selectedContainer ? String(selectedContainer.id) : "none"}
                        onValueChange={(val) => {
                          if (val === "none") {
                            setSelectedContainer(null);
                          } else {
                            const c = containerOptions.find((c) => String(c.id) === val);
                            setSelectedContainer(c || null);
                          }
                        }}
                      >
                        <SelectTrigger><SelectValue placeholder="Select container or leave empty" /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="none">None (hangar)</SelectItem>
                          {containerOptions.map((c) => (
                            <SelectItem key={c.id} value={String(c.id)}>{c.name}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </div>
              )}

              <Separator className="border-[#1e293b] mt-1" />
              <p className="text-text-secondary text-sm font-semibold">Output Location</p>
              <p className="text-text-muted text-xs">
                Where should completed items from this job be delivered?
              </p>

              {!hangarsLoaded ? (
                <p className="text-text-muted text-[13px]">Loading hangars...</p>
              ) : (
                <div className="flex flex-col gap-3">
                  <div>
                    <Label className="text-sm text-text-secondary mb-1 block">Owner</Label>
                    <Select
                      value={selectedOutputOwner ? `${selectedOutputOwner.type}-${selectedOutputOwner.id}` : "none"}
                      onValueChange={handleOutputOwnerSelect}
                    >
                      <SelectTrigger><SelectValue placeholder="Select character or corporation" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="none">None</SelectItem>
                        {ownerOptions.map((o) => (
                          <SelectItem key={`out-${o.type}-${o.id}`} value={`${o.type}-${o.id}`}>
                            {o.name} ({o.type === "character" ? "Character" : "Corporation"})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {selectedOutputOwner?.type === "corporation" && outputDivisionOptions.length > 0 && (
                    <div>
                      <Label className="text-sm text-text-secondary mb-1 block">Hangar Division</Label>
                      <Select
                        value={selectedOutputDivision ? String(selectedOutputDivision.number) : "none"}
                        onValueChange={handleOutputDivisionSelect}
                      >
                        <SelectTrigger><SelectValue placeholder="Select division" /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="none">None</SelectItem>
                          {outputDivisionOptions.map((d) => (
                            <SelectItem key={d.number} value={String(d.number)}>
                              {d.number}. {d.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}

                  {selectedOutputOwner && (
                    <div>
                      <Label className="text-sm text-text-secondary mb-1 block">Container (optional)</Label>
                      <Select
                        value={selectedOutputContainer ? String(selectedOutputContainer.id) : "none"}
                        onValueChange={(val) => {
                          if (val === "none") {
                            setSelectedOutputContainer(null);
                          } else {
                            const c = outputContainerOptions.find((c) => String(c.id) === val);
                            setSelectedOutputContainer(c || null);
                          }
                        }}
                      >
                        <SelectTrigger><SelectValue placeholder="Select container or leave empty" /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="none">None (hangar)</SelectItem>
                          {outputContainerOptions.map((c) => (
                            <SelectItem key={c.id} value={String(c.id)}>{c.name}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </div>
              )}
            </>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={() =>
              onSave({
                me_level: meLevel,
                te_level: teLevel,
                industry_skill: industrySkill,
                adv_industry_skill: advIndustrySkill,
                structure,
                rig,
                security,
                facility_tax: facilityTax,
                station_name: stationName || null,
                user_station_id: selectedUserStation?.id || null,
                source_owner_type: selectedOwner?.type || null,
                source_owner_id: selectedOwner?.id || null,
                source_division_number:
                  selectedOwner?.type === "corporation"
                    ? selectedDivision?.number ?? null
                    : null,
                source_container_id: selectedContainer?.id || null,
                source_location_id: resolvedSourceLocationId,
                output_owner_type: selectedOutputOwner?.type || null,
                output_owner_id: selectedOutputOwner?.id || null,
                output_division_number:
                  selectedOutputOwner?.type === "corporation"
                    ? selectedOutputDivision?.number ?? null
                    : null,
                output_container_id: selectedOutputContainer?.id || null,
              } as unknown as Partial<ProductionPlanStep>)
            }
          >
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function TransportSettingsTab({
  plan,
  profiles,
  onSave,
}: {
  plan: ProductionPlan;
  profiles: { id: number; name: string; transportMethod: string }[];
  onSave: (settings: {
    transport_fulfillment?: string;
    transport_method?: string;
    transport_profile_id?: number;
    courier_rate_per_m3: number;
    courier_collateral_rate: number;
  }) => void;
}) {
  const [fulfillment, setFulfillment] = useState<string>(plan.transportFulfillment || "none");
  const [method, setMethod] = useState<string>(plan.transportMethod || "none");
  const [profileId, setProfileId] = useState<string>(plan.transportProfileId ? String(plan.transportProfileId) : "none");
  const [courierRate, setCourierRate] = useState(plan.courierRatePerM3 || 0);
  const [collateralRate, setCollateralRate] = useState(plan.courierCollateralRate || 0);

  const filteredProfiles = profiles.filter(
    (p) => method === "none" || !method || p.transportMethod === method
  );

  const handleSave = () => {
    onSave({
      transport_fulfillment: fulfillment !== "none" ? fulfillment : undefined,
      transport_method: fulfillment === "self_haul" && method !== "none" ? method : undefined,
      transport_profile_id: fulfillment === "self_haul" && profileId !== "none" ? Number(profileId) : undefined,
      courier_rate_per_m3: courierRate,
      courier_collateral_rate: collateralRate,
    });
  };

  return (
    <div className="bg-background-panel rounded-sm border border-overlay-subtle p-4">
      <h3 className="text-lg font-semibold text-text-emphasis mb-2 flex items-center gap-2">
        <Truck className="h-5 w-5" />
        Transport Settings
      </h3>
      <p className="text-text-muted text-[13px] mb-3">
        Configure how items should be transported between stations when generating jobs.
        Leave fulfillment type as &ldquo;None&rdquo; to skip transport job generation.
      </p>

      <div className="flex flex-col gap-3 max-w-[500px]">
        <div>
          <Label className="text-sm text-text-secondary mb-1 block">Fulfillment Type</Label>
          <Select
            value={fulfillment}
            onValueChange={(val) => {
              setFulfillment(val);
              if (val !== "self_haul") {
                setMethod("none");
                setProfileId("none");
              }
            }}
          >
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="none">None</SelectItem>
              <SelectItem value="self_haul">Self Haul</SelectItem>
              <SelectItem value="courier_contract">Courier Contract</SelectItem>
              <SelectItem value="contact_haul">Contact Haul</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {fulfillment === "self_haul" && (
          <>
            <div>
              <Label className="text-sm text-text-secondary mb-1 block">Transport Method</Label>
              <Select
                value={method}
                onValueChange={(val) => {
                  setMethod(val);
                  setProfileId("none");
                }}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">Select method...</SelectItem>
                  <SelectItem value="freighter">Freighter</SelectItem>
                  <SelectItem value="jump_freighter">Jump Freighter</SelectItem>
                  <SelectItem value="dst">DST</SelectItem>
                  <SelectItem value="blockade_runner">Blockade Runner</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label className="text-sm text-text-secondary mb-1 block">Transport Profile</Label>
              <Select value={profileId} onValueChange={setProfileId}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">Select profile...</SelectItem>
                  {filteredProfiles.map((p) => (
                    <SelectItem key={p.id} value={String(p.id)}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </>
        )}

        {(fulfillment === "courier_contract" || fulfillment === "contact_haul") && (
          <>
            <div>
              <Label className="text-sm text-text-secondary mb-1 block">Rate per m3 (ISK)</Label>
              <Input
                type="number"
                value={courierRate}
                onChange={(e) => setCourierRate(parseFloat(e.target.value) || 0)}
                min={0}
                step={0.01}
              />
            </div>
            <div>
              <Label className="text-sm text-text-secondary mb-1 block">Collateral Rate (%)</Label>
              <Input
                type="number"
                value={collateralRate * 100}
                onChange={(e) => setCollateralRate((parseFloat(e.target.value) || 0) / 100)}
                min={0}
                step={0.1}
              />
            </div>
          </>
        )}

        <Button
          onClick={handleSave}
          className="self-start"
        >
          Save Transport Settings
        </Button>
      </div>
    </div>
  );
}

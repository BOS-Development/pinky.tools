import { useState, useEffect } from "react";
import {
  ProductionPlan,
  ProductionPlanStep,
  PlanMaterial,
  UserStation,
  HangarsResponse,
  BlueprintLevel,
} from "@industry-tool/client/data/models";
import {
  ChevronDown,
  ChevronRight,
  Wrench,
  Pencil,
  ShoppingCart,
  Sparkles,
  Loader2,
  AlertTriangle,
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
import { toast } from "@/components/ui/sonner";

type Props = {
  plan: ProductionPlan;
  planId: number;
  onUpdate: () => void;
  detectedLevels?: Record<number, BlueprintLevel>;
};

type StepGroup = {
  productTypeId: number;
  productName: string;
  activity: string;
  steps: ProductionPlanStep[];
  stepIds: number[];
  meLevel: number | "mixed";
  teLevel: number | "mixed";
  structure: string | "mixed";
  rig: string | "mixed";
  security: string | "mixed";
  stationName: string | null | "mixed";
  userStationId: number | null | "mixed";
  inputLocation: string | null | "mixed";
  outputLocation: string | null | "mixed";
};

function groupSteps(steps: ProductionPlanStep[]): StepGroup[] {
  const groupMap = new Map<string, ProductionPlanStep[]>();

  for (const step of steps) {
    const key = `${step.productTypeId}:${step.activity}`;
    const existing = groupMap.get(key) || [];
    existing.push(step);
    groupMap.set(key, existing);
  }

  const groups: StepGroup[] = [];
  for (const [, stepsInGroup] of groupMap) {
    const first = stepsInGroup[0];

    const uniform = <T,>(getter: (s: ProductionPlanStep) => T): T | "mixed" => {
      const first = getter(stepsInGroup[0]);
      return stepsInGroup.every((s) => getter(s) === first) ? first : "mixed";
    };

    groups.push({
      productTypeId: first.productTypeId,
      productName: first.productName || `Type ${first.productTypeId}`,
      activity: first.activity,
      steps: stepsInGroup,
      stepIds: stepsInGroup.map((s) => s.id),
      meLevel: uniform((s) => s.meLevel),
      teLevel: uniform((s) => s.teLevel),
      structure: uniform((s) => s.structure),
      rig: uniform((s) => s.rig),
      security: uniform((s) => s.security),
      stationName: uniform((s) => s.stationName || null),
      userStationId: uniform((s) => s.userStationId || null),
      inputLocation: uniform((s) => {
        if (!s.sourceOwnerName) return null;
        let loc = s.sourceOwnerName;
        if (s.sourceDivisionName) loc += ` / ${s.sourceDivisionName}`;
        if (s.sourceContainerName) loc += ` / ${s.sourceContainerName}`;
        return loc;
      }),
      outputLocation: uniform((s) => {
        if (!s.outputOwnerName) return null;
        let loc = s.outputOwnerName;
        if (s.outputDivisionName) loc += ` / ${s.outputDivisionName}`;
        if (s.outputContainerName) loc += ` / ${s.outputContainerName}`;
        return loc;
      }),
    });
  }

  groups.sort((a, b) => a.productName.localeCompare(b.productName));
  return groups;
}

type MaterialStatus = "all" | "none" | "mixed";

type GroupMaterial = PlanMaterial & {
  produceStatus: MaterialStatus;
};

const MixedBadge = () => (
  <Badge className="h-[18px] text-[10px] bg-[#422006] text-amber-manufacturing hover:bg-[#422006] cursor-default">
    Mixed
  </Badge>
);

export default function BatchConfigureTab({ plan, planId, onUpdate, detectedLevels = {} }: Props) {
  const [editGroup, setEditGroup] = useState<StepGroup | null>(null);
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set());
  const [groupMaterials, setGroupMaterials] = useState<Record<string, GroupMaterial[]>>({});
  const [loadingMaterials, setLoadingMaterials] = useState<Set<string>>(new Set());
  const [togglingMaterial, setTogglingMaterial] = useState<string | null>(null);
  const [togglingAllGroup, setTogglingAllGroup] = useState<string | null>(null);
  const [applyingBlueprintGroup, setApplyingBlueprintGroup] = useState<string | null>(null);

  const steps = plan.steps || [];
  const groups = groupSteps(steps);

  const groupKey = (g: StepGroup) => `${g.productTypeId}:${g.activity}`;

  const computeProduceStatus = (group: StepGroup, materialTypeId: number): MaterialStatus => {
    let hasProduced = false;
    let hasMissing = false;
    for (const step of group.steps) {
      const hasChild = steps.some(
        (s) => s.parentStepId === step.id && s.productTypeId === materialTypeId,
      );
      if (hasChild) hasProduced = true;
      else hasMissing = true;
    }
    if (hasProduced && !hasMissing) return "all";
    if (!hasProduced && hasMissing) return "none";
    return "mixed";
  };

  const fetchMaterialsForGroup = async (group: StepGroup) => {
    const key = groupKey(group);
    const firstStep = group.steps[0];
    setLoadingMaterials((prev) => new Set([...prev, key]));
    try {
      const res = await fetch(
        `/api/industry/plans/${planId}/steps/${firstStep.id}/materials`,
      );
      if (res.ok) {
        const data: PlanMaterial[] = await res.json();
        const enriched: GroupMaterial[] = (data || []).map((mat) => ({
          ...mat,
          produceStatus: computeProduceStatus(group, mat.typeId),
        }));
        setGroupMaterials((prev) => ({ ...prev, [key]: enriched }));
      }
    } catch (err) {
      console.error("Failed to fetch group materials:", err);
    } finally {
      setLoadingMaterials((prev) => {
        const next = new Set(prev);
        next.delete(key);
        return next;
      });
    }
  };

  const toggleExpand = (group: StepGroup) => {
    const key = groupKey(group);
    setExpandedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
        if (!groupMaterials[key]) {
          fetchMaterialsForGroup(group);
        }
      }
      return next;
    });
  };

  useEffect(() => {
    for (const group of groups) {
      const key = groupKey(group);
      if (groupMaterials[key]) {
        setGroupMaterials((prev) => ({
          ...prev,
          [key]: prev[key].map((mat) => ({
            ...mat,
            produceStatus: computeProduceStatus(group, mat.typeId),
          })),
        }));
      } else if (!loadingMaterials.has(key)) {
        fetchMaterialsForGroup(group);
      }
    }
  }, [steps.length]);

  const handleToggleProduce = async (group: StepGroup, material: PlanMaterial) => {
    const matKey = `${groupKey(group)}:${material.typeId}`;
    setTogglingMaterial(matKey);

    const status = computeProduceStatus(group, material.typeId);

    try {
      if (status === "all" || status === "mixed") {
        const childSteps = steps.filter(
          (s) => group.stepIds.includes(s.parentStepId!) && s.productTypeId === material.typeId,
        );
        for (const child of childSteps) {
          await fetch(`/api/industry/plans/${planId}/steps/${child.id}`, {
            method: "DELETE",
          });
        }
        toast.success(`Set ${material.typeName} to Buy across ${group.stepIds.length} step(s)`);
      } else {
        for (const step of group.steps) {
          const hasChild = steps.some(
            (s) => s.parentStepId === step.id && s.productTypeId === material.typeId,
          );
          if (!hasChild) {
            await fetch(`/api/industry/plans/${planId}/steps`, {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({
                parent_step_id: step.id,
                product_type_id: material.typeId,
              }),
            });
          }
        }
        toast.success(`Set ${material.typeName} to Produce across ${group.stepIds.length} step(s)`);
      }
      onUpdate();
    } catch (err) {
      console.error("Failed to toggle produce:", err);
      toast.error("Failed to toggle produce/buy");
    } finally {
      setTogglingMaterial(null);
    }
  };

  const handleSetAllBuild = async (group: StepGroup) => {
    const key = groupKey(group);
    setTogglingAllGroup(key);
    try {
      let mats = groupMaterials[key];
      if (!mats) {
        const firstStep = group.steps[0];
        const res = await fetch(
          `/api/industry/plans/${planId}/steps/${firstStep.id}/materials`,
        );
        if (!res.ok) throw new Error("Failed to fetch materials");
        const data: PlanMaterial[] = await res.json();
        mats = (data || []).map((mat) => ({
          ...mat,
          produceStatus: computeProduceStatus(group, mat.typeId),
        }));
        setGroupMaterials((prev) => ({ ...prev, [key]: mats! }));
      }

      const buildableMats = mats.filter((m) => m.hasBlueprint);
      let created = 0;
      for (const mat of buildableMats) {
        for (const step of group.steps) {
          const hasChild = steps.some(
            (s) => s.parentStepId === step.id && s.productTypeId === mat.typeId,
          );
          if (!hasChild) {
            await fetch(`/api/industry/plans/${planId}/steps`, {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({
                parent_step_id: step.id,
                product_type_id: mat.typeId,
              }),
            });
            created++;
          }
        }
      }
      toast.success(`Set all to Build: created ${created} step(s) across ${group.stepIds.length} ${group.productName} step(s)`);
      onUpdate();
    } catch (err) {
      console.error("Failed to set all to build:", err);
      toast.error("Failed to set all to build");
    } finally {
      setTogglingAllGroup(null);
    }
  };

  const handleSetAllBuy = async (group: StepGroup) => {
    const key = groupKey(group);
    setTogglingAllGroup(key);
    try {
      const childSteps = steps.filter(
        (s) => s.parentStepId && group.stepIds.includes(s.parentStepId),
      );
      for (const child of childSteps) {
        await fetch(`/api/industry/plans/${planId}/steps/${child.id}`, {
          method: "DELETE",
        });
      }
      toast.success(`Set all to Buy: removed ${childSteps.length} step(s) from ${group.stepIds.length} ${group.productName} step(s)`);
      onUpdate();
    } catch (err) {
      console.error("Failed to set all to buy:", err);
      toast.error("Failed to set all to buy");
    } finally {
      setTogglingAllGroup(null);
    }
  };

  const handleApplyBlueprintLevels = async (group: StepGroup) => {
    const key = groupKey(group);
    setApplyingBlueprintGroup(key);
    try {
      const blueprintTypeIds = [...new Set(group.steps.map((s) => s.blueprintTypeId))];

      let updatedCount = 0;
      for (const blueprintTypeId of blueprintTypeIds) {
        const detected = detectedLevels[blueprintTypeId];
        if (!detected) continue;

        const matchingStepIds = group.steps
          .filter((s) => s.blueprintTypeId === blueprintTypeId)
          .map((s) => s.id);

        if (matchingStepIds.length === 0) continue;

        const res = await fetch(`/api/industry/plans/${planId}/steps/batch`, {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            step_ids: matchingStepIds,
            me_level: detected.materialEfficiency,
            te_level: detected.timeEfficiency,
          }),
        });
        if (res.ok) {
          updatedCount += matchingStepIds.length;
        }
      }

      if (updatedCount > 0) {
        onUpdate();
        toast.success(`Applied detected ME/TE to ${updatedCount} step(s)`);
      } else {
        toast.error("No detected blueprint levels found for this group");
      }
    } catch (err) {
      console.error("Failed to apply blueprint levels:", err);
      toast.error("Failed to apply blueprint levels");
    } finally {
      setApplyingBlueprintGroup(null);
    }
  };

  const handleSave = async (group: StepGroup, updates: Record<string, unknown>) => {
    try {
      const res = await fetch(`/api/industry/plans/${planId}/steps/batch`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          step_ids: group.stepIds,
          ...updates,
        }),
      });
      if (res.ok) {
        setEditGroup(null);
        onUpdate();
        toast.success(`Updated ${group.stepIds.length} ${group.productName} step(s)`);
      } else {
        toast.error("Failed to update steps");
      }
    } catch (err) {
      console.error("Failed to batch update steps:", err);
      toast.error("Failed to update steps");
    }
  };

  if (steps.length === 0) {
    return (
      <div className="text-center py-4">
        <p className="text-text-muted">No steps in this plan</p>
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div>
        <p className="text-text-secondary text-[13px] mb-2">
          Steps producing the same item are grouped together. Edit a group to configure all instances at once. Expand a group to toggle materials between buy and produce.
        </p>

        <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
          <Table>
            <TableHeader>
              <TableRow className="bg-background-void">
                <TableHead>Product</TableHead>
                <TableHead>Activity</TableHead>
                <TableHead className="text-center">Count</TableHead>
                <TableHead className="text-center">Build / Buy</TableHead>
                <TableHead>ME / TE</TableHead>
                <TableHead>Structure / Rig / Sec</TableHead>
                <TableHead>Station</TableHead>
                <TableHead className="text-center">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {groups.map((group) => {
                const key = groupKey(group);
                const isExpanded = expandedGroups.has(key);
                const materials = groupMaterials[key] || [];
                const isLoadingMats = loadingMaterials.has(key);

                const childTypeIds = new Set<number>();
                for (const step of group.steps) {
                  for (const s of steps) {
                    if (s.parentStepId === step.id) {
                      childTypeIds.add(s.productTypeId);
                    }
                  }
                }
                const buildCount = childTypeIds.size;

                const loadedMats = groupMaterials[key];
                const buyCount = loadedMats
                  ? loadedMats.filter((m) => m.hasBlueprint && !childTypeIds.has(m.typeId)).length
                  : null;

                return [
                  <TableRow
                    key={key}
                    className="bg-background-panel hover:bg-[#1e2235]"
                  >
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <button
                          className="p-0.5 rounded text-text-secondary hover:bg-overlay-subtle"
                          onClick={() => toggleExpand(group)}
                        >
                          {isExpanded ? (
                            <ChevronDown className="h-4 w-4" data-testid="ExpandMoreIcon" />
                          ) : (
                            <ChevronRight className="h-4 w-4" data-testid="ChevronRightIcon" />
                          )}
                        </button>
                        <img
                          src={`https://images.evetech.net/types/${group.productTypeId}/icon?size=32`}
                          alt=""
                          width={24}
                          height={24}
                          className="rounded-sm"
                        />
                        <span className="text-text-emphasis text-sm">
                          {group.productName}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge
                        className={`h-5 text-[11px] cursor-default ${
                          group.activity === "manufacturing"
                            ? "bg-[#1e3a5f] text-blue-science hover:bg-[#1e3a5f]"
                            : "bg-[#3a1e5f] text-[#a78bfa] hover:bg-[#3a1e5f]"
                        }`}
                      >
                        {group.activity}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-center">
                      <Badge className="h-5 text-[11px] bg-[#1e293b] text-text-secondary hover:bg-[#1e293b] cursor-default">
                        {group.stepIds.length}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-center">
                      {togglingAllGroup === key ? (
                        <Loader2 className="h-4 w-4 animate-spin text-text-muted mx-auto" />
                      ) : (
                        <div className="flex items-center justify-center gap-1 flex-wrap">
                          <Badge className={`h-5 text-[11px] cursor-default flex items-center gap-0.5 ${buildCount > 0 ? "bg-[#1e3a5f] text-blue-science" : "bg-[#1e293b] text-text-muted"} hover:bg-[#1e3a5f]`}>
                            <Wrench className="h-3 w-3" data-testid="BuildIcon" />
                            {buildCount}
                          </Badge>
                          <Badge className={`h-5 text-[11px] cursor-default flex items-center gap-0.5 bg-[#1e293b] ${buyCount && buyCount > 0 ? "text-text-secondary" : "text-text-muted"} hover:bg-[#1e293b]`}>
                            <ShoppingCart className="h-3 w-3" data-testid="ShoppingCartIcon" />
                            {buyCount ?? "?"}
                          </Badge>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <button
                                className={`p-0.5 rounded ${buyCount && buyCount > 0 ? "text-teal-success hover:bg-teal-success/10" : "text-[#334155]"}`}
                                onClick={() => handleSetAllBuild(group)}
                                disabled={buyCount === 0}
                              >
                                <Wrench className="h-4 w-4" data-testid="BuildIcon" />
                              </button>
                            </TooltipTrigger>
                            <TooltipContent>Set all materials to Build</TooltipContent>
                          </Tooltip>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <button
                                className={`p-0.5 rounded ${buildCount > 0 ? "text-rose-danger hover:bg-rose-danger/10" : "text-[#334155]"}`}
                                onClick={() => handleSetAllBuy(group)}
                                disabled={buildCount === 0}
                              >
                                <ShoppingCart className="h-4 w-4" data-testid="ShoppingCartIcon" />
                              </button>
                            </TooltipTrigger>
                            <TooltipContent>Set all materials to Buy</TooltipContent>
                          </Tooltip>
                        </div>
                      )}
                    </TableCell>
                    <TableCell className="text-text-secondary text-[13px]">
                      {group.meLevel === "mixed" ? <MixedBadge /> : `ME ${group.meLevel}`}
                      {" / "}
                      {group.teLevel === "mixed" ? <MixedBadge /> : `TE ${group.teLevel}`}
                    </TableCell>
                    <TableCell className="text-text-secondary text-[13px]">
                      {group.structure === "mixed" ? <MixedBadge /> : group.structure}
                      {" / "}
                      {group.rig === "mixed" ? <MixedBadge /> : group.rig}
                      {" / "}
                      {group.security === "mixed" ? <MixedBadge /> : group.security}
                    </TableCell>
                    <TableCell className="text-text-secondary text-[13px]">
                      {group.stationName === "mixed" ? <MixedBadge /> : (group.stationName || "\u2014")}
                      {group.inputLocation === "mixed" ? (
                        <span className="block text-text-muted text-[11px]">
                          In: <MixedBadge />
                        </span>
                      ) : group.inputLocation ? (
                        <span className="block text-text-muted text-[11px]">
                          In: {group.inputLocation}
                        </span>
                      ) : null}
                      {group.outputLocation === "mixed" ? (
                        <span className="block text-text-muted text-[11px]">
                          Out: <MixedBadge />
                        </span>
                      ) : group.outputLocation ? (
                        <span className="block text-text-muted text-[11px]">
                          Out: {group.outputLocation}
                        </span>
                      ) : null}
                    </TableCell>
                    <TableCell className="text-center">
                      <div className="flex justify-center gap-1">
                        {(() => {
                          const hasDetected = group.steps.some((s) => !!detectedLevels[s.blueprintTypeId]);
                          return (
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <button
                                  className={`p-1 rounded ${hasDetected ? "text-teal-success hover:bg-teal-success/10" : "text-[#334155]"}`}
                                  onClick={() => handleApplyBlueprintLevels(group)}
                                  disabled={!hasDetected || applyingBlueprintGroup === groupKey(group)}
                                >
                                  <Sparkles className="h-4 w-4" data-testid="AutoFixHighIcon" />
                                </button>
                              </TooltipTrigger>
                              <TooltipContent>Apply detected ME/TE from blueprints</TooltipContent>
                            </Tooltip>
                          );
                        })()}
                        <button
                          className="p-1 rounded hover:bg-interactive-selected text-primary"
                          onClick={() => setEditGroup(group)}
                        >
                          <Pencil className="h-4 w-4" data-testid="EditIcon" />
                        </button>
                      </div>
                    </TableCell>
                  </TableRow>,
                  // Material rows when expanded
                  ...(isExpanded
                    ? isLoadingMats
                      ? [
                          <TableRow key={`${key}-loading`}>
                            <TableCell
                              colSpan={8}
                              className="text-text-muted text-[13px] pl-10"
                            >
                              Loading materials...
                            </TableCell>
                          </TableRow>,
                        ]
                      : materials.map((mat) => {
                          const matToggleKey = `${key}:${mat.typeId}`;
                          const isToggling = togglingMaterial === matToggleKey;
                          return (
                            <TableRow
                              key={`${key}-mat-${mat.typeId}`}
                              className="bg-background-void hover:bg-[#151825]"
                            >
                              <TableCell>
                                <div className="flex items-center gap-1 pl-10">
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
                                  {mat.produceStatus === "all" ? (
                                    <Badge className="ml-1 h-[18px] text-[10px] bg-[#1e3a5f] text-blue-science hover:bg-[#1e3a5f] cursor-default">
                                      Produce
                                    </Badge>
                                  ) : mat.produceStatus === "mixed" ? (
                                    <Badge className="ml-1 h-[18px] text-[10px] bg-[#422006] text-amber-manufacturing hover:bg-[#422006] cursor-default">
                                      Mixed
                                    </Badge>
                                  ) : (
                                    <Badge variant="outline" className="ml-1 h-[18px] text-[10px] border-[#334155] text-text-muted cursor-default">
                                      Buy
                                    </Badge>
                                  )}
                                </div>
                              </TableCell>
                              <TableCell />
                              <TableCell />
                              <TableCell />
                              <TableCell />
                              <TableCell />
                              <TableCell />
                              <TableCell className="text-center">
                                {mat.hasBlueprint && (
                                  <button
                                    className={`p-1 rounded hover:bg-overlay-subtle ${
                                      mat.produceStatus === "none"
                                        ? "text-teal-success"
                                        : "text-rose-danger"
                                    }`}
                                    disabled={isToggling}
                                    onClick={() => handleToggleProduce(group, mat)}
                                    title={
                                      mat.produceStatus === "none"
                                        ? "Switch all to Produce"
                                        : "Switch all to Buy"
                                    }
                                  >
                                    {mat.produceStatus === "none" ? (
                                      <Wrench className="h-4 w-4" data-testid="BuildIcon" />
                                    ) : (
                                      <ShoppingCart className="h-4 w-4" data-testid="ShoppingCartIcon" />
                                    )}
                                  </button>
                                )}
                              </TableCell>
                            </TableRow>
                          );
                        })
                    : []),
                ];
              })}
            </TableBody>
          </Table>
        </div>

        {editGroup && (
          <BatchEditStepDialog
            group={editGroup}
            open={!!editGroup}
            onClose={() => setEditGroup(null)}
            onSave={(updates) => handleSave(editGroup, updates)}
            detectedLevel={detectedLevels[editGroup.steps[0]?.blueprintTypeId] ?? null}
          />
        )}
      </div>
    </TooltipProvider>
  );
}

// --- Batch Edit Step Dialog ---

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

function BatchEditStepDialog({
  group,
  open,
  onClose,
  onSave,
  detectedLevel,
}: {
  group: StepGroup;
  open: boolean;
  onClose: () => void;
  onSave: (updates: Record<string, unknown>) => void;
  detectedLevel?: BlueprintLevel | null;
}) {
  const firstStep = group.steps[0];

  const [meLevel, setMeLevel] = useState(
    group.meLevel === "mixed" ? 10 : group.meLevel,
  );
  const [teLevel, setTeLevel] = useState(
    group.teLevel === "mixed" ? 20 : group.teLevel,
  );
  const [industrySkill, setIndustrySkill] = useState(firstStep.industrySkill);
  const [advIndustrySkill, setAdvIndustrySkill] = useState(firstStep.advIndustrySkill);
  const [structure, setStructure] = useState(
    group.structure === "mixed" ? "raitaru" : group.structure,
  );
  const [rig, setRig] = useState(
    group.rig === "mixed" ? "t2" : group.rig,
  );
  const [security, setSecurity] = useState(
    group.security === "mixed" ? "high" : group.security,
  );
  const [facilityTax, setFacilityTax] = useState(firstStep.facilityTax);
  const [stationName, setStationName] = useState(firstStep.stationName || "");

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
    if (!open || stationsLoaded) return;
    const fetchStations = async () => {
      try {
        const res = await fetch("/api/stations/user-stations");
        if (res.ok) {
          const data: UserStation[] = await res.json();
          setUserStations(data || []);
          if (group.userStationId !== "mixed" && group.userStationId) {
            const match = (data || []).find((s) => s.id === group.userStationId);
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
  }, [open, stationsLoaded, group]);

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
        }
      } catch (err) {
        console.error("Failed to fetch hangars:", err);
      } finally {
        setHangarsLoaded(true);
      }
    };
    fetchHangars();
  }, [open, stationIdForHangars]);

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
      if (firstStep.rigCategory) {
        const matchingRig = station.rigs.find(
          (r) => r.category === firstStep.rigCategory,
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
    firstStep.activity ? s.activities.includes(firstStep.activity) : true,
  );

  const hasStation = !!selectedUserStation;

  const resolvedSourceLocationId = selectedUserStation
    ? selectedUserStation.stationId
    : null;

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md bg-background-panel border-overlay-medium max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-text-emphasis flex items-center gap-2">
            <img
              src={`https://images.evetech.net/types/${group.productTypeId}/icon?size=32`}
              alt=""
              width={24}
              height={24}
              className="rounded-sm"
            />
            Batch Edit: {group.productName}
          </DialogTitle>
        </DialogHeader>
        <div className="p-3 rounded bg-[rgba(0,140,255,0.08)] border border-[rgba(59,130,246,0.3)] text-[#93c5fd] text-sm mb-2">
          Changes will apply to all {group.stepIds.length} {group.productName} ({group.activity}) step(s).
        </div>

        <div className="flex flex-col gap-3">
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
              <Label htmlFor="batch-me-level" className="text-sm text-text-secondary mb-1 block">ME Level</Label>
              <Input
                id="batch-me-level"
                type="number"
                value={meLevel}
                onChange={(e) => setMeLevel(parseInt(e.target.value) || 0)}
                min={0}
                max={10}
              />
            </div>
            <div>
              <Label htmlFor="batch-te-level" className="text-sm text-text-secondary mb-1 block">TE Level</Label>
              <Input
                id="batch-te-level"
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
          {detectedLevel && (
            <div className="flex items-center gap-2 flex-wrap">
              <Badge variant="outline" className="text-[11px] border-[#0ea5e9] text-[#38bdf8]">
                Blueprint detected: ME {detectedLevel.materialEfficiency} / TE {detectedLevel.timeEfficiency} ({detectedLevel.ownerName}{detectedLevel.isCopy ? ", BPC" : ""})
              </Badge>
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
            </div>
          )}

          {/* Input Location Section */}
          {hasStation && (
            <>
              <Separator className="border-[#1e293b] mt-1" />
              <p className="text-text-secondary text-sm font-semibold">Input Location</p>
              <p className="text-text-muted text-xs">
                Where should materials for these steps be pulled from?
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
                Where should completed items from these jobs be delivered?
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
              })
            }
          >
            Save ({group.stepIds.length} steps)
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

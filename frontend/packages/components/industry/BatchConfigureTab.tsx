import { useState, useEffect } from "react";
import {
  ProductionPlan,
  ProductionPlanStep,
  PlanMaterial,
  UserStation,
  HangarsResponse,
} from "@industry-tool/client/data/models";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import IconButton from "@mui/material/IconButton";
import Chip from "@mui/material/Chip";
import TextField from "@mui/material/TextField";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import Autocomplete from "@mui/material/Autocomplete";
import Snackbar from "@mui/material/Snackbar";
import Alert from "@mui/material/Alert";
import Divider from "@mui/material/Divider";
import Tooltip from "@mui/material/Tooltip";
import CircularProgress from "@mui/material/CircularProgress";
import EditIcon from "@mui/icons-material/Edit";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ChevronRightIcon from "@mui/icons-material/ChevronRight";
import BuildIcon from "@mui/icons-material/Build";
import ShoppingCartIcon from "@mui/icons-material/ShoppingCart";

type Props = {
  plan: ProductionPlan;
  planId: number;
  onUpdate: () => void;
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

export default function BatchConfigureTab({ plan, planId, onUpdate }: Props) {
  const [editGroup, setEditGroup] = useState<StepGroup | null>(null);
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set());
  const [groupMaterials, setGroupMaterials] = useState<Record<string, GroupMaterial[]>>({});
  const [loadingMaterials, setLoadingMaterials] = useState<Set<string>>(new Set());
  const [togglingMaterial, setTogglingMaterial] = useState<string | null>(null);
  const [togglingAllGroup, setTogglingAllGroup] = useState<string | null>(null);
  const [snackbar, setSnackbar] = useState<{
    open: boolean;
    message: string;
    severity: "success" | "error";
  }>({ open: false, message: "", severity: "success" });

  const steps = plan.steps || [];
  const groups = groupSteps(steps);

  const groupKey = (g: StepGroup) => `${g.productTypeId}:${g.activity}`;

  // Compute produce status for a material across all steps in a group
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

  // Eagerly fetch materials for all groups, and refresh statuses when plan data changes
  useEffect(() => {
    for (const group of groups) {
      const key = groupKey(group);
      if (groupMaterials[key]) {
        // Refresh produce statuses for already-loaded materials
        setGroupMaterials((prev) => ({
          ...prev,
          [key]: prev[key].map((mat) => ({
            ...mat,
            produceStatus: computeProduceStatus(group, mat.typeId),
          })),
        }));
      } else if (!loadingMaterials.has(key)) {
        // Eagerly fetch materials for groups that haven't been loaded yet
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
        // Remove: delete child steps for all steps in the group that have one
        const childSteps = steps.filter(
          (s) => group.stepIds.includes(s.parentStepId!) && s.productTypeId === material.typeId,
        );
        for (const child of childSteps) {
          await fetch(`/api/industry/plans/${planId}/steps/${child.id}`, {
            method: "DELETE",
          });
        }
        setSnackbar({
          open: true,
          message: `Set ${material.typeName} to Buy across ${group.stepIds.length} step(s)`,
          severity: "success",
        });
      } else {
        // Create: add child steps for all steps in the group that don't have one
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
        setSnackbar({
          open: true,
          message: `Set ${material.typeName} to Produce across ${group.stepIds.length} step(s)`,
          severity: "success",
        });
      }
      onUpdate();
    } catch (err) {
      console.error("Failed to toggle produce:", err);
      setSnackbar({
        open: true,
        message: "Failed to toggle produce/buy",
        severity: "error",
      });
    } finally {
      setTogglingMaterial(null);
    }
  };

  const handleSetAllBuild = async (group: StepGroup) => {
    const key = groupKey(group);
    setTogglingAllGroup(key);
    try {
      // Ensure materials are loaded
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

      // For each material with a blueprint, create child steps where missing
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
      setSnackbar({
        open: true,
        message: `Set all to Build: created ${created} step(s) across ${group.stepIds.length} ${group.productName} step(s)`,
        severity: "success",
      });
      onUpdate();
    } catch (err) {
      console.error("Failed to set all to build:", err);
      setSnackbar({ open: true, message: "Failed to set all to build", severity: "error" });
    } finally {
      setTogglingAllGroup(null);
    }
  };

  const handleSetAllBuy = async (group: StepGroup) => {
    const key = groupKey(group);
    setTogglingAllGroup(key);
    try {
      // Find all child steps belonging to this group's steps
      const childSteps = steps.filter(
        (s) => s.parentStepId && group.stepIds.includes(s.parentStepId),
      );
      for (const child of childSteps) {
        await fetch(`/api/industry/plans/${planId}/steps/${child.id}`, {
          method: "DELETE",
        });
      }
      setSnackbar({
        open: true,
        message: `Set all to Buy: removed ${childSteps.length} step(s) from ${group.stepIds.length} ${group.productName} step(s)`,
        severity: "success",
      });
      onUpdate();
    } catch (err) {
      console.error("Failed to set all to buy:", err);
      setSnackbar({ open: true, message: "Failed to set all to buy", severity: "error" });
    } finally {
      setTogglingAllGroup(null);
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
        setSnackbar({
          open: true,
          message: `Updated ${group.stepIds.length} ${group.productName} step(s)`,
          severity: "success",
        });
      } else {
        setSnackbar({
          open: true,
          message: "Failed to update steps",
          severity: "error",
        });
      }
    } catch (err) {
      console.error("Failed to batch update steps:", err);
      setSnackbar({
        open: true,
        message: "Failed to update steps",
        severity: "error",
      });
    }
  };

  if (steps.length === 0) {
    return (
      <Box sx={{ textAlign: "center", py: 4 }}>
        <Typography sx={{ color: "#64748b" }}>No steps in this plan</Typography>
      </Box>
    );
  }

  return (
    <Box>
      <Typography sx={{ color: "#94a3b8", fontSize: 13, mb: 2 }}>
        Steps producing the same item are grouped together. Edit a group to configure all instances at once. Expand a group to toggle materials between buy and produce.
      </Typography>

      <TableContainer component={Paper} sx={{ backgroundColor: "#12151f" }}>
        <Table size="small">
          <TableHead>
            <TableRow sx={{ backgroundColor: "#0f1219" }}>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                Product
              </TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                Activity
              </TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="center">
                Count
              </TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="center">
                Build / Buy
              </TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                ME / TE
              </TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                Structure / Rig / Sec
              </TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                Station
              </TableCell>
              <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="center">
                Actions
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {groups.map((group) => {
              const key = groupKey(group);
              const isExpanded = expandedGroups.has(key);
              const materials = groupMaterials[key] || [];
              const isLoadingMats = loadingMaterials.has(key);

              // Count distinct produced material types across all steps in this group
              const childTypeIds = new Set<number>();
              for (const step of group.steps) {
                for (const s of steps) {
                  if (s.parentStepId === step.id) {
                    childTypeIds.add(s.productTypeId);
                  }
                }
              }
              const buildCount = childTypeIds.size;

              // Buy count: from loaded materials, count buildable materials not being produced
              const loadedMats = groupMaterials[key];
              const buyCount = loadedMats
                ? loadedMats.filter((m) => m.hasBlueprint && !childTypeIds.has(m.typeId)).length
                : null; // null means materials not yet loaded

              return [
                <TableRow
                  key={key}
                  sx={{
                    backgroundColor: "#12151f",
                    "&:hover": { backgroundColor: "#1e2235" },
                  }}
                >
                  <TableCell>
                    <Box sx={{ display: "flex", alignItems: "center", gap: 0.5 }}>
                      <IconButton
                        size="small"
                        onClick={() => toggleExpand(group)}
                        sx={{ color: "#94a3b8" }}
                      >
                        {isExpanded ? (
                          <ExpandMoreIcon fontSize="small" />
                        ) : (
                          <ChevronRightIcon fontSize="small" />
                        )}
                      </IconButton>
                      <img
                        src={`https://images.evetech.net/types/${group.productTypeId}/icon?size=32`}
                        alt=""
                        width={24}
                        height={24}
                        style={{ borderRadius: 2 }}
                      />
                      <Typography sx={{ color: "#e2e8f0", fontSize: 14 }}>
                        {group.productName}
                      </Typography>
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={group.activity}
                      size="small"
                      sx={{
                        height: 20,
                        fontSize: 11,
                        backgroundColor:
                          group.activity === "manufacturing" ? "#1e3a5f" : "#3a1e5f",
                        color:
                          group.activity === "manufacturing" ? "#60a5fa" : "#a78bfa",
                      }}
                    />
                  </TableCell>
                  <TableCell align="center">
                    <Chip
                      label={group.stepIds.length}
                      size="small"
                      sx={{
                        height: 20,
                        fontSize: 11,
                        backgroundColor: "#1e293b",
                        color: "#94a3b8",
                      }}
                    />
                  </TableCell>
                  <TableCell align="center">
                    {togglingAllGroup === key ? (
                      <CircularProgress size={16} sx={{ color: "#64748b" }} />
                    ) : (
                      <Box sx={{ display: "flex", alignItems: "center", justifyContent: "center", gap: 0.5, flexWrap: "wrap" }}>
                        <Chip
                          icon={<BuildIcon sx={{ fontSize: "14px !important" }} />}
                          label={buildCount}
                          size="small"
                          sx={{
                            height: 20,
                            fontSize: 11,
                            backgroundColor: buildCount > 0 ? "#1e3a5f" : "#1e293b",
                            color: buildCount > 0 ? "#60a5fa" : "#475569",
                            "& .MuiChip-icon": { color: buildCount > 0 ? "#60a5fa" : "#475569" },
                          }}
                        />
                        <Chip
                          icon={<ShoppingCartIcon sx={{ fontSize: "14px !important" }} />}
                          label={buyCount ?? "?"}
                          size="small"
                          sx={{
                            height: 20,
                            fontSize: 11,
                            backgroundColor: buyCount && buyCount > 0 ? "#1e293b" : "#1e293b",
                            color: buyCount && buyCount > 0 ? "#94a3b8" : "#475569",
                            "& .MuiChip-icon": { color: buyCount && buyCount > 0 ? "#94a3b8" : "#475569" },
                          }}
                        />
                        <Tooltip title="Set all materials to Build">
                          <span>
                            <IconButton
                              size="small"
                              onClick={() => handleSetAllBuild(group)}
                              disabled={buyCount === 0}
                              sx={{ color: buyCount && buyCount > 0 ? "#10b981" : "#334155", p: 0.25 }}
                            >
                              <BuildIcon sx={{ fontSize: 16 }} />
                            </IconButton>
                          </span>
                        </Tooltip>
                        <Tooltip title="Set all materials to Buy">
                          <span>
                            <IconButton
                              size="small"
                              onClick={() => handleSetAllBuy(group)}
                              disabled={buildCount === 0}
                              sx={{ color: buildCount > 0 ? "#ef4444" : "#334155", p: 0.25 }}
                            >
                              <ShoppingCartIcon sx={{ fontSize: 16 }} />
                            </IconButton>
                          </span>
                        </Tooltip>
                      </Box>
                    )}
                  </TableCell>
                  <TableCell sx={{ color: "#94a3b8", fontSize: 13 }}>
                    {group.meLevel === "mixed" ? (
                      <Chip label="Mixed" size="small" sx={{ height: 18, fontSize: 10, backgroundColor: "#422006", color: "#f59e0b" }} />
                    ) : (
                      `ME ${group.meLevel}`
                    )}
                    {" / "}
                    {group.teLevel === "mixed" ? (
                      <Chip label="Mixed" size="small" sx={{ height: 18, fontSize: 10, backgroundColor: "#422006", color: "#f59e0b" }} />
                    ) : (
                      `TE ${group.teLevel}`
                    )}
                  </TableCell>
                  <TableCell sx={{ color: "#94a3b8", fontSize: 13 }}>
                    {group.structure === "mixed" ? (
                      <Chip label="Mixed" size="small" sx={{ height: 18, fontSize: 10, backgroundColor: "#422006", color: "#f59e0b" }} />
                    ) : (
                      group.structure
                    )}
                    {" / "}
                    {group.rig === "mixed" ? (
                      <Chip label="Mixed" size="small" sx={{ height: 18, fontSize: 10, backgroundColor: "#422006", color: "#f59e0b" }} />
                    ) : (
                      group.rig
                    )}
                    {" / "}
                    {group.security === "mixed" ? (
                      <Chip label="Mixed" size="small" sx={{ height: 18, fontSize: 10, backgroundColor: "#422006", color: "#f59e0b" }} />
                    ) : (
                      group.security
                    )}
                  </TableCell>
                  <TableCell sx={{ color: "#94a3b8", fontSize: 13 }}>
                    {group.stationName === "mixed" ? (
                      <Chip label="Mixed" size="small" sx={{ height: 18, fontSize: 10, backgroundColor: "#422006", color: "#f59e0b" }} />
                    ) : (
                      group.stationName || "—"
                    )}
                    {group.inputLocation === "mixed" ? (
                      <Typography component="span" sx={{ color: "#64748b", fontSize: 11, display: "block" }}>
                        In: <Chip label="Mixed" size="small" sx={{ height: 16, fontSize: 9, backgroundColor: "#422006", color: "#f59e0b" }} />
                      </Typography>
                    ) : group.inputLocation ? (
                      <Typography sx={{ color: "#64748b", fontSize: 11 }}>
                        In: {group.inputLocation}
                      </Typography>
                    ) : null}
                    {group.outputLocation === "mixed" ? (
                      <Typography component="span" sx={{ color: "#64748b", fontSize: 11, display: "block" }}>
                        Out: <Chip label="Mixed" size="small" sx={{ height: 16, fontSize: 9, backgroundColor: "#422006", color: "#f59e0b" }} />
                      </Typography>
                    ) : group.outputLocation ? (
                      <Typography sx={{ color: "#64748b", fontSize: 11 }}>
                        Out: {group.outputLocation}
                      </Typography>
                    ) : null}
                  </TableCell>
                  <TableCell align="center">
                    <IconButton
                      size="small"
                      onClick={() => setEditGroup(group)}
                      sx={{ color: "#3b82f6" }}
                    >
                      <EditIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>,
                // Material rows when expanded
                ...(isExpanded
                  ? isLoadingMats
                    ? [
                        <TableRow key={`${key}-loading`}>
                          <TableCell
                            colSpan={8}
                            sx={{ color: "#64748b", fontSize: 13, pl: 6 }}
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
                            sx={{
                              backgroundColor: "#0f1219",
                              "&:hover": { backgroundColor: "#151825" },
                            }}
                          >
                            <TableCell>
                              <Box
                                sx={{
                                  display: "flex",
                                  alignItems: "center",
                                  gap: 0.5,
                                  pl: 5,
                                }}
                              >
                                <img
                                  src={`https://images.evetech.net/types/${mat.typeId}/icon?size=32`}
                                  alt=""
                                  width={18}
                                  height={18}
                                  style={{ borderRadius: 2 }}
                                />
                                <Typography sx={{ color: "#cbd5e1", fontSize: 13 }}>
                                  {mat.typeName}
                                </Typography>
                                <Typography
                                  sx={{ color: "#64748b", fontSize: 12, ml: 1 }}
                                >
                                  x{mat.quantity}
                                </Typography>
                                {mat.produceStatus === "all" ? (
                                  <Chip
                                    label="Produce"
                                    size="small"
                                    sx={{
                                      ml: 1,
                                      height: 18,
                                      fontSize: 10,
                                      backgroundColor: "#1e3a5f",
                                      color: "#60a5fa",
                                    }}
                                  />
                                ) : mat.produceStatus === "mixed" ? (
                                  <Chip
                                    label="Mixed"
                                    size="small"
                                    sx={{
                                      ml: 1,
                                      height: 18,
                                      fontSize: 10,
                                      backgroundColor: "#422006",
                                      color: "#f59e0b",
                                    }}
                                  />
                                ) : (
                                  <Chip
                                    label="Buy"
                                    size="small"
                                    variant="outlined"
                                    sx={{
                                      ml: 1,
                                      height: 18,
                                      fontSize: 10,
                                      borderColor: "#334155",
                                      color: "#64748b",
                                    }}
                                  />
                                )}
                              </Box>
                            </TableCell>
                            <TableCell />
                            <TableCell />
                            <TableCell />
                            <TableCell />
                            <TableCell />
                            <TableCell />
                            <TableCell align="center">
                              {mat.hasBlueprint && (
                                <IconButton
                                  size="small"
                                  disabled={isToggling}
                                  onClick={() => handleToggleProduce(group, mat)}
                                  sx={{
                                    color:
                                      mat.produceStatus === "none"
                                        ? "#10b981"
                                        : "#ef4444",
                                  }}
                                  title={
                                    mat.produceStatus === "none"
                                      ? "Switch all to Produce"
                                      : "Switch all to Buy"
                                  }
                                >
                                  {mat.produceStatus === "none" ? (
                                    <BuildIcon fontSize="small" />
                                  ) : (
                                    <ShoppingCartIcon fontSize="small" />
                                  )}
                                </IconButton>
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
      </TableContainer>

      {editGroup && (
        <BatchEditStepDialog
          group={editGroup}
          open={!!editGroup}
          onClose={() => setEditGroup(null)}
          onSave={(updates) => handleSave(editGroup, updates)}
        />
      )}

      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
      >
        <Alert severity={snackbar.severity} sx={{ width: "100%" }}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
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
}: {
  group: StepGroup;
  open: boolean;
  onClose: () => void;
  onSave: (updates: Record<string, unknown>) => void;
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

  // Input location state
  const [hangarsData, setHangarsData] = useState<HangarsResponse | null>(null);
  const [hangarsLoaded, setHangarsLoaded] = useState(false);
  const [selectedOwner, setSelectedOwner] = useState<OwnerOption | null>(null);
  const [selectedDivision, setSelectedDivision] = useState<DivisionOption | null>(null);
  const [selectedContainer, setSelectedContainer] = useState<ContainerOption | null>(null);

  // Output location state
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
          // Pre-select if first step references a station and all steps share the same one
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

  // Fetch hangars when station is selected
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

  // Reset loaded state when dialog closes
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

  const handleStationSelect = (station: UserStation | null) => {
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

  // Build owner options from hangars data
  const ownerOptions: OwnerOption[] = [];
  if (hangarsData) {
    for (const char of hangarsData.characters) {
      ownerOptions.push({ id: char.id, name: char.name, type: "character" });
    }
    for (const corp of hangarsData.corporations) {
      ownerOptions.push({ id: corp.id, name: corp.name, type: "corporation" });
    }
  }

  // Build division options for selected corporation
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

  // Build container options filtered by selected owner/division
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

  const handleOwnerSelect = (owner: OwnerOption | null) => {
    setSelectedOwner(owner);
    setSelectedDivision(null);
    setSelectedContainer(null);
  };

  const handleDivisionSelect = (division: DivisionOption | null) => {
    setSelectedDivision(division);
    setSelectedContainer(null);
  };

  // Build output division options
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

  // Build output container options
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

  const handleOutputOwnerSelect = (owner: OwnerOption | null) => {
    setSelectedOutputOwner(owner);
    setSelectedOutputDivision(null);
    setSelectedOutputContainer(null);
  };

  const handleOutputDivisionSelect = (division: DivisionOption | null) => {
    setSelectedOutputDivision(division);
    setSelectedOutputContainer(null);
  };

  // Filter stations to show those with matching activity
  const filteredStations = userStations.filter((s) =>
    firstStep.activity ? s.activities.includes(firstStep.activity) : true,
  );

  const hasStation = !!selectedUserStation;

  const resolvedSourceLocationId = selectedUserStation
    ? selectedUserStation.stationId
    : null;

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: { backgroundColor: "#12151f", color: "#e2e8f0" },
      }}
    >
      <DialogTitle>
        <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
          <img
            src={`https://images.evetech.net/types/${group.productTypeId}/icon?size=32`}
            alt=""
            width={24}
            height={24}
            style={{ borderRadius: 2 }}
          />
          Batch Edit: {group.productName}
        </Box>
      </DialogTitle>
      <DialogContent>
        <Alert severity="info" sx={{ mb: 2, mt: 1 }}>
          Changes will apply to all {group.stepIds.length} {group.productName} ({group.activity}) step(s).
        </Alert>

        <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
          {/* Preferred Station Selection */}
          <Autocomplete
            value={selectedUserStation}
            onChange={(_, newValue) => handleStationSelect(newValue)}
            options={filteredStations}
            getOptionLabel={(option) =>
              `${option.stationName || "Unknown"} (${option.solarSystemName || ""})`
            }
            isOptionEqualToValue={(option, value) => option.id === value.id}
            renderOption={(props, option) => (
              <Box component="li" {...props}>
                <Box>
                  <Typography variant="body2">
                    {option.stationName || "Unknown"}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {option.solarSystemName} &middot; {option.structure} &middot;{" "}
                    {option.activities.join(", ")}
                  </Typography>
                </Box>
              </Box>
            )}
            renderInput={(params) => (
              <TextField
                {...params}
                label="Preferred Station"
                placeholder="Select a saved station or leave empty for manual config"
              />
            )}
          />

          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: 2,
            }}
          >
            <TextField
              type="number"
              label="ME Level"
              value={meLevel}
              onChange={(e) => setMeLevel(parseInt(e.target.value) || 0)}
              inputProps={{ min: 0, max: 10 }}
            />
            <TextField
              type="number"
              label="TE Level"
              value={teLevel}
              onChange={(e) => setTeLevel(parseInt(e.target.value) || 0)}
              inputProps={{ min: 0, max: 20 }}
            />
            <TextField
              type="number"
              label="Industry Skill"
              value={industrySkill}
              onChange={(e) => setIndustrySkill(parseInt(e.target.value) || 0)}
              inputProps={{ min: 0, max: 5 }}
            />
            <TextField
              type="number"
              label="Adv. Industry Skill"
              value={advIndustrySkill}
              onChange={(e) =>
                setAdvIndustrySkill(parseInt(e.target.value) || 0)
              }
              inputProps={{ min: 0, max: 5 }}
            />
            <FormControl fullWidth disabled={hasStation}>
              <InputLabel>Structure</InputLabel>
              <Select
                value={structure}
                label="Structure"
                onChange={(e) => setStructure(e.target.value)}
              >
                <MenuItem value="station">Station</MenuItem>
                <MenuItem value="raitaru">Raitaru</MenuItem>
                <MenuItem value="azbel">Azbel</MenuItem>
                <MenuItem value="sotiyo">Sotiyo</MenuItem>
                <MenuItem value="athanor">Athanor</MenuItem>
                <MenuItem value="tatara">Tatara</MenuItem>
              </Select>
            </FormControl>
            <FormControl fullWidth disabled={hasStation}>
              <InputLabel>Rig</InputLabel>
              <Select
                value={rig}
                label="Rig"
                onChange={(e) => setRig(e.target.value)}
              >
                <MenuItem value="none">None</MenuItem>
                <MenuItem value="t1">T1</MenuItem>
                <MenuItem value="t2">T2</MenuItem>
              </Select>
            </FormControl>
            <FormControl fullWidth disabled={hasStation}>
              <InputLabel>Security</InputLabel>
              <Select
                value={security}
                label="Security"
                onChange={(e) => setSecurity(e.target.value)}
              >
                <MenuItem value="high">Highsec</MenuItem>
                <MenuItem value="low">Lowsec</MenuItem>
                <MenuItem value="null">Nullsec / WH</MenuItem>
              </Select>
            </FormControl>
            <TextField
              type="number"
              label="Facility Tax %"
              value={facilityTax}
              onChange={(e) => setFacilityTax(parseFloat(e.target.value) || 0)}
              inputProps={{ min: 0, step: 0.1 }}
              disabled={hasStation}
            />
            <TextField
              label="Station Name"
              value={stationName}
              onChange={(e) => setStationName(e.target.value)}
              placeholder="e.g. Jita 4-4 or player structure name"
              sx={{ gridColumn: "1 / -1" }}
              disabled={hasStation}
            />
          </Box>

          {/* Input Location Section */}
          {hasStation && (
            <>
              <Divider sx={{ borderColor: "#1e293b", mt: 1 }} />
              <Typography sx={{ color: "#94a3b8", fontSize: 14, fontWeight: 600 }}>
                Input Location
              </Typography>
              <Typography sx={{ color: "#64748b", fontSize: 12 }}>
                Where should materials for these steps be pulled from?
              </Typography>

              {!hangarsLoaded ? (
                <Typography sx={{ color: "#64748b", fontSize: 13 }}>
                  Loading hangars...
                </Typography>
              ) : (
                <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
                  <Autocomplete
                    value={selectedOwner}
                    onChange={(_, newValue) => handleOwnerSelect(newValue)}
                    options={ownerOptions}
                    getOptionLabel={(option) =>
                      `${option.name} (${option.type === "character" ? "Character" : "Corporation"})`
                    }
                    isOptionEqualToValue={(option, value) =>
                      option.id === value.id && option.type === value.type
                    }
                    renderInput={(params) => (
                      <TextField
                        {...params}
                        label="Owner"
                        placeholder="Select character or corporation"
                        size="small"
                      />
                    )}
                  />

                  {selectedOwner?.type === "corporation" && divisionOptions.length > 0 && (
                    <Autocomplete
                      value={selectedDivision}
                      onChange={(_, newValue) => handleDivisionSelect(newValue)}
                      options={divisionOptions}
                      getOptionLabel={(option) => `${option.number}. ${option.name}`}
                      isOptionEqualToValue={(option, value) =>
                        option.number === value.number
                      }
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label="Hangar Division"
                          placeholder="Select division"
                          size="small"
                        />
                      )}
                    />
                  )}

                  {selectedOwner && (
                    <Autocomplete
                      value={selectedContainer}
                      onChange={(_, newValue) => setSelectedContainer(newValue)}
                      options={containerOptions}
                      getOptionLabel={(option) => option.name}
                      isOptionEqualToValue={(option, value) =>
                        option.id === value.id
                      }
                      noOptionsText="No containers at this station"
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label="Container (optional)"
                          placeholder="Select container or leave empty for hangar"
                          size="small"
                        />
                      )}
                    />
                  )}
                </Box>
              )}

              <Divider sx={{ borderColor: "#1e293b", mt: 1 }} />
              <Typography sx={{ color: "#94a3b8", fontSize: 14, fontWeight: 600 }}>
                Output Location
              </Typography>
              <Typography sx={{ color: "#64748b", fontSize: 12 }}>
                Where should completed items from these jobs be delivered?
              </Typography>

              {!hangarsLoaded ? (
                <Typography sx={{ color: "#64748b", fontSize: 13 }}>
                  Loading hangars...
                </Typography>
              ) : (
                <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
                  <Autocomplete
                    value={selectedOutputOwner}
                    onChange={(_, newValue) => handleOutputOwnerSelect(newValue)}
                    options={ownerOptions}
                    getOptionLabel={(option) =>
                      `${option.name} (${option.type === "character" ? "Character" : "Corporation"})`
                    }
                    isOptionEqualToValue={(option, value) =>
                      option.id === value.id && option.type === value.type
                    }
                    renderInput={(params) => (
                      <TextField
                        {...params}
                        label="Owner"
                        placeholder="Select character or corporation"
                        size="small"
                      />
                    )}
                  />

                  {selectedOutputOwner?.type === "corporation" && outputDivisionOptions.length > 0 && (
                    <Autocomplete
                      value={selectedOutputDivision}
                      onChange={(_, newValue) => handleOutputDivisionSelect(newValue)}
                      options={outputDivisionOptions}
                      getOptionLabel={(option) => `${option.number}. ${option.name}`}
                      isOptionEqualToValue={(option, value) =>
                        option.number === value.number
                      }
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label="Hangar Division"
                          placeholder="Select division"
                          size="small"
                        />
                      )}
                    />
                  )}

                  {selectedOutputOwner && (
                    <Autocomplete
                      value={selectedOutputContainer}
                      onChange={(_, newValue) => setSelectedOutputContainer(newValue)}
                      options={outputContainerOptions}
                      getOptionLabel={(option) => option.name}
                      isOptionEqualToValue={(option, value) =>
                        option.id === value.id
                      }
                      noOptionsText="No containers at this station"
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label="Container (optional)"
                          placeholder="Select container or leave empty for hangar"
                          size="small"
                        />
                      )}
                    />
                  )}
                </Box>
              )}
            </>
          )}
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} sx={{ color: "#94a3b8" }}>
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
          variant="contained"
          sx={{
            backgroundColor: "#3b82f6",
            "&:hover": { backgroundColor: "#2563eb" },
          }}
        >
          Save ({group.stepIds.length} steps)
        </Button>
      </DialogActions>
    </Dialog>
  );
}

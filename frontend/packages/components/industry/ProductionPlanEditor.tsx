import { useState, useEffect, useCallback, useRef } from "react";
import {
  ProductionPlan,
  ProductionPlanStep,
  PlanMaterial,
  GenerateJobsResult,
  UserStation,
  HangarsResponse,
} from "@industry-tool/client/data/models";
import { formatISK } from "@industry-tool/utils/formatting";
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
import Tabs from "@mui/material/Tabs";
import Tab from "@mui/material/Tab";
import Snackbar from "@mui/material/Snackbar";
import Alert from "@mui/material/Alert";
import Divider from "@mui/material/Divider";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import ChevronRightIcon from "@mui/icons-material/ChevronRight";
import BuildIcon from "@mui/icons-material/Build";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import ShoppingCartIcon from "@mui/icons-material/ShoppingCart";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import CheckIcon from "@mui/icons-material/Check";
import CloseIcon from "@mui/icons-material/Close";
import SwapHorizIcon from "@mui/icons-material/SwapHoriz";
import Tooltip from "@mui/material/Tooltip";
import BatchConfigureTab from "./BatchConfigureTab";

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
  const [snackbar, setSnackbar] = useState<{
    open: boolean;
    message: string;
    severity: "success" | "error";
  }>({ open: false, message: "", severity: "success" });
  const [editingName, setEditingName] = useState(false);
  const [nameValue, setNameValue] = useState("");
  const [tab, setTab] = useState(0);

  const initialLoadRef = useRef(true);

  const fetchPlan = useCallback(async () => {
    if (initialLoadRef.current) {
      setLoading(true);
    }
    try {
      const res = await fetch(`/api/industry/plans/${planId}`);
      if (res.ok) {
        const data = await res.json();
        setPlan(data);
        // Only auto-expand root step on initial load
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
  }, [planId]);

  useEffect(() => {
    fetchPlan();
  }, [fetchPlan]);

  const fetchMaterials = async (stepId: number) => {
    setLoadingMaterials((prev) => new Set([...prev, stepId]));
    try {
      const res = await fetch(
        `/api/industry/plans/${planId}/steps/${stepId}/materials`,
      );
      if (res.ok) {
        const data = await res.json();
        setStepMaterials((prev) => ({ ...prev, [stepId]: data || [] }));
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
      // Find the child step and delete it
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
          // Refresh materials for parent
          fetchMaterials(parentStepId);
        } catch (err) {
          console.error("Failed to remove step:", err);
        }
      }
    } else {
      // Create a new step for this material
      try {
        const res = await fetch(`/api/industry/plans/${planId}/steps`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            parent_step_id: parentStepId,
            product_type_id: material.typeId,
          }),
        });
        if (res.ok) {
          const newStep = await res.json();
          await fetchPlan();
          fetchMaterials(parentStepId);
          // Auto-expand the new child step and load its materials
          if (newStep?.id) {
            setExpandedSteps((prev) => new Set([...prev, newStep.id]));
            fetchMaterials(newStep.id);
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
        setSnackbar({
          open: true,
          message: "Step updated",
          severity: "success",
        });
      }
    } catch (err) {
      console.error("Failed to update step:", err);
    }
  };

  const handleGenerate = async () => {
    setGenerating(true);
    try {
      const res = await fetch(`/api/industry/plans/${planId}/generate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ quantity: generateQuantity }),
      });
      if (res.ok) {
        const result: GenerateJobsResult = await res.json();
        setGenerateResult(result);
        setSnackbar({
          open: true,
          message: `Created ${result.created.length} job(s), skipped ${result.skipped.length}`,
          severity: "success",
        });
      }
    } catch (err) {
      console.error("Failed to generate jobs:", err);
      setSnackbar({
        open: true,
        message: "Failed to generate jobs",
        severity: "error",
      });
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

  if (loading || !plan) {
    return (
      <Box sx={{ textAlign: "center", py: 4 }}>
        <Typography sx={{ color: "#64748b" }}>Loading plan...</Typography>
      </Box>
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

  const depthColors = ["#3b82f6", "#10b981", "#f59e0b", "#a78bfa", "#ec4899", "#06b6d4"];

  const renderDepthIndicators = (colorPath: number[]) =>
    colorPath.map((colorIndex, i) => (
      <Box
        key={`depth-bar-${i}`}
        sx={{
          position: "absolute",
          left: `${i * 32 + 8}px`,
          top: 0,
          bottom: 0,
          width: 3,
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
        sx={{
          backgroundColor: depth === 0 ? "#1a1d2e" : "#12151f",
          "&:hover": { backgroundColor: "#1e2235" },
        }}
      >
        <TableCell sx={{ position: "relative", pl: 0 }}>
          {renderDepthIndicators(colorPath)}
          <Box sx={{ display: "flex", alignItems: "center", gap: 0.5, pl: `${depth * 32 + 20}px` }}>
            <IconButton
              size="small"
              onClick={() => toggleExpand(step.id)}
              sx={{ color: "#94a3b8" }}
            >
              {isExpanded ? (
                <ExpandMoreIcon fontSize="small" />
              ) : (
                <ChevronRightIcon fontSize="small" />
              )}
            </IconButton>
            <img
              src={`https://images.evetech.net/types/${step.productTypeId}/icon?size=32`}
              alt=""
              width={20}
              height={20}
              style={{ borderRadius: 2 }}
            />
            <Typography sx={{ color: "#e2e8f0", fontSize: 14, fontWeight: depth === 0 ? 600 : 400 }}>
              {step.productName || `Type ${step.productTypeId}`}
            </Typography>
            <Chip
              label={step.activity}
              size="small"
              sx={{
                ml: 1,
                height: 20,
                fontSize: 11,
                backgroundColor:
                  step.activity === "manufacturing" ? "#1e3a5f" : "#3a1e5f",
                color:
                  step.activity === "manufacturing" ? "#60a5fa" : "#a78bfa",
              }}
            />
          </Box>
        </TableCell>
        <TableCell sx={{ color: "#94a3b8", fontSize: 13 }}>
          ME {step.meLevel} / TE {step.teLevel}
        </TableCell>
        <TableCell sx={{ color: "#94a3b8", fontSize: 13 }}>
          {step.structure} / {step.rig} / {step.security}
        </TableCell>
        <TableCell sx={{ color: "#94a3b8", fontSize: 13 }}>
          {step.stationName || "—"}
          {step.sourceOwnerName && (
            <Typography sx={{ color: "#64748b", fontSize: 11 }}>
              In: {step.sourceOwnerName}
              {step.sourceDivisionName ? ` / ${step.sourceDivisionName}` : ""}
              {step.sourceContainerName ? ` / ${step.sourceContainerName}` : ""}
            </Typography>
          )}
          {step.outputOwnerName ? (
            <Typography sx={{ color: "#64748b", fontSize: 11 }}>
              Out: {step.outputOwnerName}
              {step.outputDivisionName ? ` / ${step.outputDivisionName}` : ""}
              {step.outputContainerName ? ` / ${step.outputContainerName}` : ""}
            </Typography>
          ) : !step.parentStepId ? (
            <Typography sx={{ color: "#475569", fontSize: 11, fontStyle: "italic" }}>
              Out: set at build time
            </Typography>
          ) : null}
          {step.parentStepId && parentStep &&
           step.userStationId && parentStep.userStationId &&
           step.userStationId !== parentStep.userStationId && (
            <Tooltip title="Items must be moved between stations">
              <Box sx={{ display: "flex", alignItems: "center", gap: 0.5, mt: 0.25 }}>
                <SwapHorizIcon sx={{ fontSize: 14, color: "#f59e0b" }} />
                <Typography sx={{ color: "#f59e0b", fontSize: 11 }}>Transfer</Typography>
              </Box>
            </Tooltip>
          )}
        </TableCell>
        <TableCell align="center">
          <IconButton
            size="small"
            onClick={() => setEditStepId(step.id)}
            sx={{ color: "#3b82f6" }}
          >
            <EditIcon fontSize="small" />
          </IconButton>
          {depth > 0 && (
            <IconButton
              size="small"
              onClick={async () => {
                await fetch(
                  `/api/industry/plans/${planId}/steps/${step.id}`,
                  { method: "DELETE" },
                );
                fetchPlan();
                if (step.parentStepId) fetchMaterials(step.parentStepId);
              }}
              sx={{ color: "#ef4444" }}
            >
              <DeleteIcon fontSize="small" />
            </IconButton>
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
              sx={{ position: "relative", pl: 0, color: "#64748b", fontSize: 13 }}
            >
              {renderDepthIndicators(colorPath)}
              <Box sx={{ pl: `${(depth + 1) * 32 + 20}px` }}>Loading materials...</Box>
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
          // Check if this material has a child step (is produced)
          const childStep = children.find(
            (c) => c.productTypeId === mat.typeId,
          );

          if (childStep) {
            // Render the child step recursively
            rows.push(...renderStepRow(childStep, depth + 1, [...colorPath, childIndex], step));
            childIndex++;
          } else {
            // Render as a buy material
            rows.push(
              <TableRow
                key={`mat-${step.id}-${mat.typeId}`}
                sx={{
                  backgroundColor: "#0f1219",
                  "&:hover": { backgroundColor: "#151825" },
                }}
              >
                <TableCell sx={{ position: "relative", pl: 0 }}>
                  {renderDepthIndicators(colorPath)}
                  <Box
                    sx={{
                      display: "flex",
                      alignItems: "center",
                      gap: 0.5,
                      pl: `${(depth + 1) * 32 + 20}px`,
                    }}
                  >
                    <img
                      src={`https://images.evetech.net/types/${mat.typeId}/icon?size=32`}
                      alt=""
                      width={18}
                      height={18}
                      style={{ borderRadius: 2 }}
                    />
                    <Typography
                      sx={{ color: "#cbd5e1", fontSize: 13 }}
                    >
                      {mat.typeName}
                    </Typography>
                    <Typography
                      sx={{ color: "#64748b", fontSize: 12, ml: 1 }}
                    >
                      x{mat.quantity}
                    </Typography>
                    {mat.isProduced ? (
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
                <TableCell align="center">
                  {mat.hasBlueprint && (
                    <IconButton
                      size="small"
                      onClick={() => handleToggleProduce(step.id, mat)}
                      sx={{
                        color: mat.isProduced ? "#ef4444" : "#10b981",
                      }}
                      title={
                        mat.isProduced
                          ? "Switch to Buy"
                          : "Switch to Produce"
                      }
                    >
                      {mat.isProduced ? (
                        <ShoppingCartIcon fontSize="small" />
                      ) : (
                        <BuildIcon fontSize="small" />
                      )}
                    </IconButton>
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
    <Box>
      <Box
        sx={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          mb: 2,
        }}
      >
        <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
          <img
            src={`https://images.evetech.net/types/${plan.productTypeId}/icon?size=64`}
            alt=""
            width={40}
            height={40}
            style={{ borderRadius: 4 }}
          />
          <Box>
            {editingName ? (
              <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                <TextField
                  size="small"
                  value={nameValue}
                  onChange={(e) => setNameValue(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") handleSaveName();
                    if (e.key === "Escape") setEditingName(false);
                  }}
                  autoFocus
                  sx={{ minWidth: 250 }}
                />
                <IconButton size="small" onClick={handleSaveName} sx={{ color: "#10b981" }}>
                  <CheckIcon fontSize="small" />
                </IconButton>
                <IconButton size="small" onClick={() => setEditingName(false)} sx={{ color: "#94a3b8" }}>
                  <CloseIcon fontSize="small" />
                </IconButton>
              </Box>
            ) : (
              <Box sx={{ display: "flex", alignItems: "center", gap: 0.5 }}>
                <Typography
                  variant="h5"
                  sx={{ color: "#e2e8f0", fontWeight: 600 }}
                >
                  {plan.name}
                </Typography>
                <IconButton
                  size="small"
                  onClick={() => {
                    setNameValue(plan.name);
                    setEditingName(true);
                  }}
                  sx={{ color: "#64748b", "&:hover": { color: "#3b82f6" } }}
                >
                  <EditIcon fontSize="small" />
                </IconButton>
              </Box>
            )}
            <Typography sx={{ color: "#64748b", fontSize: 13 }}>
              {plan.steps?.length || 0} production step(s)
            </Typography>
          </Box>
        </Box>
        <Button
          variant="contained"
          startIcon={<PlayArrowIcon />}
          onClick={() => setGenerateDialogOpen(true)}
          sx={{
            backgroundColor: "#10b981",
            "&:hover": { backgroundColor: "#059669" },
          }}
        >
          Generate Jobs
        </Button>
      </Box>

      <Box sx={{ borderBottom: 1, borderColor: "rgba(148, 163, 184, 0.15)", mb: 2 }}>
        <Tabs
          value={tab}
          onChange={(_, newValue) => setTab(newValue)}
          sx={{
            "& .MuiTab-root": { color: "#64748b", textTransform: "none", fontWeight: 500 },
            "& .Mui-selected": { color: "#3b82f6" },
            "& .MuiTabs-indicator": { backgroundColor: "#3b82f6" },
          }}
        >
          <Tab label="Step Tree" />
          <Tab label="Batch Configure" />
        </Tabs>
      </Box>

      {tab === 0 && (
        <TableContainer component={Paper} sx={{ backgroundColor: "#12151f" }}>
          <Table size="small">
            <TableHead>
              <TableRow sx={{ backgroundColor: "#0f1219" }}>
                <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                  Item / Material
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
                <TableCell
                  sx={{ color: "#94a3b8", fontWeight: 600 }}
                  align="center"
                >
                  Actions
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {rootStep ? (
                renderStepRow(rootStep, 0, [0])
              ) : (
                <TableRow>
                  <TableCell colSpan={5} sx={{ textAlign: "center" }}>
                    <Typography sx={{ color: "#64748b" }}>
                      No steps in this plan
                    </Typography>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {tab === 1 && (
        <BatchConfigureTab plan={plan} planId={planId} onUpdate={fetchPlan} />
      )}

      {/* Edit Step Dialog */}
      {editStepId && (
        <EditStepDialog
          step={
            plan.steps?.find((s) => s.id === editStepId) || null
          }
          open={!!editStepId}
          onClose={() => setEditStepId(null)}
          onSave={(updates) => handleUpdateStep(editStepId, updates)}
        />
      )}

      {/* Generate Jobs Dialog */}
      <Dialog
        open={generateDialogOpen}
        onClose={() => {
          setGenerateDialogOpen(false);
          setGenerateResult(null);
        }}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: { backgroundColor: "#12151f", color: "#e2e8f0" },
        }}
      >
        <DialogTitle>Generate Production Jobs</DialogTitle>
        <DialogContent>
          {generateResult ? (
            <Box>
              <Typography sx={{ color: "#10b981", mb: 1 }}>
                Created {generateResult.created.length} job(s)
              </Typography>
              {generateResult.created.map((job) => (
                <Typography
                  key={job.id}
                  sx={{ color: "#cbd5e1", fontSize: 13, ml: 2 }}
                >
                  {job.blueprintName || `BP ${job.blueprintTypeId}`} &mdash;{" "}
                  {job.runs} runs
                  {job.estimatedCost
                    ? ` (${formatISK(job.estimatedCost)})`
                    : ""}
                </Typography>
              ))}
              {generateResult.skipped.length > 0 && (
                <>
                  <Typography sx={{ color: "#f59e0b", mt: 2, mb: 1 }}>
                    Skipped {generateResult.skipped.length} item(s)
                  </Typography>
                  {generateResult.skipped.map((skip, i) => (
                    <Typography
                      key={i}
                      sx={{ color: "#94a3b8", fontSize: 13, ml: 2 }}
                    >
                      {skip.typeName} &mdash; {skip.reason}
                    </Typography>
                  ))}
                </>
              )}
            </Box>
          ) : (
            <Box sx={{ mt: 1 }}>
              <Typography sx={{ color: "#94a3b8", mb: 2 }}>
                How many {plan.productName || "units"} do you want to produce?
                Job queue entries will be created for each step in the
                production chain.
              </Typography>
              <TextField
                type="number"
                label="Quantity"
                value={generateQuantity}
                onChange={(e) =>
                  setGenerateQuantity(Math.max(1, parseInt(e.target.value) || 1))
                }
                fullWidth
                inputProps={{ min: 1 }}
              />
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          {generateResult ? (
            <Button
              onClick={() => {
                setGenerateDialogOpen(false);
                setGenerateResult(null);
              }}
              variant="contained"
              sx={{
                backgroundColor: "#3b82f6",
                "&:hover": { backgroundColor: "#2563eb" },
              }}
            >
              Done
            </Button>
          ) : (
            <>
              <Button
                onClick={() => setGenerateDialogOpen(false)}
                sx={{ color: "#94a3b8" }}
              >
                Cancel
              </Button>
              <Button
                onClick={handleGenerate}
                disabled={generating}
                variant="contained"
                sx={{
                  backgroundColor: "#10b981",
                  "&:hover": { backgroundColor: "#059669" },
                }}
              >
                {generating ? "Generating..." : "Generate"}
              </Button>
            </>
          )}
        </DialogActions>
      </Dialog>

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
}: {
  step: ProductionPlanStep | null;
  open: boolean;
  onClose: () => void;
  onSave: (updates: Partial<ProductionPlanStep>) => void;
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
          // Pre-select if step already references a station
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

          // Pre-populate from step's existing source fields
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

              // Pre-populate division for corporation
              if (step.sourceOwnerType === "corporation" && step.sourceDivisionNumber != null) {
                const corp = data.corporations.find((c) => c.id === step.sourceOwnerId);
                if (corp) {
                  const divName = corp.divisions[String(step.sourceDivisionNumber)] || `Division ${step.sourceDivisionNumber}`;
                  setSelectedDivision({ number: step.sourceDivisionNumber, name: divName });
                }
              }

              // Pre-populate container
              if (step.sourceContainerId) {
                const container = data.containers.find((c) => c.id === step.sourceContainerId);
                if (container) {
                  setSelectedContainer({ id: container.id, name: container.name });
                }
              }
            }
          }

          // Pre-populate output location
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
    // Reset input/output location when station changes
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
      // Auto-select rig based on rigCategory
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

  // Build output division options for selected output corporation
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

  // Build output container options filtered by selected output owner/division
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

  // Filter stations to only show those with matching activity
  const filteredStations = userStations.filter((s) =>
    step?.activity ? s.activities.includes(step.activity) : true,
  );

  const hasStation = !!selectedUserStation;

  // Resolve source_location_id from selected station
  const resolvedSourceLocationId = selectedUserStation
    ? selectedUserStation.stationId
    : null;

  if (!step) return null;

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
        Edit Step: {step.productName || `Type ${step.productTypeId}`}
      </DialogTitle>
      <DialogContent>
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2, mt: 1 }}>
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
                Where should materials for this step be pulled from?
              </Typography>

              {!hangarsLoaded ? (
                <Typography sx={{ color: "#64748b", fontSize: 13 }}>
                  Loading hangars...
                </Typography>
              ) : (
                <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
                  {/* Owner Selection */}
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

                  {/* Division Selection (corporation only) */}
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

                  {/* Container Selection (optional) */}
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
                Where should completed items from this job be delivered?
              </Typography>

              {!hangarsLoaded ? (
                <Typography sx={{ color: "#64748b", fontSize: 13 }}>
                  Loading hangars...
                </Typography>
              ) : (
                <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
                  {/* Output Owner Selection */}
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

                  {/* Output Division Selection (corporation only) */}
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

                  {/* Output Container Selection (optional) */}
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
            } as unknown as Partial<ProductionPlanStep>)
          }
          variant="contained"
          sx={{
            backgroundColor: "#3b82f6",
            "&:hover": { backgroundColor: "#2563eb" },
          }}
        >
          Save
        </Button>
      </DialogActions>
    </Dialog>
  );
}

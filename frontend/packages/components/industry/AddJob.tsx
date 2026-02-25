import { useState, useEffect, useCallback } from "react";
import {
  BlueprintSearchResult,
  BlueprintLevel,
  ManufacturingCalcResult,
  ReactionSystem,
} from "@industry-tool/client/data/models";
import { formatISK, formatNumber, formatDuration } from "@industry-tool/utils/formatting";
import Autocomplete from "@mui/material/Autocomplete";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import Paper from "@mui/material/Paper";
import AddIcon from "@mui/icons-material/Add";
import WarningAmberIcon from "@mui/icons-material/WarningAmber";
import Chip from "@mui/material/Chip";

type Props = {
  onJobAdded: () => void;
};

export default function AddJob({ onJobAdded }: Props) {
  const [blueprintQuery, setBlueprintQuery] = useState("");
  const [blueprintOptions, setBlueprintOptions] = useState<BlueprintSearchResult[]>([]);
  const [selectedBlueprint, setSelectedBlueprint] = useState<BlueprintSearchResult | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);

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
  const [calcResult, setCalcResult] = useState<ManufacturingCalcResult | null>(null);
  const [calcLoading, setCalcLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  // Fetch systems
  useEffect(() => {
    fetch("/api/industry/systems")
      .then((res) => res.json())
      .then((data) => setSystems(data))
      .catch((err) => console.error("Failed to fetch systems:", err));
  }, []);

  // Search blueprints
  useEffect(() => {
    if (blueprintQuery.length < 2) {
      setBlueprintOptions([]);
      return;
    }

    const timeout = setTimeout(async () => {
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

    return () => clearTimeout(timeout);
  }, [blueprintQuery, activity]);

  // Calculate cost
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

  return (
    <Box>
      {/* Settings Row */}
      <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap", mb: 3 }}>
        <Autocomplete
          sx={{ minWidth: 300, flexGrow: 1 }}
          options={blueprintOptions}
          getOptionLabel={(opt) => opt.ProductName || opt.BlueprintName}
          value={selectedBlueprint}
          onChange={(_, value) => {
            setSelectedBlueprint(value);
            if (value) {
              // Fetch blueprint level for the selected blueprint
              fetch("/api/industry/blueprint-levels", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ type_ids: [value.BlueprintTypeID] }),
              })
                .then((res) => res.json())
                .then((data: Record<string, BlueprintLevel | null>) => {
                  const level = data[String(value.BlueprintTypeID)] ?? null;
                  setDetectedLevel(level);
                  setDetectedForBlueprintId(value.BlueprintTypeID);
                  if (level) {
                    setMeLevel(level.materialEfficiency);
                    setTeLevel(level.timeEfficiency);
                  } else {
                    setMeLevel(10);
                    setTeLevel(20);
                  }
                })
                .catch((err) => console.error("Failed to fetch blueprint levels:", err));
            } else {
              setDetectedLevel(null);
              setDetectedForBlueprintId(null);
              setMeLevel(10);
              setTeLevel(20);
            }
          }}
          inputValue={blueprintQuery}
          onInputChange={(_, value) => setBlueprintQuery(value)}
          loading={searchLoading}
          renderInput={(params) => (
            <TextField {...params} label="Search Blueprint" size="small" />
          )}
          renderOption={(props, option) => (
            <li {...props} key={option.BlueprintTypeID}>
              <Box>
                <Typography variant="body2">{option.ProductName}</Typography>
                <Typography variant="caption" color="text.secondary">
                  {option.BlueprintName} - {option.Activity}
                </Typography>
              </Box>
            </li>
          )}
          isOptionEqualToValue={(opt, val) => opt.BlueprintTypeID === val.BlueprintTypeID}
        />

        <FormControl size="small" sx={{ minWidth: 150 }}>
          <InputLabel>Activity</InputLabel>
          <Select value={activity} onChange={(e) => setActivity(e.target.value)} label="Activity">
            <MenuItem value="manufacturing">Manufacturing</MenuItem>
            <MenuItem value="reaction">Reaction</MenuItem>
            <MenuItem value="invention">Invention</MenuItem>
            <MenuItem value="copying">Copying</MenuItem>
          </Select>
        </FormControl>

        <TextField
          label="Runs"
          type="number"
          size="small"
          value={runs}
          onChange={(e) => setRuns(Math.max(1, parseInt(e.target.value) || 1))}
          sx={{ width: 100 }}
        />
      </Box>

      <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap", mb: 3 }}>
        <TextField
          label="ME Level"
          type="number"
          size="small"
          value={meLevel}
          onChange={(e) => setMeLevel(Math.max(0, Math.min(10, parseInt(e.target.value) || 0)))}
          sx={{ width: 90 }}
          inputProps={{ min: 0, max: 10 }}
        />
        <TextField
          label="TE Level"
          type="number"
          size="small"
          value={teLevel}
          onChange={(e) => setTeLevel(Math.max(0, Math.min(20, parseInt(e.target.value) || 0)))}
          sx={{ width: 90 }}
          inputProps={{ min: 0, max: 20 }}
        />
        {selectedBlueprint && detectedForBlueprintId === selectedBlueprint.BlueprintTypeID ? (
          detectedLevel ? (
            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <Chip
                label={`Detected: ME ${detectedLevel.materialEfficiency} / TE ${detectedLevel.timeEfficiency} from ${detectedLevel.ownerName}${detectedLevel.isCopy ? " (BPC)" : ""}`}
                size="small"
                color="info"
                variant="outlined"
                sx={{ fontSize: 11 }}
              />
              {(meLevel !== detectedLevel.materialEfficiency || teLevel !== detectedLevel.timeEfficiency) && (
                <Chip
                  label="Overridden"
                  size="small"
                  color="warning"
                  variant="outlined"
                  sx={{ fontSize: 11 }}
                />
              )}
            </Box>
          ) : (
            <Chip
              icon={<WarningAmberIcon />}
              label="No blueprint detected — using manual values"
              size="small"
              color="warning"
              variant="outlined"
              sx={{ fontSize: 11 }}
            />
          )
        ) : null}
        <TextField
          label="Industry Skill"
          type="number"
          size="small"
          value={industrySkill}
          onChange={(e) => setIndustrySkill(Math.max(0, Math.min(5, parseInt(e.target.value) || 0)))}
          sx={{ width: 120 }}
          inputProps={{ min: 0, max: 5 }}
        />
        <TextField
          label="Adv Industry"
          type="number"
          size="small"
          value={advIndustrySkill}
          onChange={(e) => setAdvIndustrySkill(Math.max(0, Math.min(5, parseInt(e.target.value) || 0)))}
          sx={{ width: 120 }}
          inputProps={{ min: 0, max: 5 }}
        />

        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>Structure</InputLabel>
          <Select value={structure} onChange={(e) => setStructure(e.target.value)} label="Structure">
            <MenuItem value="station">NPC Station</MenuItem>
            <MenuItem value="raitaru">Raitaru</MenuItem>
            <MenuItem value="azbel">Azbel</MenuItem>
            <MenuItem value="sotiyo">Sotiyo</MenuItem>
          </Select>
        </FormControl>

        <FormControl size="small" sx={{ minWidth: 90 }}>
          <InputLabel>Rig</InputLabel>
          <Select value={rig} onChange={(e) => setRig(e.target.value)} label="Rig">
            <MenuItem value="none">None</MenuItem>
            <MenuItem value="t1">T1</MenuItem>
            <MenuItem value="t2">T2</MenuItem>
          </Select>
        </FormControl>

        <FormControl size="small" sx={{ minWidth: 100 }}>
          <InputLabel>Security</InputLabel>
          <Select value={security} onChange={(e) => setSecurity(e.target.value)} label="Security">
            <MenuItem value="high">Highsec</MenuItem>
            <MenuItem value="low">Lowsec</MenuItem>
            <MenuItem value="null">Nullsec</MenuItem>
          </Select>
        </FormControl>

        <TextField
          label="Facility Tax %"
          type="number"
          size="small"
          value={facilityTax}
          onChange={(e) => setFacilityTax(parseFloat(e.target.value) || 0)}
          sx={{ width: 120 }}
        />
      </Box>

      <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap", mb: 3, alignItems: "center" }}>
        <Autocomplete
          sx={{ minWidth: 250 }}
          options={systems}
          getOptionLabel={(opt) => `${opt.name} (${(opt.cost_index * 100).toFixed(2)}%)`}
          value={systems.find((s) => s.system_id === systemId) || null}
          onChange={(_, value) => setSystemId(value?.system_id || 0)}
          renderInput={(params) => (
            <TextField {...params} label="System (optional)" size="small" />
          )}
          isOptionEqualToValue={(opt, val) => opt.system_id === val.system_id}
        />

        <TextField
          label="Notes"
          size="small"
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          sx={{ flexGrow: 1, minWidth: 200 }}
        />

        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={!selectedBlueprint || runs <= 0 || submitting}
          startIcon={submitting ? <CircularProgress size={16} /> : <AddIcon />}
          sx={{ height: 40 }}
        >
          Add to Queue
        </Button>
      </Box>

      {/* Calculation Result */}
      {calcLoading && (
        <Box sx={{ display: "flex", justifyContent: "center", py: 4 }}>
          <CircularProgress size={24} />
        </Box>
      )}

      {calcResult && !calcLoading && (
        <Paper sx={{ backgroundColor: "#12151f", p: 2 }}>
          <Typography variant="subtitle2" sx={{ color: "#3b82f6", mb: 1 }}>
            Cost Estimate: {calcResult.productName} x{formatNumber(calcResult.totalProducts)}
          </Typography>
          <Box sx={{ display: "flex", gap: 4, flexWrap: "wrap", mb: 2 }}>
            <Box>
              <Typography variant="caption" sx={{ color: "#64748b" }}>Input Cost</Typography>
              <Typography variant="body2" sx={{ color: "#e2e8f0" }}>{formatISK(calcResult.inputCost)}</Typography>
            </Box>
            <Box>
              <Typography variant="caption" sx={{ color: "#64748b" }}>Job Cost</Typography>
              <Typography variant="body2" sx={{ color: "#e2e8f0" }}>{formatISK(calcResult.jobCost)}</Typography>
            </Box>
            <Box>
              <Typography variant="caption" sx={{ color: "#64748b" }}>Total Cost</Typography>
              <Typography variant="body2" sx={{ color: "#e2e8f0", fontWeight: 600 }}>{formatISK(calcResult.totalCost)}</Typography>
            </Box>
            <Box>
              <Typography variant="caption" sx={{ color: "#64748b" }}>Output Value</Typography>
              <Typography variant="body2" sx={{ color: "#e2e8f0" }}>{formatISK(calcResult.outputValue)}</Typography>
            </Box>
            <Box>
              <Typography variant="caption" sx={{ color: "#64748b" }}>Profit</Typography>
              <Typography variant="body2" sx={{ color: calcResult.profit >= 0 ? "#10b981" : "#ef4444" }}>
                {formatISK(calcResult.profit)}
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" sx={{ color: "#64748b" }}>Margin</Typography>
              <Typography variant="body2" sx={{ color: calcResult.margin >= 0 ? "#10b981" : "#ef4444" }}>
                {calcResult.margin.toFixed(1)}%
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" sx={{ color: "#64748b" }}>Per Run</Typography>
              <Typography variant="body2" sx={{ color: "#e2e8f0" }}>{formatDuration(calcResult.secsPerRun)}</Typography>
            </Box>
            <Box>
              <Typography variant="caption" sx={{ color: "#64748b" }}>Total Time</Typography>
              <Typography variant="body2" sx={{ color: "#e2e8f0" }}>{formatDuration(calcResult.totalDuration)}</Typography>
            </Box>
          </Box>

          {calcResult.materials.length > 0 && (
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow sx={{ backgroundColor: "#0f1219" }}>
                    <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>Material</TableCell>
                    <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="right">Base Qty</TableCell>
                    <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="right">Required</TableCell>
                    <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="right">Price</TableCell>
                    <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }} align="right">Cost</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {calcResult.materials.map((mat) => (
                    <TableRow key={mat.typeId}>
                      <TableCell sx={{ color: "#e2e8f0" }}>{mat.name}</TableCell>
                      <TableCell align="right" sx={{ color: "#94a3b8" }}>{formatNumber(mat.baseQty)}</TableCell>
                      <TableCell align="right" sx={{ color: "#e2e8f0" }}>{formatNumber(mat.batchQty)}</TableCell>
                      <TableCell align="right" sx={{ color: "#94a3b8" }}>{formatISK(mat.price)}</TableCell>
                      <TableCell align="right" sx={{ color: "#cbd5e1" }}>{formatISK(mat.cost)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Paper>
      )}
    </Box>
  );
}

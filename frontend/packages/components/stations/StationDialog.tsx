import { useState, useEffect, useRef, useCallback } from "react";
import {
  UserStation,
  UserStationRig,
  UserStationService,
  ScanResult,
} from "@industry-tool/client/data/models";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import Autocomplete from "@mui/material/Autocomplete";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Chip from "@mui/material/Chip";
import CircularProgress from "@mui/material/CircularProgress";
import IconButton from "@mui/material/IconButton";
import DeleteIcon from "@mui/icons-material/Delete";
import AddIcon from "@mui/icons-material/Add";

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
const rigTiers = ["t1", "t2"];

const getCategoryColor = (category: string) => {
  switch (category) {
    case "ship": return "#3b82f6";
    case "component": return "#8b5cf6";
    case "equipment": return "#10b981";
    case "ammo": return "#f59e0b";
    case "drone": return "#06b6d4";
    case "reaction": return "#ec4899";
    case "reprocessing": return "#f97316";
    case "thukker": return "#d97706";
    default: return "#94a3b8";
  }
};

export default function StationDialog({ open, station, onClose }: Props) {
  const isEdit = !!station;

  const [stationOptions, setStationOptions] = useState<StationOption[]>([]);
  const [stationSearchLoading, setStationSearchLoading] = useState(false);
  const [selectedStation, setSelectedStation] = useState<StationOption | null>(null);

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
    <Dialog
      open={open}
      onClose={() => onClose(false)}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: { backgroundColor: "#12151f", color: "#e2e8f0" },
      }}
    >
      <DialogTitle>{isEdit ? "Edit Station" : "Add Station"}</DialogTitle>
      <DialogContent>
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2, mt: 1 }}>
          {/* Station Search */}
          <Autocomplete
            value={selectedStation}
            onChange={(_, newValue) => setSelectedStation(newValue)}
            onInputChange={(_, value) => handleStationSearch(value)}
            options={stationOptions}
            getOptionLabel={(option) => option.name}
            isOptionEqualToValue={(option, value) =>
              option.stationId === value.stationId
            }
            loading={stationSearchLoading}
            disabled={isEdit}
            filterOptions={(x) => x}
            renderOption={(props, option) => (
              <Box component="li" {...props}>
                <Box>
                  <Typography variant="body2">{option.name}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {option.solarSystemName}
                  </Typography>
                </Box>
              </Box>
            )}
            renderInput={(params) => (
              <TextField
                {...params}
                label="Station"
                placeholder="Search for a station..."
                InputProps={{
                  ...params.InputProps,
                  endAdornment: (
                    <>
                      {stationSearchLoading ? (
                        <CircularProgress color="inherit" size={20} />
                      ) : null}
                      {params.InputProps.endAdornment}
                    </>
                  ),
                }}
              />
            )}
          />

          {/* Scan Input */}
          <TextField
            label="Structure Fitting Scan"
            multiline
            rows={4}
            value={scanText}
            onChange={(e) => setScanText(e.target.value)}
            placeholder="Paste structure fitting scan here..."
            sx={{ "& .MuiInputBase-input": { fontSize: 13 } }}
          />
          <Button
            variant="outlined"
            onClick={handleParseScan}
            disabled={!scanText.trim() || scanParsing}
            sx={{ alignSelf: "flex-start" }}
          >
            {scanParsing ? "Parsing..." : "Parse Scan"}
          </Button>

          {/* Structure */}
          <FormControl fullWidth>
            <InputLabel>Structure</InputLabel>
            <Select
              value={structure}
              label="Structure"
              onChange={(e) => {
                const newStructure = e.target.value;
                setStructure(newStructure);
                const validCategories = getRigCategoriesForStructure(newStructure);
                setRigs((prev) => prev.filter((r) => validCategories.includes(r.category)));
              }}
            >
              <MenuItem value="raitaru">Raitaru</MenuItem>
              <MenuItem value="azbel">Azbel</MenuItem>
              <MenuItem value="sotiyo">Sotiyo</MenuItem>
              <MenuItem value="athanor">Athanor</MenuItem>
              <MenuItem value="tatara">Tatara</MenuItem>
            </Select>
          </FormControl>

          {/* Facility Tax */}
          <TextField
            type="number"
            label="Facility Tax %"
            value={facilityTax}
            onChange={(e) => setFacilityTax(parseFloat(e.target.value) || 0)}
            inputProps={{ min: 0, step: 0.1 }}
          />

          {/* Services */}
          {services.length > 0 && (
            <Box>
              <Typography
                variant="subtitle2"
                sx={{ color: "#94a3b8", mb: 0.5 }}
              >
                Services
              </Typography>
              <Box sx={{ display: "flex", gap: 0.5, flexWrap: "wrap" }}>
                {services.map((svc, i) => (
                  <Chip
                    key={i}
                    label={svc.serviceName}
                    size="small"
                    sx={{
                      backgroundColor:
                        svc.activity === "manufacturing"
                          ? "#3b82f620"
                          : "#ec489920",
                      color:
                        svc.activity === "manufacturing"
                          ? "#3b82f6"
                          : "#ec4899",
                      fontSize: "0.75rem",
                    }}
                  />
                ))}
              </Box>
            </Box>
          )}

          {/* Rigs */}
          <Box>
            <Box
              sx={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                mb: 1,
              }}
            >
              <Typography
                variant="subtitle2"
                sx={{ color: "#94a3b8" }}
              >
                Rigs
              </Typography>
              <Button
                size="small"
                startIcon={<AddIcon />}
                onClick={handleAddRig}
              >
                Add Rig
              </Button>
            </Box>

            {rigs.length === 0 && (
              <Typography sx={{ color: "#64748b", fontSize: 13 }}>
                No rigs. Paste a scan or add manually.
              </Typography>
            )}

            {rigs.map((rig, index) => (
              <Box
                key={index}
                sx={{
                  display: "flex",
                  gap: 1,
                  mb: 1,
                  alignItems: "center",
                }}
              >
                <FormControl size="small" sx={{ minWidth: 140 }}>
                  <InputLabel>Category</InputLabel>
                  <Select
                    value={rig.category}
                    label="Category"
                    onChange={(e) =>
                      handleRigChange(index, "category", e.target.value)
                    }
                  >
                    {getRigCategoriesForStructure(structure).map((cat) => (
                      <MenuItem key={cat} value={cat}>
                        <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                          <Box
                            sx={{
                              width: 8,
                              height: 8,
                              borderRadius: "50%",
                              backgroundColor: getCategoryColor(cat),
                            }}
                          />
                          <span style={{ textTransform: "capitalize" }}>{cat}</span>
                        </Box>
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>

                <FormControl size="small" sx={{ minWidth: 80 }}>
                  <InputLabel>Tier</InputLabel>
                  <Select
                    value={rig.tier}
                    label="Tier"
                    onChange={(e) =>
                      handleRigChange(index, "tier", e.target.value)
                    }
                  >
                    {rigTiers.map((t) => (
                      <MenuItem key={t} value={t}>
                        {t.toUpperCase()}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>

                <Typography
                  sx={{
                    color: "#64748b",
                    fontSize: 12,
                    flex: 1,
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                  }}
                  title={rig.rigName}
                >
                  {rig.rigName || "Manual"}
                </Typography>

                <IconButton
                  size="small"
                  onClick={() => handleRemoveRig(index)}
                  sx={{ color: "#ef4444" }}
                >
                  <DeleteIcon fontSize="small" />
                </IconButton>
              </Box>
            ))}
          </Box>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => onClose(false)} sx={{ color: "#94a3b8" }}>
          Cancel
        </Button>
        <Button
          onClick={handleSave}
          disabled={!selectedStation || saving}
          variant="contained"
          sx={{
            backgroundColor: "#3b82f6",
            "&:hover": { backgroundColor: "#2563eb" },
          }}
        >
          {saving ? "Saving..." : isEdit ? "Update" : "Add"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

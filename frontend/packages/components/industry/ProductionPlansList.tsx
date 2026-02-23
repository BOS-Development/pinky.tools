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
import Container from "@mui/material/Container";
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
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import AddIcon from "@mui/icons-material/Add";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import TextField from "@mui/material/TextField";
import Autocomplete from "@mui/material/Autocomplete";
import CircularProgress from "@mui/material/CircularProgress";

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
        <Container maxWidth="xl" sx={{ mt: 2, mb: 4 }}>
          <Button
            startIcon={<ArrowBackIcon />}
            onClick={handleBack}
            sx={{ color: "#94a3b8", mb: 2 }}
          >
            Back to Plans
          </Button>
          <ProductionPlanEditor planId={selectedPlanId} />
        </Container>
      </>
    );
  }

  return (
    <>
      <Navbar />
      <Container maxWidth="xl" sx={{ mt: 2, mb: 4 }}>
        <Box
          sx={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            mb: 2,
          }}
        >
          <Typography
            variant="h5"
            sx={{ color: "#e2e8f0", fontWeight: 600 }}
          >
            Production Plans
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setCreateDialogOpen(true)}
            sx={{
              backgroundColor: "#3b82f6",
              "&:hover": { backgroundColor: "#2563eb" },
            }}
          >
            New Plan
          </Button>
        </Box>

        {loading ? (
          <Loading />
        ) : plans.length === 0 ? (
          <Paper
            sx={{
              backgroundColor: "#12151f",
              p: 4,
              textAlign: "center",
            }}
          >
            <Typography sx={{ color: "#64748b" }}>
              No production plans yet. Create one to define how items should be
              produced.
            </Typography>
          </Paper>
        ) : (
          <TableContainer
            component={Paper}
            sx={{ backgroundColor: "#12151f" }}
          >
            <Table size="small">
              <TableHead>
                <TableRow sx={{ backgroundColor: "#0f1219" }}>
                  <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                    Product
                  </TableCell>
                  <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                    Plan Name
                  </TableCell>
                  <TableCell
                    sx={{ color: "#94a3b8", fontWeight: 600 }}
                    align="right"
                  >
                    Steps
                  </TableCell>
                  <TableCell sx={{ color: "#94a3b8", fontWeight: 600 }}>
                    Created
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
                {plans.map((plan, idx) => (
                  <TableRow
                    key={plan.id}
                    sx={{
                      backgroundColor:
                        idx % 2 === 0 ? "#12151f" : "#0f1219",
                      "&:hover": { backgroundColor: "#1a1d2e" },
                      cursor: "pointer",
                    }}
                    onClick={() => setSelectedPlanId(plan.id)}
                  >
                    <TableCell>
                      <Box
                        sx={{ display: "flex", alignItems: "center", gap: 1 }}
                      >
                        <img
                          src={`https://images.evetech.net/types/${plan.productTypeId}/icon?size=32`}
                          alt=""
                          width={24}
                          height={24}
                          style={{ borderRadius: 2 }}
                        />
                        <Typography sx={{ color: "#e2e8f0", fontSize: 14 }}>
                          {plan.productName || `Type ${plan.productTypeId}`}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell sx={{ color: "#cbd5e1", fontSize: 14 }}>
                      {plan.name}
                    </TableCell>
                    <TableCell
                      align="right"
                      sx={{ color: "#cbd5e1", fontSize: 14 }}
                    >
                      {plan.steps?.length || 0}
                    </TableCell>
                    <TableCell sx={{ color: "#64748b", fontSize: 13 }}>
                      {new Date(plan.createdAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell align="center">
                      <IconButton
                        size="small"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedPlanId(plan.id);
                        }}
                        sx={{ color: "#3b82f6" }}
                      >
                        <EditIcon fontSize="small" />
                      </IconButton>
                      <IconButton
                        size="small"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDelete(plan.id);
                        }}
                        sx={{ color: "#ef4444" }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}

        <CreatePlanDialog
          open={createDialogOpen}
          onClose={() => setCreateDialogOpen(false)}
          onCreated={handlePlanCreated}
        />
      </Container>
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
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{ sx: { backgroundColor: "#12151f", color: "#e2e8f0" } }}
    >
      <DialogTitle>Create Production Plan</DialogTitle>
      <DialogContent>
        <Autocomplete
          sx={{ mt: 1 }}
          options={searchResults}
          getOptionLabel={(option) =>
            `${option.ProductName} (${option.Activity})`
          }
          renderOption={(props, option) => (
            <Box
              component="li"
              {...props}
              key={`${option.BlueprintTypeID}-${option.Activity}`}
              sx={{ display: "flex", alignItems: "center", gap: 1 }}
            >
              <img
                src={`https://images.evetech.net/types/${option.ProductTypeID}/icon?size=32`}
                alt=""
                width={24}
                height={24}
              />
              <Box>
                <Typography sx={{ fontSize: 14 }}>
                  {option.ProductName}
                </Typography>
                <Typography sx={{ fontSize: 12, color: "#64748b" }}>
                  {option.BlueprintName} &middot; {option.Activity}
                </Typography>
              </Box>
            </Box>
          )}
          loading={searchLoading}
          onInputChange={(_, value) => handleSearch(value)}
          onChange={(_, value) => setSelectedBlueprint(value)}
          renderInput={(params) => (
            <TextField
              {...params}
              label="Search for a product"
              placeholder="e.g. Rifter, Damage Control II..."
              InputProps={{
                ...params.InputProps,
                endAdornment: (
                  <>
                    {searchLoading ? (
                      <CircularProgress color="inherit" size={20} />
                    ) : null}
                    {params.InputProps.endAdornment}
                  </>
                ),
              }}
            />
          )}
          noOptionsText={
            searchQuery.length < 2
              ? "Type to search..."
              : "No blueprints found"
          }
        />
        <Autocomplete
          sx={{ mt: 2 }}
          value={defaultManufacturingStation}
          onChange={(_, newValue) => setDefaultManufacturingStation(newValue)}
          options={manufacturingStations}
          getOptionLabel={(option) =>
            `${option.stationName || "Unknown"} (${option.solarSystemName || ""})`
          }
          isOptionEqualToValue={(option, value) => option.id === value.id}
          renderOption={(props, option) => (
            <Box component="li" {...props} key={option.id}>
              <Box>
                <Typography variant="body2">
                  {option.stationName || "Unknown"}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {option.solarSystemName} &middot; {option.structure}
                </Typography>
              </Box>
            </Box>
          )}
          renderInput={(params) => (
            <TextField
              {...params}
              label="Default Manufacturing Station"
              placeholder="Optional"
            />
          )}
        />
        <Autocomplete
          sx={{ mt: 2 }}
          value={defaultReactionStation}
          onChange={(_, newValue) => setDefaultReactionStation(newValue)}
          options={reactionStations}
          getOptionLabel={(option) =>
            `${option.stationName || "Unknown"} (${option.solarSystemName || ""})`
          }
          isOptionEqualToValue={(option, value) => option.id === value.id}
          renderOption={(props, option) => (
            <Box component="li" {...props} key={option.id}>
              <Box>
                <Typography variant="body2">
                  {option.stationName || "Unknown"}
                </Typography>
                <Typography variant="caption" color="text.secondary">
                  {option.solarSystemName} &middot; {option.structure}
                </Typography>
              </Box>
            </Box>
          )}
          renderInput={(params) => (
            <TextField
              {...params}
              label="Default Reaction Station"
              placeholder="Optional"
            />
          )}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} sx={{ color: "#94a3b8" }}>
          Cancel
        </Button>
        <Button
          onClick={handleCreate}
          disabled={!selectedBlueprint || creating}
          variant="contained"
          sx={{
            backgroundColor: "#3b82f6",
            "&:hover": { backgroundColor: "#2563eb" },
          }}
        >
          {creating ? "Creating..." : "Create Plan"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

import React, { useCallback, useEffect, useRef, useState } from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Box,
  Typography,
  Autocomplete,
  CircularProgress,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import { TransportProfile, JFRoute } from "../../pages/transport";
import { formatNumber } from "../../utils/formatting";

interface StationOption {
  stationId: number;
  name: string;
  solarSystemId: number;
  solarSystemName: string;
  security: number;
}

interface ItemTypeOption {
  TypeID: number;
  TypeName: string;
  Volume: number;
}

interface JobItemEntry {
  itemType: ItemTypeOption;
  quantity: number;
}

interface Props {
  open: boolean;
  onClose: (saved: boolean) => void;
  profiles: TransportProfile[];
  jfRoutes: JFRoute[];
}

const getSecurityColor = (sec: number) => {
  if (sec >= 0.5) return "#10b981";
  if (sec > 0.0) return "#f59e0b";
  return "#ef4444";
};

export function TransportJobDialog({ open, onClose, profiles, jfRoutes }: Props) {
  const [saving, setSaving] = useState(false);
  const [originStation, setOriginStation] = useState<StationOption | null>(null);
  const [destinationStation, setDestinationStation] = useState<StationOption | null>(null);
  const [transportMethod, setTransportMethod] = useState("freighter");
  const [fulfillmentType, setFulfillmentType] = useState("self_haul");
  const [transportProfileId, setTransportProfileId] = useState<string>("");
  const [jfRouteId, setJfRouteId] = useState<string>("");
  const [notes, setNotes] = useState("");

  // Items state
  const [items, setItems] = useState<JobItemEntry[]>([]);
  const [selectedItemType, setSelectedItemType] = useState<ItemTypeOption | null>(null);
  const [itemQuantity, setItemQuantity] = useState("");
  const [itemTypeOptions, setItemTypeOptions] = useState<ItemTypeOption[]>([]);
  const [itemTypeLoading, setItemTypeLoading] = useState(false);
  const itemTypeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Station search state
  const [originOptions, setOriginOptions] = useState<StationOption[]>([]);
  const [destOptions, setDestOptions] = useState<StationOption[]>([]);
  const [originLoading, setOriginLoading] = useState(false);
  const [destLoading, setDestLoading] = useState(false);
  const originTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const destTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (open) {
      setOriginStation(null);
      setDestinationStation(null);
      setTransportMethod("freighter");
      setFulfillmentType("self_haul");
      setTransportProfileId("");
      setJfRouteId("");
      setNotes("");
      setOriginOptions([]);
      setDestOptions([]);
      setItems([]);
      setSelectedItemType(null);
      setItemQuantity("");
      setItemTypeOptions([]);
    }
  }, [open]);

  const searchStations = useCallback(
    async (
      query: string,
      setOptions: (opts: StationOption[]) => void,
      setLoading: (l: boolean) => void,
    ) => {
      if (!query || query.length < 2) {
        setOptions([]);
        return;
      }
      setLoading(true);
      try {
        const res = await fetch(`/api/stations/search?q=${encodeURIComponent(query)}`);
        if (res.ok) {
          const data = await res.json();
          setOptions(data || []);
        }
      } catch (err) {
        console.error("Failed to search stations:", err);
        setOptions([]);
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  const handleOriginSearch = (value: string) => {
    if (originTimerRef.current) clearTimeout(originTimerRef.current);
    originTimerRef.current = setTimeout(() => {
      searchStations(value, setOriginOptions, setOriginLoading);
    }, 300);
  };

  const handleDestSearch = (value: string) => {
    if (destTimerRef.current) clearTimeout(destTimerRef.current);
    destTimerRef.current = setTimeout(() => {
      searchStations(value, setDestOptions, setDestLoading);
    }, 300);
  };

  const handleItemTypeSearch = (value: string) => {
    if (itemTypeTimerRef.current) clearTimeout(itemTypeTimerRef.current);
    if (!value || value.length < 2) {
      setItemTypeOptions([]);
      return;
    }
    itemTypeTimerRef.current = setTimeout(async () => {
      setItemTypeLoading(true);
      try {
        const res = await fetch(`/api/item-types/search?q=${encodeURIComponent(value)}`);
        if (res.ok) {
          const data = await res.json();
          setItemTypeOptions(data || []);
        }
      } catch (err) {
        console.error("Failed to search item types:", err);
        setItemTypeOptions([]);
      } finally {
        setItemTypeLoading(false);
      }
    }, 300);
  };

  const handleAddItem = () => {
    if (!selectedItemType || !itemQuantity) return;
    const qty = parseInt(itemQuantity.replace(/,/g, ""), 10);
    if (qty <= 0 || isNaN(qty)) return;

    const existing = items.find((i) => i.itemType.TypeID === selectedItemType.TypeID);
    if (existing) {
      setItems(
        items.map((i) =>
          i.itemType.TypeID === selectedItemType.TypeID
            ? { ...i, quantity: i.quantity + qty }
            : i,
        ),
      );
    } else {
      setItems([...items, { itemType: selectedItemType, quantity: qty }]);
    }
    setSelectedItemType(null);
    setItemQuantity("");
    setItemTypeOptions([]);
  };

  const handleRemoveItem = (typeId: number) => {
    setItems(items.filter((i) => i.itemType.TypeID !== typeId));
  };

  const totalVolume = items.reduce(
    (sum, i) => sum + i.itemType.Volume * i.quantity,
    0,
  );

  const handleSave = async () => {
    if (!originStation || !destinationStation || items.length === 0) return;
    setSaving(true);
    try {
      const payload: Record<string, unknown> = {
        originStationId: originStation.stationId,
        destinationStationId: destinationStation.stationId,
        originSystemId: originStation.solarSystemId,
        destinationSystemId: destinationStation.solarSystemId,
        transportMethod,
        fulfillmentType,
        items: items.map((i) => ({
          typeId: i.itemType.TypeID,
          quantity: i.quantity,
          volumeM3: i.itemType.Volume * i.quantity,
          estimatedValue: 0,
        })),
      };

      if (transportProfileId) {
        payload.transportProfileId = Number(transportProfileId);
      }
      if (jfRouteId) {
        payload.jfRouteId = Number(jfRouteId);
      }
      if (notes) {
        payload.notes = notes;
      }

      const res = await fetch("/api/transport/jobs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        onClose(true);
      }
    } catch (error) {
      console.error("Failed to create job:", error);
    } finally {
      setSaving(false);
    }
  };

  const canSave = !!originStation && !!destinationStation && items.length > 0;
  const isJF = transportMethod === "jump_freighter";
  const filteredProfiles = profiles.filter((p) => p.transportMethod === transportMethod);

  return (
    <Dialog
      open={open}
      onClose={() => onClose(false)}
      maxWidth="md"
      fullWidth
      PaperProps={{ sx: { backgroundColor: "#12151f", backgroundImage: "none" } }}
    >
      <DialogTitle>Create Transport Job</DialogTitle>
      <DialogContent>
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2, mt: 1 }}>
          <Autocomplete
            value={originStation}
            onChange={(_, newValue) => setOriginStation(newValue)}
            onInputChange={(_, inputValue) => handleOriginSearch(inputValue)}
            options={originOptions}
            getOptionLabel={(option) => option.name}
            isOptionEqualToValue={(a, b) => a.stationId === b.stationId}
            loading={originLoading}
            filterOptions={(x) => x}
            size="small"
            renderOption={(props, option) => (
              <Box component="li" {...props}>
                <Box>
                  <Typography variant="body2">{option.name}</Typography>
                  <Typography variant="caption" sx={{ color: "#94a3b8" }}>
                    {option.solarSystemName}{" "}
                    <span style={{ color: getSecurityColor(option.security ?? 0) }}>
                      ({(option.security ?? 0).toFixed(1)})
                    </span>
                  </Typography>
                </Box>
              </Box>
            )}
            renderInput={(params) => (
              <TextField
                {...params}
                label="Origin Station"
                placeholder="Search for a station..."
                InputProps={{
                  ...params.InputProps,
                  endAdornment: (
                    <>
                      {originLoading ? <CircularProgress color="inherit" size={20} /> : null}
                      {params.InputProps.endAdornment}
                    </>
                  ),
                }}
              />
            )}
          />

          {originStation && (
            <Typography variant="caption" sx={{ color: "#94a3b8", mt: -1 }}>
              System: {originStation.solarSystemName}
            </Typography>
          )}

          <Autocomplete
            value={destinationStation}
            onChange={(_, newValue) => setDestinationStation(newValue)}
            onInputChange={(_, inputValue) => handleDestSearch(inputValue)}
            options={destOptions}
            getOptionLabel={(option) => option.name}
            isOptionEqualToValue={(a, b) => a.stationId === b.stationId}
            loading={destLoading}
            filterOptions={(x) => x}
            size="small"
            renderOption={(props, option) => (
              <Box component="li" {...props}>
                <Box>
                  <Typography variant="body2">{option.name}</Typography>
                  <Typography variant="caption" sx={{ color: "#94a3b8" }}>
                    {option.solarSystemName}{" "}
                    <span style={{ color: getSecurityColor(option.security ?? 0) }}>
                      ({(option.security ?? 0).toFixed(1)})
                    </span>
                  </Typography>
                </Box>
              </Box>
            )}
            renderInput={(params) => (
              <TextField
                {...params}
                label="Destination Station"
                placeholder="Search for a station..."
                InputProps={{
                  ...params.InputProps,
                  endAdornment: (
                    <>
                      {destLoading ? <CircularProgress color="inherit" size={20} /> : null}
                      {params.InputProps.endAdornment}
                    </>
                  ),
                }}
              />
            )}
          />

          {destinationStation && (
            <Typography variant="caption" sx={{ color: "#94a3b8", mt: -1 }}>
              System: {destinationStation.solarSystemName}
            </Typography>
          )}

          {/* Items Section */}
          <Typography variant="subtitle2" sx={{ mt: 1 }}>
            Items to Transport
          </Typography>

          <Box sx={{ display: "flex", gap: 1, alignItems: "flex-start" }}>
            <Autocomplete
              value={selectedItemType}
              onChange={(_, newValue) => setSelectedItemType(newValue)}
              onInputChange={(_, inputValue) => handleItemTypeSearch(inputValue)}
              options={itemTypeOptions}
              getOptionLabel={(option) => option.TypeName}
              isOptionEqualToValue={(a, b) => a.TypeID === b.TypeID}
              loading={itemTypeLoading}
              filterOptions={(x) => x}
              size="small"
              sx={{ flex: 2 }}
              renderOption={(props, option) => (
                <Box component="li" {...props}>
                  <Box sx={{ display: "flex", gap: 1, alignItems: "center" }}>
                    <img
                      src={`https://images.evetech.net/types/${option.TypeID}/icon?size=32`}
                      alt=""
                      style={{ width: 24, height: 24 }}
                    />
                    <Box>
                      <Typography variant="body2">{option.TypeName}</Typography>
                      <Typography variant="caption" sx={{ color: "#94a3b8" }}>
                        {option.Volume.toLocaleString()} m³
                      </Typography>
                    </Box>
                  </Box>
                </Box>
              )}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Item Type"
                  placeholder="Search for an item..."
                  InputProps={{
                    ...params.InputProps,
                    endAdornment: (
                      <>
                        {itemTypeLoading ? (
                          <CircularProgress color="inherit" size={20} />
                        ) : null}
                        {params.InputProps.endAdornment}
                      </>
                    ),
                  }}
                />
              )}
            />
            <TextField
              label="Quantity"
              value={itemQuantity}
              onChange={(e) => {
                const raw = e.target.value.replace(/[^0-9]/g, "");
                if (raw === "") {
                  setItemQuantity("");
                } else {
                  setItemQuantity(Number(raw).toLocaleString());
                }
              }}
              size="small"
              sx={{ flex: 1 }}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handleAddItem();
                }
              }}
            />
            <IconButton
              onClick={handleAddItem}
              disabled={!selectedItemType || !itemQuantity}
              sx={{ color: "#3b82f6", mt: 0.5 }}
            >
              <AddIcon />
            </IconButton>
          </Box>

          {items.length > 0 && (
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow sx={{ "& th": { color: "#94a3b8", borderColor: "#1e2231" } }}>
                    <TableCell>Item</TableCell>
                    <TableCell align="right">Quantity</TableCell>
                    <TableCell align="right">Volume (m³)</TableCell>
                    <TableCell align="right" sx={{ width: 50 }} />
                  </TableRow>
                </TableHead>
                <TableBody>
                  {items.map((item) => (
                    <TableRow
                      key={item.itemType.TypeID}
                      sx={{ "& td": { borderColor: "#1e2231" } }}
                    >
                      <TableCell>
                        <Box sx={{ display: "flex", gap: 1, alignItems: "center" }}>
                          <img
                            src={`https://images.evetech.net/types/${item.itemType.TypeID}/icon?size=32`}
                            alt=""
                            style={{ width: 20, height: 20 }}
                          />
                          {item.itemType.TypeName}
                        </Box>
                      </TableCell>
                      <TableCell align="right">
                        {formatNumber(item.quantity)}
                      </TableCell>
                      <TableCell align="right">
                        {formatNumber(item.itemType.Volume * item.quantity)}
                      </TableCell>
                      <TableCell align="right">
                        <IconButton
                          size="small"
                          onClick={() => handleRemoveItem(item.itemType.TypeID)}
                          sx={{ color: "#ef4444" }}
                        >
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
                  <TableRow sx={{ "& td": { borderColor: "#1e2231", fontWeight: 600 } }}>
                    <TableCell>Total</TableCell>
                    <TableCell align="right">
                      {formatNumber(items.reduce((sum, i) => sum + i.quantity, 0))}
                    </TableCell>
                    <TableCell align="right">
                      {formatNumber(totalVolume)} m³
                    </TableCell>
                    <TableCell />
                  </TableRow>
                </TableBody>
              </Table>
            </TableContainer>
          )}

          {items.length === 0 && (
            <Typography
              variant="body2"
              sx={{ color: "#64748b", textAlign: "center", py: 1 }}
            >
              No items added yet. Search for items above and add them to this job.
            </Typography>
          )}

          <FormControl fullWidth size="small">
            <InputLabel>Transport Method</InputLabel>
            <Select
              value={transportMethod}
              onChange={(e) => {
                setTransportMethod(e.target.value);
                setTransportProfileId("");
                setJfRouteId("");
              }}
              label="Transport Method"
            >
              <MenuItem value="freighter">Freighter</MenuItem>
              <MenuItem value="jump_freighter">Jump Freighter</MenuItem>
              <MenuItem value="dst">DST</MenuItem>
              <MenuItem value="blockade_runner">Blockade Runner</MenuItem>
            </Select>
          </FormControl>

          <FormControl fullWidth size="small">
            <InputLabel>Fulfillment Type</InputLabel>
            <Select
              value={fulfillmentType}
              onChange={(e) => setFulfillmentType(e.target.value)}
              label="Fulfillment Type"
            >
              <MenuItem value="self_haul">Self Haul</MenuItem>
              <MenuItem value="courier_contract">Courier Contract</MenuItem>
              <MenuItem value="contact_haul">Contact Haul</MenuItem>
            </Select>
          </FormControl>

          {filteredProfiles.length > 0 && (
            <FormControl fullWidth size="small">
              <InputLabel>Transport Profile</InputLabel>
              <Select
                value={transportProfileId}
                onChange={(e) => setTransportProfileId(e.target.value)}
                label="Transport Profile"
              >
                <MenuItem value="">None</MenuItem>
                {filteredProfiles.map((p) => (
                  <MenuItem key={p.id} value={String(p.id)}>
                    {p.name} {p.isDefault ? "(Default)" : ""}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          )}

          {isJF && jfRoutes.length > 0 && (
            <FormControl fullWidth size="small">
              <InputLabel>JF Route</InputLabel>
              <Select
                value={jfRouteId}
                onChange={(e) => setJfRouteId(e.target.value)}
                label="JF Route"
              >
                <MenuItem value="">None</MenuItem>
                {jfRoutes.map((r) => (
                  <MenuItem key={r.id} value={String(r.id)}>
                    {r.name} ({(r.totalDistanceLy ?? 0).toFixed(1)} LY)
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          )}

          <TextField
            label="Notes"
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            fullWidth
            size="small"
            multiline
            rows={2}
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => onClose(false)} sx={{ color: "#94a3b8" }}>
          Cancel
        </Button>
        <Button variant="contained" onClick={handleSave} disabled={saving || !canSave}>
          {saving ? "Creating..." : "Create Job"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

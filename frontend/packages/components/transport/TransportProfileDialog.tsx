import React, { useEffect, useState } from "react";
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
  FormControlLabel,
  Checkbox,
} from "@mui/material";
import { TransportProfile } from "../../pages/transport";

interface Props {
  open: boolean;
  onClose: (saved: boolean) => void;
  profile: TransportProfile | null;
}

export function TransportProfileDialog({ open, onClose, profile }: Props) {
  const isEdit = !!profile;
  const [saving, setSaving] = useState(false);
  const [name, setName] = useState("");
  const [transportMethod, setTransportMethod] = useState("freighter");
  const [cargoM3, setCargoM3] = useState<number>(350000);
  const [ratePerM3PerJump, setRatePerM3PerJump] = useState<number>(0);
  const [collateralRate, setCollateralRate] = useState<number>(0.01);
  const [collateralPriceBasis, setCollateralPriceBasis] = useState("sell");
  const [fuelPerLy, setFuelPerLy] = useState<number>(0);
  const [fuelConservationLevel, setFuelConservationLevel] = useState<number>(0);
  const [routePreference, setRoutePreference] = useState("shortest");
  const [isDefault, setIsDefault] = useState(false);

  useEffect(() => {
    if (open && profile) {
      setName(profile.name);
      setTransportMethod(profile.transportMethod);
      setCargoM3(profile.cargoM3);
      setRatePerM3PerJump(profile.ratePerM3PerJump);
      setCollateralRate(profile.collateralRate);
      setCollateralPriceBasis(profile.collateralPriceBasis || "sell");
      setFuelPerLy(profile.fuelPerLy || 0);
      setFuelConservationLevel(profile.fuelConservationLevel);
      setRoutePreference(profile.routePreference || "shortest");
      setIsDefault(profile.isDefault);
    } else if (open) {
      setName("");
      setTransportMethod("freighter");
      setCargoM3(350000);
      setRatePerM3PerJump(0);
      setCollateralRate(0.01);
      setCollateralPriceBasis("sell");
      setFuelPerLy(0);
      setFuelConservationLevel(0);
      setRoutePreference("shortest");
      setIsDefault(false);
    }
  }, [open, profile]);

  const handleSave = async () => {
    setSaving(true);
    try {
      const payload = {
        name,
        transportMethod,
        cargoM3,
        ratePerM3PerJump,
        collateralRate,
        collateralPriceBasis,
        fuelPerLy: transportMethod === "jump_freighter" ? fuelPerLy : undefined,
        fuelConservationLevel: transportMethod === "jump_freighter" ? fuelConservationLevel : 0,
        routePreference,
        isDefault,
      };

      const url = isEdit
        ? `/api/transport/profiles/${profile!.id}`
        : "/api/transport/profiles";
      const method = isEdit ? "PUT" : "POST";

      const res = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        onClose(true);
      }
    } catch (error) {
      console.error("Failed to save profile:", error);
    } finally {
      setSaving(false);
    }
  };

  const isJF = transportMethod === "jump_freighter";

  return (
    <Dialog
      open={open}
      onClose={() => onClose(false)}
      maxWidth="sm"
      fullWidth
      PaperProps={{ sx: { backgroundColor: "#12151f", backgroundImage: "none" } }}
    >
      <DialogTitle>{isEdit ? "Edit Transport Profile" : "Add Transport Profile"}</DialogTitle>
      <DialogContent>
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2, mt: 1 }}>
          <TextField
            label="Profile Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            fullWidth
            size="small"
          />

          <FormControl fullWidth size="small">
            <InputLabel>Transport Method</InputLabel>
            <Select
              value={transportMethod}
              onChange={(e) => setTransportMethod(e.target.value)}
              label="Transport Method"
            >
              <MenuItem value="freighter">Freighter</MenuItem>
              <MenuItem value="jump_freighter">Jump Freighter</MenuItem>
              <MenuItem value="dst">DST</MenuItem>
              <MenuItem value="blockade_runner">Blockade Runner</MenuItem>
            </Select>
          </FormControl>

          <TextField
            label="Cargo Capacity (m3)"
            type="number"
            value={cargoM3}
            onChange={(e) => setCargoM3(Number(e.target.value))}
            fullWidth
            size="small"
          />

          {!isJF && (
            <TextField
              label="Rate per m3 per Jump (ISK)"
              type="number"
              value={ratePerM3PerJump}
              onChange={(e) => setRatePerM3PerJump(Number(e.target.value))}
              fullWidth
              size="small"
            />
          )}

          <TextField
            label="Collateral Rate"
            type="number"
            value={collateralRate}
            onChange={(e) => setCollateralRate(Number(e.target.value))}
            fullWidth
            size="small"
            inputProps={{ step: 0.001 }}
            helperText="e.g. 0.01 = 1%"
          />

          <FormControl fullWidth size="small">
            <InputLabel>Collateral Price Basis</InputLabel>
            <Select
              value={collateralPriceBasis}
              onChange={(e) => setCollateralPriceBasis(e.target.value)}
              label="Collateral Price Basis"
            >
              <MenuItem value="buy">Buy</MenuItem>
              <MenuItem value="sell">Sell</MenuItem>
              <MenuItem value="split">Split (avg)</MenuItem>
            </Select>
          </FormControl>

          {isJF && (
            <>
              <TextField
                label="Fuel per Light Year"
                type="number"
                value={fuelPerLy}
                onChange={(e) => setFuelPerLy(Number(e.target.value))}
                fullWidth
                size="small"
              />
              <FormControl fullWidth size="small">
                <InputLabel>Fuel Conservation Level</InputLabel>
                <Select
                  value={fuelConservationLevel}
                  onChange={(e) => setFuelConservationLevel(Number(e.target.value))}
                  label="Fuel Conservation Level"
                >
                  {[0, 1, 2, 3, 4, 5].map((level) => (
                    <MenuItem key={level} value={level}>
                      Level {level} ({level * 10}% reduction)
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </>
          )}

          <FormControl fullWidth size="small">
            <InputLabel>Route Preference</InputLabel>
            <Select
              value={routePreference}
              onChange={(e) => setRoutePreference(e.target.value)}
              label="Route Preference"
            >
              <MenuItem value="shortest">Shortest</MenuItem>
              <MenuItem value="secure">Secure</MenuItem>
              <MenuItem value="insecure">Insecure</MenuItem>
            </Select>
          </FormControl>

          <FormControlLabel
            control={
              <Checkbox
                checked={isDefault}
                onChange={(e) => setIsDefault(e.target.checked)}
                sx={{ color: "#94a3b8", "&.Mui-checked": { color: "#3b82f6" } }}
              />
            }
            label="Default profile for this method"
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => onClose(false)} sx={{ color: "#94a3b8" }}>
          Cancel
        </Button>
        <Button
          variant="contained"
          onClick={handleSave}
          disabled={saving || !name}
        >
          {saving ? "Saving..." : isEdit ? "Update" : "Create"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

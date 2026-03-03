import React, { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose(false); }}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit Transport Profile" : "Add Transport Profile"}</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-3 pt-1">
          <div className="flex flex-col gap-1">
            <label className="text-xs text-[#94a3b8]">Profile Name</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Profile name..."
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[#94a3b8]">Transport Method</label>
            <Select
              value={transportMethod}
              onValueChange={(v) => setTransportMethod(v)}
            >
              <SelectTrigger>
                <SelectValue placeholder="Transport Method" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="freighter">Freighter</SelectItem>
                <SelectItem value="jump_freighter">Jump Freighter</SelectItem>
                <SelectItem value="dst">DST</SelectItem>
                <SelectItem value="blockade_runner">Blockade Runner</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[#94a3b8]">Cargo Capacity (m3)</label>
            <Input
              type="number"
              value={cargoM3}
              onChange={(e) => setCargoM3(Number(e.target.value))}
            />
          </div>

          {!isJF && (
            <div className="flex flex-col gap-1">
              <label className="text-xs text-[#94a3b8]">Rate per m3 per Jump (ISK)</label>
              <Input
                type="number"
                value={ratePerM3PerJump}
                onChange={(e) => setRatePerM3PerJump(Number(e.target.value))}
              />
            </div>
          )}

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[#94a3b8]">Collateral Rate</label>
            <Input
              type="number"
              value={collateralRate}
              onChange={(e) => setCollateralRate(Number(e.target.value))}
              step={0.001}
            />
            <span className="text-xs text-[#64748b]">e.g. 0.01 = 1%</span>
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[#94a3b8]">Collateral Price Basis</label>
            <Select
              value={collateralPriceBasis}
              onValueChange={(v) => setCollateralPriceBasis(v)}
            >
              <SelectTrigger>
                <SelectValue placeholder="Collateral Price Basis" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="buy">Buy</SelectItem>
                <SelectItem value="sell">Sell</SelectItem>
                <SelectItem value="split">Split (avg)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {isJF && (
            <>
              <div className="flex flex-col gap-1">
                <label className="text-xs text-[#94a3b8]">Fuel per Light Year</label>
                <Input
                  type="number"
                  value={fuelPerLy}
                  onChange={(e) => setFuelPerLy(Number(e.target.value))}
                />
              </div>

              <div className="flex flex-col gap-1">
                <label className="text-xs text-[#94a3b8]">Fuel Conservation Level</label>
                <Select
                  value={String(fuelConservationLevel)}
                  onValueChange={(v) => setFuelConservationLevel(Number(v))}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Fuel Conservation Level" />
                  </SelectTrigger>
                  <SelectContent>
                    {[0, 1, 2, 3, 4, 5].map((level) => (
                      <SelectItem key={level} value={String(level)}>
                        Level {level} ({level * 10}% reduction)
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </>
          )}

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[#94a3b8]">Route Preference</label>
            <Select
              value={routePreference}
              onValueChange={(v) => setRoutePreference(v)}
            >
              <SelectTrigger>
                <SelectValue placeholder="Route Preference" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="shortest">Shortest</SelectItem>
                <SelectItem value="secure">Secure</SelectItem>
                <SelectItem value="insecure">Insecure</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            <Checkbox
              id="is-default"
              checked={isDefault}
              onCheckedChange={(checked) => setIsDefault(checked === true)}
            />
            <label htmlFor="is-default" className="text-sm text-[#94a3b8] cursor-pointer">
              Default profile for this method
            </label>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onClose(false)} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving || !name}>
            {saving ? "Saving..." : isEdit ? "Update" : "Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

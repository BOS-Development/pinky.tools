import React, { useCallback, useEffect, useRef, useState } from "react";
import { Plus, Trash2, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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

interface AsyncSearchDropdownProps {
  value: StationOption | null;
  onSelect: (option: StationOption) => void;
  placeholder: string;
  options: StationOption[];
  loading: boolean;
  onSearch: (value: string) => void;
  displayValue: string;
  setDisplayValue: (v: string) => void;
}

function StationSearchDropdown({
  value,
  onSelect,
  placeholder,
  options,
  loading,
  onSearch,
  displayValue,
  setDisplayValue,
}: AsyncSearchDropdownProps) {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const v = e.target.value;
    setDisplayValue(v);
    onSearch(v);
    setOpen(true);
  };

  const handleSelect = (opt: StationOption) => {
    onSelect(opt);
    setDisplayValue(opt.name);
    setOpen(false);
  };

  return (
    <div className="relative" ref={containerRef}>
      <div className="relative">
        <Input
          value={displayValue}
          onChange={handleInputChange}
          onFocus={() => { if (options.length > 0) setOpen(true); }}
          placeholder={placeholder}
        />
        {loading && (
          <Loader2 className="absolute right-3 top-2.5 h-4 w-4 animate-spin text-[#64748b]" />
        )}
      </div>
      {open && options.length > 0 && (
        <div className="absolute z-50 w-full mt-1 bg-[#1a1f2e] border border-[rgba(148,163,184,0.15)] rounded-sm shadow-lg max-h-48 overflow-y-auto">
          {options.map((opt) => (
            <button
              key={opt.stationId}
              type="button"
              className="w-full text-left px-3 py-2 hover:bg-[rgba(0,212,255,0.08)]"
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => handleSelect(opt)}
            >
              <p className="text-sm text-[#e2e8f0]">{opt.name}</p>
              <p className="text-xs text-[#94a3b8]">
                {opt.solarSystemName}{" "}
                <span style={{ color: getSecurityColor(opt.security ?? 0) }}>
                  ({(opt.security ?? 0).toFixed(1)})
                </span>
              </p>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

interface ItemTypeSearchDropdownProps {
  value: ItemTypeOption | null;
  onSelect: (option: ItemTypeOption) => void;
  options: ItemTypeOption[];
  loading: boolean;
  onSearch: (value: string) => void;
  displayValue: string;
  setDisplayValue: (v: string) => void;
}

function ItemTypeSearchDropdown({
  value,
  onSelect,
  options,
  loading,
  onSearch,
  displayValue,
  setDisplayValue,
}: ItemTypeSearchDropdownProps) {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const v = e.target.value;
    setDisplayValue(v);
    onSearch(v);
    setOpen(true);
  };

  const handleSelect = (opt: ItemTypeOption) => {
    onSelect(opt);
    setDisplayValue(opt.TypeName);
    setOpen(false);
  };

  return (
    <div className="relative flex-[2]" ref={containerRef}>
      <div className="relative">
        <Input
          value={displayValue}
          onChange={handleInputChange}
          onFocus={() => { if (options.length > 0) setOpen(true); }}
          placeholder="Search for an item..."
        />
        {loading && (
          <Loader2 className="absolute right-3 top-2.5 h-4 w-4 animate-spin text-[#64748b]" />
        )}
      </div>
      {open && options.length > 0 && (
        <div className="absolute z-50 w-full mt-1 bg-[#1a1f2e] border border-[rgba(148,163,184,0.15)] rounded-sm shadow-lg max-h-48 overflow-y-auto">
          {options.map((opt) => (
            <button
              key={opt.TypeID}
              type="button"
              className="w-full text-left px-3 py-2 hover:bg-[rgba(0,212,255,0.08)]"
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => handleSelect(opt)}
            >
              <div className="flex items-center gap-2">
                <img
                  src={`https://images.evetech.net/types/${opt.TypeID}/icon?size=32`}
                  alt=""
                  style={{ width: 24, height: 24 }}
                />
                <div>
                  <p className="text-sm text-[#e2e8f0]">{opt.TypeName}</p>
                  <p className="text-xs text-[#94a3b8]">{opt.Volume.toLocaleString()} m³</p>
                </div>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

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
  const [itemTypeDisplay, setItemTypeDisplay] = useState("");
  const itemTypeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Station search state
  const [originOptions, setOriginOptions] = useState<StationOption[]>([]);
  const [destOptions, setDestOptions] = useState<StationOption[]>([]);
  const [originLoading, setOriginLoading] = useState(false);
  const [destLoading, setDestLoading] = useState(false);
  const [originDisplay, setOriginDisplay] = useState("");
  const [destDisplay, setDestDisplay] = useState("");
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
      setOriginDisplay("");
      setDestDisplay("");
      setItems([]);
      setSelectedItemType(null);
      setItemQuantity("");
      setItemTypeOptions([]);
      setItemTypeDisplay("");
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

  const handleSelectItemType = (opt: ItemTypeOption) => {
    setSelectedItemType(opt);
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
    setItemTypeDisplay("");
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
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose(false); }}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Create Transport Job</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-3 pt-1 max-h-[70vh] overflow-y-auto pr-1">
          <div className="flex flex-col gap-1">
            <label className="text-xs text-[#94a3b8]">Origin Station</label>
            <StationSearchDropdown
              value={originStation}
              onSelect={(opt) => setOriginStation(opt)}
              placeholder="Search for a station..."
              options={originOptions}
              loading={originLoading}
              onSearch={handleOriginSearch}
              displayValue={originDisplay}
              setDisplayValue={setOriginDisplay}
            />
            {originStation && (
              <span className="text-xs text-[#94a3b8]">
                System: {originStation.solarSystemName}
              </span>
            )}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[#94a3b8]">Destination Station</label>
            <StationSearchDropdown
              value={destinationStation}
              onSelect={(opt) => setDestinationStation(opt)}
              placeholder="Search for a station..."
              options={destOptions}
              loading={destLoading}
              onSearch={handleDestSearch}
              displayValue={destDisplay}
              setDisplayValue={setDestDisplay}
            />
            {destinationStation && (
              <span className="text-xs text-[#94a3b8]">
                System: {destinationStation.solarSystemName}
              </span>
            )}
          </div>

          {/* Items Section */}
          <p className="text-sm font-medium text-[#e2e8f0] mt-1">Items to Transport</p>

          <div className="flex gap-2 items-start">
            <ItemTypeSearchDropdown
              value={selectedItemType}
              onSelect={handleSelectItemType}
              options={itemTypeOptions}
              loading={itemTypeLoading}
              onSearch={handleItemTypeSearch}
              displayValue={itemTypeDisplay}
              setDisplayValue={setItemTypeDisplay}
            />
            <div className="flex-1">
              <Input
                placeholder="Quantity"
                value={itemQuantity}
                onChange={(e) => {
                  const raw = e.target.value.replace(/[^0-9]/g, "");
                  if (raw === "") {
                    setItemQuantity("");
                  } else {
                    setItemQuantity(Number(raw).toLocaleString());
                  }
                }}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    handleAddItem();
                  }
                }}
              />
            </div>
            <Button
              variant="ghost"
              size="icon"
              className="text-[#00d4ff] hover:text-[#00d4ff] hover:bg-[rgba(0,212,255,0.1)]"
              onClick={handleAddItem}
              disabled={!selectedItemType || !itemQuantity}
            >
              <Plus className="h-4 w-4" />
            </Button>
          </div>

          {items.length > 0 && (
            <Table>
              <TableHeader>
                <TableRow className="[&>th]:text-[#94a3b8] [&>th]:border-[#1e2231]">
                  <TableHead>Item</TableHead>
                  <TableHead className="text-right">Quantity</TableHead>
                  <TableHead className="text-right">Volume (m³)</TableHead>
                  <TableHead className="text-right w-12" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.itemType.TypeID} className="[&>td]:border-[#1e2231]">
                    <TableCell>
                      <div className="flex gap-2 items-center">
                        <img
                          src={`https://images.evetech.net/types/${item.itemType.TypeID}/icon?size=32`}
                          alt=""
                          style={{ width: 20, height: 20 }}
                        />
                        {item.itemType.TypeName}
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      {formatNumber(item.quantity)}
                    </TableCell>
                    <TableCell className="text-right">
                      {formatNumber(item.itemType.Volume * item.quantity)}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6 text-[#ef4444] hover:text-[#ef4444] hover:bg-[rgba(239,68,68,0.1)]"
                        onClick={() => handleRemoveItem(item.itemType.TypeID)}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
                <TableRow className="[&>td]:border-[#1e2231] font-semibold">
                  <TableCell>Total</TableCell>
                  <TableCell className="text-right">
                    {formatNumber(items.reduce((sum, i) => sum + i.quantity, 0))}
                  </TableCell>
                  <TableCell className="text-right">
                    {formatNumber(totalVolume)} m³
                  </TableCell>
                  <TableCell />
                </TableRow>
              </TableBody>
            </Table>
          )}

          {items.length === 0 && (
            <p className="text-sm text-[#64748b] text-center py-2">
              No items added yet. Search for items above and add them to this job.
            </p>
          )}

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[#94a3b8]">Transport Method</label>
            <Select
              value={transportMethod}
              onValueChange={(v) => {
                setTransportMethod(v);
                setTransportProfileId("");
                setJfRouteId("");
              }}
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
            <label className="text-xs text-[#94a3b8]">Fulfillment Type</label>
            <Select
              value={fulfillmentType}
              onValueChange={(v) => setFulfillmentType(v)}
            >
              <SelectTrigger>
                <SelectValue placeholder="Fulfillment Type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="self_haul">Self Haul</SelectItem>
                <SelectItem value="courier_contract">Courier Contract</SelectItem>
                <SelectItem value="contact_haul">Contact Haul</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {filteredProfiles.length > 0 && (
            <div className="flex flex-col gap-1">
              <label className="text-xs text-[#94a3b8]">Transport Profile</label>
              <Select
                value={transportProfileId}
                onValueChange={(v) => setTransportProfileId(v)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="None" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">None</SelectItem>
                  {filteredProfiles.map((p) => (
                    <SelectItem key={p.id} value={String(p.id)}>
                      {p.name} {p.isDefault ? "(Default)" : ""}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {isJF && jfRoutes.length > 0 && (
            <div className="flex flex-col gap-1">
              <label className="text-xs text-[#94a3b8]">JF Route</label>
              <Select
                value={jfRouteId}
                onValueChange={(v) => setJfRouteId(v)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="None" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">None</SelectItem>
                  {jfRoutes.map((r) => (
                    <SelectItem key={r.id} value={String(r.id)}>
                      {r.name} ({(r.totalDistanceLy ?? 0).toFixed(1)} LY)
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[#94a3b8]">Notes</label>
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={2}
              className="w-full px-3 py-2 text-sm bg-transparent border border-[rgba(148,163,184,0.15)] rounded-sm text-[#e2e8f0] placeholder:text-[#64748b] focus:outline-none focus:ring-1 focus:ring-[#00d4ff] resize-none"
              placeholder="Optional notes..."
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onClose(false)} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving || !canSave}>
            {saving ? "Creating..." : "Create Job"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

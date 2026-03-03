import { useState, useCallback, useRef, useEffect } from 'react';
import { Loader2 } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Switch } from '@/components/ui/switch';
import { Separator } from '@/components/ui/separator';
import { Asset, StockpileMarker, EveInventoryType } from "@industry-tool/client/data/models";

type Owner = {
  ownerType: string;
  ownerId: number;
  ownerName: string;
};

type Props = {
  open: boolean;
  onClose: () => void;
  onSaved: (asset: Asset) => void;
  locationId: number;
  containerId?: number;
  divisionNumber?: number;
  owners: Owner[];
};

type AvailablePlan = {
  id: number;
  name: string;
  productName?: string;
};

export default function AddStockpileDialog({ open, onClose, onSaved, locationId, containerId, divisionNumber, owners }: Props) {
  const [selectedItem, setSelectedItem] = useState<EveInventoryType | null>(null);
  const [itemOptions, setItemOptions] = useState<EveInventoryType[]>([]);
  const [itemLoading, setItemLoading] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const [selectedOwner, setSelectedOwner] = useState('');
  const [desiredQuantity, setDesiredQuantity] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [autoProductionEnabled, setAutoProductionEnabled] = useState(false);
  const [selectedPlanId, setSelectedPlanId] = useState<number | null>(null);
  const [availablePlans, setAvailablePlans] = useState<AvailablePlan[]>([]);
  const [plansLoading, setPlansLoading] = useState(false);
  const [parallelism, setParallelism] = useState(0);

  // Auto-select single owner
  const effectiveOwner = owners.length === 1
    ? `${owners[0].ownerType}:${owners[0].ownerId}`
    : selectedOwner;

  const parseOwnerKey = (key: string) => {
    const [ownerType, ownerId] = key.split(':');
    return { ownerType, ownerId: parseInt(ownerId, 10) };
  };

  // Close dropdown on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const searchItems = useCallback((query: string) => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    if (!query || query.length < 2) {
      setItemOptions([]);
      return;
    }

    debounceRef.current = setTimeout(async () => {
      setItemLoading(true);
      try {
        const res = await fetch(`/api/item-types/search?q=${encodeURIComponent(query)}`);
        if (res.ok) {
          const data: EveInventoryType[] = await res.json();
          setItemOptions(data);
        }
      } finally {
        setItemLoading(false);
      }
    }, 300);
  }, []);

  useEffect(() => {
    if (!selectedItem) {
      setAvailablePlans([]);
      setSelectedPlanId(null);
      return;
    }
    const fetchPlans = async () => {
      setPlansLoading(true);
      try {
        const res = await fetch(`/api/industry/plans/by-product/${selectedItem.TypeID}`);
        if (res.ok) {
          const data: AvailablePlan[] = await res.json();
          setAvailablePlans(data || []);
        }
      } finally {
        setPlansLoading(false);
      }
    };
    fetchPlans();
  }, [selectedItem]);

  const handleSave = async () => {
    if (!selectedItem || !effectiveOwner || !desiredQuantity) return;

    const qty = parseInt(desiredQuantity.replace(/,/g, ''), 10);
    if (qty <= 0 || isNaN(qty)) return;

    const { ownerType, ownerId } = parseOwnerKey(effectiveOwner);
    const ownerInfo = owners.find(o => o.ownerType === ownerType && o.ownerId === ownerId);

    setSaving(true);
    setError(null);

    try {
      const marker: StockpileMarker = {
        userId: 0,
        typeId: selectedItem.TypeID,
        ownerType,
        ownerId,
        locationId,
        containerId,
        divisionNumber,
        desiredQuantity: qty,
        autoProductionEnabled,
        planId: autoProductionEnabled && selectedPlanId ? selectedPlanId : undefined,
        autoProductionParallelism: autoProductionEnabled ? parallelism : undefined,
      };

      const res = await fetch('/api/stockpiles/upsert', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(marker),
      });

      if (!res.ok) {
        throw new Error('Failed to save stockpile marker');
      }

      const delta = -qty;
      const phantomAsset: Asset = {
        name: selectedItem.TypeName,
        typeId: selectedItem.TypeID,
        quantity: 0,
        volume: 0,
        ownerType,
        ownerName: ownerInfo?.ownerName || '',
        ownerId,
        desiredQuantity: qty,
        stockpileDelta: delta,
      };

      onSaved(phantomAsset);
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  };

  const handleClose = () => {
    setSelectedItem(null);
    setItemOptions([]);
    setInputValue('');
    setDropdownOpen(false);
    setSelectedOwner('');
    setDesiredQuantity('');
    setError(null);
    setAutoProductionEnabled(false);
    setSelectedPlanId(null);
    setAvailablePlans([]);
    setParallelism(0);
    onClose();
  };

  const canSave = selectedItem && effectiveOwner && desiredQuantity && parseInt(desiredQuantity.replace(/,/g, ''), 10) > 0;

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) handleClose(); }}>
      <DialogContent className="sm:max-w-md bg-[#12151f] border border-[rgba(148,163,184,0.15)] text-[#e2e8f0]">
        <DialogHeader>
          <DialogTitle className="text-[#e2e8f0]">Add Stockpile Marker</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-4 pt-1">
          {/* Async item search */}
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="item-type-search" className="text-xs text-[#94a3b8]">Item Type</Label>
            <div className="relative" ref={dropdownRef}>
              <Input
                id="item-type-search"
                value={selectedItem ? selectedItem.TypeName : inputValue}
                onChange={(e) => {
                  setInputValue(e.target.value);
                  setSelectedItem(null);
                  searchItems(e.target.value);
                  setDropdownOpen(true);
                }}
                onFocus={() => { if (itemOptions.length > 0) setDropdownOpen(true); }}
                placeholder="Search for an item..."
                className="text-sm bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0] placeholder:text-[#64748b]"
              />
              {itemLoading && (
                <Loader2 className="absolute right-3 top-2.5 h-4 w-4 animate-spin text-[#64748b]" />
              )}
              {dropdownOpen && itemOptions.length > 0 && (
                <div className="absolute z-50 w-full mt-1 bg-[#1a1f2e] border border-[rgba(148,163,184,0.15)] rounded-sm shadow-lg max-h-48 overflow-y-auto">
                  {itemOptions.map((opt) => (
                    <button
                      key={opt.TypeID}
                      className="w-full text-left px-3 py-2 text-sm hover:bg-[rgba(0,212,255,0.08)] text-[#e2e8f0] flex items-center gap-2"
                      onMouseDown={() => {
                        setSelectedItem(opt);
                        setInputValue(opt.TypeName);
                        setDropdownOpen(false);
                      }}
                    >
                      <img
                        src={`https://images.evetech.net/types/${opt.TypeID}/icon?size=32`}
                        alt=""
                        className="w-5 h-5 rounded-sm"
                      />
                      {opt.TypeName}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Owner selector — only shown when multiple owners */}
          {owners.length > 1 && (
            <div className="flex flex-col gap-1.5">
              <Label className="text-xs text-[#94a3b8]">Owner</Label>
              <Select value={selectedOwner} onValueChange={setSelectedOwner}>
                <SelectTrigger className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]">
                  <SelectValue placeholder="Select owner..." />
                </SelectTrigger>
                <SelectContent className="bg-[#1a1f2e] border-[rgba(148,163,184,0.15)]">
                  {owners.map(o => (
                    <SelectItem
                      key={`${o.ownerType}:${o.ownerId}`}
                      value={`${o.ownerType}:${o.ownerId}`}
                      className="text-[#e2e8f0] focus:bg-[rgba(0,212,255,0.08)]"
                    >
                      {o.ownerName} ({o.ownerType})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {owners.length === 1 && (
            <p className="text-sm text-[#94a3b8]">
              Owner: {owners[0].ownerName} ({owners[0].ownerType})
            </p>
          )}

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="desired-qty" className="text-xs text-[#94a3b8]">Desired Quantity</Label>
            <Input
              id="desired-qty"
              type="number"
              min={1}
              value={desiredQuantity}
              onChange={(e) => setDesiredQuantity(e.target.value)}
              className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]"
            />
          </div>

          <Separator className="bg-[rgba(148,163,184,0.1)]" />

          <p className="text-xs font-semibold text-[#94a3b8]">Auto-Production</p>

          <div className="flex items-center gap-2">
            <Switch
              id="auto-production"
              checked={autoProductionEnabled}
              onCheckedChange={setAutoProductionEnabled}
            />
            <Label htmlFor="auto-production" className="text-sm text-[#e2e8f0] cursor-pointer">
              Enable Auto-Production
            </Label>
          </div>

          {autoProductionEnabled && (
            <>
              <div className="flex flex-col gap-1.5">
                <Label className="text-xs text-[#94a3b8]">Production Plan</Label>
                <Select
                  value={selectedPlanId !== null ? String(selectedPlanId) : ''}
                  onValueChange={(v) => setSelectedPlanId(v ? Number(v) : null)}
                  disabled={plansLoading || availablePlans.length === 0}
                >
                  <SelectTrigger className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]">
                    <SelectValue
                      placeholder={
                        plansLoading
                          ? 'Loading plans...'
                          : availablePlans.length === 0
                          ? 'No plans for this item'
                          : 'Select a plan...'
                      }
                    />
                  </SelectTrigger>
                  <SelectContent className="bg-[#1a1f2e] border-[rgba(148,163,184,0.15)]">
                    {availablePlans.map((plan) => (
                      <SelectItem
                        key={plan.id}
                        value={String(plan.id)}
                        className="text-[#e2e8f0] focus:bg-[rgba(0,212,255,0.08)]"
                      >
                        {plan.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="flex flex-col gap-1.5">
                <Label htmlFor="parallelism" className="text-xs text-[#94a3b8]">Max Parallelism</Label>
                <Input
                  id="parallelism"
                  type="number"
                  min={0}
                  value={parallelism}
                  onChange={(e) => setParallelism(Math.max(0, parseInt(e.target.value, 10) || 0))}
                  className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]"
                />
                <p className="text-xs text-[#64748b]">0 = no character assignment</p>
              </div>
            </>
          )}

          {error && (
            <Alert variant="destructive" className="border-[#ef4444] bg-[rgba(239,68,68,0.1)]">
              <AlertDescription className="text-[#ef4444]">{error}</AlertDescription>
            </Alert>
          )}
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={handleClose} disabled={saving}
            className="border-[rgba(148,163,184,0.2)] text-[#e2e8f0] hover:bg-[rgba(148,163,184,0.1)]">
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={saving || !canSave}
            className="bg-[#3b82f6] hover:bg-[#2563eb] text-white"
          >
            {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {saving ? 'Saving...' : 'Add Stockpile'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

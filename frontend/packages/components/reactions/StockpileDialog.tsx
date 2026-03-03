import { useState } from 'react';
import { Loader2 } from 'lucide-react';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
} from "@/components/ui/select";
import {
  Table, TableHeader, TableBody, TableRow, TableHead, TableCell,
} from "@/components/ui/table";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { ShoppingItem, StockpileMarker } from "@industry-tool/client/data/models";
import { formatNumber } from "@industry-tool/utils/formatting";
import { AssetOwner } from "@industry-tool/utils/assetAggregation";

type Props = {
  open: boolean;
  onClose: () => void;
  shoppingList: ShoppingItem[];
  locationId: number;
  locationName: string;
  owners: AssetOwner[];
};

export default function StockpileDialog({ open, onClose, shoppingList, locationId, locationName, owners }: Props) {
  const [selectedOwner, setSelectedOwner] = useState('');
  const [quantities, setQuantities] = useState<Record<number, string>>(() => {
    const initial: Record<number, string> = {};
    for (const item of shoppingList) {
      initial[item.type_id] = item.quantity.toString();
    }
    return initial;
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const ownerKey = (o: AssetOwner) => `${o.ownerType}:${o.ownerId}`;
  const parseOwnerKey = (key: string) => {
    const [ownerType, ownerId] = key.split(':');
    return { ownerType, ownerId: parseInt(ownerId, 10) };
  };

  const handleQuantityChange = (typeId: number, value: string) => {
    setQuantities(prev => ({ ...prev, [typeId]: value }));
  };

  const handleSave = async () => {
    if (!selectedOwner) return;

    const { ownerType, ownerId } = parseOwnerKey(selectedOwner);
    setSaving(true);
    setError(null);

    try {
      for (const item of shoppingList) {
        const qty = parseInt(quantities[item.type_id]?.replace(/,/g, '') || '0', 10);
        if (qty <= 0) continue;

        const marker: StockpileMarker = {
          userId: 0,
          typeId: item.type_id,
          ownerType,
          ownerId,
          locationId,
          desiredQuantity: qty,
        };

        const res = await fetch('/api/stockpiles/upsert', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(marker),
        });

        if (!res.ok) {
          throw new Error(`Failed to set stockpile for ${item.name}`);
        }
      }

      setSuccess(true);
      setTimeout(() => {
        onClose();
        setSuccess(false);
      }, 1000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save stockpile targets');
    } finally {
      setSaving(false);
    }
  };

  const itemsWithQty = shoppingList.filter(item => {
    const qty = parseInt(quantities[item.type_id]?.replace(/,/g, '') || '0', 10);
    return qty > 0;
  });

  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose(); }}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Set Stockpile Targets</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-3 pt-1">
          <p className="text-sm text-[var(--color-text-secondary)]">
            Set desired stockpile quantities at <strong>{locationName}</strong>. Items with quantity 0 will be skipped.
          </p>

          <div className="flex flex-col gap-1">
            <label className="text-xs text-[var(--color-text-secondary)]">Owner</label>
            <Select value={selectedOwner} onValueChange={setSelectedOwner}>
              <SelectTrigger>
                <SelectValue placeholder="Select owner..." />
              </SelectTrigger>
              <SelectContent>
                {owners.map(o => (
                  <SelectItem key={ownerKey(o)} value={ownerKey(o)}>
                    {o.ownerName} ({o.ownerType})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                const cleared: Record<number, string> = {};
                for (const item of shoppingList) cleared[item.type_id] = '0';
                setQuantities(cleared);
              }}
            >
              Clear All
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                const reset: Record<number, string> = {};
                for (const item of shoppingList) reset[item.type_id] = item.quantity.toString();
                setQuantities(reset);
              }}
            >
              Reset All
            </Button>
          </div>

          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          {success && (
            <Alert>
              <AlertDescription className="text-[var(--color-success-teal)]">
                Stockpile targets saved!
              </AlertDescription>
            </Alert>
          )}

          <div className="max-h-[400px] overflow-y-auto">
            <Table>
              <TableHeader>
                <TableRow className="bg-[var(--color-bg-panel)] sticky top-0">
                  <TableHead>Material</TableHead>
                  <TableHead className="text-right">Shopping List</TableHead>
                  <TableHead className="text-right">Desired Quantity</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {shoppingList.map((item) => (
                  <TableRow key={item.type_id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <img
                          src={`https://images.evetech.net/types/${item.type_id}/icon?size=32`}
                          alt=""
                          width={24}
                          height={24}
                          className="rounded-sm"
                        />
                        {item.name}
                      </div>
                    </TableCell>
                    <TableCell className="text-right">{formatNumber(item.quantity)}</TableCell>
                    <TableCell className="text-right w-40">
                      <Input
                        type="number"
                        value={quantities[item.type_id] || ''}
                        onChange={(e) => handleQuantityChange(item.type_id, e.target.value)}
                        min={0}
                        className="w-36 h-7 text-right"
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>

        <DialogFooter className="flex items-center">
          <span className="text-sm text-[var(--color-text-secondary)] mr-auto">
            {itemsWithQty.length} of {shoppingList.length} items will be set
          </span>
          <Button variant="outline" onClick={onClose} disabled={saving}>Cancel</Button>
          <Button
            onClick={handleSave}
            disabled={saving || !selectedOwner}
          >
            {saving ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin mr-1" />
                Saving...
              </>
            ) : (
              'Set Targets'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

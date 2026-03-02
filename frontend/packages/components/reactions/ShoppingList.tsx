import { useState } from 'react';
import { Loader2, Copy, CheckCircle2, Package } from 'lucide-react';
import { ChevronDown, ChevronUp, ArrowUpDown } from 'lucide-react';
import {
  Table, TableHeader, TableBody, TableRow, TableHead, TableCell,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Combobox, type ComboboxOption } from "@/components/ui/combobox";
import { cn } from "@/lib/utils";
import { PlanResponse, AssetsResponse } from "@industry-tool/client/data/models";
import { formatISK, formatNumber } from "@industry-tool/utils/formatting";
import { aggregateAssetsByTypeId, getUniqueOwners } from "@industry-tool/utils/assetAggregation";
import StockpileDialog from './StockpileDialog';

type Props = {
  planData: PlanResponse | null;
  loading: boolean;
  assets: AssetsResponse | null;
  isAuthenticated: boolean;
  stockpileLocationId: number;
  onStockpileLocationChange: (locationId: number) => void;
};

type SortKey = 'name' | 'quantity' | 'inStock' | 'delta' | 'price' | 'cost' | 'volume';
type SortDir = 'asc' | 'desc';

function SortHeader({ label, sortKey: key, activeSortKey, sortDir, onSort, align }: {
  label: string;
  sortKey: SortKey;
  activeSortKey: SortKey;
  sortDir: SortDir;
  onSort: (key: SortKey) => void;
  align?: 'left' | 'right';
}) {
  const active = activeSortKey === key;
  return (
    <TableHead className={align === 'right' ? 'text-right' : ''}>
      <button
        className={cn(
          "inline-flex items-center gap-1 text-xs font-medium uppercase tracking-wider cursor-pointer hover:text-[var(--color-primary-cyan)] transition-colors",
          active && "text-[var(--color-primary-cyan)]"
        )}
        onClick={() => onSort(key)}
      >
        {label}
        {active ? (
          sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />
        ) : (
          <ArrowUpDown className="h-3 w-3 opacity-40" />
        )}
      </button>
    </TableHead>
  );
}

export default function ShoppingList({ planData, loading, assets, isAuthenticated, stockpileLocationId, onStockpileLocationChange }: Props) {
  const [stockpileDialogOpen, setStockpileDialogOpen] = useState(false);
  const [sortKey, setSortKey] = useState<SortKey>('name');
  const [sortDir, setSortDir] = useState<SortDir>('asc');

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir(prev => prev === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortDir(key === 'name' ? 'asc' : 'desc');
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center py-16">
        <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary-cyan)]" />
      </div>
    );
  }

  if (!planData || planData.shopping_list.length === 0) {
    return (
      <p className="py-8 text-center text-[var(--color-text-secondary)]">
        Select reactions in the Pick Reactions tab to generate a shopping list.
      </p>
    );
  }

  const locations = assets?.structures ?? [];
  const selectedStructure = locations.find(s => s.id === stockpileLocationId) || null;
  const stockpileMap = selectedStructure ? aggregateAssetsByTypeId(selectedStructure) : null;

  const totalCost = planData.shopping_list.reduce((sum, item) => sum + item.cost, 0);
  const totalVolume = planData.shopping_list.reduce((sum, item) => sum + item.volume, 0);

  const deltaCost = stockpileMap
    ? planData.shopping_list.reduce((sum, item) => {
        const have = stockpileMap.get(item.type_id) || 0;
        const delta = Math.max(0, item.quantity - have);
        return sum + (delta * item.price);
      }, 0)
    : null;

  const copyMultibuy = () => {
    const text = planData.shopping_list
      .map(item => `${item.name} ${item.quantity}`)
      .join('\n');
    navigator.clipboard.writeText(text);
  };

  const copyMultibuyDelta = () => {
    if (!stockpileMap) return;
    const lines = planData.shopping_list
      .map(item => {
        const have = stockpileMap.get(item.type_id) || 0;
        const delta = Math.max(0, item.quantity - have);
        return delta > 0 ? `${item.name} ${delta}` : null;
      })
      .filter(Boolean);
    navigator.clipboard.writeText(lines.join('\n'));
  };

  const locationOptions: ComboboxOption[] = locations.map(s => ({
    value: s.id.toString(),
    label: s.name,
  }));

  return (
    <div>
      <div className="flex justify-between items-center mb-3 flex-wrap gap-2">
        <div className="flex flex-col gap-1">
          <span className="text-sm text-[var(--color-text-secondary)]">
            {planData.shopping_list.length} items | Total: {formatISK(totalCost)} | Volume: {formatNumber(totalVolume, 1)} m3
          </span>
          {deltaCost !== null && (
            <span className="text-sm text-[var(--color-success-teal)]">
              Delta Cost: {formatISK(deltaCost)}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {isAuthenticated && (
            <Combobox
              options={locationOptions}
              value={stockpileLocationId ? stockpileLocationId.toString() : ''}
              onValueChange={(val) => onStockpileLocationChange(val ? parseInt(val) : 0)}
              placeholder="Stockpile Location"
              searchPlaceholder="Search locations..."
              triggerClassName="w-64"
            />
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={copyMultibuy}
          >
            <Copy className="h-4 w-4 mr-1" />
            Copy Multibuy
          </Button>
          {stockpileMap && (
            <Button
              variant="outline"
              size="sm"
              onClick={copyMultibuyDelta}
              className="border-[var(--color-success-teal)] text-[var(--color-success-teal)] hover:border-[var(--color-success-teal)]"
            >
              <Copy className="h-4 w-4 mr-1" />
              Copy Delta
            </Button>
          )}
          {selectedStructure && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setStockpileDialogOpen(true)}
              className="border-[var(--color-manufacturing-amber)] text-[var(--color-manufacturing-amber)] hover:border-[var(--color-manufacturing-amber)]"
            >
              <Package className="h-4 w-4 mr-1" />
              Set Stockpile
            </Button>
          )}
        </div>
      </div>

      <Table>
        <TableHeader>
          <TableRow className="bg-[var(--color-bg-panel)]">
            <SortHeader label="Material" sortKey="name" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} />
            <SortHeader label="Quantity" sortKey="quantity" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} align="right" />
            {stockpileMap && (
              <SortHeader label="In Stock" sortKey="inStock" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} align="right" />
            )}
            {stockpileMap && (
              <SortHeader label="Delta" sortKey="delta" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} align="right" />
            )}
            <SortHeader label="Unit Price" sortKey="price" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} align="right" />
            <SortHeader label="Total Cost" sortKey="cost" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} align="right" />
            <SortHeader label="Volume (m3)" sortKey="volume" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} align="right" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {[...planData.shopping_list]
            .map((item) => {
              const have = stockpileMap ? (stockpileMap.get(item.type_id) || 0) : 0;
              const delta = Math.max(0, item.quantity - have);
              return { item, have, delta };
            })
            .sort((a, b) => {
              let cmp = 0;
              switch (sortKey) {
                case 'name': cmp = a.item.name.localeCompare(b.item.name); break;
                case 'quantity': cmp = a.item.quantity - b.item.quantity; break;
                case 'inStock': cmp = a.have - b.have; break;
                case 'delta': cmp = a.delta - b.delta; break;
                case 'price': cmp = a.item.price - b.item.price; break;
                case 'cost': cmp = a.item.cost - b.item.cost; break;
                case 'volume': cmp = a.item.volume - b.item.volume; break;
              }
              return sortDir === 'asc' ? cmp : -cmp;
            })
            .map(({ item, have, delta }) => {
            const fulfilled = stockpileMap ? delta === 0 : false;

            return (
              <TableRow
                key={item.type_id}
                className="odd:bg-white/[0.02]"
              >
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
                {stockpileMap && (
                  <TableCell className="text-right">{formatNumber(have)}</TableCell>
                )}
                {stockpileMap && (
                  <TableCell className={cn("text-right", fulfilled && "text-[#10b981]")}>
                    {fulfilled ? (
                      <span className="inline-flex items-center justify-end gap-1">
                        <CheckCircle2 className="h-4 w-4" />
                        0
                      </span>
                    ) : (
                      formatNumber(delta)
                    )}
                  </TableCell>
                )}
                <TableCell className="text-right">{formatISK(item.price)}</TableCell>
                <TableCell className="text-right">{formatISK(item.cost)}</TableCell>
                <TableCell className="text-right">{formatNumber(item.volume, 1)}</TableCell>
              </TableRow>
            );
          })}
          <TableRow className="[&_td]:font-bold [&_td]:border-t-2 [&_td]:border-white/10">
            <TableCell>Total</TableCell>
            <TableCell />
            {stockpileMap && <TableCell />}
            {stockpileMap && <TableCell />}
            <TableCell />
            <TableCell className="text-right">{formatISK(totalCost)}</TableCell>
            <TableCell className="text-right">{formatNumber(totalVolume, 1)}</TableCell>
          </TableRow>
        </TableBody>
      </Table>

      {selectedStructure && (
        <StockpileDialog
          open={stockpileDialogOpen}
          onClose={() => setStockpileDialogOpen(false)}
          shoppingList={planData.shopping_list}
          locationId={selectedStructure.id}
          locationName={selectedStructure.name}
          owners={getUniqueOwners(selectedStructure)}
        />
      )}
    </div>
  );
}

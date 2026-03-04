import { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import { SupplyChainResponse, SupplyChainItem, SupplyChainPlanetEntry, StockpileMarker } from '@industry-tool/client/data/models';
import { formatNumber, formatQuantity } from '@industry-tool/utils/formatting';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { cn } from '@/lib/utils';
import { ChevronDown, ChevronRight } from 'lucide-react';

const TIERS = ['All', 'R0', 'P1', 'P2', 'P3', 'P4'];
const TIER_ORDER: Record<string, number> = { R0: 0, P1: 1, P2: 2, P3: 3, P4: 4 };

const RESIZE_PRESETS = [
  { label: '1 week', hours: 168 },
  { label: '2 weeks', hours: 336 },
  { label: '1 month', hours: 720 },
];

const SOURCE_COLORS: Record<string, { bg: string; text: string }> = {
  extracted: { bg: 'var(--color-success-tint)', text: 'var(--color-success-teal)' },
  produced:  { bg: 'var(--color-info-tint)', text: 'var(--color-primary-cyan)' },
  bought:    { bg: 'var(--color-warning-tint)', text: 'var(--color-manufacturing-amber)' },
  mixed:     { bg: 'var(--color-neutral-tint)', text: 'var(--color-text-secondary)' },
};

function formatDepletion(hours: number): string {
  if (hours <= 0) return '\u2014';
  const days = Math.floor(hours / 24);
  const h = Math.floor(hours % 24);
  if (days > 0) return `${days}d ${h}h`;
  const m = Math.round((hours % 1) * 60);
  return `${h}h ${m}m`;
}

function depletionColor(hours: number): string {
  if (hours <= 0) return 'var(--color-text-muted)';
  if (hours < 24) return 'var(--color-danger-rose)';
  if (hours < 72) return 'var(--color-manufacturing-amber)';
  return 'var(--color-success-teal)';
}

function netColor(value: number): string {
  if (value > 0.01) return 'var(--color-success-teal)';
  if (value < -0.01) return 'var(--color-danger-rose)';
  return 'var(--color-text-muted)';
}

function formatRate(value: number): string {
  if (value === 0) return '\u2014';
  return formatNumber(Math.round(value));
}

const headerCellCls = "text-text-muted border-overlay-subtle bg-background-void text-[0.7rem] uppercase tracking-wide font-semibold py-2";

function SupplyChainRow({ item, expanded, onToggle, onResize }: {
  item: SupplyChainItem;
  expanded: boolean;
  onToggle: () => void;
  onResize: (item: SupplyChainItem, newQty: number) => void;
}) {
  const sourceStyle = SOURCE_COLORS[item.source] || SOURCE_COLORS.mixed;
  const hasChildren = (item.producers?.length > 0) || (item.consumers?.length > 0);
  const isDeficit = item.netPerHour < -0.01;
  const hasMarkers = (item.stockpileMarkers?.length ?? 0) > 0;
  const canResize = hasMarkers && item.consumedPerHour > 0;

  const [popoverOpen, setPopoverOpen] = useState(false);

  const handlePresetClick = (hours: number) => {
    const newQty = Math.ceil(item.consumedPerHour * hours);
    onResize(item, newQty);
    setPopoverOpen(false);
  };

  return (
    <>
      <TableRow
        className={cn(
          hasChildren ? 'cursor-pointer' : 'cursor-default',
          'hover:bg-interactive-hover',
          expanded
            ? 'bg-interactive-hover'
            : isDeficit
            ? 'bg-rose-danger/5'
            : ''
        )}
        onClick={hasChildren ? onToggle : undefined}
      >
        <TableCell className="border-overlay-subtle w-9 p-1">
          {hasChildren && (
            <button className="text-text-muted p-0.5 leading-none">
              {expanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
            </button>
          )}
        </TableCell>
        <TableCell className="border-overlay-subtle">
          <div className="flex items-center gap-2">
            <img
              src={`https://images.evetech.net/types/${item.typeId}/icon?size=32`}
              alt="" width={24} height={24} className="flex-shrink-0"
            />
            <div className="min-w-0">
              <p className="text-sm text-text-emphasis font-medium leading-tight truncate">
                {item.name || `Type ${item.typeId}`}
              </p>
              <span className="text-text-muted" style={{ fontSize: '0.65rem' }}>
                {item.tierName}
              </span>
            </div>
          </div>
        </TableCell>
        <TableCell className="border-overlay-subtle">
          <span
            className="inline-flex items-center px-2 py-0.5 rounded text-[0.65rem] font-medium"
            style={{ backgroundColor: sourceStyle.bg, color: sourceStyle.text }}
          >
            {item.source.charAt(0).toUpperCase() + item.source.slice(1)}
          </span>
        </TableCell>
        <TableCell
          className="text-right border-overlay-subtle"
          style={{ color: item.producedPerHour > 0 ? 'var(--color-success-teal)' : 'var(--color-text-muted)' }}
        >
          <span className="text-xs">{formatRate(item.producedPerHour)}</span>
        </TableCell>
        <TableCell
          className="text-right border-overlay-subtle"
          style={{ color: item.consumedPerHour > 0 ? 'var(--color-danger-rose)' : 'var(--color-text-muted)' }}
        >
          <span className="text-xs">{formatRate(item.consumedPerHour)}</span>
        </TableCell>
        <TableCell className="text-right border-overlay-subtle">
          <span className="text-xs font-semibold" style={{ color: netColor(item.netPerHour) }}>
            {item.netPerHour > 0.01 ? '+' : ''}{formatRate(item.netPerHour)}
          </span>
        </TableCell>
        <TableCell
          className="text-right border-overlay-subtle"
          onClick={(e) => { if (canResize) e.stopPropagation(); }}
        >
          {canResize ? (
            <Popover open={popoverOpen} onOpenChange={setPopoverOpen}>
              <PopoverTrigger asChild>
                <span
                  className="text-xs text-text-secondary cursor-pointer border-b border-dashed border-current hover:text-primary"
                  onClick={(e) => e.stopPropagation()}
                >
                  {item.stockpileQty > 0 ? formatQuantity(item.stockpileQty) : '\u2014'}
                </span>
              </PopoverTrigger>
              <PopoverContent
                className="p-3 w-48 bg-background-elevated border-overlay-medium"
                align="end"
                onClick={(e) => e.stopPropagation()}
              >
                <span className="block text-[0.6rem] text-text-muted font-semibold uppercase tracking-wide mb-2">
                  Set stockpile for
                </span>
                <div className="flex flex-col gap-1">
                  {RESIZE_PRESETS.map(preset => {
                    const qty = Math.ceil(item.consumedPerHour * preset.hours);
                    return (
                      <button
                        key={preset.label}
                        onClick={(e) => { e.stopPropagation(); handlePresetClick(preset.hours); }}
                        className="flex justify-between items-center text-text-emphasis text-xs py-1 px-2 rounded hover:bg-interactive-selected w-full text-left"
                      >
                        <span>{preset.label}</span>
                        <span className="text-primary font-semibold ml-4">{formatQuantity(qty)}</span>
                      </button>
                    );
                  })}
                </div>
              </PopoverContent>
            </Popover>
          ) : (
            <span className="text-xs text-text-secondary">
              {item.stockpileQty > 0 ? formatQuantity(item.stockpileQty) : '\u2014'}
            </span>
          )}
        </TableCell>
        <TableCell className="text-right border-overlay-subtle">
          <span
            className="text-xs"
            style={{
              color: depletionColor(item.depletionHours),
              fontWeight: item.depletionHours > 0 && item.depletionHours < 72 ? 600 : 400,
            }}
          >
            {formatDepletion(item.depletionHours)}
          </span>
        </TableCell>
      </TableRow>
      {expanded && (
        <>
          {item.producers?.length > 0 && (
            <>
              <TableRow>
                <TableCell className="border-overlay-subtle py-1" />
                <TableCell colSpan={7} className="border-overlay-subtle py-1">
                  <span className="text-text-muted font-semibold uppercase tracking-wide" style={{ fontSize: '0.6rem' }}>
                    Producers
                  </span>
                </TableCell>
              </TableRow>
              {item.producers.map((p, i) => (
                <PlanetEntryRow key={`prod-${i}`} entry={p} type="producer" />
              ))}
            </>
          )}
          {item.consumers?.length > 0 && (
            <>
              <TableRow>
                <TableCell className="border-overlay-subtle py-1" />
                <TableCell colSpan={7} className="border-overlay-subtle py-1">
                  <span className="text-text-muted font-semibold uppercase tracking-wide" style={{ fontSize: '0.6rem' }}>
                    Consumers
                  </span>
                </TableCell>
              </TableRow>
              {item.consumers.map((c, i) => (
                <PlanetEntryRow key={`cons-${i}`} entry={c} type="consumer" />
              ))}
            </>
          )}
        </>
      )}
    </>
  );
}

function PlanetEntryRow({ entry, type }: { entry: SupplyChainPlanetEntry; type: 'producer' | 'consumer' }) {
  const rateColor = type === 'producer' ? 'var(--color-success-teal)' : 'var(--color-danger-rose)';

  return (
    <TableRow className="bg-background-void/40">
      <TableCell className="border-overlay-subtle" />
      <TableCell colSpan={3} className="border-overlay-subtle pl-12">
        <span className="text-xs text-text-primary">{entry.solarSystemName}</span>
        <span className="text-xs text-text-muted ml-1">
          {entry.planetType} &middot; {entry.characterName}
        </span>
      </TableCell>
      <TableCell colSpan={4} className="text-right border-overlay-subtle">
        <span className="text-xs" style={{ color: rateColor }}>
          {formatNumber(Math.round(entry.ratePerHour))}/hr
        </span>
      </TableCell>
    </TableRow>
  );
}

export default function SupplyChain() {
  const { data: session } = useSession();
  const [data, setData] = useState<SupplyChainResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [tierFilter, setTierFilter] = useState('All');
  const [search, setSearch] = useState('');
  const [expandedItems, setExpandedItems] = useState<Set<number>>(new Set());
  const [bulkPopoverOpen, setBulkPopoverOpen] = useState(false);
  const hasFetchedRef = useRef(false);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchData();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session]);

  const fetchData = async () => {
    if (!session) return;
    setLoading(true);
    try {
      const response = await fetch('/api/pi/supply-chain');
      if (response.ok) {
        const result: SupplyChainResponse = await response.json();
        setData(result);
      }
    } finally {
      setLoading(false);
    }
  };

  const upsertMarkers = async (markers: StockpileMarker[], newQty: number) => {
    const totalOld = markers.reduce((sum, m) => sum + m.desiredQuantity, 0);
    const updates = markers.map((m, i) => {
      if (markers.length === 1) return { ...m, desiredQuantity: newQty };
      if (i === markers.length - 1) {
        const allocated = markers.slice(0, -1).reduce((sum, mk) => {
          return sum + Math.round(newQty * (mk.desiredQuantity / totalOld));
        }, 0);
        return { ...m, desiredQuantity: newQty - allocated };
      }
      return { ...m, desiredQuantity: Math.round(newQty * (m.desiredQuantity / totalOld)) };
    });
    for (const marker of updates) {
      await fetch('/api/stockpiles/upsert', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(marker),
      });
    }
  };

  const handleResize = useCallback(async (item: SupplyChainItem, newQty: number) => {
    const markers = item.stockpileMarkers;
    if (!markers || markers.length === 0) return;
    await upsertMarkers(markers, newQty);
    hasFetchedRef.current = false;
    await fetchData();
    hasFetchedRef.current = true;
  }, []);

  const handleResizeAll = useCallback(async (hours: number) => {
    if (!data?.items) return;
    const resizable = data.items.filter(
      item => (item.stockpileMarkers?.length ?? 0) > 0 && item.consumedPerHour > 0
    );
    if (resizable.length === 0) return;

    setLoading(true);
    for (const item of resizable) {
      const newQty = Math.ceil(item.consumedPerHour * hours);
      await upsertMarkers(item.stockpileMarkers!, newQty);
    }
    hasFetchedRef.current = false;
    await fetchData();
    hasFetchedRef.current = true;
  }, [data]);

  const toggleItem = (typeId: number) => {
    setExpandedItems(prev => {
      const next = new Set(prev);
      if (next.has(typeId)) next.delete(typeId);
      else next.add(typeId);
      return next;
    });
  };

  const filteredItems = useMemo(() => {
    if (!data?.items) return [];
    let items = data.items;

    if (tierFilter !== 'All') {
      items = items.filter(item => item.tierName === tierFilter);
    }

    if (search.trim()) {
      const q = search.trim().toLowerCase();
      items = items.filter(item => item.name?.toLowerCase().includes(q));
    }

    // Sort: tier ascending, then deficit first (net ascending), then name
    return [...items].sort((a, b) => {
      const tierDiff = (TIER_ORDER[a.tierName] ?? 99) - (TIER_ORDER[b.tierName] ?? 99);
      if (tierDiff !== 0) return tierDiff;
      if (a.netPerHour !== b.netPerHour) return a.netPerHour - b.netPerHour;
      return (a.name || '').localeCompare(b.name || '');
    });
  }, [data, tierFilter, search]);

  const summary = useMemo(() => {
    if (!data?.items) return { total: 0, deficit: 0, surplus: 0, balanced: 0 };
    let deficit = 0, surplus = 0, balanced = 0;
    for (const item of data.items) {
      if (item.netPerHour < -0.01) deficit++;
      else if (item.netPerHour > 0.01) surplus++;
      else balanced++;
    }
    return { total: data.items.length, deficit, surplus, balanced };
  }, [data]);

  if (loading) {
    return <Loading />;
  }

  const hasResizableItems = data?.items?.some(
    item => (item.stockpileMarkers?.length ?? 0) > 0 && item.consumedPerHour > 0
  );

  return (
    <div>
      {/* Summary chips */}
      <div className="flex items-center gap-3 mb-4 flex-wrap">
        <span className="text-xs text-text-muted">{summary.total} materials</span>
        {summary.deficit > 0 && (
          <span className="inline-flex items-center px-2 py-0.5 rounded text-[0.7rem] font-medium bg-rose-danger/10 text-rose-danger">
            {summary.deficit} deficit
          </span>
        )}
        {summary.surplus > 0 && (
          <span className="inline-flex items-center px-2 py-0.5 rounded text-[0.7rem] font-medium bg-teal-success/10 text-teal-success">
            {summary.surplus} surplus
          </span>
        )}
        {summary.balanced > 0 && (
          <span className="inline-flex items-center px-2 py-0.5 rounded text-[0.7rem] font-medium bg-overlay-subtle text-text-secondary">
            {summary.balanced} balanced
          </span>
        )}
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 mb-4 flex-wrap">
        <div className="flex gap-1">
          {TIERS.map(tier => (
            <button
              key={tier}
              onClick={() => setTierFilter(tier)}
              className={cn(
                'px-2.5 py-1 rounded text-xs font-medium border transition-colors',
                tierFilter === tier
                  ? 'bg-interactive-active text-primary border-border-active'
                  : 'bg-status-neutral-tint text-text-muted border-transparent hover:bg-overlay-subtle'
              )}
            >
              {tier}
            </button>
          ))}
        </div>
        {hasResizableItems && (
          <Popover open={bulkPopoverOpen} onOpenChange={setBulkPopoverOpen}>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className="h-7 text-xs text-text-secondary border-overlay-medium bg-transparent hover:bg-interactive-hover hover:border-border-active hover:text-primary"
              >
                Set all stockpiles
              </Button>
            </PopoverTrigger>
            <PopoverContent
              className="p-3 w-56 bg-background-elevated border-overlay-medium"
              align="start"
            >
              <span className="block text-[0.6rem] text-text-muted font-semibold uppercase tracking-wide mb-2">
                Set all stockpiles to cover
              </span>
              <div className="flex flex-col gap-1">
                {RESIZE_PRESETS.map(preset => (
                  <button
                    key={preset.label}
                    onClick={() => { setBulkPopoverOpen(false); handleResizeAll(preset.hours); }}
                    className="flex items-center text-text-emphasis text-xs py-1 px-2 rounded hover:bg-interactive-selected w-full text-left"
                  >
                    {preset.label}
                  </button>
                ))}
              </div>
            </PopoverContent>
          </Popover>
        )}
        <div className="ml-auto w-44">
          <Input
            placeholder="Search..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="h-8 text-[0.8rem] text-text-emphasis bg-transparent border-overlay-medium focus-visible:ring-0 focus-visible:border-primary hover:border-border-active"
          />
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className={cn(headerCellCls, "w-9")} />
              <TableHead className={headerCellCls}>Material</TableHead>
              <TableHead className={headerCellCls}>Source</TableHead>
              <TableHead className={cn(headerCellCls, "text-right")}>Produced/hr</TableHead>
              <TableHead className={cn(headerCellCls, "text-right")}>Consumed/hr</TableHead>
              <TableHead className={cn(headerCellCls, "text-right")}>Net/hr</TableHead>
              <TableHead className={cn(headerCellCls, "text-right")}>Stockpile</TableHead>
              <TableHead className={cn(headerCellCls, "text-right")}>Depletion</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredItems.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-text-muted border-overlay-subtle py-8">
                  No PI production data found
                </TableCell>
              </TableRow>
            ) : (
              filteredItems.map(item => (
                <SupplyChainRow
                  key={item.typeId}
                  item={item}
                  expanded={expandedItems.has(item.typeId)}
                  onToggle={() => toggleItem(item.typeId)}
                  onResize={handleResize}
                />
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

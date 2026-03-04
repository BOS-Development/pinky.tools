import { useState, useEffect, useCallback } from 'react';
import { useSession } from 'next-auth/react';
import Image from 'next/image';
import Link from 'next/link';
import Navbar from '@industry-tool/components/Navbar';
import Loading from '@industry-tool/components/loading';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { Separator } from '@/components/ui/separator';
import { ArrowLeft, Plus, Trash2, Pencil, ExternalLink } from 'lucide-react';
import { HaulingRun, HaulingRunItem, HaulingArbitrageRow, HaulingRunPnlEntry, HaulingRunPnlSummary } from '@industry-tool/client/data/models';
import { formatISK, formatNumber } from '@industry-tool/utils/formatting';
import { getItemIconUrl } from '@industry-tool/utils/eveImages';

const EVE_REGIONS: Record<number, string> = {
  10000002: 'The Forge',
  10000043: 'Domain',
  10000032: 'Sinq Laison',
  10000030: 'Heimatar',
  10000042: 'Metropolis',
  10000016: 'Lonetrek',
  10000033: 'The Citadel',
  10000065: 'Kor-Azor',
  10000001: 'Derelik',
  10000052: 'Kador',
};

const REGION_HUB_SYSTEMS: Record<number, number> = {
  10000002: 30000142, // The Forge → Jita
  10000043: 30002187, // Domain → Amarr
  10000032: 30002659, // Sinq Laison → Dodixie
  10000030: 30002510, // Heimatar → Rens
  10000042: 30002053, // Metropolis → Hek
  10000016: 30001161, // Lonetrek → Nourvukaiken
};

const STATUS_OPTIONS: HaulingRun['status'][] = [
  'PLANNING',
  'ACCUMULATING',
  'READY',
  'IN_TRANSIT',
  'SELLING',
  'COMPLETE',
  'CANCELLED',
];

function StatusBadge({ status }: { status: string }) {
  const config: Record<string, { color: string; bg: string }> = {
    PLANNING: { color: 'var(--color-primary-cyan)', bg: 'rgba(0, 212, 255, 0.1)' },
    ACCUMULATING: { color: 'var(--color-manufacturing-amber)', bg: 'rgba(245, 158, 11, 0.1)' },
    READY: { color: 'var(--color-success-teal)', bg: 'rgba(16, 185, 129, 0.1)' },
    IN_TRANSIT: { color: '#38bdf8', bg: 'rgba(56, 189, 248, 0.1)' },
    SELLING: { color: 'var(--color-manufacturing-amber)', bg: 'rgba(245, 158, 11, 0.1)' },
    COMPLETE: { color: 'var(--color-success-teal)', bg: 'rgba(16, 185, 129, 0.1)' },
    CANCELLED: { color: 'var(--color-danger-rose)', bg: 'rgba(239, 68, 68, 0.1)' },
  };
  const c = config[status] || { color: 'var(--color-text-secondary)', bg: 'rgba(148, 163, 184, 0.1)' };
  return (
    <span
      className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
      style={{ backgroundColor: c.bg, color: c.color }}
    >
      {status}
    </span>
  );
}

function getRowBgColor(fillPercent: number): string | undefined {
  if (fillPercent >= 100) return 'rgba(16, 185, 129, 0.08)';
  if (fillPercent < 50) return 'rgba(245, 158, 11, 0.06)';
  return undefined;
}

interface AddItemForm {
  typeId: string;
  typeName: string;
  quantityPlanned: string;
  buyPriceIsk: string;
  sellPriceIsk: string;
  volumeM3: string;
}

interface EditAcquiredForm {
  quantityAcquired: string;
}

interface PnlEntryForm {
  typeId: string;
  typeName: string;
  quantitySold: string;
  avgSellPriceIsk: string;
  totalCostIsk: string;
}

interface RouteSafetyData {
  jumps: number | null;
  fromName: string;
  toName: string;
  kills24h: number | null;
  loading: boolean;
}

interface HaulingRunDetailProps {
  runId: number;
}

export default function HaulingRunDetail({ runId }: HaulingRunDetailProps) {
  const { data: session } = useSession();
  const [run, setRun] = useState<HaulingRun | null>(null);
  const [loading, setLoading] = useState(true);
  const [scannerSuggestions, setScannerSuggestions] = useState<HaulingArbitrageRow[]>([]);

  // P&L state
  const [pnlEntries, setPnlEntries] = useState<HaulingRunPnlEntry[]>([]);
  const [pnlSummary, setPnlSummary] = useState<HaulingRunPnlSummary | null>(null);
  const [pnlDialogOpen, setPnlDialogOpen] = useState(false);
  const [editingPnlEntry, setEditingPnlEntry] = useState<HaulingRunPnlEntry | null>(null);
  const [pnlForm, setPnlForm] = useState<PnlEntryForm>({
    typeId: '',
    typeName: '',
    quantitySold: '',
    avgSellPriceIsk: '',
    totalCostIsk: '',
  });

  // Route safety state
  const [routeSafety, setRouteSafety] = useState<RouteSafetyData>({
    jumps: null,
    fromName: '',
    toName: '',
    kills24h: null,
    loading: false,
  });

  // Dialogs
  const [addItemOpen, setAddItemOpen] = useState(false);
  const [editAcquiredOpen, setEditAcquiredOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<HaulingRunItem | null>(null);

  const [addItemForm, setAddItemForm] = useState<AddItemForm>({
    typeId: '',
    typeName: '',
    quantityPlanned: '',
    buyPriceIsk: '',
    sellPriceIsk: '',
    volumeM3: '',
  });
  const [editAcquiredForm, setEditAcquiredForm] = useState<EditAcquiredForm>({ quantityAcquired: '' });
  const [submitting, setSubmitting] = useState(false);

  const fetchRun = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetch(`/api/hauling/runs/${runId}`);
      if (response.ok) {
        const data: HaulingRun = await response.json();
        setRun(data);

        // Fetch scanner suggestions for fill remaining capacity
        if (data.fromRegionId && data.toRegionId) {
          const params = new URLSearchParams({
            source_region_id: String(data.fromRegionId),
            dest_region_id: String(data.toRegionId),
          });
          const scanRes = await fetch(`/api/hauling/scanner?${params.toString()}`);
          if (scanRes.ok) {
            const scanData = await scanRes.json();
            setScannerSuggestions(Array.isArray(scanData) ? scanData : []);
          }
        }

        // Fetch P&L data for SELLING or COMPLETE runs
        if (data.status === 'SELLING' || data.status === 'COMPLETE') {
          const [pnlRes, summaryRes] = await Promise.allSettled([
            fetch(`/api/hauling/pnl?runId=${runId}`),
            fetch(`/api/hauling/pnl-summary?runId=${runId}`),
          ]);
          if (pnlRes.status === 'fulfilled' && pnlRes.value.ok) {
            const pnlData = await pnlRes.value.json();
            setPnlEntries(Array.isArray(pnlData) ? pnlData : []);
          }
          if (summaryRes.status === 'fulfilled' && summaryRes.value.ok) {
            const summaryData = await summaryRes.value.json();
            setPnlSummary(summaryData);
          }
        }
      }
    } catch (error) {
      console.error('Failed to fetch hauling run:', error);
    } finally {
      setLoading(false);
    }
  }, [runId]);

  const fetchRouteSafety = useCallback(async (fromRegionId: number, toRegionId: number) => {
    const fromSystemId = REGION_HUB_SYSTEMS[fromRegionId];
    const toSystemId = REGION_HUB_SYSTEMS[toRegionId];
    const fromHubName = EVE_REGIONS[fromRegionId] || `Region ${fromRegionId}`;
    const toHubName = EVE_REGIONS[toRegionId] || `Region ${toRegionId}`;

    setRouteSafety({ jumps: null, fromName: fromHubName, toName: toHubName, kills24h: null, loading: true });

    let jumps: number | null = null;
    let kills24h: number | null = null;

    if (fromSystemId && toSystemId) {
      try {
        const routeRes = await fetch(
          `https://esi.evetech.net/latest/route/${fromSystemId}/${toSystemId}/?flag=shortest`,
        );
        if (routeRes.ok) {
          const routeData = await routeRes.json();
          if (Array.isArray(routeData)) {
            jumps = routeData.length - 1;
          }
        }
      } catch {
        // ESI route unavailable — leave jumps as null
      }
    }

    try {
      const killsRes = await fetch(
        `https://zkillboard.com/api/kills/regionID/${fromRegionId}/pastSeconds/86400/`,
      );
      if (killsRes.ok) {
        const killsData = await killsRes.json();
        kills24h = Array.isArray(killsData) ? killsData.length : null;
      }
    } catch {
      // zKillboard unavailable — leave kills24h as null
    }

    setRouteSafety({ jumps, fromName: fromHubName, toName: toHubName, kills24h, loading: false });
  }, []);

  useEffect(() => {
    if (session) {
      fetchRun();
    }
  }, [session, fetchRun]);

  useEffect(() => {
    if (run && run.fromRegionId && run.toRegionId) {
      fetchRouteSafety(run.fromRegionId, run.toRegionId);
    }
  }, [run, fetchRouteSafety]);

  const handleStatusChange = async (newStatus: string) => {
    if (!run) return;
    try {
      const response = await fetch(`/api/hauling/runs/${runId}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus }),
      });
      if (response.ok) {
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to update status:', error);
    }
  };

  const handleAddItem = async () => {
    if (!addItemForm.typeId || !addItemForm.quantityPlanned) return;
    setSubmitting(true);
    try {
      const body: Record<string, unknown> = {
        typeId: Number(addItemForm.typeId),
        typeName: addItemForm.typeName || `Type ${addItemForm.typeId}`,
        quantityPlanned: Number(addItemForm.quantityPlanned),
      };
      if (addItemForm.buyPriceIsk) body.buyPriceIsk = Number(addItemForm.buyPriceIsk);
      if (addItemForm.sellPriceIsk) body.sellPriceIsk = Number(addItemForm.sellPriceIsk);
      if (addItemForm.volumeM3) body.volumeM3 = Number(addItemForm.volumeM3);

      const response = await fetch(`/api/hauling/runs/${runId}/items`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (response.ok) {
        setAddItemOpen(false);
        setAddItemForm({ typeId: '', typeName: '', quantityPlanned: '', buyPriceIsk: '', sellPriceIsk: '', volumeM3: '' });
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to add item:', error);
    } finally {
      setSubmitting(false);
    }
  };

  const handleEditAcquired = async () => {
    if (!editingItem) return;
    setSubmitting(true);
    try {
      const response = await fetch(`/api/hauling/runs/${runId}/items/${editingItem.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ quantityAcquired: Number(editAcquiredForm.quantityAcquired) }),
      });
      if (response.ok) {
        setEditAcquiredOpen(false);
        setEditingItem(null);
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to update acquired qty:', error);
    } finally {
      setSubmitting(false);
    }
  };

  const handleRemoveItem = async (itemId: number) => {
    try {
      const response = await fetch(`/api/hauling/runs/${runId}/items/${itemId}`, {
        method: 'DELETE',
      });
      if (response.ok) {
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to remove item:', error);
    }
  };

  const handleAddScannerItem = async (row: HaulingArbitrageRow) => {
    if (!run) return;
    const remainingM3 = (run.maxVolumeM3 || 0) - totalUsedVolume;
    const qty = row.volumeM3 && row.volumeM3 > 0
      ? Math.max(1, Math.floor(remainingM3 / row.volumeM3))
      : 1;

    try {
      const body: Record<string, unknown> = {
        typeId: row.typeId,
        typeName: row.typeName,
        quantityPlanned: qty,
      };
      if (row.buyPrice) body.buyPriceIsk = row.buyPrice;
      if (row.sellPrice) body.sellPriceIsk = row.sellPrice;
      if (row.volumeM3) body.volumeM3 = row.volumeM3;

      const response = await fetch(`/api/hauling/runs/${runId}/items`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (response.ok) {
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to add scanner item:', error);
    }
  };

  const handleOpenPnlDialog = (entry?: HaulingRunPnlEntry) => {
    if (entry) {
      setEditingPnlEntry(entry);
      setPnlForm({
        typeId: String(entry.typeId),
        typeName: entry.typeName || '',
        quantitySold: String(entry.quantitySold),
        avgSellPriceIsk: entry.avgSellPriceIsk !== undefined ? String(entry.avgSellPriceIsk) : '',
        totalCostIsk: entry.totalCostIsk !== undefined ? String(entry.totalCostIsk) : '',
      });
    } else {
      setEditingPnlEntry(null);
      setPnlForm({ typeId: '', typeName: '', quantitySold: '', avgSellPriceIsk: '', totalCostIsk: '' });
    }
    setPnlDialogOpen(true);
  };

  const handleSubmitPnl = async () => {
    if (!pnlForm.typeId || !pnlForm.quantitySold) return;
    setSubmitting(true);
    try {
      const body: Record<string, unknown> = {
        typeId: Number(pnlForm.typeId),
        typeName: pnlForm.typeName || `Type ${pnlForm.typeId}`,
        quantitySold: Number(pnlForm.quantitySold),
      };
      if (pnlForm.avgSellPriceIsk) body.avgSellPriceIsk = Number(pnlForm.avgSellPriceIsk);
      if (pnlForm.totalCostIsk) body.totalCostIsk = Number(pnlForm.totalCostIsk);

      const response = await fetch(`/api/hauling/pnl?runId=${runId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (response.ok) {
        setPnlDialogOpen(false);
        setEditingPnlEntry(null);
        setPnlForm({ typeId: '', typeName: '', quantitySold: '', avgSellPriceIsk: '', totalCostIsk: '' });
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to submit P&L entry:', error);
    } finally {
      setSubmitting(false);
    }
  };

  if (!session) return null;
  if (loading) return <Loading />;
  if (!run) {
    return (
      <>
        <Navbar />
        <div className="w-full px-4 mt-8">
          <p className="text-red-400">Run not found.</p>
        </div>
      </>
    );
  }

  const items = run.items || [];
  const totalUsedVolume = items.reduce((sum, item) => sum + (item.volumeM3 || 0) * item.quantityPlanned, 0);
  const maxVol = run.maxVolumeM3 || 0;
  const volumePercent = maxVol > 0 ? Math.min((totalUsedVolume / maxVol) * 100, 100) : 0;

  const totalOutlay = items.reduce((sum, item) => sum + (item.buyPriceIsk || 0) * item.quantityPlanned, 0);
  const totalRevenue = items.reduce((sum, item) => sum + (item.sellPriceIsk || 0) * item.quantityPlanned, 0);
  const netProfit = totalRevenue - totalOutlay;
  const overallFill = items.length > 0
    ? items.reduce((sum, item) => sum + item.fillPercent, 0) / items.length
    : 0;

  const fromName = EVE_REGIONS[run.fromRegionId] || `Region ${run.fromRegionId}`;
  const toName = EVE_REGIONS[run.toRegionId] || `Region ${run.toRegionId}`;

  // Top 3 scanner suggestions that fit remaining capacity
  const remainingM3 = maxVol > 0 ? maxVol - totalUsedVolume : Infinity;
  const topSuggestions = scannerSuggestions
    .filter((row) => !row.volumeM3 || row.volumeM3 <= remainingM3)
    .sort((a, b) => {
      const profitPerM3A = a.netProfitIsk && a.volumeM3 ? a.netProfitIsk / a.volumeM3 : 0;
      const profitPerM3B = b.netProfitIsk && b.volumeM3 ? b.netProfitIsk / b.volumeM3 : 0;
      return profitPerM3B - profitPerM3A;
    })
    .slice(0, 3);

  const capacityColor = volumePercent >= 95 ? 'var(--color-success-teal)' : volumePercent >= 70 ? 'var(--color-primary-cyan)' : 'var(--color-manufacturing-amber)';

  return (
    <TooltipProvider>
      <Navbar />
      <div className="w-full px-4 mt-8 mb-8">
        {/* Header */}
        <div className="flex items-center gap-4 mb-6">
          <Button asChild variant="ghost" size="sm" className="text-text-secondary hover:text-text-emphasis">
            <Link href="/hauling">
              <ArrowLeft className="h-4 w-4 mr-1" />
              Back
            </Link>
          </Button>
          <h1 className="text-2xl font-bold text-text-emphasis flex-1">{run.name}</h1>
          <div className="flex items-center gap-2">
            <StatusBadge status={run.status} />
            <div className="min-w-[160px]">
              <Select value="" onValueChange={(v) => handleStatusChange(v)}>
                <SelectTrigger className="bg-background-void border-overlay-strong text-text-secondary text-xs h-8">
                  <SelectValue placeholder="Change Status" />
                </SelectTrigger>
                <SelectContent className="bg-background-panel border-overlay-medium">
                  {STATUS_OPTIONS.map((s) => (
                    <SelectItem key={s} value={s} className="text-text-emphasis">
                      {s}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>

        {/* Route & Capacity */}
        <div className="flex gap-4 mb-6 flex-wrap">
          <Card className="flex-1 min-w-[200px] bg-background-panel border-overlay-subtle">
            <CardContent className="pt-4">
              <p className="text-sm text-text-secondary mb-1">Route</p>
              <p className="text-lg font-semibold text-text-emphasis">{fromName} &rarr; {toName}</p>
            </CardContent>
          </Card>
          {maxVol > 0 && (
            <Card className="flex-[2] min-w-[300px] bg-background-panel border-overlay-subtle">
              <CardContent className="pt-4">
                <p className="text-sm text-text-secondary mb-2">Capacity</p>
                <div className="flex items-center gap-4">
                  <div className="flex-1">
                    <div className="h-3 bg-[#1e2a3a] rounded-full overflow-hidden">
                      <div
                        className="h-full rounded-full transition-all"
                        style={{ width: `${volumePercent}%`, backgroundColor: capacityColor }}
                      />
                    </div>
                  </div>
                  <p className="text-sm text-text-secondary min-w-[140px] text-right">
                    {formatNumber(totalUsedVolume, 1)} / {formatNumber(maxVol)} m³
                  </p>
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Route Safety */}
        <Card className="mb-6 bg-background-panel border-overlay-subtle">
          <CardContent className="pt-4">
            <h2 className="text-lg font-semibold text-text-emphasis mb-3">Route Safety</h2>
            {routeSafety.loading ? (
              <p className="text-sm text-text-secondary">Loading route data...</p>
            ) : (
              <div className="flex gap-6 flex-wrap items-center">
                <div>
                  <p className="text-sm text-text-secondary">Shortest Route</p>
                  {routeSafety.jumps !== null ? (
                    <p className="font-semibold text-text-emphasis">
                      {routeSafety.jumps} jumps ({fromName} &rarr; {toName})
                    </p>
                  ) : (
                    <p className="text-text-emphasis">—</p>
                  )}
                </div>
                <div>
                  <p className="text-sm text-text-secondary">Source Region Kills (24h)</p>
                  {routeSafety.kills24h !== null ? (
                    <p
                      className="font-semibold"
                      style={{
                        color: routeSafety.kills24h <= 5
                          ? 'var(--color-success-teal)'
                          : routeSafety.kills24h <= 20
                          ? 'var(--color-manufacturing-amber)'
                          : 'var(--color-danger-rose)',
                      }}
                    >
                      {routeSafety.kills24h}
                    </p>
                  ) : (
                    <p className="text-text-emphasis">—</p>
                  )}
                </div>
              </div>
            )}
            <p className="text-xs text-text-muted mt-2">Data from EVE ESI and zKillboard</p>
            {/* TODO(Phase 3): Undercut alert — show warning if current dest market price
                has dropped below item sell prices since this run was created. */}
          </CardContent>
        </Card>

        {/* Items Table */}
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold text-text-emphasis">Items</h2>
          <Button size="sm" onClick={() => setAddItemOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Item
          </Button>
        </div>

        {items.length === 0 ? (
          <Card className="mb-6 bg-background-panel border-overlay-subtle">
            <CardContent className="py-6">
              <p className="text-center text-text-secondary">No items added yet.</p>
            </CardContent>
          </Card>
        ) : (
          <Card className="mb-6 bg-background-panel border-overlay-subtle">
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="bg-background-void border-overlay-subtle">
                    <TableHead className="font-bold text-text-emphasis">Item</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Buy Price</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Sell Price</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Net Profit/unit</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Planned Qty</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Acquired</TableHead>
                    <TableHead className="font-bold text-text-emphasis w-36">Fill %</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Volume</TableHead>
                    <TableHead className="font-bold text-text-emphasis">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((item) => {
                    const rowBg = getRowBgColor(item.fillPercent);
                    const netProfitUnit = item.netProfitIsk !== undefined
                      ? item.netProfitIsk
                      : (item.sellPriceIsk && item.buyPriceIsk)
                        ? item.sellPriceIsk - item.buyPriceIsk
                        : undefined;
                    const fillColor = item.fillPercent >= 100 ? 'var(--color-success-teal)' : 'var(--color-primary-cyan)';

                    return (
                      <TableRow
                        key={item.id}
                        className="border-[rgba(148,163,184,0.07)] hover:bg-[rgba(255,255,255,0.02)]"
                        style={{ backgroundColor: rowBg }}
                      >
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Image
                              src={getItemIconUrl(item.typeId, 32)}
                              alt={item.typeName}
                              width={24}
                              height={24}
                              style={{ borderRadius: 2 }}
                            />
                            <span className="text-sm font-semibold text-text-emphasis">{item.typeName}</span>
                          </div>
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">
                          {item.buyPriceIsk !== undefined ? formatISK(item.buyPriceIsk) : '—'}
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">
                          {item.sellPriceIsk !== undefined ? formatISK(item.sellPriceIsk) : '—'}
                        </TableCell>
                        <TableCell className="text-right">
                          {netProfitUnit !== undefined ? (
                            <span
                              className="text-sm font-semibold"
                              style={{ color: netProfitUnit >= 0 ? 'var(--color-success-teal)' : 'var(--color-danger-rose)' }}
                            >
                              {formatISK(netProfitUnit)}
                            </span>
                          ) : '—'}
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">
                          {item.quantityPlanned.toLocaleString()}
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">
                          {item.quantityAcquired.toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <div className="flex-1">
                              <div className="h-1.5 bg-[#1e2a3a] rounded-full overflow-hidden">
                                <div
                                  className="h-full rounded-full transition-all"
                                  style={{ width: `${Math.min(item.fillPercent, 100)}%`, backgroundColor: fillColor }}
                                />
                              </div>
                            </div>
                            <span className="text-xs text-text-secondary min-w-[36px] text-right">
                              {item.fillPercent.toFixed(0)}%
                            </span>
                          </div>
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">
                          {item.volumeM3 !== undefined
                            ? `${formatNumber(item.volumeM3 * item.quantityPlanned, 1)} m³`
                            : '—'}
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-1">
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-7 w-7 text-text-secondary hover:text-text-emphasis"
                                  onClick={() => {
                                    setEditingItem(item);
                                    setEditAcquiredForm({ quantityAcquired: String(item.quantityAcquired) });
                                    setEditAcquiredOpen(true);
                                  }}
                                >
                                  <Pencil className="h-3.5 w-3.5" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>Edit acquired qty</TooltipContent>
                            </Tooltip>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-7 w-7 text-red-400 hover:text-red-300 hover:bg-red-500/10"
                                  onClick={() => handleRemoveItem(item.id)}
                                >
                                  <Trash2 className="h-3.5 w-3.5" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>Remove item</TooltipContent>
                            </Tooltip>
                          </div>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          </Card>
        )}

        {/* Fill Remaining Capacity */}
        {topSuggestions.length > 0 && (
          <Card className="mb-6 bg-background-panel border-overlay-subtle">
            <CardContent className="pt-4">
              <div className="flex items-center justify-between mb-3">
                <h2 className="text-lg font-semibold text-text-emphasis">Fill Remaining Capacity</h2>
                <Button asChild variant="ghost" size="sm" className="text-text-secondary hover:text-text-emphasis">
                  <Link href={`/hauling/scanner?dest_region=${run.toRegionId}&source_region=${run.fromRegionId}`}>
                    See all
                    <ExternalLink className="h-3.5 w-3.5 ml-1" />
                  </Link>
                </Button>
              </div>
              <div className="flex flex-col gap-2">
                {topSuggestions.map((row) => {
                  const profitPerM3 = row.netProfitIsk && row.volumeM3
                    ? row.netProfitIsk / row.volumeM3
                    : undefined;
                  return (
                    <div
                      key={row.typeId}
                      className="flex items-center gap-4 p-2 rounded bg-background-void"
                    >
                      <Image
                        src={getItemIconUrl(row.typeId, 32)}
                        alt={row.typeName}
                        width={24}
                        height={24}
                        style={{ borderRadius: 2 }}
                      />
                      <span className="flex-1 text-sm font-semibold text-text-emphasis">{row.typeName}</span>
                      {row.netProfitIsk !== undefined && (
                        <span className="text-sm text-teal-success">{formatISK(row.netProfitIsk)}/unit</span>
                      )}
                      {profitPerM3 !== undefined && (
                        <span className="text-xs text-text-muted">{formatISK(profitPerM3)}/m³</span>
                      )}
                      {row.volumeM3 !== undefined && (
                        <span className="text-xs text-text-muted">{formatNumber(row.volumeM3, 2)} m³</span>
                      )}
                      <Button size="sm" variant="outline" onClick={() => handleAddScannerItem(row)}>
                        Add
                      </Button>
                    </div>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        )}

        {/* P&L Section — shown for SELLING or COMPLETE runs */}
        {(run.status === 'SELLING' || run.status === 'COMPLETE') && (
          <Card className="mb-6 bg-background-panel border-overlay-subtle">
            <CardContent className="pt-4">
              <div className="flex items-center justify-between mb-3">
                <h2 className="text-lg font-semibold text-text-emphasis">Profit &amp; Loss</h2>
                <Button size="sm" onClick={() => handleOpenPnlDialog()}>
                  <Plus className="h-4 w-4 mr-2" />
                  Enter P&amp;L
                </Button>
              </div>
              {pnlEntries.length === 0 ? (
                <p className="text-center text-sm text-text-secondary py-4">
                  No P&amp;L entries yet. Mark items as sold to track profit.
                </p>
              ) : (
                <div className="overflow-x-auto mb-4">
                  <Table>
                    <TableHeader>
                      <TableRow className="bg-background-void border-overlay-subtle">
                        <TableHead className="font-bold text-text-emphasis">Type</TableHead>
                        <TableHead className="font-bold text-text-emphasis text-right">Qty Sold</TableHead>
                        <TableHead className="font-bold text-text-emphasis text-right">Avg Sell Price</TableHead>
                        <TableHead className="font-bold text-text-emphasis text-right">Revenue</TableHead>
                        <TableHead className="font-bold text-text-emphasis text-right">Cost</TableHead>
                        <TableHead className="font-bold text-text-emphasis text-right">Net Profit</TableHead>
                        <TableHead className="font-bold text-text-emphasis">Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {pnlEntries.map((entry) => (
                        <TableRow key={entry.id} className="border-[rgba(148,163,184,0.07)] hover:bg-[rgba(255,255,255,0.02)]">
                          <TableCell className="text-sm text-text-emphasis">
                            {entry.typeName || `Type ${entry.typeId}`}
                          </TableCell>
                          <TableCell className="text-right text-sm text-text-secondary">
                            {formatNumber(entry.quantitySold)}
                          </TableCell>
                          <TableCell className="text-right text-sm text-text-secondary">
                            {entry.avgSellPriceIsk !== undefined ? formatISK(entry.avgSellPriceIsk) : '—'}
                          </TableCell>
                          <TableCell className="text-right text-sm" style={{ color: 'var(--color-success-teal)' }}>
                            {entry.totalRevenueIsk !== undefined ? formatISK(entry.totalRevenueIsk) : '—'}
                          </TableCell>
                          <TableCell className="text-right text-sm" style={{ color: 'var(--color-danger-rose)' }}>
                            {entry.totalCostIsk !== undefined ? formatISK(entry.totalCostIsk) : '—'}
                          </TableCell>
                          <TableCell className="text-right">
                            {entry.netProfitIsk !== undefined ? (
                              <span
                                className="text-sm font-semibold"
                                style={{ color: entry.netProfitIsk >= 0 ? 'var(--color-success-teal)' : 'var(--color-danger-rose)' }}
                              >
                                {formatISK(entry.netProfitIsk)}
                              </span>
                            ) : '—'}
                          </TableCell>
                          <TableCell>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-7 w-7 text-text-secondary hover:text-text-emphasis"
                                  onClick={() => handleOpenPnlDialog(entry)}
                                >
                                  <Pencil className="h-3.5 w-3.5" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>Edit P&L entry</TooltipContent>
                            </Tooltip>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
              {pnlSummary && (
                <div className="flex gap-6 flex-wrap mt-2">
                  <div>
                    <p className="text-sm text-text-secondary">Total Revenue</p>
                    <p className="text-lg font-semibold" style={{ color: 'var(--color-success-teal)' }}>
                      {formatISK(pnlSummary.totalRevenueIsk)}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-secondary">Total Cost</p>
                    <p className="text-lg font-semibold" style={{ color: 'var(--color-danger-rose)' }}>
                      {formatISK(pnlSummary.totalCostIsk)}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-secondary">Net Profit</p>
                    <p
                      className="text-lg font-semibold"
                      style={{ color: pnlSummary.netProfitIsk >= 0 ? 'var(--color-success-teal)' : 'var(--color-danger-rose)' }}
                    >
                      {formatISK(pnlSummary.netProfitIsk)}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-secondary">Margin %</p>
                    <p className="text-lg font-semibold text-text-emphasis">
                      {pnlSummary.marginPct.toFixed(1)}%
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-secondary">Items Sold</p>
                    <p className="text-lg font-semibold text-text-emphasis">{formatNumber(pnlSummary.itemsSold)}</p>
                  </div>
                  <div>
                    <p className="text-sm text-text-secondary">Items Pending</p>
                    <p className="text-lg font-semibold text-text-emphasis">{formatNumber(pnlSummary.itemsPending)}</p>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Stats Footer */}
        <Card className="bg-background-panel border-overlay-subtle">
          <CardContent className="pt-4">
            <h2 className="text-lg font-semibold text-text-emphasis mb-2">Run Summary</h2>
            <Separator className="mb-4 bg-overlay-subtle" />
            <div className="flex gap-6 flex-wrap">
              <div>
                <p className="text-sm text-text-secondary">Total ISK Outlay</p>
                <p className="text-lg font-semibold" style={{ color: 'var(--color-danger-rose)' }}>
                  {formatISK(totalOutlay)}
                </p>
              </div>
              {run.status === 'COMPLETE' && pnlSummary ? (
                <>
                  <div>
                    <p className="text-sm text-text-secondary">Actual Revenue</p>
                    <p className="text-lg font-semibold" style={{ color: 'var(--color-success-teal)' }}>
                      {formatISK(pnlSummary.totalRevenueIsk)}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-secondary">Actual Net Profit</p>
                    <p
                      className="text-lg font-semibold"
                      style={{ color: pnlSummary.netProfitIsk >= 0 ? 'var(--color-success-teal)' : 'var(--color-danger-rose)' }}
                    >
                      {formatISK(pnlSummary.netProfitIsk)}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-secondary">Margin %</p>
                    <p className="text-lg font-semibold text-text-emphasis">
                      {pnlSummary.marginPct.toFixed(1)}%
                    </p>
                  </div>
                </>
              ) : (
                <>
                  <div>
                    <p className="text-sm text-text-secondary">Total ISK Revenue</p>
                    <p className="text-lg font-semibold" style={{ color: 'var(--color-success-teal)' }}>
                      {formatISK(totalRevenue)}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-text-secondary">Net Profit</p>
                    <p
                      className="text-lg font-semibold"
                      style={{ color: netProfit >= 0 ? 'var(--color-success-teal)' : 'var(--color-danger-rose)' }}
                    >
                      {formatISK(netProfit)}
                    </p>
                  </div>
                </>
              )}
              <div>
                <p className="text-sm text-text-secondary">Fill % Overall</p>
                <p className="text-lg font-semibold text-text-emphasis">{overallFill.toFixed(1)}%</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Add Item Dialog */}
      <Dialog open={addItemOpen} onOpenChange={setAddItemOpen}>
        <DialogContent className="max-w-md bg-background-panel border-overlay-medium">
          <DialogHeader>
            <DialogTitle className="text-text-emphasis">Add Item</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4 mt-2">
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Type ID *</label>
              <Input
                type="number"
                value={addItemForm.typeId}
                onChange={(e) => setAddItemForm({ ...addItemForm, typeId: e.target.value })}
                autoFocus
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Item Name *</label>
              <Input
                value={addItemForm.typeName}
                onChange={(e) => setAddItemForm({ ...addItemForm, typeName: e.target.value })}
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Planned Quantity *</label>
              <Input
                type="number"
                value={addItemForm.quantityPlanned}
                onChange={(e) => setAddItemForm({ ...addItemForm, quantityPlanned: e.target.value })}
                min={1}
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Buy Price (ISK)</label>
              <Input
                type="number"
                value={addItemForm.buyPriceIsk}
                onChange={(e) => setAddItemForm({ ...addItemForm, buyPriceIsk: e.target.value })}
                min={0}
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Sell Price (ISK)</label>
              <Input
                type="number"
                value={addItemForm.sellPriceIsk}
                onChange={(e) => setAddItemForm({ ...addItemForm, sellPriceIsk: e.target.value })}
                min={0}
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Volume (m³)</label>
              <Input
                type="number"
                value={addItemForm.volumeM3}
                onChange={(e) => setAddItemForm({ ...addItemForm, volumeM3: e.target.value })}
                min={0}
                step={0.01}
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
            </div>
          </div>
          <DialogFooter className="mt-4">
            <Button variant="ghost" onClick={() => setAddItemOpen(false)} className="text-text-secondary">
              Cancel
            </Button>
            <Button
              onClick={handleAddItem}
              disabled={!addItemForm.typeId || !addItemForm.quantityPlanned || !addItemForm.typeName || submitting}
            >
              {submitting ? 'Adding...' : 'Add'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Acquired Dialog */}
      <Dialog open={editAcquiredOpen} onOpenChange={setEditAcquiredOpen}>
        <DialogContent className="max-w-xs bg-background-panel border-overlay-medium">
          <DialogHeader>
            <DialogTitle className="text-text-emphasis">Update Acquired Quantity</DialogTitle>
          </DialogHeader>
          <div className="mt-2">
            <p className="text-sm text-text-secondary mb-3">{editingItem?.typeName}</p>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Quantity Acquired</label>
              <Input
                type="number"
                value={editAcquiredForm.quantityAcquired}
                onChange={(e) => setEditAcquiredForm({ quantityAcquired: e.target.value })}
                autoFocus
                min={0}
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
            </div>
          </div>
          <DialogFooter className="mt-4">
            <Button variant="ghost" onClick={() => setEditAcquiredOpen(false)} className="text-text-secondary">
              Cancel
            </Button>
            <Button
              onClick={handleEditAcquired}
              disabled={editAcquiredForm.quantityAcquired === '' || submitting}
            >
              {submitting ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* P&L Entry Dialog */}
      <Dialog open={pnlDialogOpen} onOpenChange={setPnlDialogOpen}>
        <DialogContent className="max-w-md bg-background-panel border-overlay-medium">
          <DialogHeader>
            <DialogTitle className="text-text-emphasis">
              {editingPnlEntry ? 'Edit P&L Entry' : 'Enter P&L'}
            </DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4 mt-2">
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Item *</label>
              <Select
                value={pnlForm.typeId}
                onValueChange={(v) => {
                  const selectedItem = items.find((item) => String(item.typeId) === v);
                  setPnlForm({
                    ...pnlForm,
                    typeId: v,
                    typeName: selectedItem ? selectedItem.typeName : pnlForm.typeName,
                  });
                }}
              >
                <SelectTrigger className="bg-background-void border-overlay-strong text-text-emphasis">
                  <SelectValue placeholder="Select item" />
                </SelectTrigger>
                <SelectContent className="bg-background-panel border-overlay-medium">
                  {items.map((item) => (
                    <SelectItem key={item.typeId} value={String(item.typeId)} className="text-text-emphasis">
                      {item.typeName}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Qty Sold *</label>
              <Input
                type="number"
                value={pnlForm.quantitySold}
                onChange={(e) => setPnlForm({ ...pnlForm, quantitySold: e.target.value })}
                min={1}
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Avg Sell Price (ISK)</label>
              <Input
                type="number"
                value={pnlForm.avgSellPriceIsk}
                onChange={(e) => setPnlForm({ ...pnlForm, avgSellPriceIsk: e.target.value })}
                min={0}
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Total Cost (ISK, optional)</label>
              <Input
                type="number"
                value={pnlForm.totalCostIsk}
                onChange={(e) => setPnlForm({ ...pnlForm, totalCostIsk: e.target.value })}
                min={0}
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
              <p className="text-xs text-text-muted mt-1">Total buy cost for this item lot</p>
            </div>
          </div>
          <DialogFooter className="mt-4">
            <Button variant="ghost" onClick={() => setPnlDialogOpen(false)} className="text-text-secondary">
              Cancel
            </Button>
            <Button
              onClick={handleSubmitPnl}
              disabled={!pnlForm.typeId || !pnlForm.quantitySold || submitting}
            >
              {submitting ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </TooltipProvider>
  );
}

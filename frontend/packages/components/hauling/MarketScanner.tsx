import { useState, useEffect, useCallback, useRef } from 'react';
import { useSession } from 'next-auth/react';
import Image from 'next/image';
import Navbar from '@industry-tool/components/Navbar';
import Loading from '@industry-tool/components/loading';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
import { ScanSearch, ShoppingCart } from 'lucide-react';
import { HaulingArbitrageRow, HaulingRun } from '@industry-tool/client/data/models';
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

const REGION_OPTIONS = Object.entries(EVE_REGIONS).map(([id, name]) => ({
  id: Number(id),
  name,
}));

function getIndicatorStyle(indicator: HaulingArbitrageRow['indicator']): { color: string; bg: string } {
  switch (indicator) {
    case 'gap': return { color: 'var(--color-manufacturing-amber)', bg: 'var(--color-warning-tint)' };
    case 'markup': return { color: 'var(--color-primary-cyan)', bg: 'var(--color-info-tint)' };
    case 'thin': return { color: 'var(--color-text-secondary)', bg: 'var(--color-neutral-tint)' };
    default: return { color: 'var(--color-text-secondary)', bg: 'var(--color-neutral-tint)' };
  }
}

function getRowBgColor(indicator: HaulingArbitrageRow['indicator']): string | undefined {
  switch (indicator) {
    case 'gap': return 'var(--color-warning-tint)';
    case 'markup': return 'var(--color-info-tint)';
    case 'thin': return undefined;
    default: return undefined;
  }
}

interface AddToRunState {
  row: HaulingArbitrageRow | null;
  open: boolean;
  selectedRunId: number | '';
  quantity: string;
}

interface MarketScannerProps {
  initialSourceRegion?: number;
  initialDestRegion?: number;
}

export default function MarketScanner({ initialSourceRegion, initialDestRegion }: MarketScannerProps) {
  const { data: session } = useSession();
  const [sourceRegionId, setSourceRegionId] = useState<number>(initialSourceRegion || 10000002);
  const [destRegionId, setDestRegionId] = useState<number>(initialDestRegion || 10000043);
  const [results, setResults] = useState<HaulingArbitrageRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [scanning, setScanning] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<string | null>(null);
  const [runs, setRuns] = useState<HaulingRun[]>([]);
  const [selectedRowIndex, setSelectedRowIndex] = useState<number>(-1);

  const [addToRun, setAddToRun] = useState<AddToRunState>({
    row: null,
    open: false,
    selectedRunId: '',
    quantity: '1',
  });

  const tableRef = useRef<HTMLDivElement>(null);

  const fetchResults = useCallback(async (srcId: number, dstId: number) => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        source_region_id: String(srcId),
        dest_region_id: String(dstId),
      });
      const response = await fetch(`/api/hauling/scanner?${params.toString()}`);
      if (response.ok) {
        const data = await response.json();
        const rows: HaulingArbitrageRow[] = Array.isArray(data) ? data : [];
        // Sort by net profit descending
        rows.sort((a, b) => (b.netProfitIsk || 0) - (a.netProfitIsk || 0));
        setResults(rows);
        if (rows.length > 0) {
          setLastUpdated(rows[0].updatedAt);
        }
      }
    } catch (error) {
      console.error('Failed to fetch scanner results:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchRuns = useCallback(async () => {
    try {
      const response = await fetch('/api/hauling/runs');
      if (response.ok) {
        const data = await response.json();
        setRuns(Array.isArray(data) ? data : []);
      }
    } catch (error) {
      console.error('Failed to fetch runs:', error);
    }
  }, []);

  useEffect(() => {
    if (session) {
      fetchRuns();
    }
  }, [session, fetchRuns]);

  const handleScan = async () => {
    setScanning(true);
    try {
      await fetch('/api/hauling/scanner', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ regionId: sourceRegionId, destRegionId }),
      });
      // Scan is synchronous — results are ready after POST completes
      await fetchResults(sourceRegionId, destRegionId);
    } catch (error) {
      console.error('Failed to trigger scan:', error);
    } finally {
      setScanning(false);
    }
  };

  const handleFetchResults = () => {
    fetchResults(sourceRegionId, destRegionId);
  };

  const handleAddToRun = async () => {
    if (!addToRun.row || addToRun.selectedRunId === '') return;
    const qty = Number(addToRun.quantity) || 1;
    const row = addToRun.row;

    try {
      const body: Record<string, unknown> = {
        typeId: row.typeId,
        typeName: row.typeName,
        quantityPlanned: qty,
      };
      if (row.buyPrice) body.buyPriceIsk = row.buyPrice;
      if (row.sellPrice) body.sellPriceIsk = row.sellPrice;
      if (row.volumeM3) body.volumeM3 = row.volumeM3;

      await fetch(`/api/hauling/runs/${addToRun.selectedRunId}/items`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      setAddToRun({ row: null, open: false, selectedRunId: '', quantity: '1' });
    } catch (error) {
      console.error('Failed to add item to run:', error);
    }
  };

  // Keyboard navigation
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedRowIndex((prev) => Math.min(prev + 1, results.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedRowIndex((prev) => Math.max(prev - 1, 0));
    } else if (e.key === 'Enter' && selectedRowIndex >= 0 && runs.length > 0) {
      e.preventDefault();
      const row = results[selectedRowIndex];
      if (row) {
        setAddToRun({
          row,
          open: true,
          selectedRunId: runs[0]?.id || '',
          quantity: '1',
        });
      }
    }
  }, [results, selectedRowIndex, runs]);

  if (!session) return null;

  return (
    <>
      <Navbar />
      <div className="w-full px-4 mt-8 mb-8">
        {/* Header */}
        <div className="flex items-center gap-2 mb-6">
          <ScanSearch className="h-7 w-7 text-text-secondary" />
          <h1 className="text-2xl font-bold text-text-emphasis">Market Scanner</h1>
        </div>

        {/* Controls */}
        <Card className="mb-6 bg-background-panel border-overlay-subtle">
          <CardContent className="pt-4">
            <div className="flex gap-4 items-center flex-wrap">
              <div className="min-w-[180px]">
                <label className="text-xs text-text-secondary mb-1 block">Source Region</label>
                <Select
                  value={String(sourceRegionId)}
                  onValueChange={(v) => setSourceRegionId(Number(v))}
                >
                  <SelectTrigger className="bg-background-void border-overlay-strong text-text-emphasis">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-background-panel border-overlay-medium">
                    {REGION_OPTIONS.map((r) => (
                      <SelectItem key={r.id} value={r.id.toString()} className="text-text-emphasis">
                        {r.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="min-w-[180px]">
                <label className="text-xs text-text-secondary mb-1 block">Destination Region</label>
                <Select
                  value={String(destRegionId)}
                  onValueChange={(v) => setDestRegionId(Number(v))}
                >
                  <SelectTrigger className="bg-background-void border-overlay-strong text-text-emphasis">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-background-panel border-overlay-medium">
                    {REGION_OPTIONS.map((r) => (
                      <SelectItem key={r.id} value={r.id.toString()} className="text-text-emphasis">
                        {r.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-end gap-2 pt-4">
                <Button
                  variant="outline"
                  onClick={handleFetchResults}
                  disabled={loading}
                >
                  {loading ? 'Loading...' : 'Load'}
                </Button>
                <Button
                  onClick={handleScan}
                  disabled={scanning || loading}
                >
                  <ScanSearch className="h-4 w-4 mr-2" />
                  {scanning ? 'Scanning...' : 'Scan'}
                </Button>
              </div>
              {lastUpdated && (
                <span className="text-xs text-text-muted ml-1">
                  Last updated: {new Date(lastUpdated).toLocaleString()}
                </span>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Results Table */}
        {loading ? (
          <Card className="bg-background-panel border-overlay-subtle">
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="bg-background-void border-overlay-subtle">
                    {['Item', 'Indicator', 'Net Profit/unit', 'm³', 'Days to Sell', 'Buy Price', 'Sell Price', 'Volume Available', 'Add to Run'].map((h) => (
                      <TableHead key={h} className="font-bold text-text-emphasis">{h}</TableHead>
                    ))}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {Array.from({ length: 8 }).map((_, i) => (
                    <TableRow key={i} className="border-overlay-subtle">
                      {Array.from({ length: 9 }).map((__, j) => (
                        <TableCell key={j}>
                          <Skeleton className="h-4 w-full bg-background-elevated" />
                        </TableCell>
                      ))}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </Card>
        ) : results.length === 0 ? (
          <Card className="bg-background-panel border-overlay-subtle">
            <CardContent className="py-8">
              <p className="text-center text-text-secondary text-base">
                No arbitrage opportunities found. Try scanning to refresh data.
              </p>
            </CardContent>
          </Card>
        ) : (
          <Card className="bg-background-panel border-overlay-subtle">
            <div
              className="overflow-x-auto outline-none"
              ref={tableRef}
              tabIndex={0}
              onKeyDown={handleKeyDown}
            >
              <Table>
                <TableHeader>
                  <TableRow className="bg-background-void border-overlay-subtle">
                    <TableHead className="font-bold text-text-emphasis">Item</TableHead>
                    <TableHead className="font-bold text-text-emphasis">Indicator</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Net Profit/unit</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">m³</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Days to Sell</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Buy Price</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Sell Price</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Volume Available</TableHead>
                    <TableHead className="font-bold text-text-emphasis">Add to Run</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {results.map((row, idx) => {
                    const rowBg = getRowBgColor(row.indicator);
                    const isSelected = idx === selectedRowIndex;
                    const indicatorStyle = getIndicatorStyle(row.indicator);

                    return (
                      <TableRow
                        key={row.typeId}
                        className="cursor-pointer border-overlay-subtle"
                        style={{
                          backgroundColor: isSelected ? 'var(--color-interactive-selected)' : rowBg,
                        }}
                        onClick={() => setSelectedRowIndex(idx)}
                      >
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Image
                              src={getItemIconUrl(row.typeId, 32)}
                              alt={row.typeName}
                              width={24}
                              height={24}
                              style={{ borderRadius: 2 }}
                            />
                            <span className="text-sm font-semibold text-text-emphasis">{row.typeName}</span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <span
                            className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border"
                            style={{
                              color: indicatorStyle.color,
                              backgroundColor: indicatorStyle.bg,
                              borderColor: indicatorStyle.color + '40',
                            }}
                          >
                            {row.indicator}
                          </span>
                        </TableCell>
                        <TableCell className="text-right">
                          {row.netProfitIsk !== undefined ? (
                            <span
                              className="text-sm font-semibold"
                              style={{ color: (row.netProfitIsk || 0) >= 0 ? 'var(--color-success-teal)' : 'var(--color-danger-rose)' }}
                            >
                              {formatISK(row.netProfitIsk || 0)}
                            </span>
                          ) : '—'}
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">
                          {row.volumeM3 !== undefined ? formatNumber(row.volumeM3, 2) : '—'}
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">
                          {row.daysToSell !== undefined ? formatNumber(row.daysToSell, 1) : '—'}
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">
                          {row.buyPrice !== undefined ? formatISK(row.buyPrice) : '—'}
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">
                          {row.sellPrice !== undefined ? formatISK(row.sellPrice) : '—'}
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">
                          {row.volumeAvailable !== undefined ? formatNumber(row.volumeAvailable) : '—'}
                        </TableCell>
                        <TableCell>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={(e) => {
                              e.stopPropagation();
                              setAddToRun({
                                row,
                                open: true,
                                selectedRunId: runs[0]?.id || '',
                                quantity: '1',
                              });
                            }}
                            disabled={runs.length === 0}
                          >
                            <ShoppingCart className="h-3.5 w-3.5 mr-1" />
                            Add
                          </Button>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          </Card>
        )}

        {/* Add to Run Dialog */}
        <Dialog
          open={addToRun.open}
          onOpenChange={(open) => {
            if (!open) setAddToRun({ row: null, open: false, selectedRunId: '', quantity: '1' });
          }}
        >
          <DialogContent className="max-w-xs bg-background-panel border-overlay-medium">
            <DialogHeader>
              <DialogTitle className="text-text-emphasis">Add to Run</DialogTitle>
            </DialogHeader>
            <div className="flex flex-col gap-4 mt-2">
              <p className="text-sm font-semibold text-text-emphasis">{addToRun.row?.typeName}</p>
              <div>
                <label className="text-xs text-text-secondary mb-1 block">Hauling Run</label>
                <Select
                  value={addToRun.selectedRunId !== '' ? String(addToRun.selectedRunId) : ''}
                  onValueChange={(v) => setAddToRun({ ...addToRun, selectedRunId: Number(v) })}
                >
                  <SelectTrigger className="bg-background-void border-overlay-strong text-text-emphasis">
                    <SelectValue placeholder="Select run" />
                  </SelectTrigger>
                  <SelectContent className="bg-background-panel border-overlay-medium">
                    {runs.map((run) => (
                      <SelectItem key={run.id} value={String(run.id)} className="text-text-emphasis">
                        {run.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <label className="text-xs text-text-secondary mb-1 block">Quantity</label>
                <Input
                  type="number"
                  value={addToRun.quantity}
                  onChange={(e) => setAddToRun({ ...addToRun, quantity: e.target.value })}
                  min={1}
                  className="bg-background-void border-overlay-strong text-text-emphasis"
                />
              </div>
            </div>
            <DialogFooter className="mt-4">
              <Button
                onClick={handleAddToRun}
                disabled={addToRun.selectedRunId === ''}
                className="w-full"
              >
                Add to Run
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </>
  );
}

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
import { ScanSearch, ShoppingCart, Settings, AlertTriangle, Loader2 } from 'lucide-react';
import {
  HaulingArbitrageRow,
  HaulingRun,
  ScannerLocation,
  TradingStation,
  UserTradingStructure,
} from '@industry-tool/client/data/models';
import { formatISK, formatNumber } from '@industry-tool/utils/formatting';
import { getItemIconUrl } from '@industry-tool/utils/eveImages';
import UserStructuresDialog from './UserStructuresDialog';

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

// Build a location value string for use in Select component
function locationToValue(loc: ScannerLocation): string {
  return `${loc.type}:${loc.id}`;
}

// Default locations
const DEFAULT_SOURCE: ScannerLocation = {
  type: 'region',
  id: 10000002,
  name: 'The Forge',
  regionId: 10000002,
  systemId: 0,
};

const DEFAULT_DEST: ScannerLocation = {
  type: 'region',
  id: 10000043,
  name: 'Domain',
  regionId: 10000043,
  systemId: 0,
};

function buildRegionLocations(): ScannerLocation[] {
  return REGION_OPTIONS.map((r) => ({
    type: 'region' as const,
    id: r.id,
    name: r.name,
    regionId: r.id,
    systemId: 0,
  }));
}

function buildStationLocations(stations: TradingStation[]): ScannerLocation[] {
  return stations.map((s) => ({
    type: 'station' as const,
    id: s.stationId,
    name: s.name,
    regionId: s.regionId,
    systemId: s.systemId,
  }));
}

function buildStructureLocations(structs: UserTradingStructure[]): ScannerLocation[] {
  return structs.map((s) => ({
    type: 'structure' as const,
    dbId: s.id,
    id: s.structureId,
    name: s.name,
    regionId: s.regionId,
    systemId: s.systemId,
    structureId: s.structureId,
  }));
}

function findLocationByValue(
  value: string,
  regions: ScannerLocation[],
  stations: ScannerLocation[],
  structures: ScannerLocation[],
): ScannerLocation | undefined {
  const all = [...stations, ...regions, ...structures];
  return all.find((l) => locationToValue(l) === value);
}

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

  // Build initial locations from props
  const initSource: ScannerLocation = initialSourceRegion
    ? {
        type: 'region',
        id: initialSourceRegion,
        name: EVE_REGIONS[initialSourceRegion] || String(initialSourceRegion),
        regionId: initialSourceRegion,
        systemId: 0,
      }
    : DEFAULT_SOURCE;

  const initDest: ScannerLocation = initialDestRegion
    ? {
        type: 'region',
        id: initialDestRegion,
        name: EVE_REGIONS[initialDestRegion] || String(initialDestRegion),
        regionId: initialDestRegion,
        systemId: 0,
      }
    : DEFAULT_DEST;

  const [sourceLocation, setSourceLocation] = useState<ScannerLocation>(initSource);
  const [destLocation, setDestLocation] = useState<ScannerLocation>(initDest);

  const [stations, setStations] = useState<TradingStation[]>([]);
  const [structures, setStructures] = useState<UserTradingStructure[]>([]);

  const [results, setResults] = useState<HaulingArbitrageRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [scanning, setScanning] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<string | null>(null);
  const [runs, setRuns] = useState<HaulingRun[]>([]);
  const [selectedRowIndex, setSelectedRowIndex] = useState<number>(-1);
  const [structuresDialogOpen, setStructuresDialogOpen] = useState(false);

  const [addToRun, setAddToRun] = useState<AddToRunState>({
    row: null,
    open: false,
    selectedRunId: '',
    quantity: '1',
  });

  const tableRef = useRef<HTMLDivElement>(null);

  const fetchStations = useCallback(async () => {
    try {
      const res = await fetch('/api/hauling/stations');
      if (res.ok) {
        const data = await res.json();
        setStations(Array.isArray(data) ? data : []);
      }
    } catch (err) {
      console.error('Failed to fetch stations:', err);
    }
  }, []);

  const fetchStructures = useCallback(async () => {
    try {
      const res = await fetch('/api/hauling/structures');
      if (res.ok) {
        const data = await res.json();
        setStructures(Array.isArray(data) ? data : []);
      }
    } catch (err) {
      console.error('Failed to fetch structures:', err);
    }
  }, []);

  useEffect(() => {
    if (session) {
      fetchStations();
      fetchStructures();
    }
  }, [session, fetchStations, fetchStructures]);

  // Build location lists
  const regionLocs = buildRegionLocations();
  const stationLocs = buildStationLocations(stations);
  const structureLocs = buildStructureLocations(structures);

  const buildScannerParams = useCallback(
    (loc: ScannerLocation, prefix: 'source' | 'dest', params: URLSearchParams) => {
      if (loc.type === 'structure' && loc.structureId) {
        params.set(`${prefix}_structure_id`, String(loc.structureId));
      } else {
        if (prefix === 'source') {
          params.set('source_region_id', String(loc.regionId));
          if (loc.systemId > 0) params.set('source_system_id', String(loc.systemId));
        } else {
          params.set('dest_region_id', String(loc.regionId));
          if (loc.systemId > 0) params.set('dest_system_id', String(loc.systemId));
        }
      }
    },
    [],
  );

  const fetchResults = useCallback(
    async (srcLoc: ScannerLocation, dstLoc: ScannerLocation) => {
      setLoading(true);
      try {
        const params = new URLSearchParams();
        buildScannerParams(srcLoc, 'source', params);
        buildScannerParams(dstLoc, 'dest', params);

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
    },
    [buildScannerParams],
  );

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
      // If either location is a structure, scan it via structure-scan endpoint
      const scanPromises: Promise<void>[] = [];

      if (sourceLocation.type === 'structure' && sourceLocation.dbId) {
        scanPromises.push(
          fetch('/api/hauling/structure-scan', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: sourceLocation.dbId }),
          }).then(() => undefined),
        );
      }
      if (destLocation.type === 'structure' && destLocation.dbId) {
        scanPromises.push(
          fetch('/api/hauling/structure-scan', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: destLocation.dbId }),
          }).then(() => undefined),
        );
      }

      // If both are non-structure, use the region scan endpoint
      if (sourceLocation.type !== 'structure' || destLocation.type !== 'structure') {
        const body: Record<string, unknown> = {};
        if (sourceLocation.type === 'structure') {
          body.sourceStructureId = sourceLocation.structureId;
        } else {
          body.regionId = sourceLocation.regionId;
          body.systemId = sourceLocation.systemId;
        }
        if (destLocation.type === 'structure') {
          body.destStructureId = destLocation.structureId;
        } else {
          body.destRegionId = destLocation.regionId;
          body.destSystemId = destLocation.systemId;
        }

        scanPromises.push(
          fetch('/api/hauling/scanner', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
          }).then(() => undefined),
        );
      }

      await Promise.all(scanPromises);
      await fetchResults(sourceLocation, destLocation);
    } catch (error) {
      console.error('Failed to trigger scan:', error);
    } finally {
      setScanning(false);
    }
  };

  const handleFetchResults = () => {
    fetchResults(sourceLocation, destLocation);
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
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
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
    },
    [results, selectedRowIndex, runs],
  );

  const handleStructuresChanged = useCallback(async () => {
    await fetchStructures();
  }, [fetchStructures]);

  const sourceHasAccessWarning =
    sourceLocation.type === 'structure' &&
    structures.find((s) => s.id === sourceLocation.dbId)?.accessOk === false;

  const destHasAccessWarning =
    destLocation.type === 'structure' &&
    structures.find((s) => s.id === destLocation.dbId)?.accessOk === false;

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
            <div className="flex gap-4 items-end flex-wrap">
              {/* Source Picker */}
              <div className="min-w-[220px]">
                <div className="flex items-center justify-between mb-1">
                  <label className="text-xs text-text-secondary">Source Location</label>
                </div>
                <Select
                  value={locationToValue(sourceLocation)}
                  onValueChange={(v) => {
                    const loc = findLocationByValue(v, regionLocs, stationLocs, structureLocs);
                    if (loc) setSourceLocation(loc);
                  }}
                >
                  <SelectTrigger className="bg-background-void border-overlay-strong text-text-emphasis">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-background-panel border-overlay-medium">
                    {stationLocs.length > 0 && (
                      <>
                        <SelectItem value="__stations_label__" disabled className="text-xs text-text-muted font-semibold uppercase tracking-wide opacity-70 cursor-default">
                          Stations
                        </SelectItem>
                        {stationLocs.map((loc) => (
                          <SelectItem key={locationToValue(loc)} value={locationToValue(loc)} className="text-text-emphasis pl-4">
                            {loc.name}
                          </SelectItem>
                        ))}
                      </>
                    )}
                    <SelectItem value="__regions_label__" disabled className="text-xs text-text-muted font-semibold uppercase tracking-wide opacity-70 cursor-default">
                      Regions
                    </SelectItem>
                    {regionLocs.map((loc) => (
                      <SelectItem key={locationToValue(loc)} value={locationToValue(loc)} className="text-text-emphasis pl-4">
                        {loc.name}
                      </SelectItem>
                    ))}
                    <>
                      <SelectItem value="__structures_label__" disabled className="text-xs text-text-muted font-semibold uppercase tracking-wide opacity-70 cursor-default">
                        My Structures
                      </SelectItem>
                      {structureLocs.length > 0 ? (
                        structureLocs.map((loc) => {
                          const struct = structures.find((s) => s.id === loc.dbId);
                          return (
                            <SelectItem key={locationToValue(loc)} value={locationToValue(loc)} className="text-text-emphasis pl-4">
                              {struct?.accessOk === false ? '⚠ ' : ''}{loc.name}
                            </SelectItem>
                          );
                        })
                      ) : (
                        <SelectItem value="__no_structures__" disabled className="text-xs text-text-muted italic pl-4">
                          No structures added yet
                        </SelectItem>
                      )}
                    </>
                  </SelectContent>
                </Select>
              </div>

              {/* Dest Picker */}
              <div className="min-w-[220px]">
                <label className="text-xs text-text-secondary mb-1 block">Destination Location</label>
                <Select
                  value={locationToValue(destLocation)}
                  onValueChange={(v) => {
                    const loc = findLocationByValue(v, regionLocs, stationLocs, structureLocs);
                    if (loc) setDestLocation(loc);
                  }}
                >
                  <SelectTrigger className="bg-background-void border-overlay-strong text-text-emphasis">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-background-panel border-overlay-medium">
                    {stationLocs.length > 0 && (
                      <>
                        <SelectItem value="__stations_label_d__" disabled className="text-xs text-text-muted font-semibold uppercase tracking-wide opacity-70 cursor-default">
                          Stations
                        </SelectItem>
                        {stationLocs.map((loc) => (
                          <SelectItem key={locationToValue(loc)} value={locationToValue(loc)} className="text-text-emphasis pl-4">
                            {loc.name}
                          </SelectItem>
                        ))}
                      </>
                    )}
                    <SelectItem value="__regions_label_d__" disabled className="text-xs text-text-muted font-semibold uppercase tracking-wide opacity-70 cursor-default">
                      Regions
                    </SelectItem>
                    {regionLocs.map((loc) => (
                      <SelectItem key={locationToValue(loc)} value={locationToValue(loc)} className="text-text-emphasis pl-4">
                        {loc.name}
                      </SelectItem>
                    ))}
                    <>
                      <SelectItem value="__structures_label_d__" disabled className="text-xs text-text-muted font-semibold uppercase tracking-wide opacity-70 cursor-default">
                        My Structures
                      </SelectItem>
                      {structureLocs.length > 0 ? (
                        structureLocs.map((loc) => {
                          const struct = structures.find((s) => s.id === loc.dbId);
                          return (
                            <SelectItem key={locationToValue(loc)} value={locationToValue(loc)} className="text-text-emphasis pl-4">
                              {struct?.accessOk === false ? '⚠ ' : ''}{loc.name}
                            </SelectItem>
                          );
                        })
                      ) : (
                        <SelectItem value="__no_structures_d__" disabled className="text-xs text-text-muted italic pl-4">
                          No structures added yet
                        </SelectItem>
                      )}
                    </>
                  </SelectContent>
                </Select>
              </div>

              {/* Action Buttons */}
              <div className="flex items-center gap-2">
                <Button variant="outline" onClick={handleFetchResults} disabled={loading}>
                  {loading ? 'Loading...' : 'Load'}
                </Button>
                <Button onClick={handleScan} disabled={scanning || loading}>
                  <ScanSearch className="h-4 w-4 mr-2" />
                  {scanning ? 'Scanning...' : 'Scan'}
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setStructuresDialogOpen(true)}
                  title="Manage trading structures"
                >
                  <Settings className="h-4 w-4" />
                </Button>
              </div>

              {lastUpdated && (
                <span className="text-xs text-text-muted ml-1">
                  Last updated: {new Date(lastUpdated).toLocaleString()}
                </span>
              )}
              {structureLocs.length === 0 && (
                <span className="text-xs text-text-muted">
                  No trading structures added.{' '}
                  <button
                    className="underline cursor-pointer hover:text-text-secondary"
                    onClick={() => setStructuresDialogOpen(true)}
                  >
                    Add one via the ⚙ button
                  </button>{' '}
                  to scan player-owned markets.
                </span>
              )}
            </div>

            {/* Access warnings */}
            {(sourceHasAccessWarning || destHasAccessWarning) && (
              <div className="mt-3 flex items-start gap-2 text-sm text-amber-500 bg-status-warning-tint border border-amber-500/30 rounded-md px-3 py-2">
                <AlertTriangle className="h-4 w-4 shrink-0 mt-0.5" />
                <span>
                  Structure market access failed. Check your docking rights or rescan from the Structure Manager (
                  <button
                    className="underline cursor-pointer"
                    onClick={() => setStructuresDialogOpen(true)}
                  >
                    open
                  </button>
                  ).
                </span>
              </div>
            )}
            {scanning && (
              <div className="mt-3 flex items-center gap-2 text-xs text-text-muted">
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
                <span>
                  {(sourceLocation.type === 'structure' || destLocation.type === 'structure')
                    ? 'Scanning structure market...'
                    : 'Fetching market data from ESI...'}
                </span>
              </div>
            )}
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

        {/* Structures Manager Dialog */}
        <UserStructuresDialog
          open={structuresDialogOpen}
          onClose={() => setStructuresDialogOpen(false)}
          onStructuresChanged={handleStructuresChanged}
        />
      </div>
    </>
  );
}

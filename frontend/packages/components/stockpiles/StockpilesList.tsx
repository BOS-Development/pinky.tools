import { useState, useMemo, useEffect, useRef } from 'react';
import { useSession } from "next-auth/react";
import Navbar from "@industry-tool/components/Navbar";
import Loading from "@industry-tool/components/loading";
import { AlertTriangle, DollarSign, Copy, ExternalLink, Search } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import {
  Table, TableHeader, TableBody, TableRow, TableHead, TableCell,
} from '@/components/ui/table';
import { toast } from '@/components/ui/sonner';
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from '@/components/ui/tooltip';

export type StockpileItem = {
  name: string;
  typeId: number;
  quantity: number;
  volume: number;
  ownerType: string;
  ownerName: string;
  ownerId: number;
  desiredQuantity: number;
  stockpileDelta: number;
  deficitValue: number;
  structureName: string;
  solarSystem: string;
  region: string;
  containerName?: string;
  planId?: number;
  planName?: string;
  autoProductionEnabled?: boolean;
  autoProductionParallelism?: number;
};

export type StockpilesResponse = {
  items: StockpileItem[];
};

export default function StockpilesList() {
  const { data: session } = useSession();
  const [stockpileItems, setStockpileItems] = useState<StockpileItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [creatingAppraisal, setCreatingAppraisal] = useState(false);
  const hasFetchedRef = useRef(false);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchStockpileDeficits();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session]);

  const fetchStockpileDeficits = async () => {
    if (!session) return;

    setLoading(true);
    try {
      const response = await fetch('/api/stockpiles/deficits');
      if (response.ok) {
        const data: StockpilesResponse = await response.json();
        setStockpileItems(data.items || []);
      }
    } finally {
      setLoading(false);
    }
  };

  const filteredItems = useMemo(() => {
    if (!searchQuery) return stockpileItems;

    const query = searchQuery.toLowerCase();
    return stockpileItems.filter(
      (item) =>
        item.name.toLowerCase().includes(query) ||
        item.structureName.toLowerCase().includes(query) ||
        item.solarSystem.toLowerCase().includes(query) ||
        item.region.toLowerCase().includes(query) ||
        item.containerName?.toLowerCase().includes(query)
    );
  }, [stockpileItems, searchQuery]);

  const totalDeficit = useMemo(() => {
    return filteredItems.reduce((sum, item) => sum + Math.abs(item.stockpileDelta), 0);
  }, [filteredItems]);

  const totalVolume = useMemo(() => {
    return filteredItems.reduce((sum, item) => {
      const deficit = Math.abs(item.stockpileDelta);
      const perUnitVolume = item.quantity > 0 ? item.volume / item.quantity : item.volume;
      return sum + (deficit * perUnitVolume);
    }, 0);
  }, [filteredItems]);

  const totalDeficitISK = useMemo(() => {
    return filteredItems.reduce((sum, item) => sum + item.deficitValue, 0);
  }, [filteredItems]);

  const handleCopyForJanice = async () => {
    const janiceText = filteredItems
      .map((item) => `${item.name} ${Math.abs(item.stockpileDelta)}`)
      .join('\n');

    try {
      await navigator.clipboard.writeText(janiceText);
      toast.success('Copied to clipboard! Paste into Janice for appraisal.');
    } catch {
      toast.error('Failed to copy to clipboard');
    }
  };

  const handleOpenJanice = async () => {
    if (!session) return;

    const janiceText = filteredItems
      .map((item) => `${item.name} ${Math.abs(item.stockpileDelta)}`)
      .join('\n');

    setCreatingAppraisal(true);
    try {
      const response = await fetch('/api/janice/appraisal', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ items: janiceText }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to create Janice appraisal: ${response.status} - ${errorText}`);
      }

      const data = await response.json();

      if (data.code) {
        window.open(`https://janice.e-351.com/a/${data.code}`, '_blank');
        toast.success('Janice appraisal created and opened!');
      } else {
        throw new Error('Janice response missing appraisal code');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      toast.error(`Failed: ${errorMessage}`);
    } finally {
      setCreatingAppraisal(false);
    }
  };

  if (!session) return null;
  if (loading) return <Loading />;

  return (
    <>
      <Navbar />
      <div className="w-full px-4 py-8">
        {/* Sticky Header Section */}
        <div className="sticky top-16 z-40 bg-[var(--color-bg-void)] pb-4">
          <h1 className="text-2xl font-display font-semibold flex items-center gap-2 mb-4">
            <AlertTriangle className="h-6 w-6 text-[var(--color-danger-rose)]" />
            Stockpiles Needing Replenishment
          </h1>

          {/* Summary Stats */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
            <Card>
              <CardContent className="p-4">
                <p className="text-sm text-[var(--color-text-secondary)] mb-1">Items Below Target</p>
                <p className="text-3xl font-bold text-[var(--color-text-emphasis)]">{filteredItems.length}</p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <p className="text-sm text-[var(--color-text-secondary)] mb-1">Total Deficit</p>
                <p className="text-3xl font-bold text-[var(--color-danger-rose)]">{totalDeficit.toLocaleString()}</p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <p className="text-sm text-[var(--color-text-secondary)] mb-1">Total Volume</p>
                <p className="text-3xl font-bold text-[var(--color-text-emphasis)]">
                  {totalVolume.toLocaleString(undefined, { maximumFractionDigits: 2 })} m&sup3;
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <p className="text-sm text-[var(--color-text-secondary)] flex items-center gap-1 mb-1">
                  <DollarSign className="h-4 w-4 text-[var(--color-success-teal)]" />
                  Total Cost (ISK)
                </p>
                <p className="text-3xl font-bold text-[var(--color-danger-rose)]">
                  {totalDeficitISK.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Actions */}
          {stockpileItems.length > 0 && (
            <div className="flex gap-3 mb-3">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span>
                      <Button variant="outline" onClick={handleCopyForJanice} disabled={filteredItems.length === 0}>
                        <Copy className="h-4 w-4 mr-2" />
                        Copy for Janice
                      </Button>
                    </span>
                  </TooltipTrigger>
                  {filteredItems.length === 0 && (
                    <TooltipContent>No matching items to copy</TooltipContent>
                  )}
                </Tooltip>
              </TooltipProvider>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span>
                      <Button onClick={handleOpenJanice} disabled={filteredItems.length === 0 || creatingAppraisal}>
                        <ExternalLink className="h-4 w-4 mr-2" />
                        {creatingAppraisal ? 'Creating...' : 'Create Janice Appraisal'}
                      </Button>
                    </span>
                  </TooltipTrigger>
                  {filteredItems.length === 0 && (
                    <TooltipContent>No matching items to copy</TooltipContent>
                  )}
                </Tooltip>
              </TooltipProvider>
            </div>
          )}

          {/* Search */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-[var(--color-text-muted)]" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search items, structures, or locations..."
              className="pl-9"
            />
          </div>
        </div>

        {/* Items Table */}
        {filteredItems.length === 0 ? (
          <Card>
            <CardContent className="p-8 text-center">
              <p className="text-lg text-[var(--color-text-secondary)]">
                {stockpileItems.length === 0
                  ? 'No stockpiles need replenishment!'
                  : 'No items match your search.'}
              </p>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Item</TableHead>
                    <TableHead>Structure</TableHead>
                    <TableHead>Location</TableHead>
                    <TableHead>Container</TableHead>
                    <TableHead className="text-right">Current</TableHead>
                    <TableHead className="text-right">Target</TableHead>
                    <TableHead className="text-right">Deficit</TableHead>
                    <TableHead className="text-right">Cost (ISK)</TableHead>
                    <TableHead>Auto-Production</TableHead>
                    <TableHead>Owner</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredItems.map((item, idx) => (
                    <TableRow
                      key={idx}
                      className="border-l-4 border-l-[var(--color-danger-rose)] odd:bg-[var(--color-surface-elevated)]/30"
                    >
                      <TableCell className="font-semibold text-[var(--color-text-emphasis)]">{item.name}</TableCell>
                      <TableCell className="text-[var(--color-text-secondary)]">{item.structureName}</TableCell>
                      <TableCell className="text-[var(--color-text-secondary)]">{item.solarSystem}, {item.region}</TableCell>
                      <TableCell className="text-[var(--color-text-secondary)]">{item.containerName || '-'}</TableCell>
                      <TableCell className="text-right text-[var(--color-text-secondary)]">{item.quantity.toLocaleString()}</TableCell>
                      <TableCell className="text-right text-[var(--color-text-secondary)]">{item.desiredQuantity.toLocaleString()}</TableCell>
                      <TableCell className="text-right">
                        <span className="text-[var(--color-danger-rose)] font-semibold">
                          {item.stockpileDelta.toLocaleString()}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-[var(--color-danger-rose)] font-semibold">
                          {item.deficitValue.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                        </span>
                      </TableCell>
                      <TableCell>
                        {item.autoProductionEnabled ? (
                          <span className="text-[var(--color-success-teal)] font-semibold text-sm">
                            {item.planName || 'Linked'}
                          </span>
                        ) : (
                          <span className="text-[var(--color-text-muted)]">&mdash;</span>
                        )}
                      </TableCell>
                      <TableCell className="text-[var(--color-text-secondary)]">{item.ownerName}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        )}
      </div>
    </>
  );
}

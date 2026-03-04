import { useState, useEffect, useRef, useMemo } from 'react';
import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import { PiProfitResponse, PiFactoryProfit } from "@industry-tool/client/data/models";
import { formatISK, formatNumber, FONT_NUMERIC } from "@industry-tool/utils/formatting";
import { Card, CardContent } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { ChevronDown, ChevronRight, TrendingUp, TrendingDown } from 'lucide-react';

type ProductGroup = {
  outputTypeId: number;
  outputName: string;
  outputTier: string;
  totalRatePerHour: number;
  totalOutputValue: number;
  totalInputCost: number;
  totalExportTax: number;
  totalImportTax: number;
  totalProfit: number;
  factories: FactoryWithPlanet[];
};

type FactoryWithPlanet = PiFactoryProfit & {
  solarSystemName: string;
  characterName: string;
  planetType: string;
};

function profitColor(value: number): string {
  if (value > 0) return 'var(--color-success-teal)';
  if (value < 0) return 'var(--color-danger-rose)';
  return 'var(--color-text-secondary)';
}

const TIER_ORDER: Record<string, number> = { R0: 0, P1: 1, P2: 2, P3: 3, P4: 4 };

function ProductGroupRow({ group, expanded, onToggle }: {
  group: ProductGroup;
  expanded: boolean;
  onToggle: () => void;
}) {
  const totalTax = group.totalExportTax + group.totalImportTax;

  return (
    <>
      <TableRow
        className={cn(
          'cursor-pointer hover:bg-interactive-hover',
          expanded ? 'bg-[rgba(0,212,255,0.03)]' : ''
        )}
        onClick={onToggle}
      >
        <TableCell className="border-overlay-subtle w-10 p-1">
          <Button variant="ghost" size="icon" className="h-6 w-6 text-text-muted p-0">
            {expanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
          </Button>
        </TableCell>
        <TableCell className="border-overlay-subtle">
          <div className="flex items-center gap-2">
            {group.outputTypeId > 0 && (
              <img
                src={`https://images.evetech.net/types/${group.outputTypeId}/icon?size=32`}
                alt="" width={20} height={20} className="flex-shrink-0"
              />
            )}
            <div>
              <p className="text-sm text-text-emphasis font-medium leading-tight">
                {group.outputName}
              </p>
              <span className="text-xs text-text-muted">
                {group.outputTier} &middot; {group.factories.length} {group.factories.length === 1 ? 'factory' : 'factories'}
              </span>
            </div>
          </div>
        </TableCell>
        <TableCell className="text-right text-text-secondary border-overlay-subtle">
          {formatNumber(Math.round(group.totalRatePerHour))}/hr
        </TableCell>
        <TableCell className="text-right text-teal-success border-overlay-subtle">
          {formatISK(group.totalOutputValue)}
        </TableCell>
        <TableCell className="text-right text-rose-danger border-overlay-subtle">
          {formatISK(group.totalInputCost)}
        </TableCell>
        <TableCell className="text-right text-amber-manufacturing border-overlay-subtle">
          {formatISK(totalTax)}
        </TableCell>
        <TableCell
          className="text-right font-semibold border-overlay-subtle"
          style={{ color: profitColor(group.totalProfit) }}
        >
          {formatISK(group.totalProfit)}
        </TableCell>
      </TableRow>
      {expanded && group.factories.map((factory) => (
        <FactoryDetailRow key={`${factory.pinId}`} factory={factory} />
      ))}
    </>
  );
}

function FactoryDetailRow({ factory }: { factory: FactoryWithPlanet }) {
  return (
    <TableRow className="bg-[rgba(15,18,25,0.5)]">
      <TableCell className="border-[rgba(148,163,184,0.05)]" />
      <TableCell className="border-[rgba(148,163,184,0.05)] pl-12">
        <span className="text-xs text-text-primary">{factory.solarSystemName}</span>
        <span className="text-xs text-text-muted ml-1">({factory.characterName})</span>
      </TableCell>
      <TableCell className="text-right border-[rgba(148,163,184,0.05)]">
        <span className="text-xs text-text-muted">{formatNumber(Math.round(factory.ratePerHour))}/hr</span>
      </TableCell>
      <TableCell className="text-right border-[rgba(148,163,184,0.05)]">
        <span className="text-xs text-teal-success">{formatISK(factory.outputValuePerHour)}</span>
      </TableCell>
      <TableCell className="text-right border-[rgba(148,163,184,0.05)]">
        <span className="text-xs text-rose-danger">{formatISK(factory.inputCostPerHour)}</span>
      </TableCell>
      <TableCell className="text-right border-[rgba(148,163,184,0.05)]">
        <span className="text-xs text-amber-manufacturing">{formatISK(factory.exportTaxPerHour + factory.importTaxPerHour)}</span>
      </TableCell>
      <TableCell className="text-right border-[rgba(148,163,184,0.05)]">
        <span className="text-xs font-semibold" style={{ color: profitColor(factory.profitPerHour) }}>
          {formatISK(factory.profitPerHour)}
        </span>
      </TableCell>
    </TableRow>
  );
}

export default function ProfitTable() {
  const { data: session } = useSession();
  const [profitResponse, setProfitResponse] = useState<PiProfitResponse | null>(null);
  const [priceSource, setPriceSource] = useState<string>('sell');
  const [loading, setLoading] = useState(true);
  const [expandedGroups, setExpandedGroups] = useState<Set<number>>(new Set());
  const hasFetchedRef = useRef(false);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchProfit('sell');
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session]);

  const fetchProfit = async (source: string) => {
    if (!session) return;
    setLoading(true);
    try {
      const response = await fetch(`/api/pi/profit?priceSource=${source}`);
      if (response.ok) {
        const data: PiProfitResponse = await response.json();
        setProfitResponse(data);
      }
    } finally {
      setLoading(false);
    }
  };

  const profitData = profitResponse?.planets || [];

  const handlePriceSourceChange = (newSource: string) => {
    setPriceSource(newSource);
    hasFetchedRef.current = false;
    fetchProfit(newSource);
    hasFetchedRef.current = true;
  };

  const toggleGroup = (typeId: number) => {
    setExpandedGroups(prev => {
      const next = new Set(prev);
      if (next.has(typeId)) next.delete(typeId);
      else next.add(typeId);
      return next;
    });
  };

  // Group all factories across planets by output type
  const productGroups = useMemo(() => {
    const groupMap = new Map<number, ProductGroup>();

    for (const planet of profitData) {
      for (const factory of planet.factories) {
        let group = groupMap.get(factory.outputTypeId);
        if (!group) {
          group = {
            outputTypeId: factory.outputTypeId,
            outputName: factory.outputName,
            outputTier: factory.outputTier,
            totalRatePerHour: 0,
            totalOutputValue: 0,
            totalInputCost: 0,
            totalExportTax: 0,
            totalImportTax: 0,
            totalProfit: 0,
            factories: [],
          };
          groupMap.set(factory.outputTypeId, group);
        }
        group.totalRatePerHour += factory.ratePerHour;
        group.totalOutputValue += factory.outputValuePerHour;
        group.totalInputCost += factory.inputCostPerHour;
        group.totalExportTax += factory.exportTaxPerHour;
        group.totalImportTax += factory.importTaxPerHour;
        group.totalProfit += factory.profitPerHour;
        group.factories.push({
          ...factory,
          solarSystemName: planet.solarSystemName,
          characterName: planet.characterName,
          planetType: planet.planetType,
        });
      }
    }

    // Sort by tier then by name
    return Array.from(groupMap.values()).sort((a, b) => {
      const tierDiff = (TIER_ORDER[a.outputTier] ?? 99) - (TIER_ORDER[b.outputTier] ?? 99);
      if (tierDiff !== 0) return tierDiff;
      return a.outputName.localeCompare(b.outputName);
    });
  }, [profitData]);

  const totals = useMemo(() => {
    if (!profitResponse) return { output: 0, input: 0, exportTax: 0, importTax: 0, totalTax: 0, profit: 0 };
    return {
      output: profitResponse.totalOutputValue,
      input: profitResponse.totalInputCost,
      exportTax: profitResponse.totalExportTax,
      importTax: profitResponse.totalImportTax,
      totalTax: profitResponse.totalExportTax + profitResponse.totalImportTax,
      profit: profitResponse.totalProfit,
    };
  }, [profitResponse]);

  if (loading) {
    return <Loading />;
  }

  return (
    <div>
      {/* Summary cards + controls */}
      <div className="flex justify-between items-start mb-4 flex-wrap gap-4">
        <div className="flex gap-4 flex-wrap">
          <SummaryCard
            label="Revenue / hr"
            value={formatISK(totals.output)}
            color="var(--color-success-teal)"
            icon={<TrendingUp className="w-4 h-4 text-teal-success" />}
          />
          <SummaryCard
            label="Costs / hr"
            value={formatISK(totals.input)}
            color="var(--color-danger-rose)"
            icon={<TrendingDown className="w-4 h-4 text-rose-danger" />}
          />
          <SummaryCard
            label="Taxes / hr"
            value={formatISK(totals.totalTax)}
            color="var(--color-manufacturing-amber)"
          />
          <SummaryCard
            label="Profit / hr"
            value={formatISK(totals.profit)}
            color={profitColor(totals.profit)}
            bold
          />
        </div>
        {/* Price source toggle */}
        <div className="flex rounded overflow-hidden border border-overlay-strong">
          {(['sell', 'buy', 'split'] as const).map((source) => (
            <button
              key={source}
              onClick={() => handlePriceSourceChange(source)}
              className={cn(
                'px-3 py-1 text-xs font-medium capitalize transition-colors',
                priceSource === source
                  ? 'bg-interactive-selected text-primary border-border-active'
                  : 'text-text-muted hover:text-text-secondary hover:bg-[rgba(148,163,184,0.05)]'
              )}
            >
              {source.charAt(0).toUpperCase() + source.slice(1)}
            </button>
          ))}
        </div>
      </div>

      {/* Profit table grouped by product */}
      <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="text-text-muted border-overlay-subtle bg-background-void w-10" />
              <TableHead className="text-text-muted border-overlay-subtle bg-background-void">Product</TableHead>
              <TableHead className="text-right text-text-muted border-overlay-subtle bg-background-void">Rate</TableHead>
              <TableHead className="text-right text-text-muted border-overlay-subtle bg-background-void">Revenue/hr</TableHead>
              <TableHead className="text-right text-text-muted border-overlay-subtle bg-background-void">Costs/hr</TableHead>
              <TableHead className="text-right text-text-muted border-overlay-subtle bg-background-void">Taxes/hr</TableHead>
              <TableHead className="text-right text-text-muted border-overlay-subtle bg-background-void">Profit/hr</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {productGroups.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-text-muted border-overlay-subtle py-8">
                  No PI profit data available
                </TableCell>
              </TableRow>
            ) : (
              productGroups.map((group) => (
                <ProductGroupRow
                  key={group.outputTypeId}
                  group={group}
                  expanded={expandedGroups.has(group.outputTypeId)}
                  onToggle={() => toggleGroup(group.outputTypeId)}
                />
              ))
            )}
            {productGroups.length > 0 && (
              <TableRow className="bg-background-void">
                <TableCell className="border-overlay-subtle" />
                <TableCell className="text-text-emphasis font-semibold border-overlay-subtle">
                  Total ({productGroups.length} products)
                </TableCell>
                <TableCell className="border-overlay-subtle" />
                <TableCell className="text-right text-teal-success font-semibold border-overlay-subtle">
                  {formatISK(totals.output)}
                </TableCell>
                <TableCell className="text-right text-rose-danger font-semibold border-overlay-subtle">
                  {formatISK(totals.input)}
                </TableCell>
                <TableCell className="text-right text-amber-manufacturing font-semibold border-overlay-subtle">
                  {formatISK(totals.totalTax)}
                </TableCell>
                <TableCell
                  className="text-right font-bold border-overlay-subtle"
                  style={{ color: profitColor(totals.profit) }}
                >
                  {formatISK(totals.profit)}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

function SummaryCard({ label, value, color, icon, bold }: {
  label: string;
  value: string;
  color: string;
  icon?: React.ReactNode;
  bold?: boolean;
}) {
  return (
    <Card
      className="min-w-[140px] rounded-lg"
      style={{
        background: 'var(--color-bg-panel)',
        border: `1px solid ${color}25`,
      }}
    >
      <CardContent className="p-3 pb-3">
        <div className="flex items-center gap-1 mb-1">
          {icon}
          <span className="text-xs text-text-muted">{label}</span>
        </div>
        <p
          className={cn('text-sm', bold ? 'font-bold' : 'font-semibold')}
          style={{ color, fontFamily: FONT_NUMERIC }}
        >
          {value}
        </p>
      </CardContent>
    </Card>
  );
}

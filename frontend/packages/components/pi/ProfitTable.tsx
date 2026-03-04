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
  if (value > 0) return '#10b981';
  if (value < 0) return '#ef4444';
  return '#94a3b8';
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
          'cursor-pointer hover:bg-[rgba(0,212,255,0.04)]',
          expanded ? 'bg-[rgba(0,212,255,0.03)]' : ''
        )}
        onClick={onToggle}
      >
        <TableCell className="border-[rgba(148,163,184,0.1)] w-10 p-1">
          <Button variant="ghost" size="icon" className="h-6 w-6 text-[#64748b] p-0">
            {expanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
          </Button>
        </TableCell>
        <TableCell className="border-[rgba(148,163,184,0.1)]">
          <div className="flex items-center gap-2">
            {group.outputTypeId > 0 && (
              <img
                src={`https://images.evetech.net/types/${group.outputTypeId}/icon?size=32`}
                alt="" width={20} height={20} className="flex-shrink-0"
              />
            )}
            <div>
              <p className="text-sm text-[#e2e8f0] font-medium leading-tight">
                {group.outputName}
              </p>
              <span className="text-xs text-[#64748b]">
                {group.outputTier} &middot; {group.factories.length} {group.factories.length === 1 ? 'factory' : 'factories'}
              </span>
            </div>
          </div>
        </TableCell>
        <TableCell className="text-right text-[#94a3b8] border-[rgba(148,163,184,0.1)]">
          {formatNumber(Math.round(group.totalRatePerHour))}/hr
        </TableCell>
        <TableCell className="text-right text-[#10b981] border-[rgba(148,163,184,0.1)]">
          {formatISK(group.totalOutputValue)}
        </TableCell>
        <TableCell className="text-right text-[#ef4444] border-[rgba(148,163,184,0.1)]">
          {formatISK(group.totalInputCost)}
        </TableCell>
        <TableCell className="text-right text-[#f59e0b] border-[rgba(148,163,184,0.1)]">
          {formatISK(totalTax)}
        </TableCell>
        <TableCell
          className="text-right font-semibold border-[rgba(148,163,184,0.1)]"
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
        <span className="text-xs text-[#cbd5e1]">{factory.solarSystemName}</span>
        <span className="text-xs text-[#475569] ml-1">({factory.characterName})</span>
      </TableCell>
      <TableCell className="text-right border-[rgba(148,163,184,0.05)]">
        <span className="text-xs text-[#64748b]">{formatNumber(Math.round(factory.ratePerHour))}/hr</span>
      </TableCell>
      <TableCell className="text-right border-[rgba(148,163,184,0.05)]">
        <span className="text-xs text-[#10b981]">{formatISK(factory.outputValuePerHour)}</span>
      </TableCell>
      <TableCell className="text-right border-[rgba(148,163,184,0.05)]">
        <span className="text-xs text-[#ef4444]">{formatISK(factory.inputCostPerHour)}</span>
      </TableCell>
      <TableCell className="text-right border-[rgba(148,163,184,0.05)]">
        <span className="text-xs text-[#f59e0b]">{formatISK(factory.exportTaxPerHour + factory.importTaxPerHour)}</span>
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
            color="#10b981"
            icon={<TrendingUp className="w-4 h-4 text-[#10b981]" />}
          />
          <SummaryCard
            label="Costs / hr"
            value={formatISK(totals.input)}
            color="#ef4444"
            icon={<TrendingDown className="w-4 h-4 text-[#ef4444]" />}
          />
          <SummaryCard
            label="Taxes / hr"
            value={formatISK(totals.totalTax)}
            color="#f59e0b"
          />
          <SummaryCard
            label="Profit / hr"
            value={formatISK(totals.profit)}
            color={profitColor(totals.profit)}
            bold
          />
        </div>
        {/* Price source toggle */}
        <div className="flex rounded overflow-hidden border border-[rgba(148,163,184,0.2)]">
          {(['sell', 'buy', 'split'] as const).map((source) => (
            <button
              key={source}
              onClick={() => handlePriceSourceChange(source)}
              className={cn(
                'px-3 py-1 text-xs font-medium capitalize transition-colors',
                priceSource === source
                  ? 'bg-[rgba(0,212,255,0.1)] text-[#00d4ff] border-[rgba(0,212,255,0.3)]'
                  : 'text-[#64748b] hover:text-[#94a3b8] hover:bg-[rgba(148,163,184,0.05)]'
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
              <TableHead className="text-[#64748b] border-[rgba(148,163,184,0.1)] bg-[#0f1219] w-10" />
              <TableHead className="text-[#64748b] border-[rgba(148,163,184,0.1)] bg-[#0f1219]">Product</TableHead>
              <TableHead className="text-right text-[#64748b] border-[rgba(148,163,184,0.1)] bg-[#0f1219]">Rate</TableHead>
              <TableHead className="text-right text-[#64748b] border-[rgba(148,163,184,0.1)] bg-[#0f1219]">Revenue/hr</TableHead>
              <TableHead className="text-right text-[#64748b] border-[rgba(148,163,184,0.1)] bg-[#0f1219]">Costs/hr</TableHead>
              <TableHead className="text-right text-[#64748b] border-[rgba(148,163,184,0.1)] bg-[#0f1219]">Taxes/hr</TableHead>
              <TableHead className="text-right text-[#64748b] border-[rgba(148,163,184,0.1)] bg-[#0f1219]">Profit/hr</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {productGroups.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-[#64748b] border-[rgba(148,163,184,0.1)] py-8">
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
              <TableRow className="bg-[#0f1219]">
                <TableCell className="border-[rgba(148,163,184,0.1)]" />
                <TableCell className="text-[#e2e8f0] font-semibold border-[rgba(148,163,184,0.1)]">
                  Total ({productGroups.length} products)
                </TableCell>
                <TableCell className="border-[rgba(148,163,184,0.1)]" />
                <TableCell className="text-right text-[#10b981] font-semibold border-[rgba(148,163,184,0.1)]">
                  {formatISK(totals.output)}
                </TableCell>
                <TableCell className="text-right text-[#ef4444] font-semibold border-[rgba(148,163,184,0.1)]">
                  {formatISK(totals.input)}
                </TableCell>
                <TableCell className="text-right text-[#f59e0b] font-semibold border-[rgba(148,163,184,0.1)]">
                  {formatISK(totals.totalTax)}
                </TableCell>
                <TableCell
                  className="text-right font-bold border-[rgba(148,163,184,0.1)]"
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
        background: '#12151f',
        border: `1px solid ${color}25`,
      }}
    >
      <CardContent className="p-3 pb-3">
        <div className="flex items-center gap-1 mb-1">
          {icon}
          <span className="text-xs text-[#64748b]">{label}</span>
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

import { useState, useEffect, useCallback } from 'react';
import { useSession } from 'next-auth/react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { ChevronUp, ChevronDown, ChevronsUpDown, TrendingUp, Package, Clock, BarChart2 } from 'lucide-react';
import Image from 'next/image';
import Loading from '@industry-tool/components/loading';
import {
  HaulingRouteAnalytics,
  HaulingItemAnalytics,
  HaulingProfitDataPoint,
  HaulingRunDurationSummary,
} from '@industry-tool/client/data/models';
import { formatISK } from '@industry-tool/utils/formatting';

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

function regionName(id: number): string {
  return EVE_REGIONS[id] || `Region ${id}`;
}

type SortDir = 'asc' | 'desc';

function SortIcon({ active, dir }: { active: boolean; dir: SortDir }) {
  if (!active) return <ChevronsUpDown className="h-3 w-3 ml-1 inline opacity-40" />;
  return dir === 'asc'
    ? <ChevronUp className="h-3 w-3 ml-1 inline text-primary" />
    : <ChevronDown className="h-3 w-3 ml-1 inline text-primary" />;
}

// Distinct line colors for chart routes
const ROUTE_COLORS = [
  'var(--color-primary-cyan)',
  'var(--color-manufacturing-amber)',
  'var(--color-success-teal)',
  'var(--color-danger-rose)',
  '#a78bfa',
  '#f472b6',
  '#fb923c',
  '#34d399',
];

interface StatCardProps {
  label: string;
  value: string;
  sub?: string;
  icon: React.ReactNode;
}

function StatCard({ label, value, sub, icon }: StatCardProps) {
  return (
    <Card className="bg-background-panel border-overlay-subtle">
      <CardContent className="pt-5 pb-4 px-5">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-xs text-text-secondary mb-1">{label}</p>
            <p className="text-2xl font-bold text-text-data-value">{value}</p>
            {sub && <p className="text-xs text-text-muted mt-1">{sub}</p>}
          </div>
          <div className="text-text-secondary opacity-60">{icon}</div>
        </div>
      </CardContent>
    </Card>
  );
}

interface RouteTableProps {
  routes: HaulingRouteAnalytics[];
}

type RouteKey = keyof HaulingRouteAnalytics;

function RouteTable({ routes }: RouteTableProps) {
  const [sortKey, setSortKey] = useState<RouteKey>('totalProfitIsk');
  const [sortDir, setSortDir] = useState<SortDir>('desc');

  const handleSort = (key: RouteKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('desc');
    }
  };

  const sorted = [...routes].sort((a, b) => {
    const av = a[sortKey];
    const bv = b[sortKey];
    const cmp = typeof av === 'number' && typeof bv === 'number' ? av - bv : 0;
    return sortDir === 'asc' ? cmp : -cmp;
  });

  const th = (label: string, key: RouteKey, right?: boolean) => (
    <TableHead
      className={`font-bold text-text-emphasis cursor-pointer select-none${right ? ' text-right' : ''}`}
      onClick={() => handleSort(key)}
    >
      {label}
      <SortIcon active={sortKey === key} dir={sortDir} />
    </TableHead>
  );

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow className="bg-background-void border-overlay-subtle">
            <TableHead className="font-bold text-text-emphasis">Route</TableHead>
            {th('Runs', 'totalRuns', true)}
            {th('Total Profit', 'totalProfitIsk', true)}
            {th('Avg Profit/Run', 'avgProfitIsk', true)}
            {th('Avg Margin%', 'avgMarginPct', true)}
            {th('ISK/m³', 'avgIskPerM3', true)}
            {th('Best Run', 'bestRunProfitIsk', true)}
            {th('Worst Run', 'worstRunProfitIsk', true)}
          </TableRow>
        </TableHeader>
        <TableBody>
          {sorted.length === 0 && (
            <TableRow>
              <TableCell colSpan={8} className="text-center text-text-muted py-8">
                No route data yet. Complete some hauling runs to see analytics.
              </TableCell>
            </TableRow>
          )}
          {sorted.map((r, idx) => (
            <TableRow
              key={`${r.fromRegionId}-${r.toRegionId}`}
              className="border-overlay-subtle"
              style={{ backgroundColor: idx % 2 === 0 ? undefined : 'var(--color-overlay-subtle)' }}
            >
              <TableCell className="font-medium text-text-emphasis">
                {regionName(r.fromRegionId)} &rarr; {regionName(r.toRegionId)}
              </TableCell>
              <TableCell className="text-right text-text-secondary">{r.totalRuns}</TableCell>
              <TableCell className="text-right text-text-data-value font-semibold">{formatISK(r.totalProfitIsk)}</TableCell>
              <TableCell className="text-right text-text-secondary">{formatISK(r.avgProfitIsk)}</TableCell>
              <TableCell className="text-right text-text-secondary">{r.avgMarginPct.toFixed(1)}%</TableCell>
              <TableCell className="text-right text-text-secondary">{formatISK(r.avgIskPerM3)}</TableCell>
              <TableCell className="text-right" style={{ color: 'var(--color-success-teal)' }}>{formatISK(r.bestRunProfitIsk)}</TableCell>
              <TableCell className="text-right" style={{ color: r.worstRunProfitIsk < 0 ? 'var(--color-danger-rose)' : 'var(--color-text-secondary)' }}>
                {formatISK(r.worstRunProfitIsk)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

interface ProfitChartProps {
  timeseries: HaulingProfitDataPoint[];
  routes: HaulingRouteAnalytics[];
}

function buildRouteKey(fromRegionId: number, toRegionId: number): string {
  return `${fromRegionId}-${toRegionId}`;
}

function buildRouteLabel(fromRegionId: number, toRegionId: number): string {
  return `${regionName(fromRegionId)} → ${regionName(toRegionId)}`;
}

function ProfitChart({ timeseries, routes }: ProfitChartProps) {
  if (timeseries.length === 0) {
    return (
      <div className="flex items-center justify-center h-40 text-text-muted text-sm">
        No timeseries data yet.
      </div>
    );
  }

  // Collect all unique routes present in timeseries
  const routeKeys = Array.from(
    new Set(timeseries.map((d) => buildRouteKey(d.fromRegionId, d.toRegionId)))
  );

  // Build a map of all dates → { routeKey: profitIsk }
  const dateMap: Record<string, Record<string, number>> = {};
  for (const point of timeseries) {
    const key = buildRouteKey(point.fromRegionId, point.toRegionId);
    if (!dateMap[point.date]) dateMap[point.date] = {};
    dateMap[point.date][key] = (dateMap[point.date][key] || 0) + point.profitIsk;
  }

  const chartData = Object.entries(dateMap)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([date, vals]) => ({ date, ...vals }));

  // Map route key → color
  const colorByKey: Record<string, string> = {};
  routeKeys.forEach((k, i) => {
    colorByKey[k] = ROUTE_COLORS[i % ROUTE_COLORS.length];
  });

  // Build route labels from routes prop or timeseries
  const routeLabels: Record<string, string> = {};
  for (const r of routes) {
    const key = buildRouteKey(r.fromRegionId, r.toRegionId);
    routeLabels[key] = buildRouteLabel(r.fromRegionId, r.toRegionId);
  }
  for (const d of timeseries) {
    const key = buildRouteKey(d.fromRegionId, d.toRegionId);
    if (!routeLabels[key]) {
      routeLabels[key] = buildRouteLabel(d.fromRegionId, d.toRegionId);
    }
  }

  const iskFormatter = (v: number) => formatISK(v);

  return (
    <ResponsiveContainer width="100%" height={280}>
      <LineChart data={chartData} margin={{ top: 8, right: 16, left: 0, bottom: 8 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="var(--color-overlay-medium)" />
        <XAxis
          dataKey="date"
          tick={{ fill: 'var(--color-text-secondary)', fontSize: 11 }}
          tickLine={false}
          axisLine={{ stroke: 'var(--color-overlay-medium)' }}
        />
        <YAxis
          tickFormatter={iskFormatter}
          tick={{ fill: 'var(--color-text-secondary)', fontSize: 11 }}
          tickLine={false}
          axisLine={false}
          width={80}
        />
        <RechartsTooltip
          contentStyle={{
            backgroundColor: 'var(--color-bg-elevated)',
            border: '1px solid var(--color-overlay-medium)',
            borderRadius: '6px',
            fontSize: '12px',
            color: 'var(--color-text-primary)',
          }}
          formatter={(value: number | undefined, name: string | undefined) => [value !== undefined ? formatISK(value) : '—', name ? (routeLabels[name] || name) : ''] as [string, string]}
        />
        <Legend
          formatter={(value) => routeLabels[value] || value}
          wrapperStyle={{ fontSize: '11px', color: 'var(--color-text-secondary)' }}
        />
        {routeKeys.map((key) => (
          <Line
            key={key}
            type="monotone"
            dataKey={key}
            stroke={colorByKey[key]}
            strokeWidth={2}
            dot={false}
            connectNulls
          />
        ))}
      </LineChart>
    </ResponsiveContainer>
  );
}

interface ItemTableProps {
  items: HaulingItemAnalytics[];
}

type ItemKey = keyof HaulingItemAnalytics;

function ItemTable({ items }: ItemTableProps) {
  const [sortKey, setSortKey] = useState<ItemKey>('totalProfitIsk');
  const [sortDir, setSortDir] = useState<SortDir>('desc');

  const handleSort = (key: ItemKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('desc');
    }
  };

  const sorted = [...items].sort((a, b) => {
    const av = a[sortKey];
    const bv = b[sortKey];
    const cmp = typeof av === 'number' && typeof bv === 'number' ? av - bv : 0;
    return sortDir === 'asc' ? cmp : -cmp;
  });

  const th = (label: string, key: ItemKey, right?: boolean) => (
    <TableHead
      className={`font-bold text-text-emphasis cursor-pointer select-none${right ? ' text-right' : ''}`}
      onClick={() => handleSort(key)}
    >
      {label}
      <SortIcon active={sortKey === key} dir={sortDir} />
    </TableHead>
  );

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow className="bg-background-void border-overlay-subtle">
            <TableHead className="font-bold text-text-emphasis">Item</TableHead>
            {th('Runs', 'totalRuns', true)}
            {th('Qty Sold', 'totalQtySold', true)}
            {th('Total Profit', 'totalProfitIsk', true)}
            {th('Avg Margin%', 'avgMarginPct', true)}
          </TableRow>
        </TableHeader>
        <TableBody>
          {sorted.length === 0 && (
            <TableRow>
              <TableCell colSpan={5} className="text-center text-text-muted py-8">
                No item data yet. Sell items in completed hauling runs to see analytics.
              </TableCell>
            </TableRow>
          )}
          {sorted.map((item, idx) => (
            <TableRow
              key={item.typeId}
              className="border-overlay-subtle"
              style={{ backgroundColor: idx % 2 === 0 ? undefined : 'var(--color-overlay-subtle)' }}
            >
              <TableCell>
                <div className="flex items-center gap-2">
                  <Image
                    src={`https://images.evetech.net/types/${item.typeId}/icon`}
                    alt={item.typeName}
                    width={32}
                    height={32}
                    className="rounded"
                    unoptimized
                  />
                  <span className="text-text-emphasis font-medium">{item.typeName}</span>
                </div>
              </TableCell>
              <TableCell className="text-right text-text-secondary">{item.totalRuns}</TableCell>
              <TableCell className="text-right text-text-secondary">{item.totalQtySold.toLocaleString()}</TableCell>
              <TableCell className="text-right text-text-data-value font-semibold">{formatISK(item.totalProfitIsk)}</TableCell>
              <TableCell className="text-right text-text-secondary">{item.avgMarginPct.toFixed(1)}%</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

interface DurationSummaryProps {
  summary: HaulingRunDurationSummary;
}

function DurationSummary({ summary }: DurationSummaryProps) {
  const range = summary.maxDurationDays - summary.minDurationDays;
  const avgPct = range > 0
    ? ((summary.avgDurationDays - summary.minDurationDays) / range) * 100
    : 50;

  return (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="text-center">
          <p className="text-xs text-text-secondary mb-1">Fastest Run</p>
          <p className="text-xl font-bold text-text-data-value">{summary.minDurationDays.toFixed(1)} days</p>
        </div>
        <div className="text-center">
          <p className="text-xs text-text-secondary mb-1">Average Duration</p>
          <p className="text-xl font-bold text-text-data-value">{summary.avgDurationDays.toFixed(1)} days</p>
        </div>
        <div className="text-center">
          <p className="text-xs text-text-secondary mb-1">Slowest Run</p>
          <p className="text-xl font-bold text-text-data-value">{summary.maxDurationDays.toFixed(1)} days</p>
        </div>
      </div>

      {/* Range bar */}
      <div className="relative mt-2">
        <div className="w-full h-2 bg-background-elevated rounded-full overflow-visible">
          {/* Full range bar */}
          <div className="absolute inset-0 rounded-full" style={{ backgroundColor: 'var(--color-overlay-medium)' }} />
          {/* Average marker */}
          <div
            className="absolute top-1/2 -translate-y-1/2 w-3 h-3 rounded-full border-2"
            style={{
              left: `calc(${avgPct}% - 6px)`,
              backgroundColor: 'var(--color-primary-cyan)',
              borderColor: 'var(--color-bg-elevated)',
            }}
          />
        </div>
        <div className="flex justify-between mt-1">
          <span className="text-xs text-text-muted">{summary.minDurationDays.toFixed(1)}d</span>
          <span className="text-xs" style={{ color: 'var(--color-primary-cyan)' }}>
            avg {summary.avgDurationDays.toFixed(1)}d
          </span>
          <span className="text-xs text-text-muted">{summary.maxDurationDays.toFixed(1)}d</span>
        </div>
      </div>
    </div>
  );
}

export default function HaulingAnalytics() {
  const { data: session } = useSession();
  const [loading, setLoading] = useState(true);
  const [routes, setRoutes] = useState<HaulingRouteAnalytics[]>([]);
  const [items, setItems] = useState<HaulingItemAnalytics[]>([]);
  const [timeseries, setTimeseries] = useState<HaulingProfitDataPoint[]>([]);
  const [summary, setSummary] = useState<HaulingRunDurationSummary | null>(null);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      const [routesRes, itemsRes, tsRes, summaryRes] = await Promise.all([
        fetch('/api/hauling/analytics/routes'),
        fetch('/api/hauling/analytics/items'),
        fetch('/api/hauling/analytics/timeseries'),
        fetch('/api/hauling/analytics/summary'),
      ]);

      if (routesRes.ok) setRoutes(await routesRes.json());
      if (itemsRes.ok) setItems(await itemsRes.json());
      if (tsRes.ok) setTimeseries(await tsRes.json());
      if (summaryRes.ok) setSummary(await summaryRes.json());
    } catch (error) {
      console.error('Failed to fetch analytics:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (session) fetchAll();
  }, [session, fetchAll]);

  if (loading) return <Loading />;

  const totalRuns = summary?.totalCompletedRuns ?? 0;
  const totalProfit = summary?.totalProfitIsk ?? 0;
  const avgDuration = summary?.avgDurationDays ?? 0;
  const avgMargin = routes.length > 0
    ? routes.reduce((s, r) => s + r.avgMarginPct, 0) / routes.length
    : 0;

  return (
    <div className="flex flex-col gap-8">
      {/* Overview cards */}
      <section>
        <h2 className="text-lg font-semibold text-text-heading mb-4">Overview</h2>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <StatCard
            label="Completed Runs"
            value={String(totalRuns)}
            icon={<TrendingUp className="h-5 w-5" />}
          />
          <StatCard
            label="Total Profit"
            value={formatISK(totalProfit)}
            icon={<BarChart2 className="h-5 w-5" />}
          />
          <StatCard
            label="Avg Run Duration"
            value={`${avgDuration.toFixed(1)} days`}
            icon={<Clock className="h-5 w-5" />}
          />
          <StatCard
            label="Avg Margin"
            value={`${avgMargin.toFixed(1)}%`}
            sub="across all routes"
            icon={<Package className="h-5 w-5" />}
          />
        </div>
      </section>

      {/* Route analytics */}
      <section>
        <h2 className="text-lg font-semibold text-text-heading mb-4">Route Performance</h2>
        <Card className="bg-background-panel border-overlay-subtle mb-4">
          <RouteTable routes={routes} />
        </Card>

        <h3 className="text-sm font-semibold text-text-secondary mb-3">Profit Over Time by Route</h3>
        <Card className="bg-background-panel border-overlay-subtle">
          <CardContent className="pt-4">
            <ProfitChart timeseries={timeseries} routes={routes} />
          </CardContent>
        </Card>
      </section>

      {/* Item analytics */}
      <section>
        <h2 className="text-lg font-semibold text-text-heading mb-4">Item Performance</h2>
        <Card className="bg-background-panel border-overlay-subtle">
          <ItemTable items={items} />
        </Card>
      </section>

      {/* Run duration */}
      {summary && summary.totalCompletedRuns > 0 && (
        <section>
          <h2 className="text-lg font-semibold text-text-heading mb-4">Run Duration</h2>
          <Card className="bg-background-panel border-overlay-subtle">
            <CardContent className="pt-5 pb-5">
              <DurationSummary summary={summary} />
            </CardContent>
          </Card>
        </section>
      )}
    </div>
  );
}

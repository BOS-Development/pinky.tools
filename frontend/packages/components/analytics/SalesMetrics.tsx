import React, { useState, useEffect } from 'react';
import { Download, TrendingUp, ShoppingCart, DollarSign, Box, Users, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { cn } from '@/lib/utils';
import { formatISK, formatNumber, FONT_NUMERIC } from '@industry-tool/utils/formatting';

interface TimeSeriesData {
  date: string;
  revenue: number;
  transactions: number;
  quantitySold: number;
}

interface ItemSalesData {
  typeId: number;
  typeName: string;
  quantitySold: number;
  revenue: number;
  transactionCount: number;
  averagePricePerUnit: number;
}

interface SalesMetrics {
  totalRevenue: number;
  totalTransactions: number;
  totalQuantitySold: number;
  uniqueItemTypes: number;
  uniqueBuyers: number;
  timeSeriesData: TimeSeriesData[];
  topItems: ItemSalesData[];
}

type TimePeriod = '7d' | '30d' | '90d' | '1y' | 'all';

export default function SalesMetrics() {
  const [period, setPeriod] = useState<TimePeriod>('30d');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [metrics, setMetrics] = useState<SalesMetrics | null>(null);

  useEffect(() => {
    fetchMetrics();
  }, [period]);

  const fetchMetrics = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`/api/analytics/sales?period=${period}`);
      if (!response.ok) {
        throw new Error('Failed to fetch sales metrics');
      }
      const data = await response.json();
      setMetrics(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const exportToCSV = () => {
    if (!metrics) return;

    const csvRows = [
      ['Date', 'Revenue (ISK)', 'Transactions', 'Quantity Sold'],
      ...metrics.timeSeriesData.map(row => [
        row.date,
        row.revenue.toString(),
        row.transactions.toString(),
        row.quantitySold.toString(),
      ]),
    ];

    const csvContent = csvRows.map(row => row.join(',')).join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `sales-metrics-${period}-${new Date().toISOString().split('T')[0]}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-sm border border-rose-danger/30 bg-rose-danger/10 text-rose-danger px-4 py-3 mb-4 text-sm">
        {error}
      </div>
    );
  }

  if (!metrics) {
    return (
      <div className="rounded-sm border border-accent-blue/30 bg-accent-blue/10 text-blue-science px-4 py-3 mb-4 text-sm">
        No sales data available
      </div>
    );
  }

  const periods: TimePeriod[] = ['7d', '30d', '90d', '1y', 'all'];

  return (
    <div>
      {/* Header with time period filter and export button */}
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-text-emphasis">Sales Analytics</h1>
        <div className="flex gap-3 items-center">
          <div className="flex border border-overlay-strong rounded-sm overflow-hidden">
            {periods.map((p) => (
              <button
                key={p}
                onClick={() => setPeriod(p)}
                className={cn(
                  "px-3 py-1.5 text-sm border-r border-overlay-strong last:border-r-0",
                  period === p
                    ? "bg-primary text-background-void font-semibold"
                    : "bg-transparent text-text-secondary hover:text-text-emphasis hover:bg-status-neutral-tint"
                )}
              >
                {p === 'all' ? 'All Time' : p.toUpperCase()}
              </button>
            ))}
          </div>
          <Button variant="outline" onClick={exportToCSV}>
            <Download className="h-4 w-4 mr-1" />
            Export CSV
          </Button>
        </div>
      </div>

      {/* Info Alert */}
      <div className="rounded-sm border border-accent-blue/30 bg-accent-blue/8 text-blue-science px-4 py-3 mb-6 text-sm">
        Analytics only shows completed transactions. Pending sales (contract_created status) are not included in these metrics.
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-5 mb-8">
        <Card className="bg-background-panel border border-border-dim">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-text-secondary uppercase text-xs font-semibold tracking-wide">Total Revenue</span>
              <DollarSign className="h-5 w-5 text-text-secondary" />
            </div>
            <p className="text-2xl font-bold text-teal-success" style={{ fontFamily: FONT_NUMERIC }}>
              {formatISK(metrics.totalRevenue)}
            </p>
          </CardContent>
        </Card>

        <Card className="bg-background-panel border-overlay-subtle">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-text-secondary uppercase text-xs font-semibold tracking-wide">Transactions</span>
              <ShoppingCart className="h-5 w-5 text-category-violet" />
            </div>
            <p className="text-2xl font-bold text-[var(--color-data-value)]" style={{ fontFamily: FONT_NUMERIC }}>
              {formatNumber(metrics.totalTransactions)}
            </p>
          </CardContent>
        </Card>

        <Card className="bg-background-panel border-overlay-subtle">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-text-secondary uppercase text-xs font-semibold tracking-wide">Items Sold</span>
              <TrendingUp className="h-5 w-5 text-text-secondary" />
            </div>
            <p className="text-2xl font-bold text-[var(--color-data-value)]" style={{ fontFamily: FONT_NUMERIC }}>
              {formatNumber(metrics.totalQuantitySold)}
            </p>
          </CardContent>
        </Card>

        <Card className="bg-background-panel border-overlay-subtle">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-text-secondary uppercase text-xs font-semibold tracking-wide">Unique Items</span>
              <Box className="h-5 w-5 text-amber-manufacturing" />
            </div>
            <p className="text-2xl font-bold text-[var(--color-data-value)]" style={{ fontFamily: FONT_NUMERIC }}>
              {formatNumber(metrics.uniqueItemTypes)}
            </p>
          </CardContent>
        </Card>

        <Card className="bg-background-panel border-overlay-subtle">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-text-secondary uppercase text-xs font-semibold tracking-wide">Unique Buyers</span>
              <Users className="h-5 w-5 text-teal-success" />
            </div>
            <p className="text-2xl font-bold text-[var(--color-data-value)]" style={{ fontFamily: FONT_NUMERIC }}>
              {formatNumber(metrics.uniqueBuyers)}
            </p>
          </CardContent>
        </Card>

        <Card className="bg-background-panel border-overlay-subtle">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-text-secondary uppercase text-xs font-semibold tracking-wide">Avg / Transaction</span>
              <DollarSign className="h-5 w-5 text-text-secondary" />
            </div>
            <p className="text-2xl font-bold text-[var(--color-data-value)]" style={{ fontFamily: FONT_NUMERIC }}>
              {metrics.totalTransactions > 0
                ? formatISK(metrics.totalRevenue / metrics.totalTransactions)
                : '0 ISK'}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Top Selling Items */}
      <Card className="bg-background-panel border-overlay-subtle mb-8">
        <CardContent className="p-6">
          <h3 className="text-lg font-semibold text-text-emphasis mb-4">Top Selling Items</h3>
          <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
            <Table>
              <TableHeader>
                <TableRow className="bg-background-void">
                  <TableHead>Item Name</TableHead>
                  <TableHead className="text-right">Quantity Sold</TableHead>
                  <TableHead className="text-right">Revenue</TableHead>
                  <TableHead className="text-right">Transactions</TableHead>
                  <TableHead className="text-right">Avg Price/Unit</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {metrics.topItems.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center text-text-muted py-6">
                      No sales data available
                    </TableCell>
                  </TableRow>
                ) : (
                  metrics.topItems.map((item) => (
                    <TableRow key={item.typeId} className="bg-background-panel hover:bg-interactive-hover">
                      <TableCell>
                        <span className="font-medium text-text-emphasis">{item.typeName}</span>
                      </TableCell>
                      <TableCell className="text-right text-text-emphasis">
                        <span className="text-sm">{formatNumber(item.quantitySold)}</span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm font-semibold text-teal-success">{formatISK(item.revenue)}</span>
                      </TableCell>
                      <TableCell className="text-right text-text-emphasis">
                        <span className="text-sm">{formatNumber(item.transactionCount)}</span>
                      </TableCell>
                      <TableCell className="text-right text-text-emphasis">
                        <span className="text-sm">{formatISK(item.averagePricePerUnit)}</span>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* Sales Over Time */}
      <Card className="bg-background-panel border-overlay-subtle">
        <CardContent className="p-6">
          <h3 className="text-lg font-semibold text-text-emphasis mb-4">Sales Over Time</h3>
          <div className="overflow-x-auto rounded-sm border border-overlay-subtle">
            <Table>
              <TableHeader>
                <TableRow className="bg-background-void">
                  <TableHead>Date</TableHead>
                  <TableHead className="text-right">Revenue</TableHead>
                  <TableHead className="text-right">Transactions</TableHead>
                  <TableHead className="text-right">Quantity Sold</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {metrics.timeSeriesData.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center text-text-muted py-6">
                      No time series data available
                    </TableCell>
                  </TableRow>
                ) : (
                  metrics.timeSeriesData.map((row) => (
                    <TableRow key={row.date} className="bg-background-panel hover:bg-interactive-hover">
                      <TableCell>
                        <span className="font-medium text-text-emphasis">{row.date}</span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="font-semibold text-teal-success">{formatISK(row.revenue)}</span>
                      </TableCell>
                      <TableCell className="text-right text-text-emphasis">
                        <span className="text-sm">{formatNumber(row.transactions)}</span>
                      </TableCell>
                      <TableCell className="text-right text-text-emphasis">
                        <span className="text-sm">{formatNumber(row.quantitySold)}</span>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

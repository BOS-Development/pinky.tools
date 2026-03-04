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
        <Loader2 className="h-8 w-8 animate-spin text-[#00d4ff]" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-sm border border-[rgba(239,68,68,0.3)] bg-[rgba(239,68,68,0.1)] text-[#ef4444] px-4 py-3 mb-4 text-sm">
        {error}
      </div>
    );
  }

  if (!metrics) {
    return (
      <div className="rounded-sm border border-[rgba(59,130,246,0.3)] bg-[rgba(59,130,246,0.1)] text-[#60a5fa] px-4 py-3 mb-4 text-sm">
        No sales data available
      </div>
    );
  }

  const periods: TimePeriod[] = ['7d', '30d', '90d', '1y', 'all'];

  return (
    <div>
      {/* Header with time period filter and export button */}
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-[#e2e8f0]">Sales Analytics</h1>
        <div className="flex gap-3 items-center">
          <div className="flex border border-[rgba(148,163,184,0.2)] rounded-sm overflow-hidden">
            {periods.map((p) => (
              <button
                key={p}
                onClick={() => setPeriod(p)}
                className={cn(
                  "px-3 py-1.5 text-sm border-r border-[rgba(148,163,184,0.2)] last:border-r-0",
                  period === p
                    ? "bg-[#00d4ff] text-[#0a0e1a] font-semibold"
                    : "bg-transparent text-[#94a3b8] hover:text-[#e2e8f0] hover:bg-[rgba(148,163,184,0.08)]"
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
      <div className="rounded-sm border border-[rgba(59,130,246,0.3)] bg-[rgba(59,130,246,0.08)] text-[#60a5fa] px-4 py-3 mb-6 text-sm">
        Analytics only shows completed transactions. Pending sales (contract_created status) are not included in these metrics.
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-5 mb-8">
        <Card className="bg-[#12151f] border border-[rgba(0,212,255,0.08)]">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-[#94a3b8] uppercase text-xs font-semibold tracking-wide">Total Revenue</span>
              <DollarSign className="h-5 w-5 text-[#00d4ff]" />
            </div>
            <p className="text-2xl font-bold text-[#2dd4bf]" style={{ fontFamily: FONT_NUMERIC }}>
              {formatISK(metrics.totalRevenue)}
            </p>
          </CardContent>
        </Card>

        <Card className="bg-[#12151f] border-[rgba(148,163,184,0.1)]">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-[#94a3b8] uppercase text-xs font-semibold tracking-wide">Transactions</span>
              <ShoppingCart className="h-5 w-5 text-[#8b5cf6]" />
            </div>
            <p className="text-2xl font-bold text-[#e2e8f0]" style={{ fontFamily: FONT_NUMERIC }}>
              {formatNumber(metrics.totalTransactions)}
            </p>
          </CardContent>
        </Card>

        <Card className="bg-[#12151f] border-[rgba(148,163,184,0.1)]">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-[#94a3b8] uppercase text-xs font-semibold tracking-wide">Items Sold</span>
              <TrendingUp className="h-5 w-5 text-[#00d4ff]" />
            </div>
            <p className="text-2xl font-bold text-[#e2e8f0]" style={{ fontFamily: FONT_NUMERIC }}>
              {formatNumber(metrics.totalQuantitySold)}
            </p>
          </CardContent>
        </Card>

        <Card className="bg-[#12151f] border-[rgba(148,163,184,0.1)]">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-[#94a3b8] uppercase text-xs font-semibold tracking-wide">Unique Items</span>
              <Box className="h-5 w-5 text-[#f59e0b]" />
            </div>
            <p className="text-2xl font-bold text-[#e2e8f0]" style={{ fontFamily: FONT_NUMERIC }}>
              {formatNumber(metrics.uniqueItemTypes)}
            </p>
          </CardContent>
        </Card>

        <Card className="bg-[#12151f] border-[rgba(148,163,184,0.1)]">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-[#94a3b8] uppercase text-xs font-semibold tracking-wide">Unique Buyers</span>
              <Users className="h-5 w-5 text-[#10b981]" />
            </div>
            <p className="text-2xl font-bold text-[#e2e8f0]" style={{ fontFamily: FONT_NUMERIC }}>
              {formatNumber(metrics.uniqueBuyers)}
            </p>
          </CardContent>
        </Card>

        <Card className="bg-[#12151f] border-[rgba(148,163,184,0.1)]">
          <CardContent className="p-5">
            <div className="flex items-center justify-between mb-3">
              <span className="text-[#94a3b8] uppercase text-xs font-semibold tracking-wide">Avg / Transaction</span>
              <DollarSign className="h-5 w-5 text-[#00d4ff]" />
            </div>
            <p className="text-2xl font-bold text-[#e2e8f0]" style={{ fontFamily: FONT_NUMERIC }}>
              {metrics.totalTransactions > 0
                ? formatISK(metrics.totalRevenue / metrics.totalTransactions)
                : '0 ISK'}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Top Selling Items */}
      <Card className="bg-[#12151f] border-[rgba(148,163,184,0.1)] mb-8">
        <CardContent className="p-6">
          <h3 className="text-lg font-semibold text-[#e2e8f0] mb-4">Top Selling Items</h3>
          <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
            <Table>
              <TableHeader>
                <TableRow className="bg-[#0f1219]">
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
                    <TableCell colSpan={5} className="text-center text-[#64748b] py-6">
                      No sales data available
                    </TableCell>
                  </TableRow>
                ) : (
                  metrics.topItems.map((item) => (
                    <TableRow key={item.typeId} className="bg-[#12151f] hover:bg-[rgba(0,212,255,0.04)]">
                      <TableCell>
                        <span className="font-medium text-[#e2e8f0]">{item.typeName}</span>
                      </TableCell>
                      <TableCell className="text-right text-[#e2e8f0]">
                        <span className="text-sm">{formatNumber(item.quantitySold)}</span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="text-sm font-semibold text-[#10b981]">{formatISK(item.revenue)}</span>
                      </TableCell>
                      <TableCell className="text-right text-[#e2e8f0]">
                        <span className="text-sm">{formatNumber(item.transactionCount)}</span>
                      </TableCell>
                      <TableCell className="text-right text-[#e2e8f0]">
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
      <Card className="bg-[#12151f] border-[rgba(148,163,184,0.1)]">
        <CardContent className="p-6">
          <h3 className="text-lg font-semibold text-[#e2e8f0] mb-4">Sales Over Time</h3>
          <div className="overflow-x-auto rounded-sm border border-[rgba(148,163,184,0.1)]">
            <Table>
              <TableHeader>
                <TableRow className="bg-[#0f1219]">
                  <TableHead>Date</TableHead>
                  <TableHead className="text-right">Revenue</TableHead>
                  <TableHead className="text-right">Transactions</TableHead>
                  <TableHead className="text-right">Quantity Sold</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {metrics.timeSeriesData.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center text-[#64748b] py-6">
                      No time series data available
                    </TableCell>
                  </TableRow>
                ) : (
                  metrics.timeSeriesData.map((row) => (
                    <TableRow key={row.date} className="bg-[#12151f] hover:bg-[rgba(0,212,255,0.04)]">
                      <TableCell>
                        <span className="font-medium text-[#e2e8f0]">{row.date}</span>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className="font-semibold text-[#10b981]">{formatISK(row.revenue)}</span>
                      </TableCell>
                      <TableCell className="text-right text-[#e2e8f0]">
                        <span className="text-sm">{formatNumber(row.transactions)}</span>
                      </TableCell>
                      <TableCell className="text-right text-[#e2e8f0]">
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

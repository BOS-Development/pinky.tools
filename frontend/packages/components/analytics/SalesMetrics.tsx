import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  ButtonGroup,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  CircularProgress,
  Alert,
  Stack,
} from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart';
import AttachMoneyIcon from '@mui/icons-material/AttachMoney';
import CategoryIcon from '@mui/icons-material/Category';
import PeopleIcon from '@mui/icons-material/People';
import { formatISK, formatNumber } from '@industry-tool/utils/formatting';

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

    // Export time series data
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
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error" sx={{ mb: 2 }}>
        {error}
      </Alert>
    );
  }

  if (!metrics) {
    return (
      <Alert severity="info" sx={{ mb: 2 }}>
        No sales data available
      </Alert>
    );
  }

  return (
    <Box>
      {/* Header with time period filter and export button */}
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Sales Analytics</Typography>
        <Box display="flex" gap={2}>
          <ButtonGroup variant="outlined" size="small">
            {(['7d', '30d', '90d', '1y', 'all'] as TimePeriod[]).map((p) => (
              <Button
                key={p}
                onClick={() => setPeriod(p)}
                variant={period === p ? 'contained' : 'outlined'}
              >
                {p === 'all' ? 'All Time' : p.toUpperCase()}
              </Button>
            ))}
          </ButtonGroup>
          <Button
            variant="outlined"
            startIcon={<DownloadIcon />}
            onClick={exportToCSV}
          >
            Export CSV
          </Button>
        </Box>
      </Box>

      {/* Info Alert */}
      <Alert severity="info" sx={{ mb: 3 }}>
        Analytics only shows completed transactions. Pending sales (contract_created status) are not included in these metrics.
      </Alert>

      {/* Summary Cards */}
      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, 1fr)', md: 'repeat(3, 1fr)' },
          gap: 2.5,
          mb: 4,
        }}
      >
        <Card
          sx={{
            background: 'linear-gradient(135deg, rgba(59, 130, 246, 0.1) 0%, rgba(59, 130, 246, 0.05) 100%)',
            border: '1px solid rgba(59, 130, 246, 0.2)',
          }}
        >
          <CardContent sx={{ p: 3 }}>
            <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
              <Typography
                variant="body2"
                sx={{ color: '#94a3b8', textTransform: 'uppercase', fontSize: '0.75rem', fontWeight: 600, letterSpacing: '0.05em' }}
              >
                Total Revenue
              </Typography>
              <AttachMoneyIcon sx={{ color: '#3b82f6', fontSize: 20 }} />
            </Box>
            <Typography variant="h4" sx={{ fontWeight: 700, color: '#10b981', mb: 0.5 }}>
              {formatISK(metrics.totalRevenue)}
            </Typography>
          </CardContent>
        </Card>

        <Card>
          <CardContent sx={{ p: 3 }}>
            <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
              <Typography
                variant="body2"
                sx={{ color: '#94a3b8', textTransform: 'uppercase', fontSize: '0.75rem', fontWeight: 600, letterSpacing: '0.05em' }}
              >
                Transactions
              </Typography>
              <ShoppingCartIcon sx={{ color: '#8b5cf6', fontSize: 20 }} />
            </Box>
            <Typography variant="h4" sx={{ fontWeight: 700, mb: 0.5 }}>
              {formatNumber(metrics.totalTransactions)}
            </Typography>
          </CardContent>
        </Card>

        <Card>
          <CardContent sx={{ p: 3 }}>
            <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
              <Typography
                variant="body2"
                sx={{ color: '#94a3b8', textTransform: 'uppercase', fontSize: '0.75rem', fontWeight: 600, letterSpacing: '0.05em' }}
              >
                Items Sold
              </Typography>
              <TrendingUpIcon sx={{ color: '#3b82f6', fontSize: 20 }} />
            </Box>
            <Typography variant="h4" sx={{ fontWeight: 700, mb: 0.5 }}>
              {formatNumber(metrics.totalQuantitySold)}
            </Typography>
          </CardContent>
        </Card>

        <Card>
          <CardContent sx={{ p: 3 }}>
            <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
              <Typography
                variant="body2"
                sx={{ color: '#94a3b8', textTransform: 'uppercase', fontSize: '0.75rem', fontWeight: 600, letterSpacing: '0.05em' }}
              >
                Unique Items
              </Typography>
              <CategoryIcon sx={{ color: '#f59e0b', fontSize: 20 }} />
            </Box>
            <Typography variant="h4" sx={{ fontWeight: 700, mb: 0.5 }}>
              {formatNumber(metrics.uniqueItemTypes)}
            </Typography>
          </CardContent>
        </Card>

        <Card>
          <CardContent sx={{ p: 3 }}>
            <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
              <Typography
                variant="body2"
                sx={{ color: '#94a3b8', textTransform: 'uppercase', fontSize: '0.75rem', fontWeight: 600, letterSpacing: '0.05em' }}
              >
                Unique Buyers
              </Typography>
              <PeopleIcon sx={{ color: '#10b981', fontSize: 20 }} />
            </Box>
            <Typography variant="h4" sx={{ fontWeight: 700, mb: 0.5 }}>
              {formatNumber(metrics.uniqueBuyers)}
            </Typography>
          </CardContent>
        </Card>

        <Card>
          <CardContent sx={{ p: 3 }}>
            <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
              <Typography
                variant="body2"
                sx={{ color: '#94a3b8', textTransform: 'uppercase', fontSize: '0.75rem', fontWeight: 600, letterSpacing: '0.05em' }}
              >
                Avg / Transaction
              </Typography>
              <AttachMoneyIcon sx={{ color: '#3b82f6', fontSize: 20 }} />
            </Box>
            <Typography variant="h4" sx={{ fontWeight: 700, mb: 0.5 }}>
              {metrics.totalTransactions > 0
                ? formatISK(metrics.totalRevenue / metrics.totalTransactions)
                : '0 ISK'}
            </Typography>
          </CardContent>
        </Card>
      </Box>

      {/* Top Selling Items */}
      <Card sx={{ mb: 4 }}>
        <CardContent>
          <Typography variant="h6" mb={2}>
            Top Selling Items
          </Typography>
          <TableContainer component={Paper} variant="outlined">
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Item Name</TableCell>
                  <TableCell align="right">Quantity Sold</TableCell>
                  <TableCell align="right">Revenue</TableCell>
                  <TableCell align="right">Transactions</TableCell>
                  <TableCell align="right">Avg Price/Unit</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {metrics.topItems.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} align="center">
                      No sales data available
                    </TableCell>
                  </TableRow>
                ) : (
                  metrics.topItems.map((item) => (
                    <TableRow key={item.typeId} hover>
                      <TableCell>
                        <Typography variant="body2" sx={{ fontWeight: 500 }}>
                          {item.typeName}
                        </Typography>
                      </TableCell>
                      <TableCell align="right">
                        <Typography variant="body2">{formatNumber(item.quantitySold)}</Typography>
                      </TableCell>
                      <TableCell align="right">
                        <Typography variant="body2" sx={{ color: '#10b981', fontWeight: 600 }}>
                          {formatISK(item.revenue)}
                        </Typography>
                      </TableCell>
                      <TableCell align="right">
                        <Typography variant="body2">{formatNumber(item.transactionCount)}</Typography>
                      </TableCell>
                      <TableCell align="right">
                        <Typography variant="body2">{formatISK(item.averagePricePerUnit)}</Typography>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>

      {/* Sales Over Time */}
      <Card>
        <CardContent>
          <Typography variant="h6" mb={2}>
            Sales Over Time
          </Typography>
          <TableContainer component={Paper} variant="outlined">
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Date</TableCell>
                  <TableCell align="right">Revenue</TableCell>
                  <TableCell align="right">Transactions</TableCell>
                  <TableCell align="right">Quantity Sold</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {metrics.timeSeriesData.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={4} align="center">
                      No time series data available
                    </TableCell>
                  </TableRow>
                ) : (
                  metrics.timeSeriesData.map((row) => (
                    <TableRow key={row.date} hover>
                      <TableCell>
                        <Typography variant="body2" sx={{ fontWeight: 500 }}>
                          {row.date}
                        </Typography>
                      </TableCell>
                      <TableCell align="right">
                        <Typography variant="body2" sx={{ color: '#10b981', fontWeight: 600 }}>
                          {formatISK(row.revenue)}
                        </Typography>
                      </TableCell>
                      <TableCell align="right">
                        <Typography variant="body2">{formatNumber(row.transactions)}</Typography>
                      </TableCell>
                      <TableCell align="right">
                        <Typography variant="body2">{formatNumber(row.quantitySold)}</Typography>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </CardContent>
      </Card>
    </Box>
  );
}

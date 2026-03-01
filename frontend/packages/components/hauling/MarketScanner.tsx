import { useState, useEffect, useCallback, useRef } from 'react';
import { useSession } from 'next-auth/react';
import Image from 'next/image';
import Navbar from '@industry-tool/components/Navbar';
import Loading from '@industry-tool/components/loading';
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Skeleton from '@mui/material/Skeleton';
import Popover from '@mui/material/Popover';
import TextField from '@mui/material/TextField';
import ScannerIcon from '@mui/icons-material/Scanner';
import AddShoppingCartIcon from '@mui/icons-material/AddShoppingCart';
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

function getIndicatorColor(indicator: HaulingArbitrageRow['indicator']): 'warning' | 'info' | 'default' {
  switch (indicator) {
    case 'gap': return 'warning';
    case 'markup': return 'info';
    case 'thin': return 'default';
    default: return 'default';
  }
}

function getRowBgColor(indicator: HaulingArbitrageRow['indicator']): string | undefined {
  switch (indicator) {
    case 'gap': return 'rgba(245, 158, 11, 0.06)';
    case 'markup': return 'rgba(59, 130, 246, 0.04)';
    case 'thin': return undefined;
    default: return undefined;
  }
}

interface AddToRunState {
  row: HaulingArbitrageRow | null;
  anchorEl: HTMLElement | null;
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
    anchorEl: null,
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
        body: JSON.stringify({ regionId: sourceRegionId }),
      });
      // After triggering scan, fetch updated results
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

      setAddToRun({ row: null, anchorEl: null, selectedRunId: '', quantity: '1' });
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
          anchorEl: tableRef.current,
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
      <Container maxWidth={false} sx={{ mt: 4, mb: 4 }}>
        {/* Header */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 3 }}>
          <ScannerIcon fontSize="large" sx={{ color: '#3b82f6' }} />
          <Typography variant="h4" sx={{ fontWeight: 600 }}>
            Market Scanner
          </Typography>
        </Box>

        {/* Controls */}
        <Card sx={{ mb: 3 }}>
          <CardContent>
            <Box sx={{ display: 'flex', gap: 2, alignItems: 'center', flexWrap: 'wrap' }}>
              <FormControl sx={{ minWidth: 180 }}>
                <InputLabel>Source Region</InputLabel>
                <Select
                  value={sourceRegionId}
                  label="Source Region"
                  onChange={(e) => setSourceRegionId(Number(e.target.value))}
                >
                  {REGION_OPTIONS.map((r) => (
                    <MenuItem key={r.id} value={r.id}>{r.name}</MenuItem>
                  ))}
                </Select>
              </FormControl>
              <FormControl sx={{ minWidth: 180 }}>
                <InputLabel>Destination Region</InputLabel>
                <Select
                  value={destRegionId}
                  label="Destination Region"
                  onChange={(e) => setDestRegionId(Number(e.target.value))}
                >
                  {REGION_OPTIONS.map((r) => (
                    <MenuItem key={r.id} value={r.id}>{r.name}</MenuItem>
                  ))}
                </Select>
              </FormControl>
              <Button
                variant="outlined"
                onClick={handleFetchResults}
                disabled={loading}
              >
                {loading ? 'Loading...' : 'Load'}
              </Button>
              <Button
                variant="contained"
                startIcon={<ScannerIcon />}
                onClick={handleScan}
                disabled={scanning || loading}
              >
                {scanning ? 'Scanning...' : 'Scan'}
              </Button>
              {lastUpdated && (
                <Typography variant="caption" color="text.secondary" sx={{ ml: 1 }}>
                  Last updated: {new Date(lastUpdated).toLocaleString()}
                </Typography>
              )}
            </Box>
          </CardContent>
        </Card>

        {/* Results Table */}
        {loading ? (
          <Card>
            <CardContent sx={{ p: 0 }}>
              <TableContainer>
                <Table size="small">
                  <TableHead sx={{ backgroundColor: '#0f1219' }}>
                    <TableRow>
                      {['Item', 'Indicator', 'Net Profit/unit', 'm³', 'Days to Sell', 'Buy Price', 'Sell Price', 'Volume Available', 'Add to Run'].map((h) => (
                        <TableCell key={h} sx={{ fontWeight: 700 }}>{h}</TableCell>
                      ))}
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {Array.from({ length: 8 }).map((_, i) => (
                      <TableRow key={i}>
                        {Array.from({ length: 9 }).map((__, j) => (
                          <TableCell key={j}>
                            <Skeleton variant="text" />
                          </TableCell>
                        ))}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>
        ) : results.length === 0 ? (
          <Card>
            <CardContent>
              <Typography variant="h6" align="center" color="text.secondary" sx={{ py: 4 }}>
                No arbitrage opportunities found. Try scanning to refresh data.
              </Typography>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardContent sx={{ p: 0 }}>
              <TableContainer
                component={Paper}
                variant="outlined"
                ref={tableRef}
                tabIndex={0}
                onKeyDown={handleKeyDown}
                sx={{ outline: 'none' }}
              >
                <Table size="small" stickyHeader>
                  <TableHead>
                    <TableRow>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }}>Item</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }}>Indicator</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Net Profit/unit</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">m³</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Days to Sell</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Buy Price</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Sell Price</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Volume Available</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }}>Add to Run</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {results.map((row, idx) => {
                      const rowBg = getRowBgColor(row.indicator);
                      const isSelected = idx === selectedRowIndex;

                      return (
                        <TableRow
                          key={row.typeId}
                          hover
                          selected={isSelected}
                          onClick={() => setSelectedRowIndex(idx)}
                          sx={{
                            backgroundColor: rowBg,
                            cursor: 'pointer',
                            '&.Mui-selected': { backgroundColor: 'rgba(59, 130, 246, 0.12)' },
                          }}
                        >
                          <TableCell>
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                              <Image
                                src={getItemIconUrl(row.typeId, 32)}
                                alt={row.typeName}
                                width={24}
                                height={24}
                                style={{ borderRadius: 2 }}
                              />
                              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                                {row.typeName}
                              </Typography>
                            </Box>
                          </TableCell>
                          <TableCell>
                            <Chip
                              label={row.indicator}
                              color={getIndicatorColor(row.indicator)}
                              size="small"
                              variant="outlined"
                            />
                          </TableCell>
                          <TableCell align="right">
                            {row.netProfitIsk !== undefined ? (
                              <Typography
                                variant="body2"
                                sx={{ color: (row.netProfitIsk || 0) >= 0 ? '#10b981' : '#ef4444', fontWeight: 600 }}
                              >
                                {formatISK(row.netProfitIsk || 0)}
                              </Typography>
                            ) : '—'}
                          </TableCell>
                          <TableCell align="right">
                            {row.volumeM3 !== undefined ? formatNumber(row.volumeM3, 2) : '—'}
                          </TableCell>
                          <TableCell align="right">
                            {row.daysToSell !== undefined ? formatNumber(row.daysToSell, 1) : '—'}
                          </TableCell>
                          <TableCell align="right">
                            {row.buyPrice !== undefined ? formatISK(row.buyPrice) : '—'}
                          </TableCell>
                          <TableCell align="right">
                            {row.sellPrice !== undefined ? formatISK(row.sellPrice) : '—'}
                          </TableCell>
                          <TableCell align="right">
                            {row.volumeAvailable !== undefined ? formatNumber(row.volumeAvailable) : '—'}
                          </TableCell>
                          <TableCell>
                            <Button
                              size="small"
                              variant="outlined"
                              startIcon={<AddShoppingCartIcon fontSize="small" />}
                              onClick={(e) => {
                                e.stopPropagation();
                                setAddToRun({
                                  row,
                                  anchorEl: e.currentTarget,
                                  selectedRunId: runs[0]?.id || '',
                                  quantity: '1',
                                });
                              }}
                              disabled={runs.length === 0}
                            >
                              Add
                            </Button>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </TableContainer>
            </CardContent>
          </Card>
        )}

        {/* Add to Run Popover */}
        <Popover
          open={Boolean(addToRun.anchorEl)}
          anchorEl={addToRun.anchorEl}
          onClose={() => setAddToRun({ row: null, anchorEl: null, selectedRunId: '', quantity: '1' })}
          anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
        >
          <Box sx={{ p: 2, display: 'flex', flexDirection: 'column', gap: 2, minWidth: 240 }}>
            <Typography variant="subtitle2">
              {addToRun.row?.typeName}
            </Typography>
            <FormControl fullWidth size="small">
              <InputLabel>Hauling Run</InputLabel>
              <Select
                value={addToRun.selectedRunId}
                label="Hauling Run"
                onChange={(e) => setAddToRun({ ...addToRun, selectedRunId: e.target.value as number })}
              >
                {runs.map((run) => (
                  <MenuItem key={run.id} value={run.id}>{run.name}</MenuItem>
                ))}
              </Select>
            </FormControl>
            <TextField
              label="Quantity"
              type="number"
              size="small"
              value={addToRun.quantity}
              onChange={(e) => setAddToRun({ ...addToRun, quantity: e.target.value })}
              inputProps={{ min: 1 }}
            />
            <Button
              variant="contained"
              size="small"
              onClick={handleAddToRun}
              disabled={addToRun.selectedRunId === ''}
            >
              Add to Run
            </Button>
          </Box>
        </Popover>
      </Container>
    </>
  );
}

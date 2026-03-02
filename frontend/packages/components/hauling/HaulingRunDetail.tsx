import { useState, useEffect, useCallback } from 'react';
import { useSession } from 'next-auth/react';
import Image from 'next/image';
import Link from 'next/link';
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
import LinearProgress from '@mui/material/LinearProgress';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import Divider from '@mui/material/Divider';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import { HaulingRun, HaulingRunItem, HaulingArbitrageRow, HaulingRunPnlEntry, HaulingRunPnlSummary } from '@industry-tool/client/data/models';
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

const REGION_HUB_SYSTEMS: Record<number, number> = {
  10000002: 30000142, // The Forge → Jita
  10000043: 30002187, // Domain → Amarr
  10000032: 30002659, // Sinq Laison → Dodixie
  10000030: 30002510, // Heimatar → Rens
  10000042: 30002053, // Metropolis → Hek
  10000016: 30001161, // Lonetrek → Nourvukaiken
};

const STATUS_OPTIONS: HaulingRun['status'][] = [
  'PLANNING',
  'ACCUMULATING',
  'READY',
  'IN_TRANSIT',
  'SELLING',
  'COMPLETE',
  'CANCELLED',
];

function getStatusColor(status: HaulingRun['status']): 'primary' | 'warning' | 'success' | 'error' | 'default' | 'info' {
  switch (status) {
    case 'PLANNING': return 'primary';
    case 'ACCUMULATING': return 'warning';
    case 'READY': return 'success';
    case 'IN_TRANSIT': return 'info';
    case 'SELLING': return 'warning';
    case 'COMPLETE': return 'success';
    case 'CANCELLED': return 'error';
    default: return 'default';
  }
}

function getRowBgColor(fillPercent: number): string | undefined {
  if (fillPercent >= 100) return 'rgba(16, 185, 129, 0.08)';
  if (fillPercent < 50) return 'rgba(245, 158, 11, 0.06)';
  return undefined;
}

interface AddItemForm {
  typeId: string;
  typeName: string;
  quantityPlanned: string;
  buyPriceIsk: string;
  sellPriceIsk: string;
  volumeM3: string;
}

interface EditAcquiredForm {
  quantityAcquired: string;
}

interface PnlEntryForm {
  typeId: string;
  typeName: string;
  quantitySold: string;
  avgSellPriceIsk: string;
  totalCostIsk: string;
}

interface RouteSafetyData {
  jumps: number | null;
  fromName: string;
  toName: string;
  kills24h: number | null;
  loading: boolean;
}

interface HaulingRunDetailProps {
  runId: number;
}

export default function HaulingRunDetail({ runId }: HaulingRunDetailProps) {
  const { data: session } = useSession();
  const [run, setRun] = useState<HaulingRun | null>(null);
  const [loading, setLoading] = useState(true);
  const [scannerSuggestions, setScannerSuggestions] = useState<HaulingArbitrageRow[]>([]);

  // P&L state
  const [pnlEntries, setPnlEntries] = useState<HaulingRunPnlEntry[]>([]);
  const [pnlSummary, setPnlSummary] = useState<HaulingRunPnlSummary | null>(null);
  const [pnlDialogOpen, setPnlDialogOpen] = useState(false);
  const [editingPnlEntry, setEditingPnlEntry] = useState<HaulingRunPnlEntry | null>(null);
  const [pnlForm, setPnlForm] = useState<PnlEntryForm>({
    typeId: '',
    typeName: '',
    quantitySold: '',
    avgSellPriceIsk: '',
    totalCostIsk: '',
  });

  // Route safety state
  const [routeSafety, setRouteSafety] = useState<RouteSafetyData>({
    jumps: null,
    fromName: '',
    toName: '',
    kills24h: null,
    loading: false,
  });

  // Dialogs
  const [addItemOpen, setAddItemOpen] = useState(false);
  const [editAcquiredOpen, setEditAcquiredOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<HaulingRunItem | null>(null);
  const [statusMenuOpen, setStatusMenuOpen] = useState(false);

  const [addItemForm, setAddItemForm] = useState<AddItemForm>({
    typeId: '',
    typeName: '',
    quantityPlanned: '',
    buyPriceIsk: '',
    sellPriceIsk: '',
    volumeM3: '',
  });
  const [editAcquiredForm, setEditAcquiredForm] = useState<EditAcquiredForm>({ quantityAcquired: '' });
  const [submitting, setSubmitting] = useState(false);

  const fetchRun = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetch(`/api/hauling/runs/${runId}`);
      if (response.ok) {
        const data: HaulingRun = await response.json();
        setRun(data);

        // Fetch scanner suggestions for fill remaining capacity
        if (data.fromRegionId && data.toRegionId) {
          const params = new URLSearchParams({
            source_region_id: String(data.fromRegionId),
            dest_region_id: String(data.toRegionId),
          });
          const scanRes = await fetch(`/api/hauling/scanner?${params.toString()}`);
          if (scanRes.ok) {
            const scanData = await scanRes.json();
            setScannerSuggestions(Array.isArray(scanData) ? scanData : []);
          }
        }

        // Fetch P&L data for SELLING or COMPLETE runs
        if (data.status === 'SELLING' || data.status === 'COMPLETE') {
          const [pnlRes, summaryRes] = await Promise.allSettled([
            fetch(`/api/hauling/pnl?runId=${runId}`),
            fetch(`/api/hauling/pnl-summary?runId=${runId}`),
          ]);
          if (pnlRes.status === 'fulfilled' && pnlRes.value.ok) {
            const pnlData = await pnlRes.value.json();
            setPnlEntries(Array.isArray(pnlData) ? pnlData : []);
          }
          if (summaryRes.status === 'fulfilled' && summaryRes.value.ok) {
            const summaryData = await summaryRes.value.json();
            setPnlSummary(summaryData);
          }
        }
      }
    } catch (error) {
      console.error('Failed to fetch hauling run:', error);
    } finally {
      setLoading(false);
    }
  }, [runId]);

  const fetchRouteSafety = useCallback(async (fromRegionId: number, toRegionId: number) => {
    const fromSystemId = REGION_HUB_SYSTEMS[fromRegionId];
    const toSystemId = REGION_HUB_SYSTEMS[toRegionId];
    const fromHubName = EVE_REGIONS[fromRegionId] || `Region ${fromRegionId}`;
    const toHubName = EVE_REGIONS[toRegionId] || `Region ${toRegionId}`;

    setRouteSafety({ jumps: null, fromName: fromHubName, toName: toHubName, kills24h: null, loading: true });

    let jumps: number | null = null;
    let kills24h: number | null = null;

    if (fromSystemId && toSystemId) {
      try {
        const routeRes = await fetch(
          `https://esi.evetech.net/latest/route/${fromSystemId}/${toSystemId}/?flag=shortest`,
        );
        if (routeRes.ok) {
          const routeData = await routeRes.json();
          if (Array.isArray(routeData)) {
            jumps = routeData.length - 1;
          }
        }
      } catch {
        // ESI route unavailable — leave jumps as null
      }
    }

    try {
      const killsRes = await fetch(
        `https://zkillboard.com/api/kills/regionID/${fromRegionId}/pastSeconds/86400/`,
      );
      if (killsRes.ok) {
        const killsData = await killsRes.json();
        kills24h = Array.isArray(killsData) ? killsData.length : null;
      }
    } catch {
      // zKillboard unavailable — leave kills24h as null
    }

    setRouteSafety({ jumps, fromName: fromHubName, toName: toHubName, kills24h, loading: false });
  }, []);

  useEffect(() => {
    if (session) {
      fetchRun();
    }
  }, [session, fetchRun]);

  useEffect(() => {
    if (run && run.fromRegionId && run.toRegionId) {
      fetchRouteSafety(run.fromRegionId, run.toRegionId);
    }
  }, [run, fetchRouteSafety]);

  const handleStatusChange = async (newStatus: HaulingRun['status']) => {
    setStatusMenuOpen(false);
    if (!run) return;
    try {
      const response = await fetch(`/api/hauling/runs/${runId}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus }),
      });
      if (response.ok) {
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to update status:', error);
    }
  };

  const handleAddItem = async () => {
    if (!addItemForm.typeId || !addItemForm.quantityPlanned) return;
    setSubmitting(true);
    try {
      const body: Record<string, unknown> = {
        typeId: Number(addItemForm.typeId),
        typeName: addItemForm.typeName || `Type ${addItemForm.typeId}`,
        quantityPlanned: Number(addItemForm.quantityPlanned),
      };
      if (addItemForm.buyPriceIsk) body.buyPriceIsk = Number(addItemForm.buyPriceIsk);
      if (addItemForm.sellPriceIsk) body.sellPriceIsk = Number(addItemForm.sellPriceIsk);
      if (addItemForm.volumeM3) body.volumeM3 = Number(addItemForm.volumeM3);

      const response = await fetch(`/api/hauling/runs/${runId}/items`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (response.ok) {
        setAddItemOpen(false);
        setAddItemForm({ typeId: '', typeName: '', quantityPlanned: '', buyPriceIsk: '', sellPriceIsk: '', volumeM3: '' });
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to add item:', error);
    } finally {
      setSubmitting(false);
    }
  };

  const handleEditAcquired = async () => {
    if (!editingItem) return;
    setSubmitting(true);
    try {
      const response = await fetch(`/api/hauling/runs/${runId}/items/${editingItem.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ quantityAcquired: Number(editAcquiredForm.quantityAcquired) }),
      });
      if (response.ok) {
        setEditAcquiredOpen(false);
        setEditingItem(null);
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to update acquired qty:', error);
    } finally {
      setSubmitting(false);
    }
  };

  const handleRemoveItem = async (itemId: number) => {
    try {
      const response = await fetch(`/api/hauling/runs/${runId}/items/${itemId}`, {
        method: 'DELETE',
      });
      if (response.ok) {
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to remove item:', error);
    }
  };

  const handleAddScannerItem = async (row: HaulingArbitrageRow) => {
    if (!run) return;
    const remainingM3 = (run.maxVolumeM3 || 0) - totalUsedVolume;
    const qty = row.volumeM3 && row.volumeM3 > 0
      ? Math.max(1, Math.floor(remainingM3 / row.volumeM3))
      : 1;

    try {
      const body: Record<string, unknown> = {
        typeId: row.typeId,
        typeName: row.typeName,
        quantityPlanned: qty,
      };
      if (row.buyPrice) body.buyPriceIsk = row.buyPrice;
      if (row.sellPrice) body.sellPriceIsk = row.sellPrice;
      if (row.volumeM3) body.volumeM3 = row.volumeM3;

      const response = await fetch(`/api/hauling/runs/${runId}/items`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (response.ok) {
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to add scanner item:', error);
    }
  };

  const handleOpenPnlDialog = (entry?: HaulingRunPnlEntry) => {
    if (entry) {
      setEditingPnlEntry(entry);
      setPnlForm({
        typeId: String(entry.typeId),
        typeName: entry.typeName || '',
        quantitySold: String(entry.quantitySold),
        avgSellPriceIsk: entry.avgSellPriceIsk !== undefined ? String(entry.avgSellPriceIsk) : '',
        totalCostIsk: entry.totalCostIsk !== undefined ? String(entry.totalCostIsk) : '',
      });
    } else {
      setEditingPnlEntry(null);
      setPnlForm({ typeId: '', typeName: '', quantitySold: '', avgSellPriceIsk: '', totalCostIsk: '' });
    }
    setPnlDialogOpen(true);
  };

  const handleSubmitPnl = async () => {
    if (!pnlForm.typeId || !pnlForm.quantitySold) return;
    setSubmitting(true);
    try {
      const body: Record<string, unknown> = {
        typeId: Number(pnlForm.typeId),
        typeName: pnlForm.typeName || `Type ${pnlForm.typeId}`,
        quantitySold: Number(pnlForm.quantitySold),
      };
      if (pnlForm.avgSellPriceIsk) body.avgSellPriceIsk = Number(pnlForm.avgSellPriceIsk);
      if (pnlForm.totalCostIsk) body.totalCostIsk = Number(pnlForm.totalCostIsk);

      const response = await fetch(`/api/hauling/pnl?runId=${runId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (response.ok) {
        setPnlDialogOpen(false);
        setEditingPnlEntry(null);
        setPnlForm({ typeId: '', typeName: '', quantitySold: '', avgSellPriceIsk: '', totalCostIsk: '' });
        await fetchRun();
      }
    } catch (error) {
      console.error('Failed to submit P&L entry:', error);
    } finally {
      setSubmitting(false);
    }
  };

  if (!session) return null;
  if (loading) return <Loading />;
  if (!run) {
    return (
      <>
        <Navbar />
        <Container maxWidth={false} sx={{ mt: 4 }}>
          <Typography color="error">Run not found.</Typography>
        </Container>
      </>
    );
  }

  const items = run.items || [];
  const totalUsedVolume = items.reduce((sum, item) => sum + (item.volumeM3 || 0) * item.quantityPlanned, 0);
  const maxVol = run.maxVolumeM3 || 0;
  const volumePercent = maxVol > 0 ? Math.min((totalUsedVolume / maxVol) * 100, 100) : 0;

  const totalOutlay = items.reduce((sum, item) => sum + (item.buyPriceIsk || 0) * item.quantityPlanned, 0);
  const totalRevenue = items.reduce((sum, item) => sum + (item.sellPriceIsk || 0) * item.quantityPlanned, 0);
  const netProfit = totalRevenue - totalOutlay;
  const overallFill = items.length > 0
    ? items.reduce((sum, item) => sum + item.fillPercent, 0) / items.length
    : 0;

  const fromName = EVE_REGIONS[run.fromRegionId] || `Region ${run.fromRegionId}`;
  const toName = EVE_REGIONS[run.toRegionId] || `Region ${run.toRegionId}`;

  // Top 3 scanner suggestions that fit remaining capacity
  const remainingM3 = maxVol > 0 ? maxVol - totalUsedVolume : Infinity;
  const topSuggestions = scannerSuggestions
    .filter((row) => !row.volumeM3 || row.volumeM3 <= remainingM3)
    .sort((a, b) => {
      const profitPerM3A = a.netProfitIsk && a.volumeM3 ? a.netProfitIsk / a.volumeM3 : 0;
      const profitPerM3B = b.netProfitIsk && b.volumeM3 ? b.netProfitIsk / b.volumeM3 : 0;
      return profitPerM3B - profitPerM3A;
    })
    .slice(0, 3);

  return (
    <>
      <Navbar />
      <Container maxWidth={false} sx={{ mt: 4, mb: 4 }}>
        {/* Header */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 3 }}>
          <Button
            component={Link}
            href="/hauling"
            startIcon={<ArrowBackIcon />}
            variant="text"
            size="small"
          >
            Back
          </Button>
          <Typography variant="h4" sx={{ fontWeight: 600, flex: 1 }}>
            {run.name}
          </Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Chip
              label={run.status}
              color={getStatusColor(run.status)}
            />
            <FormControl size="small" sx={{ minWidth: 160 }}>
              <InputLabel>Change Status</InputLabel>
              <Select
                value=""
                label="Change Status"
                onChange={(e) => handleStatusChange(e.target.value as HaulingRun['status'])}
                displayEmpty
              >
                {STATUS_OPTIONS.map((s) => (
                  <MenuItem key={s} value={s}>{s}</MenuItem>
                ))}
              </Select>
            </FormControl>
          </Box>
        </Box>

        {/* Route & Capacity */}
        <Box sx={{ display: 'flex', gap: 2, mb: 3, flexWrap: 'wrap' }}>
          <Card sx={{ flex: 1, minWidth: 200 }}>
            <CardContent>
              <Typography variant="body2" color="text.secondary" gutterBottom>Route</Typography>
              <Typography variant="h6">{fromName} &rarr; {toName}</Typography>
            </CardContent>
          </Card>
          {maxVol > 0 && (
            <Card sx={{ flex: 2, minWidth: 300 }}>
              <CardContent>
                <Typography variant="body2" color="text.secondary" gutterBottom>Capacity</Typography>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                  <Box sx={{ flex: 1 }}>
                    <LinearProgress
                      variant="determinate"
                      value={volumePercent}
                      sx={{
                        height: 12,
                        borderRadius: 6,
                        backgroundColor: '#1e2a3a',
                        '& .MuiLinearProgress-bar': {
                          backgroundColor: volumePercent >= 95 ? '#10b981' : volumePercent >= 70 ? '#00d4ff' : '#f59e0b',
                        },
                      }}
                    />
                  </Box>
                  <Typography variant="body2" sx={{ minWidth: 140, textAlign: 'right' }}>
                    {formatNumber(totalUsedVolume, 1)} / {formatNumber(maxVol)} m³
                  </Typography>
                </Box>
              </CardContent>
            </Card>
          )}
        </Box>

        {/* Route Safety */}
        <Card sx={{ mb: 3 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>Route Safety</Typography>
            {routeSafety.loading ? (
              <Typography variant="body2" color="text.secondary">Loading route data...</Typography>
            ) : (
              <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap', alignItems: 'center' }}>
                <Box>
                  <Typography variant="body2" color="text.secondary">Shortest Route</Typography>
                  {routeSafety.jumps !== null ? (
                    <Typography variant="body1" sx={{ fontWeight: 600 }}>
                      {routeSafety.jumps} jumps ({fromName} &rarr; {toName})
                    </Typography>
                  ) : (
                    <Typography variant="body1">—</Typography>
                  )}
                </Box>
                <Box>
                  <Typography variant="body2" color="text.secondary">Source Region Kills (24h)</Typography>
                  {routeSafety.kills24h !== null ? (
                    <Typography
                      variant="body1"
                      sx={{
                        fontWeight: 600,
                        color: routeSafety.kills24h <= 5
                          ? '#10b981'
                          : routeSafety.kills24h <= 20
                          ? '#f59e0b'
                          : '#ef4444',
                      }}
                    >
                      {routeSafety.kills24h}
                    </Typography>
                  ) : (
                    <Typography variant="body1">—</Typography>
                  )}
                </Box>
              </Box>
            )}
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mt: 1 }}>
              Data from EVE ESI and zKillboard
            </Typography>
            {/* TODO(Phase 3): Undercut alert — show warning if current dest market price
                has dropped below item sell prices since this run was created. */}
          </CardContent>
        </Card>

        {/* Items Table */}
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
          <Typography variant="h6">Items</Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            size="small"
            onClick={() => setAddItemOpen(true)}
          >
            Add Item
          </Button>
        </Box>

        {items.length === 0 ? (
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="body1" align="center" color="text.secondary" sx={{ py: 2 }}>
                No items added yet.
              </Typography>
            </CardContent>
          </Card>
        ) : (
          <Card sx={{ mb: 3 }}>
            <CardContent sx={{ p: 0 }}>
              <TableContainer component={Paper} variant="outlined">
                <Table size="small" stickyHeader>
                  <TableHead>
                    <TableRow>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }}>Item</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Buy Price</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Sell Price</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Net Profit/unit</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Planned Qty</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Acquired</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700, width: 140 }}>Fill %</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Volume</TableCell>
                      <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }}>Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {items.map((item) => {
                      const rowBg = getRowBgColor(item.fillPercent);
                      const netProfitUnit = item.netProfitIsk !== undefined
                        ? item.netProfitIsk
                        : (item.sellPriceIsk && item.buyPriceIsk)
                          ? item.sellPriceIsk - item.buyPriceIsk
                          : undefined;

                      return (
                        <TableRow
                          key={item.id}
                          hover
                          sx={{ backgroundColor: rowBg }}
                        >
                          <TableCell>
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                              <Image
                                src={getItemIconUrl(item.typeId, 32)}
                                alt={item.typeName}
                                width={24}
                                height={24}
                                style={{ borderRadius: 2 }}
                              />
                              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                                {item.typeName}
                              </Typography>
                            </Box>
                          </TableCell>
                          <TableCell align="right">
                            {item.buyPriceIsk !== undefined ? formatISK(item.buyPriceIsk) : '—'}
                          </TableCell>
                          <TableCell align="right">
                            {item.sellPriceIsk !== undefined ? formatISK(item.sellPriceIsk) : '—'}
                          </TableCell>
                          <TableCell align="right">
                            {netProfitUnit !== undefined ? (
                              <Typography
                                variant="body2"
                                sx={{ color: netProfitUnit >= 0 ? '#10b981' : '#ef4444', fontWeight: 600 }}
                              >
                                {formatISK(netProfitUnit)}
                              </Typography>
                            ) : '—'}
                          </TableCell>
                          <TableCell align="right">{item.quantityPlanned.toLocaleString()}</TableCell>
                          <TableCell align="right">{item.quantityAcquired.toLocaleString()}</TableCell>
                          <TableCell>
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                              <Box sx={{ flex: 1 }}>
                                <LinearProgress
                                  variant="determinate"
                                  value={Math.min(item.fillPercent, 100)}
                                  sx={{
                                    height: 6,
                                    borderRadius: 3,
                                    backgroundColor: '#1e2a3a',
                                    '& .MuiLinearProgress-bar': {
                                      backgroundColor: item.fillPercent >= 100 ? '#10b981' : '#00d4ff',
                                    },
                                  }}
                                />
                              </Box>
                              <Typography variant="caption" sx={{ minWidth: 36, textAlign: 'right' }}>
                                {item.fillPercent.toFixed(0)}%
                              </Typography>
                            </Box>
                          </TableCell>
                          <TableCell align="right">
                            {item.volumeM3 !== undefined
                              ? `${formatNumber(item.volumeM3 * item.quantityPlanned, 1)} m³`
                              : '—'}
                          </TableCell>
                          <TableCell>
                            <Box sx={{ display: 'flex', gap: 0.5 }}>
                              <Tooltip title="Edit acquired qty">
                                <IconButton
                                  size="small"
                                  onClick={() => {
                                    setEditingItem(item);
                                    setEditAcquiredForm({ quantityAcquired: String(item.quantityAcquired) });
                                    setEditAcquiredOpen(true);
                                  }}
                                >
                                  <EditIcon fontSize="small" />
                                </IconButton>
                              </Tooltip>
                              <Tooltip title="Remove item">
                                <IconButton
                                  size="small"
                                  color="error"
                                  onClick={() => handleRemoveItem(item.id)}
                                >
                                  <DeleteIcon fontSize="small" />
                                </IconButton>
                              </Tooltip>
                            </Box>
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

        {/* Fill Remaining Capacity */}
        {topSuggestions.length > 0 && (
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
                <Typography variant="h6">Fill Remaining Capacity</Typography>
                <Button
                  component={Link}
                  href={`/hauling/scanner?dest_region=${run.toRegionId}&source_region=${run.fromRegionId}`}
                  endIcon={<OpenInNewIcon fontSize="small" />}
                  size="small"
                >
                  See all
                </Button>
              </Box>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                {topSuggestions.map((row) => {
                  const profitPerM3 = row.netProfitIsk && row.volumeM3
                    ? row.netProfitIsk / row.volumeM3
                    : undefined;
                  return (
                    <Box
                      key={row.typeId}
                      sx={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 2,
                        p: 1,
                        borderRadius: 1,
                        backgroundColor: '#12151f',
                      }}
                    >
                      <Image
                        src={getItemIconUrl(row.typeId, 32)}
                        alt={row.typeName}
                        width={24}
                        height={24}
                        style={{ borderRadius: 2 }}
                      />
                      <Typography variant="body2" sx={{ flex: 1, fontWeight: 600 }}>
                        {row.typeName}
                      </Typography>
                      {row.netProfitIsk !== undefined && (
                        <Typography variant="body2" sx={{ color: '#10b981' }}>
                          {formatISK(row.netProfitIsk)}/unit
                        </Typography>
                      )}
                      {profitPerM3 !== undefined && (
                        <Typography variant="caption" color="text.secondary">
                          {formatISK(profitPerM3)}/m³
                        </Typography>
                      )}
                      {row.volumeM3 !== undefined && (
                        <Typography variant="caption" color="text.secondary">
                          {formatNumber(row.volumeM3, 2)} m³
                        </Typography>
                      )}
                      <Button
                        size="small"
                        variant="outlined"
                        onClick={() => handleAddScannerItem(row)}
                      >
                        Add
                      </Button>
                    </Box>
                  );
                })}
              </Box>
            </CardContent>
          </Card>
        )}

        {/* P&L Section — shown for SELLING or COMPLETE runs */}
        {(run.status === 'SELLING' || run.status === 'COMPLETE') && (
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
                <Typography variant="h6">Profit &amp; Loss</Typography>
                <Button
                  variant="contained"
                  size="small"
                  startIcon={<AddIcon />}
                  onClick={() => handleOpenPnlDialog()}
                >
                  Enter P&amp;L
                </Button>
              </Box>
              {pnlEntries.length === 0 ? (
                <Typography variant="body2" color="text.secondary" align="center" sx={{ py: 2 }}>
                  No P&amp;L entries yet. Mark items as sold to track profit.
                </Typography>
              ) : (
                <TableContainer component={Paper} variant="outlined" sx={{ mb: 2 }}>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }}>Type</TableCell>
                        <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Qty Sold</TableCell>
                        <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Avg Sell Price</TableCell>
                        <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Revenue</TableCell>
                        <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Cost</TableCell>
                        <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }} align="right">Net Profit</TableCell>
                        <TableCell sx={{ backgroundColor: '#0f1219', fontWeight: 700 }}>Actions</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {pnlEntries.map((entry) => (
                        <TableRow key={entry.id} hover>
                          <TableCell>{entry.typeName || `Type ${entry.typeId}`}</TableCell>
                          <TableCell align="right">{formatNumber(entry.quantitySold)}</TableCell>
                          <TableCell align="right">
                            {entry.avgSellPriceIsk !== undefined ? formatISK(entry.avgSellPriceIsk) : '—'}
                          </TableCell>
                          <TableCell align="right" sx={{ color: '#10b981' }}>
                            {entry.totalRevenueIsk !== undefined ? formatISK(entry.totalRevenueIsk) : '—'}
                          </TableCell>
                          <TableCell align="right" sx={{ color: '#ef4444' }}>
                            {entry.totalCostIsk !== undefined ? formatISK(entry.totalCostIsk) : '—'}
                          </TableCell>
                          <TableCell align="right">
                            {entry.netProfitIsk !== undefined ? (
                              <Typography
                                variant="body2"
                                sx={{ color: entry.netProfitIsk >= 0 ? '#10b981' : '#ef4444', fontWeight: 600 }}
                              >
                                {formatISK(entry.netProfitIsk)}
                              </Typography>
                            ) : '—'}
                          </TableCell>
                          <TableCell>
                            <Tooltip title="Edit P&L entry">
                              <IconButton size="small" onClick={() => handleOpenPnlDialog(entry)}>
                                <EditIcon fontSize="small" />
                              </IconButton>
                            </Tooltip>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              )}
              {pnlSummary && (
                <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap', mt: 1 }}>
                  <Box>
                    <Typography variant="body2" color="text.secondary">Total Revenue</Typography>
                    <Typography variant="h6" sx={{ color: '#10b981' }}>
                      {formatISK(pnlSummary.totalRevenueIsk)}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">Total Cost</Typography>
                    <Typography variant="h6" sx={{ color: '#ef4444' }}>
                      {formatISK(pnlSummary.totalCostIsk)}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">Net Profit</Typography>
                    <Typography variant="h6" sx={{ color: pnlSummary.netProfitIsk >= 0 ? '#10b981' : '#ef4444' }}>
                      {formatISK(pnlSummary.netProfitIsk)}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">Margin %</Typography>
                    <Typography variant="h6">
                      {pnlSummary.marginPct.toFixed(1)}%
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">Items Sold</Typography>
                    <Typography variant="h6">{formatNumber(pnlSummary.itemsSold)}</Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">Items Pending</Typography>
                    <Typography variant="h6">{formatNumber(pnlSummary.itemsPending)}</Typography>
                  </Box>
                </Box>
              )}
            </CardContent>
          </Card>
        )}

        {/* Stats Footer */}
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>Run Summary</Typography>
            <Divider sx={{ mb: 2 }} />
            <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap' }}>
              <Box>
                <Typography variant="body2" color="text.secondary">Total ISK Outlay</Typography>
                <Typography variant="h6" sx={{ color: '#ef4444' }}>
                  {formatISK(totalOutlay)}
                </Typography>
              </Box>
              {run.status === 'COMPLETE' && pnlSummary ? (
                <>
                  <Box>
                    <Typography variant="body2" color="text.secondary">Actual Revenue</Typography>
                    <Typography variant="h6" sx={{ color: '#10b981' }}>
                      {formatISK(pnlSummary.totalRevenueIsk)}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">Actual Net Profit</Typography>
                    <Typography variant="h6" sx={{ color: pnlSummary.netProfitIsk >= 0 ? '#10b981' : '#ef4444' }}>
                      {formatISK(pnlSummary.netProfitIsk)}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">Margin %</Typography>
                    <Typography variant="h6">
                      {pnlSummary.marginPct.toFixed(1)}%
                    </Typography>
                  </Box>
                </>
              ) : (
                <>
                  <Box>
                    <Typography variant="body2" color="text.secondary">Total ISK Revenue</Typography>
                    <Typography variant="h6" sx={{ color: '#10b981' }}>
                      {formatISK(totalRevenue)}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">Net Profit</Typography>
                    <Typography variant="h6" sx={{ color: netProfit >= 0 ? '#10b981' : '#ef4444' }}>
                      {formatISK(netProfit)}
                    </Typography>
                  </Box>
                </>
              )}
              <Box>
                <Typography variant="body2" color="text.secondary">Fill % Overall</Typography>
                <Typography variant="h6">
                  {overallFill.toFixed(1)}%
                </Typography>
              </Box>
            </Box>
          </CardContent>
        </Card>
      </Container>

      {/* Add Item Dialog */}
      <Dialog open={addItemOpen} onClose={() => setAddItemOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Item</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}>
            <TextField
              label="Type ID"
              type="number"
              value={addItemForm.typeId}
              onChange={(e) => setAddItemForm({ ...addItemForm, typeId: e.target.value })}
              fullWidth
              required
              autoFocus
            />
            <TextField
              label="Item Name"
              value={addItemForm.typeName}
              onChange={(e) => setAddItemForm({ ...addItemForm, typeName: e.target.value })}
              fullWidth
              required
            />
            <TextField
              label="Planned Quantity"
              type="number"
              value={addItemForm.quantityPlanned}
              onChange={(e) => setAddItemForm({ ...addItemForm, quantityPlanned: e.target.value })}
              fullWidth
              required
              inputProps={{ min: 1 }}
            />
            <TextField
              label="Buy Price (ISK)"
              type="number"
              value={addItemForm.buyPriceIsk}
              onChange={(e) => setAddItemForm({ ...addItemForm, buyPriceIsk: e.target.value })}
              fullWidth
              inputProps={{ min: 0 }}
            />
            <TextField
              label="Sell Price (ISK)"
              type="number"
              value={addItemForm.sellPriceIsk}
              onChange={(e) => setAddItemForm({ ...addItemForm, sellPriceIsk: e.target.value })}
              fullWidth
              inputProps={{ min: 0 }}
            />
            <TextField
              label="Volume (m³)"
              type="number"
              value={addItemForm.volumeM3}
              onChange={(e) => setAddItemForm({ ...addItemForm, volumeM3: e.target.value })}
              fullWidth
              inputProps={{ min: 0, step: '0.01' }}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setAddItemOpen(false)}>Cancel</Button>
          <Button
            onClick={handleAddItem}
            variant="contained"
            disabled={!addItemForm.typeId || !addItemForm.quantityPlanned || !addItemForm.typeName || submitting}
          >
            {submitting ? 'Adding...' : 'Add'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Edit Acquired Dialog */}
      <Dialog open={editAcquiredOpen} onClose={() => setEditAcquiredOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>Update Acquired Quantity</DialogTitle>
        <DialogContent>
          <Box sx={{ mt: 1 }}>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              {editingItem?.typeName}
            </Typography>
            <TextField
              label="Quantity Acquired"
              type="number"
              value={editAcquiredForm.quantityAcquired}
              onChange={(e) => setEditAcquiredForm({ quantityAcquired: e.target.value })}
              fullWidth
              autoFocus
              inputProps={{ min: 0 }}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEditAcquiredOpen(false)}>Cancel</Button>
          <Button
            onClick={handleEditAcquired}
            variant="contained"
            disabled={editAcquiredForm.quantityAcquired === '' || submitting}
          >
            {submitting ? 'Saving...' : 'Save'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* P&L Entry Dialog */}
      <Dialog open={pnlDialogOpen} onClose={() => setPnlDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{editingPnlEntry ? 'Edit P&L Entry' : 'Enter P&L'}</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}>
            <FormControl fullWidth required>
              <InputLabel>Item</InputLabel>
              <Select
                value={pnlForm.typeId}
                label="Item"
                onChange={(e) => {
                  const selectedItem = items.find((item) => String(item.typeId) === e.target.value);
                  setPnlForm({
                    ...pnlForm,
                    typeId: e.target.value as string,
                    typeName: selectedItem ? selectedItem.typeName : pnlForm.typeName,
                  });
                }}
              >
                {items.map((item) => (
                  <MenuItem key={item.typeId} value={String(item.typeId)}>
                    {item.typeName}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <TextField
              label="Qty Sold"
              type="number"
              value={pnlForm.quantitySold}
              onChange={(e) => setPnlForm({ ...pnlForm, quantitySold: e.target.value })}
              fullWidth
              required
              inputProps={{ min: 1 }}
            />
            <TextField
              label="Avg Sell Price (ISK)"
              type="number"
              value={pnlForm.avgSellPriceIsk}
              onChange={(e) => setPnlForm({ ...pnlForm, avgSellPriceIsk: e.target.value })}
              fullWidth
              inputProps={{ min: 0 }}
            />
            <TextField
              label="Total Cost (ISK, optional)"
              type="number"
              value={pnlForm.totalCostIsk}
              onChange={(e) => setPnlForm({ ...pnlForm, totalCostIsk: e.target.value })}
              fullWidth
              inputProps={{ min: 0 }}
              helperText="Total buy cost for this item lot"
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setPnlDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={handleSubmitPnl}
            variant="contained"
            disabled={!pnlForm.typeId || !pnlForm.quantitySold || submitting}
          >
            {submitting ? 'Saving...' : 'Save'}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}

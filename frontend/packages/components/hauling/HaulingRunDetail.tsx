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
import { HaulingRun, HaulingRunItem, HaulingArbitrageRow } from '@industry-tool/client/data/models';
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

interface HaulingRunDetailProps {
  runId: number;
}

export default function HaulingRunDetail({ runId }: HaulingRunDetailProps) {
  const { data: session } = useSession();
  const [run, setRun] = useState<HaulingRun | null>(null);
  const [loading, setLoading] = useState(true);
  const [scannerSuggestions, setScannerSuggestions] = useState<HaulingArbitrageRow[]>([]);

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
      }
    } catch (error) {
      console.error('Failed to fetch hauling run:', error);
    } finally {
      setLoading(false);
    }
  }, [runId]);

  useEffect(() => {
    if (session) {
      fetchRun();
    }
  }, [session, fetchRun]);

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
                          backgroundColor: volumePercent >= 95 ? '#10b981' : volumePercent >= 70 ? '#3b82f6' : '#f59e0b',
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
                                      backgroundColor: item.fillPercent >= 100 ? '#10b981' : '#3b82f6',
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
    </>
  );
}

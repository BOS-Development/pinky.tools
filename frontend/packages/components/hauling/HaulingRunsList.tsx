import { useState, useEffect, useCallback } from 'react';
import { useSession } from 'next-auth/react';
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
import AddIcon from '@mui/icons-material/Add';
import LocalShippingIcon from '@mui/icons-material/LocalShipping';
import Link from 'next/link';
import { HaulingRun } from '@industry-tool/client/data/models';
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

const REGION_OPTIONS = Object.entries(EVE_REGIONS).map(([id, name]) => ({
  id: Number(id),
  name,
}));

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

function computeRunVolume(run: HaulingRun): { used: number; total: number; percent: number } {
  const items = run.items || [];
  const used = items.reduce((sum, item) => {
    return sum + (item.volumeM3 || 0) * item.quantityPlanned;
  }, 0);
  const total = run.maxVolumeM3 || 0;
  const percent = total > 0 ? Math.min((used / total) * 100, 100) : 0;
  return { used, total, percent };
}

function computeOverallFill(run: HaulingRun): number {
  const items = run.items || [];
  if (items.length === 0) return 0;
  const total = items.reduce((sum, item) => sum + item.fillPercent, 0);
  return total / items.length;
}

interface NewRunForm {
  name: string;
  fromRegionId: number | '';
  toRegionId: number | '';
  maxVolumeM3: string;
  haulThresholdIsk: string;
}

export default function HaulingRunsList() {
  const { data: session } = useSession();
  const [runs, setRuns] = useState<HaulingRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form, setForm] = useState<NewRunForm>({
    name: '',
    fromRegionId: '',
    toRegionId: '',
    maxVolumeM3: '',
    haulThresholdIsk: '',
  });

  const fetchRuns = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/hauling/runs');
      if (response.ok) {
        const data = await response.json();
        setRuns(Array.isArray(data) ? data : []);
      }
    } catch (error) {
      console.error('Failed to fetch hauling runs:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (session) {
      fetchRuns();
    }
  }, [session, fetchRuns]);

  const handleCreateRun = async () => {
    if (!form.name || form.fromRegionId === '' || form.toRegionId === '') return;

    setSubmitting(true);
    try {
      const body: Record<string, unknown> = {
        name: form.name,
        fromRegionId: form.fromRegionId,
        toRegionId: form.toRegionId,
      };
      if (form.maxVolumeM3) body.maxVolumeM3 = Number(form.maxVolumeM3);
      if (form.haulThresholdIsk) body.haulThresholdIsk = Number(form.haulThresholdIsk);

      const response = await fetch('/api/hauling/runs', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      if (response.ok) {
        setDialogOpen(false);
        setForm({ name: '', fromRegionId: '', toRegionId: '', maxVolumeM3: '', haulThresholdIsk: '' });
        await fetchRuns();
      }
    } catch (error) {
      console.error('Failed to create hauling run:', error);
    } finally {
      setSubmitting(false);
    }
  };

  const statusCounts = runs.reduce<Record<string, number>>((acc, run) => {
    acc[run.status] = (acc[run.status] || 0) + 1;
    return acc;
  }, {});

  if (!session) return null;
  if (loading) return <Loading />;

  return (
    <>
      <Navbar />
      <Container maxWidth={false} sx={{ mt: 4, mb: 4 }}>
        {/* Header */}
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <LocalShippingIcon fontSize="large" sx={{ color: '#3b82f6' }} />
            <Typography variant="h4" sx={{ fontWeight: 600 }}>
              Hauling Runs
            </Typography>
          </Box>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setDialogOpen(true)}
          >
            New Run
          </Button>
        </Box>

        {/* Status summary chips */}
        {runs.length > 0 && (
          <Box sx={{ display: 'flex', gap: 1, mb: 3, flexWrap: 'wrap' }}>
            {Object.entries(statusCounts).map(([status, count]) => (
              <Chip
                key={status}
                label={`${status}: ${count}`}
                color={getStatusColor(status as HaulingRun['status'])}
                size="small"
                variant="outlined"
              />
            ))}
          </Box>
        )}

        {/* Runs Table */}
        {runs.length === 0 ? (
          <Card>
            <CardContent>
              <Typography variant="h6" align="center" color="text.secondary" sx={{ py: 4 }}>
                No hauling runs yet. Create your first run to get started.
              </Typography>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardContent sx={{ p: 0 }}>
              <TableContainer component={Paper} variant="outlined">
                <Table size="small">
                  <TableHead sx={{ backgroundColor: '#0f1219' }}>
                    <TableRow>
                      <TableCell sx={{ fontWeight: 700 }}>Name</TableCell>
                      <TableCell sx={{ fontWeight: 700 }}>Status</TableCell>
                      <TableCell sx={{ fontWeight: 700 }}>Route</TableCell>
                      <TableCell sx={{ fontWeight: 700 }} align="right">Volume Used</TableCell>
                      <TableCell sx={{ fontWeight: 700 }} align="right">Items</TableCell>
                      <TableCell sx={{ fontWeight: 700, width: 160 }}>Progress</TableCell>
                      <TableCell sx={{ fontWeight: 700 }}>Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {runs.map((run, idx) => {
                      const { used, total, percent } = computeRunVolume(run);
                      const fillPercent = computeOverallFill(run);
                      const fromName = EVE_REGIONS[run.fromRegionId] || `Region ${run.fromRegionId}`;
                      const toName = EVE_REGIONS[run.toRegionId] || `Region ${run.toRegionId}`;
                      const itemCount = (run.items || []).length;

                      return (
                        <TableRow
                          key={run.id}
                          hover
                          sx={{
                            '&:nth-of-type(odd)': { backgroundColor: 'action.hover' },
                          }}
                        >
                          <TableCell sx={{ fontWeight: 600 }}>{run.name}</TableCell>
                          <TableCell>
                            <Chip
                              label={run.status}
                              color={getStatusColor(run.status)}
                              size="small"
                            />
                          </TableCell>
                          <TableCell>
                            <Typography variant="body2">
                              {fromName} &rarr; {toName}
                            </Typography>
                          </TableCell>
                          <TableCell align="right">
                            {total > 0
                              ? `${used.toLocaleString(undefined, { maximumFractionDigits: 1 })} / ${total.toLocaleString()} m³`
                              : used > 0
                              ? `${used.toLocaleString(undefined, { maximumFractionDigits: 1 })} m³`
                              : '—'}
                          </TableCell>
                          <TableCell align="right">{itemCount}</TableCell>
                          <TableCell>
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                              <Box sx={{ flex: 1 }}>
                                <LinearProgress
                                  variant="determinate"
                                  value={fillPercent}
                                  sx={{
                                    height: 8,
                                    borderRadius: 4,
                                    backgroundColor: '#1e2a3a',
                                    '& .MuiLinearProgress-bar': {
                                      backgroundColor: fillPercent >= 100 ? '#10b981' : fillPercent >= 50 ? '#3b82f6' : '#f59e0b',
                                    },
                                  }}
                                />
                              </Box>
                              <Typography variant="caption" sx={{ minWidth: 36, textAlign: 'right' }}>
                                {fillPercent.toFixed(0)}%
                              </Typography>
                            </Box>
                          </TableCell>
                          <TableCell>
                            <Button
                              component={Link}
                              href={`/hauling/${run.id}`}
                              size="small"
                              variant="outlined"
                            >
                              View
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
      </Container>

      {/* New Run Dialog */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>New Hauling Run</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}>
            <TextField
              label="Name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              fullWidth
              required
              autoFocus
            />
            <FormControl fullWidth required>
              <InputLabel>From Region</InputLabel>
              <Select
                value={form.fromRegionId}
                label="From Region"
                onChange={(e) => setForm({ ...form, fromRegionId: e.target.value as number })}
              >
                {REGION_OPTIONS.map((r) => (
                  <MenuItem key={r.id} value={r.id}>{r.name}</MenuItem>
                ))}
              </Select>
            </FormControl>
            <FormControl fullWidth required>
              <InputLabel>To Region</InputLabel>
              <Select
                value={form.toRegionId}
                label="To Region"
                onChange={(e) => setForm({ ...form, toRegionId: e.target.value as number })}
              >
                {REGION_OPTIONS.map((r) => (
                  <MenuItem key={r.id} value={r.id}>{r.name}</MenuItem>
                ))}
              </Select>
            </FormControl>
            <TextField
              label="Max Volume (m³)"
              type="number"
              value={form.maxVolumeM3}
              onChange={(e) => setForm({ ...form, maxVolumeM3: e.target.value })}
              fullWidth
              inputProps={{ min: 0 }}
            />
            <TextField
              label="Haul Threshold ISK"
              type="number"
              value={form.haulThresholdIsk}
              onChange={(e) => setForm({ ...form, haulThresholdIsk: e.target.value })}
              fullWidth
              inputProps={{ min: 0 }}
              helperText="Minimum net profit to consider hauling"
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={handleCreateRun}
            variant="contained"
            disabled={!form.name || form.fromRegionId === '' || form.toRegionId === '' || submitting}
          >
            {submitting ? 'Creating...' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}

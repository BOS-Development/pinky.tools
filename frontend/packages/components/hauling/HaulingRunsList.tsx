import { useState, useEffect, useCallback } from 'react';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/router';
import Navbar from '@industry-tool/components/Navbar';
import Loading from '@industry-tool/components/loading';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Plus, Truck } from 'lucide-react';
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

function StatusBadge({ status }: { status: string }) {
  const config: Record<string, { color: string; bg: string }> = {
    PLANNING: { color: 'var(--color-primary-cyan)', bg: 'rgba(0, 212, 255, 0.1)' },
    ACCUMULATING: { color: 'var(--color-manufacturing-amber)', bg: 'rgba(245, 158, 11, 0.1)' },
    READY: { color: 'var(--color-success-teal)', bg: 'rgba(16, 185, 129, 0.1)' },
    IN_TRANSIT: { color: '#38bdf8', bg: 'rgba(56, 189, 248, 0.1)' },
    SELLING: { color: 'var(--color-manufacturing-amber)', bg: 'rgba(245, 158, 11, 0.1)' },
    COMPLETE: { color: 'var(--color-success-teal)', bg: 'rgba(16, 185, 129, 0.1)' },
    CANCELLED: { color: 'var(--color-danger-rose)', bg: 'rgba(239, 68, 68, 0.1)' },
  };
  const c = config[status] || { color: 'var(--color-text-secondary)', bg: 'rgba(148, 163, 184, 0.1)' };
  return (
    <span
      className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
      style={{ backgroundColor: c.bg, color: c.color }}
    >
      {status}
    </span>
  );
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
  notifyTier2: boolean;
  notifyTier3: boolean;
  dailyDigest: boolean;
}

export default function HaulingRunsList() {
  const { data: session } = useSession();
  const router = useRouter();
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
    notifyTier2: false,
    notifyTier3: false,
    dailyDigest: false,
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
        notifyTier2: form.notifyTier2,
        notifyTier3: form.notifyTier3,
        dailyDigest: form.dailyDigest,
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
        setForm({ name: '', fromRegionId: '', toRegionId: '', maxVolumeM3: '', haulThresholdIsk: '', notifyTier2: false, notifyTier3: false, dailyDigest: false });
        await fetchRuns();
      }
    } catch (error) {
      console.error('Failed to create hauling run:', error);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteRun = async (id: number) => {
    if (!window.confirm('Delete this run?')) return;
    try {
      await fetch(`/api/hauling/runs/${id}`, { method: 'DELETE' });
      await fetchRuns();
    } catch (error) {
      console.error('Failed to delete hauling run:', error);
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
      <div className="w-full px-4 mt-8 mb-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-2">
            <Truck className="h-7 w-7 text-primary" />
            <h1 className="text-2xl font-bold text-text-emphasis">Hauling Runs</h1>
          </div>
          <Button onClick={() => setDialogOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            New Run
          </Button>
        </div>

        {/* Status summary badges */}
        {runs.length > 0 && (
          <div className="flex gap-2 mb-6 flex-wrap">
            {Object.entries(statusCounts).map(([status, count]) => (
              <StatusBadge key={status} status={`${status}: ${count}`} />
            ))}
          </div>
        )}

        {/* Runs Table */}
        {runs.length === 0 ? (
          <div className="empty-state">
            <Truck className="empty-state-icon" />
            <p className="empty-state-title">No hauling runs yet. Create your first run to get started.</p>
          </div>
        ) : (
          <Card className="bg-background-panel border-overlay-subtle">
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="bg-background-void border-overlay-subtle">
                    <TableHead className="font-bold text-text-emphasis">Name</TableHead>
                    <TableHead className="font-bold text-text-emphasis">Status</TableHead>
                    <TableHead className="font-bold text-text-emphasis">Route</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Volume Used</TableHead>
                    <TableHead className="font-bold text-text-emphasis text-right">Items</TableHead>
                    <TableHead className="font-bold text-text-emphasis w-40">Progress</TableHead>
                    <TableHead className="font-bold text-text-emphasis">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {runs.map((run, idx) => {
                    const { used, total } = computeRunVolume(run);
                    const fillPercent = computeOverallFill(run);
                    const fromName = EVE_REGIONS[run.fromRegionId] || `Region ${run.fromRegionId}`;
                    const toName = EVE_REGIONS[run.toRegionId] || `Region ${run.toRegionId}`;
                    const itemCount = (run.items || []).length;
                    const progressColor = fillPercent >= 100 ? 'var(--color-success-teal)' : fillPercent >= 50 ? 'var(--color-primary-cyan)' : 'var(--color-manufacturing-amber)';

                    return (
                      <TableRow
                        key={run.id}
                        className="cursor-pointer hover:bg-[rgba(0,212,255,0.03)] border-[rgba(148,163,184,0.07)]"
                        style={{ backgroundColor: idx % 2 === 0 ? undefined : 'rgba(255,255,255,0.02)' }}
                        onClick={() => router.push(`/hauling/${run.id}`)}
                      >
                        <TableCell className="font-semibold text-text-emphasis">
                          <Link href={`/hauling/${run.id}`} style={{ color: 'inherit', textDecoration: 'none' }}>
                            {run.name}
                          </Link>
                        </TableCell>
                        <TableCell>
                          <StatusBadge status={run.status} />
                        </TableCell>
                        <TableCell className="text-sm text-text-secondary">
                          {fromName} &rarr; {toName}
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">
                          {total > 0
                            ? `${used.toLocaleString(undefined, { maximumFractionDigits: 1 })} / ${total.toLocaleString()} m³`
                            : used > 0
                            ? `${used.toLocaleString(undefined, { maximumFractionDigits: 1 })} m³`
                            : '—'}
                        </TableCell>
                        <TableCell className="text-right text-sm text-text-secondary">{itemCount}</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <div className="flex-1">
                              <div className="h-2 bg-[#1e2a3a] rounded-full overflow-hidden">
                                <div
                                  className="h-full rounded-full transition-all"
                                  style={{ width: `${fillPercent}%`, backgroundColor: progressColor }}
                                />
                              </div>
                            </div>
                            <span className="text-xs text-text-secondary min-w-[36px] text-right">
                              {fillPercent.toFixed(0)}%
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-1" onClick={(e) => e.stopPropagation()}>
                            <Button asChild size="sm" variant="outline">
                              <Link href={`/hauling/${run.id}`}>View</Link>
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              className="border-red-500 text-red-400 hover:bg-red-500/10 ml-1"
                              onClick={() => handleDeleteRun(run.id)}
                            >
                              Delete
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          </Card>
        )}
      </div>

      {/* New Run Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-md bg-background-panel border-overlay-medium">
          <DialogHeader>
            <DialogTitle className="text-text-emphasis">New Hauling Run</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4 mt-2">
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Name *</label>
              <Input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="Run name"
                autoFocus
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">From Region *</label>
              <Select
                value={form.fromRegionId !== '' ? String(form.fromRegionId) : ''}
                onValueChange={(v) => setForm({ ...form, fromRegionId: Number(v) })}
              >
                <SelectTrigger className="bg-background-void border-overlay-strong text-text-emphasis">
                  <SelectValue placeholder="Select region" />
                </SelectTrigger>
                <SelectContent className="bg-background-panel border-overlay-medium">
                  {REGION_OPTIONS.map((r) => (
                    <SelectItem key={r.id} value={r.id.toString()} className="text-text-emphasis">
                      {r.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">To Region *</label>
              <Select
                value={form.toRegionId !== '' ? String(form.toRegionId) : ''}
                onValueChange={(v) => setForm({ ...form, toRegionId: Number(v) })}
              >
                <SelectTrigger className="bg-background-void border-overlay-strong text-text-emphasis">
                  <SelectValue placeholder="Select region" />
                </SelectTrigger>
                <SelectContent className="bg-background-panel border-overlay-medium">
                  {REGION_OPTIONS.map((r) => (
                    <SelectItem key={r.id} value={r.id.toString()} className="text-text-emphasis">
                      {r.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Max Volume (m³)</label>
              <Input
                type="number"
                value={form.maxVolumeM3}
                onChange={(e) => setForm({ ...form, maxVolumeM3: e.target.value })}
                min={0}
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
            </div>
            <div>
              <label className="text-xs text-text-secondary mb-1 block">Haul Threshold ISK</label>
              <Input
                type="number"
                value={form.haulThresholdIsk}
                onChange={(e) => setForm({ ...form, haulThresholdIsk: e.target.value })}
                min={0}
                className="bg-background-void border-overlay-strong text-text-emphasis"
              />
              <p className="text-xs text-text-muted mt-1">Minimum net profit to consider hauling</p>
            </div>
            <div>
              <p className="text-sm font-semibold text-text-emphasis mb-1">Discord Notifications</p>
              <p className="text-xs text-text-muted mb-2">Requires Discord linked in Settings.</p>
              <div className="flex flex-col gap-2">
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="notifyTier2"
                    checked={form.notifyTier2}
                    onCheckedChange={(checked) => setForm({ ...form, notifyTier2: Boolean(checked) })}
                  />
                  <label htmlFor="notifyTier2" className="text-sm text-text-secondary cursor-pointer">
                    Notify when fill crosses 80% (requires Discord)
                  </label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="notifyTier3"
                    checked={form.notifyTier3}
                    onCheckedChange={(checked) => setForm({ ...form, notifyTier3: Boolean(checked) })}
                  />
                  <label htmlFor="notifyTier3" className="text-sm text-text-secondary cursor-pointer">
                    Notify when items are slow to fill (requires Discord)
                  </label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="dailyDigest"
                    checked={form.dailyDigest}
                    onCheckedChange={(checked) => setForm({ ...form, dailyDigest: Boolean(checked) })}
                  />
                  <label htmlFor="dailyDigest" className="text-sm text-text-secondary cursor-pointer">
                    Daily digest in Discord
                  </label>
                </div>
              </div>
            </div>
          </div>
          <DialogFooter className="mt-4">
            <Button variant="ghost" onClick={() => setDialogOpen(false)} className="text-text-secondary">
              Cancel
            </Button>
            <Button
              onClick={handleCreateRun}
              disabled={!form.name || form.fromRegionId === '' || form.toRegionId === '' || submitting}
            >
              {submitting ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

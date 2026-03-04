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
    PLANNING: { color: '#00d4ff', bg: 'rgba(0, 212, 255, 0.1)' },
    ACCUMULATING: { color: '#f59e0b', bg: 'rgba(245, 158, 11, 0.1)' },
    READY: { color: '#10b981', bg: 'rgba(16, 185, 129, 0.1)' },
    IN_TRANSIT: { color: '#38bdf8', bg: 'rgba(56, 189, 248, 0.1)' },
    SELLING: { color: '#f59e0b', bg: 'rgba(245, 158, 11, 0.1)' },
    COMPLETE: { color: '#10b981', bg: 'rgba(16, 185, 129, 0.1)' },
    CANCELLED: { color: '#ef4444', bg: 'rgba(239, 68, 68, 0.1)' },
  };
  const c = config[status] || { color: '#94a3b8', bg: 'rgba(148, 163, 184, 0.1)' };
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
            <Truck className="h-7 w-7 text-[#00d4ff]" />
            <h1 className="text-2xl font-bold text-[#e2e8f0]">Hauling Runs</h1>
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
          <Card className="bg-[#12151f] border-[rgba(148,163,184,0.1)]">
            <CardContent className="py-8">
              <p className="text-center text-[#94a3b8] text-base">
                No hauling runs yet. Create your first run to get started.
              </p>
            </CardContent>
          </Card>
        ) : (
          <Card className="bg-[#12151f] border-[rgba(148,163,184,0.1)]">
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="bg-[#0f1219] border-[rgba(148,163,184,0.1)]">
                    <TableHead className="font-bold text-[#e2e8f0]">Name</TableHead>
                    <TableHead className="font-bold text-[#e2e8f0]">Status</TableHead>
                    <TableHead className="font-bold text-[#e2e8f0]">Route</TableHead>
                    <TableHead className="font-bold text-[#e2e8f0] text-right">Volume Used</TableHead>
                    <TableHead className="font-bold text-[#e2e8f0] text-right">Items</TableHead>
                    <TableHead className="font-bold text-[#e2e8f0] w-40">Progress</TableHead>
                    <TableHead className="font-bold text-[#e2e8f0]">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {runs.map((run, idx) => {
                    const { used, total } = computeRunVolume(run);
                    const fillPercent = computeOverallFill(run);
                    const fromName = EVE_REGIONS[run.fromRegionId] || `Region ${run.fromRegionId}`;
                    const toName = EVE_REGIONS[run.toRegionId] || `Region ${run.toRegionId}`;
                    const itemCount = (run.items || []).length;
                    const progressColor = fillPercent >= 100 ? '#10b981' : fillPercent >= 50 ? '#00d4ff' : '#f59e0b';

                    return (
                      <TableRow
                        key={run.id}
                        className="cursor-pointer hover:bg-[rgba(0,212,255,0.03)] border-[rgba(148,163,184,0.07)]"
                        style={{ backgroundColor: idx % 2 === 0 ? undefined : 'rgba(255,255,255,0.02)' }}
                        onClick={() => router.push(`/hauling/${run.id}`)}
                      >
                        <TableCell className="font-semibold text-[#e2e8f0]">
                          <Link href={`/hauling/${run.id}`} style={{ color: 'inherit', textDecoration: 'none' }}>
                            {run.name}
                          </Link>
                        </TableCell>
                        <TableCell>
                          <StatusBadge status={run.status} />
                        </TableCell>
                        <TableCell className="text-sm text-[#94a3b8]">
                          {fromName} &rarr; {toName}
                        </TableCell>
                        <TableCell className="text-right text-sm text-[#94a3b8]">
                          {total > 0
                            ? `${used.toLocaleString(undefined, { maximumFractionDigits: 1 })} / ${total.toLocaleString()} m³`
                            : used > 0
                            ? `${used.toLocaleString(undefined, { maximumFractionDigits: 1 })} m³`
                            : '—'}
                        </TableCell>
                        <TableCell className="text-right text-sm text-[#94a3b8]">{itemCount}</TableCell>
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
                            <span className="text-xs text-[#94a3b8] min-w-[36px] text-right">
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
        <DialogContent className="max-w-md bg-[#12151f] border-[rgba(148,163,184,0.15)]">
          <DialogHeader>
            <DialogTitle className="text-[#e2e8f0]">New Hauling Run</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4 mt-2">
            <div>
              <label className="text-xs text-[#94a3b8] mb-1 block">Name *</label>
              <Input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="Run name"
                autoFocus
                className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]"
              />
            </div>
            <div>
              <label className="text-xs text-[#94a3b8] mb-1 block">From Region *</label>
              <Select
                value={form.fromRegionId !== '' ? String(form.fromRegionId) : ''}
                onValueChange={(v) => setForm({ ...form, fromRegionId: Number(v) })}
              >
                <SelectTrigger className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]">
                  <SelectValue placeholder="Select region" />
                </SelectTrigger>
                <SelectContent className="bg-[#12151f] border-[rgba(148,163,184,0.15)]">
                  {REGION_OPTIONS.map((r) => (
                    <SelectItem key={r.id} value={r.id.toString()} className="text-[#e2e8f0]">
                      {r.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="text-xs text-[#94a3b8] mb-1 block">To Region *</label>
              <Select
                value={form.toRegionId !== '' ? String(form.toRegionId) : ''}
                onValueChange={(v) => setForm({ ...form, toRegionId: Number(v) })}
              >
                <SelectTrigger className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]">
                  <SelectValue placeholder="Select region" />
                </SelectTrigger>
                <SelectContent className="bg-[#12151f] border-[rgba(148,163,184,0.15)]">
                  {REGION_OPTIONS.map((r) => (
                    <SelectItem key={r.id} value={r.id.toString()} className="text-[#e2e8f0]">
                      {r.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="text-xs text-[#94a3b8] mb-1 block">Max Volume (m³)</label>
              <Input
                type="number"
                value={form.maxVolumeM3}
                onChange={(e) => setForm({ ...form, maxVolumeM3: e.target.value })}
                min={0}
                className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]"
              />
            </div>
            <div>
              <label className="text-xs text-[#94a3b8] mb-1 block">Haul Threshold ISK</label>
              <Input
                type="number"
                value={form.haulThresholdIsk}
                onChange={(e) => setForm({ ...form, haulThresholdIsk: e.target.value })}
                min={0}
                className="bg-[#0f1219] border-[rgba(148,163,184,0.2)] text-[#e2e8f0]"
              />
              <p className="text-xs text-[#64748b] mt-1">Minimum net profit to consider hauling</p>
            </div>
            <div>
              <p className="text-sm font-semibold text-[#e2e8f0] mb-1">Discord Notifications</p>
              <p className="text-xs text-[#64748b] mb-2">Requires Discord linked in Settings.</p>
              <div className="flex flex-col gap-2">
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="notifyTier2"
                    checked={form.notifyTier2}
                    onCheckedChange={(checked) => setForm({ ...form, notifyTier2: Boolean(checked) })}
                  />
                  <label htmlFor="notifyTier2" className="text-sm text-[#94a3b8] cursor-pointer">
                    Notify when fill crosses 80% (requires Discord)
                  </label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="notifyTier3"
                    checked={form.notifyTier3}
                    onCheckedChange={(checked) => setForm({ ...form, notifyTier3: Boolean(checked) })}
                  />
                  <label htmlFor="notifyTier3" className="text-sm text-[#94a3b8] cursor-pointer">
                    Notify when items are slow to fill (requires Discord)
                  </label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="dailyDigest"
                    checked={form.dailyDigest}
                    onCheckedChange={(checked) => setForm({ ...form, dailyDigest: Boolean(checked) })}
                  />
                  <label htmlFor="dailyDigest" className="text-sm text-[#94a3b8] cursor-pointer">
                    Daily digest in Discord
                  </label>
                </div>
              </div>
            </div>
          </div>
          <DialogFooter className="mt-4">
            <Button variant="ghost" onClick={() => setDialogOpen(false)} className="text-[#94a3b8]">
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

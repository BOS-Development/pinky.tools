import { useState, useEffect, useCallback } from 'react';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/router';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { ChevronUp, ChevronDown, ChevronsUpDown, ChevronLeft, ChevronRight, History } from 'lucide-react';
import Loading from '@industry-tool/components/loading';
import { HaulingRun } from '@industry-tool/client/data/models';
import moment from 'moment';

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
type SortKey = 'name' | 'status' | 'completedAt' | 'createdAt' | 'maxVolumeM3';

function StatusBadge({ status }: { status: string }) {
  const config: Record<string, { color: string; bg: string }> = {
    COMPLETE: { color: 'var(--color-success-teal)', bg: 'var(--color-success-tint)' },
    CANCELLED: { color: 'var(--color-danger-rose)', bg: 'var(--color-error-tint)' },
  };
  const c = config[status] || { color: 'var(--color-text-secondary)', bg: 'var(--color-neutral-tint)' };
  return (
    <span
      className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
      style={{ backgroundColor: c.bg, color: c.color }}
    >
      {status}
    </span>
  );
}

function SortIcon({ active, dir }: { active: boolean; dir: SortDir }) {
  if (!active) return <ChevronsUpDown className="h-3 w-3 ml-1 inline opacity-40" />;
  return dir === 'asc'
    ? <ChevronUp className="h-3 w-3 ml-1 inline text-primary" />
    : <ChevronDown className="h-3 w-3 ml-1 inline text-primary" />;
}

const PAGE_SIZE = 25;

export default function HaulingHistory() {
  const { data: session } = useSession();
  const router = useRouter();
  const [runs, setRuns] = useState<HaulingRun[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(0);
  const [statusFilter, setStatusFilter] = useState<'ALL' | 'COMPLETE' | 'CANCELLED'>('ALL');
  const [sortKey, setSortKey] = useState<SortKey>('completedAt');
  const [sortDir, setSortDir] = useState<SortDir>('desc');

  const fetchHistory = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        limit: String(PAGE_SIZE),
        offset: String(page * PAGE_SIZE),
      });
      const res = await fetch(`/api/hauling/history?${params.toString()}`);
      if (res.ok) {
        const data = await res.json();
        setRuns(Array.isArray(data.runs) ? data.runs : []);
        setTotal(typeof data.total === 'number' ? data.total : 0);
      }
    } catch (error) {
      console.error('Failed to fetch hauling history:', error);
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    if (session) fetchHistory();
  }, [session, fetchHistory]);

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('desc');
    }
  };

  const filtered = statusFilter === 'ALL'
    ? runs
    : runs.filter((r) => r.status === statusFilter);

  const sorted = [...filtered].sort((a, b) => {
    let cmp = 0;
    if (sortKey === 'name') {
      cmp = a.name.localeCompare(b.name);
    } else if (sortKey === 'status') {
      cmp = a.status.localeCompare(b.status);
    } else if (sortKey === 'completedAt') {
      cmp = (a.completedAt ?? '').localeCompare(b.completedAt ?? '');
    } else if (sortKey === 'createdAt') {
      cmp = a.createdAt.localeCompare(b.createdAt);
    } else if (sortKey === 'maxVolumeM3') {
      cmp = (a.maxVolumeM3 ?? 0) - (b.maxVolumeM3 ?? 0);
    }
    return sortDir === 'asc' ? cmp : -cmp;
  });

  const totalPages = Math.ceil(total / PAGE_SIZE);

  const th = (label: string, key: SortKey, right?: boolean) => (
    <TableHead
      className={`font-bold text-text-emphasis cursor-pointer select-none${right ? ' text-right' : ''}`}
      onClick={() => handleSort(key)}
    >
      {label}
      <SortIcon active={sortKey === key} dir={sortDir} />
    </TableHead>
  );

  if (loading) return <Loading />;

  return (
    <div className="flex flex-col gap-4">
      {/* Filter bar */}
      <div className="flex items-center gap-3">
        <label className="text-sm text-text-secondary">Status:</label>
        <Select
          value={statusFilter}
          onValueChange={(v) => {
            setStatusFilter(v as 'ALL' | 'COMPLETE' | 'CANCELLED');
            setPage(0);
          }}
        >
          <SelectTrigger className="w-40 bg-background-void border-overlay-strong text-text-emphasis h-8 text-sm">
            <SelectValue />
          </SelectTrigger>
          <SelectContent className="bg-background-panel border-overlay-medium">
            <SelectItem value="ALL" className="text-text-emphasis">All</SelectItem>
            <SelectItem value="COMPLETE" className="text-text-emphasis">Complete</SelectItem>
            <SelectItem value="CANCELLED" className="text-text-emphasis">Cancelled</SelectItem>
          </SelectContent>
        </Select>
        <span className="text-sm text-text-muted ml-auto">{total} total runs</span>
      </div>

      {/* Table */}
      {sorted.length === 0 ? (
        <div className="empty-state">
          <History className="empty-state-icon" />
          <p className="empty-state-title">No completed runs found.</p>
        </div>
      ) : (
        <Card className="bg-background-panel border-overlay-subtle">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="bg-background-void border-overlay-subtle">
                  {th('Name', 'name')}
                  <TableHead className="font-bold text-text-emphasis">Route</TableHead>
                  {th('Status', 'status')}
                  {th('Completed', 'completedAt')}
                  {th('Created', 'createdAt')}
                  {th('Max Volume', 'maxVolumeM3', true)}
                </TableRow>
              </TableHeader>
              <TableBody>
                {sorted.map((run, idx) => {
                  return (
                    <TableRow
                      key={run.id}
                      className="cursor-pointer hover:bg-interactive-hover border-overlay-subtle"
                      style={{ backgroundColor: idx % 2 === 0 ? undefined : 'var(--color-overlay-subtle)' }}
                      onClick={() => router.push(`/hauling/${run.id}`)}
                    >
                      <TableCell className="font-semibold text-text-emphasis">{run.name}</TableCell>
                      <TableCell className="text-sm text-text-secondary">
                        {regionName(run.fromRegionId)} &rarr; {regionName(run.toRegionId)}
                      </TableCell>
                      <TableCell>
                        <StatusBadge status={run.status} />
                      </TableCell>
                      <TableCell className="text-sm text-text-secondary">
                        {run.completedAt ? moment(run.completedAt).format('MMM D, YYYY') : '—'}
                      </TableCell>
                      <TableCell className="text-sm text-text-secondary">
                        {moment(run.createdAt).format('MMM D, YYYY')}
                      </TableCell>
                      <TableCell className="text-right text-sm text-text-secondary">
                        {run.maxVolumeM3 ? `${run.maxVolumeM3.toLocaleString()} m³` : '—'}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        </Card>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-end gap-2">
          <Button
            size="sm"
            variant="outline"
            disabled={page === 0}
            onClick={() => setPage((p) => p - 1)}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm text-text-secondary">
            Page {page + 1} of {totalPages}
          </span>
          <Button
            size="sm"
            variant="outline"
            disabled={page >= totalPages - 1}
            onClick={() => setPage((p) => p + 1)}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  );
}

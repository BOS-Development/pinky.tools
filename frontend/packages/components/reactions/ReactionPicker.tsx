import { useState, useMemo } from 'react';
import { ChevronDown, ChevronUp, ArrowUpDown } from 'lucide-react';
import {
  Table, TableHeader, TableBody, TableRow, TableHead, TableCell,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import {
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import { Reaction } from "@industry-tool/client/data/models";
import { formatISK, formatNumber, getValueColor } from "@industry-tool/utils/formatting";

const SIMPLE_GROUPS = ['Intermediate Materials', 'Unrefined Mineral'];

type SortKey = 'product_name' | 'group_name' | 'profit_per_run' | 'profit_per_cycle' | 'margin' | 'output_value_per_run';

type Props = {
  reactions: Reaction[];
  meFactor: number;
  selections: Record<number, number>;
  onSelectionChange: (reactionTypeId: number, instances: number) => void;
};

function SortHeader({ label, sortKey: key, activeSortKey, sortDir, onSort, align }: {
  label: string;
  sortKey: SortKey;
  activeSortKey: SortKey;
  sortDir: 'asc' | 'desc';
  onSort: (key: SortKey) => void;
  align?: 'left' | 'right';
}) {
  const active = activeSortKey === key;
  return (
    <TableHead className={align === 'right' ? 'text-right' : ''}>
      <button
        className={cn(
          "inline-flex items-center gap-1 text-xs font-medium uppercase tracking-wider cursor-pointer hover:text-[var(--color-primary-cyan)] transition-colors",
          active && "text-[var(--color-primary-cyan)]"
        )}
        onClick={() => onSort(key)}
      >
        {label}
        {active ? (
          sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />
        ) : (
          <ArrowUpDown className="h-3 w-3 opacity-40" />
        )}
      </button>
    </TableHead>
  );
}

export default function ReactionPicker({ reactions, meFactor, selections, onSelectionChange }: Props) {
  const [search, setSearch] = useState('');
  const [groupFilter, setGroupFilter] = useState('all');
  const [sortKey, setSortKey] = useState<SortKey>('product_name');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');
  const [expandedRow, setExpandedRow] = useState<number | null>(null);

  const complexReactions = useMemo(() =>
    reactions.filter(r => !SIMPLE_GROUPS.includes(r.group_name)),
    [reactions]
  );

  const groups = useMemo(() => {
    const g = new Set(complexReactions.map(r => r.group_name));
    return Array.from(g).sort();
  }, [complexReactions]);

  const filtered = useMemo(() => {
    let result = complexReactions;

    if (groupFilter !== 'all') {
      result = result.filter(r => r.group_name === groupFilter);
    }

    if (search) {
      const lower = search.toLowerCase();
      result = result.filter(r => r.product_name.toLowerCase().includes(lower));
    }

    result = [...result].sort((a, b) => {
      let cmp = 0;
      switch (sortKey) {
        case 'product_name': cmp = a.product_name.localeCompare(b.product_name); break;
        case 'group_name': cmp = a.group_name.localeCompare(b.group_name); break;
        case 'profit_per_run': cmp = a.profit_per_run - b.profit_per_run; break;
        case 'profit_per_cycle': cmp = a.profit_per_cycle - b.profit_per_cycle; break;
        case 'margin': cmp = a.margin - b.margin; break;
        case 'output_value_per_run': cmp = a.output_value_per_run - b.output_value_per_run; break;
      }
      return sortDir === 'asc' ? cmp : -cmp;
    });

    return result;
  }, [complexReactions, groupFilter, search, sortKey, sortDir]);

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir(prev => prev === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortDir(key === 'product_name' || key === 'group_name' ? 'asc' : 'desc');
    }
  };

  return (
    <div>
      <div className="flex gap-3 mb-3 flex-wrap items-center">
        <Input
          placeholder="Search..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-52 h-8"
        />
        <Select value={groupFilter} onValueChange={setGroupFilter}>
          <SelectTrigger className="w-44 h-8">
            <SelectValue placeholder="All Groups" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Groups</SelectItem>
            {groups.map(g => (
              <SelectItem key={g} value={g}>{g}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <span className="text-sm text-[var(--color-text-secondary)] self-center">
          {filtered.length} reactions | ME: {meFactor}
        </span>
      </div>

      <Table>
        <TableHeader>
          <TableRow className="bg-[var(--color-bg-panel)]">
            <TableHead className="w-10" />
            <TableHead className="w-20">Instances</TableHead>
            <SortHeader label="Product" sortKey="product_name" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} />
            <SortHeader label="Group" sortKey="group_name" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} />
            <TableHead className="text-right">Produced</TableHead>
            <TableHead className="text-right">Slots/Inst</TableHead>
            <SortHeader label="Output/Run" sortKey="output_value_per_run" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} align="right" />
            <SortHeader label="Profit/Run" sortKey="profit_per_run" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} align="right" />
            <SortHeader label="Profit/Cycle" sortKey="profit_per_cycle" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} align="right" />
            <SortHeader label="Margin" sortKey="margin" activeSortKey={sortKey} sortDir={sortDir} onSort={handleSort} align="right" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {filtered.map((r) => (
            <ReactionRow
              key={r.reaction_type_id}
              reaction={r}
              instances={selections[r.reaction_type_id] || 0}
              onInstanceChange={(val) => onSelectionChange(r.reaction_type_id, val)}
              expanded={expandedRow === r.reaction_type_id}
              onToggle={() => setExpandedRow(prev => prev === r.reaction_type_id ? null : r.reaction_type_id)}
            />
          ))}
          {filtered.length === 0 && (
            <TableRow>
              <TableCell colSpan={10} className="py-8 text-center text-[var(--color-text-secondary)]">
                No reactions found
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  );
}

type RowProps = {
  reaction: Reaction;
  instances: number;
  onInstanceChange: (val: number) => void;
  expanded: boolean;
  onToggle: () => void;
};

function ReactionRow({ reaction: r, instances, onInstanceChange, expanded, onToggle }: RowProps) {
  return (
    <>
      <TableRow
        className={cn(
          "odd:bg-status-neutral-tint",
          instances > 0 && "bg-interactive-hover"
        )}
      >
        <TableCell>
          <button
            className="p-1 rounded-sm hover:bg-[var(--color-surface-elevated)] text-[var(--color-text-secondary)] transition-colors"
            onClick={onToggle}
          >
            {expanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
          </button>
        </TableCell>
        <TableCell>
          <Input
            type="number"
            value={instances || ''}
            onChange={(e) => onInstanceChange(parseInt(e.target.value) || 0)}
            className="w-16 h-7 text-center"
            placeholder="0"
            min={0}
          />
        </TableCell>
        <TableCell>
          <div className="flex items-center gap-2">
            <img
              src={`https://images.evetech.net/types/${r.product_type_id}/icon?size=32`}
              alt=""
              width={24}
              height={24}
              className="rounded-sm"
            />
            {r.product_name}
          </div>
        </TableCell>
        <TableCell>{r.group_name}</TableCell>
        <TableCell className="text-right">{formatNumber(instances * r.complex_instances * r.runs_per_cycle * r.product_qty_per_run)}</TableCell>
        <TableCell className="text-right">{r.num_intermediates + r.complex_instances}</TableCell>
        <TableCell className="text-right">{formatISK(r.output_value_per_run)}</TableCell>
        <TableCell className="text-right" style={{ color: getValueColor(r.profit_per_run) }}>
          {formatISK(r.profit_per_run)}
        </TableCell>
        <TableCell className="text-right" style={{ color: getValueColor(r.profit_per_cycle) }}>
          {formatISK(r.profit_per_cycle)}
        </TableCell>
        <TableCell className="text-right" style={{ color: getValueColor(r.margin) }}>
          {r.margin.toFixed(2)}%
        </TableCell>
      </TableRow>
      {expanded && (
        <TableRow>
          <TableCell colSpan={10} className="py-0">
            <div className="py-2 px-4">
              <p className="text-sm font-medium mb-2 text-[var(--color-text-primary)]">Materials per run</p>
              <Table>
                <TableHeader>
                  <TableRow className="bg-[var(--color-bg-panel)]">
                    <TableHead>Material</TableHead>
                    <TableHead className="text-right">Base Qty</TableHead>
                    <TableHead className="text-right">Adj Qty</TableHead>
                    <TableHead className="text-right">Price</TableHead>
                    <TableHead className="text-right">Cost</TableHead>
                    <TableHead>Type</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {r.materials.map(m => (
                    <TableRow key={m.type_id}>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <img
                            src={`https://images.evetech.net/types/${m.type_id}/icon?size=32`}
                            alt=""
                            width={20}
                            height={20}
                            className="rounded-sm"
                          />
                          {m.name}
                        </div>
                      </TableCell>
                      <TableCell className="text-right">{formatNumber(m.base_qty)}</TableCell>
                      <TableCell className="text-right">{formatNumber(m.adj_qty)}</TableCell>
                      <TableCell className="text-right">{formatISK(m.price)}</TableCell>
                      <TableCell className="text-right">{formatISK(m.cost)}</TableCell>
                      <TableCell>
                        <span className={cn(
                          "text-xs",
                          m.is_intermediate ? "text-[var(--color-primary-cyan)]" : "text-[var(--color-text-secondary)]"
                        )}>
                          {m.is_intermediate ? 'Intermediate' : 'Raw'}
                        </span>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              <div className="mt-2 flex gap-6">
                <span className="text-xs text-[var(--color-text-secondary)]">
                  Input Cost: {formatISK(r.input_cost_per_run)} | Job Cost: {formatISK(r.job_cost_per_run)} | Output: {formatISK(r.output_value_per_run)}
                </span>
              </div>
            </div>
          </TableCell>
        </TableRow>
      )}
    </>
  );
}

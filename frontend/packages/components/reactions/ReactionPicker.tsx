import { useState, useMemo } from 'react';
import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import IconButton from '@mui/material/IconButton';
import Collapse from '@mui/material/Collapse';
import Typography from '@mui/material/Typography';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
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
    <Box>
      <Box sx={{ display: 'flex', gap: 2, mb: 2, flexWrap: 'wrap' }}>
        <TextField
          size="small"
          label="Search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          sx={{ minWidth: 200 }}
        />
        <FormControl size="small" sx={{ minWidth: 160 }}>
          <InputLabel>Group</InputLabel>
          <Select
            value={groupFilter}
            label="Group"
            onChange={(e) => setGroupFilter(e.target.value)}
          >
            <MenuItem value="all">All Groups</MenuItem>
            {groups.map(g => (
              <MenuItem key={g} value={g}>{g}</MenuItem>
            ))}
          </Select>
        </FormControl>
        <Typography variant="body2" color="text.secondary" sx={{ alignSelf: 'center' }}>
          {filtered.length} reactions | ME: {meFactor}
        </Typography>
      </Box>

      <TableContainer>
        <Table size="small" sx={{ '& th': { backgroundColor: '#0f1219', fontWeight: 'bold' } }}>
          <TableHead>
            <TableRow>
              <TableCell sx={{ width: 40 }} />
              <TableCell sx={{ width: 80 }}>Instances</TableCell>
              <TableCell>
                <TableSortLabel active={sortKey === 'product_name'} direction={sortDir} onClick={() => handleSort('product_name')}>
                  Product
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel active={sortKey === 'group_name'} direction={sortDir} onClick={() => handleSort('group_name')}>
                  Group
                </TableSortLabel>
              </TableCell>
              <TableCell align="right">Produced</TableCell>
              <TableCell align="right">Slots/Inst</TableCell>
              <TableCell align="right">
                <TableSortLabel active={sortKey === 'output_value_per_run'} direction={sortDir} onClick={() => handleSort('output_value_per_run')}>
                  Output/Run
                </TableSortLabel>
              </TableCell>
              <TableCell align="right">
                <TableSortLabel active={sortKey === 'profit_per_run'} direction={sortDir} onClick={() => handleSort('profit_per_run')}>
                  Profit/Run
                </TableSortLabel>
              </TableCell>
              <TableCell align="right">
                <TableSortLabel active={sortKey === 'profit_per_cycle'} direction={sortDir} onClick={() => handleSort('profit_per_cycle')}>
                  Profit/Cycle
                </TableSortLabel>
              </TableCell>
              <TableCell align="right">
                <TableSortLabel active={sortKey === 'margin'} direction={sortDir} onClick={() => handleSort('margin')}>
                  Margin
                </TableSortLabel>
              </TableCell>
            </TableRow>
          </TableHead>
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
                <TableCell colSpan={10} align="center" sx={{ py: 4, color: 'text.secondary' }}>
                  No reactions found
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
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
        sx={{
          '&:nth-of-type(odd)': { backgroundColor: 'rgba(255,255,255,0.02)' },
          backgroundColor: instances > 0 ? 'rgba(59, 130, 246, 0.08)' : undefined,
        }}
      >
        <TableCell>
          <IconButton size="small" onClick={onToggle}>
            {expanded ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>
        <TableCell>
          <TextField
            size="small"
            type="number"
            value={instances || ''}
            onChange={(e) => onInstanceChange(parseInt(e.target.value) || 0)}
            sx={{ width: 60 }}
            inputProps={{ min: 0, style: { textAlign: 'center' } }}
            placeholder="0"
          />
        </TableCell>
        <TableCell>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <img
              src={`https://images.evetech.net/types/${r.product_type_id}/icon?size=32`}
              alt=""
              width={24}
              height={24}
              style={{ borderRadius: 2 }}
            />
            {r.product_name}
          </Box>
        </TableCell>
        <TableCell>{r.group_name}</TableCell>
        <TableCell align="right">{formatNumber(instances * r.complex_instances * r.runs_per_cycle * r.product_qty_per_run)}</TableCell>
        <TableCell align="right">{r.num_intermediates + r.complex_instances}</TableCell>
        <TableCell align="right">{formatISK(r.output_value_per_run)}</TableCell>
        <TableCell align="right" sx={{ color: getValueColor(r.profit_per_run) }}>
          {formatISK(r.profit_per_run)}
        </TableCell>
        <TableCell align="right" sx={{ color: getValueColor(r.profit_per_cycle) }}>
          {formatISK(r.profit_per_cycle)}
        </TableCell>
        <TableCell align="right" sx={{ color: getValueColor(r.margin) }}>
          {r.margin.toFixed(2)}%
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell colSpan={10} sx={{ py: 0, borderBottom: expanded ? undefined : 'none' }}>
          <Collapse in={expanded} timeout="auto" unmountOnExit>
            <Box sx={{ py: 1, px: 2 }}>
              <Typography variant="subtitle2" sx={{ mb: 1 }}>Materials per run</Typography>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Material</TableCell>
                    <TableCell align="right">Base Qty</TableCell>
                    <TableCell align="right">Adj Qty</TableCell>
                    <TableCell align="right">Price</TableCell>
                    <TableCell align="right">Cost</TableCell>
                    <TableCell>Type</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {r.materials.map(m => (
                    <TableRow key={m.type_id}>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <img
                            src={`https://images.evetech.net/types/${m.type_id}/icon?size=32`}
                            alt=""
                            width={20}
                            height={20}
                            style={{ borderRadius: 2 }}
                          />
                          {m.name}
                        </Box>
                      </TableCell>
                      <TableCell align="right">{formatNumber(m.base_qty)}</TableCell>
                      <TableCell align="right">{formatNumber(m.adj_qty)}</TableCell>
                      <TableCell align="right">{formatISK(m.price)}</TableCell>
                      <TableCell align="right">{formatISK(m.cost)}</TableCell>
                      <TableCell>
                        <Typography
                          variant="caption"
                          sx={{ color: m.is_intermediate ? '#3b82f6' : '#94a3b8' }}
                        >
                          {m.is_intermediate ? 'Intermediate' : 'Raw'}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              <Box sx={{ mt: 1, display: 'flex', gap: 3 }}>
                <Typography variant="caption" color="text.secondary">
                  Input Cost: {formatISK(r.input_cost_per_run)} | Job Cost: {formatISK(r.job_cost_per_run)} | Output: {formatISK(r.output_value_per_run)}
                </Typography>
              </Box>
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  );
}

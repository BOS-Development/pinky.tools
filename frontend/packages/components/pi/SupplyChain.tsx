import { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import { SupplyChainResponse, SupplyChainItem, SupplyChainPlanetEntry, StockpileMarker } from '@industry-tool/client/data/models';
import { formatNumber, formatQuantity } from '@industry-tool/utils/formatting';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TextField from '@mui/material/TextField';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Popover from '@mui/material/Popover';
import Button from '@mui/material/Button';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowRightIcon from '@mui/icons-material/KeyboardArrowRight';

const TIERS = ['All', 'R0', 'P1', 'P2', 'P3', 'P4'];
const TIER_ORDER: Record<string, number> = { R0: 0, P1: 1, P2: 2, P3: 3, P4: 4 };

const RESIZE_PRESETS = [
  { label: '1 week', hours: 168 },
  { label: '2 weeks', hours: 336 },
  { label: '1 month', hours: 720 },
];

const SOURCE_COLORS: Record<string, { bg: string; text: string }> = {
  extracted: { bg: 'rgba(16, 185, 129, 0.1)', text: '#10b981' },
  produced:  { bg: 'rgba(59, 130, 246, 0.1)', text: '#3b82f6' },
  bought:    { bg: 'rgba(245, 158, 11, 0.1)', text: '#f59e0b' },
  mixed:     { bg: 'rgba(148, 163, 184, 0.1)', text: '#94a3b8' },
};

function formatDepletion(hours: number): string {
  if (hours <= 0) return '\u2014';
  const days = Math.floor(hours / 24);
  const h = Math.floor(hours % 24);
  if (days > 0) return `${days}d ${h}h`;
  const m = Math.round((hours % 1) * 60);
  return `${h}h ${m}m`;
}

function depletionColor(hours: number): string {
  if (hours <= 0) return '#64748b';
  if (hours < 24) return '#ef4444';
  if (hours < 72) return '#f59e0b';
  return '#10b981';
}

function netColor(value: number): string {
  if (value > 0.01) return '#10b981';
  if (value < -0.01) return '#ef4444';
  return '#64748b';
}

function formatRate(value: number): string {
  if (value === 0) return '\u2014';
  return formatNumber(Math.round(value));
}

const headerCellSx = {
  color: '#64748b',
  borderColor: 'rgba(148, 163, 184, 0.1)',
  bgcolor: '#0f1219',
  fontSize: '0.7rem',
  textTransform: 'uppercase' as const,
  letterSpacing: 0.5,
  fontWeight: 600,
  py: 1,
};

function SupplyChainRow({ item, expanded, onToggle, onResize }: {
  item: SupplyChainItem;
  expanded: boolean;
  onToggle: () => void;
  onResize: (item: SupplyChainItem, newQty: number) => void;
}) {
  const sourceStyle = SOURCE_COLORS[item.source] || SOURCE_COLORS.mixed;
  const hasChildren = (item.producers?.length > 0) || (item.consumers?.length > 0);
  const isDeficit = item.netPerHour < -0.01;
  const hasMarkers = (item.stockpileMarkers?.length ?? 0) > 0;
  const canResize = hasMarkers && item.consumedPerHour > 0;

  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);

  const handleStockpileClick = (e: React.MouseEvent<HTMLElement>) => {
    if (!canResize) return;
    e.stopPropagation();
    setAnchorEl(e.currentTarget);
  };

  const handlePresetClick = (hours: number) => {
    const newQty = Math.ceil(item.consumedPerHour * hours);
    onResize(item, newQty);
    setAnchorEl(null);
  };

  return (
    <>
      <TableRow
        sx={{
          cursor: hasChildren ? 'pointer' : 'default',
          '&:hover': { bgcolor: 'rgba(59, 130, 246, 0.04)' },
          bgcolor: expanded ? 'rgba(59, 130, 246, 0.03)' : isDeficit ? 'rgba(239, 68, 68, 0.03)' : 'transparent',
        }}
        onClick={hasChildren ? onToggle : undefined}
      >
        <TableCell sx={{ color: '#e2e8f0', borderColor: 'rgba(148, 163, 184, 0.08)', width: 36, p: 0.5 }}>
          {hasChildren && (
            <IconButton size="small" sx={{ color: '#64748b', p: 0.25 }}>
              {expanded ? <KeyboardArrowDownIcon fontSize="small" /> : <KeyboardArrowRightIcon fontSize="small" />}
            </IconButton>
          )}
        </TableCell>
        <TableCell sx={{ borderColor: 'rgba(148, 163, 184, 0.08)' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <img
              src={`https://images.evetech.net/types/${item.typeId}/icon?size=32`}
              alt="" width={24} height={24} style={{ flexShrink: 0 }}
            />
            <Box sx={{ minWidth: 0 }}>
              <Typography variant="body2" sx={{ color: '#e2e8f0', fontWeight: 500, lineHeight: 1.2 }} noWrap>
                {item.name || `Type ${item.typeId}`}
              </Typography>
              <Typography variant="caption" sx={{ color: '#475569', fontSize: '0.65rem' }}>
                {item.tierName}
              </Typography>
            </Box>
          </Box>
        </TableCell>
        <TableCell sx={{ borderColor: 'rgba(148, 163, 184, 0.08)' }}>
          <Chip
            label={item.source.charAt(0).toUpperCase() + item.source.slice(1)}
            size="small"
            sx={{
              bgcolor: sourceStyle.bg,
              color: sourceStyle.text,
              fontSize: '0.65rem',
              height: 20,
              fontWeight: 500,
            }}
          />
        </TableCell>
        <TableCell align="right" sx={{ borderColor: 'rgba(148, 163, 184, 0.08)', color: item.producedPerHour > 0 ? '#10b981' : '#475569' }}>
          <Typography variant="caption">{formatRate(item.producedPerHour)}</Typography>
        </TableCell>
        <TableCell align="right" sx={{ borderColor: 'rgba(148, 163, 184, 0.08)', color: item.consumedPerHour > 0 ? '#ef4444' : '#475569' }}>
          <Typography variant="caption">{formatRate(item.consumedPerHour)}</Typography>
        </TableCell>
        <TableCell align="right" sx={{ borderColor: 'rgba(148, 163, 184, 0.08)' }}>
          <Typography variant="caption" sx={{ color: netColor(item.netPerHour), fontWeight: 600 }}>
            {item.netPerHour > 0.01 ? '+' : ''}{formatRate(item.netPerHour)}
          </Typography>
        </TableCell>
        <TableCell
          align="right"
          onClick={canResize ? handleStockpileClick : undefined}
          sx={{
            borderColor: 'rgba(148, 163, 184, 0.08)',
            color: '#94a3b8',
            cursor: canResize ? 'pointer' : 'default',
            '&:hover': canResize ? { color: '#3b82f6' } : {},
          }}
        >
          <Typography variant="caption" sx={{ borderBottom: canResize ? '1px dashed currentColor' : 'none' }}>
            {item.stockpileQty > 0 ? formatQuantity(item.stockpileQty) : '\u2014'}
          </Typography>
          <Popover
            open={Boolean(anchorEl)}
            anchorEl={anchorEl}
            onClose={(e: React.SyntheticEvent) => { e.stopPropagation?.(); setAnchorEl(null); }}
            anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
            transformOrigin={{ vertical: 'top', horizontal: 'right' }}
            slotProps={{
              paper: {
                sx: { bgcolor: '#1a1f2e', border: '1px solid rgba(148, 163, 184, 0.15)', p: 1.5, minWidth: 200 },
                onClick: (e: React.MouseEvent) => e.stopPropagation(),
              },
            }}
          >
            <Typography variant="caption" sx={{ color: '#64748b', fontWeight: 600, textTransform: 'uppercase', letterSpacing: 0.5, fontSize: '0.6rem', mb: 1, display: 'block' }}>
              Set stockpile for
            </Typography>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
              {RESIZE_PRESETS.map(preset => {
                const qty = Math.ceil(item.consumedPerHour * preset.hours);
                return (
                  <Button
                    key={preset.label}
                    size="small"
                    onClick={(e) => { e.stopPropagation(); handlePresetClick(preset.hours); }}
                    sx={{
                      justifyContent: 'space-between',
                      textTransform: 'none',
                      color: '#e2e8f0',
                      fontSize: '0.75rem',
                      py: 0.5,
                      px: 1,
                      '&:hover': { bgcolor: 'rgba(59, 130, 246, 0.1)' },
                    }}
                  >
                    <span>{preset.label}</span>
                    <Typography component="span" sx={{ color: '#3b82f6', fontSize: '0.75rem', fontWeight: 600, ml: 2 }}>
                      {formatQuantity(qty)}
                    </Typography>
                  </Button>
                );
              })}
            </Box>
          </Popover>
        </TableCell>
        <TableCell align="right" sx={{ borderColor: 'rgba(148, 163, 184, 0.08)' }}>
          <Typography variant="caption" sx={{ color: depletionColor(item.depletionHours), fontWeight: item.depletionHours > 0 && item.depletionHours < 72 ? 600 : 400 }}>
            {formatDepletion(item.depletionHours)}
          </Typography>
        </TableCell>
      </TableRow>
      {expanded && (
        <>
          {item.producers?.length > 0 && (
            <>
              <TableRow>
                <TableCell sx={{ borderColor: 'rgba(148, 163, 184, 0.03)', py: 0.5 }} />
                <TableCell colSpan={7} sx={{ borderColor: 'rgba(148, 163, 184, 0.03)', py: 0.5 }}>
                  <Typography variant="caption" sx={{ color: '#64748b', fontWeight: 600, textTransform: 'uppercase', letterSpacing: 0.5, fontSize: '0.6rem' }}>
                    Producers
                  </Typography>
                </TableCell>
              </TableRow>
              {item.producers.map((p, i) => (
                <PlanetEntryRow key={`prod-${i}`} entry={p} type="producer" />
              ))}
            </>
          )}
          {item.consumers?.length > 0 && (
            <>
              <TableRow>
                <TableCell sx={{ borderColor: 'rgba(148, 163, 184, 0.03)', py: 0.5 }} />
                <TableCell colSpan={7} sx={{ borderColor: 'rgba(148, 163, 184, 0.03)', py: 0.5 }}>
                  <Typography variant="caption" sx={{ color: '#64748b', fontWeight: 600, textTransform: 'uppercase', letterSpacing: 0.5, fontSize: '0.6rem' }}>
                    Consumers
                  </Typography>
                </TableCell>
              </TableRow>
              {item.consumers.map((c, i) => (
                <PlanetEntryRow key={`cons-${i}`} entry={c} type="consumer" />
              ))}
            </>
          )}
        </>
      )}
    </>
  );
}

function PlanetEntryRow({ entry, type }: { entry: SupplyChainPlanetEntry; type: 'producer' | 'consumer' }) {
  const rateColor = type === 'producer' ? '#10b981' : '#ef4444';

  return (
    <TableRow sx={{ bgcolor: 'rgba(15, 18, 25, 0.4)' }}>
      <TableCell sx={{ borderColor: 'rgba(148, 163, 184, 0.03)' }} />
      <TableCell colSpan={3} sx={{ borderColor: 'rgba(148, 163, 184, 0.03)', pl: 5 }}>
        <Typography variant="caption" sx={{ color: '#cbd5e1' }}>
          {entry.solarSystemName}
        </Typography>
        <Typography variant="caption" sx={{ color: '#475569', ml: 0.5 }}>
          {entry.planetType} &middot; {entry.characterName}
        </Typography>
      </TableCell>
      <TableCell colSpan={4} align="right" sx={{ borderColor: 'rgba(148, 163, 184, 0.03)' }}>
        <Typography variant="caption" sx={{ color: rateColor }}>
          {formatNumber(Math.round(entry.ratePerHour))}/hr
        </Typography>
      </TableCell>
    </TableRow>
  );
}

export default function SupplyChain() {
  const { data: session } = useSession();
  const [data, setData] = useState<SupplyChainResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [tierFilter, setTierFilter] = useState('All');
  const [search, setSearch] = useState('');
  const [expandedItems, setExpandedItems] = useState<Set<number>>(new Set());
  const [bulkAnchorEl, setBulkAnchorEl] = useState<HTMLElement | null>(null);
  const hasFetchedRef = useRef(false);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchData();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session]);

  const fetchData = async () => {
    if (!session) return;
    setLoading(true);
    try {
      const response = await fetch('/api/pi/supply-chain');
      if (response.ok) {
        const result: SupplyChainResponse = await response.json();
        setData(result);
      }
    } finally {
      setLoading(false);
    }
  };

  const upsertMarkers = async (markers: StockpileMarker[], newQty: number) => {
    const totalOld = markers.reduce((sum, m) => sum + m.desiredQuantity, 0);
    const updates = markers.map((m, i) => {
      if (markers.length === 1) return { ...m, desiredQuantity: newQty };
      if (i === markers.length - 1) {
        const allocated = markers.slice(0, -1).reduce((sum, mk) => {
          return sum + Math.round(newQty * (mk.desiredQuantity / totalOld));
        }, 0);
        return { ...m, desiredQuantity: newQty - allocated };
      }
      return { ...m, desiredQuantity: Math.round(newQty * (m.desiredQuantity / totalOld)) };
    });
    for (const marker of updates) {
      await fetch('/api/stockpiles/upsert', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(marker),
      });
    }
  };

  const handleResize = useCallback(async (item: SupplyChainItem, newQty: number) => {
    const markers = item.stockpileMarkers;
    if (!markers || markers.length === 0) return;
    await upsertMarkers(markers, newQty);
    hasFetchedRef.current = false;
    await fetchData();
    hasFetchedRef.current = true;
  }, []);

  const handleResizeAll = useCallback(async (hours: number) => {
    if (!data?.items) return;
    const resizable = data.items.filter(
      item => (item.stockpileMarkers?.length ?? 0) > 0 && item.consumedPerHour > 0
    );
    if (resizable.length === 0) return;

    setLoading(true);
    for (const item of resizable) {
      const newQty = Math.ceil(item.consumedPerHour * hours);
      await upsertMarkers(item.stockpileMarkers!, newQty);
    }
    hasFetchedRef.current = false;
    await fetchData();
    hasFetchedRef.current = true;
  }, [data]);

  const toggleItem = (typeId: number) => {
    setExpandedItems(prev => {
      const next = new Set(prev);
      if (next.has(typeId)) next.delete(typeId);
      else next.add(typeId);
      return next;
    });
  };

  const filteredItems = useMemo(() => {
    if (!data?.items) return [];
    let items = data.items;

    if (tierFilter !== 'All') {
      items = items.filter(item => item.tierName === tierFilter);
    }

    if (search.trim()) {
      const q = search.trim().toLowerCase();
      items = items.filter(item => item.name?.toLowerCase().includes(q));
    }

    // Sort: tier ascending, then deficit first (net ascending), then name
    return [...items].sort((a, b) => {
      const tierDiff = (TIER_ORDER[a.tierName] ?? 99) - (TIER_ORDER[b.tierName] ?? 99);
      if (tierDiff !== 0) return tierDiff;
      if (a.netPerHour !== b.netPerHour) return a.netPerHour - b.netPerHour;
      return (a.name || '').localeCompare(b.name || '');
    });
  }, [data, tierFilter, search]);

  const summary = useMemo(() => {
    if (!data?.items) return { total: 0, deficit: 0, surplus: 0, balanced: 0 };
    let deficit = 0, surplus = 0, balanced = 0;
    for (const item of data.items) {
      if (item.netPerHour < -0.01) deficit++;
      else if (item.netPerHour > 0.01) surplus++;
      else balanced++;
    }
    return { total: data.items.length, deficit, surplus, balanced };
  }, [data]);

  if (loading) {
    return <Loading />;
  }

  return (
    <Box>
      {/* Summary chips */}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2, flexWrap: 'wrap' }}>
        <Typography variant="caption" sx={{ color: '#64748b' }}>
          {summary.total} materials
        </Typography>
        {summary.deficit > 0 && (
          <Chip
            label={`${summary.deficit} deficit`}
            size="small"
            sx={{ bgcolor: 'rgba(239, 68, 68, 0.1)', color: '#ef4444', fontSize: '0.7rem', height: 22 }}
          />
        )}
        {summary.surplus > 0 && (
          <Chip
            label={`${summary.surplus} surplus`}
            size="small"
            sx={{ bgcolor: 'rgba(16, 185, 129, 0.1)', color: '#10b981', fontSize: '0.7rem', height: 22 }}
          />
        )}
        {summary.balanced > 0 && (
          <Chip
            label={`${summary.balanced} balanced`}
            size="small"
            sx={{ bgcolor: 'rgba(148, 163, 184, 0.1)', color: '#94a3b8', fontSize: '0.7rem', height: 22 }}
          />
        )}
      </Box>

      {/* Filters */}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2, flexWrap: 'wrap' }}>
        <Box sx={{ display: 'flex', gap: 0.5 }}>
          {TIERS.map(tier => (
            <Chip
              key={tier}
              label={tier}
              size="small"
              onClick={() => setTierFilter(tier)}
              sx={{
                bgcolor: tierFilter === tier ? 'rgba(59, 130, 246, 0.15)' : 'rgba(148, 163, 184, 0.06)',
                color: tierFilter === tier ? '#3b82f6' : '#64748b',
                border: tierFilter === tier ? '1px solid rgba(59, 130, 246, 0.3)' : '1px solid transparent',
                fontSize: '0.7rem',
                fontWeight: 500,
                height: 26,
                cursor: 'pointer',
                '&:hover': { bgcolor: tierFilter === tier ? 'rgba(59, 130, 246, 0.2)' : 'rgba(148, 163, 184, 0.1)' },
              }}
            />
          ))}
        </Box>
        {data?.items?.some(item => (item.stockpileMarkers?.length ?? 0) > 0 && item.consumedPerHour > 0) && (
          <>
            <Button
              size="small"
              onClick={(e) => setBulkAnchorEl(e.currentTarget)}
              sx={{
                textTransform: 'none',
                color: '#94a3b8',
                fontSize: '0.75rem',
                border: '1px solid rgba(148, 163, 184, 0.15)',
                px: 1.5,
                '&:hover': { bgcolor: 'rgba(59, 130, 246, 0.08)', borderColor: 'rgba(59, 130, 246, 0.3)', color: '#3b82f6' },
              }}
            >
              Set all stockpiles
            </Button>
            <Popover
              open={Boolean(bulkAnchorEl)}
              anchorEl={bulkAnchorEl}
              onClose={() => setBulkAnchorEl(null)}
              anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
              transformOrigin={{ vertical: 'top', horizontal: 'left' }}
              slotProps={{
                paper: {
                  sx: { bgcolor: '#1a1f2e', border: '1px solid rgba(148, 163, 184, 0.15)', p: 1.5, minWidth: 220 },
                },
              }}
            >
              <Typography variant="caption" sx={{ color: '#64748b', fontWeight: 600, textTransform: 'uppercase', letterSpacing: 0.5, fontSize: '0.6rem', mb: 1, display: 'block' }}>
                Set all stockpiles to cover
              </Typography>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                {RESIZE_PRESETS.map(preset => (
                  <Button
                    key={preset.label}
                    size="small"
                    onClick={() => { setBulkAnchorEl(null); handleResizeAll(preset.hours); }}
                    sx={{
                      justifyContent: 'flex-start',
                      textTransform: 'none',
                      color: '#e2e8f0',
                      fontSize: '0.75rem',
                      py: 0.5,
                      px: 1,
                      '&:hover': { bgcolor: 'rgba(59, 130, 246, 0.1)' },
                    }}
                  >
                    {preset.label}
                  </Button>
                ))}
              </Box>
            </Popover>
          </>
        )}
        <TextField
          placeholder="Search..."
          value={search}
          onChange={e => setSearch(e.target.value)}
          size="small"
          variant="outlined"
          sx={{
            ml: 'auto',
            width: 180,
            '& .MuiOutlinedInput-root': {
              color: '#e2e8f0',
              fontSize: '0.8rem',
              '& fieldset': { borderColor: 'rgba(148, 163, 184, 0.15)' },
              '&:hover fieldset': { borderColor: 'rgba(59, 130, 246, 0.3)' },
              '&.Mui-focused fieldset': { borderColor: '#3b82f6' },
            },
            '& .MuiInputBase-input': { py: 0.75 },
          }}
        />
      </Box>

      {/* Table */}
      <TableContainer>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell sx={{ ...headerCellSx, width: 36 }} />
              <TableCell sx={headerCellSx}>Material</TableCell>
              <TableCell sx={headerCellSx}>Source</TableCell>
              <TableCell align="right" sx={headerCellSx}>Produced/hr</TableCell>
              <TableCell align="right" sx={headerCellSx}>Consumed/hr</TableCell>
              <TableCell align="right" sx={headerCellSx}>Net/hr</TableCell>
              <TableCell align="right" sx={headerCellSx}>Stockpile</TableCell>
              <TableCell align="right" sx={headerCellSx}>Depletion</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredItems.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} align="center" sx={{ color: '#64748b', borderColor: 'rgba(148, 163, 184, 0.1)', py: 4 }}>
                  No PI production data found
                </TableCell>
              </TableRow>
            ) : (
              filteredItems.map(item => (
                <SupplyChainRow
                  key={item.typeId}
                  item={item}
                  expanded={expandedItems.has(item.typeId)}
                  onToggle={() => toggleItem(item.typeId)}
                  onResize={handleResize}
                />
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}

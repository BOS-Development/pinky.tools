import { useState, useEffect, useRef, useMemo } from 'react';
import { useSession } from "next-auth/react";
import Loading from "@industry-tool/components/loading";
import { PiProfitResponse, PiFactoryProfit } from "@industry-tool/client/data/models";
import { formatISK, formatNumber } from "@industry-tool/utils/formatting";
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import IconButton from '@mui/material/IconButton';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';
import KeyboardArrowRightIcon from '@mui/icons-material/KeyboardArrowRight';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import TrendingDownIcon from '@mui/icons-material/TrendingDown';

type ProductGroup = {
  outputTypeId: number;
  outputName: string;
  outputTier: string;
  totalRatePerHour: number;
  totalOutputValue: number;
  totalInputCost: number;
  totalExportTax: number;
  totalImportTax: number;
  totalProfit: number;
  factories: FactoryWithPlanet[];
};

type FactoryWithPlanet = PiFactoryProfit & {
  solarSystemName: string;
  characterName: string;
  planetType: string;
};

function profitColor(value: number): string {
  if (value > 0) return '#10b981';
  if (value < 0) return '#ef4444';
  return '#94a3b8';
}

const TIER_ORDER: Record<string, number> = { R0: 0, P1: 1, P2: 2, P3: 3, P4: 4 };

function ProductGroupRow({ group, expanded, onToggle }: {
  group: ProductGroup;
  expanded: boolean;
  onToggle: () => void;
}) {
  const totalTax = group.totalExportTax + group.totalImportTax;

  return (
    <>
      <TableRow
        sx={{
          cursor: 'pointer',
          '&:hover': { bgcolor: 'rgba(59, 130, 246, 0.05)' },
          bgcolor: expanded ? 'rgba(59, 130, 246, 0.03)' : 'transparent',
        }}
        onClick={onToggle}
      >
        <TableCell sx={{ color: '#e2e8f0', borderColor: 'rgba(148, 163, 184, 0.1)', width: 40, p: 1 }}>
          <IconButton size="small" sx={{ color: '#64748b', p: 0.5 }}>
            {expanded ? <KeyboardArrowDownIcon fontSize="small" /> : <KeyboardArrowRightIcon fontSize="small" />}
          </IconButton>
        </TableCell>
        <TableCell sx={{ color: '#e2e8f0', borderColor: 'rgba(148, 163, 184, 0.1)' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            {group.outputTypeId > 0 && (
              <img
                src={`https://images.evetech.net/types/${group.outputTypeId}/icon?size=32`}
                alt="" width={20} height={20} style={{ flexShrink: 0 }}
              />
            )}
            <Box>
              <Typography variant="body2" sx={{ color: '#e2e8f0', fontWeight: 500, lineHeight: 1.2 }}>
                {group.outputName}
              </Typography>
              <Typography variant="caption" sx={{ color: '#64748b' }}>
                {group.outputTier} &middot; {group.factories.length} {group.factories.length === 1 ? 'factory' : 'factories'}
              </Typography>
            </Box>
          </Box>
        </TableCell>
        <TableCell align="right" sx={{ color: '#94a3b8', borderColor: 'rgba(148, 163, 184, 0.1)' }}>
          {formatNumber(Math.round(group.totalRatePerHour))}/hr
        </TableCell>
        <TableCell align="right" sx={{ color: '#10b981', borderColor: 'rgba(148, 163, 184, 0.1)' }}>
          {formatISK(group.totalOutputValue)}
        </TableCell>
        <TableCell align="right" sx={{ color: '#ef4444', borderColor: 'rgba(148, 163, 184, 0.1)' }}>
          {formatISK(group.totalInputCost)}
        </TableCell>
        <TableCell align="right" sx={{ color: '#f59e0b', borderColor: 'rgba(148, 163, 184, 0.1)' }}>
          {formatISK(totalTax)}
        </TableCell>
        <TableCell align="right" sx={{ color: profitColor(group.totalProfit), fontWeight: 600, borderColor: 'rgba(148, 163, 184, 0.1)' }}>
          {formatISK(group.totalProfit)}
        </TableCell>
      </TableRow>
      {expanded && group.factories.map((factory) => (
        <FactoryDetailRow key={`${factory.pinId}`} factory={factory} />
      ))}
    </>
  );
}

function FactoryDetailRow({ factory }: { factory: FactoryWithPlanet }) {
  return (
    <TableRow sx={{ bgcolor: 'rgba(15, 18, 25, 0.5)' }}>
      <TableCell sx={{ borderColor: 'rgba(148, 163, 184, 0.05)' }} />
      <TableCell sx={{ borderColor: 'rgba(148, 163, 184, 0.05)', pl: 6 }}>
        <Typography variant="caption" sx={{ color: '#cbd5e1' }}>
          {factory.solarSystemName}
        </Typography>
        <Typography variant="caption" sx={{ color: '#475569', ml: 0.5 }}>
          ({factory.characterName})
        </Typography>
      </TableCell>
      <TableCell align="right" sx={{ borderColor: 'rgba(148, 163, 184, 0.05)' }}>
        <Typography variant="caption" sx={{ color: '#64748b' }}>
          {formatNumber(Math.round(factory.ratePerHour))}/hr
        </Typography>
      </TableCell>
      <TableCell align="right" sx={{ borderColor: 'rgba(148, 163, 184, 0.05)' }}>
        <Typography variant="caption" sx={{ color: '#10b981' }}>
          {formatISK(factory.outputValuePerHour)}
        </Typography>
      </TableCell>
      <TableCell align="right" sx={{ borderColor: 'rgba(148, 163, 184, 0.05)' }}>
        <Typography variant="caption" sx={{ color: '#ef4444' }}>
          {formatISK(factory.inputCostPerHour)}
        </Typography>
      </TableCell>
      <TableCell align="right" sx={{ borderColor: 'rgba(148, 163, 184, 0.05)' }}>
        <Typography variant="caption" sx={{ color: '#f59e0b' }}>
          {formatISK(factory.exportTaxPerHour + factory.importTaxPerHour)}
        </Typography>
      </TableCell>
      <TableCell align="right" sx={{ borderColor: 'rgba(148, 163, 184, 0.05)' }}>
        <Typography variant="caption" sx={{ color: profitColor(factory.profitPerHour), fontWeight: 600 }}>
          {formatISK(factory.profitPerHour)}
        </Typography>
      </TableCell>
    </TableRow>
  );
}

export default function ProfitTable() {
  const { data: session } = useSession();
  const [profitResponse, setProfitResponse] = useState<PiProfitResponse | null>(null);
  const [priceSource, setPriceSource] = useState<string>('sell');
  const [loading, setLoading] = useState(true);
  const [expandedGroups, setExpandedGroups] = useState<Set<number>>(new Set());
  const hasFetchedRef = useRef(false);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchProfit('sell');
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session]);

  const fetchProfit = async (source: string) => {
    if (!session) return;
    setLoading(true);
    try {
      const response = await fetch(`/api/pi/profit?priceSource=${source}`);
      if (response.ok) {
        const data: PiProfitResponse = await response.json();
        setProfitResponse(data);
      }
    } finally {
      setLoading(false);
    }
  };

  const profitData = profitResponse?.planets || [];

  const handlePriceSourceChange = (_: React.MouseEvent<HTMLElement>, newSource: string | null) => {
    if (newSource) {
      setPriceSource(newSource);
      hasFetchedRef.current = false;
      fetchProfit(newSource);
      hasFetchedRef.current = true;
    }
  };

  const toggleGroup = (typeId: number) => {
    setExpandedGroups(prev => {
      const next = new Set(prev);
      if (next.has(typeId)) next.delete(typeId);
      else next.add(typeId);
      return next;
    });
  };

  // Group all factories across planets by output type
  const productGroups = useMemo(() => {
    const groupMap = new Map<number, ProductGroup>();

    for (const planet of profitData) {
      for (const factory of planet.factories) {
        let group = groupMap.get(factory.outputTypeId);
        if (!group) {
          group = {
            outputTypeId: factory.outputTypeId,
            outputName: factory.outputName,
            outputTier: factory.outputTier,
            totalRatePerHour: 0,
            totalOutputValue: 0,
            totalInputCost: 0,
            totalExportTax: 0,
            totalImportTax: 0,
            totalProfit: 0,
            factories: [],
          };
          groupMap.set(factory.outputTypeId, group);
        }
        group.totalRatePerHour += factory.ratePerHour;
        group.totalOutputValue += factory.outputValuePerHour;
        group.totalInputCost += factory.inputCostPerHour;
        group.totalExportTax += factory.exportTaxPerHour;
        group.totalImportTax += factory.importTaxPerHour;
        group.totalProfit += factory.profitPerHour;
        group.factories.push({
          ...factory,
          solarSystemName: planet.solarSystemName,
          characterName: planet.characterName,
          planetType: planet.planetType,
        });
      }
    }

    // Sort by tier then by name
    return Array.from(groupMap.values()).sort((a, b) => {
      const tierDiff = (TIER_ORDER[a.outputTier] ?? 99) - (TIER_ORDER[b.outputTier] ?? 99);
      if (tierDiff !== 0) return tierDiff;
      return a.outputName.localeCompare(b.outputName);
    });
  }, [profitData]);

  const totals = useMemo(() => {
    if (!profitResponse) return { output: 0, input: 0, exportTax: 0, importTax: 0, totalTax: 0, profit: 0 };
    return {
      output: profitResponse.totalOutputValue,
      input: profitResponse.totalInputCost,
      exportTax: profitResponse.totalExportTax,
      importTax: profitResponse.totalImportTax,
      totalTax: profitResponse.totalExportTax + profitResponse.totalImportTax,
      profit: profitResponse.totalProfit,
    };
  }, [profitResponse]);

  if (loading) {
    return <Loading />;
  }

  return (
    <Box>
      {/* Summary cards + controls */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2, flexWrap: 'wrap', gap: 2 }}>
        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
          <SummaryCard
            label="Revenue / hr"
            value={formatISK(totals.output)}
            color="#10b981"
            icon={<TrendingUpIcon sx={{ fontSize: 18, color: '#10b981' }} />}
          />
          <SummaryCard
            label="Costs / hr"
            value={formatISK(totals.input)}
            color="#ef4444"
            icon={<TrendingDownIcon sx={{ fontSize: 18, color: '#ef4444' }} />}
          />
          <SummaryCard
            label="Taxes / hr"
            value={formatISK(totals.totalTax)}
            color="#f59e0b"
          />
          <SummaryCard
            label="Profit / hr"
            value={formatISK(totals.profit)}
            color={profitColor(totals.profit)}
            bold
          />
        </Box>
        <ToggleButtonGroup
          value={priceSource}
          exclusive
          onChange={handlePriceSourceChange}
          size="small"
          sx={{
            '& .MuiToggleButton-root': {
              color: '#64748b',
              borderColor: 'rgba(148, 163, 184, 0.2)',
              textTransform: 'none',
              fontSize: '0.75rem',
              px: 1.5,
              py: 0.5,
              '&.Mui-selected': {
                color: '#3b82f6',
                bgcolor: 'rgba(59, 130, 246, 0.1)',
                borderColor: 'rgba(59, 130, 246, 0.3)',
              },
            },
          }}
        >
          <ToggleButton value="sell">Sell</ToggleButton>
          <ToggleButton value="buy">Buy</ToggleButton>
          <ToggleButton value="split">Split</ToggleButton>
        </ToggleButtonGroup>
      </Box>

      {/* Profit table grouped by product */}
      <TableContainer>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell sx={{ color: '#64748b', borderColor: 'rgba(148, 163, 184, 0.1)', bgcolor: '#0f1219', width: 40 }} />
              <TableCell sx={{ color: '#64748b', borderColor: 'rgba(148, 163, 184, 0.1)', bgcolor: '#0f1219' }}>Product</TableCell>
              <TableCell align="right" sx={{ color: '#64748b', borderColor: 'rgba(148, 163, 184, 0.1)', bgcolor: '#0f1219' }}>Rate</TableCell>
              <TableCell align="right" sx={{ color: '#64748b', borderColor: 'rgba(148, 163, 184, 0.1)', bgcolor: '#0f1219' }}>Revenue/hr</TableCell>
              <TableCell align="right" sx={{ color: '#64748b', borderColor: 'rgba(148, 163, 184, 0.1)', bgcolor: '#0f1219' }}>Costs/hr</TableCell>
              <TableCell align="right" sx={{ color: '#64748b', borderColor: 'rgba(148, 163, 184, 0.1)', bgcolor: '#0f1219' }}>Taxes/hr</TableCell>
              <TableCell align="right" sx={{ color: '#64748b', borderColor: 'rgba(148, 163, 184, 0.1)', bgcolor: '#0f1219' }}>Profit/hr</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {productGroups.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} align="center" sx={{ color: '#64748b', borderColor: 'rgba(148, 163, 184, 0.1)', py: 4 }}>
                  No PI profit data available
                </TableCell>
              </TableRow>
            ) : (
              productGroups.map((group) => (
                <ProductGroupRow
                  key={group.outputTypeId}
                  group={group}
                  expanded={expandedGroups.has(group.outputTypeId)}
                  onToggle={() => toggleGroup(group.outputTypeId)}
                />
              ))
            )}
            {productGroups.length > 0 && (
              <TableRow sx={{ bgcolor: '#0f1219' }}>
                <TableCell sx={{ borderColor: 'rgba(148, 163, 184, 0.1)' }} />
                <TableCell sx={{ color: '#e2e8f0', fontWeight: 600, borderColor: 'rgba(148, 163, 184, 0.1)' }}>
                  Total ({productGroups.length} products)
                </TableCell>
                <TableCell sx={{ borderColor: 'rgba(148, 163, 184, 0.1)' }} />
                <TableCell align="right" sx={{ color: '#10b981', fontWeight: 600, borderColor: 'rgba(148, 163, 184, 0.1)' }}>
                  {formatISK(totals.output)}
                </TableCell>
                <TableCell align="right" sx={{ color: '#ef4444', fontWeight: 600, borderColor: 'rgba(148, 163, 184, 0.1)' }}>
                  {formatISK(totals.input)}
                </TableCell>
                <TableCell align="right" sx={{ color: '#f59e0b', fontWeight: 600, borderColor: 'rgba(148, 163, 184, 0.1)' }}>
                  {formatISK(totals.totalTax)}
                </TableCell>
                <TableCell align="right" sx={{ color: profitColor(totals.profit), fontWeight: 700, borderColor: 'rgba(148, 163, 184, 0.1)' }}>
                  {formatISK(totals.profit)}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}

function SummaryCard({ label, value, color, icon, bold }: {
  label: string;
  value: string;
  color: string;
  icon?: React.ReactNode;
  bold?: boolean;
}) {
  return (
    <Card sx={{
      background: `linear-gradient(135deg, ${color}08 0%, #12151f 100%)`,
      border: `1px solid ${color}25`,
      borderRadius: 2,
      minWidth: 140,
    }}>
      <CardContent sx={{ p: 1.5, '&:last-child': { pb: 1.5 } }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mb: 0.5 }}>
          {icon}
          <Typography variant="caption" sx={{ color: '#64748b' }}>{label}</Typography>
        </Box>
        <Typography variant="body2" sx={{ color, fontWeight: bold ? 700 : 600 }}>
          {value}
        </Typography>
      </CardContent>
    </Card>
  );
}

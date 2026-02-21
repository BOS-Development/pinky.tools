import { useState, useEffect, useRef, useMemo } from 'react';
import { useSession } from "next-auth/react";
import Navbar from "@industry-tool/components/Navbar";
import Loading from "@industry-tool/components/loading";
import LaunchpadDetail from "@industry-tool/components/pi/LaunchpadDetail";
import { PiPlanet, PiPlanetsResponse } from "@industry-tool/client/data/models";
import { formatNumber } from "@industry-tool/utils/formatting";
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import Grid from '@mui/material/Grid';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import SearchIcon from '@mui/icons-material/Search';
import PublicIcon from '@mui/icons-material/Public';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import ErrorIcon from '@mui/icons-material/Error';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import AccessTimeIcon from '@mui/icons-material/AccessTime';

const PLANET_TYPE_IDS: Record<string, number> = {
  temperate: 11,
  ice: 12,
  gas: 13,
  oceanic: 2014,
  lava: 2015,
  barren: 2016,
  storm: 2017,
  plasma: 2063,
};

const PLANET_TYPE_COLORS: Record<string, string> = {
  temperate: '#10b981',
  barren: '#94a3b8',
  oceanic: '#3b82f6',
  ice: '#67e8f9',
  gas: '#f59e0b',
  lava: '#ef4444',
  storm: '#8b5cf6',
  plasma: '#ec4899',
};

function getStatusColor(status: string): string {
  switch (status) {
    case 'running': return '#10b981';
    case 'expired': return '#ef4444';
    case 'stalled': return '#f59e0b';
    case 'stale_data': return '#94a3b8';
    default: return '#94a3b8';
  }
}

function getStatusLabel(status: string): string {
  switch (status) {
    case 'running': return 'Running';
    case 'expired': return 'Expired';
    case 'stalled': return 'Stalled';
    case 'stale_data': return 'Stale Data';
    default: return status;
  }
}

function getStatusIcon(status: string) {
  switch (status) {
    case 'running': return <CheckCircleIcon sx={{ fontSize: 14 }} />;
    case 'expired': return <ErrorIcon sx={{ fontSize: 14 }} />;
    case 'stalled': return <WarningAmberIcon sx={{ fontSize: 14 }} />;
    case 'stale_data': return <AccessTimeIcon sx={{ fontSize: 14 }} />;
    default: return null;
  }
}

function formatTimeAgo(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffHours / 24);

  if (diffDays > 0) return `${diffDays}d ${diffHours % 24}h ago`;
  if (diffHours > 0) return `${diffHours}h ago`;
  const diffMins = Math.floor(diffMs / (1000 * 60));
  return `${diffMins}m ago`;
}

function formatTimeUntil(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = date.getTime() - now.getTime();
  if (diffMs <= 0) return 'Expired';
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffHours / 24);

  if (diffDays > 0) return `${diffDays}d ${diffHours % 24}h`;
  if (diffHours > 0) return `${diffHours}h ${Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))}m`;
  return `${Math.floor(diffMs / (1000 * 60))}m`;
}

type LaunchpadSelection = {
  characterId: number;
  planetId: number;
  pinId: number;
  planetName: string;
};

function PlanetCard({ planet, onLaunchpadClick }: { planet: PiPlanet; onLaunchpadClick: (sel: LaunchpadSelection) => void }) {
  const typeColor = PLANET_TYPE_COLORS[planet.planetType] || '#94a3b8';
  const statusColor = getStatusColor(planet.status);
  const hasIssues = planet.status !== 'running';

  return (
    <Card
      sx={{
        background: hasIssues
          ? `linear-gradient(135deg, rgba(239, 68, 68, 0.05) 0%, #12151f 100%)`
          : `linear-gradient(135deg, rgba(59, 130, 246, 0.05) 0%, #12151f 100%)`,
        border: `1px solid ${hasIssues ? 'rgba(239, 68, 68, 0.2)' : 'rgba(59, 130, 246, 0.15)'}`,
        borderRadius: 2,
        height: '100%',
      }}
    >
      <CardContent sx={{ p: 2, '&:last-child': { pb: 2 } }}>
        {/* Header */}
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 1.5 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, minWidth: 0 }}>
            <img
              src={`https://images.evetech.net/types/${PLANET_TYPE_IDS[planet.planetType] || 11}/icon?size=64`}
              alt={planet.planetType}
              width={28}
              height={28}
              style={{ flexShrink: 0, borderRadius: '50%' }}
            />
            <Box sx={{ minWidth: 0 }}>
              <Typography variant="body2" sx={{ color: '#e2e8f0', fontWeight: 600, lineHeight: 1.2 }} noWrap>
                {planet.solarSystemName}
              </Typography>
              <Typography variant="caption" sx={{ color: '#64748b' }}>
                {planet.planetType.charAt(0).toUpperCase() + planet.planetType.slice(1)} - {planet.characterName}
              </Typography>
            </Box>
          </Box>
          <Chip
            icon={getStatusIcon(planet.status) || undefined}
            label={getStatusLabel(planet.status)}
            size="small"
            sx={{
              bgcolor: `${statusColor}20`,
              color: statusColor,
              border: `1px solid ${statusColor}40`,
              fontWeight: 600,
              fontSize: '0.7rem',
              height: 24,
              flexShrink: 0,
              '& .MuiChip-icon': { color: statusColor },
            }}
          />
        </Box>

        {/* Extractors */}
        {planet.extractors.length > 0 && (
          <Box sx={{ mb: 1 }}>
            <Typography variant="caption" sx={{ color: '#64748b', fontWeight: 600, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              Extractors
            </Typography>
            {planet.extractors.map((ext) => (
              <Box key={ext.pinId} sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 0.25 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, minWidth: 0 }}>
                  <img src={`https://images.evetech.net/types/${ext.productTypeId}/icon?size=32`} alt="" width={16} height={16} style={{ flexShrink: 0 }} />
                  <Typography variant="caption" sx={{ color: ext.status === 'expired' ? '#ef4444' : '#cbd5e1' }} noWrap>
                    {ext.productName || `Type ${ext.productTypeId}`}
                  </Typography>
                </Box>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Typography variant="caption" sx={{ color: '#94a3b8' }}>
                    {formatNumber(Math.round(ext.ratePerHour))}/hr
                  </Typography>
                  {ext.expiryTime && (
                    <Typography variant="caption" sx={{ color: ext.status === 'expired' ? '#ef4444' : '#10b981', fontWeight: 500 }}>
                      {formatTimeUntil(ext.expiryTime)}
                    </Typography>
                  )}
                </Box>
              </Box>
            ))}
          </Box>
        )}

        {/* Factories */}
        {planet.factories.length > 0 && (
          <Box sx={{ mb: 1 }}>
            <Typography variant="caption" sx={{ color: '#64748b', fontWeight: 600, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              Factories ({planet.factories.length})
            </Typography>
            {/* Group by schematic */}
            {Object.entries(
              planet.factories.reduce<Record<string, { count: number; ratePerHour: number; status: string; outputTypeId: number }>>((acc, f) => {
                const key = f.schematicName || `Unknown (${f.schematicId})`;
                if (!acc[key]) acc[key] = { count: 0, ratePerHour: 0, status: 'running', outputTypeId: f.outputTypeId };
                acc[key].count++;
                acc[key].ratePerHour += f.ratePerHour;
                if (f.status === 'stalled') acc[key].status = 'stalled';
                return acc;
              }, {})
            ).map(([name, info]) => (
              <Box key={name} sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 0.25 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, minWidth: 0 }}>
                  {info.outputTypeId > 0 && (
                    <img src={`https://images.evetech.net/types/${info.outputTypeId}/icon?size=32`} alt="" width={16} height={16} style={{ flexShrink: 0 }} />
                  )}
                  <Typography variant="caption" sx={{ color: info.status === 'stalled' ? '#f59e0b' : '#cbd5e1' }} noWrap>
                    {info.count}x {name}
                  </Typography>
                </Box>
                <Typography variant="caption" sx={{ color: '#94a3b8' }}>
                  {formatNumber(Math.round(info.ratePerHour))}/hr
                </Typography>
              </Box>
            ))}
          </Box>
        )}

        {/* Launchpads */}
        {planet.launchpads.length > 0 && (
          <Box sx={{ mb: 1 }}>
            <Typography variant="caption" sx={{ color: '#64748b', fontWeight: 600, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              Storage ({planet.launchpads.length})
            </Typography>
            {planet.launchpads.map((lp) => (
              <Box
                key={lp.pinId}
                onClick={() => onLaunchpadClick({
                  characterId: planet.characterId,
                  planetId: planet.planetId,
                  pinId: lp.pinId,
                  planetName: `${planet.solarSystemName} - ${planet.planetType.charAt(0).toUpperCase() + planet.planetType.slice(1)}`,
                })}
                sx={{
                  mt: 0.5,
                  px: 0.75,
                  py: 0.5,
                  borderRadius: 0.75,
                  cursor: 'pointer',
                  border: '1px solid transparent',
                  '&:hover': { bgcolor: 'rgba(59, 130, 246, 0.06)', borderColor: 'rgba(59, 130, 246, 0.15)' },
                }}
              >
                <Typography variant="caption" sx={{ color: lp.label ? '#94a3b8' : '#475569', fontWeight: lp.label ? 500 : 400, fontStyle: lp.label ? 'normal' : 'italic' }}>
                  {lp.label || `Launchpad`}
                </Typography>
                {lp.contents.length === 0 ? (
                  <Typography variant="caption" sx={{ color: '#475569', display: 'block', fontSize: '0.65rem' }}>
                    Empty
                  </Typography>
                ) : (
                  lp.contents.sort((a, b) => b.amount - a.amount).slice(0, 3).map((item) => (
                    <Box key={item.typeId} sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, minWidth: 0 }}>
                        <img src={`https://images.evetech.net/types/${item.typeId}/icon?size=32`} alt="" width={14} height={14} style={{ flexShrink: 0 }} />
                        <Typography variant="caption" sx={{ color: '#cbd5e1', fontSize: '0.65rem' }} noWrap>
                          {item.name || `Type ${item.typeId}`}
                        </Typography>
                      </Box>
                      <Typography variant="caption" sx={{ color: '#94a3b8', flexShrink: 0, ml: 1, fontSize: '0.65rem' }}>
                        {formatNumber(item.amount)}
                      </Typography>
                    </Box>
                  ))
                )}
                {lp.contents.length > 3 && (
                  <Typography variant="caption" sx={{ color: '#475569', fontSize: '0.6rem' }}>
                    +{lp.contents.length - 3} more
                  </Typography>
                )}
              </Box>
            ))}
          </Box>
        )}

        {/* Footer */}
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 1, pt: 1, borderTop: '1px solid rgba(148, 163, 184, 0.1)' }}>
          <Typography variant="caption" sx={{ color: '#475569' }}>
            CC{planet.upgradeLevel}
          </Typography>
          <Typography variant="caption" sx={{ color: '#475569' }}>
            Updated {formatTimeAgo(planet.lastUpdate)}
          </Typography>
        </Box>
      </CardContent>
    </Card>
  );
}

export default function PlanetOverview({ embedded }: { embedded?: boolean }) {
  const { data: session } = useSession();
  const [planets, setPlanets] = useState<PiPlanet[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const hasFetchedRef = useRef(false);
  const [selectedLaunchpad, setSelectedLaunchpad] = useState<LaunchpadSelection | null>(null);

  useEffect(() => {
    if (session && !hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchPlanets();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session]);

  const fetchPlanets = async () => {
    if (!session) return;
    setLoading(true);
    try {
      const response = await fetch('/api/pi/planets');
      if (response.ok) {
        const data: PiPlanetsResponse = await response.json();
        setPlanets(data.planets || []);
      }
    } finally {
      setLoading(false);
    }
  };

  const filteredPlanets = useMemo(() => {
    if (!searchQuery) return planets;
    const query = searchQuery.toLowerCase();
    return planets.filter(
      (p) =>
        p.solarSystemName.toLowerCase().includes(query) ||
        p.characterName.toLowerCase().includes(query) ||
        p.planetType.toLowerCase().includes(query)
    );
  }, [planets, searchQuery]);

  const stats = useMemo(() => {
    const total = planets.length;
    const running = planets.filter(p => p.status === 'running').length;
    const stalled = planets.filter(p => p.status === 'stalled' || p.status === 'expired').length;
    const stale = planets.filter(p => p.status === 'stale_data').length;
    const totalExtractors = planets.reduce((sum, p) => sum + p.extractors.length, 0);
    const totalFactories = planets.reduce((sum, p) => sum + p.factories.length, 0);
    return { total, running, stalled, stale, totalExtractors, totalFactories };
  }, [planets]);

  if (loading) {
    if (embedded) return <Loading />;
    return (
      <>
        <Navbar />
        <Loading />
      </>
    );
  }

  const content = (
    <>
        {/* Stats Row */}
        <Box sx={{ display: 'flex', gap: 2, mb: 2, flexWrap: 'wrap' }}>
          <StatChip label="Planets" value={stats.total} color="#3b82f6" />
          <StatChip label="Running" value={stats.running} color="#10b981" />
          {stats.stalled > 0 && <StatChip label="Issues" value={stats.stalled} color="#ef4444" />}
          {stats.stale > 0 && <StatChip label="Stale" value={stats.stale} color="#94a3b8" />}
          <StatChip label="Extractors" value={stats.totalExtractors} color="#f59e0b" />
          <StatChip label="Factories" value={stats.totalFactories} color="#8b5cf6" />
        </Box>

        {/* Search */}
        <TextField
          size="small"
          placeholder="Search planets..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon sx={{ color: '#64748b', fontSize: 20 }} />
                </InputAdornment>
              ),
            },
          }}
          sx={{
            mb: 2,
            width: 300,
            '& .MuiOutlinedInput-root': {
              bgcolor: '#12151f',
              '& fieldset': { borderColor: 'rgba(148, 163, 184, 0.15)' },
              '&:hover fieldset': { borderColor: 'rgba(59, 130, 246, 0.3)' },
            },
            '& .MuiInputBase-input': { color: '#e2e8f0', fontSize: '0.875rem' },
          }}
        />

        {/* Planet Grid */}
        {filteredPlanets.length === 0 ? (
          <Box sx={{ textAlign: 'center', py: 8 }}>
            <PublicIcon sx={{ fontSize: 48, color: '#475569', mb: 1 }} />
            <Typography variant="body1" sx={{ color: '#64748b' }}>
              {planets.length === 0
                ? 'No planets found. Make sure your characters have the PI scope and data has been refreshed.'
                : 'No planets match your search.'}
            </Typography>
          </Box>
        ) : (
          <Grid container spacing={2}>
            {filteredPlanets.map((planet) => (
              <Grid key={`${planet.characterId}-${planet.planetId}`} size={{ xs: 12, sm: 6, md: 4, lg: 3 }}>
                <PlanetCard planet={planet} onLaunchpadClick={setSelectedLaunchpad} />
              </Grid>
            ))}
          </Grid>
        )}
    </>
  );

  const handleLabelChange = (characterId: number, planetId: number, pinId: number, label: string) => {
    setPlanets(prev => prev.map(p => {
      if (p.characterId !== characterId || p.planetId !== planetId) return p;
      return {
        ...p,
        launchpads: p.launchpads.map(lp =>
          lp.pinId === pinId ? { ...lp, label: label || undefined } : lp
        ),
      };
    }));
  };

  const drawer = (
    <LaunchpadDetail
      open={selectedLaunchpad !== null}
      onClose={() => setSelectedLaunchpad(null)}
      characterId={selectedLaunchpad?.characterId ?? 0}
      planetId={selectedLaunchpad?.planetId ?? 0}
      pinId={selectedLaunchpad?.pinId ?? 0}
      planetName={selectedLaunchpad?.planetName ?? ''}
      onLabelChange={handleLabelChange}
    />
  );

  if (embedded) return <>{content}{drawer}</>;

  return (
    <>
      <Navbar />
      <Container maxWidth="xl" sx={{ mt: 2, mb: 4 }}>
        <Typography variant="h5" sx={{ color: '#e2e8f0', mb: 2, fontWeight: 600 }}>
          Planetary Industry
        </Typography>
        {content}
      </Container>
      {drawer}
    </>
  );
}

function StatChip({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        gap: 0.75,
        px: 1.5,
        py: 0.5,
        borderRadius: 1,
        bgcolor: `${color}10`,
        border: `1px solid ${color}30`,
      }}
    >
      <Typography variant="caption" sx={{ color: '#94a3b8', fontWeight: 500 }}>
        {label}
      </Typography>
      <Typography variant="body2" sx={{ color, fontWeight: 700 }}>
        {value}
      </Typography>
    </Box>
  );
}

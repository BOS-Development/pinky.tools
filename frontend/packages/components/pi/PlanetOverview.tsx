import { useState, useEffect, useRef, useMemo } from 'react';
import { useSession } from "next-auth/react";
import Navbar from "@industry-tool/components/Navbar";
import Loading from "@industry-tool/components/loading";
import LaunchpadDetail from "@industry-tool/components/pi/LaunchpadDetail";
import { PiPlanet, PiPlanetsResponse } from "@industry-tool/client/data/models";
import { formatNumber } from "@industry-tool/utils/formatting";
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';
import { Search, Globe, CheckCircle, XCircle, AlertTriangle, Clock } from 'lucide-react';

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
  oceanic: '#00d4ff',
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
  const cls = "w-3 h-3";
  switch (status) {
    case 'running': return <CheckCircle className={cls} />;
    case 'expired': return <XCircle className={cls} />;
    case 'stalled': return <AlertTriangle className={cls} />;
    case 'stale_data': return <Clock className={cls} />;
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
  const statusColor = getStatusColor(planet.status);
  const hasIssues = planet.status !== 'running';

  return (
    <Card
      className={cn(
        'h-full',
        hasIssues
          ? 'border-[rgba(244,63,94,0.2)]'
          : 'border-[rgba(0,212,255,0.08)]'
      )}
      style={{ background: '#12151f' }}
    >
      <CardContent className="p-3 pb-3">
        {/* Header */}
        <div className="flex justify-between items-start mb-3">
          <div className="flex items-center gap-2 min-w-0">
            <img
              src={`https://images.evetech.net/types/${PLANET_TYPE_IDS[planet.planetType] || 11}/icon?size=64`}
              alt={planet.planetType}
              width={28}
              height={28}
              className="flex-shrink-0 rounded-full"
            />
            <div className="min-w-0">
              <p className="text-sm font-semibold text-[#e2e8f0] leading-tight truncate">
                {planet.solarSystemName}
              </p>
              <span className="text-xs text-[#64748b]">
                {planet.planetType.charAt(0).toUpperCase() + planet.planetType.slice(1)} - {planet.characterName}
              </span>
            </div>
          </div>
          {/* Status badge */}
          <span
            className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[0.7rem] font-semibold flex-shrink-0"
            style={{
              backgroundColor: `${statusColor}20`,
              color: statusColor,
              border: `1px solid ${statusColor}40`,
            }}
          >
            {getStatusIcon(planet.status)}
            {getStatusLabel(planet.status)}
          </span>
        </div>

        {/* Extractors */}
        {planet.extractors.length > 0 && (
          <div className="mb-2">
            <span className="text-xs text-[#64748b] font-semibold uppercase tracking-wide">
              Extractors
            </span>
            {planet.extractors.map((ext) => (
              <div key={ext.pinId} className="flex justify-between items-center mt-0.5">
                <div className="flex items-center gap-1 min-w-0">
                  <img src={`https://images.evetech.net/types/${ext.productTypeId}/icon?size=32`} alt="" width={16} height={16} className="flex-shrink-0" />
                  <span className={cn("text-xs truncate", ext.status === 'expired' ? 'text-[#ef4444]' : 'text-[#cbd5e1]')}>
                    {ext.productName || `Type ${ext.productTypeId}`}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-[#94a3b8]">
                    {formatNumber(Math.round(ext.ratePerHour))}/hr
                  </span>
                  {ext.expiryTime && (
                    <span className={cn("text-xs font-medium", ext.status === 'expired' ? 'text-[#ef4444]' : 'text-[#10b981]')}>
                      {formatTimeUntil(ext.expiryTime)}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Factories */}
        {planet.factories.length > 0 && (
          <div className="mb-2">
            <span className="text-xs text-[#64748b] font-semibold uppercase tracking-wide">
              Factories ({planet.factories.length})
            </span>
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
              <div key={name} className="flex justify-between items-center mt-0.5">
                <div className="flex items-center gap-1 min-w-0">
                  {info.outputTypeId > 0 && (
                    <img src={`https://images.evetech.net/types/${info.outputTypeId}/icon?size=32`} alt="" width={16} height={16} className="flex-shrink-0" />
                  )}
                  <span className={cn("text-xs truncate", info.status === 'stalled' ? 'text-[#f59e0b]' : 'text-[#cbd5e1]')}>
                    {info.count}x {name}
                  </span>
                </div>
                <span className="text-xs text-[#94a3b8]">
                  {formatNumber(Math.round(info.ratePerHour))}/hr
                </span>
              </div>
            ))}
          </div>
        )}

        {/* Launchpads */}
        {planet.launchpads.length > 0 && (
          <div className="mb-2">
            <span className="text-xs text-[#64748b] font-semibold uppercase tracking-wide">
              Storage ({planet.launchpads.length})
            </span>
            {planet.launchpads.map((lp) => (
              <div
                key={lp.pinId}
                onClick={() => onLaunchpadClick({
                  characterId: planet.characterId,
                  planetId: planet.planetId,
                  pinId: lp.pinId,
                  planetName: `${planet.solarSystemName} - ${planet.planetType.charAt(0).toUpperCase() + planet.planetType.slice(1)}`,
                })}
                className="mt-1 px-1.5 py-1 rounded cursor-pointer border border-transparent hover:bg-[rgba(0,212,255,0.06)] hover:border-[rgba(0,212,255,0.15)] transition-colors"
              >
                <span className={cn("text-xs", lp.label ? 'text-[#94a3b8] font-medium' : 'text-[#475569] italic')}>
                  {lp.label || 'Launchpad'}
                </span>
                {lp.contents.length === 0 ? (
                  <span className="text-[#475569] block" style={{ fontSize: '0.65rem' }}>
                    Empty
                  </span>
                ) : (
                  lp.contents.sort((a, b) => b.amount - a.amount).slice(0, 3).map((item) => (
                    <div key={item.typeId} className="flex justify-between items-center">
                      <div className="flex items-center gap-1 min-w-0">
                        <img src={`https://images.evetech.net/types/${item.typeId}/icon?size=32`} alt="" width={14} height={14} className="flex-shrink-0" />
                        <span className="text-[#cbd5e1] truncate" style={{ fontSize: '0.65rem' }}>
                          {item.name || `Type ${item.typeId}`}
                        </span>
                      </div>
                      <span className="text-[#94a3b8] flex-shrink-0 ml-2" style={{ fontSize: '0.65rem' }}>
                        {formatNumber(item.amount)}
                      </span>
                    </div>
                  ))
                )}
                {lp.contents.length > 3 && (
                  <span className="text-[#475569]" style={{ fontSize: '0.6rem' }}>
                    +{lp.contents.length - 3} more
                  </span>
                )}
              </div>
            ))}
          </div>
        )}

        {/* Footer */}
        <div className="flex justify-between items-center mt-2 pt-2 border-t border-[rgba(148,163,184,0.1)]">
          <span className="text-xs text-[#475569]">CC{planet.upgradeLevel}</span>
          <span className="text-xs text-[#475569]">Updated {formatTimeAgo(planet.lastUpdate)}</span>
        </div>
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
      <div className="flex gap-2 mb-4 flex-wrap">
        <StatChip label="Planets" value={stats.total} color="#00d4ff" />
        <StatChip label="Running" value={stats.running} color="#10b981" />
        {stats.stalled > 0 && <StatChip label="Issues" value={stats.stalled} color="#ef4444" />}
        {stats.stale > 0 && <StatChip label="Stale" value={stats.stale} color="#94a3b8" />}
        <StatChip label="Extractors" value={stats.totalExtractors} color="#f59e0b" />
        <StatChip label="Factories" value={stats.totalFactories} color="#8b5cf6" />
      </div>

      {/* Search */}
      <div className="relative mb-4 w-[300px]">
        <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-[#64748b]" />
        <Input
          placeholder="Search planets..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-8 h-8 text-sm bg-[#12151f] border-[rgba(148,163,184,0.15)] text-[#e2e8f0] focus-visible:ring-0 focus-visible:border-[rgba(0,212,255,0.3)] hover:border-[rgba(0,212,255,0.3)]"
        />
      </div>

      {/* Planet Grid */}
      {filteredPlanets.length === 0 ? (
        <div className="empty-state">
          <Globe className="empty-state-icon" />
          <p className="empty-state-title">
            {planets.length === 0
              ? 'No planets found. Make sure your characters have the PI scope and data has been refreshed.'
              : 'No planets match your search.'}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {filteredPlanets.map((planet) => (
            <PlanetCard
              key={`${planet.characterId}-${planet.planetId}`}
              planet={planet}
              onLaunchpadClick={setSelectedLaunchpad}
            />
          ))}
        </div>
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
      <div className="max-w-screen-xl mx-auto px-4 mt-2 mb-8">
        <h2 className="text-xl font-semibold text-[#e2e8f0] mb-4">
          Planetary Industry
        </h2>
        {content}
      </div>
      {drawer}
    </>
  );
}

function StatChip({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div
      className="flex items-center gap-1.5 px-3 py-1 rounded"
      style={{
        backgroundColor: `${color}10`,
        border: `1px solid ${color}30`,
      }}
    >
      <span className="text-xs text-[#94a3b8] font-medium">{label}</span>
      <span className="text-sm font-bold" style={{ color }}>{value}</span>
    </div>
  );
}

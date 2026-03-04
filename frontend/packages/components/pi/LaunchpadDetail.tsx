import { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import {
  LaunchpadDetailResponse,
  PiPinContent,
} from '@industry-tool/client/data/models';
import { formatNumber } from '@industry-tool/utils/formatting';
import { Separator } from '@/components/ui/separator';
import { Button } from '@/components/ui/button';
import { X, Rocket, Factory, Package, Edit, Loader2 } from 'lucide-react';

type LaunchpadDetailProps = {
  open: boolean;
  onClose: () => void;
  characterId: number;
  planetId: number;
  pinId: number;
  planetName: string;
  onLabelChange?: (characterId: number, planetId: number, pinId: number, label: string) => void;
};

function formatDepletion(hours: number): string {
  if (hours <= 0) return 'Empty';
  const days = Math.floor(hours / 24);
  const h = Math.floor(hours % 24);
  if (days > 0) return `${days}d ${h}h`;
  const m = Math.round((hours % 1) * 60);
  return `${h}h ${m}m`;
}

function depletionColor(hours: number): string {
  if (hours <= 0) return '#ef4444';
  if (hours < 24) return '#ef4444';
  if (hours < 72) return '#f59e0b';
  return '#10b981';
}

export default function LaunchpadDetail({
  open,
  onClose,
  characterId,
  planetId,
  pinId,
  planetName,
  onLabelChange,
}: LaunchpadDetailProps) {
  const [data, setData] = useState<LaunchpadDetailResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [label, setLabel] = useState('');
  const [editing, setEditing] = useState(false);
  const [editValue, setEditValue] = useState('');
  const [savingLabel, setSavingLabel] = useState(false);
  const labelInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      fetchDetail();
    } else {
      setData(null);
      setError(null);
      setEditing(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, characterId, planetId, pinId]);

  useEffect(() => {
    if (data) {
      setLabel(data.label || '');
    }
  }, [data]);

  useEffect(() => {
    if (editing && labelInputRef.current) {
      labelInputRef.current.focus();
    }
  }, [editing]);

  const fetchDetail = async () => {
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams({
        characterId: String(characterId),
        planetId: String(planetId),
        pinId: String(pinId),
      });
      const response = await fetch(`/api/pi/launchpad-detail?${params}`);
      if (response.ok) {
        const result: LaunchpadDetailResponse = await response.json();
        setData(result);
      } else {
        setError('Failed to load launchpad details');
      }
    } catch {
      setError('Failed to load launchpad details');
    } finally {
      setLoading(false);
    }
  };

  const startEditing = useCallback(() => {
    setEditValue(label);
    setEditing(true);
  }, [label]);

  const cancelEditing = useCallback(() => {
    setEditing(false);
    setEditValue('');
  }, []);

  const saveLabel = useCallback(async () => {
    const trimmed = editValue.trim();
    setSavingLabel(true);

    try {
      if (trimmed === '') {
        await fetch('/api/pi/launchpad-labels', {
          method: 'DELETE',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ characterId, planetId, pinId }),
        });
        setLabel('');
        onLabelChange?.(characterId, planetId, pinId, '');
      } else {
        await fetch('/api/pi/launchpad-labels', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ characterId, planetId, pinId, label: trimmed }),
        });
        setLabel(trimmed);
        onLabelChange?.(characterId, planetId, pinId, trimmed);
      }
    } catch {
      // Silently fail, keep old label
    } finally {
      setSavingLabel(false);
      setEditing(false);
    }
  }, [editValue, characterId, planetId, pinId, onLabelChange]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        saveLabel();
      } else if (e.key === 'Escape') {
        cancelEditing();
      }
    },
    [saveLabel, cancelEditing]
  );

  const sortedContents = data
    ? [...data.contents].sort((a, b) => b.amount - a.amount)
    : [];

  // Aggregate input requirements across all connected factories
  const aggregatedInputs = useMemo(() => {
    if (!data) return [];
    const map = new Map<number, { typeId: number; name: string; consumedPerHour: number; currentStock: number; depletionHours: number }>();
    for (const factory of data.factories) {
      for (const inp of factory.inputs) {
        const existing = map.get(inp.typeId);
        if (existing) {
          existing.consumedPerHour += inp.consumedPerHour;
        } else {
          map.set(inp.typeId, {
            typeId: inp.typeId,
            name: inp.name,
            consumedPerHour: inp.consumedPerHour,
            currentStock: inp.currentStock,
            depletionHours: 0,
          });
        }
      }
    }
    // Recalculate depletion from aggregated consumption
    for (const item of map.values()) {
      if (item.consumedPerHour > 0 && item.currentStock > 0) {
        item.depletionHours = item.currentStock / item.consumedPerHour;
      }
    }
    return Array.from(map.values()).sort((a, b) => a.depletionHours - b.depletionHours);
  }, [data]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      {/* Panel */}
      <div className="absolute right-0 top-0 h-full w-[450px] max-w-full bg-[#0a0e1a] border-l border-[rgba(0,212,255,0.15)] flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-[rgba(148,163,184,0.1)] bg-[#0a0e1a]">
          <div className="flex items-center gap-3 min-w-0 flex-1">
            <Rocket className="w-5 h-5 text-[#00d4ff] flex-shrink-0" />
            <div className="min-w-0 flex-1">
              <p className="text-sm font-semibold text-[#e2e8f0] leading-snug truncate">
                {planetName}
              </p>
              {/* Editable label */}
              {editing ? (
                <input
                  ref={labelInputRef}
                  value={editValue}
                  onChange={(e) => setEditValue(e.target.value)}
                  onBlur={saveLabel}
                  onKeyDown={handleKeyDown}
                  disabled={savingLabel}
                  placeholder="Enter label..."
                  className="bg-transparent border-b border-[rgba(0,212,255,0.3)] text-[#94a3b8] text-xs outline-none w-full py-0.5 mt-0.5 focus:border-[#00d4ff]"
                />
              ) : (
                <div
                  onClick={startEditing}
                  className="flex items-center gap-1 cursor-pointer group"
                >
                  <span className={label ? 'text-xs text-[#94a3b8]' : 'text-xs text-[#475569] italic'}>
                    {label || 'Add label...'}
                  </span>
                  <Edit className="w-3 h-3 text-[#475569] opacity-0 group-hover:opacity-100 transition-opacity" />
                </div>
              )}
            </div>
          </div>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-[#64748b] hover:text-[#e2e8f0] ml-2"
            onClick={onClose}
          >
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Scrollable content */}
        <div className="flex-1 overflow-y-auto px-6 py-4">
          {loading && (
            <div className="flex justify-center items-center py-16">
              <Loader2 className="h-8 w-8 animate-spin text-[#00d4ff]" />
            </div>
          )}

          {error && !loading && (
            <div className="text-center py-16">
              <p className="text-sm text-[#ef4444]">{error}</p>
            </div>
          )}

          {data && !loading && (
            <>
              {/* Input Requirements */}
              <SectionHeader
                icon={<Factory className="w-4 h-4 text-[#00d4ff]" />}
                label={`Input Requirements (${aggregatedInputs.length})`}
              />
              {aggregatedInputs.length === 0 ? (
                <span className="text-xs text-[#475569] block mb-4">
                  No factory inputs tracked
                </span>
              ) : (
                <div className="mb-6">
                  {aggregatedInputs.map((input) => (
                    <InputRow key={input.typeId} input={input} />
                  ))}
                </div>
              )}

              <Separator className="bg-[rgba(148,163,184,0.08)] mb-4" />

              {/* Current Contents */}
              <SectionHeader
                icon={<Package className="w-4 h-4 text-[#00d4ff]" />}
                label={`Current Contents (${sortedContents.length})`}
              />
              {sortedContents.length === 0 ? (
                <span className="text-xs text-[#475569]">Launchpad is empty</span>
              ) : (
                <div>
                  {sortedContents.map((item) => (
                    <ContentRow key={item.typeId} item={item} />
                  ))}
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function SectionHeader({ icon, label }: { icon: React.ReactNode; label: string }) {
  return (
    <div className="flex items-center gap-1.5 mb-3">
      {icon}
      <span className="text-xs text-[#64748b] font-semibold uppercase tracking-wide">
        {label}
      </span>
    </div>
  );
}

type AggregatedInput = {
  typeId: number;
  name: string;
  consumedPerHour: number;
  currentStock: number;
  depletionHours: number;
};

function InputRow({ input }: { input: AggregatedInput }) {
  const color = depletionColor(input.depletionHours);

  return (
    <div className="flex items-center justify-between py-1.5 border-b border-[rgba(148,163,184,0.05)] last:border-b-0">
      <div className="flex items-center gap-1.5 min-w-0 flex-1">
        <img
          src={`https://images.evetech.net/types/${input.typeId}/icon?size=32`}
          alt=""
          width={18}
          height={18}
          className="flex-shrink-0"
        />
        <div className="min-w-0">
          <span className="text-xs text-[#cbd5e1] block truncate">{input.name}</span>
          <span className="text-[#475569]" style={{ fontSize: '0.65rem' }}>
            {formatNumber(input.consumedPerHour)}/hr
          </span>
        </div>
      </div>
      <div className="text-right flex-shrink-0 ml-2">
        <span className="text-xs text-[#94a3b8] block">{formatNumber(input.currentStock)}</span>
        <span className="font-semibold" style={{ color, fontSize: '0.65rem' }}>
          {formatDepletion(input.depletionHours)}
        </span>
      </div>
    </div>
  );
}

function ContentRow({ item }: { item: PiPinContent }) {
  return (
    <div className="flex items-center justify-between py-1.5 border-b border-[rgba(148,163,184,0.05)] last:border-b-0">
      <div className="flex items-center gap-1.5 min-w-0">
        <img
          src={`https://images.evetech.net/types/${item.typeId}/icon?size=32`}
          alt=""
          width={18}
          height={18}
          className="flex-shrink-0"
        />
        <span className="text-xs text-[#cbd5e1] truncate">
          {item.name || `Type ${item.typeId}`}
        </span>
      </div>
      <span className="text-xs text-[#94a3b8] font-medium flex-shrink-0 ml-2">
        {formatNumber(item.amount)}
      </span>
    </div>
  );
}

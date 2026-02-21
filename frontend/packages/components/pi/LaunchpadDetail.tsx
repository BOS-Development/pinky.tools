import { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import {
  LaunchpadDetailResponse,
  PiPinContent,
} from '@industry-tool/client/data/models';
import { formatNumber } from '@industry-tool/utils/formatting';
import Drawer from '@mui/material/Drawer';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import CircularProgress from '@mui/material/CircularProgress';
import TextField from '@mui/material/TextField';
import CloseIcon from '@mui/icons-material/Close';
import RocketLaunchIcon from '@mui/icons-material/RocketLaunch';
import FactoryIcon from '@mui/icons-material/Factory';
import InventoryIcon from '@mui/icons-material/Inventory';
import EditIcon from '@mui/icons-material/Edit';

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

  return (
    <Drawer
      anchor="right"
      open={open}
      onClose={onClose}
      PaperProps={{
        sx: {
          width: 450,
          maxWidth: '100vw',
          bgcolor: '#0a0e1a',
          borderLeft: '1px solid rgba(59, 130, 246, 0.15)',
        },
      }}
    >
      {/* Header */}
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          px: 2.5,
          py: 2,
          borderBottom: '1px solid rgba(148, 163, 184, 0.1)',
          background: 'linear-gradient(135deg, rgba(59, 130, 246, 0.05) 0%, #0a0e1a 100%)',
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, minWidth: 0, flex: 1 }}>
          <RocketLaunchIcon sx={{ color: '#3b82f6', fontSize: 24, flexShrink: 0 }} />
          <Box sx={{ minWidth: 0, flex: 1 }}>
            <Typography
              variant="body1"
              sx={{ color: '#e2e8f0', fontWeight: 600, lineHeight: 1.3 }}
              noWrap
            >
              {planetName}
            </Typography>
            {/* Editable label */}
            {editing ? (
              <TextField
                inputRef={labelInputRef}
                value={editValue}
                onChange={(e) => setEditValue(e.target.value)}
                onBlur={saveLabel}
                onKeyDown={handleKeyDown}
                disabled={savingLabel}
                size="small"
                placeholder="Enter label..."
                variant="standard"
                sx={{
                  mt: 0.25,
                  '& .MuiInput-root': {
                    color: '#94a3b8',
                    fontSize: '0.75rem',
                    '&:before': { borderColor: 'rgba(59, 130, 246, 0.3)' },
                    '&:after': { borderColor: '#3b82f6' },
                  },
                  '& .MuiInputBase-input': { py: 0.25 },
                }}
              />
            ) : (
              <Box
                onClick={startEditing}
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 0.5,
                  cursor: 'pointer',
                  '&:hover': { '& .edit-icon': { opacity: 1 } },
                }}
              >
                <Typography
                  variant="caption"
                  sx={{
                    color: label ? '#94a3b8' : '#475569',
                    fontStyle: label ? 'normal' : 'italic',
                  }}
                >
                  {label || 'Add label...'}
                </Typography>
                <EditIcon
                  className="edit-icon"
                  sx={{
                    fontSize: 12,
                    color: '#475569',
                    opacity: label ? 0 : 0.5,
                    transition: 'opacity 0.15s',
                  }}
                />
              </Box>
            )}
          </Box>
        </Box>
        <IconButton onClick={onClose} size="small" sx={{ color: '#64748b', ml: 1 }}>
          <CloseIcon fontSize="small" />
        </IconButton>
      </Box>

      {/* Content */}
      <Box sx={{ flex: 1, overflow: 'auto', px: 2.5, py: 2 }}>
        {loading && (
          <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', py: 8 }}>
            <CircularProgress size={36} sx={{ color: '#3b82f6' }} />
          </Box>
        )}

        {error && !loading && (
          <Box sx={{ textAlign: 'center', py: 8 }}>
            <Typography variant="body2" sx={{ color: '#ef4444' }}>
              {error}
            </Typography>
          </Box>
        )}

        {data && !loading && (
          <>
            {/* Input Requirements (aggregated across all factories) */}
            <SectionHeader
              icon={<FactoryIcon sx={{ fontSize: 16, color: '#3b82f6' }} />}
              label={`Input Requirements (${aggregatedInputs.length})`}
            />
            {aggregatedInputs.length === 0 ? (
              <Typography variant="caption" sx={{ color: '#475569', display: 'block', mb: 2 }}>
                No factory inputs tracked
              </Typography>
            ) : (
              <Box sx={{ mb: 3 }}>
                {aggregatedInputs.map((input) => (
                  <InputRow key={input.typeId} input={input} />
                ))}
              </Box>
            )}

            <Divider sx={{ borderColor: 'rgba(148, 163, 184, 0.08)', mb: 2 }} />

            {/* Current Contents */}
            <SectionHeader
              icon={<InventoryIcon sx={{ fontSize: 16, color: '#3b82f6' }} />}
              label={`Current Contents (${sortedContents.length})`}
            />
            {sortedContents.length === 0 ? (
              <Typography variant="caption" sx={{ color: '#475569' }}>
                Launchpad is empty
              </Typography>
            ) : (
              <Box>
                {sortedContents.map((item) => (
                  <ContentRow key={item.typeId} item={item} />
                ))}
              </Box>
            )}
          </>
        )}
      </Box>
    </Drawer>
  );
}

function SectionHeader({ icon, label }: { icon: React.ReactNode; label: string }) {
  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, mb: 1.5 }}>
      {icon}
      <Typography
        variant="caption"
        sx={{
          color: '#64748b',
          fontWeight: 600,
          textTransform: 'uppercase',
          letterSpacing: 0.5,
        }}
      >
        {label}
      </Typography>
    </Box>
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
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        py: 0.5,
        '&:not(:last-child)': {
          borderBottom: '1px solid rgba(148, 163, 184, 0.05)',
        },
      }}
    >
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, minWidth: 0, flex: 1 }}>
        <img
          src={`https://images.evetech.net/types/${input.typeId}/icon?size=32`}
          alt=""
          width={18}
          height={18}
          style={{ flexShrink: 0 }}
        />
        <Box sx={{ minWidth: 0 }}>
          <Typography variant="caption" sx={{ color: '#cbd5e1', display: 'block' }} noWrap>
            {input.name}
          </Typography>
          <Typography variant="caption" sx={{ color: '#475569', fontSize: '0.65rem' }}>
            {formatNumber(input.consumedPerHour)}/hr
          </Typography>
        </Box>
      </Box>

      <Box sx={{ textAlign: 'right', flexShrink: 0, ml: 1 }}>
        <Typography variant="caption" sx={{ color: '#94a3b8', display: 'block' }}>
          {formatNumber(input.currentStock)}
        </Typography>
        <Typography
          variant="caption"
          sx={{
            color,
            fontWeight: 600,
            fontSize: '0.65rem',
          }}
        >
          {formatDepletion(input.depletionHours)}
        </Typography>
      </Box>
    </Box>
  );
}

function ContentRow({ item }: { item: PiPinContent }) {
  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        py: 0.5,
        '&:not(:last-child)': {
          borderBottom: '1px solid rgba(148, 163, 184, 0.05)',
        },
      }}
    >
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, minWidth: 0 }}>
        <img
          src={`https://images.evetech.net/types/${item.typeId}/icon?size=32`}
          alt=""
          width={18}
          height={18}
          style={{ flexShrink: 0 }}
        />
        <Typography variant="caption" sx={{ color: '#cbd5e1' }} noWrap>
          {item.name || `Type ${item.typeId}`}
        </Typography>
      </Box>
      <Typography variant="caption" sx={{ color: '#94a3b8', fontWeight: 500, flexShrink: 0, ml: 1 }}>
        {formatNumber(item.amount)}
      </Typography>
    </Box>
  );
}
